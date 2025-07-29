package models

import (
	"time"

	"gorm.io/gorm"
)

// EmailLog 邮件日志表
type EmailLog struct {
	gorm.Model
	MessageID   string     `gorm:"uniqueIndex;not null;comment:邮件消息ID" json:"message_id"`
	Subject     string     `gorm:"comment:邮件主题" json:"subject"`
	FromEmail   string     `gorm:"comment:发件人邮箱" json:"from_email"`
	RecipientID uint       `gorm:"comment:转发对象ID" json:"recipient_id"`
	Recipient   Recipient  `gorm:"foreignKey:RecipientID" json:"recipient,omitempty"`
	Status      string     `gorm:"comment:转发状态(success/failed)" json:"status"`
	Error       string     `gorm:"comment:错误信息" json:"error"`
	ForwardedAt *time.Time `gorm:"comment:转发时间" json:"forwarded_at"`
}