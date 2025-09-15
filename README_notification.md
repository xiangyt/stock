# 机器人通知功能

本项目集成了钉钉机器人和企微机器人通知功能，支持发送文本、Markdown和卡片消息。

## 功能特性

- ✅ 支持钉钉机器人和企微机器人
- ✅ 支持文本、Markdown、卡片消息格式
- ✅ 支持@指定用户和@所有人
- ✅ 支持批量发送和定时通知
- ✅ 提供丰富的消息模板
- ✅ 完整的错误处理和日志记录
- ✅ RESTful API接口

## 快速开始

### 1. 配置机器人

#### 钉钉机器人配置

1. 在钉钉群中添加自定义机器人
2. 复制Webhook地址
3. 如果启用了加签，获取密钥

#### 企微机器人配置

1. 在企微群中添加群机器人
2. 复制Webhook地址

### 2. 配置文件

创建 `configs/notification.yaml` 文件：

```yaml
notification:
  dingtalk:
    enabled: true
    webhook: "https://oapi.dingtalk.com/robot/send?access_token=YOUR_ACCESS_TOKEN"
    secret: "YOUR_SECRET_KEY"  # 可选
  
  wework:
    enabled: true
    webhook: "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=YOUR_KEY"
```

### 3. 代码示例

```go
package main

import (
    "context"
    "stock/internal/notification"
    "stock/internal/utils"
)

func main() {
    // 创建logger
    logger := utils.NewLogger("notification", "info")
    
    // 创建配置
    config := &notification.Config{
        DingTalk: &notification.DingTalkConfig{
            Enabled: true,
            Webhook: "YOUR_WEBHOOK_URL",
            Secret:  "YOUR_SECRET",
        },
    }
    
    // 创建通知管理器
    factory := notification.NewFactory(logger)
    manager, _ := factory.CreateManager(config)
    
    // 创建通知服务
    service := notification.NewService(manager, logger)
    
    // 发送消息
    ctx := context.Background()
    service.SendTextMessage(ctx, "Hello, World!", nil, false)
}
```

## API接口

### 发送文本消息

```bash
POST /api/v1/notification/text
Content-Type: application/json

{
    "content": "这是一条测试消息",
    "atMobiles": ["13800138000"],
    "atAll": false,
    "botType": "dingtalk"  // 可选，不指定则发送到所有机器人
}
```

### 发送Markdown消息

```bash
POST /api/v1/notification/markdown
Content-Type: application/json

{
    "title": "系统通知",
    "content": "## 标题\n\n**内容**"
}
```

### 发送股票提醒

```bash
POST /api/v1/notification/stock-alert
Content-Type: application/json

{
    "tsCode": "000001.SZ",
    "stockName": "平安银行",
    "alertType": "买入信号",
    "alertReason": "技术指标突破",
    "extra": {
        "价格": "10.50",
        "涨幅": "2.35%"
    }
}
```

### 发送系统通知

```bash
POST /api/v1/notification/system
Content-Type: application/json

{
    "title": "系统维护通知",
    "message": "系统将于今晚进行维护",
    "level": "warning",
    "extra": {
        "维护时间": "22:00-24:00"
    }
}
```

### 获取机器人状态

```bash
GET /api/v1/notification/status
```

### 发送健康检查

```bash
POST /api/v1/notification/health-check
```

## 消息模板

### 股票提醒模板

```go
template := &notification.StockAlertTemplate{
    Stock: &model.Stock{
        TSCode: "000001.SZ",
        Name:   "平安银行",
    },
    AlertType:   "买入信号",
    AlertReason: "技术指标MACD金叉",
    Timestamp:   time.Now(),
    Extra: map[string]interface{}{
        "价格": "10.50",
        "涨幅": "2.35%",
    },
}

service.SendStockAlert(ctx, template)
```

### 业绩报表模板

```go
template := &notification.PerformanceReportTemplate{
    Stock:  stock,
    Report: performanceReport,
}

service.SendPerformanceReport(ctx, template)
```

### 系统通知模板

```go
template := &notification.SystemNotificationTemplate{
    Title:     "系统通知",
    Message:   "数据同步完成",
    Level:     "info",
    Timestamp: time.Now(),
}

service.SendSystemNotification(ctx, template)
```

## 命令行工具

### 测试模式

```bash
go run cmd/notification/main.go -test -message "测试消息" -bot dingtalk
```

### 发送健康检查

```bash
go run cmd/notification/main.go -config configs/notification.yaml
```

## 错误处理

所有API都会返回统一的错误格式：

```json
{
    "success": false,
    "error": {
        "code": "NOTIFICATION_ERROR",
        "message": "发送消息失败",
        "details": "具体错误信息"
    }
}
```

## 注意事项

1. **频率限制**: 钉钉和企微都有频率限制，建议在批量发送时添加延迟
2. **消息长度**: 注意消息内容长度限制
3. **安全性**: 妥善保管Webhook URL和密钥
4. **测试**: 建议先在测试群中验证功能

## 扩展功能

- 支持更多机器人类型（飞书、Slack等）
- 定时任务和计划通知
- 消息模板管理
- 通知历史记录
- 消息发送统计

## 故障排除

### 常见问题

1. **消息发送失败**
   - 检查Webhook URL是否正确
   - 检查网络连接
   - 查看日志获取详细错误信息

2. **签名验证失败**
   - 检查密钥是否正确
   - 确认时间同步

3. **@功能不生效**
   - 确认手机号格式正确
   - 检查机器人权限设置

### 日志查看

```bash
# 查看通知相关日志
grep "notification" logs/app.log

# 查看错误日志
grep "ERROR" logs/app.log | grep "notification"