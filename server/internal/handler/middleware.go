// Package handler 提供 Eino Career Agent 的 HTTP 请求处理器
// JWT 中间件负责验证请求中的 JWT Token 并注入用户信息到上下文
package handler

import (
	"net/http"
	"strings"

	einojwt "github.com/hautmz/eino-carrer-agent/server/internal/pkg/jwt"
	"github.com/hautmz/eino-carrer-agent/server/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// 上下文键常量，用于在 Gin 上下文中存储用户信息
const (
	ContextKeyUserID   = "user_id"   // 用户 ID
	ContextKeyUsername = "username"  // 用户名
)

// JWTAuthMiddleware 创建 JWT 认证中间件
// 从 Authorization: Bearer {token} 头解析 Token
// 验证成功后将 userID 和 username 注入 Gin 上下文
func JWTAuthMiddleware(jwtManager *einojwt.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Authorization 头获取 Token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "缺少 Authorization 头")
			c.Abort()
			return
		}

		// 验证 Bearer 格式
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "Authorization 格式错误，应为 Bearer {token}")
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 解析并验证 Token
		claims, err := jwtManager.ParseToken(tokenString)
		if err != nil {
			if _, ok := err.(*einojwt.TokenExpireError); ok {
				response.Unauthorized(c, "Token 已过期，请重新登录")
			} else {
				response.Unauthorized(c, "无效的 Token")
			}
			c.Abort()
			return
		}

		// 将用户信息注入上下文
		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyUsername, claims.Username)

		c.Next()
	}
}

// GetUserID 从 Gin 上下文中获取已认证的用户 ID
// 必须在 JWTAuthMiddleware 之后调用
func GetUserID(c *gin.Context) int64 {
	userID, exists := c.Get(ContextKeyUserID)
	if !exists {
		return 0
	}
	return userID.(int64)
}

// GetUsername 从 Gin 上下文中获取已认证的用户名
func GetUsername(c *gin.Context) string {
	username, exists := c.Get(ContextKeyUsername)
	if !exists {
		return ""
	}
	return username.(string)
}

// handleUnauthorized 处理未授权响应（统一格式）
func handleUnauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, response.Response{
		Success: false,
		Message: message,
	})
}
