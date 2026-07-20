// Package jwt 提供 Eino Career Agent 的 JWT Token 生成和解析功能
// 使用 golang-jwt/jwt/v5 库实现，支持 HS256 签名算法
package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenExpireError 表示 Token 已过期的错误
type TokenExpireError struct {
	Message string
}

func (e *TokenExpireError) Error() string {
	return fmt.Sprintf("token 已过期: %s", e.Message)
}

// Claims 是自定义的 JWT Claims 结构体
type Claims struct {
	UserID   int64  `json:"user_id"`   // 用户 ID
	Username string `json:"username"`  // 用户名
	jwt.RegisteredClaims
}

// JWTManager 管理 JWT Token 的生成和解析
type JWTManager struct {
	secret     []byte        // 签名密钥
	expiration time.Duration // Token 过期时间
}

// NewJWTManager 创建新的 JWT 管理器
func NewJWTManager(secret string, expiration time.Duration) *JWTManager {
	return &JWTManager{
		secret:     []byte(secret),
		expiration: expiration,
	}
}

// GenerateToken 生成 JWT Token
func (m *JWTManager) GenerateToken(userID int64, username string) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.expiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "eino-career-agent",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("生成 token 失败: %w", err)
	}

	return tokenString, nil
}

// ParseToken 解析并验证 JWT Token
func (m *JWTManager) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("不支持的签名算法: %v", token.Header["alg"])
		}
		return m.secret, nil
	})

	if err != nil {
		if err == jwt.ErrTokenExpired {
			return nil, &TokenExpireError{Message: "token 已过期"}
		}
		return nil, fmt.Errorf("解析 token 失败: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("无效的 token claims 类型")
	}

	return claims, nil
}
