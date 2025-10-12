package model

import (
	"fmt"
	"strings"
	"time"
)

// Stock 股票基础信息模型 - A股市场
type Stock struct {
	TsCode    string     `json:"ts_code" gorm:"primaryKey;size:20;not null"` // Tushare股票代码，如：000001.SZ、600000.SH，主键
	Symbol    string     `json:"symbol" gorm:"size:10;not null"`             // 股票代码，如：000001、600000（不含交易所后缀）
	Name      string     `json:"name" gorm:"size:100;not null"`              // 股票简称，如：平安银行、浦发银行
	Area      string     `json:"area" gorm:"size:50"`                        // 所在地区，如：深圳、上海、北京
	Industry  string     `json:"industry" gorm:"size:100"`                   // 所属行业，如：银行、房地产开发、软件开发
	Market    string     `json:"market" gorm:"size:10"`                      // 交易市场，SZ=深交所、SH=上交所、BJ=北交所
	ListDate  *time.Time `json:"list_date"`                                  // 上市日期，首次公开发行日期
	IsActive  bool       `json:"is_active" gorm:"default:true"`              // 是否活跃交易，false表示停牌、退市等
	CreatedAt time.Time  `json:"created_at"`                                 // 记录创建时间
	UpdatedAt time.Time  `json:"updated_at"`                                 // 记录更新时间
}

// TableName 指定表名
func (Stock) TableName() string {
	return "stocks"
}

// DailyData 日线数据模型 - A股K线数据
type DailyData struct {
	TsCode    string    `json:"ts_code" gorm:"size:20;not null;primaryKey"` // 股票代码，如：000001.SZ，联合主键1
	TradeDate int       `json:"trade_date" gorm:"not null;primaryKey"`      // 交易日期，YYYYMMDD格式，如：20250910，联合主键2
	Open      float64   `json:"open" gorm:"type:decimal(10,3)"`             // 开盘价，单位：元
	High      float64   `json:"high" gorm:"type:decimal(10,3)"`             // 最高价，单位：元
	Low       float64   `json:"low" gorm:"type:decimal(10,3)"`              // 最低价，单位：元
	Close     float64   `json:"close" gorm:"type:decimal(10,3)"`            // 收盘价，单位：元
	Volume    int64     `json:"volume"`                                     // 成交量，单位：股（A股以股为单位）
	Amount    float64   `json:"amount" gorm:"type:decimal(20,2)"`           // 成交额，单位：元
	CreatedAt time.Time `json:"created_at"`                                 // 记录创建时间戳
	UpdatedAt time.Time `json:"updated_at"`                                 // 记录更新时间戳
}

// Get4Price 获取最高、最低、开盘、收盘价
func (d DailyData) Get4Price() (float64, float64, float64, float64) {
	return d.High, d.Low, d.Open, d.Close
}

// GetSymbol 获取股票代码
func (d DailyData) GetSymbol() string {
	return strings.Split(d.TsCode, ".")[0]
}

// GetTradeDate 获取交易日期
func (d DailyData) GetTradeDate() int {
	return d.TradeDate
}

// TableName 指定表名 - 根据股票代码动态选择表名
func (d DailyData) TableName() string {
	// 提取股票代码的前三位数字
	var prefix string
	if len(d.TsCode) >= 3 {
		// 处理带后缀的情况，如 "000001.SZ"
		code := strings.Split(d.TsCode, ".")[0]
		if len(code) >= 3 {
			prefix = code[:3]
		}
	}

	// 根据前三位确定表名
	switch prefix {
	case "000", "001", "002", "300", "301", "600", "601", "603", "605", "688":
		return fmt.Sprintf("daily_data_%s", prefix)
	default:
		return "daily_data_other"
	}
}

// getExchange 根据股票代码获取交易所类型
func (d DailyData) getExchange() string {
	tsCode := d.TsCode
	if len(tsCode) == 0 {
		return "SH" // 默认上海
	}

	// 检查后缀
	if len(tsCode) > 3 {
		if tsCode[len(tsCode)-3:] == ".SH" {
			return "SH"
		}
		if tsCode[len(tsCode)-3:] == ".SZ" {
			return "SZ"
		}
	}

	// 根据代码前缀判断
	if len(tsCode) > 0 {
		firstChar := tsCode[0]
		if firstChar == '6' {
			return "SH"
		}
		if firstChar == '0' || firstChar == '3' {
			return "SZ"
		}
	}

	return "SH" // 默认上海
}

