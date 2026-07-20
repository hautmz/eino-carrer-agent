package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/hautmz/eino-carrer-agent/server/internal/domain"
)

// ConversationRepo 对话数据访问接口
type ConversationRepo interface {
	// Create 创建新对话
	Create(ctx context.Context, conv *domain.Conversation) error
	// GetByID 根据对话 ID 获取对话（可预加载消息列表）
	GetByID(ctx context.Context, id string, withMessages bool) (*domain.Conversation, error)
	// ListByUserID 获取用户的对话列表（分页，按更新时间倒序）
	ListByUserID(ctx context.Context, userID int64, offset, limit int) ([]domain.Conversation, int64, error)
	// Update 更新对话信息（如标题）
	Update(ctx context.Context, conv *domain.Conversation) error
	// Delete 删除对话（级联删除消息）
	Delete(ctx context.Context, id string) error
}

// conversationRepo 对话数据访问实现
type conversationRepo struct {
	db *gorm.DB
}

// NewConversationRepo 创建对话数据访问实例
func NewConversationRepo(db *gorm.DB) ConversationRepo {
	return &conversationRepo{db: db}
}

// Create 创建新对话
func (r *conversationRepo) Create(ctx context.Context, conv *domain.Conversation) error {
	if err := r.db.WithContext(ctx).Create(conv).Error; err != nil {
		return fmt.Errorf("创建对话失败: %w", err)
	}
	return nil
}

// GetByID 根据对话 ID 获取对话
func (r *conversationRepo) GetByID(ctx context.Context, id string, withMessages bool) (*domain.Conversation, error) {
	var conv domain.Conversation
	query := r.db.WithContext(ctx)
	if withMessages {
		query = query.Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		})
	}
	if err := query.First(&conv, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("根据 ID 获取对话失败: %w", err)
	}
	return &conv, nil
}

// ListByUserID 获取用户的对话列表
func (r *conversationRepo) ListByUserID(ctx context.Context, userID int64, offset, limit int) ([]domain.Conversation, int64, error) {
	var conversations []domain.Conversation
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.Conversation{}).Where("user_id = ?", userID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计对话数量失败: %w", err)
	}

	if err := db.Order("updated_at DESC").Offset(offset).Limit(limit).Find(&conversations).Error; err != nil {
		return nil, 0, fmt.Errorf("查询对话列表失败: %w", err)
	}

	return conversations, total, nil
}

// Update 更新对话信息
func (r *conversationRepo) Update(ctx context.Context, conv *domain.Conversation) error {
	if err := r.db.WithContext(ctx).Save(conv).Error; err != nil {
		return fmt.Errorf("更新对话失败: %w", err)
	}
	return nil
}

// Delete 删除对话及其关联消息
func (r *conversationRepo) Delete(ctx context.Context, id string) error {
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Where("conversation_id = ?", id).Delete(&domain.Message{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("删除对话消息失败: %w", err)
	}

	if err := tx.Delete(&domain.Conversation{}, "id = ?", id).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("删除对话失败: %w", err)
	}

	return tx.Commit().Error
}
