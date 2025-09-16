-- 数据迁移脚本：将daily_data表的数据迁移到按交易所分表
-- 执行前请确保已经创建了daily_data_sh和daily_data_sz表

-- 迁移上海证券交易所的数据
INSERT INTO daily_data_sh (ts_code, trade_date, open, high, low, close, volume, amount, created_at)
SELECT ts_code, trade_date, open, high, low, close, volume, amount, created_at
FROM daily_data 
WHERE ts_code LIKE '%.SH' OR ts_code LIKE '6%'
ON DUPLICATE KEY UPDATE
    open = VALUES(open),
    high = VALUES(high),
    low = VALUES(low),
    close = VALUES(close),
    volume = VALUES(volume),
    amount = VALUES(amount),
    created_at = VALUES(created_at);

-- 迁移深圳证券交易所的数据
INSERT INTO daily_data_sz (ts_code, trade_date, open, high, low, close, volume, amount, created_at)
SELECT ts_code, trade_date, open, high, low, close, volume, amount, created_at
FROM daily_data 
WHERE ts_code LIKE '%.SZ' OR ts_code LIKE '0%' OR ts_code LIKE '3%'
ON DUPLICATE KEY UPDATE
    open = VALUES(open),
    high = VALUES(high),
    low = VALUES(low),
    close = VALUES(close),
    volume = VALUES(volume),
    amount = VALUES(amount),
    created_at = VALUES(created_at);

-- 验证迁移结果
SELECT 
    'Original daily_data count' as description,
    COUNT(*) as count
FROM daily_data
UNION ALL
SELECT 
    'SH data count' as description,
    COUNT(*) as count
FROM daily_data_sh
UNION ALL
SELECT 
    'SZ data count' as description,
    COUNT(*) as count
FROM daily_data_sz
UNION ALL
SELECT 
    'Total migrated count' as description,
    (SELECT COUNT(*) FROM daily_data_sh) + (SELECT COUNT(*) FROM daily_data_sz) as count;