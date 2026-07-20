// Package service 提供 Eino Career Agent 的业务逻辑层
// Upload Service 负责文件上传、存储、解析等业务
package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hautmz/eino-carrer-agent/server/internal/config"
	"github.com/hautmz/eino-carrer-agent/server/internal/domain"
	"github.com/hautmz/eino-carrer-agent/server/internal/parser"
	"github.com/hautmz/eino-carrer-agent/server/internal/pkg/logger"
	"github.com/hautmz/eino-carrer-agent/server/internal/repository"
)

// UploadService 文件上传服务
type UploadService struct {
	fileRepo repository.UploadedFileRepo
	cfg      *config.Config
}

// NewUploadService 创建文件上传服务实例
func NewUploadService(fileRepo repository.UploadedFileRepo, cfg *config.Config) *UploadService {
	return &UploadService{fileRepo: fileRepo, cfg: cfg}
}

// UploadResult 文件上传结果
type UploadResult struct {
	ID        int64  `json:"id"`         // 文件 ID
	Filename  string `json:"filename"`   // 原始文件名
	FileType  string `json:"file_type"`  // 文件类型
	FileSize  int64  `json:"file_size"`  // 文件大小
	Parsed    bool   `json:"parsed"`     // 是否已解析
}

// Upload 处理文件上传
// 1. 验证文件类型和大小
// 2. 保存文件到磁盘
// 3. 创建数据库记录
// 4. 异步解析文件内容
func (svc *UploadService) Upload(ctx context.Context, userID int64, filename string, fileData []byte) (*UploadResult, error) {
	// 1. 验证文件大小
	if int64(len(fileData)) > svc.cfg.Upload.MaxSize {
		return nil, fmt.Errorf("文件大小超过限制（最大 %d 字节）", svc.cfg.Upload.MaxSize)
	}

	// 2. 验证文件类型
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
	if !svc.isAllowedType(ext) {
		return nil, fmt.Errorf("不支持的文件类型: %s，仅支持 %v", ext, svc.cfg.Upload.AllowedTypes)
	}

	// 3. 确保上传目录存在
	uploadDir := svc.cfg.Upload.StoragePath
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("创建上传目录失败: %w", err)
	}

	// 4. 生成唯一文件名保存到磁盘
	savedName := fmt.Sprintf("%d_%d_%s", userID, len(fileData), filename)
	savedPath := filepath.Join(uploadDir, savedName)
	if err := os.WriteFile(savedPath, fileData, 0644); err != nil {
		return nil, fmt.Errorf("保存文件失败: %w", err)
	}

	// 5. 创建数据库记录
	fileRecord := &domain.UploadedFile{
		UserID:   userID,
		Filename: filename,
		FilePath: savedPath,
		FileType: ext,
		FileSize: int64(len(fileData)),
	}
	if err := svc.fileRepo.Create(ctx, fileRecord); err != nil {
		os.Remove(savedPath)
		return nil, fmt.Errorf("创建文件记录失败: %w", err)
	}

	// 6. 异步解析文件内容
	go svc.parseFileAsync(fileRecord.ID, savedPath, ext)

	logger.S().Infof("文件上传成功: %s (ID: %d, 大小: %d)", filename, fileRecord.ID, len(fileData))

	return &UploadResult{
		ID:       fileRecord.ID,
		Filename: filename,
		FileType: ext,
		FileSize: int64(len(fileData)),
		Parsed:   false,
	}, nil
}

// GetFile 获取文件信息
func (svc *UploadService) GetFile(ctx context.Context, fileID int64) (*domain.UploadedFile, error) {
	file, err := svc.fileRepo.GetByID(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %w", err)
	}
	return file, nil
}

// isAllowedType 检查文件类型是否在允许列表中
func (svc *UploadService) isAllowedType(ext string) bool {
	for _, allowed := range svc.cfg.Upload.AllowedTypes {
		if ext == allowed {
			return true
		}
	}
	return false
}

// parseFileAsync 异步解析文件内容
func (svc *UploadService) parseFileAsync(fileID int64, filePath string, fileType string) {
	ctx := context.Background()

	var content string
	var err error

	switch fileType {
	case "pdf":
		content, err = parser.ParsePDF(filePath)
	case "docx":
		content, err = parser.ParseDOCX(filePath)
	default:
		logger.S().Warnf("不支持解析的文件类型: %s", fileType)
		return
	}

	if err != nil {
		logger.S().Errorf("解析文件失败 (ID: %d): %v", fileID, err)
		return
	}

	if err := svc.fileRepo.UpdateParsedContent(ctx, fileID, content); err != nil {
		logger.S().Errorf("保存解析内容失败 (ID: %d): %v", fileID, err)
		return
	}

	logger.S().Infof("文件解析完成 (ID: %d, 内容长度: %d)", fileID, len(content))
}
