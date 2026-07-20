// Package agent 提供 Eino Career Agent 的核心 AI Agent 编排功能
// 包括 ChatModel 初始化、Agent 构建、Graph 编排、Tool 定义等
package agent

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/hautmz/eino-carrer-agent/server/internal/config"
	"github.com/hautmz/eino-carrer-agent/server/internal/pkg/logger"
)

// NewChatModel 根据配置创建 OpenAI 兼容的 ChatModel 实例
// 从环境变量读取 OPENAI_API_KEY、OPENAI_BASE_URL、OPENAI_MODEL
// 支持所有兼容 OpenAI 协议的 LLM 提供商（Qwen、DeepSeek 等）
func NewChatModel(ctx context.Context, cfg *config.OpenAIConfig) (model.BaseChatModel, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY 环境变量未设置")
	}
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("OPENAI_BASE_URL 环境变量未设置")
	}
	if cfg.Model == "" {
		return nil, fmt.Errorf("OPENAI_MODEL 环境变量未设置")
	}

	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  cfg.APIKey,
		BaseURL: cfg.BaseURL,
		Model:   cfg.Model,
	})
	if err != nil {
		return nil, fmt.Errorf("创建 ChatModel 失败: %w", err)
	}

	logger.S().Infof("ChatModel 创建成功: %s (BaseURL: %s)", cfg.Model, cfg.BaseURL)

	return chatModel, nil
}

// AgentService 是 Agent 服务，封装 ChatModel 和对话逻辑
// 提供聊天、报告生成等核心 AI 功能
type AgentService struct {
	chatModel model.BaseChatModel
	cfg       *config.Config
}

// NewAgentService 创建 Agent 服务实例
func NewAgentService(ctx context.Context, cfg *config.Config) (*AgentService, error) {
	chatModel, err := NewChatModel(ctx, &cfg.OpenAI)
	if err != nil {
		return nil, err
	}

	return &AgentService{
		chatModel: chatModel,
		cfg:       cfg,
	}, nil
}

// Chat 执行普通对话
// 将对话历史 + 用户消息发送给 LLM，返回流式响应
func (s *AgentService) Chat(ctx context.Context, systemPrompt string, history []*schema.Message, userMessage string) (*schema.Message, error) {
	// 构造消息列表
	messages := make([]*schema.Message, 0, len(history)+2)

	// 添加系统提示词
	if systemPrompt != "" {
		messages = append(messages, &schema.Message{
			Role:    schema.System,
			Content: systemPrompt,
		})
	}

	// 添加历史消息
	messages = append(messages, history...)

	// 添加用户消息
	messages = append(messages, &schema.Message{
		Role:    schema.User,
		Content: userMessage,
	})

	// 调用 LLM
	resp, err := s.chatModel.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("调用 LLM 失败: %w", err)
	}

	return resp, nil
}

// ChatStream 执行流式对话
// 返回 StreamReader，可逐步读取 LLM 响应
func (s *AgentService) ChatStream(ctx context.Context, systemPrompt string, history []*schema.Message, userMessage string) (*schema.StreamReader[*schema.Message], error) {
	messages := make([]*schema.Message, 0, len(history)+2)

	if systemPrompt != "" {
		messages = append(messages, &schema.Message{
			Role:    schema.System,
			Content: systemPrompt,
		})
	}

	messages = append(messages, history...)

	messages = append(messages, &schema.Message{
		Role:    schema.User,
		Content: userMessage,
	})

	stream, err := s.chatModel.Stream(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("调用 LLM Stream 失败: %w", err)
	}

	return stream, nil
}

// GetChatModel 返回 ChatModel 实例（供 Tool 使用）
func (s *AgentService) GetChatModel() model.BaseChatModel {
	return s.chatModel
}
