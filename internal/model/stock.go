package model

import (
	"time"
)

// Stock 股票基础信息模型 - A股市场
type Stock struct {
	ID        uint       `json:"id" gorm:"primaryKey"`                        // 主键ID，数据库自增
	TsCode    string     `json:"ts_code" gorm:"uniqueIndex;size:20;not null"` // Tushare股票代码，如：000001.SZ、600000.SH
	Symbol    string     `json:"symbol" gorm:"size:10;not null"`              // 股票代码，如：000001、600000（不含交易所后缀）
	Name      string     `json:"name" gorm:"size:100;not null"`               // 股票简称，如：平安银行、浦发银行
	Area      string     `json:"area" gorm:"size:50"`                         // 所在地区，如：深圳、上海、北京
	Industry  string     `json:"industry" gorm:"size:100"`                    // 所属行业，如：银行、房地产开发、软件开发
	Market    string     `json:"market" gorm:"size:10"`                       // 交易市场，SZ=深交所、SH=上交所、BJ=北交所
	ListDate  *time.Time `json:"list_date"`                                   // 上市日期，首次公开发行日期
	IsActive  bool       `json:"is_active" gorm:"default:true"`               // 是否活跃交易，false表示停牌、退市等
	CreatedAt time.Time  `json:"created_at"`                                  // 记录创建时间
	UpdatedAt time.Time  `json:"updated_at"`                                  // 记录更新时间
}

// TableName 指定表名
func (Stock) TableName() string {
	return "stocks"
}

// DailyData 日线数据模型 - A股K线数据
type DailyData struct {
	ID        uint    `json:"id" gorm:"primaryKey"`                  // 主键ID，数据库自增
	TsCode    string  `json:"ts_code" gorm:"size:20;not null;index"` // 股票代码，如：000001.SZ
	TradeDate int     `json:"trade_date" gorm:"not null;index"`      // 交易日期，YYYYMMDD格式，如：20250910
	Open      float64 `json:"open" gorm:"type:decimal(10,3)"`        // 开盘价，单位：元
	High      float64 `json:"high" gorm:"type:decimal(10,3)"`        // 最高价，单位：元
	Low       float64 `json:"low" gorm:"type:decimal(10,3)"`         // 最低价，单位：元
	Close     float64 `json:"close" gorm:"type:decimal(10,3)"`       // 收盘价，单位：元
	Volume    int64   `json:"volume"`                                // 成交量，单位：股（A股以股为单位）
	Amount    float64 `json:"amount" gorm:"type:decimal(20,2)"`      // 成交额，单位：元
	CreatedAt int64   `json:"created_at"`                            // 记录创建时间戳
}

// TableName 指定表名
func (DailyData) TableName() string {
	return "daily_data"
}

// TechnicalIndicator 技术指标模型 - A股技术分析指标
type TechnicalIndicator struct {
	ID         uint    `json:"id" gorm:"primaryKey"`                  // 主键ID，数据库自增
	TsCode     string  `json:"ts_code" gorm:"size:20;not null;index"` // 股票代码，如：000001.SZ
	TradeDate  int     `json:"trade_date" gorm:"not null;index"`      // 交易日期，YYYYMMDD格式，如：20250910
	MA5        float64 `json:"ma5" gorm:"type:decimal(10,3)"`         // 5日移动平均线，单位：元，短期趋势指标
	MA10       float64 `json:"ma10" gorm:"type:decimal(10,3)"`        // 10日移动平均线，单位：元，短期趋势指标
	MA20       float64 `json:"ma20" gorm:"type:decimal(10,3)"`        // 20日移动平均线，单位：元，中期趋势指标
	MA60       float64 `json:"ma60" gorm:"type:decimal(10,3)"`        // 60日移动平均线，单位：元，长期趋势指标
	RSI        float64 `json:"rsi" gorm:"type:decimal(8,4)"`          // 相对强弱指数，范围0-100，>70超买，<30超卖
	MACD       float64 `json:"macd" gorm:"type:decimal(10,6)"`        // MACD指标，趋势跟踪指标，正值看涨，负值看跌
	MACDSignal float64 `json:"macd_signal" gorm:"type:decimal(10,6)"` // MACD信号线，MACD的EMA平滑线
	MACDHist   float64 `json:"macd_hist" gorm:"type:decimal(10,6)"`   // MACD柱状图，MACD-Signal，反映趋势变化
	KDJK       float64 `json:"kdj_k" gorm:"type:decimal(8,4)"`        // KDJ指标K值，范围0-100，随机指标
	KDJD       float64 `json:"kdj_d" gorm:"type:decimal(8,4)"`        // KDJ指标D值，范围0-100，K值的平滑线
	KDJJ       float64 `json:"kdj_j" gorm:"type:decimal(8,4)"`        // KDJ指标J值，3K-2D，敏感度最高
	CreatedAt  int64   `json:"created_at"`                            // 记录创建时间戳
}

