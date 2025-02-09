package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"github.com/zhengxiaod/gochat/pkg/protocol/pb"
	"github.com/zhengxiaod/gochat/pkg/utils"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

const (
	ResendCountMax = 3 // 超时重传最大次数
)

type Client struct {
	conn                 *websocket.Conn
	connMutex            sync.Mutex
	token                string
	userId               uint64
	clientId             uint64
	clientId2Cancel      map[uint64]context.CancelFunc // clientId 到 context 的映射
	clientId2CancelMutex sync.Mutex
	seq                  uint64 // 本地消息最大同步序列号

	sendCh chan []byte // 写入
}

func NewClient(userId, token, host string) *Client {
	// 创建 client
	c := &Client{
		clientId2Cancel: make(map[uint64]context.CancelFunc),
		token:           token,
		userId:          utils.StrToUint64(userId),
		sendCh:          make(chan []byte, 1024),
	}

	// 连接 websocket
	conn, _, err := websocket.DefaultDialer.Dial(host+"/ws", http.Header{})
	if err != nil {
		panic(err)
	}
	c.conn = conn
	// 向 websocket 发送登录请求
	c.Login()

	// 心跳
	go c.Heartbeat()

	// 写
	go c.write()

	// 读
	go c.read()

	return c
}

func NewClientInPB(userId, token, host string) *Client {
	// 创建 client
	c := &Client{
		clientId2Cancel: make(map[uint64]context.CancelFunc),
		token:           token,
		userId:          utils.StrToUint64(userId),
		sendCh:          make(chan []byte, 1024),
	}

	// 连接 websocket
	conn, _, err := websocket.DefaultDialer.Dial(host+"/ws", http.Header{})
	if err != nil {
		panic(err)
	}
	c.conn = conn
	// 向 websocket 发送登录请求
	c.LoginInPB()

	//c.Login()

	// 心跳
	go c.HeartbeatInPB()
	//go c.Heartbeat()

	// 写
	go c.write()

	// 读
	go c.readInPB()
	//go c.read()

	return c
}

