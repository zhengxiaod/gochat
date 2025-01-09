package model

import (
	"github.com/zhengxiaod/gochat/pkg/db"
	"time"
)

// Message 消息
type Message struct {
	ID         uint64    `gorm:"primary_key;auto_increment;comment:'自增主键'" json:"id"`
	UserID     uint64    `gorm:"not null;comment:'用户id，指接受者用户id'" json:"user_id"`
	GroupID    uint64    `gorm:"not null;comment:'房间id'"`
	Content    []byte    `gorm:"not null;comment:'消息内容'" json:"content"`
	CreateTime time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;comment:'创建时间'" json:"create_time"`
	UpdateTime time.Time `gorm:"not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:'更新时间'" json:"update_time"`
}

func (Message) TableName() string {
	return "message"
}

func CreateMessage(message *Message) error {
	return db.DB.Create(message).Error
}

func GetMessageListByGroupID(groupID uint64, pageIndex, pageSize int) ([]*Message, error) {
	skip := (pageIndex - 1) * pageSize
	var messages []*Message
	// 修改 SQL 查询以支持分页和排序
	err := db.DB.Where("group_id = ?", groupID).
		Order("create_time DESC"). // 按照 CreateTime 降序排序
		Offset(skip).              // 设置偏移量
		Limit(pageSize).           // 设置限制条数
		Find(&messages).Error
	return messages, err
}