// TableName 指定表名
func (TechnicalIndicator) TableName() string {
	return "technical_indicators"
}

// SelectionResult 选股结果模型 - A股选股策略执行结果
type SelectionResult struct {
	ID            uint      `json:"id" gorm:"primaryKey"`                        // 主键ID，数据库自增
	StrategyName  string    `json:"strategy_name" gorm:"size:50;not null;index"` // 选股策略名称，如：technical、fundamental、combined
	TsCode        string    `json:"ts_code" gorm:"size:20;not null;index"`       // 股票代码，如：000001.SZ
	SelectionDate time.Time `json:"selection_date" gorm:"not null;index"`        // 选股日期，策略执行日期
	Score         float64   `json:"score" gorm:"type:decimal(8,4)"`              // 选股评分，0-100分，分数越高越优质
	Reason        string    `json:"reason" gorm:"type:text"`                     // 选股理由，详细说明为什么选中该股票
	CreatedAt     time.Time `json:"created_at"`                                  // 记录创建时间

	// 关联股票信息
	Stock Stock `json:"stock" gorm:"foreignKey:TsCode;references:TsCode"` // 关联的股票基础信息
}

// TableName 指定表名
func (SelectionResult) TableName() string {
	return "selection_results"
}

// PerformanceReport 业绩报表模型 - A股上市公司业绩报表数据
type PerformanceReport struct {
	ID         uint      `json:"id" gorm:"primaryKey"`                  // 主键ID，数据库自增
	TsCode     string    `json:"ts_code" gorm:"size:20;not null;index"` // 股票代码，如：000001.SZ
	ReportDate time.Time `json:"report_date" gorm:"not null;index"`     // 报告期，如：2025-06-30

	// 每股收益相关
	EPS float64 `json:"eps" gorm:"type:decimal(10,4)"` // 每股收益，单位：元
	//EPSYoY    float64 `json:"eps_yoy" gorm:"type:decimal(8,4)"`     // 每股收益同比增长，单位：%
	WeightEPS float64 `json:"weight_eps" gorm:"type:decimal(10,4)"` // 加权每股收益，单位：元

	// 营业收入相关
	Revenue    float64 `json:"revenue" gorm:"type:decimal(20,2)"`    // 营业总收入，单位：元
	RevenueQoQ float64 `json:"revenue_qoq" gorm:"type:decimal(8,4)"` // 营业总收入同比增长，单位：%
	RevenueYoY float64 `json:"revenue_yoy" gorm:"type:decimal(8,4)"` // 营业总收入季度环比增长，单位：%

	// 净利润相关
	NetProfit    float64 `json:"net_profit" gorm:"type:decimal(20,2)"`    // 净利润，单位：元
	NetProfitQoQ float64 `json:"net_profit_qoq" gorm:"type:decimal(8,4)"` // 净利润同比增长，单位：%
	NetProfitYoY float64 `json:"net_profit_yoy" gorm:"type:decimal(8,4)"` // 净利润季度环比增长，单位：%

	// 每股净资产
	BVPS float64 `json:"bvps" gorm:"type:decimal(10,4)"` // 每股净资产，单位：元

	// 销售毛利率
	GrossMargin float64 `json:"gross_margin" gorm:"type:decimal(8,4)"` // 销售毛利率，单位：%

	// 股息率
	DividendYield float64 `json:"dividend_yield" gorm:"type:decimal(8,4)"` // 股息率，单位：%

	// 最新公告日期
	LatestAnnouncementDate *time.Time `json:"latest_announcement_date"` // 最新公告日期
	FirstAnnouncementDate  *time.Time `json:"first_announcement_date"`  // 首次公告日期

	CreatedAt time.Time `json:"created_at"` // 记录创建时间
	UpdatedAt time.Time `json:"updated_at"` // 记录更新时间
}

