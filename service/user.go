package service

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/zhengxiaod/gochat/model"
	"github.com/zhengxiaod/gochat/pkg/utils"
	"log"
	"net/http"
	"time"
)

type UserQueryResult struct {
	PhoneNumber string `json:"phone_number"`
	Nickname    string `json:"nickname"`
	IsFriend    bool   `json:"is_friend"` // 是否是好友 【true-是，false-否】
}

func Register(c *gin.Context) {
	// 获取参数并验证
	phoneNumber := c.PostForm("phone_number")
	nickname := c.PostForm("nickname")
	password := c.PostForm("password")
	if phoneNumber == "" || password == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "参数不正确",
		})
		return
	}
	// 查询手机号是否已存在
	cnt, err := model.GetUserCountByPhone(phoneNumber)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "系统错误:" + err.Error(),
		})
		return
	}
	if cnt > 0 {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "账号已被注册",
		})
		return
	}
	// 插入用户信息
	ub := &model.User{
		PhoneNumber: phoneNumber,
		Nickname:    nickname,
		Password:    utils.GetMd5(password),
	}
	err = model.CreateUser(ub)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "系统错误" + err.Error(),
		})
		return
	}

	// 生成 token
	token, err := utils.GenerateToken(ub.ID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "系统错误:" + err.Error(),
		})
		return
	}

	// 发放 token
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "登录成功",
		"data": gin.H{
			"token": token,
			"id":    utils.Uint64ToStr(ub.ID),
		},
	})
}

func Login(c *gin.Context) {
	// 验证参数
	phoneNumber := c.PostForm("phone_number")
	password := c.PostForm("password")
	if phoneNumber == "" || password == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "参数不正确",
		})
		return
	}

	// 验证账号名和密码是否正确
	ub, err := model.GetUserByPhoneAndPassword(phoneNumber, utils.GetMd5(password))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "手机号或密码错误",
		})
		return
	}
	// 生成 token
	token, err := utils.GenerateToken(ub.ID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "系统错误:" + err.Error(),
		})
		return
	}

	// 发放 token
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "登录成功",
		"data": gin.H{
			"token":   token,
			"user_id": utils.Uint64ToStr(ub.ID),
		},
	})
}

func UserDetail(c *gin.Context) {
	u, _ := c.Get("user_claims")
	uc := u.(*utils.UserClaims)
	userDetail, err := model.GetUserByUserId(uc.UserId)
	if err != nil {
		log.Printf("[DB ERROR]:%v\n", err)
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "数据查询异常",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "数据加载成功",
		"data": userDetail,
	})
}

func UserQuery(c *gin.Context) {
	account := c.Query("phone_number")
	if account == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "参数不正确",
		})
		return
	}
	uc := c.MustGet("user_claims").(*utils.UserClaims)
	userQuery, err := model.GetUserByPhone(account)
	if err != nil {
		log.Printf("[DB ERROR]:%v\n", err)
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "数据查询异常",
		})
		return
	}
	data := UserQueryResult{
		PhoneNumber: userQuery.PhoneNumber,
		Nickname:    userQuery.Nickname,
		IsFriend:    false,
	}

	// 判断是否是好友
	if model.JudgeUserIsFriend(userQuery.ID, uc.UserId) {
		data.IsFriend = true
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "数据加载成功",
		"data": data,
	})
}

func UserAdd(c *gin.Context) {
	account := c.PostForm("phone_number")
	if account == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "参数不正确",
		})
		return
	}
	ub, err := model.GetUserByPhone(account)
	if err != nil {
		log.Printf("[DB ERROR]:%v\n", err)
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "数据查询异常",
		})
		return
	}
	uc := c.MustGet("user_claims").(*utils.UserClaims)
	// 判断是否是好友
	if model.JudgeUserIsFriend(ub.ID, uc.UserId) {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "互为好友，不可重复添加",
		})
		return
	}

	// 添加好友
	// 添加私聊房间Group
	// 私聊房间Name: 用户ID_好友ID
	rb := &model.Group{
		Name:       fmt.Sprintf("%d_%d", uc.UserId, ub.ID),
		OwnerID:    uc.UserId,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}
	if err = model.CreateGroup(rb); err != nil {
		log.Printf("[DB ERROR]:%v\n", err)
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "数据查询异常",
		})
		return
	}

	// 保存用户与房间的关联记录group_user
	// 创建两条
	gu := &model.GroupUser{
		GroupID:    rb.ID,
		UserID:     uc.UserId,
		GroupType:  1,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}
	if err = model.CreateGroupUser(gu); err != nil {
		log.Printf("[DB ERROR]:%v\n", err)
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "数据库异常",
		})
		return
	}

	gu2 := &model.GroupUser{
		GroupID:    rb.ID,
		UserID:     ub.ID,
		GroupType:  1,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}
	if err = model.CreateGroupUser(gu2); err != nil {
		log.Printf("[DB ERROR]:%v\n", err)
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "数据库异常",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "添加成功",
	})
}

func UserDelete(c *gin.Context) {
	account := c.PostForm("phone_number")
	if account == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "参数不正确",
		})
		return
	}
	ub, err := model.GetUserByPhone(account)
	if err != nil {
		log.Printf("[DB ERROR]:%v\n", err)
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "数据查询异常",
		})
		return
	}
	uc := c.MustGet("user_claims").(*utils.UserClaims)
	// 判断是否是好友
	if !model.JudgeUserIsFriend(ub.ID, uc.UserId) {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "不为好友，无法删除",
		})
		return
	}
	// 获取group
	group, err := model.GetGroupByGroupName(uc.UserId, ub.ID)
	if err != nil {
		log.Printf("[DB ERROR]:%v\n", err)
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "数据查询异常",
		})
		return
	}

	// 删除group
	if err = model.DeleteGroupByID(group.ID); err != nil {
		log.Printf("[DB ERROR]:%v\n", err)
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "系统异常",
		})
		return
	}

	// 删除group_user关联关系
	if err = model.DeleteGroupUserByGroupID(group.ID); err != nil {
		log.Printf("[DB ERROR]:%v\n", err)
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "系统异常",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "删除成功",
	})
}
