package notification

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"stock/internal/model"
)

// StockAlertTemplate 股票提醒模板
type StockAlertTemplate struct {
	Stock       *model.Stock
	AlertType   string
	AlertReason string
	Timestamp   time.Time
	Extra       map[string]interface{}
}

// ToMessage 转换为通用消息
func (t *StockAlertTemplate) ToMessage() *Message {
	content := t.generateTextContent()
	return &Message{
		Content: content,
		MsgType: MessageTypeText,
	}
}

// ToMarkdown 转换为Markdown消息
func (t *StockAlertTemplate) ToMarkdown() (string, string) {
	title := fmt.Sprintf("股票提醒 - %s", t.AlertType)
	content := t.generateMarkdownContent()
	return title, content
}

// ToCard 转换为卡片消息
func (t *StockAlertTemplate) ToCard() *Card {
	return &Card{
		Title:       fmt.Sprintf("股票提醒 - %s", t.AlertType),
		Description: t.AlertReason,
		Color:       t.getAlertColor(),
		Fields:      t.generateCardFields(),
	}
}

// generateTextContent 生成文本内容
func (t *StockAlertTemplate) generateTextContent() string {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("📈 股票提醒 - %s\n", t.AlertType))
	content.WriteString(fmt.Sprintf("股票代码: %s\n", t.Stock.TsCode))
	content.WriteString(fmt.Sprintf("股票名称: %s\n", t.Stock.Name))
	content.WriteString(fmt.Sprintf("提醒原因: %s\n", t.AlertReason))
	content.WriteString(fmt.Sprintf("提醒时间: %s\n", t.Timestamp.Format("2006-01-02 15:04:05")))

	// 添加额外信息
	if t.Extra != nil {
		for key, value := range t.Extra {
			content.WriteString(fmt.Sprintf("%s: %v\n", key, value))
		}
	}

	return content.String()
}

// generateMarkdownContent 生成Markdown内容
func (t *StockAlertTemplate) generateMarkdownContent() string {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("## 📈 股票提醒 - %s\n\n", t.AlertType))
	content.WriteString(fmt.Sprintf("**股票代码**: %s\n\n", t.Stock.TsCode))
	content.WriteString(fmt.Sprintf("**股票名称**: %s\n\n", t.Stock.Name))
	content.WriteString(fmt.Sprintf("**提醒原因**: %s\n\n", t.AlertReason))
	content.WriteString(fmt.Sprintf("**提醒时间**: %s\n\n", t.Timestamp.Format("2006-01-02 15:04:05")))

	// 添加额外信息
	if t.Extra != nil && len(t.Extra) > 0 {
		content.WriteString("### 详细信息\n\n")
		for key, value := range t.Extra {
			content.WriteString(fmt.Sprintf("- **%s**: %v\n", key, value))
		}
		content.WriteString("\n")
	}

	return content.String()
}

// generateCardFields 生成卡片字段
func (t *StockAlertTemplate) generateCardFields() []CardField {
	fields := []CardField{
		{Name: "股票代码", Value: t.Stock.TsCode, Short: true},
		{Name: "股票名称", Value: t.Stock.Name, Short: true},
		{Name: "提醒原因", Value: t.AlertReason, Short: false},
		{Name: "提醒时间", Value: t.Timestamp.Format("2006-01-02 15:04:05"), Short: true},
	}

	// 添加额外字段
	if t.Extra != nil {
		for key, value := range t.Extra {
			fields = append(fields, CardField{
				Name:  key,
				Value: fmt.Sprintf("%v", value),
				Short: true,
			})
		}
	}

	return fields
}

// getAlertColor 获取提醒颜色
func (t *StockAlertTemplate) getAlertColor() string {
	switch strings.ToLower(t.AlertType) {
	case "买入", "buy":
		return "#00FF00" // 绿色
	case "卖出", "sell":
		return "#FF0000" // 红色
	case "警告", "warning":
		return "#FFA500" // 橙色
	default:
		return "#0000FF" // 蓝色
	}
}

// PerformanceReportTemplate 业绩报表模板
type PerformanceReportTemplate struct {
	Stock  *model.Stock
	Report *model.PerformanceReport
}

// ToMarkdown 转换为Markdown消息
func (t *PerformanceReportTemplate) ToMarkdown() (string, string) {
	title := fmt.Sprintf("业绩报表 - %s(%s)", t.Stock.Name, t.Stock.TsCode)
	content := t.generateMarkdownContent()
	return title, content
}