// TableName 指定表名
func (PerformanceReport) TableName() string {
	return "performance_reports"
}

// BacktestResult 回测结果模型 - A股策略回测数据
type BacktestResult struct {
	ID             uint      `json:"id" gorm:"primaryKey"`                      // 主键ID，数据库自增
	StrategyName   string    `json:"strategy_name" gorm:"size:50;not null"`     // 回测策略名称，如：technical、fundamental
	StartDate      time.Time `json:"start_date" gorm:"not null"`                // 回测开始日期
	EndDate        time.Time `json:"end_date" gorm:"not null"`                  // 回测结束日期
	InitialCapital float64   `json:"initial_capital" gorm:"type:decimal(20,2)"` // 初始资金，单位：元
	FinalCapital   float64   `json:"final_capital" gorm:"type:decimal(20,2)"`   // 最终资金，单位：元
	TotalReturn    float64   `json:"total_return" gorm:"type:decimal(8,4)"`     // 总收益率，单位：%，(最终资金-初始资金)/初始资金
	AnnualReturn   float64   `json:"annual_return" gorm:"type:decimal(8,4)"`    // 年化收益率，单位：%
	MaxDrawdown    float64   `json:"max_drawdown" gorm:"type:decimal(8,4)"`     // 最大回撤，单位：%，衡量风险水平
	SharpeRatio    float64   `json:"sharpe_ratio" gorm:"type:decimal(8,4)"`     // 夏普比率，风险调整后收益指标
	WinRate        float64   `json:"win_rate" gorm:"type:decimal(8,4)"`         // 胜率，单位：%，盈利交易次数/总交易次数
	CreatedAt      time.Time `json:"created_at"`                                // 记录创建时间
}

// TableName 指定表名
func (BacktestResult) TableName() string {
	return "backtest_results"
}

// ShareholderCount 股东户数模型 - A股上市公司股东户数数据
type ShareholderCount struct {
	ID              uint       `json:"id" gorm:"primaryKey"`                       // 主键ID，数据库自增
	TsCode          string     `json:"ts_code" gorm:"size:20;not null;index"`      // 股票代码，如：000001.SZ
	SecurityCode    string     `json:"security_code" gorm:"size:10;not null"`      // 证券代码，如：000001
	SecurityName    string     `json:"security_name" gorm:"size:100;not null"`     // 证券简称，如：平安银行
	EndDate         time.Time  `json:"end_date" gorm:"not null;index"`             // 统计截止日期，如：2025-06-30
	HolderNum       int64      `json:"holder_num" gorm:"not null"`                 // 股东户数，单位：户
	PreHolderNum    int64      `json:"pre_holder_num"`                             // 上期股东户数，单位：户
	HolderNumChange int64      `json:"holder_num_change"`                          // 股东户数变化，单位：户
	HolderNumRatio  float64    `json:"holder_num_ratio" gorm:"type:decimal(8,4)"`  // 股东户数变化比例，单位：%
	AvgMarketCap    float64    `json:"avg_market_cap" gorm:"type:decimal(15,2)"`   // 户均市值，单位：元
	AvgHoldNum      float64    `json:"avg_hold_num" gorm:"type:decimal(15,2)"`     // 户均持股数，单位：股
	TotalMarketCap  float64    `json:"total_market_cap" gorm:"type:decimal(20,2)"` // 总市值，单位：元
	TotalAShares    int64      `json:"total_a_shares"`                             // 总股本，单位：股
	IntervalChrate  float64    `json:"interval_chrate" gorm:"type:decimal(8,4)"`   // 区间涨跌幅，单位：%
	ChangeShares    int64      `json:"change_shares"`                              // 股本变动，单位：股
	ChangeReason    string     `json:"change_reason" gorm:"size:100"`              // 变动原因，如：发行融资
	HoldNoticeDate  *time.Time `json:"hold_notice_date"`                           // 公告日期
	PreEndDate      *time.Time `json:"pre_end_date"`                               // 上期截止日期
	CreatedAt       time.Time  `json:"created_at"`                                 // 记录创建时间
	UpdatedAt       time.Time  `json:"updated_at"`                                 // 记录更新时间

	// 关联股票信息
	Stock Stock `json:"stock" gorm:"foreignKey:TsCode;references:TsCode"` // 关联的股票基础信息
}

// TableName 指定表名
func (ShareholderCount) TableName() string {
	return "shareholder_counts"
}
