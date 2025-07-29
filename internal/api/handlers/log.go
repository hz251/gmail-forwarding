package handlers

import (
	"net/http"
	"strconv"

	"gmail-forwarding/internal/database"
	"gmail-forwarding/internal/models"

	"github.com/gin-gonic/gin"
)

// LogResponse 邮件日志响应结构
type LogResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// GetLogs 获取邮件日志列表
func GetLogs(c *gin.Context) {
	db := database.GetDB()
	var logs []models.EmailLog

	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")

	// 计算偏移量
	offset := (page - 1) * pageSize

	// 构建查询
	query := db.Preload("Recipient").Order("created_at DESC")

	// 状态过滤
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	var total int64
	query.Model(&models.EmailLog{}).Count(&total)

	// 分页查询
	if err := query.Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, LogResponse{
			Success: false,
			Message: "获取邮件日志失败",
			Error:   err.Error(),
		})
		return
	}

	// 构建响应数据
	data := map[string]interface{}{
		"logs":       logs,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
		"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
	}

	c.JSON(http.StatusOK, LogResponse{
		Success: true,
		Message: "获取邮件日志成功",
		Data:    data,
	})
}

// GetLog 获取单个邮件日志
func GetLog(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, LogResponse{
			Success: false,
			Message: "无效的ID参数",
			Error:   err.Error(),
		})
		return
	}

	db := database.GetDB()
	var log models.EmailLog

	if err := db.Preload("Recipient").First(&log, id).Error; err != nil {
		c.JSON(http.StatusNotFound, LogResponse{
			Success: false,
			Message: "邮件日志不存在",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, LogResponse{
		Success: true,
		Message: "获取邮件日志成功",
		Data:    log,
	})
}

// GetLogStats 获取邮件日志统计信息
func GetLogStats(c *gin.Context) {
	db := database.GetDB()

	// 统计总数
	var total int64
	db.Model(&models.EmailLog{}).Count(&total)

	// 统计成功数
	var success int64
	db.Model(&models.EmailLog{}).Where("status = ?", "success").Count(&success)

	// 统计失败数
	var failed int64
	db.Model(&models.EmailLog{}).Where("status = ?", "failed").Count(&failed)

	// 最近24小时统计
	var recent24h int64
	db.Raw("SELECT COUNT(*) FROM email_logs WHERE created_at > DATE_SUB(NOW(), INTERVAL 24 HOUR)").Scan(&recent24h)

	stats := map[string]interface{}{
		"total":      total,
		"success":    success,
		"failed":     failed,
		"recent_24h": recent24h,
		"success_rate": func() float64 {
			if total > 0 {
				return float64(success) / float64(total) * 100
			}
			return 0
		}(),
	}

	c.JSON(http.StatusOK, LogResponse{
		Success: true,
		Message: "获取邮件统计信息成功",
		Data:    stats,
	})
}