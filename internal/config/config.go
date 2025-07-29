package config

import (
	"log"
	"os"
)

// Config 应用配置结构
type Config struct {
	// 数据库配置
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// Gmail 配置
	GmailUser     string
	GmailPassword string

	// 应用配置
	AppPort       string
	CheckInterval string
}

// GlobalConfig 全局配置实例
var GlobalConfig *Config

// Load 加载配置
func Load() {
	GlobalConfig = &Config{
		// 数据库配置
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "gmail_forwarding"),

		// Gmail 配置
		GmailUser:     getEnv("GMAIL_USER", ""),
		GmailPassword: getEnv("GMAIL_APP_PASSWORD", ""),

		// 应用配置
		AppPort:       getEnv("APP_PORT", "8080"),
		CheckInterval: getEnv("CHECK_INTERVAL", "5m"),
	}

	// 验证必需的配置
	validateConfig()

	log.Println("配置加载完成")
}

// getEnv 获取环境变量，如果不存在则使用默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// validateConfig 验证配置
func validateConfig() {
	if GlobalConfig.GmailUser == "" {
		log.Fatal("GMAIL_USER 环境变量未设置")
	}

	if GlobalConfig.GmailPassword == "" {
		log.Fatal("GMAIL_APP_PASSWORD 环境变量未设置")
	}

	if GlobalConfig.DBPassword == "" {
		log.Println("警告: DB_PASSWORD 环境变量未设置，可能导致数据库连接失败")
	}

	log.Printf("Gmail 账户: %s", GlobalConfig.GmailUser)
	log.Printf("数据库: %s:%s/%s", GlobalConfig.DBHost, GlobalConfig.DBPort, GlobalConfig.DBName)
	log.Printf("应用端口: %s", GlobalConfig.AppPort)
	log.Printf("检查间隔: %s", GlobalConfig.CheckInterval)
}