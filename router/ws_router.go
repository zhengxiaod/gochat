package router

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/zhengxiaod/gochat/service/ws"
)

var upgrader = websocket.Upgrader{}

func WSRouter() *gin.Engine {

	cm := ws.NewClientManager()

	r := gin.Default()

	r.GET("/ws", func(c *gin.Context) {
		// 升级协议  http -> websocket
		WsConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			fmt.Println("websocket conn err :", err)
			return
		}

		// 初始化连接
		conn := ws.NewConnection(cm, WsConn)

		// 开启读写线程
		go conn.Start()
	})

	return r
}