// WeeklyData 周K线数据模型 - A股周K线行情数据
type WeeklyData struct {
	TsCode    string    `json:"ts_code" gorm:"size:20;not null;primaryKey"` // 股票代码，如：000001.SZ，联合主键1
	TradeDate int       `json:"trade_date" gorm:"not null;primaryKey"`      // 周结束交易日期，YYYYMMDD格式，如：20250910，联合主键2
	Open      float64   `json:"open" gorm:"type:decimal(10,3)"`             // 周开盘价，单位：元
	High      float64   `json:"high" gorm:"type:decimal(10,3)"`             // 周最高价，单位：元
	Low       float64   `json:"low" gorm:"type:decimal(10,3)"`              // 周最低价，单位：元
	Close     float64   `json:"close" gorm:"type:decimal(10,3)"`            // 周收盘价，单位：元
	Volume    int64     `json:"volume"`                                     // 周成交量，单位：股
	Amount    float64   `json:"amount" gorm:"type:decimal(20,2)"`           // 周成交额，单位：元
	CreatedAt time.Time `json:"created_at"`                                 // 记录创建时间戳
	UpdatedAt time.Time `json:"updated_at"`                                 // 记录更新时间戳
}

// Get4Price 获取最高、最低、开盘、收盘价
func (w WeeklyData) Get4Price() (float64, float64, float64, float64) {
	return w.High, w.Low, w.Open, w.Close
}

// GetSymbol 获取股票代码
func (w WeeklyData) GetSymbol() string {
	return strings.Split(w.TsCode, ".")[0]
}

// GetTradeDate 获取交易日期
func (w WeeklyData) GetTradeDate() int {
	return w.TradeDate
}

// TableName 指定表名 - 根据股票代码动态选择表名
func (w WeeklyData) TableName() string {
	// 提取股票代码的前三位数字
	var prefix string
	if len(w.TsCode) >= 3 {
		// 处理带后缀的情况，如 "000001.SZ"
		code := strings.Split(w.TsCode, ".")[0]
		if len(code) >= 3 {
			prefix = code[:3]
		}
	}

	// 根据前三位确定表名
	switch prefix {
	case "000", "001", "002", "300", "301", "600", "601", "603", "605", "688":
		return fmt.Sprintf("weekly_data_%s", prefix)
	default:
		return "weekly_data_other"
	}
}

// MonthlyData 月K线数据模型 - A股月K线行情数据
type MonthlyData struct {
	TsCode    string    `json:"ts_code" gorm:"size:20;not null;primaryKey"` // 股票代码，如：000001.SZ，联合主键1
	TradeDate int       `json:"trade_date" gorm:"not null;primaryKey"`      // 月结束交易日期，YYYYMMDD格式，如：20250930，联合主键2
	Open      float64   `json:"open" gorm:"type:decimal(10,3)"`             // 月开盘价，单位：元
	High      float64   `json:"high" gorm:"type:decimal(10,3)"`             // 月最高价，单位：元
	Low       float64   `json:"low" gorm:"type:decimal(10,3)"`              // 月最低价，单位：元
	Close     float64   `json:"close" gorm:"type:decimal(10,3)"`            // 月收盘价，单位：元
	Volume    int64     `json:"volume"`                                     // 月成交量，单位：股
	Amount    float64   `json:"amount" gorm:"type:decimal(20,2)"`           // 月成交额，单位：元
	CreatedAt time.Time `json:"created_at"`                                 // 记录创建时间戳
	UpdatedAt time.Time `json:"updated_at"`                                 // 记录更新时间戳
}

// Get4Price 获取最高、最低、开盘、收盘价
func (m MonthlyData) Get4Price() (float64, float64, float64, float64) {
	return m.High, m.Low, m.Open, m.Close
}

// GetSymbol 获取股票代码
func (m MonthlyData) GetSymbol() string {
	return strings.Split(m.TsCode, ".")[0]
}

// GetTradeDate 获取交易日期
func (m MonthlyData) GetTradeDate() int {
	return m.TradeDate
}

// TableName 指定表名 - 根据股票代码动态选择表名
func (m MonthlyData) TableName() string {
	// 提取股票代码的前三位数字
	var prefix string
	if len(m.TsCode) >= 3 {
		// 处理带后缀的情况，如 "000001.SZ"
		code := strings.Split(m.TsCode, ".")[0]
		if len(code) >= 3 {
			prefix = code[:3]
		}
	}

	// 根据前三位确定表名
	switch prefix {
	case "000", "001", "002", "300", "301", "600", "601", "603", "605", "688":
		return fmt.Sprintf("monthly_data_%s", prefix)
	default:
		return "monthly_data_other"
	}
}

