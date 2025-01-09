package router

import (
	"github.com/gin-gonic/gin"
	"github.com/zhengxiaod/gochat/pkg/middlewares"
	"github.com/zhengxiaod/gochat/service"
)

func HttpRouter() *gin.Engine {
	r := gin.Default()

	// 用户注册
	r.POST("/register", service.Register)

	//用户登录
	r.POST("/login", service.Login)

	auth := r.Group("", middlewares.AuthCheck())
	{
		// 发送、接收消息
		auth.GET("/websocket/message", service.WebSocketMessage)

		// 获取聊天记录列表
		auth.GET("/message/list", service.MessageList)
	}

	return r
}
