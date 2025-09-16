# 日K线数据按交易所分表设计

## 概述

为了优化日K线数据的存储和查询性能，系统现在按照交易所将日K线数据分为两个表：
- `daily_data_sh`: 上海证券交易所日K线数据
- `daily_data_sz`: 深圳证券交易所日K线数据

## 表结构

### daily_data_sh (上海证券交易所)
```sql
CREATE TABLE daily_data_sh (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    ts_code VARCHAR(20) NOT NULL,           -- 股票代码，如：600000.SH
    trade_date INT NOT NULL,                -- 交易日期，YYYYMMDD格式
    open DECIMAL(10,3) DEFAULT 0,           -- 开盘价
    high DECIMAL(10,3) DEFAULT 0,           -- 最高价
    low DECIMAL(10,3) DEFAULT 0,            -- 最低价
    close DECIMAL(10,3) DEFAULT 0,          -- 收盘价
    volume BIGINT DEFAULT 0,                -- 成交量
    amount DECIMAL(20,2) DEFAULT 0,         -- 成交额
    created_at BIGINT DEFAULT 0,            -- 创建时间戳
    
    INDEX idx_ts_code (ts_code),
    INDEX idx_trade_date (trade_date),
    INDEX idx_ts_code_trade_date (ts_code, trade_date),
    UNIQUE KEY uk_ts_code_trade_date (ts_code, trade_date)
);
```

### daily_data_sz (深圳证券交易所)
```sql
CREATE TABLE daily_data_sz (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    ts_code VARCHAR(20) NOT NULL,           -- 股票代码，如：000001.SZ
    trade_date INT NOT NULL,                -- 交易日期，YYYYMMDD格式
    open DECIMAL(10,3) DEFAULT 0,           -- 开盘价
    high DECIMAL(10,3) DEFAULT 0,           -- 最高价
    low DECIMAL(10,3) DEFAULT 0,            -- 最低价
    close DECIMAL(10,3) DEFAULT 0,          -- 收盘价
    volume BIGINT DEFAULT 0,                -- 成交量
    amount DECIMAL(20,2) DEFAULT 0,         -- 成交额
    created_at BIGINT DEFAULT 0,            -- 创建时间戳
    
    INDEX idx_ts_code (ts_code),
    INDEX idx_trade_date (trade_date),
    INDEX idx_ts_code_trade_date (ts_code, trade_date),
    UNIQUE KEY uk_ts_code_trade_date (ts_code, trade_date)
);
```

## 交易所识别规则

系统根据以下规则自动识别股票所属的交易所：

### 上海证券交易所 (SH)
- 股票代码以 `.SH` 结尾
- 股票代码以 `6` 开头（如：600000）

### 深圳证券交易所 (SZ)
- 股票代码以 `.SZ` 结尾
- 股票代码以 `0` 开头（主板，如：000001）
- 股票代码以 `3` 开头（创业板，如：300001）

## 核心组件

### 1. 数据模型
```

### 2. Repository层

#### DailyDataSHRepository (internal/repository/daily_data_sh_repository.go)
负责上海证券交易所日K线数据的数据库操作。

#### DailyDataSZRepository (internal/repository/daily_data_sz_repository.go)
负责深圳证券交易所日K线数据的数据库操作。

### 3. 服务层

#### DailyKLineManager (internal/service/daily_kline_manager.go)
统一的日K线数据管理器，提供以下功能：
- 根据股票代码自动选择对应的交易所表
- 统一的数据保存、查询、删除接口
- 跨交易所的统计信息

主要方法：
```go
// 保存数据到对应的交易所表
func (m *DailyKLineManager) SaveDailyData(data []model.DailyData) error

// 更新或插入数据
func (m *DailyKLineManager) UpsertDailyData(data []model.DailyData) error

// 获取指定股票的日K线数据
func (m *DailyKLineManager) GetDailyData(tsCode string, startDate, endDate time.Time, limit int) ([]model.DailyData, error)

// 获取最新的日K线数据
func (m *DailyKLineManager) GetLatestDailyData(tsCode string) (*model.DailyData, error)

// 获取数据统计信息
func (m *DailyKLineManager) GetAllExchangeStats() (map[string]interface{}, error)
```

## 数据迁移

### 1. 创建新表
执行SQL脚本创建新的交易所分表：
```bash
mysql -u username -p database_name < scripts/migrations/003_create_daily_data_exchange_tables.sql
```

### 2. 数据迁移
使用迁移工具将现有数据迁移到新表：
```bash
go run cmd/migrate-daily-data/main.go
```

或者使用SQL脚本直接迁移：
```bash
mysql -u username -p database_name < scripts/migrations/004_migrate_daily_data_to_exchange_tables.sql
```

### 3. 验证迁移结果
迁移完成后，系统会显示统计信息：
```json
{
  "sh_count": 12345,
  "sz_count": 23456,
  "total_count": 35801,
  "exchanges": ["SH", "SZ"]
}
```

## 使用示例

### 1. 在服务中使用DailyKLineManager
```go
// 创建管理器
dailyManager := service.NewDailyKLineManager(db, logger)

// 保存数据（自动分表）
err := dailyManager.UpsertDailyData(dailyDataList)

// 查询数据
data, err := dailyManager.GetDailyData("600000.SH", startDate, endDate, 100)

// 获取统计信息
stats, err := dailyManager.GetAllExchangeStats()
```

### 2. 在KLineService中的集成
KLineService已经集成了DailyKLineManager，所有原有的API保持不变：
```go
// 原有的方法调用方式不变
klineService := service.NewKLineService(db, logger, collectorManager)
data, err := klineService.GetKLineData("600000.SH", startDate, endDate)
```

## 性能优化

### 1. 查询性能
- 按交易所分表减少了单表数据量，提高查询效率
- 针对每个交易所的查询模式优化索引设计
- 减少跨表查询的复杂度

### 2. 存储优化
- 分表存储便于数据归档和清理
- 可以针对不同交易所设置不同的存储策略
- 支持分别备份和恢复

### 3. 维护优化
- 可以分别对不同交易所的表进行维护
- 支持独立的表结构优化
- 便于监控和统计各交易所的数据情况

## 注意事项

1. **向后兼容性**: 现有的API接口保持不变，业务代码无需修改
2. **数据一致性**: 迁移过程中确保数据完整性和一致性
3. **索引优化**: 根据实际查询模式调整索引策略
4. **监控告警**: 建议对新表设置监控和告警机制
5. **备份策略**: 更新备份脚本以包含新的分表结构

## 未来扩展

1. **更多交易所**: 可以轻松扩展支持其他交易所（如港股、美股）
2. **时间分片**: 可以进一步按时间维度分表（如按年分表）
3. **读写分离**: 可以为不同交易所配置不同的读写策略
4. **缓存策略**: 可以针对不同交易所设置不同的缓存策略