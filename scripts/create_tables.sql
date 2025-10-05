-- 股票数据库建表脚本
-- 生成时间: 2025-10-01
-- 数据库: MySQL 8.0+

-- 删除现有表（如果存在）
DROP TABLE IF EXISTS `selection_results`;
DROP TABLE IF EXISTS `shareholder_counts`;
DROP TABLE IF EXISTS `backtest_results`;
DROP TABLE IF EXISTS `performance_reports`;
DROP TABLE IF EXISTS `technical_indicators`;
DROP TABLE IF EXISTS `yearly_data`;
DROP TABLE IF EXISTS `quarterly_data`;
DROP TABLE IF EXISTS `monthly_data_other`;
DROP TABLE IF EXISTS `monthly_data_688`;
DROP TABLE IF EXISTS `monthly_data_605`;
DROP TABLE IF EXISTS `monthly_data_603`;
DROP TABLE IF EXISTS `monthly_data_601`;
DROP TABLE IF EXISTS `monthly_data_600`;
DROP TABLE IF EXISTS `monthly_data_301`;
DROP TABLE IF EXISTS `monthly_data_300`;
DROP TABLE IF EXISTS `monthly_data_002`;
DROP TABLE IF EXISTS `monthly_data_001`;
DROP TABLE IF EXISTS `monthly_data_000`;
DROP TABLE IF EXISTS `monthly_data`;
DROP TABLE IF EXISTS `weekly_data_other`;
DROP TABLE IF EXISTS `weekly_data_688`;
DROP TABLE IF EXISTS `weekly_data_605`;
DROP TABLE IF EXISTS `weekly_data_603`;
DROP TABLE IF EXISTS `weekly_data_601`;
DROP TABLE IF EXISTS `weekly_data_600`;
DROP TABLE IF EXISTS `weekly_data_301`;
DROP TABLE IF EXISTS `weekly_data_300`;
DROP TABLE IF EXISTS `weekly_data_002`;
DROP TABLE IF EXISTS `weekly_data_001`;
DROP TABLE IF EXISTS `weekly_data_000`;
DROP TABLE IF EXISTS `weekly_data`;
DROP TABLE IF EXISTS `daily_data_other`;
DROP TABLE IF EXISTS `daily_data_688`;
DROP TABLE IF EXISTS `daily_data_605`;
DROP TABLE IF EXISTS `daily_data_603`;
DROP TABLE IF EXISTS `daily_data_601`;
DROP TABLE IF EXISTS `daily_data_600`;
DROP TABLE IF EXISTS `daily_data_301`;
DROP TABLE IF EXISTS `daily_data_300`;
DROP TABLE IF EXISTS `daily_data_002`;
DROP TABLE IF EXISTS `daily_data_001`;
DROP TABLE IF EXISTS `daily_data_000`;
DROP TABLE IF EXISTS `daily_data_sz`;
DROP TABLE IF EXISTS `daily_data_sh`;
DROP TABLE IF EXISTS `stocks`;

