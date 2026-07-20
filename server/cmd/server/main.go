// Package main 是 Eino Career Agent 服务端程序入口
// 负责初始化配置、日志、数据库、路由等组件，并启动 HTTP 服务
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/hautmz/eino-carrer-agent/server/internal/config"
	"github.com/hautmz/eino-carrer-agent/server/internal/pkg/database"
	"github.com/hautmz/eino-carrer-agent/server/internal/pkg/logger"
	"github.com/hautmz/eino-carrer-agent/server/internal/pkg/response"
)

func main() {
	fmt.Println("=== Eino Career Agent Starting ===")

	// ===== 1. 加载配置 =====
	configPath := "./configs/config.yaml"
	if len(os.Args) > 1 {
		// 支持通过命令行参数指定配置文件路径
		configPath = os.Args[1]
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	fmt.Printf("Config loaded: port=%d, mode=%s\n", cfg.Server.Port, cfg.Server.Mode)

	// ===== 2. 初始化日志 =====
	if err := logger.Init(cfg.Server.Mode); err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}
	defer logger.Sync()

	logger.S().Infof("配置加载完成，服务端口: %d, 模式: %s", cfg.Server.Port, cfg.Server.Mode)

	// ===== 3. 初始化数据库 =====
	db, err := database.Init(cfg.Database.Path)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	logger.S().Infof("数据库初始化完成: %s", cfg.Database.Path)

	// ===== 4. 设置 Gin 运行模式 =====
	gin.SetMode(cfg.Server.Mode)

	// ===== 5. 创建 Gin 引擎和路由 =====
	r := setupRouter(cfg)

	// 将 db 实例存入 Gin 的自定义上下文，供 Handler 使用
	// 后续各 Handler 通过 c.MustGet("db") 获取数据库实例
	r.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Set("config", cfg)
		c.Next()
	})

	// ===== 6. 启动 HTTP 服务 =====
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	fmt.Printf("Eino Career Agent 服务启动，监听地址: %s\n", addr)
	logger.S().Infof("Eino Career Agent 服务启动，监听地址: %s", addr)

	// 直接启动 HTTP 服务（阻塞直到服务停止）
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

// setupRouter 配置所有 HTTP 路由
// 包括健康检查、认证、聊天、文件上传、报告、对话管理等路由
func setupRouter(cfg *config.Config) *gin.Engine {
	r := gin.New()

	// 使用 gin 的 Recovery 中间件（捕获 panic 防止服务崩溃）
	r.Use(gin.Recovery())

	// CORS 中间件（开发环境允许所有来源）
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// ===== 公开路由（无需认证） =====
	public := r.Group("/api")
	{
		// 健康检查接口
		public.GET("/health", healthCheck)

		// 认证路由（后续 Task 4.1 实现）
		// auth := public.Group("/auth")
		// {
		//     auth.POST("/register", handler.Register)
		//     auth.POST("/login", handler.Login)
		// }
	}

	// ===== 认证路由（需要 JWT Token） =====
	// authRequired := r.Group("/api")
	// authRequired.Use(middleware.JWTAuth())
	// {
	//     // 聊天 SSE 流式接口
	//     authRequired.POST("/chat/stream", handler.ChatStream)
	//     // 文件上传
	//     authRequired.POST("/upload", handler.Upload)
	//     authRequired.GET("/upload/:id", handler.GetUpload)
	//     // 报告
	//     authRequired.GET("/report/list", handler.ReportList)
	//     authRequired.GET("/report/:id", handler.ReportDetail)
	//     // 对话管理
	//     authRequired.GET("/conversation/list", handler.ConversationList)
	//     authRequired.GET("/conversation/:id", handler.ConversationDetail)
	//     authRequired.DELETE("/conversation/:id", handler.DeleteConversation)
	// }

	return r
}

// healthCheck 健康检查处理函数
// 返回服务运行状态，可用于负载均衡健康检测
func healthCheck(c *gin.Context) {
	response.OK(c, gin.H{
		"status":  "running",
		"service": "eino-career-agent",
	})
}
