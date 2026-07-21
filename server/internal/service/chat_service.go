// Package service 提供 Eino Career Agent 的业务逻辑层
// Chat Service 负责聊天对话管理、消息持久化、Agent 事件流到 SSE 推送
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/hautmz/eino-carrer-agent/server/internal/agent"
	"github.com/hautmz/eino-carrer-agent/server/internal/config"
	"github.com/hautmz/eino-carrer-agent/server/internal/domain"
	"github.com/hautmz/eino-carrer-agent/server/internal/pkg/logger"
	"github.com/hautmz/eino-carrer-agent/server/internal/pkg/sse"
	"github.com/hautmz/eino-carrer-agent/server/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ChatService 聊天服务，处理对话上下文、Agent 调用、SSE 推送
type ChatService struct {
	careerAgent *agent.CareerAgent
	convRepo    repository.ConversationRepo
	msgRepo     repository.MessageRepo
	cfg         *config.Config
}

// NewChatService 创建聊天服务实例
func NewChatService(
	careerAgent *agent.CareerAgent,
	convRepo repository.ConversationRepo,
	msgRepo repository.MessageRepo,
	cfg *config.Config,
) *ChatService {
	return &ChatService{
		careerAgent: careerAgent,
		convRepo:    convRepo,
		msgRepo:     msgRepo,
		cfg:         cfg,
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
// 4. 调用 Agent.StreamChat 获取事件流
// 5. 读取 Agent 事件流，推送 SSE 事件
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
	_, cancelHeartbeat := svc.startHeartbeat(c)
	defer cancelHeartbeat()

	// 5. 调用 Agent 流式对话
	iterator, err := svc.careerAgent.StreamChat(ctx, history, req.Message)
	if err != nil {
		sse.WriteError(c.Writer, fmt.Sprintf("调用 Agent 失败: %v", err))
		sse.WriteDone(c.Writer)
		return
	}

	// 6. 读取 Agent 事件流，推送 SSE 事件
	var fullContent string
	agent.ReadAgentStream(iterator, &agent.StreamCallbacks{
		OnMessage: func(content string) {
			fullContent += content
			data, _ := json.Marshal(map[string]string{"content": content})
			if err := sse.WriteEvent(c.Writer, sse.EventMessage, string(data)); err != nil {
				logger.S().Warnf("写入 SSE message 事件失败: %v", err)
			}
		},
		OnToolCall: func(name, result string) {
			data, _ := json.Marshal(map[string]interface{}{
				"tool_name": name,
				"result":    result,
			})
			if err := sse.WriteEvent(c.Writer, sse.EventToolCall, string(data)); err != nil {
				logger.S().Warnf("写入 SSE tool_call 事件失败: %v", err)
			}
		},
		OnError: func(errMsg string) {
			sse.WriteError(c.Writer, errMsg)
		},
		OnDone: func(content string) {
			fullContent = content
		},
	})

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
	}

	// 8. 发送完成事件
	sse.WriteDone(c.Writer)
}

// getOrCreateConversation 获取现有对话或创建新对话
func (svc *ChatService) getOrCreateConversation(ctx context.Context, userID int64, convID string, firstMsg string) (*domain.Conversation, error) {
	if convID != "" {
		conv, err := svc.convRepo.GetByID(ctx, convID, false)
		if err != nil {
			return nil, fmt.Errorf("获取对话失败: %w", err)
		}
		if conv.UserID != userID {
			return nil, fmt.Errorf("无权访问该对话")
		}
		return conv, nil
	}

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
	limit := svc.cfg.Agent.MaxHistoryMessages
	messages, err := svc.msgRepo.GetRecentByConversationID(ctx, convID, limit)
	if err != nil {
		return nil, err
	}

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
