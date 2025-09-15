package api

import (
	"net/http"
	"time"

	"stock/internal/notification"

	"github.com/gin-gonic/gin"
)

// NotificationHandler 通知处理器
type NotificationHandler struct {
	service *notification.Service
}

// NewNotificationHandler 创建通知处理器
func NewNotificationHandler(service *notification.Service) *NotificationHandler {
	return &NotificationHandler{
		service: service,
	}
}

// SendTextMessage 发送文本消息
func (h *NotificationHandler) SendTextMessage(c *gin.Context) {
	var req struct {
		Content   string   `json:"content" binding:"required"`
		AtMobiles []string `json:"atMobiles"`
		AtAll     bool     `json:"atAll"`
		BotType   string   `json:"botType"` // 可选，指定机器人类型
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	ctx := c.Request.Context()

	// 如果指定了机器人类型，只发送到指定类型
	if req.BotType != "" {
		botType := notification.BotType(req.BotType)
		message := &notification.Message{
			Content:   req.Content,
			MsgType:   notification.MessageTypeText,
			AtMobiles: req.AtMobiles,
			AtAll:     req.AtAll,
		}

		if err := h.service.SendToBotType(ctx, botType, message); err != nil {
			Error(c, http.StatusInternalServerError, "发送消息失败", err.Error())
			return
		}
	} else {
		// 发送到所有机器人
		if err := h.service.SendTextMessage(ctx, req.Content, req.AtMobiles, req.AtAll); err != nil {
			Error(c, http.StatusInternalServerError, "发送消息失败", err.Error())
			return
		}
	}

	Success(c, gin.H{"message": "消息发送成功"})
}

// SendMarkdownMessage 发送Markdown消息
func (h *NotificationHandler) SendMarkdownMessage(c *gin.Context) {
	var req struct {
		Title   string `json:"title" binding:"required"`
		Content string `json:"content" binding:"required"`
		BotType string `json:"botType"` // 可选，指定机器人类型
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	ctx := c.Request.Context()

	if err := h.service.SendMarkdownMessage(ctx, req.Title, req.Content); err != nil {
		Error(c, http.StatusInternalServerError, "发送Markdown消息失败", err.Error())
		return
	}

	Success(c, gin.H{"message": "Markdown消息发送成功"})
}

// SendStockAlert 发送股票提醒
func (h *NotificationHandler) SendStockAlert(c *gin.Context) {
	var req struct {
		TSCode      string                 `json:"tsCode" binding:"required"`
		StockName   string                 `json:"stockName" binding:"required"`
		AlertType   string                 `json:"alertType" binding:"required"`
		AlertReason string                 `json:"alertReason" binding:"required"`
		Extra       map[string]interface{} `json:"extra"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	// 创建模拟股票对象
	stock := &MockStock{
		TSCode: req.TSCode,
		Name:   req.StockName,
	}

	template := &notification.StockAlertTemplate{
		Stock:       stock,
		AlertType:   req.AlertType,
		AlertReason: req.AlertReason,
		Timestamp:   time.Now(),
		Extra:       req.Extra,
	}

	ctx := c.Request.Context()
	if err := h.service.SendStockAlert(ctx, template); err != nil {
		Error(c, http.StatusInternalServerError, "发送股票提醒失败", err.Error())
		return
	}

	Success(c, gin.H{"message": "股票提醒发送成功"})
}

// SendSystemNotification 发送系统通知
func (h *NotificationHandler) SendSystemNotification(c *gin.Context) {
	var req struct {
		Title   string                 `json:"title" binding:"required"`
		Message string                 `json:"message" binding:"required"`
		Level   string                 `json:"level"`
		Extra   map[string]interface{} `json:"extra"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, "参数错误", err.Error())
		return
	}

	if req.Level == "" {
		req.Level = "info"
	}

	template := &notification.SystemNotificationTemplate{
		Title:     req.Title,
		Message:   req.Message,
		Level:     req.Level,
		Timestamp: time.Now(),
		Extra:     req.Extra,
	}

	ctx := c.Request.Context()
	if err := h.service.SendSystemNotification(ctx, template); err != nil {
		Error(c, http.StatusInternalServerError, "发送系统通知失败", err.Error())
		return
	}

	Success(c, gin.H{"message": "系统通知发送成功"})
}

// GetBotStatus 获取机器人状态
func (h *NotificationHandler) GetBotStatus(c *gin.Context) {
	bots := h.service.GetAvailableBots()
	isHealthy := h.service.IsHealthy()

	Success(c, gin.H{
		"healthy":        isHealthy,
		"availableBots":  bots,
		"botCount":       len(bots),
		"supportedTypes": []string{"dingtalk", "wework"},
	})
}

// SendHealthCheck 发送健康检查
func (h *NotificationHandler) SendHealthCheck(c *gin.Context) {
	ctx := c.Request.Context()
	if err := h.service.SendHealthCheck(ctx); err != nil {
		Error(c, http.StatusInternalServerError, "发送健康检查失败", err.Error())
		return
	}

	Success(c, gin.H{"message": "健康检查发送成功"})
}

// MockStock 模拟股票结构（临时使用，实际应该使用 model.Stock）
type MockStock struct {
	TSCode string `json:"tsCode"`
	Name   string `json:"name"`
}
