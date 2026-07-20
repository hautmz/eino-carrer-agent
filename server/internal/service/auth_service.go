// Package service 提供 Eino Career Agent 的业务逻辑层
// Auth Service 负责用户注册、登录、Token 生成等认证相关业务
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/hautmz/eino-carrer-agent/server/internal/config"
	"github.com/hautmz/eino-carrer-agent/server/internal/domain"
	einojwt "github.com/hautmz/eino-carrer-agent/server/internal/pkg/jwt"
	"github.com/hautmz/eino-carrer-agent/server/internal/pkg/logger"
	"github.com/hautmz/eino-carrer-agent/server/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

// AuthService 认证服务，处理用户注册和登录
type AuthService struct {
	userRepo   repository.UserRepo
	jwtManager *einojwt.JWTManager
}

// NewAuthService 创建认证服务实例
func NewAuthService(userRepo repository.UserRepo, cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		jwtManager: einojwt.NewJWTManager(
			cfg.JWT.Secret,
			cfg.JWT.Expiration,
		),
	}
}

// RegisterRequest 注册请求参数
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"` // 用户名，3-50字符
	Password string `json:"password" binding:"required,min=6,max=50"` // 密码，6-50字符
}

// LoginRequest 登录请求参数
type LoginRequest struct {
	Username string `json:"username" binding:"required"` // 用户名
	Password string `json:"password" binding:"required"` // 密码
}

// AuthResponse 认证响应（注册和登录共用）
type AuthResponse struct {
	UserID   int64  `json:"user_id"`   // 用户 ID
	Username string `json:"username"`  // 用户名
	Token    string `json:"token"`     // JWT Token
	ExpireAt string `json:"expire_at"` // Token 过期时间
}

// Register 用户注册
// 1. 检查用户名是否已存在
// 2. 使用 bcrypt 哈希密码
// 3. 创建用户记录
// 4. 生成 JWT Token
func (s *AuthService) Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error) {
	// 检查用户名是否已存在
	existingUser, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("用户名已存在")
	}

	// 使用 bcrypt 哈希密码（cost=10 是推荐的安全强度）
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		return nil, fmt.Errorf("密码哈希失败: %w", err)
	}

	// 创建用户
	user := &domain.User{
		Username:     req.Username,
		PasswordHash: string(hashedPassword),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	// 生成 JWT Token
	token, err := s.jwtManager.GenerateToken(user.ID, req.Username)
	if err != nil {
		return nil, fmt.Errorf("生成 Token 失败: %w", err)
	}

	logger.S().Infof("用户注册成功: %s (ID: %d)", req.Username, user.ID)

	return &AuthResponse{
		UserID:   user.ID,
		Username: req.Username,
		Token:    token,
		ExpireAt: time.Now().Add(s.jwtManager.Expiration()).Format(time.RFC3339),
	}, nil
}

// Login 用户登录
// 1. 根据用户名查找用户
// 2. 验证密码（bcrypt 比对）
// 3. 生成 JWT Token
func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error) {
	// 查找用户
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("用户名或密码错误")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("用户名或密码错误")
	}

	// 生成 JWT Token
	token, err := s.jwtManager.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, fmt.Errorf("生成 Token 失败: %w", err)
	}

	logger.S().Infof("用户登录成功: %s (ID: %d)", user.Username, user.ID)

	return &AuthResponse{
		UserID:   user.ID,
		Username: user.Username,
		Token:    token,
		ExpireAt: time.Now().Add(s.jwtManager.Expiration()).Format(time.RFC3339),
	}, nil
}

// GetJWTManager 返回 JWT 管理器（供中间件使用）
func (s *AuthService) GetJWTManager() *einojwt.JWTManager {
	return s.jwtManager
}
