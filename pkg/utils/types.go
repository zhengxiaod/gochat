package utils

import "encoding/json"

type MessageStruct struct {
	Cmd     string          `json:"cmd"`
	Content json.RawMessage `json:"content"`
}

type LoginStruct struct {
	Token string `json:"token"`
}

type SendMsgStruct struct {
	ReceiverId uint64 `json:"receiver_id"`
	Content    string `json:"content"`
}
