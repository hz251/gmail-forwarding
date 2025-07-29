package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"gmail-forwarding/internal/api"
	"gmail-forwarding/internal/config"
	"gmail-forwarding/internal/database"
	"gmail-forwarding/internal/scheduler"
)

func main() {
	log.Println("启动 Gmail 邮件转发服务...")

	// 1. 加载配置
	config.Load()

	// 2. 初始化数据库
	if err := database.InitDB(); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}

	// 3. 启动定时任务
	emailScheduler := scheduler.NewScheduler()
	emailScheduler.Start()

	// 4. 设置路由并启动HTTP服务器
	router := api.SetupRoutes()
	
	// 启动服务器的goroutine
	go func() {
		port := config.GlobalConfig.AppPort
		log.Printf("HTTP 服务器启动在端口: %s", port)
		if err := router.Run(":" + port); err != nil {
			log.Fatalf("启动HTTP服务器失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("收到退出信号，正在关闭服务...")

	// 停止定时任务
	emailScheduler.Stop()

	log.Println("服务已关闭")
}