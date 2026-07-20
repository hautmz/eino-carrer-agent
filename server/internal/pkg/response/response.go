// Package response 提供 Eino Career Agent 的统一 HTTP 响应格式
// 所有 API 响应均使用此格式，保持一致性
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 是统一的 API 响应结构体
type Response struct {
	Success bool        `json:"success"`            // 请求是否成功
	Message string      `json:"message"`            // 响应消息
	Data    interface{} `json:"data"`               // 响应数据，成功时填充
	Errors  interface{} `json:"errors,omitempty"`   // 错误详情，失败时填充
}

// OK 返回成功响应（HTTP 200）
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "OK",
		Data:    data,
		Errors:  nil,
	})
}

// OKWithMessage 返回成功响应，可自定义消息
func OKWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: message,
		Data:    data,
		Errors:  nil,
	})
}

// Fail 返回失败响应，可自定义 HTTP 状态码和消息
func Fail(c *gin.Context, httpCode int, message string) {
	c.JSON(httpCode, Response{
		Success: false,
		Message: message,
		Data:    nil,
		Errors:  nil,
	})
}

// FailWithErrors 返回失败响应，附带错误详情
func FailWithErrors(c *gin.Context, httpCode int, message string, errors interface{}) {
	c.JSON(httpCode, Response{
		Success: false,
		Message: message,
		Data:    nil,
		Errors:  errors,
	})
}

// Unauthorized 返回 401 未授权响应
func Unauthorized(c *gin.Context, message string) {
	Fail(c, http.StatusUnauthorized, message)
}

// BadRequest 返回 400 请求参数错误响应
func BadRequest(c *gin.Context, message string) {
	Fail(c, http.StatusBadRequest, message)
}

// InternalError 返回 500 内部服务器错误响应
func InternalError(c *gin.Context, message string) {
	Fail(c, http.StatusInternalServerError, message)
}

// NotFound 返回 404 未找到响应
func NotFound(c *gin.Context, message string) {
	Fail(c, http.StatusNotFound, message)
}
