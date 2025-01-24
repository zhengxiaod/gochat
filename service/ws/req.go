package ws

import (
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/zhengxiaod/gochat/common"
	"github.com/zhengxiaod/gochat/model"
	"github.com/zhengxiaod/gochat/pkg/protocol/pb"
	"github.com/zhengxiaod/gochat/pkg/utils"
	"log"
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

	// LoginACK
	LoginAck := &utils.AckMsgStruct{
		AckType: "ACK_Login",
	}
	// 将结构体编码为 JSON
	jsonData, err := json.Marshal(LoginAck)
	if err != nil {
		log.Fatalf("Error marshaling to JSON: %v", err)
	}
	bytes, err := GetOutputMsg("ACK", int32(common.OK), jsonData)

	if err != nil {
		fmt.Println("[用户登录] proto.Marshal err:", err)
		return
	}

	// 告诉客户端登陆成功
	r.conn.SendMsg(userClaims.UserId, bytes)
}

func (r *Req) LoginInPB() {
	// 检查用户是否已登录 只能防止同一个连接多次调用 Login
	if r.conn.GetUserId() != 0 {
		fmt.Println("[用户登录] 用户已登录")
		return
	}

	// 消息解析 proto string -> struct
	loginMsg := new(pb.LoginMsg)
	err := proto.Unmarshal(r.data, loginMsg)
	if err != nil {
		fmt.Println("[用户登录] unmarshal error,err:", err)
		return
	}
	// 登录校验
	userClaims, err := utils.AnalyseToken(string(loginMsg.Token))
	if err != nil {
		fmt.Println("[用户登录] AnalyseToken err:", err)
		return
	}

	// 设置 user_id
	r.conn.SetUserId(userClaims.UserId)

	// 加入到 connMap 中
	r.conn.ClientManager.AddConn(userClaims.UserId, r.conn)

	// 回复ACK
	bytes, err := GetOutputMsgInPB(pb.CmdType_CT_ACK, int32(common.OK), &pb.ACKMsg{Type: pb.ACKType_AT_Login})
	if err != nil {
		fmt.Println("[用户登录] proto.Marshal err:", err)
		return
	}

	// 回复发送 Login 请求的客户端
	r.conn.SendMsg(userClaims.UserId, bytes)
}

func (r *Req) SendMessage() {

	Msg := new(utils.MessageStruct)

	err := json.Unmarshal(r.data, Msg)
	if err != nil {
		fmt.Println("[发送消息] unmarshal error,err:", err)
		return
	}

	// 实现上行消息可靠性，clientID保证
	if !r.conn.CompareAndIncrClientID(Msg.ClientId) {
		fmt.Println("不是想要收到的 clientID，不进行处理, msg:", Msg)
		return
	}

	if Msg.SenderId != r.conn.GetUserId() {
		fmt.Println("[消息处理] 发送者有误")
		return
	}

	// 给自己发一份，消息落库但是不推送
	// 返回seqID用于保证消息顺序
	seq, err := SendToUser(Msg, Msg.SenderId)
	if err != nil {
		fmt.Println("[消息处理] send to 自己出错, err:", err)
		return
	}

	// 单聊、群聊
	switch Msg.SessionType {
	case 1:
		// 查看双方是否为好友
		isFriend, err := model.IsFriend(r.conn.GetUserId(), Msg.ReceiverId)
		if err != nil {
			fmt.Println("[发送消息] 查询好友失败")
			return
		}
		if !isFriend {
			fmt.Println("[发送消息] 不是好友")
			return
		}
		_, err = SendToUser(Msg, Msg.ReceiverId)
	case 2:
		err = SendToGroup(Msg)
	default:
		fmt.Println("[消息处理] 会话类型错误")
		return
	}

	// 发送ACK
	// 回复发送上行消息的客户端 ACK
	// MessageAck
	MessageAck := &utils.AckMsgStruct{
		AckType:  "ACK_Message",
		ClientId: Msg.ClientId,
		SeqId:    seq,
	}
	// 将结构体编码为 JSON
	jsonData, err := json.Marshal(MessageAck)
	if err != nil {
		log.Fatalf("Error marshaling to JSON: %v", err)
	}
	ackBytes, err := GetOutputMsg("ACK", int32(common.OK), jsonData)
	if err != nil {
		fmt.Println("[发送ACK消息] proto.Marshal err:", err)
		return
	}
	// 回复发送 Message 请求的客户端 A
	r.conn.SendMsg(Msg.SenderId, ackBytes)
}

