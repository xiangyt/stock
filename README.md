# 智能选股系统

一个基于Go语言开发的智能选股系统，支持多数据源采集、策略分析和自动化通知。

## 🎯 系统概述

智能选股系统是一个全流程的股票投资辅助平台，从数据采集、策略分析到结果通知，为投资者提供完整的技术支持。

## 🏗️ 系统架构

### 核心模块

#### 1. 数据采集管理模块
- **数据源配置**: 支持东方财富、同花顺等多个数据源
- **采集任务调度**: 自动化数据采集任务管理
- **数据质量监控**: 实时监控数据完整性和准确性
- **历史数据管理**: 海量历史数据存储和查询

#### 2. 策略配置模块
- **选股策略编辑器**: 可视化策略配置界面
- **策略参数配置**: 灵活的参数调整机制
- **策略回测功能**: 历史数据验证策略有效性
- **策略性能分析**: 多维度策略效果评估

#### 3. 通知机器人配置模块
- **钉钉/企微机器人管理**: 多平台消息推送支持
- **消息模板配置**: 自定义通知内容格式
- **通知规则设置**: 智能化通知触发条件
- **发送记录查询**: 完整的消息发送历史

#### 4. 用户管理模块
- **用户认证授权**: 安全的登录和权限控制
- **角色权限管理**: 细粒度的功能权限分配
- **个人偏好设置**: 个性化用户体验
- **操作日志记录**: 完整的用户行为追踪

#### 5. 股票数据管理模块
- **股票基础信息管理**: 完整的股票档案信息
- **K线数据查询与展示**: 多周期K线图表分析
- **财务数据分析**: 深度财务指标挖掘
- **实时行情监控**: 实时价格和成交数据
- **技术指标计算**: 丰富的技术分析工具

#### 6. 选股结果管理模块
- **选股结果展示**: 直观的选股结果呈现
- **股票池管理**: 多维度股票分组管理
- **关注列表功能**: 个性化股票关注
- **历史选股记录**: 完整的选股历史追踪
- **结果导出功能**: 多格式数据导出支持

#### 7. 风险控制模块
- **风险指标监控**: 实时风险评估
- **止损止盈设置**: 智能化风险控制
- **仓位管理建议**: 科学的资金配置
- **风险预警机制**: 及时的风险提醒

#### 8. 报表分析模块
- **策略收益统计**: 详细的收益分析报告
- **选股成功率分析**: 策略有效性评估
- **市场热点分析**: 市场趋势洞察
- **自定义报表生成**: 灵活的报表定制

#### 9. 系统监控模块
- **系统性能监控**: 实时系统状态监控
- **数据采集状态**: 采集任务执行情况
- **API调用统计**: 接口使用情况分析
- **错误日志管理**: 系统异常处理和追踪

#### 10. 投资组合模块
- **模拟投资组合**: 虚拟投资组合管理
- **收益跟踪**: 投资收益实时跟踪
- **资产配置建议**: 智能化资产配置
- **投资记录管理**: 完整的投资历史记录

## 🎨 前端页面结构

```
智能选股系统前端
├── 登录/注册页面
├── 仪表板 (Dashboard)
│   ├── 系统概览
│   ├── 今日选股结果
│   └── 关键指标监控
├── 股票管理
│   ├── 股票列表
│   ├── 股票详情
│   └── K线图表
├── 策略中心
│   ├── 策略列表
│   ├── 策略编辑器
│   ├── 回测分析
│   └── 策略性能
├── 选股结果
│   ├── 今日选股
│   ├── 历史记录
│   └── 股票池管理
├── 数据管理
│   ├── 数据源配置
│   ├── 采集任务
│   └── 数据质量
├── 通知配置
│   ├── 机器人管理
│   ├── 消息模板
│   └── 发送记录
├── 系统设置
│   ├── 用户管理
│   ├── 权限配置
│   └── 系统监控
└── 个人中心
    ├── 个人信息
    ├── 偏好设置
    └── 操作日志
```

## 🔧 技术栈

- **后端**: Go 1.21+, Gin, GORM, MySQL
- **前端**: Vue 3, Element Plus, ECharts, Vite
- **数据库**: MySQL 8.0+
- **缓存**: Redis (可选)
- **消息队列**: 支持多种MQ (可选)
- **部署**: Docker, Docker Compose

## 🚀 快速开始

### 环境要求

- Go 1.21+
- MySQL 8.0+
- Node.js 16+

### 安装步骤

1. 克隆项目
```bash
git clone <repository-url>
cd stock
```

