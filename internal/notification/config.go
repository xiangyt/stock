package notification

import (
	"fmt"
	"os"
	"strconv"
)

// Config 通知配置
type Config struct {
	DingTalk *DingTalkConfig `mapstructure:"dingtalk"`
	WeWork   *WeWorkConfig   `mapstructure:"wework"`
}

// DingTalkConfig 钉钉机器人配置
type DingTalkConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Webhook string `mapstructure:"webhook"`
	Secret  string `mapstructure:"secret"`
}

// WeWorkConfig 企微机器人配置
type WeWorkConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Webhook string `mapstructure:"webhook"`
}

// LoadConfigFromEnv 从环境变量加载配置
func LoadConfigFromEnv() (*Config, error) {
	config := &Config{}

	// 钉钉配置
	if webhook := os.Getenv("DINGTALK_WEBHOOK"); webhook != "" {
		config.DingTalk = &DingTalkConfig{
			Enabled: getBoolEnv("DINGTALK_ENABLED", true),
			Webhook: webhook,
			Secret:  os.Getenv("DINGTALK_SECRET"),
		}
	}

	// 企微配置
	if webhook := os.Getenv("WEWORK_WEBHOOK"); webhook != "" {
		config.WeWork = &WeWorkConfig{
			Enabled: getBoolEnv("WEWORK_ENABLED", true),
			Webhook: webhook,
		}
	}

	return config, nil
}

// MergeWithFileConfig 将环境变量配置与文件配置合并
// 环境变量优先级更高
func MergeWithFileConfig(envConfig, fileConfig *Config) *Config {
	if envConfig == nil {
		return fileConfig
	}
	if fileConfig == nil {
		return envConfig
	}

	merged := &Config{}

	// 合并钉钉配置
	if envConfig.DingTalk != nil {
		merged.DingTalk = envConfig.DingTalk
	} else {
		merged.DingTalk = fileConfig.DingTalk
	}

	// 合并企微配置
	if envConfig.WeWork != nil {
		merged.WeWork = envConfig.WeWork
	} else {
		merged.WeWork = fileConfig.WeWork
	}

	return merged
}

// ValidateConfig 验证配置的有效性
func ValidateConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("配置不能为空")
	}

	// 检查是否至少启用了一个机器人
	hasEnabled := false

	if config.DingTalk != nil && config.DingTalk.Enabled {
		if config.DingTalk.Webhook == "" {
			return fmt.Errorf("钉钉机器人已启用但Webhook为空")
		}
		hasEnabled = true
	}

	if config.WeWork != nil && config.WeWork.Enabled {
		if config.WeWork.Webhook == "" {
			return fmt.Errorf("企微机器人已启用但Webhook为空")
		}
		hasEnabled = true
	}

	if !hasEnabled {
		return fmt.Errorf("至少需要启用一个机器人")
	}

	return nil
}

// getBoolEnv 获取布尔类型的环境变量
func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// MaskSensitiveInfo 隐藏敏感信息用于日志输出
func MaskSensitiveInfo(config *Config) *Config {
	if config == nil {
		return nil
	}

	masked := &Config{}

	if config.DingTalk != nil {
		masked.DingTalk = &DingTalkConfig{
			Enabled: config.DingTalk.Enabled,
			Webhook: maskWebhook(config.DingTalk.Webhook),
			Secret:  maskSecret(config.DingTalk.Secret),
		}
	}

	if config.WeWork != nil {
		masked.WeWork = &WeWorkConfig{
			Enabled: config.WeWork.Enabled,
			Webhook: maskWebhook(config.WeWork.Webhook),
		}
	}

	return masked
}

// maskWebhook 隐藏Webhook URL中的敏感部分
func maskWebhook(webhook string) string {
	if webhook == "" {
		return ""
	}
	if len(webhook) > 50 {
		return webhook[:30] + "..." + webhook[len(webhook)-10:]
	}
	return webhook
}

// maskSecret 隐藏密钥
func maskSecret(secret string) string {
	if secret == "" {
		return ""
	}
	if len(secret) > 10 {
		return secret[:6] + "..." + secret[len(secret)-4:]
	}
	return "***"
}
