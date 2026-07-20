package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/hautmz/eino-carrer-agent/server/internal/domain"
)

// UserRepo 用户数据访问接口
type UserRepo interface {
	// Create 创建新用户
	Create(ctx context.Context, user *domain.User) error
	// GetByID 根据 ID 获取用户
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	// GetByUsername 根据用户名获取用户（用于登录验证）
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	// Update 更新用户信息
	Update(ctx context.Context, user *domain.User) error
}

// userRepo 用户数据访问实现
type userRepo struct {
	db *gorm.DB
}

// NewUserRepo 创建用户数据访问实例
func NewUserRepo(db *gorm.DB) UserRepo {
	return &userRepo{db: db}
}

// Create 创建新用户
func (r *userRepo) Create(ctx context.Context, user *domain.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return fmt.Errorf("创建用户失败: %w", err)
	}
	return nil
}

// GetByID 根据 ID 获取用户
func (r *userRepo) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	var user domain.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, fmt.Errorf("根据 ID 获取用户失败: %w", err)
	}
	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (r *userRepo) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return nil, fmt.Errorf("根据用户名获取用户失败: %w", err)
	}
	return &user, nil
}

// Update 更新用户信息
func (r *userRepo) Update(ctx context.Context, user *domain.User) error {
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		return fmt.Errorf("更新用户失败: %w", err)
	}
	return nil
}
