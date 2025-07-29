package processor

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"gmail-forwarding/internal/database"
	"gmail-forwarding/internal/gmail"
	"gmail-forwarding/internal/models"
)

// EmailProcessor 邮件处理器
type EmailProcessor struct {
	imapClient *gmail.IMAPClient
	smtpClient *gmail.SMTPClient
	mu         sync.Mutex
}

// NewEmailProcessor 创建新的邮件处理器
func NewEmailProcessor(imapClient *gmail.IMAPClient, smtpClient *gmail.SMTPClient) *EmailProcessor {
	return &EmailProcessor{
		imapClient: imapClient,
		smtpClient: smtpClient,
	}
}

// SubjectParseResult 主题解析结果
type SubjectParseResult struct {
	Keyword   string
	Recipient string
}

// parseSubject 解析邮件主题，提取关键字和转发对象
func (ep *EmailProcessor) parseSubject(subject string) (*SubjectParseResult, error) {
	// 使用正则表达式解析主题格式：关键字 - 转发对象
	// 支持中英文字符、数字、空格等
	re := regexp.MustCompile(`^(.+?)\s*-\s*(.+?)$`)
	matches := re.FindStringSubmatch(strings.TrimSpace(subject))

	if len(matches) != 3 {
		return nil, fmt.Errorf("邮件主题格式不正确，应为：关键字 - 转发对象")
	}

	return &SubjectParseResult{
		Keyword:   strings.TrimSpace(matches[1]),
		Recipient: strings.TrimSpace(matches[2]),
	}, nil
}

// checkForwardingRule 检查转发规则是否存在且启用
func (ep *EmailProcessor) checkForwardingRule(keyword string) (bool, error) {
	db := database.GetDB()
	var rule models.ForwardingRule

	err := db.Where("keyword = ? AND active = ?", keyword, true).First(&rule).Error
	if err != nil {
		return false, err
	}

	return true, nil
}

// findRecipient 根据姓名查找转发对象
func (ep *EmailProcessor) findRecipient(name string) (*models.Recipient, error) {
	db := database.GetDB()
	var recipient models.Recipient

	err := db.Where("name = ?", name).First(&recipient).Error
	if err != nil {
		return nil, err
	}

	return &recipient, nil
}

// logEmailAction 记录邮件处理日志
func (ep *EmailProcessor) logEmailAction(email *gmail.Email, recipient *models.Recipient, status, errorMsg string) error {
	db := database.GetDB()

	now := time.Now()
	emailLog := &models.EmailLog{
		MessageID:   email.MessageID,
		Subject:     email.Subject,
		FromEmail:   email.From,
		RecipientID: recipient.ID,
		Status:      status,
		Error:       errorMsg,
		ForwardedAt: &now,
	}

	return db.Create(emailLog).Error
}

// ProcessEmails 处理邮件主函数
func (ep *EmailProcessor) ProcessEmails() error {
	ep.mu.Lock()
	defer ep.mu.Unlock()

	log.Println("开始处理邮件...")

	// 连接 IMAP 服务器
	if err := ep.imapClient.Connect(); err != nil {
		return fmt.Errorf("连接IMAP服务器失败: %w", err)
	}
	defer ep.imapClient.Disconnect()

	// 获取未读邮件
	emails, err := ep.imapClient.FetchUnreadEmails()
	if err != nil {
		return fmt.Errorf("获取邮件失败: %w", err)
	}

	if len(emails) == 0 {
		log.Println("没有未读邮件")
		return nil
	}

	log.Printf("找到 %d 封未读邮件", len(emails))

	// 处理每封邮件
	for _, email := range emails {
		if err := ep.processEmail(email); err != nil {
			log.Printf("处理邮件失败 [%s]: %v", email.Subject, err)
		}

		// 标记邮件为已读
		if err := ep.imapClient.MarkAsRead(email.MessageID); err != nil {
			log.Printf("标记邮件已读失败 [%s]: %v", email.MessageID, err)
		}
	}

	log.Println("邮件处理完成")
	return nil
}

// processEmail 处理单封邮件
func (ep *EmailProcessor) processEmail(email *gmail.Email) error {
	log.Printf("处理邮件: %s", email.Subject)

	// 解析邮件主题
	parseResult, err := ep.parseSubject(email.Subject)
	if err != nil {
		log.Printf("邮件主题解析失败: %v", err)
		return nil // 不是转发格式的邮件，跳过
	}

	log.Printf("解析结果 - 关键字: %s, 转发对象: %s", parseResult.Keyword, parseResult.Recipient)

	// 检查转发规则
	ruleExists, err := ep.checkForwardingRule(parseResult.Keyword)
	if err != nil {
		log.Printf("检查转发规则失败: %v", err)
		return nil // 规则不存在，跳过
	}

	if !ruleExists {
		log.Printf("关键字 '%s' 没有对应的转发规则", parseResult.Keyword)
		return nil
	}

	// 查找转发对象
	recipient, err := ep.findRecipient(parseResult.Recipient)
	if err != nil {
		log.Printf("查找转发对象失败: %v", err)
		// 记录失败日志
		dummyRecipient := &models.Recipient{Model: models.Recipient{}.Model, Name: parseResult.Recipient}
		ep.logEmailAction(email, dummyRecipient, "failed", fmt.Sprintf("转发对象不存在: %v", err))
		return nil
	}

	log.Printf("找到转发对象: %s <%s>", recipient.Name, recipient.Email)

	// 转发邮件
	err = ep.smtpClient.ForwardEmail(email, recipient.Email)
	if err != nil {
		log.Printf("转发邮件失败: %v", err)
		ep.logEmailAction(email, recipient, "failed", err.Error())
		return err
	}

	// 记录成功日志
	ep.logEmailAction(email, recipient, "success", "")
	log.Printf("邮件成功转发给: %s", recipient.Email)

	return nil
}