package notification

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"stock/internal/utils"
)

// DingTalkBot 钉钉机器人
type DingTalkBot struct {
	webhook string
	secret  string
	logger  *utils.Logger
	client  *http.Client
}

// NewDingTalkBot 创建钉钉机器人
func NewDingTalkBot(webhook, secret string, logger *utils.Logger) *DingTalkBot {
	return &DingTalkBot{
		webhook: webhook,
		secret:  secret,
		logger:  logger,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetBotType 获取机器人类型
func (d *DingTalkBot) GetBotType() BotType {
	return BotTypeDingTalk
}

// SendMessage 发送消息
func (d *DingTalkBot) SendMessage(ctx context.Context, message *Message) error {
	switch message.MsgType {
	case MessageTypeText:
		return d.sendTextMessage(ctx, message)
	case MessageTypeMarkdown:
		return d.sendMarkdownMessage(ctx, message)
	default:
		return fmt.Errorf("unsupported message type: %s", message.MsgType)
	}
}

// SendMarkdown 发送Markdown格式消息
func (d *DingTalkBot) SendMarkdown(ctx context.Context, title, content string) error {
	message := &Message{
		Content: content,
		MsgType: MessageTypeMarkdown,
		Extra: map[string]interface{}{
			"title": title,
		},
	}
	return d.SendMessage(ctx, message)
}

// SendCard 发送卡片消息（钉钉不直接支持，转换为Markdown）
func (d *DingTalkBot) SendCard(ctx context.Context, card *Card) error {
	// 将卡片转换为Markdown格式
	markdown := d.cardToMarkdown(card)
	return d.SendMarkdown(ctx, card.Title, markdown)
}

// sendTextMessage 发送文本消息
func (d *DingTalkBot) sendTextMessage(ctx context.Context, message *Message) error {
	payload := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]interface{}{
			"content": message.Content,
		},
	}

	// 添加@信息
	if len(message.AtMobiles) > 0 || message.AtAll {
		at := map[string]interface{}{
			"atMobiles": message.AtMobiles,
			"isAtAll":   message.AtAll,
		}
		payload["at"] = at
	}

	return d.sendRequest(ctx, payload)
}

// sendMarkdownMessage 发送Markdown消息
func (d *DingTalkBot) sendMarkdownMessage(ctx context.Context, message *Message) error {
	title := "通知"
	if titleVal, ok := message.Extra["title"]; ok {
		if titleStr, ok := titleVal.(string); ok {
			title = titleStr
		}
	}

	payload := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]interface{}{
			"title": title,
			"text":  message.Content,
		},
	}

	// 添加@信息
	if len(message.AtMobiles) > 0 || message.AtAll {
		at := map[string]interface{}{
			"atMobiles": message.AtMobiles,
			"isAtAll":   message.AtAll,
		}
		payload["at"] = at
	}

	return d.sendRequest(ctx, payload)
}

// sendRequest 发送HTTP请求
func (d *DingTalkBot) sendRequest(ctx context.Context, payload map[string]interface{}) error {
	// 生成签名
	webhookURL, err := d.generateSignedURL()
	if err != nil {
		return fmt.Errorf("failed to generate signed URL: %v", err)
	}

	// 序列化payload
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := d.client.Do(req)
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
			return fmt.Errorf("dingtalk API error: code=%v, msg=%s", errCode, errMsg)
		}
	}

	d.logger.Infof("DingTalk message sent successfully")
	return nil
}

// generateSignedURL 生成带签名的URL
func (d *DingTalkBot) generateSignedURL() (string, error) {
	if d.secret == "" {
		return d.webhook, nil
	}

	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	stringToSign := timestamp + "\n" + d.secret

	h := hmac.New(sha256.New, []byte(d.secret))
	h.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	u, err := url.Parse(d.webhook)
	if err != nil {
		return "", err
	}

	query := u.Query()
	query.Set("timestamp", timestamp)
	query.Set("sign", signature)
	u.RawQuery = query.Encode()

	return u.String(), nil
}

// cardToMarkdown 将卡片转换为Markdown格式
func (d *DingTalkBot) cardToMarkdown(card *Card) string {
	var markdown strings.Builder

	// 标题
	if card.Title != "" {
		markdown.WriteString(fmt.Sprintf("## %s\n\n", card.Title))
	}

	// 描述
	if card.Description != "" {
		markdown.WriteString(fmt.Sprintf("%s\n\n", card.Description))
	}

	// 字段
	if len(card.Fields) > 0 {
		for _, field := range card.Fields {
			markdown.WriteString(fmt.Sprintf("**%s**: %s\n\n", field.Name, field.Value))
		}
	}

	// 链接
	if card.URL != "" {
		markdown.WriteString(fmt.Sprintf("[查看详情](%s)\n", card.URL))
	}

	return markdown.String()
}
