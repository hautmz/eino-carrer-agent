// Package handler 提供 Eino Career Agent 的 HTTP 请求处理器
// Auth Handler 处理用户注册和登录请求
package handler

import (
	"github.com/hautmz/eino-carrer-agent/server/internal/pkg/response"
	"github.com/hautmz/eino-carrer-agent/server/internal/service"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证处理器，处理注册和登录
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler 创建认证处理器实例
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register 用户注册
// POST /api/auth/register
// 请求体: {"username": "xxx", "password": "xxx"}
func (h *AuthHandler) Register(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	result, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		response.Fail(c, 409, err.Error())
		return
	}

	response.OKWithMessage(c, "注册成功", result)
}

// Login 用户登录
// POST /api/auth/login
// 请求体: {"username": "xxx", "password": "xxx"}
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	result, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.OKWithMessage(c, "登录成功", result)
}
