package notification

import "context"

// NotificationBot 机器人通知接口
type NotificationBot interface {
	// SendMessage 发送消息
	SendMessage(ctx context.Context, message *Message) error

	// SendMarkdown 发送Markdown格式消息
	SendMarkdown(ctx context.Context, title, content string) error

	// SendCard 发送卡片消息（如果支持）
	SendCard(ctx context.Context, card *Card) error

	// GetBotType 获取机器人类型
	GetBotType() BotType
}

// BotType 机器人类型
type BotType string

const (
	BotTypeDingTalk BotType = "dingtalk" // 钉钉机器人
	BotTypeWeWork   BotType = "wework"   // 企微机器人
)

// Message 通用消息结构
type Message struct {
	Content   string                 `json:"content"`   // 消息内容
	MsgType   MessageType            `json:"msgType"`   // 消息类型
	AtMobiles []string               `json:"atMobiles"` // @的手机号列表
	AtAll     bool                   `json:"atAll"`     // 是否@所有人
	Extra     map[string]interface{} `json:"extra"`     // 额外参数
}

// MessageType 消息类型
type MessageType string

const (
	MessageTypeText     MessageType = "text"     // 文本消息
	MessageTypeMarkdown MessageType = "markdown" // Markdown消息
	MessageTypeCard     MessageType = "card"     // 卡片消息
)

// Card 卡片消息结构
type Card struct {
	Title       string      `json:"title"`       // 卡片标题
	Description string      `json:"description"` // 卡片描述
	URL         string      `json:"url"`         // 跳转链接
	Color       string      `json:"color"`       // 卡片颜色
	Fields      []CardField `json:"fields"`      // 卡片字段
}

// CardField 卡片字段
type CardField struct {
	Name  string `json:"name"`  // 字段名
	Value string `json:"value"` // 字段值
	Short bool   `json:"short"` // 是否短字段
}

// NotificationManager 通知管理器接口
type NotificationManager interface {
	// RegisterBot 注册机器人
	RegisterBot(botType BotType, bot NotificationBot) error

	// SendToBot 发送消息到指定类型的机器人
	SendToBot(ctx context.Context, botType BotType, message *Message) error

	// SendToAllBots 发送消息到所有机器人
	SendToAllBots(ctx context.Context, message *Message) error

	// GetBot 获取指定类型的机器人
	GetBot(botType BotType) (NotificationBot, bool)
}
