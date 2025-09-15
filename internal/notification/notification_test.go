package notification

import (
	"context"
	"testing"
	"time"

	"stock/internal/config"
	"stock/internal/model"
	"stock/internal/utils"

	"github.com/stretchr/testify/assert"
)

func TestNotificationManager(t *testing.T) {
	// 创建logger
	logger := utils.NewLogger(config.LogConfig{
		Level:  "debug",
		Format: "text",
	})

	// 创建管理器
	manager := NewManager(logger)

	// 测试注册机器人
	mockBot := &MockBot{botType: BotTypeDingTalk}
	err := manager.RegisterBot(BotTypeDingTalk, mockBot)
	assert.NoError(t, err)

	// 测试获取机器人
	bot, exists := manager.GetBot(BotTypeDingTalk)
	assert.True(t, exists)
	assert.Equal(t, mockBot, bot)

	// 测试发送消息
	message := &Message{
		Content: "测试消息",
		MsgType: MessageTypeText,
	}

	err = manager.SendToBot(context.Background(), BotTypeDingTalk, message)
	assert.NoError(t, err)
	assert.True(t, mockBot.messageSent)
}

func TestNotificationFactory(t *testing.T) {
	logger := utils.NewLogger(config.LogConfig{
		Level:  "debug",
		Format: "text",
	})
	factory := NewFactory(logger)

	// 测试配置
	config := &Config{
		DingTalk: &DingTalkConfig{
			Enabled: true,
			Webhook: "https://oapi.dingtalk.com/robot/send?access_token=test",
			Secret:  "test-secret",
		},
		WeWork: &WeWorkConfig{
			Enabled: true,
			Webhook: "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=test",
		},
	}

	// 创建管理器
	manager, err := factory.CreateManager(config)
	assert.NoError(t, err)
	assert.NotNil(t, manager)

	// 检查注册的机器人
	enabledBots := manager.GetRegisteredBots()
	assert.Len(t, enabledBots, 2)
	assert.Contains(t, enabledBots, BotTypeDingTalk)
	assert.Contains(t, enabledBots, BotTypeWeWork)
}

func TestStockAlertTemplate(t *testing.T) {
	template := &StockAlertTemplate{
		Stock: &model.Stock{
			TsCode: "000001.SZ",
			Name:   "平安银行",
		},
		AlertType:   "买入",
		AlertReason: "技术指标突破",
		Timestamp:   time.Now(),
		Extra: map[string]interface{}{
			"价格": "10.50",
			"涨幅": "5.2%",
		},
	}

	// 测试转换为消息
	message := template.ToMessage()
	assert.Equal(t, MessageTypeText, message.MsgType)
	assert.Contains(t, message.Content, "000001.SZ")
	assert.Contains(t, message.Content, "平安银行")

	// 测试转换为Markdown
	title, content := template.ToMarkdown()
	assert.Contains(t, title, "买入")
	assert.Contains(t, content, "000001.SZ")
	assert.Contains(t, content, "平安银行")

	// 测试转换为卡片
	card := template.ToCard()
	assert.Contains(t, card.Title, "买入")
	assert.Equal(t, "#00FF00", card.Color)
	assert.True(t, len(card.Fields) > 0)
}

// MockBot 模拟机器人
type MockBot struct {
	botType     BotType
	messageSent bool
}

func (m *MockBot) SendMessage(ctx context.Context, message *Message) error {
	m.messageSent = true
	return nil
}

func (m *MockBot) SendMarkdown(ctx context.Context, title, content string) error {
	m.messageSent = true
	return nil
}

func (m *MockBot) SendCard(ctx context.Context, card *Card) error {
	m.messageSent = true
	return nil
}

func (m *MockBot) GetBotType() BotType {
	return m.botType
}
