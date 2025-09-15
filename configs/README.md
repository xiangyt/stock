# 配置文件说明

## 配置文件结构

本项目使用 YAML 格式的配置文件，主要配置文件为 `app.yaml`。

## 快速开始

1. 复制配置模板：
   ```bash
   cp configs/app.yaml.example configs/app.yaml
   ```

2. 根据你的环境修改 `app.yaml` 中的配置项

## 配置项说明

### 应用配置 (app)
- `name`: 应用名称
- `version`: 应用版本
- `env`: 运行环境 (development/production/test)
- `port`: 应用监听端口
- `debug`: 是否开启调试模式

### 数据库配置 (database)
- `driver`: 数据库驱动，目前支持 mysql
- `host`: 数据库服务器地址
- `port`: 数据库端口
- `name`: 数据库名称
- `user`: 数据库用户名
- `password`: 数据库密码
- `ssl_mode`: SSL模式
- `max_open_conns`: 最大打开连接数
- `max_idle_conns`: 最大空闲连接数
- `conn_max_lifetime`: 连接最大生存时间

### Redis配置 (redis)
- `host`: Redis服务器地址
- `port`: Redis端口
- `password`: Redis密码
- `db`: Redis数据库编号
- `pool_size`: 连接池大小
- `min_idle_conns`: 最小空闲连接数

### 日志配置 (log)
- `level`: 日志级别 (debug/info/warn/error)
- `format`: 日志格式 (json/text)
- `file`: 日志文件路径
- `max_size`: 单个日志文件最大大小(MB)
- `max_backups`: 保留的日志文件数量
- `max_age`: 日志文件保留天数
- `compress`: 是否压缩旧日志文件

### JWT配置 (jwt)
- `secret`: JWT密钥（生产环境必须修改）
- `expire_hours`: JWT过期时间(小时)
- `issuer`: JWT签发者

### 限流配置 (rate_limit)
- `enabled`: 是否启用限流
- `rps`: 每秒请求数限制
- `burst`: 突发请求数限制

### 监控配置 (metrics)
- `enabled`: 是否启用监控指标
- `port`: 监控端口
- `path`: 监控路径

### CORS配置 (cors)
- `allowed_origins`: 允许的源
- `allowed_methods`: 允许的HTTP方法
- `allowed_headers`: 允许的请求头
- `exposed_headers`: 暴露的响应头
- `allow_credentials`: 是否允许携带凭证
- `max_age`: 预检请求缓存时间

## 环境变量

可以通过环境变量覆盖配置文件中的设置，环境变量格式为：`STOCK_<SECTION>_<KEY>`

例如：
```bash
export STOCK_DATABASE_HOST=localhost
export STOCK_DATABASE_PASSWORD=your_password
export STOCK_JWT_SECRET=your_secret_key
export STOCK_APP_PORT=8080
```

## 安全建议

1. **生产环境配置**：
   - 修改默认的JWT密钥
   - 使用强密码
   - 限制CORS允许的源
   - 关闭调试模式

2. **敏感信息**：
   - 数据库密码建议使用环境变量
   - JWT密钥建议使用环境变量
   - 不要将包含敏感信息的配置文件提交到版本控制

3. **网络安全**：
   - 生产环境建议使用SSL/TLS
   - 配置防火墙规则
   - 使用VPN或内网访问数据库

## 配置验证

可以使用以下命令验证配置是否正确：

```bash
# 验证配置文件语法
go run -ldflags "-X main.configFile=configs/app.yaml" cmd/cli/main.go -cmd validate-config

# 测试数据库连接
go run -ldflags "-X main.configFile=configs/app.yaml" cmd/cli/main.go -cmd test-db