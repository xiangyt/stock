# 智能选股系统 (Smart Stock Selection System)

基于Go语言开发的高性能智能选股系统，集成技术分析、基本面分析和多种选股策略。

## 项目架构

```
stock/
├── README.md                    # 项目说明文档
├── go.mod                       # Go模块文件
├── go.sum                       # 依赖校验文件
├── Makefile                     # 构建脚本
├── Dockerfile                   # Docker构建文件
├── docker-compose.yml           # Docker编排文件
├── .env.example                 # 环境变量示例
├── .gitignore                   # Git忽略文件
│
├── cmd/                         # 应用程序入口
│   ├── server/                  # Web服务器
│   │   └── main.go
│   ├── cli/                     # 命令行工具
│   │   └── main.go
│   └── worker/                  # 后台任务处理器
│       └── main.go
│
├── internal/                    # 内部包（不对外暴露）
│   ├── config/                  # 配置管理
│   │   ├── config.go
│   │   └── database.go
│   │
│   ├── model/                   # 数据模型
│   │   ├── stock.go
│   │   ├── daily_data.go
│   │   ├── financial_data.go
│   │   ├── technical_indicator.go
│   │   └── selection_result.go
│   │
│   ├── repository/              # 数据访问层
│   │   ├── interface.go
│   │   ├── stock_repo.go
│   │   ├── daily_data_repo.go
│   │   ├── financial_repo.go
│   │   └── selection_repo.go
│   │
│   ├── service/                 # 业务逻辑层
│   │   ├── data_collector.go
│   │   ├── technical_analyzer.go
│   │   ├── fundamental_analyzer.go
│   │   ├── strategy_engine.go
│   │   ├── backtest_engine.go
│   │   └── risk_manager.go
│   │
│   ├── handler/                 # HTTP处理器
│   │   ├── stock_handler.go
│   │   ├── analysis_handler.go
│   │   ├── strategy_handler.go
│   │   └── middleware.go
│   │
│   ├── collector/               # 数据采集器
│   │   ├── interface.go
│   │   ├── tushare_collector.go
│   │   ├── akshare_collector.go
│   │   └── yahoo_collector.go
│   │
│   ├── strategy/                # 选股策略
│   │   ├── interface.go
│   │   ├── base_strategy.go
│   │   ├── technical_strategy.go
│   │   ├── fundamental_strategy.go
│   │   └── combined_strategy.go
│   │
│   ├── indicator/               # 技术指标计算
│   │   ├── ma.go
│   │   ├── rsi.go
│   │   ├── macd.go
│   │   ├── kdj.go
│   │   └── bollinger.go
│   │
│   └── utils/                   # 工具包
│       ├── logger.go
│       ├── database.go
│       ├── http_client.go
│       ├── math.go
│       └── time.go
│
├── pkg/                         # 公共包（可对外暴露）
│   ├── errors/                  # 错误定义
│   │   └── errors.go
│   ├── response/                # HTTP响应格式
│   │   └── response.go
│   └── validator/               # 数据验证
│       └── validator.go
│
├── api/                         # API定义
│   ├── openapi.yaml            # OpenAPI规范
│   └── proto/                  # gRPC协议定义（可选）
│       └── stock.proto
│
├── web/                         # 前端资源
│   ├── static/                 # 静态文件
│   │   ├── css/
│   │   ├── js/
│   │   └── images/
│   └── templates/              # HTML模板
│       ├── index.html
│       ├── analysis.html
│       └── strategy.html
│
├── scripts/                     # 脚本文件
│   ├── init_db.sql             # 数据库初始化脚本
│   ├── migrate.sh              # 数据库迁移脚本
│   └── deploy.sh               # 部署脚本
│
├── configs/                     # 配置文件
│   ├── app.yaml                # 应用配置
│   ├── database.yaml           # 数据库配置
│   └── strategy.yaml           # 策略配置
│
├── docs/                        # 文档
│   ├── api.md                  # API文档
│   ├── deployment.md           # 部署文档
│   └── development.md          # 开发文档
│
├── test/                        # 测试文件
│   ├── integration/            # 集成测试
│   ├── unit/                   # 单元测试
│   └── testdata/               # 测试数据
│
└── vendor/                      # 依赖包（可选，使用go mod vendor生成）
```

## 核心模块说明

### 1. **应用入口** (`cmd/`)
- **server**: Web API服务器
- **cli**: 命令行工具（数据更新、选股执行等）
- **worker**: 后台任务处理（定时数据更新、策略执行）

### 2. **内部包** (`internal/`)
- **config**: 配置管理，支持YAML、环境变量
- **model**: 数据模型定义，对应数据库表结构
- **repository**: 数据访问层，封装数据库操作
- **service**: 业务逻辑层，核心业务处理
- **handler**: HTTP处理器，处理Web请求
- **collector**: 数据采集器，支持多数据源
- **strategy**: 选股策略实现
- **indicator**: 技术指标计算
- **utils**: 通用工具函数

### 3. **公共包** (`pkg/`)
- **errors**: 统一错误定义
- **response**: HTTP响应格式标准化
- **validator**: 数据验证工具

### 4. **API定义** (`api/`)
- **openapi.yaml**: RESTful API规范
- **proto**: gRPC服务定义（可选）

### 5. **前端资源** (`web/`)
- **static**: CSS、JS、图片等静态资源
- **templates**: HTML模板文件

## 技术栈

### 后端框架
- **Web框架**: Gin/Echo/Fiber
- **ORM**: GORM
- **数据库**: PostgreSQL/MySQL/SQLite
- **缓存**: Redis
- **消息队列**: RabbitMQ/Kafka（可选）

### 数据处理
- **HTTP客户端**: Resty
- **JSON处理**: encoding/json
- **时间处理**: time包
- **数学计算**: math包 + 自定义指标算法

### 工具库
- **配置管理**: Viper
- **日志**: Logrus/Zap
- **验证**: go-playground/validator
- **测试**: Testify
- **文档**: Swagger

### 部署工具
- **容器化**: Docker + Docker Compose
- **构建**: Makefile
- **CI/CD**: GitHub Actions（可选）

## 项目特点

### 1. **高性能**
- Go语言原生并发支持
- 高效的内存管理
- 快速的编译和执行

### 2. **模块化设计**
- 清晰的分层架构
- 接口驱动开发
- 易于测试和维护

### 3. **可扩展性**
- 插件化的数据源
- 策略模式的选股算法
- 微服务架构支持

### 4. **生产就绪**
- 完整的错误处理
- 日志和监控
- Docker化部署

## 快速开始

```bash
# 1. 初始化项目
go mod init stock

# 2. 安装依赖
go mod tidy

# 3. 初始化数据库
make init-db

# 4. 启动服务
make run-server

# 5. 启动Web界面
open http://localhost:8080
```

## 开发命令

```bash
# 构建
make build

# 测试
make test

# 代码检查
make lint

# 生成文档
make docs

# Docker构建
make docker-build

# 部署
make deploy