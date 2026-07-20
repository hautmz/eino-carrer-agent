// Package service 提供 Eino Career Agent 的业务逻辑层
// Chat Service 负责聊天对话管理、消息持久化、SSE 流式推送
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/hautmz/eino-carrer-agent/server/internal/agent"
	"github.com/hautmz/eino-carrer-agent/server/internal/agent/graph"
	"github.com/hautmz/eino-carrer-agent/server/internal/agent/prompts"
	"github.com/hautmz/eino-carrer-agent/server/internal/config"
	"github.com/hautmz/eino-carrer-agent/server/internal/domain"
	"github.com/hautmz/eino-carrer-agent/server/internal/pkg/logger"
	"github.com/hautmz/eino-carrer-agent/server/internal/pkg/sse"
	"github.com/hautmz/eino-carrer-agent/server/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ChatService 聊天服务，处理对话上下文、LLM 调用、SSE 推送
type ChatService struct {
	agentService    *agent.AgentService
	convRepo        repository.ConversationRepo
	msgRepo         repository.MessageRepo
	reportRepo      repository.ReportRepo
	cfg             *config.Config
}

// NewChatService 创建聊天服务实例
func NewChatService(
	agentService *agent.AgentService,
	convRepo repository.ConversationRepo,
	msgRepo repository.MessageRepo,
	reportRepo repository.ReportRepo,
	cfg *config.Config,
) *ChatService {
	return &ChatService{
		agentService: agentService,
		convRepo:     convRepo,
		msgRepo:      msgRepo,
		reportRepo:   reportRepo,
		cfg:          cfg,
	}
}

// ChatStreamRequest 聊天流式请求参数
type ChatStreamRequest struct {
	ConversationID string `json:"conversation_id"` // 对话 ID（为空则新建）
	Message        string `json:"message"`         // 用户消息内容
	FileID         *int64 `json:"file_id"`         // 关联文件 ID（可选）
}

// HandleChatStream 处理聊天 SSE 流式请求
// 核心流程:
// 1. 解析请求，获取或创建对话
// 2. 保存用户消息到 DB
// 3. 加载历史消息转为 Eino Message 格式
// 4. 调用 AgentService.ChatStream 获取流式响应
// 5. 逐 chunk 推送 SSE 事件，同时拼接完整回复
// 6. 保存 assistant 回复到 DB
func (svc *ChatService) HandleChatStream(c *gin.Context, userID int64, req *ChatStreamRequest) {
	ctx := c.Request.Context()

	// 设置 SSE 流式响应头
	sse.SetupStream(c)

	// 1. 获取或创建对话
	conv, err := svc.getOrCreateConversation(ctx, userID, req.ConversationID, req.Message)
	if err != nil {
		sse.WriteError(c.Writer, fmt.Sprintf("获取对话失败: %v", err))
		sse.WriteDone(c.Writer)
		return
	}

	// 2. 保存用户消息到 DB
	userMsg := &domain.Message{
		ConversationID: conv.ID,
		Role:           "user",
		Content:        req.Message,
		FileID:         req.FileID,
	}
	if err := svc.msgRepo.Create(ctx, userMsg); err != nil {
		sse.WriteError(c.Writer, fmt.Sprintf("保存用户消息失败: %v", err))
		sse.WriteDone(c.Writer)
		return
	}

	// 3. 加载历史消息
	history, err := svc.loadHistoryMessages(ctx, conv.ID)
	if err != nil {
		sse.WriteError(c.Writer, fmt.Sprintf("加载历史消息失败: %v", err))
		sse.WriteDone(c.Writer)
		return
	}

	// 4. 启动心跳 goroutine
	heartbeatCtx, cancelHeartbeat := svc.startHeartbeat(c)

	// 5. 调用 LLM 流式生成
	stream, err := svc.agentService.ChatStream(ctx, prompts.SystemPrompt, history, req.Message)
	if err != nil {
		cancelHeartbeat()
		sse.WriteError(c.Writer, fmt.Sprintf("调用 AI 失败: %v", err))
		sse.WriteDone(c.Writer)
		return
	}

	// 6. 读取流式响应，推送 SSE 事件
	var fullContent string
	defer cancelHeartbeat()

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.S().Errorf("读取流式响应失败: %v", err)
			break
		}

		if chunk.Content != "" {
			fullContent += chunk.Content

			// 推送 message 事件
			data, _ := json.Marshal(map[string]string{
				"content": chunk.Content,
			})
			if err := sse.WriteEvent(c.Writer, sse.EventMessage, string(data)); err != nil {
				logger.S().Warnf("写入 SSE 事件失败: %v", err)
				break
			}
		}
	}

	// 7. 保存 assistant 回复到 DB
	if fullContent != "" {
		assistantMsg := &domain.Message{
			ConversationID: conv.ID,
			Role:           "assistant",
			Content:        fullContent,
		}
		if err := svc.msgRepo.Create(ctx, assistantMsg); err != nil {
			logger.S().Errorf("保存 assistant 消息失败: %v", err)
		}

		// 检查是否包含报告生成意图，如需要则触发报告生成
		svc.checkAndTriggerReport(heartbeatCtx, c, conv.ID, userID, fullContent)
	}

	// 8. 发送完成事件
	sse.WriteDone(c.Writer)
}

