package model

import (
	"fmt"
	"github.com/zhengxiaod/gochat/pkg/db"
	"time"
)

type GroupUser struct {
	ID         uint64    `gorm:"primary_key;auto_increment;comment:'自增主键'" json:"id"`
	GroupID    uint64    `gorm:"not null;comment:'组id'" json:"group_id"`
	UserID     uint64    `gorm:"not null;comment:'用户id'" json:"user_id"`
	GroupType  int       `gorm:"not null;comment:'房间类型'" json:"group_type"` // 房间 类型 【1-独聊房间 2-群聊房间】
	CreateTime time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;comment:'创建时间'" json:"create_time"`
	UpdateTime time.Time `gorm:"not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:'更新时间'" json:"update_time"`
}

func (GroupUser) TableName() string {
	return "group_user"
}

func CreateGroupUser(gu *GroupUser) error {
	return db.DB.Create(gu).Error
}

// IsBelongToGroup 验证用户是否属于群
func IsBelongToGroup(userId, groupId uint64) (bool, error) {
	var cnt int64
	err := db.DB.Model(&GroupUser{}).
		Where("user_id = ? and group_id = ?", userId, groupId).
		Count(&cnt).Error
	return cnt > 0, err
}

func GetGroupUserIdsByGroupId(groupId uint64) ([]uint64, error) {
	var ids []uint64
	err := db.DB.Model(&GroupUser{}).Where("group_id = ?", groupId).Pluck("user_id", &ids).Error
	return ids, err
}

func JudgeUserIsFriend(userId, friendId uint64) bool {
	// todo: 查询userId下的group type为1的group
	// todo：查询friendId下的group type为1的group
	// todo: 判断两个group是否有交集
	// todo: 有交集则返回true

	var userGroups []GroupUser
	var friendGroups []GroupUser

	// 查询 userId 下的 group type 为 1 的群组
	err := db.DB.Model(&GroupUser{}).
		Where("user_id = ? AND group_type = ?", userId, 1).
		Find(&userGroups).Error
	if err != nil {
		fmt.Println("Error querying user groups:", err)
		return false
	}

	// 查询 friendId 下的 group type 为 1 的群组
	err = db.DB.Model(&GroupUser{}).
		Where("user_id = ? AND group_type = ?", friendId, 1).
		Find(&friendGroups).Error
	if err != nil {
		fmt.Println("Error querying friend groups:", err)
		return false
	}

	// 判断两个群组是否有交集
	userGroupIDs := make(map[uint64]struct{})
	for _, group := range userGroups {
		userGroupIDs[group.GroupID] = struct{}{}
	}

	for _, group := range friendGroups {
		if _, exists := userGroupIDs[group.GroupID]; exists {
			return true // 找到交集，返回 true
		}
	}

	return false // 没有交集，返回 false
}

func DeleteGroupUserByGroupID(groupId uint64) error {
	return db.DB.Where("group_id = ?", groupId).Delete(&GroupUser{}).Error
}
