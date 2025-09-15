package response

import (
	"net/http"
	"time"

	"stock/pkg/errors"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// PagedResponse 分页响应结构
type PagedResponse struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	Total     int64       `json:"total"`
	Page      int         `json:"page"`
	Limit     int         `json:"limit"`
	Timestamp time.Time   `json:"timestamp"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:      http.StatusOK,
		Message:   "success",
		Data:      data,
		Timestamp: time.Now(),
	})
}

// SuccessWithMessage 带消息的成功响应
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:      http.StatusOK,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
	})
}

// SuccessPaged 分页成功响应
func SuccessPaged(c *gin.Context, data interface{}, total int64, page, limit int) {
	c.JSON(http.StatusOK, PagedResponse{
		Code:      http.StatusOK,
		Message:   "success",
		Data:      data,
		Total:     total,
		Page:      page,
		Limit:     limit,
		Timestamp: time.Now(),
	})
}

// Error 错误响应
func Error(c *gin.Context, err error) {
	if appErr, ok := errors.IsAppError(err); ok {
		c.JSON(appErr.Code, Response{
			Code:      appErr.Code,
			Message:   appErr.Message,
			Data:      appErr.Details,
			Timestamp: time.Now(),
		})
		return
	}

	// 默认内部服务器错误
	c.JSON(http.StatusInternalServerError, Response{
		Code:      http.StatusInternalServerError,
		Message:   "Internal server error",
		Timestamp: time.Now(),
	})
}

// ErrorWithCode 带状态码的错误响应
func ErrorWithCode(c *gin.Context, code int, message string) {
	c.JSON(code, Response{
		Code:      code,
		Message:   message,
		Timestamp: time.Now(),
	})
}

// BadRequest 400错误响应
func BadRequest(c *gin.Context, message string) {
	ErrorWithCode(c, http.StatusBadRequest, message)
}

// Unauthorized 401错误响应
func Unauthorized(c *gin.Context, message string) {
	ErrorWithCode(c, http.StatusUnauthorized, message)
}

// Forbidden 403错误响应
func Forbidden(c *gin.Context, message string) {
	ErrorWithCode(c, http.StatusForbidden, message)
}

// NotFound 404错误响应
func NotFound(c *gin.Context, message string) {
	ErrorWithCode(c, http.StatusNotFound, message)
}

// InternalServerError 500错误响应
func InternalServerError(c *gin.Context, message string) {
	ErrorWithCode(c, http.StatusInternalServerError, message)
}
