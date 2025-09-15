package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"stock/internal/notification"
	"stock/internal/utils"
)

func main() {
	var (
		configFile = flag.String("config", "configs/notification.yaml", "配置文件路径")
		testMode   = flag.Bool("test", false, "测试模式")
		message    = flag.String("message", "", "要发送的测试消息")
		botType    = flag.String("bot", "", "机器人类型 (dingtalk/wework)")
	)
	flag.Parse()

	// 创建logger
	logger := utils.NewLogger("notification", "info")

	if *testMode {
		runTestMode(logger, *message, *botType)
		return
	}

	// 加载配置
	config, err := loadConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 创建通知管理器
	factory := notification.NewFactory(logger)
	manager, err := factory.CreateManager(config)
	if err != nil {
		log.Fatalf("Failed to create notification manager: %v", err)
	}

	// 创建通知服务
	service := notification.NewService(manager, logger)

	// 发送健康检查消息
	ctx := context.Background()
	if err := service.SendHealthCheck(ctx); err != nil {
		log.Printf("Failed to send health check: %v", err)
	} else {
		log.Println("Health check sent successfully")
	}

	// 示例：发送系统通知
	systemNotification := &notification.SystemNotificationTemplate{
		Title:     "通知系统启动",
		Message:   "股票选择系统通知功能已启动",
		Level:     "info",
		Timestamp: time.Now(),
		Extra: map[string]interface{}{
			"版本":   "1.0.0",
			"环境":   "production",
			"启动时间": time.Now().Format("2006-01-02 15:04:05"),
		},
	}

	if err := service.SendSystemNotification(ctx, systemNotification); err != nil {
		log.Printf("Failed to send system notification: %v", err)
	} else {
		log.Println("System notification sent successfully")
	}
}

func runTestMode(logger *utils.Logger, message, botType string) {
	if message == "" {
		message = "这是一条测试消息"
	}

	// 创建测试配置
	config := &notification.Config{
		DingTalk: &notification.DingTalkConfig{
			Enabled: true,
			Webhook: "https://oapi.dingtalk.com/robot/send?access_token=test",
			Secret:  "",
		},
		WeWork: &notification.WeWorkConfig{
			Enabled: true,
			Webhook: "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=test",
		},
	}

	// 根据指定的机器人类型调整配置
	switch botType {
	case "dingtalk":
		config.WeWork.Enabled = false
	case "wework":
		config.DingTalk.Enabled = false
	}

	factory := notification.NewFactory(logger)
	manager, err := factory.CreateManager(config)
	if err != nil {
		log.Fatalf("Failed to create notification manager: %v", err)
	}

	service := notification.NewService(manager, logger)

	// 发送测试消息
	ctx := context.Background()
	if err := service.SendTextMessage(ctx, message, nil, false); err != nil {
		log.Printf("Failed to send test message: %v", err)
	} else {
		log.Println("Test message sent successfully")
	}

	// 发送测试Markdown消息
	markdownContent := fmt.Sprintf(`## 📊 测试通知

**消息内容**: %s

**发送时间**: %s

**机器人类型**: %s

---
*这是一条测试消息*`, message, time.Now().Format("2006-01-02 15:04:05"), botType)

	if err := service.SendMarkdownMessage(ctx, "测试通知", markdownContent); err != nil {
		log.Printf("Failed to send markdown message: %v", err)
	} else {
		log.Println("Markdown message sent successfully")
	}
}

func loadConfig(configFile string) (*notification.Config, error) {
	// 这里应该实现配置文件加载逻辑
	// 为了简化，返回一个默认配置
	return &notification.Config{
		DingTalk: &notification.DingTalkConfig{
			Enabled: false, // 默认禁用，需要在配置文件中启用
			Webhook: "",
			Secret:  "",
		},
		WeWork: &notification.WeWorkConfig{
			Enabled: false, // 默认禁用，需要在配置文件中启用
			Webhook: "",
		},
	}, nil
}
