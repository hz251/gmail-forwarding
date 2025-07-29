package models

import (
	"gorm.io/gorm"
)

// Recipient 转发对象表
type Recipient struct {
	gorm.Model
	Name  string `gorm:"uniqueIndex;not null;comment:转发对象姓名" json:"name"`
	Email string `gorm:"not null;comment:转发对象邮箱" json:"email"`
}