// QuarterlyData 季K线数据模型 - A股季K线行情数据
type QuarterlyData struct {
	TsCode    string    `json:"ts_code" gorm:"size:20;not null;primaryKey"` // 股票代码，如：000001.SZ，联合主键1
	TradeDate int       `json:"trade_date" gorm:"not null;primaryKey"`      // 季结束交易日期，YYYYMMDD格式，如：20250930，联合主键2
	Open      float64   `json:"open" gorm:"type:decimal(10,3)"`             // 季开盘价，单位：元
	High      float64   `json:"high" gorm:"type:decimal(10,3)"`             // 季最高价，单位：元
	Low       float64   `json:"low" gorm:"type:decimal(10,3)"`              // 季最低价，单位：元
	Close     float64   `json:"close" gorm:"type:decimal(10,3)"`            // 季收盘价，单位：元
	Volume    int64     `json:"volume"`                                     // 季成交量，单位：股
	Amount    float64   `json:"amount" gorm:"type:decimal(20,2)"`           // 季成交额，单位：元
	CreatedAt time.Time `json:"created_at"`                                 // 记录创建时间戳
	UpdatedAt time.Time `json:"updated_at"`                                 // 记录更新时间戳
}

// TableName 指定表名
func (QuarterlyData) TableName() string {
	return "quarterly_data"
}

// YearlyData 年K线数据模型 - A股年K线行情数据
type YearlyData struct {
	TsCode    string    `json:"ts_code" gorm:"size:20;not null;primaryKey"` // 股票代码，如：000001.SZ，联合主键1
	TradeDate int       `json:"trade_date" gorm:"not null;primaryKey"`      // 年结束交易日期，YYYYMMDD格式，如：20251231，联合主键2
	Open      float64   `json:"open" gorm:"type:decimal(10,3)"`             // 年开盘价，单位：元
	High      float64   `json:"high" gorm:"type:decimal(10,3)"`             // 年最高价，单位：元
	Low       float64   `json:"low" gorm:"type:decimal(10,3)"`              // 年最低价，单位：元
	Close     float64   `json:"close" gorm:"type:decimal(10,3)"`            // 年收盘价，单位：元
	Volume    int64     `json:"volume"`                                     // 年成交量，单位：股
	Amount    float64   `json:"amount" gorm:"type:decimal(20,2)"`           // 年成交额，单位：元
	CreatedAt time.Time `json:"created_at"`                                 // 记录创建时间戳
	UpdatedAt time.Time `json:"updated_at"`                                 // 记录更新时间戳
}

// Get4Price 获取最高、最低、开盘、收盘价
func (y YearlyData) Get4Price() (float64, float64, float64, float64) {
	return y.High, y.Low, y.Open, y.Close
}

// GetSymbol 获取股票代码
func (y YearlyData) GetSymbol() string {
	return strings.Split(y.TsCode, ".")[0]
}

// GetTradeDate 获取交易日期
func (y YearlyData) GetTradeDate() int {
	return y.TradeDate
}

// TableName 指定表名
func (YearlyData) TableName() string {
	return "yearly_data"
}

// TechnicalIndicatorPeriod 技术指标周期类型
type TechnicalIndicatorPeriod string

const (
	TechnicalIndicatorPeriodDaily   TechnicalIndicatorPeriod = "daily"
	TechnicalIndicatorPeriodWeekly  TechnicalIndicatorPeriod = "weekly"
	TechnicalIndicatorPeriodMonthly TechnicalIndicatorPeriod = "monthly"
	TechnicalIndicatorPeriodYearly  TechnicalIndicatorPeriod = "yearly"
)

