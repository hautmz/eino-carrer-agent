package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
	"github.com/hautmz/eino-carrer-agent/server/internal/agent/graph"
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
	ReportID    string `json:"report_id"`     // 生成的报告 ID
	Status      string `json:"status"`        // 状态: generating/completed/failed
	Message     string `json:"message"`       // 提示消息
	SectionCount int   `json:"section_count"` // 成功生成的章节数
}

// chatModelForTool 是 Tool 内部使用的 ChatModel 最小接口
type chatModelForTool interface {
	Generate(ctx context.Context, messages []*schema.Message, opts ...interface{}) (*schema.Message, error)
}

// NewReportTool 创建职业报告生成 Tool
// 当用户请求生成职业规划报告时，Agent 会调用此 Tool
// Tool 内部触发 Report Graph 的 12 章节并行生成流程
func NewReportTool(chatModel chatModelForTool, msgRepo repository.MessageRepo, reportRepo repository.ReportRepo, agentCfg AgentConfig) tool.BaseTool {
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
			messages, err := msgRepo.GetRecentByConversationID(ctx, input.ConversationID, agentCfg.MaxHistoryMessages)
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
				// 从对话中推断用户 ID（需要后续从 context 传入）
				userID = 0 // TODO: 从 context 获取真实 userID
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

			// 3. 执行报告生成
			graphInput := graph.ReportGraphInput{
				ConversationHistory: einoMessages,
				ReportID:            reportID,
			}

			result := graph.GenerateReport(ctx, chatModel, graphInput, agentCfg.SectionTimeout)

			// 4. 将结果写入数据库
			if result.Err != nil {
				reportRepo.UpdateStatus(ctx, reportID, "failed")
				return &ReportToolOutput{
					ReportID: reportID,
					Status:   "failed",
					Message:  fmt.Sprintf("报告生成失败: %v", result.Err),
				}, nil
			}

			// 将各章节 JSON 序列化后写入 Report
			sectionMap := mapSectionResultsToDBFields(result.Sections)
			for field, content := range sectionMap {
				jsonBytes, _ := json.Marshal(content)
				reportRepo.UpdateSection(ctx, reportID, field, string(jsonBytes))
			}
			reportRepo.UpdateStatus(ctx, reportID, "completed")

			successCount := 0
			for _, sr := range result.Sections {
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

// AgentConfig 是 Tool 需要的 Agent 配置
type AgentConfig struct {
	MaxHistoryMessages int // 最大历史消息数
	SectionTimeout     int // 单章节超时（秒）
	ReportTimeout      int // 总报告超时（秒）
}

// mapSectionResultsToDBFields 将章节结果映射到数据库字段名
func mapSectionResultsToDBFields(sections map[string]interface{}) map[string]interface{} {
	return sections // key 已经是数据库字段名（如 professional_index）
}

// NewReportToolWithUserID 创建带用户 ID 的报告生成 Tool
// 这是一个工厂方法，返回一个闭包，在调用时注入用户 ID
func NewReportToolWithUserID(chatModel chatModelForTool, msgRepo repository.MessageRepo, reportRepo repository.ReportRepo, agentCfg AgentConfig, userID int64) tool.BaseTool {
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
			logger.S().Infof("报告生成 Tool 被调用，对话ID: %s, 用户ID: %d", input.ConversationID, userID)

			messages, err := msgRepo.GetRecentByConversationID(ctx, input.ConversationID, agentCfg.MaxHistoryMessages)
			if err != nil {
				return nil, fmt.Errorf("获取对话历史失败: %w", err)
			}

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

			reportID := uuid.New().String()
			report := &domain.Report{
				ID:             reportID,
				ConversationID: input.ConversationID,
				UserID:         userID,
				Status:         "generating",
				CreatedAt:      time.Now(),
			}
			if err := reportRepo.Create(ctx, report); err != nil {
				return nil, fmt.Errorf("创建报告记录失败: %w", err)
			}

			reportCtx, cancel := context.WithTimeout(ctx, time.Duration(agentCfg.ReportTimeout)*time.Second)
			defer cancel()

			graphInput := graph.ReportGraphInput{
				ConversationHistory: einoMessages,
				ReportID:            reportID,
			}

			result := graph.GenerateReport(reportCtx, chatModel, graphInput, agentCfg.SectionTimeout)

			if result.Err != nil {
				reportRepo.UpdateStatus(ctx, reportID, "failed")
				return &ReportToolOutput{
					ReportID: reportID,
					Status:   "failed",
					Message:  fmt.Sprintf("报告生成失败: %v", result.Err),
				}, nil
			}

			sectionMap := mapSectionResultsToDBFields(result.Sections)
			for field, content := range sectionMap {
				jsonBytes, _ := json.Marshal(content)
				reportRepo.UpdateSection(ctx, reportID, field, string(jsonBytes))
			}
			reportRepo.UpdateStatus(ctx, reportID, "completed")

			successCount := 0
			for name, sr := range result.Sections {
				if sr != nil {
					if m, ok := sr.(map[string]interface{}); ok {
						if m["status"] != "failed" {
							successCount++
						}
					} else {
						successCount++
					}
				}
				_ = name
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
