package ws

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/zhengxiaod/gochat/pkg/utils"
	"sync"
)

type Conn struct {
	ClientManager *ClientManager  // 当前连接属于哪个ClientManager
	UserId        uint64          // 连接所属用户id
	UserIdMutex   sync.RWMutex    // 保护 userId 的锁
	Socket        *websocket.Conn // 用户连接
	sendCh        chan []byte     // 用户要发送的数据
}

func NewConnection(cm *ClientManager, wsConn *websocket.Conn) *Conn {
	return &Conn{
		ClientManager: cm,
		UserId:        0, // 此时用户未登录， userID 为 0
		Socket:        wsConn,
		sendCh:        make(chan []byte, 10),
	}
}

func (c *Conn) Start() {
	// 开启从客户端读取数据流程的 goroutine
	go c.StartReader()

	// 开启用于写回客户端数据流程的 goroutine
	go c.StartWriter()
}

// StartReader 用于从客户端中读取数据
func (c *Conn) StartReader() {
	fmt.Println("[Reader Goroutine is running]")

	for {
		// 阻塞读
		_, data, err := c.Socket.ReadMessage()
		if err != nil {
			fmt.Println("read msg data error ", err)
			return
		}

		// 消息处理
		c.HandlerMessage(data)
	}
}

// StartWriter 向客户端写数据
func (c *Conn) StartWriter() {
	fmt.Println("[Writer Goroutine is running]")

	var err error
	for {
		select {
		case data := <-c.sendCh:
			if err = c.Socket.WriteMessage(websocket.TextMessage, data); err != nil {
				fmt.Println("Send Data error:, ", err, " Conn Writer exit")
				return
			}
		}
	}
}

func (c *Conn) HandlerMessage(data []byte) {
	UpMsg := new(utils.MessageStruct)
	err := json.Unmarshal(data, UpMsg)
	if err != nil {
		fmt.Println("json unmarshal error ", err)
		return
	}

	// 对未登录用户进行拦截
	if UpMsg.Cmd != "Login" && c.GetUserId() == 0 {
		fmt.Println("user not login")
		return
	}

	req := &Req{
		conn: c,
		data: UpMsg.Content,
		f:    nil,
	}

	// 判断cmd类型
	switch UpMsg.Cmd {
	case "Login":
		req.f = req.Login
	case "Message":
		req.f = req.SendMessage
	}

	// 执行对应逻辑
	go req.f()
}

// GetUserId 获取 userId
func (c *Conn) GetUserId() uint64 {
	c.UserIdMutex.RLock()
	defer c.UserIdMutex.RUnlock()

	return c.UserId
}

// SetUserId 设置 UserId
func (c *Conn) SetUserId(userId uint64) {
	c.UserIdMutex.Lock()
	defer c.UserIdMutex.Unlock()

	c.UserId = userId
}

// SendMsg 根据 userId 给相应 socket 发送消息
func (c *Conn) SendMsg(userId uint64, bytes []byte) {
	// 根据 userId 找到对应 socket
	conn := c.ClientManager.GetConn(userId)
	if conn == nil {
		return
	}

	// 发送
	conn.sendCh <- bytes

	return
}
