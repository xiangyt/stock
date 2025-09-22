package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// WeWorkBot 企微机器人
type WeWorkBot struct {
	webhook string
	client  *http.Client
}

// NewWeWorkBot 创建企微机器人
func NewWeWorkBot(webhook string) *WeWorkBot {
	return &WeWorkBot{
		webhook: webhook,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetBotType 获取机器人类型
func (w *WeWorkBot) GetBotType() BotType {
	return BotTypeWeWork
}

// SendMessage 发送消息
func (w *WeWorkBot) SendMessage(ctx context.Context, message *Message) error {
	switch message.MsgType {
	case MessageTypeText:
		return w.sendTextMessage(ctx, message)
	case MessageTypeMarkdown:
		return w.sendMarkdownMessage(ctx, message)
	case MessageTypeCard:
		return w.sendCardMessage(ctx, message)
	default:
		return fmt.Errorf("unsupported message type: %s", message.MsgType)
	}
}

// SendMarkdown 发送Markdown格式消息
func (w *WeWorkBot) SendMarkdown(ctx context.Context, title, content string) error {
	message := &Message{
		Content: content,
		MsgType: MessageTypeMarkdown,
	}
	return w.SendMessage(ctx, message)
}

// SendCard 发送卡片消息
func (w *WeWorkBot) SendCard(ctx context.Context, card *Card) error {
	message := &Message{
		MsgType: MessageTypeCard,
		Extra: map[string]interface{}{
			"card": card,
		},
	}
	return w.SendMessage(ctx, message)
}

// sendTextMessage 发送文本消息
func (w *WeWorkBot) sendTextMessage(ctx context.Context, message *Message) error {
	content := message.Content

	// 添加@信息
	if len(message.AtMobiles) > 0 {
		var mentions []string
		for _, mobile := range message.AtMobiles {
			mentions = append(mentions, fmt.Sprintf("<@%s>", mobile))
		}
		content = strings.Join(mentions, " ") + " " + content
	}

	payload := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]interface{}{
			"content": content,
		},
	}

	return w.sendRequest(ctx, payload)
}

// sendMarkdownMessage 发送Markdown消息
func (w *WeWorkBot) sendMarkdownMessage(ctx context.Context, message *Message) error {
	content := message.Content

	// 添加@信息
	if len(message.AtMobiles) > 0 {
		var mentions []string
		for _, mobile := range message.AtMobiles {
			mentions = append(mentions, fmt.Sprintf("<@%s>", mobile))
		}
		content = strings.Join(mentions, " ") + "\n" + content
	}

	payload := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]interface{}{
			"content": content,
		},
	}

	return w.sendRequest(ctx, payload)
}

// sendCardMessage 发送卡片消息
func (w *WeWorkBot) sendCardMessage(ctx context.Context, message *Message) error {
	cardData, ok := message.Extra["card"]
	if !ok {
		return fmt.Errorf("card data not found in message extra")
	}

	card, ok := cardData.(*Card)
	if !ok {
		return fmt.Errorf("invalid card data type")
	}

	// 构建模板卡片
	templateCard := map[string]interface{}{
		"card_type": "text_notice",
		"source": map[string]interface{}{
			"icon_url":   "",
			"desc":       "股票选择系统",
			"desc_color": 0,
		},
		"main_title": map[string]interface{}{
			"title": card.Title,
			"desc":  card.Description,
		},
	}

	// 添加字段信息
	if len(card.Fields) > 0 {
		var emphasisContent []map[string]interface{}
		for _, field := range card.Fields {
			emphasisContent = append(emphasisContent, map[string]interface{}{
				"title": field.Name,
				"desc":  field.Value,
			})
		}
		templateCard["emphasis_content"] = emphasisContent
	}

	// 添加跳转链接
	if card.URL != "" {
		templateCard["jump_list"] = []map[string]interface{}{
			{
				"type":  1,
				"url":   card.URL,
				"title": "查看详情",
			},
		}
	}

	payload := map[string]interface{}{
		"msgtype":       "template_card",
		"template_card": templateCard,
	}

	return w.sendRequest(ctx, payload)
}

// sendRequest 发送HTTP请求
func (w *WeWorkBot) sendRequest(ctx context.Context, payload map[string]interface{}) error {
	// 序列化payload
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "POST", w.webhook, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	// 检查错误码
	if errCode, ok := result["errcode"]; ok {
		if code, ok := errCode.(float64); ok && code != 0 {
			errMsg := "unknown error"
			if msg, ok := result["errmsg"]; ok {
				if msgStr, ok := msg.(string); ok {
					errMsg = msgStr
				}
			}
			return fmt.Errorf("wework API error: code=%v, msg=%s", errCode, errMsg)
		}
	}

	return nil
}
