package notification

import (
	"context"
	"fmt"
	"sync"

	"stock/internal/utils"
)

// Manager 通知管理器实现
type Manager struct {
	bots   map[BotType]NotificationBot
	mutex  sync.RWMutex
	logger *utils.Logger
}

// NewManager 创建通知管理器
func NewManager(logger *utils.Logger) *Manager {
	return &Manager{
		bots:   make(map[BotType]NotificationBot),
		logger: logger,
	}
}

// RegisterBot 注册机器人
func (m *Manager) RegisterBot(botType BotType, bot NotificationBot) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if bot == nil {
		return fmt.Errorf("bot cannot be nil")
	}

	m.bots[botType] = bot
	m.logger.Infof("Registered %s bot successfully", botType)
	return nil
}

// SendToBot 发送消息到指定类型的机器人
func (m *Manager) SendToBot(ctx context.Context, botType BotType, message *Message) error {
	m.mutex.RLock()
	bot, exists := m.bots[botType]
	m.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("bot type %s not registered", botType)
	}

	m.logger.Infof("Sending message to %s bot", botType)
	return bot.SendMessage(ctx, message)
}

// SendToAllBots 发送消息到所有机器人
func (m *Manager) SendToAllBots(ctx context.Context, message *Message) error {
	m.mutex.RLock()
	bots := make(map[BotType]NotificationBot)
	for botType, bot := range m.bots {
		bots[botType] = bot
	}
	m.mutex.RUnlock()

	if len(bots) == 0 {
		return fmt.Errorf("no bots registered")
	}

	var errors []error
	for botType, bot := range bots {
		if err := bot.SendMessage(ctx, message); err != nil {
			m.logger.Errorf("Failed to send message to %s bot: %v", botType, err)
			errors = append(errors, fmt.Errorf("%s: %v", botType, err))
		} else {
			m.logger.Infof("Message sent to %s bot successfully", botType)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to send to some bots: %v", errors)
	}

	return nil
}

// GetBot 获取指定类型的机器人
func (m *Manager) GetBot(botType BotType) (NotificationBot, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	bot, exists := m.bots[botType]
	return bot, exists
}

// GetRegisteredBots 获取已注册的机器人类型列表
func (m *Manager) GetRegisteredBots() []BotType {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var botTypes []BotType
	for botType := range m.bots {
		botTypes = append(botTypes, botType)
	}
	return botTypes
}

// SendMarkdownToBot 发送Markdown消息到指定机器人
func (m *Manager) SendMarkdownToBot(ctx context.Context, botType BotType, title, content string) error {
	m.mutex.RLock()
	bot, exists := m.bots[botType]
	m.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("bot type %s not registered", botType)
	}

	return bot.SendMarkdown(ctx, title, content)
}

// SendMarkdownToAllBots 发送Markdown消息到所有机器人
func (m *Manager) SendMarkdownToAllBots(ctx context.Context, title, content string) error {
	m.mutex.RLock()
	bots := make(map[BotType]NotificationBot)
	for botType, bot := range m.bots {
		bots[botType] = bot
	}
	m.mutex.RUnlock()

	if len(bots) == 0 {
		return fmt.Errorf("no bots registered")
	}

	var errors []error
	for botType, bot := range bots {
		if err := bot.SendMarkdown(ctx, title, content); err != nil {
			m.logger.Errorf("Failed to send markdown to %s bot: %v", botType, err)
			errors = append(errors, fmt.Errorf("%s: %v", botType, err))
		} else {
			m.logger.Infof("Markdown sent to %s bot successfully", botType)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to send to some bots: %v", errors)
	}

	return nil
}

// SendCardToBot 发送卡片消息到指定机器人
func (m *Manager) SendCardToBot(ctx context.Context, botType BotType, card *Card) error {
	m.mutex.RLock()
	bot, exists := m.bots[botType]
	m.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("bot type %s not registered", botType)
	}

	return bot.SendCard(ctx, card)
}

// SendCardToAllBots 发送卡片消息到所有机器人
func (m *Manager) SendCardToAllBots(ctx context.Context, card *Card) error {
	m.mutex.RLock()
	bots := make(map[BotType]NotificationBot)
	for botType, bot := range m.bots {
		bots[botType] = bot
	}
	m.mutex.RUnlock()

	if len(bots) == 0 {
		return fmt.Errorf("no bots registered")
	}

	var errors []error
	for botType, bot := range bots {
		if err := bot.SendCard(ctx, card); err != nil {
			m.logger.Errorf("Failed to send card to %s bot: %v", botType, err)
			errors = append(errors, fmt.Errorf("%s: %v", botType, err))
		} else {
			m.logger.Infof("Card sent to %s bot successfully", botType)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to send to some bots: %v", errors)
	}

	return nil
}
