package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/zhengxiaod/gochat/pkg/utils"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

const (
	httpAddr       = "http://localhost:8080"
	websocketAddr  = "ws://localhost:8081"
	ResendCountMax = 3 // 超时重传最大次数
)

type Client struct {
	conn                 *websocket.Conn
	token                string
	userId               uint64
	clientId             uint64
	clientId2Cancel      map[uint64]context.CancelFunc // clientId 到 context 的映射
	clientId2CancelMutex sync.Mutex
	seq                  uint64 // 本地消息最大同步序列号
}

// websocket 客户端
func main() {
	// http 登录，获取 token
	client := Login()

	// 连接 websocket 服务
	client.Start()
}

func (c *Client) Start() {
	// 连接 websocket
	conn, _, err := websocket.DefaultDialer.Dial(websocketAddr+"/ws", http.Header{})
	if err != nil {
		panic(err)
	}
	c.conn = conn

	fmt.Println("与 websocket 建立连接")
	// 向 websocket 发送登录请求
	c.Login()

	// 心跳
	go c.Heartbeat()

	time.Sleep(time.Millisecond * 100)

	// 离线消息同步
	go c.Sync()

	// 收取消息
	go c.Receive()

	time.Sleep(time.Millisecond * 100)

	c.ReadLine()
}

// ReadLine 读取用户消息并发送
func (c *Client) ReadLine() {
	var (
		msg         string
		receiverId  uint64
		sessionType int8
		clientId    uint64
	)

	readLine := func(hint string, v interface{}) {
		fmt.Println(hint)
		_, err := fmt.Scanln(v)
		if err != nil {
			panic(err)
		}
	}
	for {
		readLine("发送消息", &msg)
		readLine("接收人id(user_id/group_id)：", &receiverId)
		readLine("发送消息类型(1-单聊、2-群聊)：", &sessionType)
		readLine("clientID：", &clientId)
		UpMsg := &utils.MessageStruct{
			SessionType: sessionType,
			ReceiverId:  receiverId,
			SenderId:    c.userId,
			ClientId:    clientId,
			SeqId:       c.seq,
			Content:     msg,
			SendTime:    time.Now().UnixMilli(),
		}

		// 将内容转换为 JSON
		contentJSON, err := json.Marshal(UpMsg)
		if err != nil {
			fmt.Println("Error marshalling content:", err)
			return
		}

		c.SendMsg("Message", contentJSON)

		// 启动超时重传
		ctx, cancel := context.WithCancel(context.Background())

		go func(ctx context.Context) {
			maxRetry := ResendCountMax // 最大重试次数
			retryCount := 0
			retryInterval := time.Millisecond * 100 // 重试间隔
			ticker := time.NewTicker(retryInterval)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					fmt.Println("收到 ACK，不再重试")
					return
				case <-ticker.C:
					if retryCount >= maxRetry {
						fmt.Println("达到最大超时次数，不再重试")
						// TODO 进行消息发送失败处理
						return
					}
					fmt.Println("消息超时 msg:", msg, "，第", retryCount+1, "次重试")
					c.SendMsg("Message", contentJSON)
					retryCount++
				}
			}
		}(ctx)

		c.clientId2CancelMutex.Lock()
		c.clientId2Cancel[UpMsg.ClientId] = cancel
		c.clientId2CancelMutex.Unlock()

		time.Sleep(time.Second)
	}
}

// Login websocket 登录
func (c *Client) Login() {
	fmt.Println("websocket login...")
	// 组装底层数据
	loginMsg := &utils.LoginStruct{
		Token: c.token,
	}
	// 将内容转换为 JSON
	contentJSON, err := json.Marshal(loginMsg)
	if err != nil {
		fmt.Println("Error marshalling content:", err)
		return
	}
	c.SendMsg("Login", contentJSON)
}

func (c *Client) Heartbeat() {
	//  2min 一次
	ticker := time.NewTicker(time.Second * 2 * 60)
	for range ticker.C {
		c.SendMsg("Heartbeat", json.RawMessage{})
		//fmt.Println("发送心跳", time.Now().Format("2006-01-02 15:04:05"))
	}
}

func (c *Client) Sync() {
	// 组装底层数据
	syncMsg := &utils.SyncStruct{
		SeqId: c.seq,
	}
	// 将内容转换为 JSON
	contentJSON, err := json.Marshal(syncMsg)
	if err != nil {
		fmt.Println("Error marshalling content:", err)
		return
	}

	c.SendMsg("Sync", contentJSON)
}

func (c *Client) Receive() {
	for {
		_, bytes, err := c.conn.ReadMessage()
		if err != nil {
			panic(err)
		}
		c.HandlerMessage(bytes)
	}
}

