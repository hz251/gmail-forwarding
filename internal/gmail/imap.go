package gmail

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
)

// Email 邮件结构体
type Email struct {
	MessageID string
	Subject   string
	From      string
	To        string
	Body      string
	HTML      string
}

// IMAPClient IMAP 客户端
type IMAPClient struct {
	client   *client.Client
	username string
	password string
}

// NewIMAPClient 创建新的 IMAP 客户端
func NewIMAPClient(username, password string) *IMAPClient {
	return &IMAPClient{
		username: username,
		password: password,
	}
}

// Connect 连接到 Gmail IMAP 服务器
func (ic *IMAPClient) Connect() error {
	// 连接到 Gmail IMAP 服务器
	c, err := client.DialTLS("imap.gmail.com:993", nil)
	if err != nil {
		return fmt.Errorf("failed to connect to IMAP server: %w", err)
	}
	ic.client = c

	// 登录
	if err := c.Login(ic.username, ic.password); err != nil {
		return fmt.Errorf("failed to login: %w", err)
	}

	log.Println("Successfully connected to Gmail IMAP server")
	return nil
}

// FetchUnreadEmails 获取未读邮件
func (ic *IMAPClient) FetchUnreadEmails() ([]*Email, error) {
	// 选择收件箱
	mbox, err := ic.client.Select("INBOX", false)
	if err != nil {
		return nil, fmt.Errorf("failed to select INBOX: %w", err)
	}
	log.Printf("Mailbox contains %d messages", mbox.Messages)

	// 搜索未读邮件
	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{imap.SeenFlag}
	ids, err := ic.client.Search(criteria)
	if err != nil {
		return nil, fmt.Errorf("failed to search emails: %w", err)
	}

	if len(ids) == 0 {
		log.Println("No unread emails found")
		return nil, nil
	}

	log.Printf("Found %d unread emails", len(ids))

	// 创建序列集
	seqset := new(imap.SeqSet)
	seqset.AddNum(ids...)

	// 获取邮件内容
	messages := make(chan *imap.Message, len(ids))
	done := make(chan error, 1)

	go func() {
		done <- ic.client.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope, imap.FetchRFC822}, messages)
	}()

	emails := []*Email{}
	for msg := range messages {
		email, err := ic.parseMessage(msg)
		if err != nil {
			log.Printf("Failed to parse message: %v", err)
			continue
		}
		emails = append(emails, email)
	}

	if err := <-done; err != nil {
		return nil, fmt.Errorf("failed to fetch messages: %w", err)
	}

	return emails, nil
}

// parseMessage 解析邮件消息
func (ic *IMAPClient) parseMessage(msg *imap.Message) (*Email, error) {
	email := &Email{}

	// 获取邮件头信息
	if msg.Envelope != nil {
		email.MessageID = msg.Envelope.MessageId
		email.Subject = msg.Envelope.Subject
		if len(msg.Envelope.From) > 0 {
			email.From = fmt.Sprintf("%s <%s>", msg.Envelope.From[0].PersonalName, msg.Envelope.From[0].Address())
		}
		if len(msg.Envelope.To) > 0 {
			email.To = fmt.Sprintf("%s <%s>", msg.Envelope.To[0].PersonalName, msg.Envelope.To[0].Address())
		}
	}

	// 获取邮件正文
	for _, value := range msg.Body {
		mr, err := mail.CreateReader(value)
		if err != nil {
			continue
		}

		// 读取邮件各部分
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				continue
			}

			switch h := p.Header.(type) {
			case *mail.InlineHeader:
				// 读取内联内容
				b, _ := io.ReadAll(p.Body)
				contentType, _, _ := h.ContentType()
				if strings.HasPrefix(contentType, "text/plain") {
					email.Body = string(b)
				} else if strings.HasPrefix(contentType, "text/html") {
					email.HTML = string(b)
				}
			}
		}
	}

	return email, nil
}

// MarkAsRead 标记邮件为已读
func (ic *IMAPClient) MarkAsRead(messageID string) error {
	// 搜索特定消息ID的邮件
	criteria := imap.NewSearchCriteria()
	criteria.Header.Add("Message-ID", messageID)
	ids, err := ic.client.Search(criteria)
	if err != nil {
		return fmt.Errorf("failed to search email by message ID: %w", err)
	}

	if len(ids) == 0 {
		return fmt.Errorf("email not found with message ID: %s", messageID)
	}

	// 创建序列集
	seqset := new(imap.SeqSet)
	seqset.AddNum(ids...)

	// 添加已读标记
	item := imap.FormatFlagsOp(imap.AddFlags, true)
	flags := []interface{}{imap.SeenFlag}
	err = ic.client.Store(seqset, item, flags, nil)
	if err != nil {
		return fmt.Errorf("failed to mark email as read: %w", err)
	}

	log.Printf("Marked email %s as read", messageID)
	return nil
}

// Disconnect 断开连接
func (ic *IMAPClient) Disconnect() error {
	if ic.client != nil {
		return ic.client.Logout()
	}
	return nil
}