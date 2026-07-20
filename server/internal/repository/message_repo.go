package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/hautmz/eino-carrer-agent/server/internal/domain"
)

// MessageRepo 消息数据访问接口
type MessageRepo interface {
	// Create 创建单条消息
	Create(ctx context.Context, msg *domain.Message) error
	// BatchCreate 批量创建消息
	BatchCreate(ctx context.Context, messages []domain.Message) error
	// ListByConversationID 根据对话 ID 查询消息列表（分页，按时间正序）
	ListByConversationID(ctx context.Context, conversationID string, offset, limit int) ([]domain.Message, error)
	// GetRecentByConversationID 获取对话最近 N 条消息
	GetRecentByConversationID(ctx context.Context, conversationID string, limit int) ([]domain.Message, error)
	// DeleteByConversationID 删除对话的所有消息
	DeleteByConversationID(ctx context.Context, conversationID string) error
}

// messageRepo 消息数据访问实现
type messageRepo struct {
	db *gorm.DB
}

// NewMessageRepo 创建消息数据访问实例
func NewMessageRepo(db *gorm.DB) MessageRepo {
	return &messageRepo{db: db}
}

// Create 创建单条消息
func (r *messageRepo) Create(ctx context.Context, msg *domain.Message) error {
	if err := r.db.WithContext(ctx).Create(msg).Error; err != nil {
		return fmt.Errorf("创建消息失败: %w", err)
	}
	return nil
}

// BatchCreate 批量创建消息
func (r *messageRepo) BatchCreate(ctx context.Context, messages []domain.Message) error {
	if len(messages) == 0 {
		return nil
	}
	if err := r.db.WithContext(ctx).CreateInBatches(messages, 100).Error; err != nil {
		return fmt.Errorf("批量创建消息失败: %w", err)
	}
	return nil
}

// ListByConversationID 根据对话 ID 分页查询消息
func (r *messageRepo) ListByConversationID(ctx context.Context, conversationID string, offset, limit int) ([]domain.Message, error) {
	var messages []domain.Message
	if err := r.db.WithContext(ctx).
		Where("conversation_id = ?", conversationID).
		Order("created_at ASC").
		Offset(offset).
		Limit(limit).
		Find(&messages).Error; err != nil {
		return nil, fmt.Errorf("查询对话消息列表失败: %w", err)
	}
	return messages, nil
}

// GetRecentByConversationID 获取对话最近 N 条消息
func (r *messageRepo) GetRecentByConversationID(ctx context.Context, conversationID string, limit int) ([]domain.Message, error) {
	var messages []domain.Message
	if err := r.db.WithContext(ctx).
		Where("conversation_id = ?", conversationID).
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error; err != nil {
		return nil, fmt.Errorf("获取对话最近消息失败: %w", err)
	}
	// 反转为正序（最早在前）
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	return messages, nil
}

// DeleteByConversationID 删除对话的所有消息
func (r *messageRepo) DeleteByConversationID(ctx context.Context, conversationID string) error {
	if err := r.db.WithContext(ctx).Where("conversation_id = ?", conversationID).Delete(&domain.Message{}).Error; err != nil {
		return fmt.Errorf("删除对话消息失败: %w", err)
	}
	return nil
}
