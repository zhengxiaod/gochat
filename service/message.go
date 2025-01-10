package service

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/zhengxiaod/gochat/model"
	"github.com/zhengxiaod/gochat/pkg/utils"
	"net/http"
	"strconv"
)

func MessageList(c *gin.Context) {
	GroupID := c.Query("group_id")
	if GroupID == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "房间号不能为空",
		})
		return
	}
	// 判断用户是否属于该房间
	// 使用 ParseUint 将字符串转换为 uint64
	GroupId, err := strconv.ParseUint(GroupID, 10, 64) // 10 表示十进制，64 表示返回值的位数
	if err != nil {
		fmt.Println("转换错误:", err)
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "房间号输入错误",
		})
		return
	}

	// 判断用户是否属于该房间
	uc := c.MustGet("user_claims").(*utils.UserClaims)
	isBelong, err := model.IsBelongToGroup(uc.UserId, GroupId)
	if !isBelong {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "非法访问",
		})
		return
	}

	pageIndex, _ := strconv.ParseInt(c.Query("page_index"), 10, 32)
	pageSize, _ := strconv.ParseInt(c.Query("page_size"), 10, 32)
	// 聊天记录查询
	data, err := model.GetMessageListByGroupID(GroupId, int(pageIndex), int(pageSize))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "系统异常:" + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "数据加载成功",
		"data": gin.H{
			"list": data,
		},
	})

}
