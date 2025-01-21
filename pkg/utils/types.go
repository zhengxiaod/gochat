package utils

import "encoding/json"

type InputMessageStruct struct {
	Cmd     string          `json:"cmd"`
	Content json.RawMessage `json:"content"`
}

type LoginStruct struct {
	Token string `json:"token"`
}

type SyncStruct struct {
	SeqId uint64 `json:"seq_id"`
}

type MessageStruct struct {
	SeqId       uint64 `json:"seq_id"`
	SenderId    uint64 `json:"sender_id"`
	ClientId    uint64 `json:"client_id"`
	SessionType int8   `json:"session_type"`
	ReceiverId  uint64 `json:"receiver_id"` // 如果是单聊则是接收者id，如果是群聊则是groupid
	Content     string `json:"content"`
	SendTime    int64  `json:"send_time"`
}

type OutputMessageStruct struct {
	Cmd     string          `json:"cmd"`
	Code    int32           `json:"code"`
	CodeMsg string          `json:"code_msg"`
	Content json.RawMessage `json:"content"`
}

type AckMsgStruct struct {
	AckType  string `json:"type"`
	SeqId    uint64 `json:"seq_id"`
	ClientId uint64 `json:"client_id"`
}

type OutputSyncMsgStruct struct {
	Content []MessageStruct `json:"content"`
	HasMore bool            `json:"has_more"`
}
