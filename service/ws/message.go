package ws

import (
	"encoding/json"
	"fmt"
	"github.com/zhengxiaod/gochat/common"
	"github.com/zhengxiaod/gochat/model"
	"github.com/zhengxiaod/gochat/pkg/utils"
	"log"
	"time"
)

// GetOutputMsg 组装出下行消息
func GetOutputMsg(Cmd string, code int32, message json.RawMessage) ([]byte, error) {
	output := &utils.OutputMessageStruct{
		Cmd:     Cmd,
		Code:    code,
		CodeMsg: common.GetErrorMessage(uint32(code), ""),
		Content: message,
	}

	OutputBytes, err := json.Marshal(output)
	if err != nil {
		fmt.Println("[GetOutputMsg] output marshal err:", err)
		return nil, err
	}
	return OutputBytes, nil
}

func SendToUser(msg *utils.MessageStruct, receiverId uint64) (uint64, error) {
	// 确保服务端消息可靠

	// 获取接受者 seqId
	seq, err := GetUserNextSeq(receiverId)
	if err != nil {
		fmt.Println("[消息处理] 获取 seq 失败,err:", err)
		return 0, err
	}
	msg.SeqId = seq

	// 消息落库
	message := &model.Message{
		UserID:      receiverId,
		SenderID:    msg.SenderId,
		SessionType: 1,
		ReceiverId:  receiverId,
		MessageType: 1,
		Content:     []byte(msg.Content),
		Seq:         msg.SeqId,
		SendTime:    time.Now(),
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}
	err = model.CreateMessage(message)
	if err != nil {
		fmt.Println("[发送消息] 消息落库失败")
		return 0, err
	}

	// 如果是发送给自己的，只落库不推送
	if receiverId == msg.SenderId {
		return seq, nil
	}

	// 组装消息
	// 把msg转为json
	// 将结构体编码为 JSON
	jsonData, err := json.Marshal(msg)
	if err != nil {
		log.Fatalf("Error marshaling to JSON: %v", err)
	}

	bytes, err := GetOutputMsg("Message", int32(common.OK), jsonData)
	if err != nil {
		fmt.Println("[消息处理] GetOutputMsg Marshal error,err:", err)
		return 0, err
	}

	// 进行推送
	return 0, Send(receiverId, bytes)

}

// Send 消息转发
// 是否在线 ---否---> 不进行推送
//    |
//    是
//    ↓
//  消息发送

func Send(receiverId uint64, bytes []byte) error {
	// 查询是否在线
	conn := ConnManager.GetConn(receiverId)
	if conn != nil {
		// 发送消息
		conn.SendMsg(receiverId, bytes)
		fmt.Println("[消息处理]， 发送本地消息给用户, ", receiverId)
		return nil
	}
	return nil
}

// SendToGroup 发送消息到群
func SendToGroup(msg *utils.MessageStruct) error {
	// 获取群成员信息
	userIds, err := model.GetGroupUserIdsByGroupId(msg.ReceiverId)
	if err != nil {
		fmt.Println("[群聊消息处理] 查询失败，err:", err, msg)
		return err
	}

	// userId set
	m := make(map[uint64]struct{}, len(userIds))
	for _, userId := range userIds {
		m[userId] = struct{}{}
	}

	// 检查当前用户是否属于该群
	if _, ok := m[msg.SenderId]; !ok {
		fmt.Println("[群聊消息处理] 用户不属于该群组，msg:", msg)
		return nil
	}

	// 自己不再进行推送
	delete(m, msg.SenderId)

	sendUserIds := make([]uint64, 0, len(m))
	for userId := range m {
		sendUserIds = append(sendUserIds, userId)
	}
	// 批量获取 seqId
	seqs, err := GetUserNextSeqBatch(sendUserIds)
	if err != nil {
		fmt.Println("[批量获取 seq] 失败,err:", err)
		return err
	}

	//  k:userid v:该userId的seq
	sendUserSet := make(map[uint64]uint64, len(seqs))
	for i, userId := range sendUserIds {
		sendUserSet[userId] = seqs[i]
	}

	// 创建 Message 对象，进行落库
	messages := make([]*model.Message, 0, len(m))
	for userId, seq := range sendUserSet {
		messages = append(messages, &model.Message{
			UserID:      userId,
			SenderID:    msg.SenderId,
			SessionType: msg.SessionType,
			ReceiverId:  msg.ReceiverId,
			MessageType: 1,
			Content:     []byte(msg.Content),
			Seq:         seq,
			SendTime:    time.UnixMilli(msg.SendTime),
		})
	}

	err = model.CreateMessage(messages...)
	if err != nil {
		fmt.Println("[消息处理] 消息落库失败")
		return err
	}

	// 组装消息，进行推送
	userId2Msg := make(map[uint64][]byte, len(m))
	for userId, seq := range sendUserSet {
		msg.SeqId = seq

		jsonData, err := json.Marshal(msg)
		if err != nil {
			log.Fatalf("Error marshaling to JSON: %v", err)
		}

		bytes, err := GetOutputMsg("Message", int32(common.OK), jsonData)
		if err != nil {
			fmt.Println("[消息处理] GetOutputMsg Marshal error,err:", err)
			return err
		}
		userId2Msg[userId] = bytes
	}

	// 消息推送
	ConnManager.SendMessageAll(userId2Msg)

	return nil
}