// getOrCreateConversation 获取现有对话或创建新对话
func (svc *ChatService) getOrCreateConversation(ctx context.Context, userID int64, convID string, firstMsg string) (*domain.Conversation, error) {
	// 如果提供了对话 ID，获取已有对话
	if convID != "" {
		conv, err := svc.convRepo.GetByID(ctx, convID, false)
		if err != nil {
			return nil, fmt.Errorf("获取对话失败: %w", err)
		}
		// 验证对话属于当前用户
		if conv.UserID != userID {
			return nil, fmt.Errorf("无权访问该对话")
		}
		return conv, nil
	}

	// 创建新对话
	title := firstMsg
	if len(title) > 50 {
		title = title[:50] + "..."
	}

	conv := &domain.Conversation{
		ID:     uuid.New().String(),
		UserID: userID,
		Title:  title,
	}
	if err := svc.convRepo.Create(ctx, conv); err != nil {
		return nil, fmt.Errorf("创建对话失败: %w", err)
	}

	return conv, nil
}

// loadHistoryMessages 加载对话的历史消息并转为 Eino Message 格式
func (svc *ChatService) loadHistoryMessages(ctx context.Context, convID string) ([]*schema.Message, error) {
	// 获取最近 N 条消息
	limit := svc.cfg.Agent.MaxHistoryMessages
	messages, err := svc.msgRepo.GetRecentByConversationID(ctx, convID, limit)
	if err != nil {
		return nil, err
	}

	// 转换为 Eino Message 格式
	var result []*schema.Message
	for _, msg := range messages {
		var role schema.RoleType
		switch msg.Role {
		case "user":
			role = schema.User
		case "assistant":
			role = schema.Assistant
		case "tool":
			role = schema.Tool
		default:
			continue
		}
		result = append(result, &schema.Message{
			Role:    role,
			Content: msg.Content,
		})
	}

	return result, nil
}

// startHeartbeat 启动心跳 goroutine
func (svc *ChatService) startHeartbeat(c *gin.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	interval := time.Duration(svc.cfg.SSE.HeartbeatInterval) * time.Second

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := sse.WriteHeartbeat(c.Writer); err != nil {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return ctx, cancel
}

// checkAndTriggerReport 检查是否需要触发报告生成
// 简化实现：如果 assistant 回复中包含"生成报告"关键词，则触发
func (svc *ChatService) checkAndTriggerReport(ctx context.Context, c *gin.Context, convID string, userID int64, assistantContent string) {
	// TODO: 后续由 Agent 自动通过 Tool 调用触发报告生成
	// 当前简化版本通过关键词匹配触发
}

// GenerateReportAndPush 生成职业规划报告并通过 SSE 推送进度
// 在后台 goroutine 中执行报告生成，实时推送每个章节的完成状态
func (svc *ChatService) GenerateReportAndPush(c *gin.Context, convID string, userID int64, profileContext string) {
	ctx := c.Request.Context()

	// 创建报告记录
	report := &domain.Report{
		ID:             uuid.New().String(),
		ConversationID: convID,
		UserID:         userID,
		Status:         "generating",
	}
	if err := svc.reportRepo.Create(ctx, report); err != nil {
		sse.WriteError(c.Writer, fmt.Sprintf("创建报告记录失败: %v", err))
		return
	}

	// 推送报告开始事件
	startData, _ := json.Marshal(map[string]string{
		"report_id": report.ID,
		"status":    "generating",
	})
	sse.WriteEvent(c.Writer, sse.EventReportProgress, string(startData))

	// 调用 Report Graph 生成报告
	chatModel := svc.agentService.GetChatModel()
	result, err := graph.GenerateReport(ctx, chatModel, profileContext)
	if err != nil {
		// 更新报告状态为失败
		svc.reportRepo.UpdateStatus(ctx, report.ID, "failed")
		sse.WriteError(c.Writer, fmt.Sprintf("报告生成失败: %v", err))
		return
	}

	// 更新报告内容和状态
	// 从 ReportResult.Sections 映射中提取各章节 JSON 字符串
	sectionMap := result.Sections
	report.ProfessionalIndex = toJSONString(sectionMap["professional_index"])
	report.MyselfReport = toJSONString(sectionMap["myself_report"])
	report.AchievementSuperiority = toJSONString(sectionMap["achievement_superiority"])
	report.CareerExperience = toJSONString(sectionMap["career_experience"])
	report.MotivationValues = toJSONString(sectionMap["motivation_values"])
	report.SkillHeatmap = toJSONString(sectionMap["skill_heatmap"])
	report.InterestAssessment = toJSONString(sectionMap["interest_assessment"])
	report.CareerRecommendations = toJSONString(sectionMap["career_recommendations"])
	report.IndustryAnalysis = toJSONString(sectionMap["industry_analysis"])
	report.GoalSetting = toJSONString(sectionMap["goal_setting"])
	report.ActionPlan = toJSONString(sectionMap["action_plan"])
	report.SummaryOutlook = toJSONString(sectionMap["summary_outlook"])
	report.Status = "completed"

	if err := svc.reportRepo.Update(ctx, report); err != nil {
		logger.S().Errorf("更新报告内容失败: %v", err)
	}

	// 推送报告完成事件
	resultData, _ := json.Marshal(map[string]interface{}{
		"report_id": report.ID,
		"status":    "completed",
	})
	sse.WriteEvent(c.Writer, sse.EventReportResult, string(resultData))
}

// toJSONString 将 interface{} 转换为 JSON 字符串
func toJSONString(v interface{}) string {
	if v == nil {
		return ""
	}
	bytes, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(bytes)
}
