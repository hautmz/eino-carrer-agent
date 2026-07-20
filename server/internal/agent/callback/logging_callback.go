// Package callback 提供 Eino Career Agent 的回调日志功能
// 实现 Eino Callback Handler 接口，记录 Agent 运行过程中的关键事件
package callback

import (
	"context"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/schema"
	"github.com/hautmz/eino-carrer-agent/server/internal/pkg/logger"
)

// LoggingCallback 实现 Eino Callback Handler 接口
// 记录 Agent/Graph/Chain/Tool 运行的开始、结束和错误事件
type LoggingCallback struct{}

// NewLoggingCallback 创建日志回调实例
func NewLoggingCallback() *LoggingCallback {
	return &LoggingCallback{}
}

// OnStart 记录组件开始运行
func (c *LoggingCallback) OnStart(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
	logger.S().Debugf("[Callback] OnStart - 组件: %s, 类型: %s", info.Name, info.Type)
	return ctx
}

// OnEnd 记录组件运行结束
func (c *LoggingCallback) OnEnd(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
	logger.S().Debugf("[Callback] OnEnd - 组件: %s, 类型: %s", info.Name, info.Type)
	return ctx
}

// OnError 记录组件运行错误
func (c *LoggingCallback) OnError(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
	logger.S().Errorf("[Callback] OnError - 组件: %s, 类型: %s, 错误: %v", info.Name, info.Type, err)
	return ctx
}

// OnStartWithStreamInput 记录流式输入开始
func (c *LoggingCallback) OnStartWithStreamInput(ctx context.Context, info *callbacks.RunInfo, input *schema.StreamReader[callbacks.CallbackInput]) context.Context {
	logger.S().Debugf("[Callback] OnStartWithStreamInput - 组件: %s", info.Name)
	return ctx
}

// OnEndWithStreamOutput 记录流式输出结束
func (c *LoggingCallback) OnEndWithStreamOutput(ctx context.Context, info *callbacks.RunInfo, output *schema.StreamReader[callbacks.CallbackOutput]) context.Context {
	logger.S().Debugf("[Callback] OnEndWithStreamOutput - 组件: %s", info.Name)
	return ctx
}