// generateMarkdownContent 生成Markdown内容
func (t *PerformanceReportTemplate) generateMarkdownContent() string {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("## 📊 业绩报表 - %s(%s)\n\n", t.Stock.Name, t.Stock.TsCode))
	content.WriteString(fmt.Sprintf("**报告期**: %s\n\n", strconv.Itoa(t.Report.ReportDate)))

	content.WriteString("### 每股指标\n")
	content.WriteString(fmt.Sprintf("- **每股收益(EPS)**: %.4f元\n", t.Report.EPS))
	content.WriteString(fmt.Sprintf("- **每股净资产(BVPS)**: %.4f元\n", t.Report.BVPS))
	content.WriteString(fmt.Sprintf("- **加权每股收益**: %.4f元\n\n", t.Report.WeightEPS))

	content.WriteString("### 经营指标\n")
	content.WriteString(fmt.Sprintf("- **营业总收入**: %.2f亿元\n", t.Report.Revenue/100000000))
	content.WriteString(fmt.Sprintf("- **净利润**: %.2f亿元\n", t.Report.NetProfit/100000000))
	content.WriteString(fmt.Sprintf("- **销售毛利率**: %.2f%%\n\n", t.Report.GrossMargin))

	content.WriteString("### 增长指标\n")
	content.WriteString(fmt.Sprintf("- **营收同比增长**: %.2f%%\n", t.Report.RevenueYoY))
	content.WriteString(fmt.Sprintf("- **营收环比增长**: %.2f%%\n", t.Report.RevenueQoQ))
	content.WriteString(fmt.Sprintf("- **净利润同比增长**: %.2f%%\n", t.Report.NetProfitYoY))
	content.WriteString(fmt.Sprintf("- **净利润环比增长**: %.2f%%\n\n", t.Report.NetProfitQoQ))

	if t.Report.DividendYield > 0 {
		content.WriteString("### 分红信息\n")
		content.WriteString(fmt.Sprintf("- **股息率**: %.2f%%\n\n", t.Report.DividendYield))
	}

	if t.Report.LatestAnnouncementDate != nil {
		content.WriteString(fmt.Sprintf("**公告日期**: %s\n", t.Report.LatestAnnouncementDate.Format("2006-01-02")))
	}

	return content.String()
}

// SystemNotificationTemplate 系统通知模板
type SystemNotificationTemplate struct {
	Title     string
	Message   string
	Level     string // info, warning, error
	Timestamp time.Time
	Extra     map[string]interface{}
}

// ToMessage 转换为通用消息
func (t *SystemNotificationTemplate) ToMessage() *Message {
	content := fmt.Sprintf("[%s] %s\n%s\n时间: %s",
		strings.ToUpper(t.Level),
		t.Title,
		t.Message,
		t.Timestamp.Format("2006-01-02 15:04:05"))

	return &Message{
		Content: content,
		MsgType: MessageTypeText,
	}
}

// ToMarkdown 转换为Markdown消息
func (t *SystemNotificationTemplate) ToMarkdown() (string, string) {
	title := fmt.Sprintf("系统通知 - %s", t.Title)

	var content strings.Builder
	content.WriteString(fmt.Sprintf("## %s 系统通知\n\n", t.getLevelEmoji()))
	content.WriteString(fmt.Sprintf("**标题**: %s\n\n", t.Title))
	content.WriteString(fmt.Sprintf("**消息**: %s\n\n", t.Message))
	content.WriteString(fmt.Sprintf("**级别**: %s\n\n", strings.ToUpper(t.Level)))
	content.WriteString(fmt.Sprintf("**时间**: %s\n\n", t.Timestamp.Format("2006-01-02 15:04:05")))

	if t.Extra != nil && len(t.Extra) > 0 {
		content.WriteString("### 详细信息\n\n")
		for key, value := range t.Extra {
			content.WriteString(fmt.Sprintf("- **%s**: %v\n", key, value))
		}
	}

	return title, content.String()
}

// getLevelEmoji 获取级别对应的emoji
func (t *SystemNotificationTemplate) getLevelEmoji() string {
	switch strings.ToLower(t.Level) {
	case "info":
		return "ℹ️"
	case "warning":
		return "⚠️"
	case "error":
		return "❌"
	default:
		return "📢"
	}
}
