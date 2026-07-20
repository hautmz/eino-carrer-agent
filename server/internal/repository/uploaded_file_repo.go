package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/hautmz/eino-carrer-agent/server/internal/domain"
)

// UploadedFileRepo 文件上传数据访问接口
type UploadedFileRepo interface {
	// Create 创建文件上传记录
	Create(ctx context.Context, file *domain.UploadedFile) error
	// GetByID 根据文件 ID 获取文件信息
	GetByID(ctx context.Context, id int64) (*domain.UploadedFile, error)
	// UpdateParsedContent 更新文件的解析内容
	UpdateParsedContent(ctx context.Context, id int64, content string) error
	// Delete 删除文件记录
	Delete(ctx context.Context, id int64) error
}

// uploadedFileRepo 文件上传数据访问实现
type uploadedFileRepo struct {
	db *gorm.DB
}

// NewUploadedFileRepo 创建文件上传数据访问实例
func NewUploadedFileRepo(db *gorm.DB) UploadedFileRepo {
	return &uploadedFileRepo{db: db}
}

// Create 创建文件上传记录
func (r *uploadedFileRepo) Create(ctx context.Context, file *domain.UploadedFile) error {
	if err := r.db.WithContext(ctx).Create(file).Error; err != nil {
		return fmt.Errorf("创建文件记录失败: %w", err)
	}
	return nil
}

// GetByID 根据文件 ID 获取文件信息
func (r *uploadedFileRepo) GetByID(ctx context.Context, id int64) (*domain.UploadedFile, error) {
	var file domain.UploadedFile
	if err := r.db.WithContext(ctx).First(&file, id).Error; err != nil {
		return nil, fmt.Errorf("根据 ID 获取文件失败: %w", err)
	}
	return &file, nil
}

// UpdateParsedContent 更新文件的解析内容
func (r *uploadedFileRepo) UpdateParsedContent(ctx context.Context, id int64, content string) error {
	if err := r.db.WithContext(ctx).Model(&domain.UploadedFile{}).
		Where("id = ?", id).
		Update("parsed_content", content).Error; err != nil {
		return fmt.Errorf("更新文件解析内容失败: %w", err)
	}
	return nil
}

// Delete 删除文件记录
func (r *uploadedFileRepo) Delete(ctx context.Context, id int64) error {
	if err := r.db.WithContext(ctx).Delete(&domain.UploadedFile{}, id).Error; err != nil {
		return fmt.Errorf("删除文件记录失败: %w", err)
	}
	return nil
}
