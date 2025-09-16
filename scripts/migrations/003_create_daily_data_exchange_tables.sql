-- 创建上海证券交易所日线数据表
CREATE TABLE IF NOT EXISTS daily_data_sh (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    ts_code VARCHAR(20) NOT NULL COMMENT '股票代码，如：600000.SH',
    trade_date INT NOT NULL COMMENT '交易日期，YYYYMMDD格式',
    open DECIMAL(10,3) DEFAULT 0 COMMENT '开盘价，单位：元',
    high DECIMAL(10,3) DEFAULT 0 COMMENT '最高价，单位：元',
    low DECIMAL(10,3) DEFAULT 0 COMMENT '最低价，单位：元',
    close DECIMAL(10,3) DEFAULT 0 COMMENT '收盘价，单位：元',
    volume BIGINT DEFAULT 0 COMMENT '成交量，单位：股',
    amount DECIMAL(20,2) DEFAULT 0 COMMENT '成交额，单位：元',
    created_at BIGINT DEFAULT 0 COMMENT '记录创建时间戳',
    
    INDEX idx_ts_code (ts_code),
    INDEX idx_trade_date (trade_date),
    INDEX idx_ts_code_trade_date (ts_code, trade_date),
    UNIQUE KEY uk_ts_code_trade_date (ts_code, trade_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='上海证券交易所日线数据表';

-- 创建深圳证券交易所日线数据表
CREATE TABLE IF NOT EXISTS daily_data_sz (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    ts_code VARCHAR(20) NOT NULL COMMENT '股票代码，如：000001.SZ',
    trade_date INT NOT NULL COMMENT '交易日期，YYYYMMDD格式',
    open DECIMAL(10,3) DEFAULT 0 COMMENT '开盘价，单位：元',
    high DECIMAL(10,3) DEFAULT 0 COMMENT '最高价，单位：元',
    low DECIMAL(10,3) DEFAULT 0 COMMENT '最低价，单位：元',
    close DECIMAL(10,3) DEFAULT 0 COMMENT '收盘价，单位：元',
    volume BIGINT DEFAULT 0 COMMENT '成交量，单位：股',
    amount DECIMAL(20,2) DEFAULT 0 COMMENT '成交额，单位：元',
    created_at BIGINT DEFAULT 0 COMMENT '记录创建时间戳',
    
    INDEX idx_ts_code (ts_code),
    INDEX idx_trade_date (trade_date),
    INDEX idx_ts_code_trade_date (ts_code, trade_date),
    UNIQUE KEY uk_ts_code_trade_date (ts_code, trade_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='深圳证券交易所日线数据表';