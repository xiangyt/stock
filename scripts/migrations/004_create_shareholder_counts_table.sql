-- 创建股东户数表
CREATE TABLE IF NOT EXISTS shareholder_counts (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    ts_code VARCHAR(20) NOT NULL COMMENT '股票代码',
    security_code VARCHAR(10) NOT NULL COMMENT '证券代码',
    security_name VARCHAR(100) NOT NULL COMMENT '证券名称',
    end_date DATE NOT NULL COMMENT '统计截止日期',
    holder_num BIGINT NOT NULL COMMENT '股东户数',
    pre_holder_num BIGINT DEFAULT 0 COMMENT '上期股东户数',
    holder_num_change BIGINT DEFAULT 0 COMMENT '股东户数变化',
    holder_num_ratio DECIMAL(8,4) DEFAULT 0 COMMENT '股东户数变化比例(%)',
    avg_market_cap DECIMAL(15,2) DEFAULT 0 COMMENT '户均市值(元)',
    avg_hold_num DECIMAL(15,2) DEFAULT 0 COMMENT '户均持股数(股)',
    total_market_cap DECIMAL(20,2) DEFAULT 0 COMMENT '总市值(元)',
    total_a_shares BIGINT DEFAULT 0 COMMENT '总股本(股)',
    interval_chrate DECIMAL(8,4) DEFAULT 0 COMMENT '区间涨跌幅(%)',
    change_shares BIGINT DEFAULT 0 COMMENT '变动股本',
    change_reason VARCHAR(100) DEFAULT '' COMMENT '变动原因',
    hold_notice_date DATE DEFAULT NULL COMMENT '股东户数公告日期',
    pre_end_date DATE DEFAULT NULL COMMENT '上期统计截止日期',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    -- 索引
    INDEX idx_ts_code (ts_code),
    INDEX idx_end_date (end_date),
    INDEX idx_ts_code_end_date (ts_code, end_date),
    INDEX idx_holder_num (holder_num),
    INDEX idx_avg_market_cap (avg_market_cap),
    
    -- 唯一约束：同一股票同一日期只能有一条记录
    UNIQUE KEY uk_ts_code_end_date (ts_code, end_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='股东户数表';