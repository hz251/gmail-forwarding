package handlers

import (
	"net/http"
	"strconv"

	"gmail-forwarding/internal/database"
	"gmail-forwarding/internal/models"

	"github.com/gin-gonic/gin"
)

// RecipientResponse 转发对象响应结构
type RecipientResponse struct {
	Success bool                `json:"success"`
	Message string              `json:"message"`
	Data    interface{}         `json:"data,omitempty"`
	Error   string              `json:"error,omitempty"`
}

// GetRecipients 获取所有转发对象
func GetRecipients(c *gin.Context) {
	db := database.GetDB()
	var recipients []models.Recipient

	if err := db.Find(&recipients).Error; err != nil {
		c.JSON(http.StatusInternalServerError, RecipientResponse{
			Success: false,
			Message: "获取转发对象列表失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, RecipientResponse{
		Success: true,
		Message: "获取转发对象列表成功",
		Data:    recipients,
	})
}

// GetRecipient 获取单个转发对象
func GetRecipient(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, RecipientResponse{
			Success: false,
			Message: "无效的ID参数",
			Error:   err.Error(),
		})
		return
	}

	db := database.GetDB()
	var recipient models.Recipient

	if err := db.First(&recipient, id).Error; err != nil {
		c.JSON(http.StatusNotFound, RecipientResponse{
			Success: false,
			Message: "转发对象不存在",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, RecipientResponse{
		Success: true,
		Message: "获取转发对象成功",
		Data:    recipient,
	})
}

// CreateRecipient 创建转发对象
func CreateRecipient(c *gin.Context) {
	var recipient models.Recipient

	if err := c.ShouldBindJSON(&recipient); err != nil {
		c.JSON(http.StatusBadRequest, RecipientResponse{
			Success: false,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	// 验证必填字段
	if recipient.Name == "" || recipient.Email == "" {
		c.JSON(http.StatusBadRequest, RecipientResponse{
			Success: false,
			Message: "姓名和邮箱地址不能为空",
		})
		return
	}

	db := database.GetDB()
	if err := db.Create(&recipient).Error; err != nil {
		c.JSON(http.StatusInternalServerError, RecipientResponse{
			Success: false,
			Message: "创建转发对象失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, RecipientResponse{
		Success: true,
		Message: "创建转发对象成功",
		Data:    recipient,
	})
}

// UpdateRecipient 更新转发对象
func UpdateRecipient(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, RecipientResponse{
			Success: false,
			Message: "无效的ID参数",
			Error:   err.Error(),
		})
		return
	}

	db := database.GetDB()
	var recipient models.Recipient

	// 检查记录是否存在
	if err := db.First(&recipient, id).Error; err != nil {
		c.JSON(http.StatusNotFound, RecipientResponse{
			Success: false,
			Message: "转发对象不存在",
			Error:   err.Error(),
		})
		return
	}

	// 绑定更新数据
	var updateData models.Recipient
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, RecipientResponse{
			Success: false,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	// 验证必填字段
	if updateData.Name == "" || updateData.Email == "" {
		c.JSON(http.StatusBadRequest, RecipientResponse{
			Success: false,
			Message: "姓名和邮箱地址不能为空",
		})
		return
	}

	// 更新数据
	recipient.Name = updateData.Name
	recipient.Email = updateData.Email

	if err := db.Save(&recipient).Error; err != nil {
		c.JSON(http.StatusInternalServerError, RecipientResponse{
			Success: false,
			Message: "更新转发对象失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, RecipientResponse{
		Success: true,
		Message: "更新转发对象成功",
		Data:    recipient,
	})
}

// DeleteRecipient 删除转发对象
func DeleteRecipient(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, RecipientResponse{
			Success: false,
			Message: "无效的ID参数",
			Error:   err.Error(),
		})
		return
	}

	db := database.GetDB()
	var recipient models.Recipient

	// 检查记录是否存在
	if err := db.First(&recipient, id).Error; err != nil {
		c.JSON(http.StatusNotFound, RecipientResponse{
			Success: false,
			Message: "转发对象不存在",
			Error:   err.Error(),
		})
		return
	}

	// 删除记录
	if err := db.Delete(&recipient).Error; err != nil {
		c.JSON(http.StatusInternalServerError, RecipientResponse{
			Success: false,
			Message: "删除转发对象失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, RecipientResponse{
		Success: true,
		Message: "删除转发对象成功",
	})
}