package notification

import (
	"context"
	"fmt"
	"time"

	"stock/internal/utils"
)

// Service 通知服务
type Service struct {
	manager *Manager
	logger  *utils.Logger
}

// NewService 创建通知服务
func NewService(manager *Manager, logger *utils.Logger) *Service {
	return &Service{
		manager: manager,
		logger:  logger,
	}
}

// SendStockAlert 发送股票提醒
func (s *Service) SendStockAlert(ctx context.Context, template *StockAlertTemplate) error {
	s.logger.Infof("Sending stock alert for %s: %s", template.Stock.TsCode, template.AlertType)

	// 发送Markdown格式消息
	title, content := template.ToMarkdown()
	return s.manager.SendMarkdownToAllBots(ctx, title, content)
}

// SendPerformanceReport 发送业绩报表通知
func (s *Service) SendPerformanceReport(ctx context.Context, template *PerformanceReportTemplate) error {
	s.logger.Infof("Sending performance report for %s", template.Stock.TsCode)

	title, content := template.ToMarkdown()
	return s.manager.SendMarkdownToAllBots(ctx, title, content)
}

// SendSystemNotification 发送系统通知
func (s *Service) SendSystemNotification(ctx context.Context, template *SystemNotificationTemplate) error {
	s.logger.Infof("Sending system notification: %s", template.Title)

	title, content := template.ToMarkdown()
	return s.manager.SendMarkdownToAllBots(ctx, title, content)
}

// SendTextMessage 发送简单文本消息
func (s *Service) SendTextMessage(ctx context.Context, content string, atMobiles []string, atAll bool) error {
	message := &Message{
		Content:   content,
		MsgType:   MessageTypeText,
		AtMobiles: atMobiles,
		AtAll:     atAll,
	}

	return s.manager.SendToAllBots(ctx, message)
}

// SendMarkdownMessage 发送Markdown消息
func (s *Service) SendMarkdownMessage(ctx context.Context, title, content string) error {
	return s.manager.SendMarkdownToAllBots(ctx, title, content)
}

// SendCardMessage 发送卡片消息
func (s *Service) SendCardMessage(ctx context.Context, card *Card) error {
	return s.manager.SendCardToAllBots(ctx, card)
}

// SendToBotType 发送消息到指定类型的机器人
func (s *Service) SendToBotType(ctx context.Context, botType BotType, message *Message) error {
	return s.manager.SendToBot(ctx, botType, message)
}

// GetAvailableBots 获取可用的机器人类型
func (s *Service) GetAvailableBots() []BotType {
	return s.manager.GetRegisteredBots()
}

// IsHealthy 检查服务健康状态
func (s *Service) IsHealthy() bool {
	bots := s.manager.GetRegisteredBots()
	return len(bots) > 0
}

// SendHealthCheck 发送健康检查消息
func (s *Service) SendHealthCheck(ctx context.Context) error {
	template := &SystemNotificationTemplate{
		Title:     "系统健康检查",
		Message:   "通知系统运行正常",
		Level:     "info",
		Timestamp: time.Now(),
		Extra: map[string]interface{}{
			"可用机器人数量": len(s.manager.GetRegisteredBots()),
			"机器人类型":   s.manager.GetRegisteredBots(),
		},
	}

	return s.SendSystemNotification(ctx, template)
}

// BatchSendStockAlerts 批量发送股票提醒
func (s *Service) BatchSendStockAlerts(ctx context.Context, templates []*StockAlertTemplate) error {
	if len(templates) == 0 {
		return nil
	}

	s.logger.Infof("Batch sending %d stock alerts", len(templates))

	var errors []error
	for i, template := range templates {
		if err := s.SendStockAlert(ctx, template); err != nil {
			s.logger.Errorf("Failed to send stock alert %d: %v", i, err)
			errors = append(errors, err)
		}

		// 添加延迟避免频率限制
		if i < len(templates)-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to send %d/%d alerts: %v", len(errors), len(templates), errors)
	}

	return nil
}

// ScheduledNotification 定时通知结构
type ScheduledNotification struct {
	ID       string
	Title    string
	Content  string
	BotTypes []BotType
	Schedule string // cron表达式
	Enabled  bool
}

// SendScheduledNotification 发送定时通知
func (s *Service) SendScheduledNotification(ctx context.Context, notification *ScheduledNotification) error {
	s.logger.Infof("Sending scheduled notification: %s", notification.Title)

	message := &Message{
		Content: fmt.Sprintf("📅 定时通知\n\n%s\n\n%s", notification.Title, notification.Content),
		MsgType: MessageTypeText,
	}

	// 如果指定了机器人类型，只发送到指定类型
	if len(notification.BotTypes) > 0 {
		var errors []error
		for _, botType := range notification.BotTypes {
			if err := s.manager.SendToBot(ctx, botType, message); err != nil {
				errors = append(errors, fmt.Errorf("%s: %v", botType, err))
			}
		}
		if len(errors) > 0 {
			return fmt.Errorf("failed to send to some bots: %v", errors)
		}
		return nil
	}

	// 否则发送到所有机器人
	return s.manager.SendToAllBots(ctx, message)
}