// SendMessageInPB 消息处理，处理客户端发送给服务端的消息
// A 发送消息给服务端，服务端收到消息处理后发给 B
// 包括：单聊、群聊
func (r *Req) SendMessageInPB() {
	// 消息解析 proto string -> struct
	msg := new(pb.UpMsg)
	err := proto.Unmarshal(r.data, msg)
	if err != nil {
		fmt.Println("[消息处理] unmarshal error,err:", err)
		return
	}

	// 实现消息可靠性
	if !r.conn.CompareAndIncrClientID(msg.ClientId) {
		fmt.Println("不是想要收到的 clientID，不进行处理, msg:", msg)
		return
	}

	if msg.Msg.SenderId != r.conn.GetUserId() {
		fmt.Println("[消息处理] 发送者有误")
		return
	}

	// 单聊不能发给自己
	if msg.Msg.SessionType == pb.SessionType_ST_Single && msg.Msg.ReceiverId == r.conn.GetUserId() {
		fmt.Println("[消息处理] 接收者有误")
		return
	}

	// 给自己发一份，消息落库但是不推送
	seq, err := SendToUserInPB(msg.Msg, msg.Msg.SenderId)
	if err != nil {
		fmt.Println("[消息处理] send to 自己出错, err:", err)
		return
	}

	// 单聊、群聊
	switch msg.Msg.SessionType {
	case pb.SessionType_ST_Single:
		_, err = SendToUserInPB(msg.Msg, msg.Msg.ReceiverId)
	case pb.SessionType_ST_Group:
		err = SendToGroupInPB(msg.Msg)
	default:
		fmt.Println("[消息处理] 会话类型错误")
		return
	}
	if err != nil {
		fmt.Println("[消息处理] 系统错误")
		return
	}

	// 回复发送上行消息的客户端 ACK
	ackBytes, err := GetOutputMsgInPB(pb.CmdType_CT_ACK, common.OK, &pb.ACKMsg{
		Type:     pb.ACKType_AT_Up,
		ClientId: msg.ClientId, // 回复客户端，当前已 ACK 的消息
		Seq:      seq,          // 回复客户端当前其 seq
	})
	if err != nil {
		fmt.Println("[消息处理] proto.Marshal err:", err)
		return
	}
	// 回复发送 Message 请求的客户端 A
	r.conn.SendMsg(msg.Msg.SenderId, ackBytes)
}

func (r *Req) Heartbeat() {
	// TODO 更新当前用户状态，不做回复
}

// Sync  消息同步，拉取离线消息
func (r *Req) Sync() {
	msg := new(utils.SyncStruct)
	err := json.Unmarshal(r.data, msg)
	if err != nil {
		fmt.Println("[离线消息] unmarshal error,err:", err)
		return
	}

	// 根据 seq 查询，得到比 seq 大的用户消息
	messages, hasMore, err := model.ListByUserIdAndSeq(r.conn.GetUserId(), msg.SeqId, model.MessageLimit)
	if err != nil {
		fmt.Println("[离线消息] model.ListByUserIdAndSeq error, err:", err)
		return
	}
	jsonMessage := model.MessagesToJson(messages)
	OutputSyncMsg := utils.OutputSyncMsgStruct{
		Content: jsonMessage,
		HasMore: hasMore,
	}
	// 将结构体编码为 JSON
	jsonData, err := json.Marshal(OutputSyncMsg)
	if err != nil {
		log.Fatalf("Error marshaling to JSON: %v", err)
	}

	syncBytes, err := GetOutputMsg("Sync", int32(common.OK), jsonData)
	if err != nil {
		fmt.Println("[离线消息] proto.Marshal err:", err)
		return
	}
	// 回复
	r.conn.SendMsg(r.conn.GetUserId(), syncBytes)
}

// SyncInPB  消息同步，拉取离线消息
func (r *Req) SyncInPB() {
	msg := new(pb.SyncInputMsg)
	err := proto.Unmarshal(r.data, msg)
	if err != nil {
		fmt.Println("[离线消息] unmarshal error,err:", err)
		return
	}

	// 根据 seq 查询，得到比 seq 大的用户消息
	messages, hasMore, err := model.ListByUserIdAndSeq(r.conn.GetUserId(), msg.Seq, model.MessageLimit)
	if err != nil {
		fmt.Println("[离线消息] model.ListByUserIdAndSeq error, err:", err)
		return
	}
	pbMessage := model.MessagesToPB(messages)

	ackBytes, err := GetOutputMsgInPB(pb.CmdType_CT_Sync, int32(common.OK), &pb.SyncOutputMsg{
		Messages: pbMessage,
		HasMore:  hasMore,
	})
	if err != nil {
		fmt.Println("[离线消息] proto.Marshal err:", err)
		return
	}
	// 回复
	r.conn.SendMsg(r.conn.GetUserId(), ackBytes)
}
