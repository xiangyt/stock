package logger

import (
	"io"
	"os"
	"path/filepath"

	"stock/internal/config"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger 日志记录器
type Logger struct {
	*logrus.Logger
}

// NewLogger 创建新的日志记录器
func NewLogger(cfg config.LogConfig) *Logger {
	logger := logrus.New()

	// 设置日志级别
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// 设置日志格式
	if cfg.Format == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	// 设置输出
	var writers []io.Writer
	writers = append(writers, os.Stdout)

	// 如果配置了日志文件，添加文件输出
	if cfg.File != "" {
		// 确保日志目录存在
		if err := os.MkdirAll(filepath.Dir(cfg.File), 0755); err != nil {
			logger.Errorf("Failed to create logger directory: %v", err)
		} else {
			fileWriter := &lumberjack.Logger{
				Filename:   cfg.File,
				MaxSize:    cfg.MaxSize,
				MaxBackups: cfg.MaxBackups,
				MaxAge:     cfg.MaxAge,
				Compress:   cfg.Compress,
			}
			writers = append(writers, fileWriter)
		}
	}

	logger.SetOutput(io.MultiWriter(writers...))

	return &Logger{Logger: logger}
}

// WithField 添加字段
func (l *Logger) WithField(key string, value interface{}) *logrus.Entry {
	return l.Logger.WithField(key, value)
}

// WithFields 添加多个字段
func (l *Logger) WithFields(fields logrus.Fields) *logrus.Entry {
	return l.Logger.WithFields(fields)
}

// WithError 添加错误字段
func (l *Logger) WithError(err error) *logrus.Entry {
	return l.Logger.WithError(err)
}
