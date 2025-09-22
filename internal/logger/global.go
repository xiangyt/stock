package logger

import (
	"sync"
)

var (
	globalLogger *Logger
	loggerOnce   sync.Once
	loggerMutex  sync.RWMutex
)

// InitGlobalLogger 初始化全局日志器（只能调用一次）
func InitGlobalLogger(cfg LogConfig) {
	loggerOnce.Do(func() {
		loggerMutex.Lock()
		defer loggerMutex.Unlock()
		globalLogger = NewLogger(cfg)
	})
}

// GetGlobalLogger 获取全局日志器
func GetGlobalLogger() *Logger {
	loggerMutex.RLock()
	defer loggerMutex.RUnlock()

	if globalLogger == nil {
		// 如果没有初始化，使用默认配置
		defaultConfig := LogConfig{
			Level:  "info",
			Format: "text",
		}
		globalLogger = NewLogger(defaultConfig)
	}

	return globalLogger
}

// SetGlobalLogger 设置全局日志器（用于测试或特殊场景）
func SetGlobalLogger(logger *Logger) {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()
	globalLogger = logger
}

// ResetGlobalLogger 重置全局日志器（主要用于测试）
func ResetGlobalLogger() {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()
	globalLogger = nil
	loggerOnce = sync.Once{}
}

// 全局日志方法 - 直接使用，无需传递 logger 参数

// Debug 记录调试信息
func Debug(args ...interface{}) {
	GetGlobalLogger().Debug(args...)
}

// Debugf 记录格式化调试信息
func Debugf(format string, args ...interface{}) {
	GetGlobalLogger().Debugf(format, args...)
}

// Info 记录信息
func Info(args ...interface{}) {
	GetGlobalLogger().Info(args...)
}

// Infof 记录格式化信息
func Infof(format string, args ...interface{}) {
	GetGlobalLogger().Infof(format, args...)
}

// Warn 记录警告信息
func Warn(args ...interface{}) {
	GetGlobalLogger().Warn(args...)
}

// Warnf 记录格式化警告信息
func Warnf(format string, args ...interface{}) {
	GetGlobalLogger().Warnf(format, args...)
}

// Error 记录错误信息
func Error(args ...interface{}) {
	GetGlobalLogger().Error(args...)
}

// Errorf 记录格式化错误信息
func Errorf(format string, args ...interface{}) {
	GetGlobalLogger().Errorf(format, args...)
}

// Fatal 记录致命错误并退出程序
func Fatal(args ...interface{}) {
	GetGlobalLogger().Fatal(args...)
}

// Fatalf 记录格式化致命错误并退出程序
func Fatalf(format string, args ...interface{}) {
	GetGlobalLogger().Fatalf(format, args...)
}

// WithField 添加字段并返回 Entry
func WithField(key string, value interface{}) *Logger {
	entry := GetGlobalLogger().WithField(key, value)
	return &Logger{Logger: entry.Logger}
}

// WithFields 添加多个字段并返回 Entry
func WithFields(fields map[string]interface{}) *Logger {
	logrusFields := make(map[string]interface{})
	for k, v := range fields {
		logrusFields[k] = v
	}
	entry := GetGlobalLogger().WithFields(logrusFields)
	return &Logger{Logger: entry.Logger}
}

// WithError 添加错误字段并返回 Entry
func WithError(err error) *Logger {
	entry := GetGlobalLogger().WithError(err)
	return &Logger{Logger: entry.Logger}
}