// TechnicalIndicator 技术指标模型 - A股技术分析指标
type TechnicalIndicator struct {
	Symbol    string  `json:"symbol" gorm:"column:symbol;size:10;not null;primaryKey"` // 股票代码，如：000001，联合主键1
	TradeDate int     `json:"trade_date" gorm:"column:trade_date;not null;primaryKey"` // 交易日期，YYYYMMDD格式，如：20250910，联合主键2
	Ma5       float64 `json:"ma5" gorm:"column:ma5;type:decimal(10,3)"`                // 5期移动平均线，单位：元，短期趋势指标
	Ma10      float64 `json:"ma10" gorm:"column:ma10;type:decimal(10,3)"`              // 10期移动平均线，单位：元，短期趋势指标
	Ma20      float64 `json:"ma20" gorm:"column:ma20;type:decimal(10,3)"`              // 20期移动平均线，单位：元，中期趋势指标
	Ma60      float64 `json:"ma60" gorm:"column:ma60;type:decimal(10,3)"`              // 60期移动平均线，单位：元，长期趋势指标
	Rsi6      float64 `json:"rsi6" gorm:"column:rsi6;type:decimal(8,4)"`               // 6期相对强弱指数，范围0-100，>70超买，<30超卖
	Rsi12     float64 `json:"rsi12" gorm:"column:rsi12;type:decimal(8,4)"`             // 12期相对强弱指数，范围0-100，>70超买，<30超卖
	Rsi24     float64 `json:"rsi24" gorm:"column:rsi24;type:decimal(8,4)"`             // 24期相对强弱指数，范围0-100，>70超买，<30超卖
	Macd      float64 `json:"macd" gorm:"column:macd;type:decimal(10,6)"`              // Macd指标，趋势跟踪指标，正值看涨，负值看跌
	MacdEma1  float64 `json:"macd_ema1" gorm:"column:macd_ema1;type:decimal(10,6)"`    // Macd Ema1
	MacdEma2  float64 `json:"macd_ema2" gorm:"column:macd_ema2;type:decimal(10,6)"`    // Macd Ema2
	MacdDif   float64 `json:"macd_dif" gorm:"column:macd_dif;type:decimal(10,6)"`      // Macd DIF线，快线减慢线的差值
	MacdDea   float64 `json:"macd_dea" gorm:"column:macd_dea;type:decimal(10,6)"`      // Macd DEA线，DIF的EMA平滑线
	KdjK      float64 `json:"kdj_k" gorm:"column:kdj_k;type:decimal(8,4)"`             // Kdj指标K值，范围0-100，随机指标
	KdjD      float64 `json:"kdj_d" gorm:"column:kdj_d;type:decimal(8,4)"`             // Kdj指标D值，范围0-100，K值的平滑线
	KdjJ      float64 `json:"kdj_j" gorm:"column:kdj_j;type:decimal(8,4)"`             // Kdj指标J值，3K-2D，敏感度最高

	// 内部字段，用于动态表名
	Period TechnicalIndicatorPeriod `json:"-" gorm:"-"` // 周期类型，不存储到数据库
}

// TableName 动态指定表名
func (t TechnicalIndicator) TableName() string {
	switch t.Period {
	case TechnicalIndicatorPeriodDaily:
		return "daily_technical_indicators"
	case TechnicalIndicatorPeriodWeekly:
		return "weekly_technical_indicators"
	case TechnicalIndicatorPeriodMonthly:
		return "monthly_technical_indicators"
	case TechnicalIndicatorPeriodYearly:
		return "yearly_technical_indicators"
	default:
		return "daily_technical_indicators" // 默认使用日表
	}
}

// NewTechnicalIndicator 创建技术指标实例
func NewTechnicalIndicator(period TechnicalIndicatorPeriod) *TechnicalIndicator {
	return &TechnicalIndicator{Period: period}
}

// NewDailyTechnicalIndicator 创建日技术指标实例
func NewDailyTechnicalIndicator() *TechnicalIndicator {
	return NewTechnicalIndicator(TechnicalIndicatorPeriodDaily)
}

// NewWeeklyTechnicalIndicator 创建周技术指标实例
func NewWeeklyTechnicalIndicator() *TechnicalIndicator {
	return NewTechnicalIndicator(TechnicalIndicatorPeriodWeekly)
}

// NewMonthlyTechnicalIndicator 创建月技术指标实例
func NewMonthlyTechnicalIndicator() *TechnicalIndicator {
	return NewTechnicalIndicator(TechnicalIndicatorPeriodMonthly)
}

