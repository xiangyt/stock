package notification

import (
	"fmt"

	"stock/internal/utils"
)

// Factory 通知工厂
type Factory struct {
	logger *utils.Logger
}

// NewFactory 创建通知工厂
func NewFactory(logger *utils.Logger) *Factory {
	return &Factory{
		logger: logger,
	}
}

// CreateManager 根据配置创建通知管理器
func (f *Factory) CreateManager(config *Config) (*Manager, error) {
	if config == nil {
		return nil, fmt.Errorf("notification config is nil")
	}

	if err := ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid notification config: %w", err)
	}

	manager := NewManager(f.logger)

	// 创建钉钉机器人
	if config.DingTalk != nil && config.DingTalk.Enabled && config.DingTalk.Webhook != "" {
		dingTalkBot := NewDingTalkBot(config.DingTalk.Webhook, config.DingTalk.Secret, f.logger)
		if err := manager.RegisterBot(BotTypeDingTalk, dingTalkBot); err != nil {
			return nil, fmt.Errorf("failed to register dingtalk bot: %v", err)
		}
		f.logger.Infof("DingTalk bot registered successfully")
	}

	// 创建企微机器人
	if config.WeWork != nil && config.WeWork.Enabled && config.WeWork.Webhook != "" {
		weWorkBot := NewWeWorkBot(config.WeWork.Webhook, f.logger)
		if err := manager.RegisterBot(BotTypeWeWork, weWorkBot); err != nil {
			return nil, fmt.Errorf("failed to register wework bot: %v", err)
		}
		f.logger.Infof("WeWork bot registered successfully")
	}

	return manager, nil
}

// CreateDingTalkBot 创建钉钉机器人
func (f *Factory) CreateDingTalkBot(webhook, secret string) NotificationBot {
	return NewDingTalkBot(webhook, secret, f.logger)
}

// CreateWeWorkBot 创建企微机器人
func (f *Factory) CreateWeWorkBot(webhook string) NotificationBot {
	return NewWeWorkBot(webhook, f.logger)
}
