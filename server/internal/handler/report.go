// Package handler 提供 Eino Career Agent 的 HTTP 请求处理器
// Report Handler 处理报告列表和详情查询
package handler

import (
	"strconv"

	"github.com/hautmz/eino-carrer-agent/server/internal/pkg/response"
	"github.com/hautmz/eino-carrer-agent/server/internal/repository"

	"github.com/gin-gonic/gin"
)

// ReportHandler 报告处理器
type ReportHandler struct {
	reportRepo repository.ReportRepo
}

// NewReportHandler 创建报告处理器实例
func NewReportHandler(reportRepo repository.ReportRepo) *ReportHandler {
	return &ReportHandler{reportRepo: reportRepo}
}

// ReportList 报告列表接口
// GET /api/report/list?page=1&page_size=10
func (h *ReportHandler) ReportList(c *gin.Context) {
	userID := GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "未认证")
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	// 查询报告列表
	reports, total, err := h.reportRepo.ListByUserID(c.Request.Context(), userID, offset, pageSize)
	if err != nil {
		response.InternalError(c, "查询报告列表失败")
		return
	}

	response.OK(c, gin.H{
		"list":      reports,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ReportDetail 报告详情接口
// GET /api/report/:id
func (h *ReportHandler) ReportDetail(c *gin.Context) {
	userID := GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "未认证")
		return
	}

	reportID := c.Param("id")
	if reportID == "" {
		response.BadRequest(c, "缺少报告 ID")
		return
	}

	// 查询报告
	report, err := h.reportRepo.GetByID(c.Request.Context(), reportID)
	if err != nil {
		response.NotFound(c, "报告不存在")
		return
	}

	// 验证报告属于当前用户
	if report.UserID != userID {
		response.Unauthorized(c, "无权访问该报告")
		return
	}

	response.OK(c, report)
}