// NewYearlyTechnicalIndicator 创建年技术指标实例
func NewYearlyTechnicalIndicator() *TechnicalIndicator {
	return NewTechnicalIndicator(TechnicalIndicatorPeriodYearly)
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
	TsCode     string `json:"ts_code" gorm:"column:ts_code;size:20;not null;primaryKey"` // 股票代码，如：000001.SZ，联合主键1
	ReportDate int    `json:"report_date" gorm:"column:report_date;not null;primaryKey"` // 报告期，YYYYMMDD格式，如：20250630，联合主键2

	// 每股收益相关
	EPS       float64 `json:"eps" gorm:"column:eps;type:decimal(10,4)"`               // 每股收益，单位：元
	WeightEPS float64 `json:"weight_eps" gorm:"column:weight_eps;type:decimal(10,4)"` // 加权每股收益，单位：元

	// 营业收入相关
	Revenue    float64 `json:"revenue" gorm:"column:revenue;type:decimal(20,2)"`        // 营业总收入，单位：元
	RevenueQoQ float64 `json:"revenue_qoq" gorm:"column:revenue_qoq;type:decimal(8,4)"` // 营业总收入同比增长，单位：%
	RevenueYoY float64 `json:"revenue_yoy" gorm:"column:revenue_yoy;type:decimal(8,4)"` // 营业总收入季度环比增长，单位：%

	// 净利润相关
	NetProfit    float64 `json:"net_profit" gorm:"column:net_profit;type:decimal(20,2)"`        // 净利润，单位：元
	NetProfitQoQ float64 `json:"net_profit_qoq" gorm:"column:net_profit_qoq;type:decimal(8,4)"` // 净利润同比增长，单位：%
	NetProfitYoY float64 `json:"net_profit_yoy" gorm:"column:net_profit_yoy;type:decimal(8,4)"` // 净利润季度环比增长，单位：%

	// 每股净资产
	BVPS float64 `json:"bvps" gorm:"column:bvps;type:decimal(10,4)"` // 每股净资产，单位：元

	// 销售毛利率
	GrossMargin float64 `json:"gross_margin" gorm:"column:gross_margin;type:decimal(8,4)"` // 销售毛利率，单位：%

	// 股息率
	DividendYield float64 `json:"dividend_yield" gorm:"column:dividend_yield;type:decimal(8,4)"` // 股息率，单位：%

	// 最新公告日期
	LatestAnnouncementDate *time.Time `json:"latest_announcement_date" gorm:"column:latest_announcement_date;type:datetime(3)"` // 最新公告日期
	FirstAnnouncementDate  *time.Time `json:"first_announcement_date" gorm:"column:first_announcement_date;type:datetime(3)"`   // 首次公告日期

	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;type:datetime(3)"` // 记录创建时间
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at;type:datetime(3)"` // 记录更新时间
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
	TsCode          string     `json:"ts_code" gorm:"column:ts_code;size:20;not null;primaryKey"`          // 股票代码，如：000001.SZ，联合主键1
	EndDate         int        `json:"end_date" gorm:"column:end_date;not null;primaryKey"`                // 统计截止日期，YYYYMMDD格式，如：20250630，联合主键2
	SecurityCode    string     `json:"security_code" gorm:"column:security_code;size:10;not null"`         // 证券代码，如：000001
	SecurityName    string     `json:"security_name" gorm:"column:security_name;size:100;not null"`        // 证券简称，如：平安银行
	HolderNum       int64      `json:"holder_num" gorm:"column:holder_num;not null"`                       // 股东户数，单位：户
	PreHolderNum    int64      `json:"pre_holder_num" gorm:"column:pre_holder_num"`                        // 上期股东户数，单位：户
	HolderNumChange int64      `json:"holder_num_change" gorm:"column:holder_num_change"`                  // 股东户数变化，单位：户
	HolderNumRatio  float64    `json:"holder_num_ratio" gorm:"column:holder_num_ratio;type:decimal(8,4)"`  // 股东户数变化比例，单位：%
	AvgMarketCap    float64    `json:"avg_market_cap" gorm:"column:avg_market_cap;type:decimal(15,2)"`     // 户均市值，单位：元
	AvgHoldNum      float64    `json:"avg_hold_num" gorm:"column:avg_hold_num;type:decimal(15,2)"`         // 户均持股数，单位：股
	TotalMarketCap  float64    `json:"total_market_cap" gorm:"column:total_market_cap;type:decimal(20,2)"` // 总市值，单位：元
	TotalAShares    int64      `json:"total_a_shares" gorm:"column:total_a_shares"`                        // 总股本，单位：股
	IntervalChrate  float64    `json:"interval_chrate" gorm:"column:interval_chrate;type:decimal(8,4)"`    // 区间涨跌幅，单位：%
	ChangeShares    int64      `json:"change_shares" gorm:"column:change_shares"`                          // 股本变动，单位：股
	ChangeReason    string     `json:"change_reason" gorm:"column:change_reason;size:100"`                 // 变动原因，如：发行融资
	HoldNoticeDate  *time.Time `json:"hold_notice_date" gorm:"column:hold_notice_date;type:datetime(3)"`   // 公告日期
	PreEndDate      *int       `json:"pre_end_date" gorm:"column:pre_end_date"`                            // 上期截止日期，YYYYMMDD格式，如：20250331
	CreatedAt       time.Time  `json:"created_at" gorm:"column:created_at;type:datetime(3)"`               // 记录创建时间
	UpdatedAt       time.Time  `json:"updated_at" gorm:"column:updated_at;type:datetime(3)"`               // 记录更新时间

	// 关联股票信息
	Stock Stock `json:"stock" gorm:"foreignKey:TsCode;references:TsCode"` // 关联的股票基础信息
}

// TableName 指定表名
func (ShareholderCount) TableName() string {
	return "shareholder_counts"
}