// HandlerMessage 客户端消息处理
func (c *Client) HandlerMessage(bytes []byte) {
	outputMsg := new(utils.OutputMessageStruct)
	err := json.Unmarshal(bytes, outputMsg)
	if err != nil {
		panic(err)
	}
	// 将顶层消息以 JSON 格式打印
	outputJson, _ := json.MarshalIndent(outputMsg, "", "  ")
	fmt.Println("收到顶层 OutPut 消息：", string(outputJson))

	switch outputMsg.Cmd {
	case "Sync":
		syncMsg := new(utils.OutputSyncMsgStruct)
		err = json.Unmarshal(outputMsg.Content, syncMsg)
		if err != nil {
			panic(err)
		}

		seq := c.seq
		for _, message := range syncMsg.Content {
			fmt.Println("收到离线消息：", message)
			if seq < message.SeqId {
				seq = message.SeqId
			}
		}
		c.seq = seq
		// 如果还有，继续拉取
		if syncMsg.HasMore {
			c.Sync()
		}
	case "Message":
		pushMsg := new(utils.MessageStruct)
		err = json.Unmarshal(outputMsg.Content, pushMsg)
		if err != nil {
			panic(err)
		}
		fmt.Printf("收到消息：%s, 发送人id：%d, 会话类型：%d, 接收时间:%s\n", pushMsg.Content, pushMsg.SenderId, pushMsg.SessionType, time.Now().Format("2006-01-02 15:04:05"))
		// 更新 seq
		seq := pushMsg.SeqId
		if c.seq < seq {
			c.seq = seq
		}
		fmt.Println("更新 seq:", c.seq)
	case "ACK": // 收到 ACK
		ackMsg := new(utils.AckMsgStruct)
		err = json.Unmarshal(outputMsg.Content, ackMsg)
		if err != nil {
			panic(err)
		}

		switch ackMsg.AckType {
		case "ACK_Login":
			fmt.Println("登录成功")
		case "ACK_Message": // 收到上行消息的 ACK
			// 取消超时重传
			clientId := ackMsg.ClientId
			c.clientId2CancelMutex.Lock()
			if cancel, ok := c.clientId2Cancel[clientId]; ok {
				// 取消超时重传
				cancel()
				delete(c.clientId2Cancel, clientId)
				fmt.Println("取消超时重传，clientId:", clientId)
			}
			c.clientId2CancelMutex.Unlock()
			// 更新客户端本地维护的 seq
			seq := ackMsg.SeqId
			if c.seq < seq {
				c.seq = seq
			}
		}
	default:
		fmt.Println("未知消息类型")
	}
}

// SendMsg 客户端向服务端发送上行消息
func (c *Client) SendMsg(Cmd string, content json.RawMessage) {
	// 组装顶层数据
	cmdMsg := &utils.InputMessageStruct{
		Cmd:     Cmd,
		Content: nil,
	}
	if len(content) > 0 {
		data, err := json.Marshal(content)
		if err != nil {
			panic(err)
		}
		cmdMsg.Content = data
	}

	bytes, err := json.Marshal(cmdMsg)
	if err != nil {
		panic(err)
	}

	// 发送
	err = c.conn.WriteMessage(websocket.BinaryMessage, bytes)
	if err != nil {
		panic(err)
	}
}

// Login 用户http登录获取 token
func Login() *Client {
	// 读取 phone_number 和 password 参数
	var phoneNumber, password string
	fmt.Print("Enter phone_number: ")
	fmt.Scanln(&phoneNumber)
	fmt.Print("Enter password: ")
	fmt.Scanln(&password)

	// 创建一个 url.Values 对象，并将 phone_number 和 password 参数添加到其中
	data := url.Values{}
	data.Set("phone_number", phoneNumber)
	data.Set("password", password)

	// 向服务器发送 POST 请求
	resp, err := http.PostForm(httpAddr+"/login", data)
	if err != nil {
		fmt.Println("Error sending HTTP request:", err)
		panic(err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Unexpected HTTP status code: %d\n", resp.StatusCode)
		panic(err)
	}

	// 读取返回数据
	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	// 获取 token
	var respData struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Token  string `json:"token"`
			UserId string `json:"user_id"`
		} `json:"data"`
	}
	err = json.Unmarshal(bodyData, &respData)
	if err != nil {
		panic(err)
	}

	if respData.Code != 200 {
		panic(fmt.Sprintf("登录失败, %s", respData))
	}
	// 获取客户端 seq 序列号
	var seq uint64
	fmt.Print("Enter seq: ")
	fmt.Scanln(&seq)

	client := &Client{
		clientId2Cancel: make(map[uint64]context.CancelFunc),
		seq:             seq,
	}

	client.token = respData.Data.Token
	clientStr := respData.Data.UserId
	client.userId, err = strconv.ParseUint(clientStr, 10, 64)
	if err != nil {
		panic(err)
	}

	fmt.Println("client code:", respData.Code, " ", respData.Msg)
	fmt.Println("token:", client.token, "userId", client.userId)
	return client
}
