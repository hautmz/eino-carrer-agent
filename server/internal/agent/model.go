// Package agent 提供 Eino Career Agent 的核心 AI Agent 编排功能
// 包括 ChatModel 初始化、Agent 构建、Graph 编排、Tool 定义等
package agent

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/hautmz/eino-carrer-agent/server/internal/config"
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

	// 使用 eino-ext 的 OpenAI ChatModel 实现
	// 通过自定义 BaseURL 兼容所有 OpenAI 协议的 LLM 服务
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  cfg.APIKey,
		BaseURL: cfg.BaseURL,
		Model:   cfg.Model,
	})
	if err != nil {
		return nil, fmt.Errorf("创建 ChatModel 失败: %w", err)
	}

	return chatModel, nil
}
