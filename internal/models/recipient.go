package models

import (
	"gorm.io/gorm"
)

// Recipient 转发对象表
type Recipient struct {
	gorm.Model
	Name  string `gorm:"uniqueIndex;not null;size:100;comment:转发对象姓名" json:"name"`
	Email string `gorm:"not null;size:255;comment:转发对象邮箱" json:"email"`
}