// Package handler 提供 Eino Career Agent 的 HTTP 请求处理器
// Upload Handler 处理文件上传和文件信息查询
package handler

import (
	"fmt"

	"github.com/hautmz/eino-carrer-agent/server/internal/pkg/response"
	"github.com/hautmz/eino-carrer-agent/server/internal/service"

	"github.com/gin-gonic/gin"
)

// UploadHandler 文件上传处理器
type UploadHandler struct {
	uploadService *service.UploadService
}

// NewUploadHandler 创建文件上传处理器实例
func NewUploadHandler(uploadService *service.UploadService) *UploadHandler {
	return &UploadHandler{uploadService: uploadService}
}

// Upload 文件上传接口
// POST /api/upload
// multipart/form-data: file=xxx
// 限制: 10MB, 仅 PDF/DOCX
func (h *UploadHandler) Upload(c *gin.Context) {
	userID := GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "未认证")
		return
	}

	// 获取上传文件
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择要上传的文件")
		return
	}

	// 读取文件内容
	fileData, err := file.Open()
	if err != nil {
		response.InternalError(c, "读取文件失败")
		return
	}
	defer fileData.Close()

	// 读取文件字节
	buf := make([]byte, file.Size)
	if _, err := fileData.Read(buf); err != nil {
		response.InternalError(c, "读取文件内容失败")
		return
	}

	// 调用服务层处理上传
	result, err := h.uploadService.Upload(c.Request.Context(), userID, file.Filename, buf)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OKWithMessage(c, "文件上传成功", result)
}

// GetUpload 获取文件信息接口
// GET /api/upload/:id
func (h *UploadHandler) GetUpload(c *gin.Context) {
	userID := GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "未认证")
		return
	}

	// 解析文件 ID
	fileIDStr := c.Param("id")
	var fileID int64
	if _, err := fmt.Sscanf(fileIDStr, "%d", &fileID); err != nil {
		response.BadRequest(c, "无效的文件 ID")
		return
	}

	// 查询文件信息
	file, err := h.uploadService.GetFile(c.Request.Context(), fileID)
	if err != nil {
		response.NotFound(c, "文件不存在")
		return
	}

	// 验证文件属于当前用户
	if file.UserID != userID {
		response.Unauthorized(c, "无权访问该文件")
		return
	}

	response.OK(c, file)
}
