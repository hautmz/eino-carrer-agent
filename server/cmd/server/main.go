// Package main 是 Eino Career Agent 服务端程序入口
// 负责初始化配置、日志、数据库、Agent、服务层、路由等组件，并启动 HTTP 服务
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/hautmz/eino-carrer-agent/server/internal/agent"
	"github.com/hautmz/eino-carrer-agent/server/internal/config"
	"github.com/hautmz/eino-carrer-agent/server/internal/handler"
	"github.com/hautmz/eino-carrer-agent/server/internal/pkg/database"
	"github.com/hautmz/eino-carrer-agent/server/internal/pkg/logger"
	"github.com/hautmz/eino-carrer-agent/server/internal/pkg/response"
	"github.com/hautmz/eino-carrer-agent/server/internal/repository"
	"github.com/hautmz/eino-carrer-agent/server/internal/service"
)

func main() {
	fmt.Println("=== Eino Career Agent Starting ===")

	ctx := context.Background()

	// ===== 1. 加载配置 =====
	configPath := "./configs/config.yaml"
	if len(os.Args) > 1 {
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

	// ===== 4. 初始化 Repository 层 =====
	userRepo := repository.NewUserRepo(db)
	convRepo := repository.NewConversationRepo(db)
	msgRepo := repository.NewMessageRepo(db)
	reportRepo := repository.NewReportRepo(db)
	fileRepo := repository.NewUploadedFileRepo(db)

	// ===== 5. 初始化 Agent 服务 =====
	agentService, err := agent.NewAgentService(ctx, cfg)
	if err != nil {
		log.Fatalf("初始化 Agent 服务失败: %v", err)
	}
	logger.S().Info("Agent 服务初始化完成")

	// ===== 6. 初始化 Service 层 =====
	authService := service.NewAuthService(userRepo, cfg)
	chatService := service.NewChatService(agentService, convRepo, msgRepo, reportRepo, cfg)
	uploadService := service.NewUploadService(fileRepo, cfg)

	// ===== 7. 初始化 Handler 层 =====
	authHandler := handler.NewAuthHandler(authService)
	chatHandler := handler.NewChatHandler(chatService)
	uploadHandler := handler.NewUploadHandler(uploadService)
	reportHandler := handler.NewReportHandler(reportRepo)
	convHandler := handler.NewConversationHandler(convRepo, msgRepo)

	// ===== 8. 设置 Gin 运行模式 =====
	gin.SetMode(cfg.Server.Mode)

	// ===== 9. 创建路由 =====
	r := setupRouter(cfg, authService, authHandler, chatHandler, uploadHandler, reportHandler, convHandler)

	// ===== 10. 启动 HTTP 服务 =====
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	fmt.Printf("Eino Career Agent 服务启动，监听地址: %s\n", addr)
	logger.S().Infof("Eino Career Agent 服务启动，监听地址: %s", addr)

	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

// setupRouter 配置所有 HTTP 路由
// 包括健康检查、认证、聊天、文件上传、报告、对话管理等路由
func setupRouter(
	cfg *config.Config,
	authService *service.AuthService,
	authHandler *handler.AuthHandler,
	chatHandler *handler.ChatHandler,
	uploadHandler *handler.UploadHandler,
	reportHandler *handler.ReportHandler,
	convHandler *handler.ConversationHandler,
) *gin.Engine {
	r := gin.New()

	// Recovery 中间件（捕获 panic 防止服务崩溃）
	r.Use(gin.Recovery())

	// CORS 中间件
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
		public.GET("/health", healthCheck)

		auth := public.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}
	}

	// ===== 认证路由（需要 JWT Token） =====
	authRequired := r.Group("/api")
	authRequired.Use(handler.JWTAuthMiddleware(authService.GetJWTManager()))
	{
		authRequired.POST("/chat/stream", chatHandler.ChatStream)
		authRequired.POST("/upload", uploadHandler.Upload)
		authRequired.GET("/upload/:id", uploadHandler.GetUpload)
		authRequired.GET("/report/list", reportHandler.ReportList)
		authRequired.GET("/report/:id", reportHandler.ReportDetail)
		authRequired.GET("/conversation/list", convHandler.ConversationList)
		authRequired.GET("/conversation/:id", convHandler.ConversationDetail)
		authRequired.DELETE("/conversation/:id", convHandler.DeleteConversation)
	}

	// ===== 托管前端 SPA 静态文件 =====
	// 前端构建产物放在 web/dist 目录
	spaDir := filepath.Join("..", "web", "dist")
	if info, err := os.Stat(spaDir); err == nil && info.IsDir() {
		r.Use(staticServe(spaDir))
		logger.S().Infof("前端静态文件目录: %s", spaDir)
	}

	return r
}

// staticServe 返回一个中间件，托管前端 SPA 静态文件
// 对于非 /api 路径的 GET 请求，优先匹配静态文件，未匹配则返回 index.html（SPA 回退）
func staticServe(root string) gin.HandlerFunc {
	fileServer := http.FileServer(http.Dir(root))

	return func(c *gin.Context) {
		// 仅处理非 /api 路径的 GET 请求
		if c.Request.Method != "GET" || len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:5] == "/api/" {
			c.Next()
			return
		}

		// 尝试匹配静态文件
		path := c.Request.URL.Path
		if path == "/" {
			path = "/index.html"
		}

		fullPath := filepath.Join(root, filepath.Clean(path))
		if _, err := os.Stat(fullPath); err == nil {
			c.Request.URL.Path = path
			fileServer.ServeHTTP(c.Writer, c.Request)
			c.Abort()
			return
		}

		// SPA 回退：所有未匹配的路径返回 index.html
		indexFile := filepath.Join(root, "index.html")
		if _, err := os.Stat(indexFile); err == nil {
			c.Request.URL.Path = "/index.html"
			fileServer.ServeHTTP(c.Writer, c.Request)
			c.Abort()
			return
		}

		c.Next()
	}
}

// healthCheck 健康检查处理函数
func healthCheck(c *gin.Context) {
	response.OK(c, gin.H{
		"status":  "running",
		"service": "eino-career-agent",
	})
}
