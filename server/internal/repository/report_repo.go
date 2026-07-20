package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/hautmz/eino-carrer-agent/server/internal/domain"
)

// ReportRepo 报告数据访问接口
type ReportRepo interface {
	// Create 创建新报告记录
	Create(ctx context.Context, report *domain.Report) error
	// GetByID 根据报告 ID 获取报告
	GetByID(ctx context.Context, id string) (*domain.Report, error)
	// ListByUserID 获取用户的报告列表（分页）
	ListByUserID(ctx context.Context, userID int64, offset, limit int) ([]domain.Report, int64, error)
	// UpdateSection 更新报告的某个章节内容
	UpdateSection(ctx context.Context, id string, sectionName string, content string) error
	// UpdateStatus 更新报告状态
	UpdateStatus(ctx context.Context, id string, status string) error
	// Update 批量更新报告
	Update(ctx context.Context, report *domain.Report) error
}

// reportRepo 报告数据访问实现
type reportRepo struct {
	db *gorm.DB
}

// NewReportRepo 创建报告数据访问实例
func NewReportRepo(db *gorm.DB) ReportRepo {
	return &reportRepo{db: db}
}

// Create 创建新报告记录
func (r *reportRepo) Create(ctx context.Context, report *domain.Report) error {
	if err := r.db.WithContext(ctx).Create(report).Error; err != nil {
		return fmt.Errorf("创建报告失败: %w", err)
	}
	return nil
}

// GetByID 根据报告 ID 获取报告
func (r *reportRepo) GetByID(ctx context.Context, id string) (*domain.Report, error) {
	var report domain.Report
	if err := r.db.WithContext(ctx).First(&report, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("根据 ID 获取报告失败: %w", err)
	}
	return &report, nil
}

// ListByUserID 获取用户的报告列表
func (r *reportRepo) ListByUserID(ctx context.Context, userID int64, offset, limit int) ([]domain.Report, int64, error) {
	var reports []domain.Report
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.Report{}).Where("user_id = ?", userID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计报告数量失败: %w", err)
	}

	if err := db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&reports).Error; err != nil {
		return nil, 0, fmt.Errorf("查询报告列表失败: %w", err)
	}

	return reports, total, nil
}

// UpdateSection 更新报告的某个章节内容
func (r *reportRepo) UpdateSection(ctx context.Context, id string, sectionName string, content string) error {
	if err := r.db.WithContext(ctx).Model(&domain.Report{}).
		Where("id = ?", id).
		Update(sectionName, content).Error; err != nil {
		return fmt.Errorf("更新报告章节 %s 失败: %w", sectionName, err)
	}
	return nil
}

// UpdateStatus 更新报告状态
func (r *reportRepo) UpdateStatus(ctx context.Context, id string, status string) error {
	if err := r.db.WithContext(ctx).Model(&domain.Report{}).
		Where("id = ?", id).
		Update("status", status).Error; err != nil {
		return fmt.Errorf("更新报告状态失败: %w", err)
	}
	return nil
}

// Update 批量更新报告
func (r *reportRepo) Update(ctx context.Context, report *domain.Report) error {
	if err := r.db.WithContext(ctx).Save(report).Error; err != nil {
		return fmt.Errorf("更新报告失败: %w", err)
	}
	return nil
}
