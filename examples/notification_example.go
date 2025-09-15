package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"stock/internal/notification"
	"stock/internal/utils"
)

func main() {
	// 创建logger
	logger := utils.NewLogger("notification-example", "info")

	// 创建通知配置
	config := &notification.Config{
		DingTalk: &notification.DingTalkConfig{
			Enabled: true,
			// 请替换为您的钉钉机器人Webhook URL
			Webhook: "https://oapi.dingtalk.com/robot/send?access_token=YOUR_ACCESS_TOKEN",
			// 如果启用了加签，请填入密钥
			Secret: "YOUR_SECRET_KEY",
		},
		WeWork: &notification.WeWorkConfig{
			Enabled: true,
			// 请替换为您的企微机器人Webhook URL
			Webhook: "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=YOUR_KEY",
		},
	}

	// 创建通知管理器
	factory := notification.NewFactory(logger)
	manager, err := factory.CreateManager(config)
	if err != nil {
		log.Fatalf("Failed to create notification manager: %v", err)
	}

	// 创建通知服务
	service := notification.NewService(manager, logger)

	ctx := context.Background()

	// 示例1: 发送简单文本消息
	fmt.Println("=== 示例1: 发送简单文本消息 ===")
	if err := service.SendTextMessage(ctx, "这是一条测试消息", nil, false); err != nil {
		log.Printf("发送文本消息失败: %v", err)
	} else {
		fmt.Println("文本消息发送成功")
	}

	time.Sleep(2 * time.Second)

	// 示例2: 发送Markdown消息
	fmt.Println("=== 示例2: 发送Markdown消息 ===")
	markdownContent := `## 📊 系统状态报告

**系统名称**: 股票选择系统
**运行状态**: 正常
**最后更新**: ` + time.Now().Format("2006-01-02 15:04:05") + `

### 关键指标
- **CPU使用率**: 45%
- **内存使用率**: 62%
- **磁盘使用率**: 78%

---
*这是一条自动生成的状态报告*`

	if err := service.SendMarkdownMessage(ctx, "系统状态报告", markdownContent); err != nil {
		log.Printf("发送Markdown消息失败: %v", err)
	} else {
		fmt.Println("Markdown消息发送成功")
	}

	time.Sleep(2 * time.Second)

	// 示例3: 发送股票提醒
	fmt.Println("=== 示例3: 发送股票提醒 ===")
	stockAlert := &notification.StockAlertTemplate{
		Stock: &MockStock{
			TSCode: "000001.SZ",
			Name:   "平安银行",
		},
		AlertType:   "买入信号",
		AlertReason: "技术指标MACD金叉，成交量放大",
		Timestamp:   time.Now(),
		Extra: map[string]interface{}{
			"当前价格": "10.50元",
			"涨跌幅":  "+2.35%",
			"成交量":  "1.2亿股",
			"建议操作": "分批买入",
		},
	}

	if err := service.SendStockAlert(ctx, stockAlert); err != nil {
		log.Printf("发送股票提醒失败: %v", err)
	} else {
		fmt.Println("股票提醒发送成功")
	}

	time.Sleep(2 * time.Second)

	// 示例4: 发送业绩报表通知
	fmt.Println("=== 示例4: 发送业绩报表通知 ===")
	performanceReport := &notification.PerformanceReportTemplate{
		Stock: &MockStock{
			TSCode: "000001.SZ",
			Name:   "平安银行",
		},
		Report: &MockPerformanceReport{
			ReportDate:    time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC),
			EPS:           1.18,
			BVPS:          15.23,
			WeightEPS:     1.15,
			Revenue:       69385000000, // 693.85亿元
			NetProfit:     24870000000, // 248.70亿元
			GrossMargin:   0,           // 银行业一般不计算毛利率
			RevenueYoY:    8.5,
			RevenueQoQ:    2.1,
			NetProfitYoY:  12.3,
			NetProfitQoQ:  -1.5,
			DividendYield: 3.2,
			NoticeDate:    time.Date(2024, 8, 20, 0, 0, 0, 0, time.UTC),
		},
	}

	if err := service.SendPerformanceReport(ctx, performanceReport); err != nil {
		log.Printf("发送业绩报表通知失败: %v", err)
	} else {
		fmt.Println("业绩报表通知发送成功")
	}

	time.Sleep(2 * time.Second)

	// 示例5: 发送系统通知
	fmt.Println("=== 示例5: 发送系统通知 ===")
	systemNotification := &notification.SystemNotificationTemplate{
		Title:     "数据同步完成",
		Message:   "今日股票数据同步已完成，共处理4523只股票的K线数据",
		Level:     "info",
		Timestamp: time.Now(),
		Extra: map[string]interface{}{
			"处理股票数量": 4523,
			"同步耗时":   "15分32秒",
			"数据大小":   "2.3GB",
			"错误数量":   0,
		},
	}

	if err := service.SendSystemNotification(ctx, systemNotification); err != nil {
		log.Printf("发送系统通知失败: %v", err)
	} else {
		fmt.Println("系统通知发送成功")
	}

	time.Sleep(2 * time.Second)

	// 示例6: 发送卡片消息（仅企微支持）
	fmt.Println("=== 示例6: 发送卡片消息 ===")
	card := &notification.Card{
		Title:       "股票推荐",
		Description: "基于技术分析的股票推荐",
		Color:       "#00FF00",
		URL:         "https://example.com/stock/000001",
		Fields: []notification.CardField{
			{Name: "股票代码", Value: "000001.SZ", Short: true},
			{Name: "股票名称", Value: "平安银行", Short: true},
			{Name: "推荐理由", Value: "技术指标良好，基本面稳健", Short: false},
			{Name: "目标价位", Value: "12.00元", Short: true},
			{Name: "风险等级", Value: "中等", Short: true},
		},
	}

	if err := service.SendCardMessage(ctx, card); err != nil {
		log.Printf("发送卡片消息失败: %v", err)
	} else {
		fmt.Println("卡片消息发送成功")
	}

	// 示例7: 批量发送股票提醒
	fmt.Println("=== 示例7: 批量发送股票提醒 ===")
	var alerts []*notification.StockAlertTemplate
	stocks := []struct {
		code, name, alertType, reason string
	}{
		{"000002.SZ", "万科A", "卖出信号", "技术指标MACD死叉"},
		{"600036.SH", "招商银行", "买入信号", "突破重要阻力位"},
		{"000858.SZ", "五粮液", "警告", "成交量异常放大"},
	}

	for _, stock := range stocks {
		alert := &notification.StockAlertTemplate{
			Stock: &MockStock{
				TSCode: stock.code,
				Name:   stock.name,
			},
			AlertType:   stock.alertType,
			AlertReason: stock.reason,
			Timestamp:   time.Now(),
		}
		alerts = append(alerts, alert)
	}

	if err := service.BatchSendStockAlerts(ctx, alerts); err != nil {
		log.Printf("批量发送股票提醒失败: %v", err)
	} else {
		fmt.Println("批量股票提醒发送成功")
	}

	fmt.Println("=== 所有示例执行完成 ===")
}

// MockStock 模拟股票结构
type MockStock struct {
	TSCode string
	Name   string
}

// MockPerformanceReport 模拟业绩报表结构
type MockPerformanceReport struct {
	ReportDate    time.Time
	EPS           float64
	BVPS          float64
	WeightEPS     float64
	Revenue       float64
	NetProfit     float64
	GrossMargin   float64
	RevenueYoY    float64
	RevenueQoQ    float64
	NetProfitYoY  float64
	NetProfitQoQ  float64
	DividendYield float64
	NoticeDate    time.Time
}
