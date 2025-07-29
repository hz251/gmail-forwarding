package handlers

import (
	"net/http"
	"os"

	"gmail-forwarding/internal/gmail"
	"gmail-forwarding/internal/processor"

	"github.com/gin-gonic/gin"
)

// ProcessResponse 处理响应结构
type ProcessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// ProcessEmails 手动触发邮件处理
func ProcessEmails(c *gin.Context) {
	// 获取Gmail配置
	gmailUser := os.Getenv("GMAIL_USER")
	gmailPassword := os.Getenv("GMAIL_APP_PASSWORD")

	if gmailUser == "" || gmailPassword == "" {
		c.JSON(http.StatusInternalServerError, ProcessResponse{
			Success: false,
			Message: "Gmail配置不完整",
			Error:   "请检查GMAIL_USER和GMAIL_APP_PASSWORD环境变量",
		})
		return
	}

	// 创建客户端
	imapClient := gmail.NewIMAPClient(gmailUser, gmailPassword)
	smtpClient := gmail.NewSMTPClient(gmailUser, gmailPassword)

	// 创建处理器
	emailProcessor := processor.NewEmailProcessor(imapClient, smtpClient)

	// 处理邮件
	if err := emailProcessor.ProcessEmails(); err != nil {
		c.JSON(http.StatusInternalServerError, ProcessResponse{
			Success: false,
			Message: "邮件处理失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ProcessResponse{
		Success: true,
		Message: "邮件处理完成",
	})
}