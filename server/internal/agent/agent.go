// Package agent 提供 Eino Career Agent 的核心 AI Agent 编排功能
// 使用 Eino ADK 的 ChatModelAgent 构建真正的 ReAct Agent
// Agent 可自动判断用户意图，选择调用 Tools（生成报告/解析文件/查询报告）
package agent

import (
	"context"
	"fmt"
	"io"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/hautmz/eino-carrer-agent/server/internal/agent/prompts"
	"github.com/hautmz/eino-carrer-agent/server/internal/config"
	"github.com/hautmz/eino-carrer-agent/server/internal/pkg/logger"
	"github.com/hautmz/eino-carrer-agent/server/internal/repository"
	agenttools "github.com/hautmz/eino-carrer-agent/server/internal/agent/tools"
)

// CareerAgent 是职业规划 AI Agent，封装 Eino ADK ChatModelAgent
// 提供流式对话、Tool 自动调用等能力
type CareerAgent struct {
	agent    *adk.ChatModelAgent
	chatModel model.BaseChatModel
}

// NewCareerAgent 创建职业规划 Agent
// 1. 初始化 ChatModel（OpenAI 兼容协议）
// 2. 创建 3 个 Agent Tools
// 3. 用 Eino ADK NewChatModelAgent 构建真正的 ReAct Agent
func NewCareerAgent(ctx context.Context, cfg *config.Config, msgRepo repository.MessageRepo, reportRepo repository.ReportRepo, fileRepo repository.UploadedFileRepo) (*CareerAgent, error) {
	// 1. 初始化 ChatModel
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  cfg.OpenAI.APIKey,
		BaseURL: cfg.OpenAI.BaseURL,
		Model:   cfg.OpenAI.Model,
	})
	if err != nil {
		return nil, fmt.Errorf("创建 ChatModel 失败: %w", err)
	}
	logger.S().Infof("ChatModel 创建成功: %s (BaseURL: %s)", cfg.OpenAI.Model, cfg.OpenAI.BaseURL)

	// 2. 创建 Tools
	var tools []tool.BaseTool

	reportTool := agenttools.NewReportTool(chatModel, msgRepo, reportRepo, cfg)
	tools = append(tools, reportTool)

	fileParseTool := agenttools.NewFileParseTool(fileRepo)
	tools = append(tools, fileParseTool)

	reportQueryTool := agenttools.NewReportQueryTool(reportRepo, 0)
	tools = append(tools, reportQueryTool)

	// 3. 构建 ChatModelAgent
	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "CareerAgent",
		Description: "专业的职业规划 AI 助手，能够根据用户对话自动生成职业规划报告、解析简历文件、查询历史报告",
		Instruction: prompts.SystemPrompt,
		Model:       chatModel,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: tools,
			},
		},
		MaxIterations: 10,
	})
	if err != nil {
		return nil, fmt.Errorf("创建 ChatModelAgent 失败: %w", err)
	}

	logger.S().Info("CareerAgent 创建成功（含3个Tools: generate_career_report/parse_resume_file/query_career_report）")

	return &CareerAgent{
		agent:    agent,
		chatModel: chatModel,
	}, nil
}

// StreamChat 流式执行 Agent 对话
// 返回 AsyncIterator，逐事件读取 Agent 的推理和 Tool 调用过程
func (a *CareerAgent) StreamChat(ctx context.Context, history []*schema.Message, userMessage string) (*adk.AsyncIterator[*adk.AgentEvent], error) {
	// 构建输入消息列表
	messages := make([]*schema.Message, 0, len(history)+1)
	messages = append(messages, history...)
	messages = append(messages, schema.UserMessage(userMessage))

	input := &adk.AgentInput{
		Messages:        messages,
		EnableStreaming: true,
	}

	// 运行 Agent，返回事件迭代器
	iterator := a.agent.Run(ctx, input)

	return iterator, nil
}

// GetChatModel 返回 ChatModel 实例（供 Report Graph 等直接使用）
func (a *CareerAgent) GetChatModel() model.BaseChatModel {
	return a.chatModel
}

// ReadAgentStream 读取 Agent 事件流，通过回调函数推送 SSE 事件
// 处理 Agent 的三种事件类型：
//   - Assistant 消息（流式文本）→ 推送 message 事件
//   - Tool 调用结果 → 推送 tool_call 事件
//   - 错误 → 推送 error 事件
func ReadAgentStream(iterator *adk.AsyncIterator[*adk.AgentEvent], callbacks *StreamCallbacks) {
	var fullContent string

	for {
		event, ok := iterator.Next()
		if !ok {
			break
		}

		if event.Err != nil {
			if callbacks.OnError != nil {
				callbacks.OnError(event.Err.Error())
			}
			break
		}

		if event.Output == nil || event.Output.MessageOutput == nil {
			continue
		}

		mv := event.Output.MessageOutput

		// 处理流式消息
		if mv.IsStreaming && mv.MessageStream != nil {
			role := mv.Role
			toolName := mv.ToolName

			if role == schema.Assistant {
				// 流式 Assistant 消息
				for {
					chunk, err := mv.MessageStream.Recv()
					if err == io.EOF {
						break
					}
					if err != nil {
						break
					}
					if chunk.Content != "" {
						fullContent += chunk.Content
						if callbacks.OnMessage != nil {
							callbacks.OnMessage(chunk.Content)
						}
					}
				}
			} else if role == schema.Tool {
				// Tool 调用结果
				toolResult, _ := concatStream(mv.MessageStream)
				if callbacks.OnToolCall != nil {
					callbacks.OnToolCall(toolName, toolResult)
				}
			}
		} else if mv.Message != nil {
			// 非流式消息
			if mv.Role == schema.Assistant && mv.Message.Content != "" {
				fullContent += mv.Message.Content
				if callbacks.OnMessage != nil {
					callbacks.OnMessage(mv.Message.Content)
				}
			} else if mv.Role == schema.Tool {
				if callbacks.OnToolCall != nil {
					callbacks.OnToolCall(mv.ToolName, mv.Message.Content)
				}
			}
		}
	}

	if callbacks.OnDone != nil {
		callbacks.OnDone(fullContent)
	}
}

// StreamCallbacks Agent 事件流回调函数
type StreamCallbacks struct {
	OnMessage  func(content string)     // 收到 Assistant 文本片段
	OnToolCall func(name, result string) // 收到 Tool 调用结果
	OnError    func(err string)          // 发生错误
	OnDone     func(fullContent string)  // 流结束，返回完整内容
}

// concatStream 将 StreamReader 中的所有消息拼接为字符串
func concatStream(stream *schema.StreamReader[*schema.Message]) (string, error) {
	var result string
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return result, err
		}
		result += msg.Content
	}
	return result, nil
}
