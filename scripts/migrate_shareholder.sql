-- 股东户数数据表迁移脚本
-- 创建股东户数表

CREATE TABLE IF NOT EXISTS `shareholder_counts` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID，数据库自增',
    `ts_code` varchar(20) NOT NULL COMMENT '股票代码，如：000001.SZ',
    `security_code` varchar(10) NOT NULL COMMENT '证券代码，如：000001',
    `security_name` varchar(100) NOT NULL COMMENT '证券简称，如：平安银行',
    `end_date` date NOT NULL COMMENT '统计截止日期，如：2025-06-30',
    `holder_num` bigint NOT NULL COMMENT '股东户数，单位：户',
    `pre_holder_num` bigint DEFAULT NULL COMMENT '上期股东户数，单位：户',
    `holder_num_change` bigint DEFAULT NULL COMMENT '股东户数变化，单位：户',
    `holder_num_ratio` decimal(8,4) DEFAULT NULL COMMENT '股东户数变化比例，单位：%',
    `avg_market_cap` decimal(15,2) DEFAULT NULL COMMENT '户均市值，单位：元',
    `avg_hold_num` decimal(15,2) DEFAULT NULL COMMENT '户均持股数，单位：股',
    `total_market_cap` decimal(20,2) DEFAULT NULL COMMENT '总市值，单位：元',
    `total_a_shares` bigint DEFAULT NULL COMMENT '总股本，单位：股',
    `interval_chrate` decimal(8,4) DEFAULT NULL COMMENT '区间涨跌幅，单位：%',
    `change_shares` bigint DEFAULT NULL COMMENT '股本变动，单位：股',
    `change_reason` varchar(100) DEFAULT NULL COMMENT '变动原因，如：发行融资',
    `hold_notice_date` date DEFAULT NULL COMMENT '公告日期',
    `pre_end_date` date DEFAULT NULL COMMENT '上期截止日期',
    `created_at` datetime(3) DEFAULT NULL COMMENT '记录创建时间',
    `updated_at` datetime(3) DEFAULT NULL COMMENT '记录更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_shareholder_counts_ts_code` (`ts_code`),
    KEY `idx_shareholder_counts_end_date` (`end_date`),
    KEY `idx_shareholder_counts_ts_code_end_date` (`ts_code`,`end_date`),
    KEY `idx_shareholder_counts_holder_num` (`holder_num`),
    KEY `idx_shareholder_counts_avg_market_cap` (`avg_market_cap`),
    KEY `idx_shareholder_counts_holder_num_change` (`holder_num_change`),
    UNIQUE KEY `uk_shareholder_counts_ts_code_end_date` (`ts_code`,`end_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='股东户数数据表';

-- 创建索引以优化查询性能
CREATE INDEX IF NOT EXISTS `idx_shareholder_counts_security_code` ON `shareholder_counts` (`security_code`);
CREATE INDEX IF NOT EXISTS `idx_shareholder_counts_security_name` ON `shareholder_counts` (`security_name`);
CREATE INDEX IF NOT EXISTS `idx_shareholder_counts_created_at` ON `shareholder_counts` (`created_at`);
CREATE INDEX IF NOT EXISTS `idx_shareholder_counts_updated_at` ON `shareholder_counts` (`updated_at`);

-- 添加外键约束（如果stocks表存在）
-- ALTER TABLE `shareholder_counts` 
-- ADD CONSTRAINT `fk_shareholder_counts_stock` 
-- FOREIGN KEY (`ts_code`) REFERENCES `stocks` (`ts_code`) ON DELETE CASCADE ON UPDATE CASCADE;