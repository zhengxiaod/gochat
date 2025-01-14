package service

import (
	"github.com/gin-gonic/gin"
	"github.com/zhengxiaod/gochat/model"
	"github.com/zhengxiaod/gochat/pkg/utils"
	"net/http"
)

// CreateGroup 创建群聊
func CreateGroup(c *gin.Context) {
	// 参数校验
	name := c.PostForm("name")
	idsStr := c.PostFormArray("ids") // 群成员 id，不包括群创建者
	if name == "" || len(idsStr) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "参数不正确",
		})
		return
	}
	ids := make([]uint64, 0, len(idsStr)+1)
	for i := range idsStr {
		ids = append(ids, utils.StrToUint64(idsStr[i]))
	}
	// 获取用户信息
	uc := c.MustGet("user_claims").(*utils.UserClaims)
	ids = append(ids, uc.UserId)

	// 获取 ids 用户信息
	ids, err := model.GetUserIdByIds(ids)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "系统错误:" + err.Error(),
		})
		return
	}

	// 创建群组
	group := &model.Group{
		Name:    name,
		OwnerID: uc.UserId,
	}
	err = model.CreateGroup(group, ids)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "系统错误:" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "群组创建成功",
		"data": gin.H{
			"id": utils.Uint64ToStr(group.ID),
		},
	})
}
