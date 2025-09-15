-- 创建数据库（如果不存在）
CREATE DATABASE IF NOT EXISTS stock_db;

-- 使用数据库
USE stock_db;

-- 创建股票基础信息表
CREATE TABLE IF NOT EXISTS stocks (
    id BIGSERIAL PRIMARY KEY,
    ts_code VARCHAR(20) UNIQUE NOT NULL,
    symbol VARCHAR(10) NOT NULL,
    name VARCHAR(100) NOT NULL,
    area VARCHAR(50),
    industry VARCHAR(100),
    market VARCHAR(10),
    list_date DATE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建日线数据表
CREATE TABLE IF NOT EXISTS daily_data (
    id BIGSERIAL PRIMARY KEY,
    ts_code VARCHAR(20) NOT NULL,
    trade_date DATE NOT NULL,
    open DECIMAL(10,3),
    high DECIMAL(10,3),
    low DECIMAL(10,3),
    close DECIMAL(10,3),
    volume BIGINT,
    amount DECIMAL(20,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(ts_code, trade_date)
);

-- 创建财务数据表
CREATE TABLE IF NOT EXISTS financial_data (
    id BIGSERIAL PRIMARY KEY,
    ts_code VARCHAR(20) NOT NULL,
    end_date DATE NOT NULL,
    roe DECIMAL(8,4),
    roa DECIMAL(8,4),
    gross_profit_margin DECIMAL(8,4),
    net_profit_margin DECIMAL(8,4),
    debt_to_assets DECIMAL(8,4),
    pe_ratio DECIMAL(10,2),
    pb_ratio DECIMAL(10,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(ts_code, end_date)
);

-- 创建技术指标表
CREATE TABLE IF NOT EXISTS technical_indicators (
    id BIGSERIAL PRIMARY KEY,
    ts_code VARCHAR(20) NOT NULL,
    trade_date DATE NOT NULL,
    ma5 DECIMAL(10,3),
    ma10 DECIMAL(10,3),
    ma20 DECIMAL(10,3),
    ma60 DECIMAL(10,3),
    rsi DECIMAL(8,4),
    macd DECIMAL(10,6),
    macd_signal DECIMAL(10,6),
    macd_hist DECIMAL(10,6),
    kdj_k DECIMAL(8,4),
    kdj_d DECIMAL(8,4),
    kdj_j DECIMAL(8,4),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(ts_code, trade_date)
);

-- 创建选股结果表
CREATE TABLE IF NOT EXISTS selection_results (
    id BIGSERIAL PRIMARY KEY,
    strategy_name VARCHAR(50) NOT NULL,
    ts_code VARCHAR(20) NOT NULL,
    selection_date DATE NOT NULL,
    score DECIMAL(8,4),
    reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建回测结果表
CREATE TABLE IF NOT EXISTS backtest_results (
    id BIGSERIAL PRIMARY KEY,
    strategy_name VARCHAR(50) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    initial_capital DECIMAL(20,2),
    final_capital DECIMAL(20,2),
    total_return DECIMAL(8,4),
    annual_return DECIMAL(8,4),
    max_drawdown DECIMAL(8,4),
    sharpe_ratio DECIMAL(8,4),
    win_rate DECIMAL(8,4),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_stocks_ts_code ON stocks(ts_code);
CREATE INDEX IF NOT EXISTS idx_stocks_market ON stocks(market);
CREATE INDEX IF NOT EXISTS idx_stocks_industry ON stocks(industry);

CREATE INDEX IF NOT EXISTS idx_daily_data_ts_code ON daily_data(ts_code);
CREATE INDEX IF NOT EXISTS idx_daily_data_trade_date ON daily_data(trade_date);

CREATE INDEX IF NOT EXISTS idx_financial_data_ts_code ON financial_data(ts_code);
CREATE INDEX IF NOT EXISTS idx_financial_data_end_date ON financial_data(end_date);

CREATE INDEX IF NOT EXISTS idx_technical_indicators_ts_code ON technical_indicators(ts_code);
CREATE INDEX IF NOT EXISTS idx_technical_indicators_trade_date ON technical_indicators(trade_date);

CREATE INDEX IF NOT EXISTS idx_selection_results_strategy ON selection_results(strategy_name);
CREATE INDEX IF NOT EXISTS idx_selection_results_date ON selection_results(selection_date);

-- 插入一些示例数据
INSERT INTO stocks (ts_code, symbol, name, area, industry, market, list_date) VALUES
('000001.SZ', '000001', '平安银行', '深圳', '银行', 'SZ', '1991-04-03'),
('000002.SZ', '000002', '万科A', '深圳', '房地产开发', 'SZ', '1991-01-29'),
('600000.SH', '600000', '浦发银行', '上海', '银行', 'SH', '1999-11-10'),
('600036.SH', '600036', '招商银行', '深圳', '银行', 'SH', '2002-04-09')
ON CONFLICT (ts_code) DO NOTHING;