package errors

import (
	"fmt"
	"net/http"
)

// AppError 应用错误
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	return e.Message
}

// 预定义错误
var (
	// 通用错误
	ErrInternalServer = &AppError{
		Code:    http.StatusInternalServerError,
		Message: "Internal server error",
	}
	ErrBadRequest = &AppError{
		Code:    http.StatusBadRequest,
		Message: "Bad request",
	}
	ErrUnauthorized = &AppError{
		Code:    http.StatusUnauthorized,
		Message: "Unauthorized",
	}
	ErrForbidden = &AppError{
		Code:    http.StatusForbidden,
		Message: "Forbidden",
	}
	ErrNotFound = &AppError{
		Code:    http.StatusNotFound,
		Message: "Resource not found",
	}

	// 业务错误
	ErrStockNotFound = &AppError{
		Code:    http.StatusNotFound,
		Message: "Stock not found",
	}
	ErrInvalidStockCode = &AppError{
		Code:    http.StatusBadRequest,
		Message: "Invalid stock code",
	}
	ErrDataSourceUnavailable = &AppError{
		Code:    http.StatusServiceUnavailable,
		Message: "Data source unavailable",
	}
	ErrStrategyNotFound = &AppError{
		Code:    http.StatusNotFound,
		Message: "Strategy not found",
	}
	ErrInvalidStrategy = &AppError{
		Code:    http.StatusBadRequest,
		Message: "Invalid strategy parameters",
	}
	ErrInsufficientData = &AppError{
		Code:    http.StatusBadRequest,
		Message: "Insufficient data for analysis",
	}
)

// New 创建新的应用错误
func New(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// Newf 创建格式化的应用错误
func Newf(code int, format string, args ...interface{}) *AppError {
	return &AppError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

// WithDetails 添加错误详情
func (e *AppError) WithDetails(details string) *AppError {
	return &AppError{
		Code:    e.Code,
		Message: e.Message,
		Details: details,
	}
}

// WithDetailsf 添加格式化的错误详情
func (e *AppError) WithDetailsf(format string, args ...interface{}) *AppError {
	return &AppError{
		Code:    e.Code,
		Message: e.Message,
		Details: fmt.Sprintf(format, args...),
	}
}

// IsAppError 检查是否为应用错误
func IsAppError(err error) (*AppError, bool) {
	if appErr, ok := err.(*AppError); ok {
		return appErr, true
	}
	return nil, false
}
