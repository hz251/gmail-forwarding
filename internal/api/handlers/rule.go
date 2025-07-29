package handlers

import (
	"net/http"
	"strconv"

	"gmail-forwarding/internal/database"
	"gmail-forwarding/internal/models"

	"github.com/gin-gonic/gin"
)

// RuleResponse 转发规则响应结构
type RuleResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// GetRules 获取所有转发规则
func GetRules(c *gin.Context) {
	db := database.GetDB()
	var rules []models.ForwardingRule

	if err := db.Find(&rules).Error; err != nil {
		c.JSON(http.StatusInternalServerError, RuleResponse{
			Success: false,
			Message: "获取转发规则列表失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, RuleResponse{
		Success: true,
		Message: "获取转发规则列表成功",
		Data:    rules,
	})
}

// GetRule 获取单个转发规则
func GetRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, RuleResponse{
			Success: false,
			Message: "无效的ID参数",
			Error:   err.Error(),
		})
		return
	}

	db := database.GetDB()
	var rule models.ForwardingRule

	if err := db.First(&rule, id).Error; err != nil {
		c.JSON(http.StatusNotFound, RuleResponse{
			Success: false,
			Message: "转发规则不存在",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, RuleResponse{
		Success: true,
		Message: "获取转发规则成功",
		Data:    rule,
	})
}

// CreateRule 创建转发规则
func CreateRule(c *gin.Context) {
	var rule models.ForwardingRule

	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, RuleResponse{
			Success: false,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	// 验证必填字段
	if rule.Keyword == "" {
		c.JSON(http.StatusBadRequest, RuleResponse{
			Success: false,
			Message: "关键字不能为空",
		})
		return
	}

	// 设置默认值
	if !c.Request.URL.Query().Has("active") {
		rule.Active = true
	}

	db := database.GetDB()
	if err := db.Create(&rule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, RuleResponse{
			Success: false,
			Message: "创建转发规则失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, RuleResponse{
		Success: true,
		Message: "创建转发规则成功",
		Data:    rule,
	})
}

// UpdateRule 更新转发规则
func UpdateRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, RuleResponse{
			Success: false,
			Message: "无效的ID参数",
			Error:   err.Error(),
		})
		return
	}

	db := database.GetDB()
	var rule models.ForwardingRule

	// 检查记录是否存在
	if err := db.First(&rule, id).Error; err != nil {
		c.JSON(http.StatusNotFound, RuleResponse{
			Success: false,
			Message: "转发规则不存在",
			Error:   err.Error(),
		})
		return
	}

	// 绑定更新数据
	var updateData models.ForwardingRule
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, RuleResponse{
			Success: false,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	// 验证必填字段
	if updateData.Keyword == "" {
		c.JSON(http.StatusBadRequest, RuleResponse{
			Success: false,
			Message: "关键字不能为空",
		})
		return
	}

	// 更新数据
	rule.Keyword = updateData.Keyword
	rule.Active = updateData.Active

	if err := db.Save(&rule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, RuleResponse{
			Success: false,
			Message: "更新转发规则失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, RuleResponse{
		Success: true,
		Message: "更新转发规则成功",
		Data:    rule,
	})
}

// DeleteRule 删除转发规则
func DeleteRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, RuleResponse{
			Success: false,
			Message: "无效的ID参数",
			Error:   err.Error(),
		})
		return
	}

	db := database.GetDB()
	var rule models.ForwardingRule

	// 检查记录是否存在
	if err := db.First(&rule, id).Error; err != nil {
		c.JSON(http.StatusNotFound, RuleResponse{
			Success: false,
			Message: "转发规则不存在",
			Error:   err.Error(),
		})
		return
	}

	// 删除记录
	if err := db.Delete(&rule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, RuleResponse{
			Success: false,
			Message: "删除转发规则失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, RuleResponse{
		Success: true,
		Message: "删除转发规则成功",
	})
}