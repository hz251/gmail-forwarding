package gmail

import (
	"fmt"
	"log"
	"net/smtp"
	"strings"
	"time"
)

// SMTPClient SMTP 客户端
type SMTPClient struct {
	host        string
	port        string
	username    string
	password    string
	maxRetries  int
	retryDelay time.Duration
	timeout     time.Duration
}

// NewSMTPClient 创建新的 SMTP 客户端
func NewSMTPClient(username, password string) *SMTPClient {
	return &SMTPClient{
		host:        "smtp.gmail.com",
		port:        "587",
		username:    username,
		password:    password,
		maxRetries:  3,
		retryDelay:  2 * time.Second,
		timeout:     30 * time.Second,
	}
}

// ForwardEmail 转发邮件 - 使用改进的SMTP实现和重试机制
func (sc *SMTPClient) ForwardEmail(email *Email, toEmail string) error {
	log.Printf("开始发送邮件到: %s", toEmail)
	
	// 构建邮件内容
	message := sc.buildForwardMessage(email, toEmail)
	
	// 使用重试机制发送邮件
	var lastErr error
	for attempt := 1; attempt <= sc.maxRetries; attempt++ {
		log.Printf("尝试发送邮件 - 第 %d/%d 次", attempt, sc.maxRetries)
		
		err := sc.sendEmailWithManualSMTP(toEmail, message)
		if err == nil {
			log.Printf("邮件成功发送到: %s (第%d次尝试)", toEmail, attempt)
			return nil
		}
		
		lastErr = err
		log.Printf("第%d次尝试失败: %v", attempt, err)
		
		// 如果不是最后一次尝试，等待一段时间再重试
		if attempt < sc.maxRetries {
			log.Printf("等待 %v 后重试...", sc.retryDelay)
			time.Sleep(sc.retryDelay)
		}
	}
	
	return fmt.Errorf("发送邮件失败，已经进行%d次尝试: %w", sc.maxRetries, lastErr)
}

// sendEmailWithManualSMTP 直接使用smtp.SendMail，简化实现
func (sc *SMTPClient) sendEmailWithManualSMTP(toEmail, message string) error {
	addr := fmt.Sprintf("%s:%s", sc.host, sc.port)
	log.Printf("使用smtp.SendMail发送到: %s", addr)
	
	// 使用简化的认证和发送
	auth := smtp.PlainAuth("", sc.username, sc.password, sc.host)
	
	err := smtp.SendMail(addr, auth, sc.username, []string{toEmail}, []byte(message))
	if err != nil {
		log.Printf("smtp.SendMail失败: %v", err)
		return fmt.Errorf("smtp.SendMail失败: %w", err)
	}
	
	log.Printf("邮件成功发送")
	return nil
}

// buildForwardMessage 构建转发邮件内容
func (sc *SMTPClient) buildForwardMessage(email *Email, toEmail string) string {
	var message strings.Builder

	// 邮件头
	message.WriteString(fmt.Sprintf("To: %s\r\n", toEmail))
	message.WriteString(fmt.Sprintf("From: %s\r\n", sc.username))
	message.WriteString(fmt.Sprintf("Subject: [转发] %s\r\n", email.Subject))
	message.WriteString("MIME-Version: 1.0\r\n")
	message.WriteString("Content-Type: multipart/alternative; boundary=\"boundary123\"\r\n")
	message.WriteString("\r\n")

	// 添加转发说明
	message.WriteString("--boundary123\r\n")
	message.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	message.WriteString("\r\n")
	message.WriteString("---------- 转发邮件 ----------\r\n")
	message.WriteString(fmt.Sprintf("发件人: %s\r\n", email.From))
	message.WriteString(fmt.Sprintf("主题: %s\r\n", email.Subject))
	message.WriteString(fmt.Sprintf("收件人: %s\r\n", email.To))
	message.WriteString("---------- 邮件内容 ----------\r\n\r\n")

	// 添加原邮件正文（纯文本）
	if email.Body != "" {
		message.WriteString(email.Body)
	} else {
		message.WriteString("（此邮件无纯文本内容）")
	}
	message.WriteString("\r\n")

	// 如果有 HTML 内容，也添加进去
	if email.HTML != "" {
		message.WriteString("\r\n--boundary123\r\n")
		message.WriteString("Content-Type: text/html; charset=utf-8\r\n")
		message.WriteString("\r\n")
		message.WriteString("<div style=\"border-left: 3px solid #ccc; padding-left: 10px; margin: 10px 0;\">")
		message.WriteString("<h4>---------- 转发邮件 ----------</h4>")
		message.WriteString(fmt.Sprintf("<p><strong>发件人:</strong> %s</p>", email.From))
		message.WriteString(fmt.Sprintf("<p><strong>主题:</strong> %s</p>", email.Subject))
		message.WriteString(fmt.Sprintf("<p><strong>收件人:</strong> %s</p>", email.To))
		message.WriteString("<h4>---------- 邮件内容 ----------</h4>")
		message.WriteString(email.HTML)
		message.WriteString("</div>")
	}

	message.WriteString("\r\n--boundary123--\r\n")

	return message.String()
}

