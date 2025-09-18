-- 更新K线数据表的主键结构
-- 将自增ID主键改为股票代码+交易日期的联合主键
-- 执行时间：2025-09-17

-- 1. 备份现有数据（可选，建议在生产环境执行）
-- CREATE TABLE daily_data_sh_backup AS SELECT * FROM daily_data_sh;
-- CREATE TABLE daily_data_sz_backup AS SELECT * FROM daily_data_sz;
-- CREATE TABLE weekly_data_backup AS SELECT * FROM weekly_data;
-- CREATE TABLE monthly_data_backup AS SELECT * FROM monthly_data;
-- CREATE TABLE yearly_data_backup AS SELECT * FROM yearly_data;

-- 2. 更新日K线数据表（上海交易所）
-- 删除现有主键和自增ID列
ALTER TABLE daily_data_sh DROP PRIMARY KEY;
ALTER TABLE daily_data_sh DROP COLUMN id;

-- 添加新的联合主键
ALTER TABLE daily_data_sh ADD PRIMARY KEY (ts_code, trade_date);

-- 添加更新时间字段
ALTER TABLE daily_data_sh ADD COLUMN updated_at BIGINT DEFAULT 0 COMMENT '记录更新时间戳';

-- 3. 更新日K线数据表（深圳交易所）
-- 删除现有主键和自增ID列
ALTER TABLE daily_data_sz DROP PRIMARY KEY;
ALTER TABLE daily_data_sz DROP COLUMN id;

-- 添加新的联合主键
ALTER TABLE daily_data_sz ADD PRIMARY KEY (ts_code, trade_date);

-- 添加更新时间字段
ALTER TABLE daily_data_sz ADD COLUMN updated_at BIGINT DEFAULT 0 COMMENT '记录更新时间戳';

-- 4. 更新周K线数据表
-- 删除现有主键和自增ID列
ALTER TABLE weekly_data DROP PRIMARY KEY;
ALTER TABLE weekly_data DROP COLUMN id;

-- 添加新的联合主键
ALTER TABLE weekly_data ADD PRIMARY KEY (ts_code, trade_date);

-- 添加更新时间字段
ALTER TABLE weekly_data ADD COLUMN updated_at BIGINT DEFAULT 0 COMMENT '记录更新时间戳';

-- 5. 更新月K线数据表
-- 删除现有主键和自增ID列
ALTER TABLE monthly_data DROP PRIMARY KEY;
ALTER TABLE monthly_data DROP COLUMN id;

-- 添加新的联合主键
ALTER TABLE monthly_data ADD PRIMARY KEY (ts_code, trade_date);

-- 添加更新时间字段
ALTER TABLE monthly_data ADD COLUMN updated_at BIGINT DEFAULT 0 COMMENT '记录更新时间戳';

-- 6. 更新年K线数据表
-- 删除现有主键和自增ID列
ALTER TABLE yearly_data DROP PRIMARY KEY;
ALTER TABLE yearly_data DROP COLUMN id;

-- 添加新的联合主键
ALTER TABLE yearly_data ADD PRIMARY KEY (ts_code, trade_date);

-- 添加更新时间字段
ALTER TABLE yearly_data ADD COLUMN updated_at BIGINT DEFAULT 0 COMMENT '记录更新时间戳';

-- 7. 创建索引以提高查询性能
-- 为股票代码创建索引（如果不存在）
CREATE INDEX IF NOT EXISTS idx_daily_data_sh_ts_code ON daily_data_sh (ts_code);
CREATE INDEX IF NOT EXISTS idx_daily_data_sz_ts_code ON daily_data_sz (ts_code);
CREATE INDEX IF NOT EXISTS idx_weekly_data_ts_code ON weekly_data (ts_code);
CREATE INDEX IF NOT EXISTS idx_monthly_data_ts_code ON monthly_data (ts_code);
CREATE INDEX IF NOT EXISTS idx_yearly_data_ts_code ON yearly_data (ts_code);

-- 为交易日期创建索引（如果不存在）
CREATE INDEX IF NOT EXISTS idx_daily_data_sh_trade_date ON daily_data_sh (trade_date);
CREATE INDEX IF NOT EXISTS idx_daily_data_sz_trade_date ON daily_data_sz (trade_date);
CREATE INDEX IF NOT EXISTS idx_weekly_data_trade_date ON weekly_data (trade_date);
CREATE INDEX IF NOT EXISTS idx_monthly_data_trade_date ON monthly_data (trade_date);
CREATE INDEX IF NOT EXISTS idx_yearly_data_trade_date ON yearly_data (trade_date);

-- 8. 更新现有记录的updated_at字段
UPDATE daily_data_sh SET updated_at = created_at WHERE updated_at = 0;
UPDATE daily_data_sz SET updated_at = created_at WHERE updated_at = 0;
UPDATE weekly_data SET updated_at = created_at WHERE updated_at = 0;
UPDATE monthly_data SET updated_at = created_at WHERE updated_at = 0;
UPDATE yearly_data SET updated_at = created_at WHERE updated_at = 0;

-- 迁移完成
-- 新的表结构特点：
-- 1. 使用股票代码(ts_code) + 交易日期(trade_date)作为联合主键
-- 2. 去掉了自增ID字段，节省存储空间
-- 3. 添加了updated_at字段，便于跟踪数据更新
-- 4. 保持了原有的索引结构，确保查询性能
-- 5. 联合主键确保了数据的唯一性，避免重复插入