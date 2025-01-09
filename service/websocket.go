package service

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/zhengxiaod/gochat/model"
	"github.com/zhengxiaod/gochat/pkg/utils"
	"log"
	"net/http"
	"time"
)

type MessageStruct struct {
	Message string `json:"message"`
	GroupID uint64 `json:"group_id"`
}

var upgrader = websocket.Upgrader{} // use default options

// ws 用于保存所有连接的客户端
var ws = make(map[uint64]*websocket.Conn)

func WebSocketMessage(c *gin.Context) {
	w := c.Writer
	r := c.Request
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "系统异常：" + err.Error(),
		})
		return
	}
	defer conn.Close()

	uc := c.MustGet("user_claims").(*utils.UserClaims)
	ws[uc.UserId] = conn

	for {
		ms := new(MessageStruct)
		err := conn.ReadJSON(ms)
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", ms.Message)
		// 判断用户是否属于消息体的房间
		isBelong, err := model.IsBelongToGroup(uc.UserId, ms.GroupID)
		if err != nil {
			log.Printf("IsBelongToGroup Error:%v\n", err)
			return
		}
		if !isBelong {
			log.Printf("UserIdentity:%v RoomIdentity:%v Not Exits\n", uc.UserId, ms.GroupID)
			return
		}
		// 保存消息
		mb := &model.Message{
			UserID:     uc.UserId,
			GroupID:    ms.GroupID,
			Content:    []byte(ms.Message),
			CreateTime: time.Now(),
			UpdateTime: time.Now(),
		}
		err = model.CreateMessage(mb)
		if err != nil {
			log.Printf("CreateMessage Error:%v\n", err)
			return
		}
		// 获取在特定房间的在线用户
		users, err := model.GetGroupUserIdsByGroupId(ms.GroupID)
		if err != nil {
			log.Printf("GetGroupUserIdsByGroupId Error:%v\n", err)
			return
		}
		// 往客户端发送数据
		for _, user := range users {
			if conn, ok := ws[user]; ok {
				err = conn.WriteMessage(websocket.TextMessage, []byte(ms.Message))
				if err != nil {
					log.Println("write:", err)
					break
				}
			}
		}
	}
}
