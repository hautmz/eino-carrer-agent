package tools

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
	"github.com/hautmz/eino-carrer-agent/server/internal/parser"
	"github.com/hautmz/eino-carrer-agent/server/internal/pkg/logger"
	"github.com/hautmz/eino-carrer-agent/server/internal/repository"
)

// FileParseToolInput 文件解析 Tool 的输入参数
type FileParseToolInput struct {
	FileID int64 `json:"file_id" jsonschema:"description=上传文件的ID"` // 文件 ID
}

// FileParseToolOutput 文件解析 Tool 的输出
type FileParseToolOutput struct {
	FileID      int64  `json:"file_id"`       // 文件 ID
	Filename    string `json:"filename"`      // 原始文件名
	FileType    string `json:"file_type"`     // 文件类型
	Content     string `json:"content"`       // 解析出的文本内容
	ContentLen  int    `json:"content_len"`   // 文本内容长度
	Message     string `json:"message"`       // 提示消息
}

// NewFileParseTool 创建文件解析 Tool
// 当用户上传了简历/文档文件时，Agent 调用此 Tool 解析文件内容
func NewFileParseTool(fileRepo repository.UploadedFileRepo) tool.BaseTool {
	return utils.NewTool(
		&schema.ToolInfo{
			Name: "parse_resume_file",
			Desc: "当用户上传了简历文件（PDF或DOCX格式）时调用此工具。该工具会解析文件内容，提取文本信息用于后续的职业分析和报告生成。",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"file_id": {
					Type: "integer",
					Desc: "用户上传的文件ID",
				},
			}),
		},
		func(ctx context.Context, input *FileParseToolInput) (*FileParseToolOutput, error) {
			logger.S().Infof("文件解析 Tool 被调用，文件ID: %d", input.FileID)

			// 1. 从数据库获取文件信息
			file, err := fileRepo.GetByID(ctx, input.FileID)
			if err != nil {
				return nil, fmt.Errorf("获取文件信息失败: %w", err)
			}

			// 2. 如果已有解析内容，直接返回
			if file.ParsedContent != "" {
				return &FileParseToolOutput{
					FileID:     file.ID,
					Filename:   file.Filename,
					FileType:   file.FileType,
					Content:    file.ParsedContent,
					ContentLen: len(file.ParsedContent),
					Message:    "文件内容已解析完成",
				}, nil
			}

			// 3. 根据文件类型调用对应解析器
			var content string
			switch file.FileType {
			case "pdf":
				content, err = parser.ParsePDF(file.FilePath)
			case "docx":
				content, err = parser.ParseDOCX(file.FilePath)
			default:
				return nil, fmt.Errorf("不支持的文件类型: %s", file.FileType)
			}

			if err != nil {
				return nil, fmt.Errorf("解析文件失败: %w", err)
			}

			// 4. 将解析内容存入数据库
			if err := fileRepo.UpdateParsedContent(ctx, file.ID, content); err != nil {
				logger.S().Warnf("保存文件解析内容失败: %v", err)
			}

			logger.S().Infof("文件解析完成，文件: %s，内容长度: %d", file.Filename, len(content))

			return &FileParseToolOutput{
				FileID:     file.ID,
				Filename:   file.Filename,
				FileType:   file.FileType,
				Content:    content,
				ContentLen: len(content),
				Message:    "文件内容解析完成",
			}, nil
		},
	)
}
