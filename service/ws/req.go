package ws

import (
	"encoding/json"
	"fmt"
	"github.com/zhengxiaod/gochat/model"
	"github.com/zhengxiaod/gochat/pkg/utils"
	"time"
)

// Handler 路由函数
type Handler func()

// Req 请求
type Req struct {
	conn *Conn   // 连接
	data []byte  // 客户端发送的请求数据
	f    Handler // 该请求需要执行的路由函数
}

func (r *Req) Login() {
	// 检查用户是否已登录 只能防止同一个连接多次调用 Login
	if r.conn.GetUserId() != 0 {
		fmt.Println("[用户登录] 用户已登录")
		return
	}

	loginMsg := new(utils.LoginStruct)
	err := json.Unmarshal(r.data, loginMsg)
	if err != nil {
		fmt.Println("[用户登录] unmarshal error,err:", err)
		return
	}
	// 登录校验
	userClaims, err := utils.AnalyseToken(loginMsg.Token)
	if err != nil {
		fmt.Println("[用户登录] AnalyseToken err:", err)
		return
	}

	// 设置 user_id
	r.conn.SetUserId(userClaims.UserId)

	// 加入到 connMap 中
	r.conn.ClientManager.AddConn(userClaims.UserId, r.conn)

	// 告诉客户端登陆成功
	r.conn.SendMsg(userClaims.UserId, []byte("login success"))
}

func (r *Req) SendMessage() {

	Msg := new(utils.SendMsgStruct)

	err := json.Unmarshal(r.data, Msg)
	if err != nil {
		fmt.Println("[发送消息] unmarshal error,err:", err)
		return
	}

	// 查看双方是否为好友
	isFriend, err := model.IsFriend(r.conn.GetUserId(), Msg.ReceiverId)
	if !isFriend {
		fmt.Println("[发送消息] 不是好友")
		return
	}

	// 消息落库
	message := &model.Message{
		UserID:      Msg.ReceiverId,
		SenderID:    r.conn.GetUserId(),
		SessionType: 1,
		ReceiverId:  Msg.ReceiverId,
		MessageType: 1,
		Content:     []byte(Msg.Content),
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}

	err = model.CreateMessage(message)
	if err != nil {
		fmt.Println("[发送消息] 消息落库失败")
		return
	}

	// 发送消息
	r.conn.SendMsg(Msg.ReceiverId, []byte(Msg.Content))

}
