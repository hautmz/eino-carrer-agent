// Package handler 提供 Eino Career Agent 的 HTTP 请求处理器
// Conversation Handler 处理对话列表、详情、删除等请求
package handler

import (
	"strconv"

	"github.com/hautmz/eino-carrer-agent/server/internal/pkg/response"
	"github.com/hautmz/eino-carrer-agent/server/internal/repository"

	"github.com/gin-gonic/gin"
)

// ConversationHandler 对话管理处理器
type ConversationHandler struct {
	convRepo repository.ConversationRepo
	msgRepo  repository.MessageRepo
}

// NewConversationHandler 创建对话管理处理器实例
func NewConversationHandler(convRepo repository.ConversationRepo, msgRepo repository.MessageRepo) *ConversationHandler {
	return &ConversationHandler{convRepo: convRepo, msgRepo: msgRepo}
}

// ConversationList 对话列表接口
// GET /api/conversation/list?page=1&page_size=20
func (h *ConversationHandler) ConversationList(c *gin.Context) {
	userID := GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "未认证")
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	// 查询对话列表
	conversations, total, err := h.convRepo.ListByUserID(c.Request.Context(), userID, offset, pageSize)
	if err != nil {
		response.InternalError(c, "查询对话列表失败")
		return
	}

	response.OK(c, gin.H{
		"list":      conversations,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ConversationDetail 对话详情接口（含消息列表）
// GET /api/conversation/:id
func (h *ConversationHandler) ConversationDetail(c *gin.Context) {
	userID := GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "未认证")
		return
	}

	convID := c.Param("id")
	if convID == "" {
		response.BadRequest(c, "缺少对话 ID")
		return
	}

	// 查询对话及其消息
	conv, err := h.convRepo.GetByID(c.Request.Context(), convID, true)
	if err != nil {
		response.NotFound(c, "对话不存在")
		return
	}

	// 验证对话属于当前用户
	if conv.UserID != userID {
		response.Unauthorized(c, "无权访问该对话")
		return
	}

	response.OK(c, conv)
}

// DeleteConversation 删除对话接口
// DELETE /api/conversation/:id
func (h *ConversationHandler) DeleteConversation(c *gin.Context) {
	userID := GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "未认证")
		return
	}

	convID := c.Param("id")
	if convID == "" {
		response.BadRequest(c, "缺少对话 ID")
		return
	}

	// 验证对话属于当前用户
	conv, err := h.convRepo.GetByID(c.Request.Context(), convID, false)
	if err != nil {
		response.NotFound(c, "对话不存在")
		return
	}
	if conv.UserID != userID {
		response.Unauthorized(c, "无权删除该对话")
		return
	}

	// 删除对话（级联删除消息）
	if err := h.convRepo.Delete(c.Request.Context(), convID); err != nil {
		response.InternalError(c, "删除对话失败")
		return
	}

	response.OKWithMessage(c, "删除成功", nil)
}
