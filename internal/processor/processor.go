package processor

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"

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
	Keyword string
	Email   string
}

// parseSubject 解析邮件主题，提取关键字和邮箱地址
func (ep *EmailProcessor) parseSubject(subject string) (*SubjectParseResult, error) {
	// 使用正则表达式解析主题格式：关键字 - 邮箱地址
	re := regexp.MustCompile(`^(.+?)\s*-\s*(.+?)$`)
	matches := re.FindStringSubmatch(strings.TrimSpace(subject))

	if len(matches) != 3 {
		return nil, fmt.Errorf("邮件主题格式不正确，应为：关键字 - 邮箱地址")
	}

	keyword := strings.TrimSpace(matches[1])
	email := strings.TrimSpace(matches[2])

	// 验证邮箱格式
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return nil, fmt.Errorf("邮箱地址格式不正确: %s", email)
	}

	return &SubjectParseResult{
		Keyword: keyword,
		Email:   email,
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

// findOrCreateRecipient 根据邮箱地址查找或创建转发对象
func (ep *EmailProcessor) findOrCreateRecipient(email string) (*models.Recipient, error) {
	db := database.GetDB()
	var recipient models.Recipient

	// 首先尝试根据邮箱地址查找现有记录
	err := db.Where("email = ?", email).First(&recipient).Error
	if err == nil {
		// 找到现有记录
		return &recipient, nil
	}

	// 如果不存在，创建新记录
	// 使用邮箱地址的用户名部分作为名称
	atIndex := strings.Index(email, "@")
	name := email
	if atIndex > 0 {
		name = email[:atIndex]
	}

	recipient = models.Recipient{
		Name:  name,
		Email: email,
	}

	if err := db.Create(&recipient).Error; err != nil {
		return nil, fmt.Errorf("创建转发对象失败: %w", err)
	}

	log.Printf("创建新的转发对象: %s <%s>", recipient.Name, recipient.Email)
	return &recipient, nil
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

	log.Printf("解析结果 - 关键字: %s, 转发邮箱: %s", parseResult.Keyword, parseResult.Email)

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

	// 查找或创建转发对象
	recipient, err := ep.findOrCreateRecipient(parseResult.Email)
	if err != nil {
		log.Printf("查找或创建转发对象失败: %v", err)
		return nil
	}

	log.Printf("找到转发对象: %s <%s>", recipient.Name, recipient.Email)

	// 转发邮件
	err = ep.smtpClient.ForwardEmail(email, recipient.Email)
	if err != nil {
		log.Printf("转发邮件失败: %v", err)
		return err
	}

	log.Printf("邮件成功转发给: %s", recipient.Email)

	return nil
}