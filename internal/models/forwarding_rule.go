package models

import (
	"gorm.io/gorm"
)

// ForwardingRule 转发规则表
type ForwardingRule struct {
	gorm.Model
	Keyword string `gorm:"uniqueIndex;not null;size:100;comment:匹配关键字" json:"keyword"`
	Active  bool   `gorm:"default:true;comment:是否启用" json:"active"`
}