package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
	"github.com/hautmz/eino-carrer-agent/server/internal/agent/graph"
	"github.com/hautmz/eino-carrer-agent/server/internal/config"
	"github.com/hautmz/eino-carrer-agent/server/internal/domain"
	"github.com/hautmz/eino-carrer-agent/server/internal/pkg/logger"
	"github.com/hautmz/eino-carrer-agent/server/internal/repository"
	"github.com/google/uuid"
)

// ReportToolInput 报告生成 Tool 的输入参数
type ReportToolInput struct {
	ConversationID string `json:"conversation_id" jsonschema:"description=对话ID，用于获取对话历史"` // 对话 ID
}

// ReportToolOutput 报告生成 Tool 的输出
type ReportToolOutput struct {
	ReportID     string `json:"report_id"`     // 生成的报告 ID
	Status       string `json:"status"`        // 状态: generating/completed/failed
	Message      string `json:"message"`       // 提示消息
	SectionCount int    `json:"section_count"` // 成功生成的章节数
}

// NewReportTool 创建职业报告生成 Tool
// 当用户请求生成职业规划报告时，Agent 会自动调用此 Tool
// Tool 内部触发 Report Graph 的 12 章节并行生成流程
func NewReportTool(chatModel model.BaseChatModel, msgRepo repository.MessageRepo, reportRepo repository.ReportRepo, cfg *config.Config) tool.BaseTool {
	graph.SetTimeouts(cfg.Agent.ReportTimeout, cfg.Agent.SectionTimeout)
	graph.SetMaxConcurrentSections(cfg.Agent.MaxConcurrentSections)

	return utils.NewTool(
		&schema.ToolInfo{
			Name: "generate_career_report",
			Desc: "当用户明确要求生成职业规划报告时调用此工具。该工具会根据对话历史中的用户信息，并行生成包含12个章节的完整职业规划报告。生成过程需要一些时间，请告知用户耐心等待。",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"conversation_id": {
					Type: "string",
					Desc: "当前对话的ID，用于获取对话历史作为报告生成的输入",
				},
			}),
		},
		func(ctx context.Context, input *ReportToolInput) (*ReportToolOutput, error) {
			logger.S().Infof("报告生成 Tool 被调用，对话ID: %s", input.ConversationID)

			// 1. 从数据库加载对话历史
			messages, err := msgRepo.GetRecentByConversationID(ctx, input.ConversationID, cfg.Agent.MaxHistoryMessages)
			if err != nil {
				return nil, fmt.Errorf("获取对话历史失败: %w", err)
			}

			// 转换为 Eino Message 格式
			einoMessages := make([]*schema.Message, 0, len(messages))
			for _, msg := range messages {
				role := schema.User
				if msg.Role == "assistant" {
					role = schema.Assistant
				}
				einoMessages = append(einoMessages, &schema.Message{
					Role:    role,
					Content: msg.Content,
				})
			}

			// 2. 创建报告记录（初始状态为 generating）
			reportID := uuid.New().String()
			var userID int64
			if len(messages) > 0 {
				userID = 0
			}

			report := &domain.Report{
				ID:             reportID,
				ConversationID: input.ConversationID,
				UserID:         userID,
				Status:         "generating",
			}
			if err := reportRepo.Create(ctx, report); err != nil {
				return nil, fmt.Errorf("创建报告记录失败: %w", err)
			}

			// 3. 设置报告总超时 context
			reportCtx, cancel := context.WithTimeout(ctx, time.Duration(cfg.Agent.ReportTimeout)*time.Second)
			defer cancel()

			// 4. 执行报告生成（传入对话历史，由 GenerateReport 自动提取画像并并行生成章节）
			result, err := graph.GenerateReport(reportCtx, chatModel, "", einoMessages...)
			if err != nil {
				reportRepo.UpdateStatus(ctx, reportID, "failed")
				return &ReportToolOutput{
					ReportID: reportID,
					Status:   "failed",
					Message:  fmt.Sprintf("报告生成失败: %v", err),
				}, nil
			}

			// 5. 将各章节结果写入数据库
			sectionMap := result.Sections
			for field, content := range sectionMap {
				jsonBytes, _ := json.Marshal(content)
				reportRepo.UpdateSection(ctx, reportID, field, string(jsonBytes))
			}
			reportRepo.UpdateStatus(ctx, reportID, "completed")

			// 统计成功生成的章节数
			successCount := 0
			for _, sr := range sectionMap {
				if sr != nil {
					if m, ok := sr.(map[string]interface{}); ok {
						if m["status"] != "failed" {
							successCount++
						}
					} else {
						successCount++
					}
				}
			}

			logger.S().Infof("报告生成完成，ID: %s，成功章节: %d/12", reportID, successCount)

			return &ReportToolOutput{
				ReportID:     reportID,
				Status:       "completed",
				Message:      fmt.Sprintf("职业规划报告已生成完成，共 %d 个章节", successCount),
				SectionCount: successCount,
			}, nil
		},
	)
}
