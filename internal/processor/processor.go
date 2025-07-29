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

// loadActiveRules 预加载所有启用的转发规则
func (ep *EmailProcessor) loadActiveRules() (map[string]bool, error) {
	db := database.GetDB()
	var rules []models.ForwardingRule

	err := db.Where("active = ?", true).Find(&rules).Error
	if err != nil {
		return nil, fmt.Errorf("加载转发规则失败: %w", err)
	}

	rulesMap := make(map[string]bool)
	for _, rule := range rules {
		rulesMap[rule.Keyword] = true
	}

	log.Printf("已加载 %d 个启用的转发规则", len(rulesMap))
	return rulesMap, nil
}

// shouldForward 检查邮件是否应该转发
func (ep *EmailProcessor) shouldForward(email *gmail.Email, activeRules map[string]bool) (*SubjectParseResult, bool) {
	// 解析邮件主题
	parseResult, err := ep.parseSubject(email.Subject)
	if err != nil {
		log.Printf("邮件主题解析失败: %v", err)
		return nil, false // 不是转发格式的邮件，跳过
	}

	// 内存中快速匹配关键字
	if !activeRules[parseResult.Keyword] {
		log.Printf("关键字 '%s' 没有对应的转发规则", parseResult.Keyword)
		return parseResult, false
	}

	log.Printf("匹配到转发规则 - 关键字: %s, 转发邮箱: %s", parseResult.Keyword, parseResult.Email)
	return parseResult, true
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

	// 预加载所有启用的转发规则
	activeRules, err := ep.loadActiveRules()
	if err != nil {
		return fmt.Errorf("加载转发规则失败: %w", err)
	}

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
		if err := ep.processEmailWithRules(email, activeRules); err != nil {
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

// processEmail 处理单封邮件（旧方法，保留兼容性）
func (ep *EmailProcessor) processEmail(email *gmail.Email) error {
	// 加载规则并调用新方法
	activeRules, err := ep.loadActiveRules()
	if err != nil {
		return fmt.Errorf("加载转发规则失败: %w", err)
	}
	return ep.processEmailWithRules(email, activeRules)
}

// processEmailWithRules 使用预加载规则处理单封邮件
func (ep *EmailProcessor) processEmailWithRules(email *gmail.Email, activeRules map[string]bool) error {
	log.Printf("处理邮件: %s", email.Subject)

	// 检查邮件是否应该转发
	parseResult, shouldForward := ep.shouldForward(email, activeRules)
	if !shouldForward {
		return nil // 不需要转发，跳过
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