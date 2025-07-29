package scheduler

import (
	"fmt"
	"log"
	"os"
	"time"

	"gmail-forwarding/internal/gmail"
	"gmail-forwarding/internal/processor"

	"github.com/robfig/cron/v3"
)

// Scheduler 定时任务调度器
type Scheduler struct {
	cron           *cron.Cron
	emailProcessor *processor.EmailProcessor
}

// NewScheduler 创建新的调度器
func NewScheduler() *Scheduler {
	// 获取Gmail配置
	gmailUser := os.Getenv("GMAIL_USER")
	gmailPassword := os.Getenv("GMAIL_APP_PASSWORD")

	if gmailUser == "" || gmailPassword == "" {
		log.Fatal("Gmail配置不完整，请检查GMAIL_USER和GMAIL_APP_PASSWORD环境变量")
	}

	// 创建客户端
	imapClient := gmail.NewIMAPClient(gmailUser, gmailPassword)
	smtpClient := gmail.NewSMTPClient(gmailUser, gmailPassword)

	// 创建处理器
	emailProcessor := processor.NewEmailProcessor(imapClient, smtpClient)

	// 创建cron实例，支持秒级调度
	c := cron.New(cron.WithSeconds())

	return &Scheduler{
		cron:           c,
		emailProcessor: emailProcessor,
	}
}

// Start 启动定时任务
func (s *Scheduler) Start() {
	// 获取检查间隔配置，默认5分钟
	interval := os.Getenv("CHECK_INTERVAL")
	if interval == "" {
		interval = "5m"
	}

	// 解析间隔时间
	duration, err := time.ParseDuration(interval)
	if err != nil {
		log.Printf("无效的检查间隔配置 %s，使用默认值5分钟", interval)
		duration = 5 * time.Minute
	}

	// 构建cron表达式（每N分钟执行一次）
	cronExpr := buildCronExpression(duration)
	log.Printf("设置邮件检查间隔: %s (cron: %s)", duration, cronExpr)

	// 添加定时任务
	_, err = s.cron.AddFunc(cronExpr, func() {
		log.Println("定时任务触发，开始处理邮件...")
		if err := s.emailProcessor.ProcessEmails(); err != nil {
			log.Printf("定时处理邮件失败: %v", err)
		}
	})

	if err != nil {
		log.Fatalf("添加定时任务失败: %v", err)
	}

	// 启动调度器
	s.cron.Start()
	log.Println("邮件处理定时任务已启动")

	// 立即执行一次
	go func() {
		log.Println("启动时执行一次邮件处理...")
		if err := s.emailProcessor.ProcessEmails(); err != nil {
			log.Printf("启动时处理邮件失败: %v", err)
		}
	}()
}

// Stop 停止定时任务
func (s *Scheduler) Stop() {
	if s.cron != nil {
		s.cron.Stop()
		log.Println("定时任务已停止")
	}
}

// buildCronExpression 根据时间间隔构建cron表达式
func buildCronExpression(duration time.Duration) string {
	minutes := int(duration.Minutes())
	
	if minutes < 1 {
		// 小于1分钟的，按秒处理
		seconds := int(duration.Seconds())
		return formatSeconds(seconds)
	}
	
	if minutes == 1 {
		return "0 * * * * *" // 每分钟执行
	}
	
	if minutes < 60 {
		return formatMinutes(minutes)
	}
	
	// 大于等于60分钟的，按小时处理
	hours := minutes / 60
	return formatHours(hours)
}

// formatSeconds 格式化秒级cron表达式
func formatSeconds(seconds int) string {
	if seconds < 60 {
		return "0 0/" + fmt.Sprintf("%d", seconds) + " * * * *"
	}
	return "0 * * * * *" // 默认每分钟
}

// formatMinutes 格式化分钟级cron表达式  
func formatMinutes(minutes int) string {
	return "0 0/" + fmt.Sprintf("%d", minutes) + " * * * *"
}

// formatHours 格式化小时级cron表达式
func formatHours(hours int) string {
	return "0 0 0/" + fmt.Sprintf("%d", hours) + " * * *"
}