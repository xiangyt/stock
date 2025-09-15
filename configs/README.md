# 配置文件说明

本目录包含项目的配置文件模板和说明。

## 配置文件列表

### 通知配置
- `notification.yaml.example` - 通知系统配置模板
- `notification.yaml` - 实际配置文件（包含敏感信息，已被Git忽略）

## 使用方法

### 1. 复制配置模板

```bash
# 复制通知配置模板
cp configs/notification.yaml.example configs/notification.yaml
```

### 2. 编辑配置文件

编辑 `configs/notification.yaml` 文件，填入真实的配置信息：

```yaml
notification:
  dingtalk:
    enabled: true
    webhook: "https://oapi.dingtalk.com/robot/send?access_token=YOUR_REAL_TOKEN"
    secret: "YOUR_REAL_SECRET"
  
  wework:
    enabled: true
    webhook: "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=YOUR_REAL_KEY"
```

### 3. 获取配置信息

#### 钉钉机器人配置
1. 在钉钉群中点击群设置 → 智能群助手 → 添加机器人
2. 选择"自定义"机器人
3. 设置机器人名称和头像
4. 选择安全设置（推荐使用"加签"方式）
5. 复制Webhook地址和密钥

#### 企微机器人配置
1. 在企微群中点击群设置 → 群机器人 → 添加机器人
2. 选择"Webhook"类型
3. 设置机器人名称
4. 复制Webhook地址

## 安全注意事项

⚠️ **重要安全提醒**

1. **不要提交敏感信息到Git**
   - `configs/notification.yaml` 已被添加到 `.gitignore`
   - 只提交 `.example` 模板文件

2. **保护配置文件**
   - 设置适当的文件权限：`chmod 600 configs/notification.yaml`
   - 不要在日志中输出敏感信息

3. **生产环境建议**
   - 使用环境变量存储敏感配置
   - 使用密钥管理服务（如AWS Secrets Manager、Azure Key Vault等）
   - 定期轮换密钥和Token

## 环境变量配置（推荐）

除了配置文件，您也可以使用环境变量：

```bash
export DINGTALK_WEBHOOK="https://oapi.dingtalk.com/robot/send?access_token=YOUR_TOKEN"
export DINGTALK_SECRET="YOUR_SECRET"
export WEWORK_WEBHOOK="https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=YOUR_KEY"
```

程序会优先使用环境变量，如果环境变量不存在则使用配置文件。

## 配置验证

使用以下命令验证配置是否正确：

```bash
# 测试钉钉机器人
go run test_notification.go -type dingtalk

# 测试企微机器人  
go run test_notification.go -type wework

# 测试所有机器人
go run test_notification.go -type all