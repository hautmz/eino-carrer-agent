// Package sse 提供 Eino Career Agent 的 Server-Sent Events (SSE) 工具函数
// 用于在 HTTP 长连接中向客户端推送流式数据
package sse

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SSE 事件类型常量
const (
	EventMessage        = "message"         // 普通对话文本片段
	EventToolCall       = "tool_call"        // Agent 调用 Tool 通知
	EventReportProgress = "report_progress"  // 报告生成进度
	EventReportResult   = "report_result"    // 报告完整结果
	EventError          = "error"            // 错误信息
	EventDone           = "done"             // 结束标记
	EventHeartbeat      = "heartbeat"        // 心跳保活事件
)

// SetupStream 设置 SSE 流式响应的 HTTP 头
func SetupStream(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)
}

// WriteEvent 向客户端写入一个 SSE 事件
func WriteEvent(w io.Writer, event string, data string) error {
	if _, err := fmt.Fprintf(w, "event: %s\n", event); err != nil {
		return fmt.Errorf("写入 SSE event 行失败: %w", err)
	}
	if _, err := fmt.Fprintf(w, "data: %s\n", data); err != nil {
		return fmt.Errorf("写入 SSE data 行失败: %w", err)
	}
	if _, err := fmt.Fprint(w, "\n"); err != nil {
		return fmt.Errorf("写入 SSE 分隔行失败: %w", err)
	}

	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	return nil
}

// WriteHeartbeat 写入心跳保活事件
func WriteHeartbeat(w io.Writer) error {
	if _, err := fmt.Fprint(w, ": heartbeat\n\n"); err != nil {
		return fmt.Errorf("写入心跳失败: %w", err)
	}
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
	return nil
}

// WriteDone 写入结束事件
func WriteDone(w io.Writer) error {
	return WriteEvent(w, EventDone, "")
}

// WriteError 写入错误事件
func WriteError(w io.Writer, errMsg string) error {
	return WriteEvent(w, EventError, errMsg)
}
