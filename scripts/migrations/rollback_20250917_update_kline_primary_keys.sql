-- 回滚K线数据表的主键结构更新
-- 将联合主键改回自增ID主键
-- 回滚时间：2025-09-17

-- 警告：此回滚脚本会丢失updated_at字段的数据
-- 建议在执行前备份数据

-- 1. 回滚日K线数据表（上海交易所）
-- 删除联合主键
ALTER TABLE daily_data_sh DROP PRIMARY KEY;

-- 添加自增ID列
ALTER TABLE daily_data_sh ADD COLUMN id INT AUTO_INCREMENT PRIMARY KEY FIRST;

-- 删除updated_at字段
ALTER TABLE daily_data_sh DROP COLUMN updated_at;

-- 重新创建原有索引
CREATE INDEX idx_daily_data_sh_ts_code ON daily_data_sh (ts_code);
CREATE INDEX idx_daily_data_sh_trade_date ON daily_data_sh (trade_date);

-- 2. 回滚日K线数据表（深圳交易所）
-- 删除联合主键
ALTER TABLE daily_data_sz DROP PRIMARY KEY;

-- 添加自增ID列
ALTER TABLE daily_data_sz ADD COLUMN id INT AUTO_INCREMENT PRIMARY KEY FIRST;

-- 删除updated_at字段
ALTER TABLE daily_data_sz DROP COLUMN updated_at;

-- 重新创建原有索引
CREATE INDEX idx_daily_data_sz_ts_code ON daily_data_sz (ts_code);
CREATE INDEX idx_daily_data_sz_trade_date ON daily_data_sz (trade_date);

-- 3. 回滚周K线数据表
-- 删除联合主键
ALTER TABLE weekly_data DROP PRIMARY KEY;

-- 添加自增ID列
ALTER TABLE weekly_data ADD COLUMN id INT AUTO_INCREMENT PRIMARY KEY FIRST;

-- 删除updated_at字段
ALTER TABLE weekly_data DROP COLUMN updated_at;

-- 重新创建原有索引
CREATE INDEX idx_weekly_data_ts_code ON weekly_data (ts_code);
CREATE INDEX idx_weekly_data_trade_date ON weekly_data (trade_date);

-- 4. 回滚月K线数据表
-- 删除联合主键
ALTER TABLE monthly_data DROP PRIMARY KEY;

-- 添加自增ID列
ALTER TABLE monthly_data ADD COLUMN id INT AUTO_INCREMENT PRIMARY KEY FIRST;

-- 删除updated_at字段
ALTER TABLE monthly_data DROP COLUMN updated_at;

-- 重新创建原有索引
CREATE INDEX idx_monthly_data_ts_code ON monthly_data (ts_code);
CREATE INDEX idx_monthly_data_trade_date ON monthly_data (trade_date);

-- 5. 回滚年K线数据表
-- 删除联合主键
ALTER TABLE yearly_data DROP PRIMARY KEY;

-- 添加自增ID列
ALTER TABLE yearly_data ADD COLUMN id INT AUTO_INCREMENT PRIMARY KEY FIRST;

-- 删除updated_at字段
ALTER TABLE yearly_data DROP COLUMN updated_at;

-- 重新创建原有索引
CREATE INDEX idx_yearly_data_ts_code ON yearly_data (ts_code);
CREATE INDEX idx_yearly_data_trade_date ON yearly_data (trade_date);

-- 回滚完成
-- 表结构已恢复到原始状态：
-- 1. 使用自增ID作为主键
-- 2. 保留了ts_code和trade_date的索引
-- 3. 删除了updated_at字段