-- 1. 股票基础信息表
CREATE TABLE `stocks` (
  `ts_code` varchar(20) NOT NULL COMMENT 'Tushare股票代码，如：000001.SZ、600000.SH',
  `symbol` varchar(10) NOT NULL COMMENT '股票代码，如：000001、600000（不含交易所后缀）',
  `name` varchar(100) NOT NULL COMMENT '股票简称，如：平安银行、浦发银行',
  `area` varchar(50) DEFAULT NULL COMMENT '所在地区，如：深圳、上海、北京',
  `industry` varchar(100) DEFAULT NULL COMMENT '所属行业，如：银行、房地产开发、软件开发',
  `market` varchar(10) DEFAULT NULL COMMENT '交易市场，SZ=深交所、SH=上交所、BJ=北交所',
  `list_date` datetime(3) DEFAULT NULL COMMENT '上市日期，首次公开发行日期',
  `is_active` tinyint(1) DEFAULT '1' COMMENT '是否活跃交易，false表示停牌、退市等',
  `created_at` datetime(3) DEFAULT NULL COMMENT '记录创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '记录更新时间',
  PRIMARY KEY (`ts_code`),
  KEY `idx_stocks_symbol` (`symbol`),
  KEY `idx_stocks_market` (`market`),
  KEY `idx_stocks_industry` (`industry`),
  KEY `idx_stocks_list_date` (`list_date`),
  KEY `idx_stocks_is_active` (`is_active`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='股票基础信息表 - A股市场';

-- 2. 日线数据表 - 000开头股票（深交所主板）
CREATE TABLE `daily_data_000` (
  `ts_code` varchar(20) NOT NULL COMMENT '股票代码，如：000001.SZ',
  `trade_date` int NOT NULL COMMENT '交易日期，YYYYMMDD格式，如：20250910',
  `open` decimal(10,3) DEFAULT NULL COMMENT '开盘价，单位：元',
  `high` decimal(10,3) DEFAULT NULL COMMENT '最高价，单位：元',
  `low` decimal(10,3) DEFAULT NULL COMMENT '最低价，单位：元',
  `close` decimal(10,3) DEFAULT NULL COMMENT '收盘价，单位：元',
  `volume` bigint DEFAULT NULL COMMENT '成交量，单位：股（A股以股为单位）',
  `amount` decimal(20,2) DEFAULT NULL COMMENT '成交额，单位：元',
  `created_at` datetime(3) DEFAULT NULL COMMENT '记录创建时间戳',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '记录更新时间戳',
  PRIMARY KEY (`ts_code`,`trade_date`),
  KEY `idx_daily_000_trade_date` (`trade_date`),
  KEY `idx_daily_000_ts_code_date` (`ts_code`,`trade_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='日线数据表 - 000开头股票（深交所主板）';

-- 3. 日线数据表 - 001开头股票（深交所主板）
CREATE TABLE `daily_data_001` (
  `ts_code` varchar(20) NOT NULL COMMENT '股票代码，如：001979.SZ',
  `trade_date` int NOT NULL COMMENT '交易日期，YYYYMMDD格式，如：20250910',
  `open` decimal(10,3) DEFAULT NULL COMMENT '开盘价，单位：元',
  `high` decimal(10,3) DEFAULT NULL COMMENT '最高价，单位：元',
  `low` decimal(10,3) DEFAULT NULL COMMENT '最低价，单位：元',
  `close` decimal(10,3) DEFAULT NULL COMMENT '收盘价，单位：元',
  `volume` bigint DEFAULT NULL COMMENT '成交量，单位：股（A股以股为单位）',
  `amount` decimal(20,2) DEFAULT NULL COMMENT '成交额，单位：元',
  `created_at` datetime(3) DEFAULT NULL COMMENT '记录创建时间戳',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '记录更新时间戳',
  PRIMARY KEY (`ts_code`,`trade_date`),
  KEY `idx_daily_001_trade_date` (`trade_date`),
  KEY `idx_daily_001_ts_code_date` (`ts_code`,`trade_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='日线数据表 - 001开头股票（深交所主板）';

-- 4. 日线数据表 - 002开头股票（深交所中小板）
CREATE TABLE `daily_data_002` (
  `ts_code` varchar(20) NOT NULL COMMENT '股票代码，如：002415.SZ',
  `trade_date` int NOT NULL COMMENT '交易日期，YYYYMMDD格式，如：20250910',
  `open` decimal(10,3) DEFAULT NULL COMMENT '开盘价，单位：元',
  `high` decimal(10,3) DEFAULT NULL COMMENT '最高价，单位：元',
  `low` decimal(10,3) DEFAULT NULL COMMENT '最低价，单位：元',
  `close` decimal(10,3) DEFAULT NULL COMMENT '收盘价，单位：元',
  `volume` bigint DEFAULT NULL COMMENT '成交量，单位：股（A股以股为单位）',
  `amount` decimal(20,2) DEFAULT NULL COMMENT '成交额，单位：元',
  `created_at` datetime(3) DEFAULT NULL COMMENT '记录创建时间戳',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '记录更新时间戳',
  PRIMARY KEY (`ts_code`,`trade_date`),
  KEY `idx_daily_002_trade_date` (`trade_date`),
  KEY `idx_daily_002_ts_code_date` (`ts_code`,`trade_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='日线数据表 - 002开头股票（深交所中小板）';

-- 5. 日线数据表 - 300开头股票（深交所创业板）
CREATE TABLE `daily_data_300` (
  `ts_code` varchar(20) NOT NULL COMMENT '股票代码，如：300750.SZ',
  `trade_date` int NOT NULL COMMENT '交易日期，YYYYMMDD格式，如：20250910',
  `open` decimal(10,3) DEFAULT NULL COMMENT '开盘价，单位：元',
  `high` decimal(10,3) DEFAULT NULL COMMENT '最高价，单位：元',
  `low` decimal(10,3) DEFAULT NULL COMMENT '最低价，单位：元',
  `close` decimal(10,3) DEFAULT NULL COMMENT '收盘价，单位：元',
  `volume` bigint DEFAULT NULL COMMENT '成交量，单位：股（A股以股为单位）',
  `amount` decimal(20,2) DEFAULT NULL COMMENT '成交额，单位：元',
  `created_at` datetime(3) DEFAULT NULL COMMENT '记录创建时间戳',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '记录更新时间戳',
  PRIMARY KEY (`ts_code`,`trade_date`),
  KEY `idx_daily_300_trade_date` (`trade_date`),
  KEY `idx_daily_300_ts_code_date` (`ts_code`,`trade_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='日线数据表 - 300开头股票（深交所创业板）';

-- 6. 日线数据表 - 301开头股票（深交所创业板）
CREATE TABLE `daily_data_301` (
  `ts_code` varchar(20) NOT NULL COMMENT '股票代码，如：301088.SZ',
  `trade_date` int NOT NULL COMMENT '交易日期，YYYYMMDD格式，如：20250910',
  `open` decimal(10,3) DEFAULT NULL COMMENT '开盘价，单位：元',
  `high` decimal(10,3) DEFAULT NULL COMMENT '最高价，单位：元',
  `low` decimal(10,3) DEFAULT NULL COMMENT '最低价，单位：元',
  `close` decimal(10,3) DEFAULT NULL COMMENT '收盘价，单位：元',
  `volume` bigint DEFAULT NULL COMMENT '成交量，单位：股（A股以股为单位）',
  `amount` decimal(20,2) DEFAULT NULL COMMENT '成交额，单位：元',
  `created_at` datetime(3) DEFAULT NULL COMMENT '记录创建时间戳',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '记录更新时间戳',
  PRIMARY KEY (`ts_code`,`trade_date`),
  KEY `idx_daily_301_trade_date` (`trade_date`),
  KEY `idx_daily_301_ts_code_date` (`ts_code`,`trade_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='日线数据表 - 301开头股票（深交所创业板）';

-- 7. 日线数据表 - 600开头股票（上交所主板）
CREATE TABLE `daily_data_600` (
  `ts_code` varchar(20) NOT NULL COMMENT '股票代码，如：600000.SH',
  `trade_date` int NOT NULL COMMENT '交易日期，YYYYMMDD格式，如：20250910',
  `open` decimal(10,3) DEFAULT NULL COMMENT '开盘价，单位：元',
  `high` decimal(10,3) DEFAULT NULL COMMENT '最高价，单位：元',
  `low` decimal(10,3) DEFAULT NULL COMMENT '最低价，单位：元',
  `close` decimal(10,3) DEFAULT NULL COMMENT '收盘价，单位：元',
  `volume` bigint DEFAULT NULL COMMENT '成交量，单位：股（A股以股为单位）',
  `amount` decimal(20,2) DEFAULT NULL COMMENT '成交额，单位：元',
  `created_at` datetime(3) DEFAULT NULL COMMENT '记录创建时间戳',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '记录更新时间戳',
  PRIMARY KEY (`ts_code`,`trade_date`),
  KEY `idx_daily_600_trade_date` (`trade_date`),
  KEY `idx_daily_600_ts_code_date` (`ts_code`,`trade_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='日线数据表 - 600开头股票（上交所主板）';

-- 8. 日线数据表 - 601开头股票（上交所主板）
CREATE TABLE `daily_data_601` (
  `ts_code` varchar(20) NOT NULL COMMENT '股票代码，如：601318.SH',
  `trade_date` int NOT NULL COMMENT '交易日期，YYYYMMDD格式，如：20250910',
  `open` decimal(10,3) DEFAULT NULL COMMENT '开盘价，单位：元',
  `high` decimal(10,3) DEFAULT NULL COMMENT '最高价，单位：元',
  `low` decimal(10,3) DEFAULT NULL COMMENT '最低价，单位：元',
  `close` decimal(10,3) DEFAULT NULL COMMENT '收盘价，单位：元',
  `volume` bigint DEFAULT NULL COMMENT '成交量，单位：股（A股以股为单位）',
  `amount` decimal(20,2) DEFAULT NULL COMMENT '成交额，单位：元',
  `created_at` datetime(3) DEFAULT NULL COMMENT '记录创建时间戳',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '记录更新时间戳',
  PRIMARY KEY (`ts_code`,`trade_date`),
  KEY `idx_daily_601_trade_date` (`trade_date`),
  KEY `idx_daily_601_ts_code_date` (`ts_code`,`trade_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='日线数据表 - 601开头股票（上交所主板）';

-- 9. 日线数据表 - 603开头股票（上交所主板）
CREATE TABLE `daily_data_603` (
  `ts_code` varchar(20) NOT NULL COMMENT '股票代码，如：603259.SH',
  `trade_date` int NOT NULL COMMENT '交易日期，YYYYMMDD格式，如：20250910',
  `open` decimal(10,3) DEFAULT NULL COMMENT '开盘价，单位：元',
  `high` decimal(10,3) DEFAULT NULL COMMENT '最高价，单位：元',
  `low` decimal(10,3) DEFAULT NULL COMMENT '最低价，单位：元',
  `close` decimal(10,3) DEFAULT NULL COMMENT '收盘价，单位：元',
  `volume` bigint DEFAULT NULL COMMENT '成交量，单位：股（A股以股为单位）',
  `amount` decimal(20,2) DEFAULT NULL COMMENT '成交额，单位：元',
  `created_at` datetime(3) DEFAULT NULL COMMENT '记录创建时间戳',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '记录更新时间戳',
  PRIMARY KEY (`ts_code`,`trade_date`),
  KEY `idx_daily_603_trade_date` (`trade_date`),
  KEY `idx_daily_603_ts_code_date` (`ts_code`,`trade_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='日线数据表 - 603开头股票（上交所主板）';

-- 10. 日线数据表 - 605开头股票（上交所主板）
CREATE TABLE `daily_data_605` (
  `ts_code` varchar(20) NOT NULL COMMENT '股票代码，如：605499.SH',
  `trade_date` int NOT NULL COMMENT '交易日期，YYYYMMDD格式，如：20250910',
  `open` decimal(10,3) DEFAULT NULL COMMENT '开盘价，单位：元',
  `high` decimal(10,3) DEFAULT NULL COMMENT '最高价，单位：元',
  `low` decimal(10,3) DEFAULT NULL COMMENT '最低价，单位：元',
  `close` decimal(10,3) DEFAULT NULL COMMENT '收盘价，单位：元',
  `volume` bigint DEFAULT NULL COMMENT '成交量，单位：股（A股以股为单位）',
  `amount` decimal(20,2) DEFAULT NULL COMMENT '成交额，单位：元',
  `created_at` datetime(3) DEFAULT NULL COMMENT '记录创建时间戳',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '记录更新时间戳',
  PRIMARY KEY (`ts_code`,`trade_date`),
  KEY `idx_daily_605_trade_date` (`trade_date`),
  KEY `idx_daily_605_ts_code_date` (`ts_code`,`trade_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='日线数据表 - 605开头股票（上交所主板）';

-- 11. 日线数据表 - 688开头股票（上交所科创板）
CREATE TABLE `daily_data_688` (
  `ts_code` varchar(20) NOT NULL COMMENT '股票代码，如：688981.SH',
  `trade_date` int NOT NULL COMMENT '交易日期，YYYYMMDD格式，如：20250910',
  `open` decimal(10,3) DEFAULT NULL COMMENT '开盘价，单位：元',
  `high` decimal(10,3) DEFAULT NULL COMMENT '最高价，单位：元',
  `low` decimal(10,3) DEFAULT NULL COMMENT '最低价，单位：元',
  `close` decimal(10,3) DEFAULT NULL COMMENT '收盘价，单位：元',
  `volume` bigint DEFAULT NULL COMMENT '成交量，单位：股（A股以股为单位）',
  `amount` decimal(20,2) DEFAULT NULL COMMENT '成交额，单位：元',
  `created_at` datetime(3) DEFAULT NULL COMMENT '记录创建时间戳',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '记录更新时间戳',
  PRIMARY KEY (`ts_code`,`trade_date`),
  KEY `idx_daily_688_trade_date` (`trade_date`),
  KEY `idx_daily_688_ts_code_date` (`ts_code`,`trade_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='日线数据表 - 688开头股票（上交所科创板）';

-- 12. 日线数据表 - 其他股票
CREATE TABLE `daily_data_other` (
  `ts_code` varchar(20) NOT NULL COMMENT '股票代码，如：其他前缀的股票',
  `trade_date` int NOT NULL COMMENT '交易日期，YYYYMMDD格式，如：20250910',
  `open` decimal(10,3) DEFAULT NULL COMMENT '开盘价，单位：元',
  `high` decimal(10,3) DEFAULT NULL COMMENT '最高价，单位：元',
  `low` decimal(10,3) DEFAULT NULL COMMENT '最低价，单位：元',
  `close` decimal(10,3) DEFAULT NULL COMMENT '收盘价，单位：元',
  `volume` bigint DEFAULT NULL COMMENT '成交量，单位：股（A股以股为单位）',
  `amount` decimal(20,2) DEFAULT NULL COMMENT '成交额，单位：元',
  `created_at` datetime(3) DEFAULT NULL COMMENT '记录创建时间戳',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '记录更新时间戳',
  PRIMARY KEY (`ts_code`,`trade_date`),
  KEY `idx_daily_other_trade_date` (`trade_date`),
  KEY `idx_daily_other_ts_code_date` (`ts_code`,`trade_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='日线数据表 - 其他前缀股票';

-- 13. 周K线数据表 - 000开头股票（深交所主板）
CREATE TABLE `weekly_data_000` (
  `ts_code` varchar(20) NOT NULL COMMENT '股票代码，如：000001.SZ',
  `trade_date` int NOT NULL COMMENT '周结束交易日期，YYYYMMDD格式，如：20250910',
  `open` decimal(10,3) DEFAULT NULL COMMENT '周开盘价，单位：元',
  `high` decimal(10,3) DEFAULT NULL COMMENT '周最高价，单位：元',
  `low` decimal(10,3) DEFAULT NULL COMMENT '周最低价，单位：元',
  `close` decimal(10,3) DEFAULT NULL COMMENT '周收盘价，单位：元',
  `volume` bigint DEFAULT NULL COMMENT '周成交量，单位：股',
  `amount` decimal(20,2) DEFAULT NULL COMMENT '周成交额，单位：元',
  `created_at` datetime(3) DEFAULT NULL COMMENT '记录创建时间戳',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '记录更新时间戳',
  PRIMARY KEY (`ts_code`,`trade_date`),
  KEY `idx_weekly_000_trade_date` (`trade_date`),
  KEY `idx_weekly_000_ts_code_date` (`ts_code`,`trade_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='周K线数据表 - 000开头股票（深交所主板）';

-- 14. 周K线数据表 - 001开头股票（深交所主板）
CREATE TABLE `weekly_data_001` (
  `ts_code` varchar(20) NOT NULL COMMENT '股票代码，如：001979.SZ',
  `trade_date` int NOT NULL COMMENT '周结束交易日期，YYYYMMDD格式，如：20250910',
  `open` decimal(10,3) DEFAULT NULL COMMENT '周开盘价，单位：元',
  `high` decimal(10,3) DEFAULT NULL COMMENT '周最高价，单位：元',
  `low` decimal(10,3) DEFAULT NULL COMMENT '周最低价，单位：元',
  `close` decimal(10,3) DEFAULT NULL COMMENT '周收盘价，单位：元',
  `volume` bigint DEFAULT NULL COMMENT '周成交量，单位：股',
  `amount` decimal(20,2) DEFAULT NULL COMMENT '周成交额，单位：元',
  `created_at` datetime(3) DEFAULT NULL COMMENT '记录创建时间戳',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '记录更新时间戳',
  PRIMARY KEY (`ts_code`,`trade_date`),
  KEY `idx_weekly_001_trade_date` (`trade_date`),
  KEY `idx_weekly_001_ts_code_date` (`ts_code`,`trade_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='周K线数据表 - 001开头股票（深交所主板）';

-- 15. 周K线数据表 - 002开头股票（深交所中小板）
CREATE TABLE `weekly_data_002` (
  `ts_code` varchar(20) NOT NULL COMMENT '股票代码，如：002415.SZ',
  `trade_date` int NOT NULL COMMENT '周结束交易日期，YYYYMMDD格式，如：20250910',
  `open` decimal(10,3) DEFAULT NULL COMMENT '周开盘价，单位：元',
  `high` decimal(10,3) DEFAULT NULL COMMENT '周最高价，单位：元',
  `low` decimal(10,3) DEFAULT NULL COMMENT '周最低价，单位：元',
  `close` decimal(10,3) DEFAULT NULL COMMENT '周收盘价，单位：元',
  `volume` bigint DEFAULT NULL COMMENT '周成交量，单位：股',
  `amount` decimal(20,2) DEFAULT NULL COMMENT '周成交额，单位：元',
  `created_at` datetime(3) DEFAULT NULL COMMENT '记录创建时间戳',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '记录更新时间戳',
  PRIMARY KEY (`ts_code`,`trade_date`),
  KEY `idx_weekly_002_trade_date` (`trade_date`),
  KEY `idx_weekly_002_ts_code_date` (`ts_code`,`trade_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='周K线数据表 - 002开头股票（深交所中小板）';

-- 16. 周K线数据表 - 300开头股票（深交所创业板）
CREATE TABLE `weekly_data_300` (
  `ts_code` varchar(20) NOT NULL COMMENT '股票代码，如：300059.SZ',
  `trade_date` int NOT NULL COMMENT '周结束交易日期，YYYYMMDD格式，如：20250910',
  `open` decimal(10,3) DEFAULT NULL COMMENT '周开盘价，单位：元',
  `high` decimal(10,3) DEFAULT NULL COMMENT '周最高价，单位：元',
  `low` decimal(10,3) DEFAULT NULL COMMENT '周最低价，单位：元',
  `close` decimal(10,3) DEFAULT NULL COMMENT '周收盘价，单位：元',
  `volume` bigint DEFAULT NULL COMMENT '周成交量，单位：股',
  `amount` decimal(20,2) DEFAULT NULL COMMENT '周成交额，单位：元',
  `created_at` datetime(3) DEFAULT NULL COMMENT '记录创建时间戳',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '记录更新时间戳',
  PRIMARY KEY (`ts_code`,`trade_date`),
  KEY `idx_weekly_300_trade_date` (`trade_date`),
  KEY `idx_weekly_300_ts_code_date` (`ts_code`,`trade_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='周K线数据表 - 300开头股票（深交所创业板）';

-- 17. 周K线数据表 - 301开头股票（深交所创业板）
CREATE TABLE `weekly_data_301` (
  `ts_code` varchar(20) NOT NULL COMMENT '股票代码，如：301236.SZ',
  `trade_date` int NOT NULL COMMENT '周结束交易日期，YYYYMMDD格式，如：20250910',
  `open` decimal(10,3) DEFAULT NULL COMMENT '周开盘价，单位：元',
  `high` decimal(10,3) DEFAULT NULL COMMENT '周最高价，单位：元',
  `low` decimal(10,3) DEFAULT NULL COMMENT '周最低价，单位：元',
  `close` decimal(10,3) DEFAULT NULL COMMENT '周收盘价，单位：元',
  `volume` bigint DEFAULT NULL COMMENT '周成交量，单位：股',
  `amount` decimal(20,2) DEFAULT NULL COMMENT '周成交额，单位：元',
  `created_at` datetime(3) DEFAULT NULL COMMENT '记录创建时间戳',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '记录更新时间戳',
  PRIMARY KEY (`ts_code`,`trade_date`),
  KEY `idx_weekly_301_trade_date` (`trade_date`),
  KEY `idx_weekly_301_ts_code_date` (`ts_code`,`trade_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='周K线数据表 - 301开头股票（深交所创业板）';

-- 18. 周K线数据表 - 600开头股票（上交所主板）
CREATE TABLE `weekly_data_600` (
  `ts_code` varchar(20) NOT NULL COMMENT '股票代码，如：600000.SH',
  `trade_date` int NOT NULL COMMENT '周结束交易日期，YYYYMMDD格式，如：20250910',
  `open` decimal(10,3) DEFAULT NULL COMMENT '周开盘价，单位：元',
  `high` decimal(10,3) DEFAULT NULL COMMENT '周最高价，单位：元',
  `low` decimal(10,3) DEFAULT NULL COMMENT '周最低价，单位：元',
  `close` decimal(10,3) DEFAULT NULL COMMENT '周收盘价，单位：元',
  `volume` bigint DEFAULT NULL COMMENT '周成交量，单位：股',
  `amount` decimal(20,2) DEFAULT NULL COMMENT '周成交额，单位：元',
  `created_at` datetime(3) DEFAULT NULL COMMENT '记录创建时间戳',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '记录更新时间戳',
  PRIMARY KEY (`ts_code`,`trade_date`),
  KEY `idx_weekly_600_trade_date` (`trade_date`),
  KEY `idx_weekly_600_ts_code_date` (`ts_code`,`trade_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='周K线数据表 - 600开头股票（上交所主板）';

-- 19. 周K线数据表 - 601开头股票（上交所主板）
CREATE TABLE `weekly_data_601` (
  `ts_code` varchar(20) NOT NULL COMMENT '股票代码，如：601318.SH',
  `trade_date` int NOT NULL COMMENT '周结束交易日期，YYYYMMDD格式，如：20250910',
  `open` decimal(10,3) DEFAULT NULL COMMENT '周开盘价，单位：元',
  `high` decimal(10,3) DEFAULT NULL COMMENT '周最高价，单位：元',
  `low` decimal(10,3) DEFAULT NULL COMMENT '周最低价，单位：元',
  `close` decimal(10,3) DEFAULT NULL COMMENT '周收盘价，单位：元',
  `volume` bigint DEFAULT NULL COMMENT '周成交量，单位：股',
  `amount` decimal(20,2) DEFAULT NULL COMMENT '周成交额，单位：元',
  `created_at` datetime(3) DEFAULT NULL COMMENT '记录创建时间戳',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '记录更新时间戳',
  PRIMARY KEY (`ts_code`,`trade_date`),
  KEY `idx_weekly_601_trade_date` (`trade_date`),
  KEY `idx_weekly_601_ts_code_date` (`ts_code`,`trade_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='周K线数据表 - 601开头股票（上交所主板）';

-- 20. 周K线数据表 - 603开头股票（上交所主板）
CREATE TABLE `weekly_data_603` (
  `ts_code` varchar(20) NOT NULL COMMENT '股票代码，如：603259.SH',
  `trade_date` int NOT NULL COMMENT '周结束交易日期，YYYYMMDD格式，如：20250910',
  `open` decimal(10,3) DEFAULT NULL COMMENT '周开盘价，单位：元',
  `high` decimal(10,3) DEFAULT NULL COMMENT '周最高价，单位：元',
  `low` decimal(10,3) DEFAULT NULL COMMENT '周最低价，单位：元',
  `close` decimal(10,3) DEFAULT NULL COMMENT '周收盘价，单位：元',
  `volume` bigint DEFAULT NULL COMMENT '周成交量，单位：股',
  `amount` decimal(20,2) DEFAULT NULL COMMENT '周成交额，单位：元',
  `created_at` datetime(3) DEFAULT NULL COMMENT '记录创建时间戳',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '记录更新时间戳',
  PRIMARY KEY (`ts_code`,`trade_date`),
  KEY `idx_weekly_603_trade_date` (`trade_date`),
  KEY `idx_weekly_603_ts_code_date` (`ts_code`,`trade_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='周K线数据表 - 603开头股票（上交所主板）';

-- 21. 周K线数据表 - 605开头股票（上交所主板）
CREATE TABLE `weekly_data_605` (
  `ts_code` varchar(20) NOT NULL COMMENT '股票代码，如：605117.SH',
  `trade_date` int NOT NULL COMMENT '周结束交易日期，YYYYMMDD格式，如：20250910',
  `open` decimal(10,3) DEFAULT NULL COMMENT '周开盘价，单位：元',
  `high` decimal(10,3) DEFAULT NULL COMMENT '周最高价，单位：元',
  `low` decimal(10,3) DEFAULT NULL COMMENT '周最低价，单位：元',
  `close` decimal(10,3) DEFAULT NULL COMMENT '周收盘价，单位：元',
  `volume` bigint DEFAULT NULL COMMENT '周成交量，单位：股',
  `amount` decimal(20,2) DEFAULT NULL COMMENT '周成交额，单位：元',
  `created_at` datetime(3) DEFAULT NULL COMMENT '记录创建时间戳',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '记录更新时间戳',
  PRIMARY KEY (`ts_code`,`trade_date`),
  KEY `idx_weekly_605_trade_date` (`trade_date`),
  KEY `idx_weekly_605_ts_code_date` (`ts_code`,`trade_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='周K线数据表 - 605开头股票（上交所主板）';

-- 22. 周K线数据表 - 688开头股票（上交所科创板）
CREATE TABLE `weekly_data_688` (
  `ts_code` varchar(20) NOT NULL COMMENT '股票代码，如：688009.SH',
  `trade_date` int NOT NULL COMMENT '周结束交易日期，YYYYMMDD格式，如：20250910',
  `open` decimal(10,3) DEFAULT NULL COMMENT '周开盘价，单位：元',
  `high` decimal(10,3) DEFAULT NULL COMMENT '周最高价，单位：元',
  `low` decimal(10,3) DEFAULT NULL COMMENT '周最低价，单位：元',
  `close` decimal(10,3) DEFAULT NULL COMMENT '周收盘价，单位：元',
  `volume` bigint DEFAULT NULL COMMENT '周成交量，单位：股',
  `amount` decimal(20,2) DEFAULT NULL COMMENT '周成交额，单位：元',
  `created_at` datetime(3) DEFAULT NULL COMMENT '记录创建时间戳',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '记录更新时间戳',
  PRIMARY KEY (`ts_code`,`trade_date`),
  KEY `idx_weekly_688_trade_date` (`trade_date`),
  KEY `idx_weekly_688_ts_code_date` (`ts_code`,`trade_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='周K线数据表 - 688开头股票（上交所科创板）';

-- 23. 周K线数据表 - 其他股票
CREATE TABLE `weekly_data_other` (
  `ts_code` varchar(20) NOT NULL COMMENT '股票代码，如：其他前缀的股票',
  `trade_date` int NOT NULL COMMENT '周结束交易日期，YYYYMMDD格式，如：20250910',
  `open` decimal(10,3) DEFAULT NULL COMMENT '周开盘价，单位：元',
  `high` decimal(10,3) DEFAULT NULL COMMENT '周最高价，单位：元',
  `low` decimal(10,3) DEFAULT NULL COMMENT '周最低价，单位：元',
  `close` decimal(10,3) DEFAULT NULL COMMENT '周收盘价，单位：元',
  `volume` bigint DEFAULT NULL COMMENT '周成交量，单位：股',
  `amount` decimal(20,2) DEFAULT NULL COMMENT '周成交额，单位：元',
  `created_at` datetime(3) DEFAULT NULL COMMENT '记录创建时间戳',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '记录更新时间戳',
  PRIMARY KEY (`ts_code`,`trade_date`),
  KEY `idx_weekly_other_trade_date` (`trade_date`),
  KEY `idx_weekly_other_ts_code_date` (`ts_code`,`trade_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='周K线数据表 - 其他前缀股票';

-- 24. 月K线数据表 - 000开头股票（深交所主板）
CREATE TABLE `monthly_data_000` (
  `ts_code` varchar(20) NOT NULL COMMENT '股票代码，如：000001.SZ',
  `trade_date` int NOT NULL COMMENT '月结束交易日期，YYYYMMDD格式，如：20250930',
  `open` decimal(10,3) DEFAULT NULL COMMENT '月开盘价，单位：元',
  `high` decimal(10,3) DEFAULT NULL COMMENT '月最高价，单位：元',
  `low` decimal(10,3) DEFAULT NULL COMMENT '月最低价，单位：元',
  `close` decimal(10,3) DEFAULT NULL COMMENT '月收盘价，单位：元',
  `volume` bigint DEFAULT NULL COMMENT '月成交量，单位：股',
  `amount` decimal(20,2) DEFAULT NULL COMMENT '月成交额，单位：元',
  `created_at` datetime(3) DEFAULT NULL COMMENT '记录创建时间戳',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '记录更新时间戳',
  PRIMARY KEY (`ts_code`,`trade_date`),
  KEY `idx_monthly_000_trade_date` (`trade_date`),
  KEY `idx_monthly_000_ts_code_date` (`ts_code`,`trade_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='月K线数据表 - 000开头股票（深交所主板）';

-- 25. 月K线数据表 - 001开头股票（深交所主板）
CREATE TABLE `monthly_data_001` (
  `ts_code` varchar(20) NOT NULL COMMENT '股票代码，如：001979.SZ',
  `trade_date` int NOT NULL COMMENT '月结束交易日期，YYYYMMDD格式，如：20250930',
  `open` decimal(10,3) DEFAULT NULL COMMENT '月开盘价，单位：元',
  `high` decimal(10,3) DEFAULT NULL COMMENT '月最高价，单位：元',
  `low` decimal(10,3) DEFAULT NULL COMMENT '月最低价，单位：元',
  `close` decimal(10,3) DEFAULT NULL COMMENT '月收盘价，单位：元',
  `volume` bigint DEFAULT NULL COMMENT '月成交量，单位：股',
  `amount` decimal(20,2) DEFAULT NULL COMMENT '月成交额，单位：元',
  `created_at` datetime(3) DEFAULT NULL COMMENT '记录创建时间戳',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '记录更新时间戳',
  PRIMARY KEY (`ts_code`,`trade_date`),
  KEY `idx_monthly_001_trade_date` (`trade_date`),
  KEY `idx_monthly_001_ts_code_date` (`ts_code`,`trade_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='月K线数据表 - 001开头股票（深交所主板）';

-- 26. 月K线数据表 - 002开头股票（深交所中小板）
CREATE TABLE `monthly_data_002` (
  `ts_code` varchar(20) NOT NULL COMMENT '股票代码，如：002415.SZ',
  `trade_date` int NOT NULL COMMENT '月结束交易日期，YYYYMMDD格式，如：20250930',
  `open` decimal(10,3) DEFAULT NULL COMMENT '月开盘价，单位：元',
  `high` decimal(10,3) DEFAULT NULL COMMENT '月最高价，单位：元',
  `low` decimal(10,3) DEFAULT NULL COMMENT '月最低价，单位：元',
  `close` decimal(10,3) DEFAULT NULL COMMENT '月收盘价，单位：元',
  `volume` bigint DEFAULT NULL COMMENT '月成交量，单位：股',
  `amount` decimal(20,2) DEFAULT NULL COMMENT '月成交额，单位：元',
  `created_at` datetime(3) DEFAULT NULL COMMENT '记录创建时间戳',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '记录更新时间戳',
  PRIMARY KEY (`ts_code`,`trade_date`),
  KEY `idx_monthly_002_trade_date` (`trade_date`),
  KEY `idx_monthly_002_ts_code_date` (`ts_code`,`trade_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='月K线数据表 - 002开头股票（深交所中小板）';

-- 27. 月K线数据表 - 300开头股票（深交所创业板）
CREATE TABLE `monthly_data_300` (
  `ts_code` varchar(20) NOT NULL COMMENT '股票代码，如：300059.SZ',
  `trade_date` int NOT NULL COMMENT '月结束交易日期，YYYYMMDD格式，如：20250930',
  `open` decimal(10,3) DEFAULT NULL COMMENT '月开盘价，单位：元',
  `high` decimal(10,3) DEFAULT NULL COMMENT '月最高价，单位：元',
  `low` decimal(10,3) DEFAULT NULL COMMENT '月最低价，单位：元',
  `close` decimal(10,3) DEFAULT NULL COMMENT '月收盘价，单位：元',
  `volume` bigint DEFAULT NULL COMMENT '月成交量，单位：股',
  `amount` decimal(20,2) DEFAULT NULL COMMENT '月成交额，单位：元',
  `created_at` datetime(3) DEFAULT NULL COMMENT '记录创建时间戳',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '记录更新时间戳',
  PRIMARY KEY (`ts_code`,`trade_date`),
  KEY `idx_monthly_300_trade_date` (`trade_date`),
  KEY `idx_monthly_300_ts_code_date` (`ts_code`,`trade_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='月K线数据表 - 300开头股票（深交所创业板）';

-- 28. 月K线数据表 - 301开头股票（深交所创业板）
CREATE TABLE `monthly_data_301` (
  `ts_code` varchar(20) NOT NULL COMMENT '股票代码，如：301236.SZ',
  `trade_date` int NOT NULL COMMENT '月结束交易日期，YYYYMMDD格式，如：20250930',
  `open` decimal(10,3) DEFAULT NULL COMMENT '月开盘价，单位：元',
  `high` decimal(10,3) DEFAULT NULL COMMENT '月最高价，单位：元',
  `low` decimal(10,3) DEFAULT NULL COMMENT '月最低价，单位：元',
  `close` decimal(10,3) DEFAULT NULL COMMENT '月收盘价，单位：元',
  `volume` bigint DEFAULT NULL COMMENT '月成交量，单位：股',
  `amount` decimal(20,2) DEFAULT NULL COMMENT '月成交额，单位：元',
  `created_at` datetime(3) DEFAULT NULL COMMENT '记录创建时间戳',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '记录更新时间戳',
  PRIMARY KEY (`ts_code`,`trade_date`),
  KEY `idx_monthly_301_trade_date` (`trade_date`),
  KEY `idx_monthly_301_ts_code_date` (`ts_code`,`trade_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='月K线数据表 - 301开头股票（深交所创业板）';

-- 29. 月K线数据表 - 600开头股票（上交所主板）
CREATE TABLE `monthly_data_600` (
  `ts_code` varchar(20) NOT NULL COMMENT '股票代码，如：600000.SH',
  `trade_date` int NOT NULL COMMENT '月结束交易日期，YYYYMMDD格式，如：20250930',
  `open` decimal(10,3) DEFAULT NULL COMMENT '月开盘价，单位：元',
  `high` decimal(10,3) DEFAULT NULL COMMENT '月最高价，单位：元',
  `low` decimal(10,3) DEFAULT NULL COMMENT '月最低价，单位：元',
  `close` decimal(10,3) DEFAULT NULL COMMENT '月收盘价，单位：元',
  `volume` bigint DEFAULT NULL COMMENT '月成交量，单位：股',
  `amount` decimal(20,2) DEFAULT NULL COMMENT '月成交额，单位：元',
  `created_at` datetime(3) DEFAULT NULL COMMENT '记录创建时间戳',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '记录更新时间戳',
  PRIMARY KEY (`ts_code`,`trade_date`),
  KEY `idx_monthly_600_trade_date` (`trade_date`),
  KEY `idx_monthly_600_ts_code_date` (`ts_code`,`trade_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='月K线数据表 - 600开头股票（上交所主板）';

-- 30. 月K线数据表 - 601开头股票（上交所主板）
CREATE TABLE `monthly_data_601` (
  `ts_code` varchar(20) NOT NULL COMMENT '股票代码，如：601318.SH',
  `trade_date` int NOT NULL COMMENT '月结束交易日期，YYYYMMDD格式，如：20250930',
  `open` decimal(10,3) DEFAULT NULL COMMENT '月开盘价，单位：元',
  `high` decimal(10,3) DEFAULT NULL COMMENT '月最高价，单位：元',
  `low` decimal(10,3) DEFAULT NULL COMMENT '月最低价，单位：元',
  `close` decimal(10,3) DEFAULT NULL COMMENT '月收盘价，单位：元',
  `volume` bigint DEFAULT NULL COMMENT '月成交量，单位：股',
  `amount` decimal(20,2) DEFAULT NULL COMMENT '月成交额，单位：元',
  `created_at` datetime(3) DEFAULT NULL COMMENT '记录创建时间戳',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '记录更新时间戳',
  PRIMARY KEY (`ts_code`,`trade_date`),
  KEY `idx_monthly_601_trade_date` (`trade_date`),
  KEY `idx_monthly_601_ts_code_date` (`ts_code`,`trade_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='月K线数据表 - 601开头股票（上交所主板）';

-- 31. 月K线数据表 - 603开头股票（上交所主板）
CREATE TABLE `monthly_data_603` (
  `ts_code` varchar(20) NOT NULL COMMENT '股票代码，如：603259.SH',
  `trade_date` int NOT NULL COMMENT '月结束交易日期，YYYYMMDD格式，如：20250930',
  `open` decimal(10,3) DEFAULT NULL COMMENT '月开盘价，单位：元',
  `high` decimal(10,3) DEFAULT NULL COMMENT '月最高价，单位：元',
  `low` decimal(10,3) DEFAULT NULL COMMENT '月最低价，单位：元',
  `close` decimal(10,3) DEFAULT NULL COMMENT '月收盘价，单位：元',
  `volume` bigint DEFAULT NULL COMMENT '月成交量，单位：股',
  `amount` decimal(20,2) DEFAULT NULL COMMENT '月成交额，单位：元',
  `created_at` datetime(3) DEFAULT NULL COMMENT '记录创建时间戳',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '记录更新时间戳',
  PRIMARY KEY (`ts_code`,`trade_date`),
  KEY `idx_monthly_603_trade_date` (`trade_date`),
  KEY `idx_monthly_603_ts_code_date` (`ts_code`,`trade_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='月K线数据表 - 603开头股票（上交所主板）';

-- 32. 月K线数据表 - 605开头股票（上交所主板）
CREATE TABLE `monthly_data_605` (
  `ts_code` varchar(20) NOT NULL COMMENT '股票代码，如：605117.SH',
  `trade_date` int NOT NULL COMMENT '月结束交易日期，YYYYMMDD格式，如：20250930',
  `open` decimal(10,3) DEFAULT NULL COMMENT '月开盘价，单位：元',
  `high` decimal(10,3) DEFAULT NULL COMMENT '月最高价，单位：元',
  `low` decimal(10,3) DEFAULT NULL COMMENT '月最低价，单位：元',
  `close` decimal(10,3) DEFAULT NULL COMMENT '月收盘价，单位：元',
  `volume` bigint DEFAULT NULL COMMENT '月成交量，单位：股',
  `amount` decimal(20,2) DEFAULT NULL COMMENT '月成交额，单位：元',
  `created_at` datetime(3) DEFAULT NULL COMMENT '记录创建时间戳',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '记录更新时间戳',
  PRIMARY KEY (`ts_code`,`trade_date`),
  KEY `idx_monthly_605_trade_date` (`trade_date`),
  KEY `idx_monthly_605_ts_code_date` (`ts_code`,`trade_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='月K线数据表 - 605开头股票（上交所主板）';

-- 33. 月K线数据表 - 688开头股票（上交所科创板）
CREATE TABLE `monthly_data_688` (
  `ts_code` varchar(20) NOT NULL COMMENT '股票代码，如：688009.SH',
  `trade_date` int NOT NULL COMMENT '月结束交易日期，YYYYMMDD格式，如：20250930',
  `open` decimal(10,3) DEFAULT NULL COMMENT '月开盘价，单位：元',
  `high` decimal(10,3) DEFAULT NULL COMMENT '月最高价，单位：元',
  `low` decimal(10,3) DEFAULT NULL COMMENT '月最低价，单位：元',
  `close` decimal(10,3) DEFAULT NULL COMMENT '月收盘价，单位：元',
  `volume` bigint DEFAULT NULL COMMENT '月成交量，单位：股',
  `amount` decimal(20,2) DEFAULT NULL COMMENT '月成交额，单位：元',
  `created_at` datetime(3) DEFAULT NULL COMMENT '记录创建时间戳',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '记录更新时间戳',
  PRIMARY KEY (`ts_code`,`trade_date`),
  KEY `idx_monthly_688_trade_date` (`trade_date`),
  KEY `idx_monthly_688_ts_code_date` (`ts_code`,`trade_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='月K线数据表 - 688开头股票（上交所科创板）';

-- 34. 月K线数据表 - 其他股票
CREATE TABLE `monthly_data_other` (
  `ts_code` varchar(20) NOT NULL COMMENT '股票代码，如：其他前缀的股票',
  `trade_date` int NOT NULL COMMENT '月结束交易日期，YYYYMMDD格式，如：20250930',
  `open` decimal(10,3) DEFAULT NULL COMMENT '月开盘价，单位：元',
  `high` decimal(10,3) DEFAULT NULL COMMENT '月最高价，单位：元',
  `low` decimal(10,3) DEFAULT NULL COMMENT '月最低价，单位：元',
  `close` decimal(10,3) DEFAULT NULL COMMENT '月收盘价，单位：元',
  `volume` bigint DEFAULT NULL COMMENT '月成交量，单位：股',
  `amount` decimal(20,2) DEFAULT NULL COMMENT '月成交额，单位：元',
  `created_at` datetime(3) DEFAULT NULL COMMENT '记录创建时间戳',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '记录更新时间戳',
  PRIMARY KEY (`ts_code`,`trade_date`),
  KEY `idx_monthly_other_trade_date` (`trade_date`),
  KEY `idx_monthly_other_ts_code_date` (`ts_code`,`trade_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='月K线数据表 - 其他前缀股票';

-- 35. 季K线数据表
CREATE TABLE `quarterly_data` (
  `ts_code` varchar(20) NOT NULL COMMENT '股票代码，如：000001.SZ',
  `trade_date` int NOT NULL COMMENT '季结束交易日期，YYYYMMDD格式，如：20250930',
  `open` decimal(10,3) DEFAULT NULL COMMENT '季开盘价，单位：元',
  `high` decimal(10,3) DEFAULT NULL COMMENT '季最高价，单位：元',
  `low` decimal(10,3) DEFAULT NULL COMMENT '季最低价，单位：元',
  `close` decimal(10,3) DEFAULT NULL COMMENT '季收盘价，单位：元',
  `volume` bigint DEFAULT NULL COMMENT '季成交量，单位：股',
  `amount` decimal(20,2) DEFAULT NULL COMMENT '季成交额，单位：元',
  `created_at` datetime(3) DEFAULT NULL COMMENT '记录创建时间戳',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '记录更新时间戳',
  PRIMARY KEY (`ts_code`,`trade_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='季K线数据表 - A股季K线行情数据';

-- 36. 年K线数据表
CREATE TABLE `yearly_data` (
  `ts_code` varchar(20) NOT NULL COMMENT '股票代码，如：000001.SZ',
  `trade_date` int NOT NULL COMMENT '年结束交易日期，YYYYMMDD格式，如：20251231',
  `open` decimal(10,3) DEFAULT NULL COMMENT '年开盘价，单位：元',
  `high` decimal(10,3) DEFAULT NULL COMMENT '年最高价，单位：元',
  `low` decimal(10,3) DEFAULT NULL COMMENT '年最低价，单位：元',
  `close` decimal(10,3) DEFAULT NULL COMMENT '年收盘价，单位：元',
  `volume` bigint DEFAULT NULL COMMENT '年成交量，单位：股',
  `amount` decimal(20,2) DEFAULT NULL COMMENT '年成交额，单位：元',
  `created_at` datetime(3) DEFAULT NULL COMMENT '记录创建时间戳',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '记录更新时间戳',
  PRIMARY KEY (`ts_code`,`trade_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='年K线数据表 - A股年K线行情数据';

-- 37. 技术指标表
CREATE TABLE `technical_indicators` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID，数据库自增',
  `ts_code` varchar(20) NOT NULL COMMENT '股票代码，如：000001.SZ',
  `trade_date` int NOT NULL COMMENT '交易日期，YYYYMMDD格式，如：20250910',
  `ma5` decimal(10,3) DEFAULT NULL COMMENT '5日移动平均线，单位：元，短期趋势指标',
  `ma10` decimal(10,3) DEFAULT NULL COMMENT '10日移动平均线，单位：元，短期趋势指标',
  `ma20` decimal(10,3) DEFAULT NULL COMMENT '20日移动平均线，单位：元，中期趋势指标',
  `ma60` decimal(10,3) DEFAULT NULL COMMENT '60日移动平均线，单位：元，长期趋势指标',
  `rsi` decimal(8,4) DEFAULT NULL COMMENT '相对强弱指数，范围0-100，>70超买，<30超卖',
  `macd` decimal(10,6) DEFAULT NULL COMMENT 'MACD指标，趋势跟踪指标，正值看涨，负值看跌',
  `macd_signal` decimal(10,6) DEFAULT NULL COMMENT 'MACD信号线，MACD的EMA平滑线',
  `macd_hist` decimal(10,6) DEFAULT NULL COMMENT 'MACD柱状图，MACD-Signal，反映趋势变化',
  `kdj_k` decimal(8,4) DEFAULT NULL COMMENT 'KDJ指标K值，范围0-100，随机指标',
  `kdj_d` decimal(8,4) DEFAULT NULL COMMENT 'KDJ指标D值，范围0-100，K值的平滑线',
  `kdj_j` decimal(8,4) DEFAULT NULL COMMENT 'KDJ指标J值，3K-2D，敏感度最高',
  `created_at` bigint DEFAULT NULL COMMENT '记录创建时间戳',
  PRIMARY KEY (`id`),
  KEY `idx_tech_ts_code` (`ts_code`),
  KEY `idx_tech_trade_date` (`trade_date`),
  KEY `idx_tech_ts_code_date` (`ts_code`,`trade_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='技术指标表 - A股技术分析指标';

-- 38. 业绩报表表
CREATE TABLE `performance_reports` (
  `ts_code` varchar(20) NOT NULL COMMENT '股票代码，如：000001.SZ',
  `report_date` int NOT NULL COMMENT '报告期，YYYYMMDD格式，如：20250630',
  `eps` decimal(10,4) DEFAULT NULL COMMENT '每股收益，单位：元',
  `weight_eps` decimal(10,4) DEFAULT NULL COMMENT '加权每股收益，单位：元',
  `revenue` decimal(20,2) DEFAULT NULL COMMENT '营业总收入，单位：元',
  `revenue_qoq` decimal(8,4) DEFAULT NULL COMMENT '营业总收入同比增长，单位：%',
  `revenue_yoy` decimal(8,4) DEFAULT NULL COMMENT '营业总收入季度环比增长，单位：%',
  `net_profit` decimal(20,2) DEFAULT NULL COMMENT '净利润，单位：元',
  `net_profit_qoq` decimal(8,4) DEFAULT NULL COMMENT '净利润同比增长，单位：%',
  `net_profit_yoy` decimal(8,4) DEFAULT NULL COMMENT '净利润季度环比增长，单位：%',
  `bvps` decimal(10,4) DEFAULT NULL COMMENT '每股净资产，单位：元',
  `gross_margin` decimal(8,4) DEFAULT NULL COMMENT '销售毛利率，单位：%',
  `dividend_yield` decimal(8,4) DEFAULT NULL COMMENT '股息率，单位：%',
  `latest_announcement_date` datetime(3) DEFAULT NULL COMMENT '最新公告日期',
  `first_announcement_date` datetime(3) DEFAULT NULL COMMENT '首次公告日期',
  `created_at` datetime(3) DEFAULT NULL COMMENT '记录创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '记录更新时间',
  PRIMARY KEY (`ts_code`,`report_date`),
  KEY `idx_perf_report_date` (`report_date`),
  KEY `idx_perf_eps` (`eps`),
  KEY `idx_perf_revenue` (`revenue`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='业绩报表表 - A股上市公司业绩报表数据';

-- 39. 股东户数表
CREATE TABLE `shareholder_counts` (
  `ts_code` varchar(20) NOT NULL COMMENT '股票代码，如：000001.SZ',
  `end_date` int NOT NULL COMMENT '统计截止日期，YYYYMMDD格式，如：20250630',
  `security_code` varchar(10) NOT NULL COMMENT '证券代码，如：000001',
  `security_name` varchar(100) NOT NULL COMMENT '证券简称，如：平安银行',
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
  `hold_notice_date` datetime(3) DEFAULT NULL COMMENT '公告日期',
  `pre_end_date` int DEFAULT NULL COMMENT '上期截止日期，YYYYMMDD格式，如：20250331',
  `created_at` datetime(3) DEFAULT NULL COMMENT '记录创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '记录更新时间',
  PRIMARY KEY (`ts_code`,`end_date`),
  KEY `idx_shareholder_end_date` (`end_date`),
  KEY `idx_shareholder_holder_num` (`holder_num`),
  KEY `idx_shareholder_avg_market_cap` (`avg_market_cap`),
  CONSTRAINT `fk_shareholder_counts_stock` FOREIGN KEY (`ts_code`) REFERENCES `stocks` (`ts_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='股东户数表 - A股上市公司股东户数数据';

-- 40. 选股结果表
CREATE TABLE `selection_results` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID，数据库自增',
  `strategy_name` varchar(50) NOT NULL COMMENT '选股策略名称，如：technical、fundamental、combined',
  `ts_code` varchar(20) NOT NULL COMMENT '股票代码，如：000001.SZ',
  `selection_date` datetime(3) NOT NULL COMMENT '选股日期，策略执行日期',
  `score` decimal(8,4) DEFAULT NULL COMMENT '选股评分，0-100分，分数越高越优质',
  `reason` text COMMENT '选股理由，详细说明为什么选中该股票',
  `created_at` datetime(3) DEFAULT NULL COMMENT '记录创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_selection_strategy` (`strategy_name`),
  KEY `idx_selection_date` (`selection_date`),
  KEY `idx_selection_score` (`score`),
  KEY `fk_selection_results_stock` (`ts_code`),
  CONSTRAINT `fk_selection_results_stock` FOREIGN KEY (`ts_code`) REFERENCES `stocks` (`ts_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='选股结果表 - A股选股策略执行结果';

-- 41. 回测结果表
CREATE TABLE `backtest_results` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID，数据库自增',
  `strategy_name` varchar(50) NOT NULL COMMENT '回测策略名称，如：technical、fundamental',
  `start_date` datetime(3) NOT NULL COMMENT '回测开始日期',
  `end_date` datetime(3) NOT NULL COMMENT '回测结束日期',
  `initial_capital` decimal(20,2) DEFAULT NULL COMMENT '初始资金，单位：元',
  `final_capital` decimal(20,2) DEFAULT NULL COMMENT '最终资金，单位：元',
  `total_return` decimal(8,4) DEFAULT NULL COMMENT '总收益率，单位：%，(最终资金-初始资金)/初始资金',
  `annual_return` decimal(8,4) DEFAULT NULL COMMENT '年化收益率，单位：%',
  `max_drawdown` decimal(8,4) DEFAULT NULL COMMENT '最大回撤，单位：%，衡量风险水平',
  `sharpe_ratio` decimal(8,4) DEFAULT NULL COMMENT '夏普比率，风险调整后收益指标',
  `win_rate` decimal(8,4) DEFAULT NULL COMMENT '胜率，单位：%，盈利交易次数/总交易次数',
  `created_at` datetime(3) DEFAULT NULL COMMENT '记录创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_backtest_strategy` (`strategy_name`),
  KEY `idx_backtest_return` (`total_return`),
  KEY `idx_backtest_date` (`start_date`,`end_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='回测结果表 - A股策略回测数据';

-- 插入一些示例数据
INSERT INTO `stocks` (`ts_code`, `symbol`, `name`, `area`, `industry`, `market`, `list_date`, `is_active`, `created_at`, `updated_at`) VALUES
('000001.SZ', '000001', '平安银行', '深圳', '银行', 'SZ', '1991-04-03 00:00:00.000', 1, NOW(3), NOW(3)),
('000002.SZ', '000002', '万科A', '深圳', '房地产开发', 'SZ', '1991-01-29 00:00:00.000', 1, NOW(3), NOW(3)),
('600000.SH', '600000', '浦发银行', '上海', '银行', 'SH', '1999-11-10 00:00:00.000', 1, NOW(3), NOW(3)),
('600036.SH', '600036', '招商银行', '深圳', '银行', 'SH', '2002-04-09 00:00:00.000', 1, NOW(3), NOW(3)),
('000858.SZ', '000858', '五粮液', '宜宾', '白酒', 'SZ', '1998-04-27 00:00:00.000', 1, NOW(3), NOW(3));

-- 创建完成提示
SELECT '数据库表创建完成！' AS message;
SELECT COUNT(*) AS stock_count FROM stocks;