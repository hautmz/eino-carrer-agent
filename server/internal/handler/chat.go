// Package handler 提供 Eino Career Agent 的 HTTP 请求处理器
// Chat Handler 处理 SSE 流式聊天请求
package handler

import (
	"github.com/hautmz/eino-carrer-agent/server/internal/pkg/response"
	"github.com/hautmz/eino-carrer-agent/server/internal/service"

	"github.com/gin-gonic/gin"
)

// ChatHandler 聊天处理器
type ChatHandler struct {
	chatService *service.ChatService
}

// NewChatHandler 创建聊天处理器实例
func NewChatHandler(chatService *service.ChatService) *ChatHandler {
	return &ChatHandler{chatService: chatService}
}

// ChatStream SSE 流式聊天接口
// POST /api/chat/stream
// 请求体: {"conversation_id": "", "message": "xxx", "file_id": null}
// 返回: SSE 事件流（message/tool_call/report_progress/report_result/error/done）
func (h *ChatHandler) ChatStream(c *gin.Context) {
	userID := GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "未认证")
		return
	}

	var req service.ChatStreamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	if req.Message == "" {
		response.BadRequest(c, "消息内容不能为空")
		return
	}

	// 委托 ChatService 处理 SSE 流式响应
	h.chatService.HandleChatStream(c, userID, &req)
}