// Login websocket 登录
func (c *Client) Login() {
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

func (c *Client) LoginInPB() {
	c.sendMsgInPB(pb.CmdType_CT_Login, &pb.LoginMsg{
		Token: []byte(c.token),
	})
}

func (c *Client) Heartbeat() {
	//  2min 一次
	ticker := time.NewTicker(time.Second * 2 * 60)
	for range ticker.C {
		c.SendMsg("Heartbeat", json.RawMessage{})
		//fmt.Println("发送心跳", time.Now().Format("2006-01-02 15:04:05"))
	}
}

func (c *Client) HeartbeatInPB() {
	ticker := time.NewTicker(time.Minute * 2)
	for range ticker.C {
		c.sendMsgInPB(pb.CmdType_CT_Heartbeat, &pb.HeartbeatMsg{})
	}
}

func (c *Client) write() {
	for {
		select {
		case bytes, ok := <-c.sendCh:
			if !ok {
				return
			}
			if err := c.conn.WriteMessage(websocket.BinaryMessage, bytes); err != nil {
				return
			}
		}
	}
}

func (c *Client) read() {
	for {
		_, bytes, err := c.conn.ReadMessage()
		if err != nil {
			panic(err)
		}

		outputMsg := new(utils.OutputMessageStruct)
		err = json.Unmarshal(bytes, outputMsg)
		if err != nil {
			panic(err)
		}

		// 只收两种，Message 收取下行消息和 ACK，上行消息ACK回复
		switch outputMsg.Cmd {
		case "Message":
			// 计算接收消息数量
			atomic.AddInt64(&receivedMessageCount, 1)
			msgTimer.updateEndTime()

			pushMsg := new(utils.MessageStruct)
			err = json.Unmarshal(outputMsg.Content, pushMsg)
			if err != nil {
				panic(err)
			}
			// 更新 seq
			seq := pushMsg.SeqId
			if c.seq < seq {
				c.seq = seq
			}
		case "ACK": // 收到 ACK
			ackMsg := new(utils.AckMsgStruct)
			err = json.Unmarshal(outputMsg.Content, ackMsg)
			if err != nil {
				panic(err)
			}

			switch ackMsg.AckType {
			case "ACK_Message": // 收到上行消息的 ACK
				// 计算接收消息数量
				atomic.AddInt64(&receivedMessageCount, 1)
				msgTimer.updateEndTime()

				// 取消超时重传
				clientId := ackMsg.ClientId
				c.clientId2CancelMutex.Lock()
				if cancel, ok := c.clientId2Cancel[clientId]; ok {
					// 取消超时重传
					cancel()
					delete(c.clientId2Cancel, clientId)
					//fmt.Println("取消超时重传，clientId:", clientId)
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
}

func (c *Client) readInPB() {
	for {
		_, bytes, err := c.conn.ReadMessage()
		if err != nil {
			panic(err)
		}

		msg := new(pb.Output)
		err = proto.Unmarshal(bytes, msg)
		if err != nil {
			panic(err)
		}

		// 只收两种，Message 收取下行消息和 ACK，上行消息ACK回复
		switch msg.Type {
		case pb.CmdType_CT_Message:
			// 计算接收消息数量
			atomic.AddInt64(&receivedMessageCount, 1)
			msgTimer.updateEndTime()

			pushMsg := new(pb.PushMsg)
			err = proto.Unmarshal(msg.Data, pushMsg)
			if err != nil {
				panic(err)
			}
			// 更新 seq
			seq := pushMsg.Msg.Seq
			if c.seq < seq {
				c.seq = seq
			}
		case pb.CmdType_CT_ACK: // 收到 ACK
			ackMsg := new(pb.ACKMsg)
			err = proto.Unmarshal(msg.Data, ackMsg)
			if err != nil {
				panic(err)
			}

			switch ackMsg.Type {
			case pb.ACKType_AT_Up: // 收到上行消息的 ACK
				// 计算接收消息数量
				atomic.AddInt64(&receivedMessageCount, 1)
				msgTimer.updateEndTime()

				// 取消超时重传
				clientId := ackMsg.ClientId
				c.clientId2CancelMutex.Lock()
				if cancel, ok := c.clientId2Cancel[clientId]; ok {
					// 取消超时重传
					cancel()
					delete(c.clientId2Cancel, clientId)
					//fmt.Println("取消超时重传，clientId:", clientId)
				}
				c.clientId2CancelMutex.Unlock()
				// 更新客户端本地维护的 seq
				seq := ackMsg.Seq
				if c.seq < seq {
					c.seq = seq
				}
			}
		default:
			fmt.Println("未知消息类型")
		}
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
	c.connMutex.Lock()
	defer c.connMutex.Unlock()
	err = c.conn.WriteMessage(websocket.BinaryMessage, bytes)
	if err != nil {
		panic(err)
	}
}

// 客户端向服务端发送上行消息
func (c *Client) sendMsgInPB(cmdType pb.CmdType, msg proto.Message) {
	// 组装顶层数据
	cmdMsg := &pb.Input{
		Type: cmdType,
		Data: nil,
	}
	if msg != nil {
		data, err := proto.Marshal(msg)
		if err != nil {
			panic(err)
		}
		cmdMsg.Data = data
	}

	bytes, err := proto.Marshal(cmdMsg)
	if err != nil {
		panic(err)
	}

	// 发送
	c.sendCh <- bytes
}

// send 发送消息，启动超时重试
func (c *Client) send(chatId int64) {
	message := &utils.MessageStruct{
		SessionType: 2,              // 群聊
		ReceiverId:  uint64(chatId), // 发送到该群
		SenderId:    c.userId,       // 发送者 		// 文本
		ClientId:    c.getClientId(),
		Content:     "文本聊天消息" + utils.Uint64ToStr(c.userId), // 消息
		SendTime:    time.Now().UnixMilli(),                 // 发送时间
	}
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		panic(err)
	}

	// 发送消息
	c.SendMsg("Message", jsonMessage)

	// 启动超时重传
	ctx, cancel := context.WithCancel(context.Background())

	go func(ctx context.Context) {
		maxRetry := ResendCountMax // 最大重试次数
		retryCount := 0
		retryInterval := time.Millisecond * 500 // 重试间隔
		ticker := time.NewTicker(retryInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if retryCount >= maxRetry {
					return
				}
				c.SendMsg("Message", jsonMessage)
				retryCount++
			}
		}
	}(ctx)

	c.clientId2CancelMutex.Lock()
	c.clientId2Cancel[message.ClientId] = cancel
	c.clientId2CancelMutex.Unlock()
}

// sendInPB 发送消息，启动超时重试
func (c *Client) sendInPB(chatId int64) {
	message := &pb.Message{
		SessionType: pb.SessionType_ST_Group,                        // 群聊
		ReceiverId:  uint64(chatId),                                 // 发送到该群
		SenderId:    c.userId,                                       // 发送者
		MessageType: pb.MessageType_MT_Text,                         // 文本
		Content:     []byte("文本聊天消息" + utils.Uint64ToStr(c.userId)), // 消息
		SendTime:    time.Now().UnixMilli(),                         // 发送时间
	}
	UpMsg := &pb.UpMsg{
		Msg:      message,
		ClientId: c.getClientId(),
	}
	// 发送消息
	c.sendMsgInPB(pb.CmdType_CT_Message, UpMsg)

	// 启动超时重传
	ctx, cancel := context.WithCancel(context.Background())

	go func(ctx context.Context) {
		maxRetry := ResendCountMax // 最大重试次数
		retryCount := 0
		retryInterval := time.Millisecond * 500 // 重试间隔
		ticker := time.NewTicker(retryInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if retryCount >= maxRetry {
					return
				}
				c.sendMsgInPB(pb.CmdType_CT_Message, UpMsg)
				retryCount++
			}
		}
	}(ctx)

	c.clientId2CancelMutex.Lock()
	c.clientId2Cancel[UpMsg.ClientId] = cancel
	c.clientId2CancelMutex.Unlock()

}

func (c *Client) getClientId() uint64 {
	c.clientId++
	return c.clientId
}
