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
		// 个人账号详情
		auth.GET("/user/detail", service.UserDetail)

		// 查询指定用户详情
		auth.GET("/user/list", service.UserQuery)

		// 发送、接收消息
		auth.GET("/websocket/message", service.WebSocketMessage)

		// 获取聊天记录列表
		auth.GET("/message/list", service.MessageList)

		// 添加好友
		auth.POST("/user/add", service.UserAdd)

		// 删除好友
		auth.DELETE("/user/delete", service.UserDelete)
	}

	return r
}
