package api

import (
	"gmail-forwarding/internal/api/handlers"
	"gmail-forwarding/internal/api/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置路由
func SetupRoutes() *gin.Engine {
	// 设置 Gin 模式
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	// 添加中间件
	router.Use(middleware.CORS())
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Gmail forwarding service is running",
		})
	})

	// API 路由组
	api := router.Group("/api")
	{
		// 转发对象管理
		recipients := api.Group("/recipients")
		{
			recipients.GET("", handlers.GetRecipients)
			recipients.GET("/:id", handlers.GetRecipient)
			recipients.POST("", handlers.CreateRecipient)
			recipients.PUT("/:id", handlers.UpdateRecipient)
			recipients.DELETE("/:id", handlers.DeleteRecipient)
		}

		// 转发规则管理
		rules := api.Group("/rules")
		{
			rules.GET("", handlers.GetRules)
			rules.GET("/:id", handlers.GetRule)
			rules.POST("", handlers.CreateRule)
			rules.PUT("/:id", handlers.UpdateRule)
			rules.DELETE("/:id", handlers.DeleteRule)
		}


		// 邮件处理
		api.POST("/process", handlers.ProcessEmails)
	}

	return router
}