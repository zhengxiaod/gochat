package model

import (
	"fmt"
	"github.com/zhengxiaod/gochat/pkg/db"
	"time"
)

type Group struct {
	ID         uint64    `gorm:"primary_key;auto_increment;comment:'自增主键'" json:"id"`
	Name       string    `gorm:"not null;comment:'群组名称'" json:"name"`
	OwnerID    uint64    `gorm:"not null;comment:'群主id'" json:"owner_id"`
	CreateTime time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;comment:'创建时间'" json:"create_time"`
	UpdateTime time.Time `gorm:"not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:'更新时间'" json:"update_time"`
}

func (*Group) TableName() string {
	return "group"
}

func CreateGroup(group *Group) error {
	return db.DB.Create(group).Error
}

func GetGroupByGroupName(user1, user2 uint64) (*Group, error) {
	var group Group
	name1 := fmt.Sprintf("%d_%d", user1, user2)
	name2 := fmt.Sprintf("%d_%d", user2, user1)

	err := db.DB.Where("name = ? OR name = ?", name1, name2).First(&group).Error
	if err != nil {
		fmt.Println("GetGroupByGroupName Error:", err)
		return nil, err
	}

	return &group, nil
}

func DeleteGroupByID(id uint64) error {
	return db.DB.Delete(&Group{}, id).Error
}