2. 配置数据库
```bash
# 创建数据库
mysql -u root -p -e "CREATE DATABASE stock CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 导入表结构
mysql -u root -p stock < scripts/create_tables.sql
```

3. 配置文件
```bash
cp configs/app.yaml.example configs/app.yaml
# 编辑配置文件，修改数据库连接信息
```

4. 启动后端服务
```bash
go mod tidy
go run cmd/web/main.go
```

5. 启动前端服务
```bash
cd frontend
npm install
npm run dev
```

### 访问地址

- 前端界面: http://localhost:3000
- 后端API: http://localhost:8080
- API文档: http://localhost:8080/swagger/index.html

## 📁 项目结构

```
stock/
├── cmd/                    # 应用程序入口
│   ├── api/               # API服务
│   ├── cli/               # 命令行工具
│   ├── web/               # Web服务
│   └── worker/            # 后台任务
├── internal/              # 内部包
│   ├── api/               # API处理器
│   ├── collector/         # 数据采集器
│   ├── config/            # 配置管理
│   ├── database/          # 数据库连接
│   ├── handler/           # 业务处理器
│   ├── model/             # 数据模型
│   ├── repository/        # 数据访问层
│   ├── service/           # 业务逻辑层
│   └── utils/             # 工具函数
├── frontend/              # 前端项目
├── configs/               # 配置文件
├── scripts/               # 脚本文件
├── docs/                  # 文档
└── test/                  # 测试文件
```

## 🔌 API接口设计

### RESTful API 结构
```
/api/v1/
├── auth/           # 认证相关
│   ├── POST /login
│   ├── POST /logout
│   └── POST /refresh
├── users/          # 用户管理
│   ├── GET /users
│   ├── POST /users
│   ├── PUT /users/:id
│   └── DELETE /users/:id
├── stocks/         # 股票数据
│   ├── GET /stocks
│   ├── GET /stocks/:code
│   └── GET /stocks/:code/kline
├── strategies/     # 策略管理
│   ├── GET /strategies
│   ├── POST /strategies
│   ├── PUT /strategies/:id
│   └── DELETE /strategies/:id
├── selections/     # 选股结果
│   ├── GET /selections
│   ├── POST /selections
│   └── GET /selections/history
├── collectors/     # 数据采集
│   ├── GET /collectors
│   ├── POST /collectors/sync
│   └── GET /collectors/status
├── notifications/  # 通知管理
│   ├── GET /robots
│   ├── POST /robots
│   ├── PUT /robots/:id
│   └── POST /notifications/send
├── portfolios/     # 投资组合
│   ├── GET /portfolios
│   ├── POST /portfolios
│   └── GET /portfolios/:id/performance
├── reports/        # 报表分析
│   ├── GET /reports/strategy-performance
│   ├── GET /reports/market-analysis
│   └── GET /reports/custom
├── monitoring/     # 系统监控
│   ├── GET /monitoring/system
│   ├── GET /monitoring/api-stats
│   └── GET /monitoring/logs
└── settings/       # 系统设置
    ├── GET /settings
    └── PUT /settings
```

## 🗄️ 数据库设计

### 核心数据表
```sql
-- 用户相关
users, roles, permissions, user_roles

-- 股票数据
stocks, daily_data, kline_data, financial_data

-- 策略相关  
strategies, strategy_params, strategy_results, backtests

-- 选股相关
stock_selections, stock_pools, watchlists

-- 投资组合
portfolios, portfolio_stocks, transactions

-- 通知机器人
robot_configs, notification_logs

-- 系统监控
system_logs, api_logs, performance_metrics
```

## 🔧 配置说明

### 数据库配置
```yaml
database:
  host: localhost
  port: 3306
  username: root
  password: your_password
  database: stock
```

### 数据源配置
```yaml
collectors:
  eastmoney:
    enabled: true
    rate_limit: 100
  tushare:
    enabled: false
    token: your_token
```

## 📚 开发指南

### 添加新的数据采集器

1. 实现 `DataCollector` 接口
2. 在 `collector` 包中注册
3. 添加配置项

### 添加新的选股策略

1. 实现 `Strategy` 接口
2. 在 `strategy` 包中注册
3. 添加参数配置

## 🚢 部署

### Docker部署
```bash
docker-compose up -d
```

### 手动部署
```bash
# 构建后端
go build -o stock cmd/web/main.go

# 构建前端
cd frontend && npm run build

# 启动服务
./stock
```

## 🤝 贡献

欢迎提交Issue和Pull Request来改进项目。

## 📄 许可证

MIT License