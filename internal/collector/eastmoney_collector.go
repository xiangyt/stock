package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"stock/internal/logger"
	"stock/internal/model"

	"golang.org/x/time/rate"
)

type KLineType = string

// K线类型常量
const (
	KLineTypeDaily   KLineType = "101" // 日K线
	KLineTypeWeekly  KLineType = "102" // 周K线
	KLineTypeMonthly KLineType = "103" // 月K线
	KLineTypeYearly  KLineType = "104" // 年K线
)

// KLineData 通用K线数据接口
type KLineData interface {
	model.WeeklyData | model.MonthlyData | model.YearlyData
}

// KLineResponse 东方财富K线API响应结构
type KLineResponse struct {
	RC   int `json:"rc"`
	RT   int `json:"rt"`
	Data struct {
		Code   string   `json:"code"`
		Market int      `json:"market"`
		Name   string   `json:"name"`
		Klines []string `json:"klines"`
	} `json:"data"`
}

// EastMoneyCollector 东方财富数据采集器
type EastMoneyCollector struct {
	BaseCollector
	client  *http.Client
	logger  *logger.Logger
	parser  *KLineParser
	limiter *rate.Limiter // 限流器
}

// NewEastMoneyCollector 创建东方财富采集器
func NewEastMoneyCollector(logger *logger.Logger) *EastMoneyCollector {
	config := CollectorConfig{
		Name:      "eastmoney",
		BaseURL:   "https://push2.eastmoney.com",
		Timeout:   30 * time.Second,
		RateLimit: 10, // 每秒10个请求
		Headers: map[string]string{
			"Accept":             "*/*",
			"Accept-Language":    "zh-CN,zh;q=0.9,en;q=0.8",
			"Connection":         "keep-alive",
			"Referer":            "https://data.eastmoney.com/",
			"User-Agent":         "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36",
			"sec-ch-ua":          `"Not;A=Brand";v="99", "Google Chrome";v="139", "Chromium";v="139"`,
			"sec-ch-ua-mobile":   "?0",
			"sec-ch-ua-platform": "macOS",
			"sec-fetch-dest":     "script",
			"sec-fetch-mode":     "no-cors",
			"sec-fetch-site":     "same-site",
			//"cookie":             "qgqp_b_id=70e1191db491f7e84374e18218beb159; st_nvi=t19Iuw0pv7cziS9NpDFvwac78; nid=00c18f59b20816388614a11f44a7a467; nid_create_time=1755334507172; gvi=SPp0K7jZ5OHSqiz2pigL09303; gvi_create_time=1755334507172; st_si=67468401804019; fullscreengg=1; fullscreengg2=1; websitepoptg_api_time=1758373501952; st_asi=delete; wsc_checkuser_ok=1; st_pvi=30544878065704; st_sp=2025-08-16%2016%3A55%3A06; st_inirUrl=https%3A%2F%2Fdata.eastmoney.com%2Fgphg%2F; st_sn=166; st_psi=20250921003620276-113300300813-8448178144",
		},
	}

	// 创建限流器，每秒允许 RateLimit 个请求，突发容量为 RateLimit*2
	limiter := rate.NewLimiter(rate.Limit(config.RateLimit), config.RateLimit*2)

	return &EastMoneyCollector{
		BaseCollector: BaseCollector{
			Config:    config,
			Connected: false,
		},
		client: &http.Client{
			Timeout: config.Timeout,
		},
		logger:  logger,
		parser:  NewKLineParser(),
		limiter: limiter,
	}
}

// Connect 连接数据源
func (e *EastMoneyCollector) Connect() error {
	e.logger.Infof("Connecting to %s...", e.Config.Name)

	// 测试连接
	if err := e.testConnection(); err != nil {
		e.logger.Errorf("Failed to connect to %s: %v", e.Config.Name, err)
		return err
	}

	e.Connected = true
	e.logger.Infof("Successfully connected to %s", e.Config.Name)
	return nil
}

// testConnection 测试连接
func (e *EastMoneyCollector) testConnection() error {
	// 获取少量数据测试连接
	return nil
}

// makeRequest 发送HTTP请求（带限流）
func (e *EastMoneyCollector) makeRequest(url string) (*http.Response, error) {
	return e.makeRequestWithContext(context.Background(), url)
}

// makeRequestWithContext 发送HTTP请求（带限流和上下文）
func (e *EastMoneyCollector) makeRequestWithContext(ctx context.Context, url string) (*http.Response, error) {
	// 应用限流
	if err := e.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait failed: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// 添加请求头
	for key, value := range e.Config.Headers {
		req.Header.Set(key, value)
	}

	e.logger.Debugf("Making rate-limited request: %s", url)

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return resp, nil
}

// SetRateLimit 动态设置限流速率
func (e *EastMoneyCollector) SetRateLimit(requestsPerSecond int) {
	if requestsPerSecond <= 0 {
		requestsPerSecond = 1 // 最小值为1
	}

	e.Config.RateLimit = requestsPerSecond
	e.limiter.SetLimit(rate.Limit(requestsPerSecond))
	e.limiter.SetBurst(requestsPerSecond * 2) // 突发容量为速率的2倍

	e.logger.Infof("Rate limit updated to %d requests/second", requestsPerSecond)
}

// GetRateLimit 获取当前限流速率
func (e *EastMoneyCollector) GetRateLimit() int {
	return e.Config.RateLimit
}

// GetRateLimitStats 获取限流统计信息
func (e *EastMoneyCollector) GetRateLimitStats() map[string]interface{} {
	return map[string]interface{}{
		"rate_limit":    e.Config.RateLimit,
		"current_limit": float64(e.limiter.Limit()),
		"burst_size":    e.limiter.Burst(),
		"tokens":        e.limiter.Tokens(), // 当前可用令牌数
	}
}

// EastMoneyStockListResponse 东方财富股票列表响应结构
type EastMoneyStockListResponse struct {
	RC     int    `json:"rc"`
	RT     int    `json:"rt"`
	SVRID  int    `json:"svrid"`
	LT     int    `json:"lt"`
	FULL   int    `json:"full"`
	DLMKTS string `json:"dlmkts"`
	Data   struct {
		Total int `json:"total"`
		Diff  []struct {
			F1   interface{} `json:"f1"`   // 未知字段
			F2   interface{} `json:"f2"`   // 最新价
			F3   interface{} `json:"f3"`   // 涨跌幅
			F12  string      `json:"f12"`  // 股票代码
			F13  int         `json:"f13"`  // 市场标识 0=深市 1=沪市
			F14  string      `json:"f14"`  // 股票名称
			F62  interface{} `json:"f62"`  // 主力净流入
			F66  interface{} `json:"f66"`  // 超大单净流入
			F69  interface{} `json:"f69"`  // 超大单净流入占比
			F72  interface{} `json:"f72"`  // 大单净流入
			F75  interface{} `json:"f75"`  // 大单净流入占比
			F78  interface{} `json:"f78"`  // 中单净流入
			F81  interface{} `json:"f81"`  // 中单净流入占比
			F84  interface{} `json:"f84"`  // 小单净流入
			F87  interface{} `json:"f87"`  // 小单净流入占比
			F124 interface{} `json:"f124"` // 更新时间戳
			F184 interface{} `json:"f184"` // 主力净流入占比
			F204 interface{} `json:"f204"` // 5日主力净流入
			F205 interface{} `json:"f205"` // 10日主力净流入
		} `json:"diff"`
	} `json:"data"`
}

// fetchStockListPage 获取股票列表分页数据
func (e *EastMoneyCollector) fetchStockListPage(page, pageSize int) (*EastMoneyStockListResponse, error) {
	// 构建请求URL
	baseURL := "https://push2.eastmoney.com/api/qt/clist/get"
	params := url.Values{}

	// 基础参数
	params.Set("cb", fmt.Sprintf("jQuery112303251051388385584_%d", time.Now().UnixMilli()))
	params.Set("fid", "f62")
	params.Set("po", "1")
	params.Set("pz", strconv.Itoa(pageSize))
	params.Set("pn", strconv.Itoa(page))
	params.Set("np", "1")
	params.Set("fltt", "2")
	params.Set("invt", "2")
	params.Set("ut", "8dec03ba335b81bf4ebdf7b29ec27d15")

	// 市场筛选参数 - 所有A股
	params.Set("fs", "m:0+t:6+f:!2,m:0+t:13+f:!2,m:0+t:80+f:!2,m:1+t:2+f:!2,m:1+t:23+f:!2,m:0+t:7+f:!2,m:1+t:3+f:!2")

	// 返回字段
	params.Set("fields", "f12,f14,f2,f3,f62,f184,f66,f69,f72,f75,f78,f81,f84,f87,f204,f205,f124,f1,f13")

	requestURL := baseURL + "?" + params.Encode()

	resp, err := e.makeRequest(requestURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 东方财富返回的是JSONP格式，需要提取JSON部分
	bodyStr := string(body)

	// 使用正则表达式提取JSON部分
	re := regexp.MustCompile(`jQuery\d+_\d+\((.*)\)`)
	matches := re.FindStringSubmatch(bodyStr)
	if len(matches) < 2 {
		return nil, fmt.Errorf("invalid JSONP response format")
	}

	jsonStr := matches[1]

	var result EastMoneyStockListResponse
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	return &result, nil
}

// parseFloat 安全地将interface{}转换为float64
func parseFloat(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return val
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
		return 0
	case int:
		return float64(val)
	case int64:
		return float64(val)
	default:
		return 0
	}
}

// parseTimeString 安全地将字符串转换为时间，支持多种时间格式
// 返回解析后的时间和是否成功的标志
func parseTimeString(timeStr string) (time.Time, bool) {
	if timeStr == "" {
		return time.Time{}, false
	}

	// 尝试多种时间格式进行解析
	formats := []string{
		"2006-01-02T15:04:05", // ISO 8601 格式
		"2006-01-02 15:04:05", // 标准日期时间格式
		"2006-01-02",          // 日期格式
		"20060102",            // 紧凑日期格式
		"2006/01/02",          // 斜杠分隔日期格式
		"2006/01/02 15:04:05", // 斜杠分隔日期时间格式
		time.DateTime,         // Go标准日期时间格式
		time.DateOnly,         // Go标准日期格式
	}

	for _, format := range formats {
		if parsedTime, err := time.ParseInLocation(format, timeStr, time.Local); err == nil {
			return parsedTime, true
		}
	}

	return time.Time{}, false
}

// parseTimeToInt 将时间字符串解析为 YYYYMMDD 格式的整数
func parseTimeToInt(timeStr string) (int, bool) {
	if timeStr == "" {
		return 0, false
	}

	// 先使用现有的 parseTimeString 函数解析时间
	if parsedTime, success := parseTimeString(timeStr); success {
		// 转换为 YYYYMMDD 格式的整数
		year := parsedTime.Year()
		month := int(parsedTime.Month())
		day := parsedTime.Day()
		dateInt := year*10000 + month*100 + day
		return dateInt, true
	}

	return 0, false
}

// GetStockList 获取股票列表
func (e *EastMoneyCollector) GetStockList() ([]model.Stock, error) {
	e.logger.Info("Fetching stock list from EastMoney...")

	var allStocks []model.Stock
	page := 1
	pageSize := 50 // 每次只获取50条，避免API限制

	for {
		e.logger.Infof("Fetching page %d...", page)

		response, err := e.fetchStockListPage(page, pageSize)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch page %d: %v", page, err)
		}

		if response.RC != 0 {
			return nil, fmt.Errorf("API error: rc=%d", response.RC)
		}

		if len(response.Data.Diff) == 0 {
			break // 没有更多数据
		}

		// 记录API返回的总数信息
		if page == 1 {
			e.logger.Infof("API reports total stocks: %d", response.Data.Total)
		}

		// 转换数据格式
		for _, item := range response.Data.Diff {
			// 确定市场
			var market string
			switch item.F13 {
			case 0:
				market = "SZ" // 深市
			case 1:
				market = "SH" // 沪市
			default:
				market = "UNKNOWN"
			}

			// 构建TsCode
			tsCode := fmt.Sprintf("%s.%s", item.F12, market)

			stock := model.Stock{
				TsCode:    tsCode,
				Symbol:    item.F12,
				Name:      item.F14,
				Market:    market,
				IsActive:  true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			// 根据股票代码判断板块和地区
			if len(item.F12) >= 3 {
				switch {
				case strings.HasPrefix(item.F12, "000"), strings.HasPrefix(item.F12, "001"), strings.HasPrefix(item.F12, "002"):
					stock.Industry = "主板" // 深市主板/中小板
					stock.Area = "深圳"
				case strings.HasPrefix(item.F12, "300"):
					stock.Industry = "创业板"
					stock.Area = "深圳"
				case strings.HasPrefix(item.F12, "600"), strings.HasPrefix(item.F12, "601"), strings.HasPrefix(item.F12, "603"), strings.HasPrefix(item.F12, "605"):
					stock.Industry = "主板" // 沪市主板
					stock.Area = "上海"
				case strings.HasPrefix(item.F12, "688"):
					stock.Industry = "科创板"
					stock.Area = "上海"
				case strings.HasPrefix(item.F12, "8"), strings.HasPrefix(item.F12, "4"):
					stock.Industry = "北交所"
					stock.Area = "北京"
				default:
					stock.Industry = "其他"
					stock.Area = "未知"
				}
			}

			allStocks = append(allStocks, stock)
		}

		e.logger.Infof("Fetched %d stocks from page %d", len(response.Data.Diff), page)

		// 如果返回的数据少于页面大小，说明已经是最后一页
		if len(response.Data.Diff) < pageSize {
			break
		}

		page++

		// 添加延迟避免请求过快
		time.Sleep(100 * time.Millisecond)
	}

	e.logger.Infof("Total fetched %d stocks from EastMoney", len(allStocks))
	return allStocks, nil
}

// GetStockDetail 获取股票详情
func (e *EastMoneyCollector) GetStockDetail(tsCode string) (*model.Stock, error) {
	e.logger.Infof("Fetching stock detail for %s from EastMoney", tsCode)

	// 从tsCode解析股票代码和市场
	parts := strings.Split(tsCode, ".")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid tsCode format: %s", tsCode)
	}

	symbol := parts[0]
	market := parts[1]

	// 构建请求URL - 获取股票基本信息
	baseURL := "https://push2.eastmoney.com/api/qt/stock/get"
	params := url.Values{}
	params.Set("ut", "b2884a393a59ad64002292a3e90d46a5")
	params.Set("invt", "2")
	params.Set("fltt", "2")
	params.Set("cb", fmt.Sprintf("jQuery112309283015113892927_%d", time.Now().UnixMilli()))

	// 设置股票代码，东方财富格式：市场代码.股票代码
	var secid string
	if market == "SH" {
		secid = fmt.Sprintf("1.%s", symbol)
	} else if market == "SZ" {
		secid = fmt.Sprintf("0.%s", symbol)
	} else {
		return nil, fmt.Errorf("unsupported market: %s", market)
	}
	params.Set("secid", secid)

	// 返回字段 - 使用你提供的字段列表
	params.Set("fields", "f57,f58,f107,f43,f169,f170,f171,f47,f48,f60,f46,f44,f45,f168,f50,f162,f177,f803")

	requestURL := baseURL + "?" + params.Encode()

	resp, err := e.makeRequest(requestURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 解析JSONP响应
	bodyStr := string(body)
	// 尝试多种JSONP格式
	var jsonData string

	// 尝试jQuery格式
	re1 := regexp.MustCompile(`jQuery\d+_\d+\((.*)\)`)
	matches1 := re1.FindStringSubmatch(bodyStr)
	if len(matches1) >= 2 {
		jsonData = matches1[1]
	} else {
		// 尝试jsonp格式
		re2 := regexp.MustCompile(`jsonp\d+\((.*)\)`)
		matches2 := re2.FindStringSubmatch(bodyStr)
		if len(matches2) >= 2 {
			jsonData = matches2[1]
		} else {
			return nil, fmt.Errorf("invalid JSONP response format: %s", bodyStr[:100])
		}
	}

	var response struct {
		RC   int `json:"rc"`
		RT   int `json:"rt"`
		Data struct {
			F43  float64 `json:"f43"`  // 最新价
			F44  float64 `json:"f44"`  // 最高价
			F45  float64 `json:"f45"`  // 最低价
			F46  float64 `json:"f46"`  // 今开
			F47  int64   `json:"f47"`  // 成交量
			F48  float64 `json:"f48"`  // 成交额
			F50  float64 `json:"f50"`  // 量比
			F57  string  `json:"f57"`  // 股票代码
			F58  string  `json:"f58"`  // 股票名称
			F60  float64 `json:"f60"`  // 昨收
			F107 int     `json:"f107"` // 停牌状态
			F162 float64 `json:"f162"` // 涨跌幅
			F168 float64 `json:"f168"` // 换手率
			F169 float64 `json:"f169"` // 涨跌额
			F170 float64 `json:"f170"` // 市盈率动
			F171 float64 `json:"f171"` // 市净率
			F177 int     `json:"f177"` // 流通股本
			F803 string  `json:"f803"` // 板块
		} `json:"data"`
	}

	if err := json.Unmarshal([]byte(jsonData), &response); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	if response.RC != 0 {
		return nil, fmt.Errorf("API error: rc=%d", response.RC)
	}

	stock := &model.Stock{
		TsCode:   tsCode,
		Symbol:   symbol,
		Name:     response.Data.F58,
		Market:   market,
		IsActive: true,
	}

	// 根据股票代码判断板块和地区
	if len(symbol) >= 3 {
		switch {
		case strings.HasPrefix(symbol, "000"), strings.HasPrefix(symbol, "001"), strings.HasPrefix(symbol, "002"):
			stock.Industry = "主板"
			stock.Area = "深圳"
		case strings.HasPrefix(symbol, "300"):
			stock.Industry = "创业板"
			stock.Area = "深圳"
		case strings.HasPrefix(symbol, "600"), strings.HasPrefix(symbol, "601"), strings.HasPrefix(symbol, "603"), strings.HasPrefix(symbol, "605"):
			stock.Industry = "主板"
			stock.Area = "上海"
		case strings.HasPrefix(symbol, "688"):
			stock.Industry = "科创板"
			stock.Area = "上海"
		default:
			stock.Industry = "其他"
			stock.Area = "未知"
		}
	}

	return stock, nil
}

// GetStockData 获取股票历史数据 (兼容旧接口)
func (e *EastMoneyCollector) GetStockData(tsCode string, startDate, endDate time.Time) ([]model.DailyData, error) {
	return e.GetDailyKLine(tsCode, startDate, endDate)
}

// GetRealtimeData 获取实时数据
func (e *EastMoneyCollector) GetRealtimeData(tsCodes []string) ([]model.DailyData, error) {
	e.logger.Infof("Fetching realtime data for %d stocks from EastMoney", len(tsCodes))

	// 东方财富的实时数据可以通过股票列表API获取
	// 这里简化处理，获取前50只股票的实时数据
	response, err := e.fetchStockListPage(1, 50)
	if err != nil {
		return nil, err
	}

	var realtimeData []model.DailyData
	now := time.Now()

	for _, item := range response.Data.Diff {
		var market string
		switch item.F13 {
		case 0:
			market = "SZ"
		case 1:
			market = "SH"
		default:
			continue
		}

		tsCode := fmt.Sprintf("%s.%s", item.F12, market)

		// 检查是否在请求的股票列表中
		found := false
		for _, requestedCode := range tsCodes {
			if requestedCode == tsCode {
				found = true
				break
			}
		}

		if !found && len(tsCodes) > 0 {
			continue
		}

		// 转换当前日期为YYYYMMDD格式的int
		nowDateInt := now.Year()*10000 + int(now.Month())*100 + now.Day()

		data := model.DailyData{
			TsCode:    tsCode,
			TradeDate: nowDateInt,
			Close:     parseFloat(item.F2), // 最新价作为收盘价
			CreatedAt: now.Unix(),
			// 其他字段暂时无法从该API获取
		}
		realtimeData = append(realtimeData, data)
	}

	e.logger.Infof("Fetched realtime data for %d stocks", len(realtimeData))
	return realtimeData, nil
}

// parseInt64 解析64位整数
func parseInt64(s string) int64 {
	if s == "" || s == "-" {
		return 0
	}
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return i
}

// parseStockCode 解析股票代码，返回代码和市场
func (e *EastMoneyCollector) parseStockCode(tsCode string) (string, string, error) {
	parts := strings.Split(tsCode, ".")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid tsCode format: %s", tsCode)
	}
	return parts[0], parts[1], nil
}

// calculateNetProfitMargin 计算净利润率
func calculateNetProfitMargin(netProfit, revenue float64) float64 {
	if revenue == 0 {
		return 0
	}
	return (netProfit / revenue) * 100
}

// isValidTsCode 验证股票代码格式
func isValidTsCode(tsCode string) bool {
	// 检查格式是否为 XXXXXX.XX (6位数字.2位字母)
	matched, _ := regexp.MatchString(`^\d{6}\.(SH|SZ)$`, tsCode)
	return matched
}

// GetPerformanceReports 获取业绩报表数据
func (e *EastMoneyCollector) GetPerformanceReports(tsCode string) ([]model.PerformanceReport, error) {
	e.logger.Infof("Fetching performance reports for %s from EastMoney", tsCode)

	// 验证股票代码格式
	if !isValidTsCode(tsCode) {
		return nil, fmt.Errorf("invalid tsCode format: %s", tsCode)
	}

	// 提取股票代码（去掉交易所后缀）
	stockCode := strings.Split(tsCode, ".")[0]

	// 构建业绩报表API URL
	baseURL := "https://datacenter-web.eastmoney.com/api/data/v1/get"
	params := url.Values{}
	params.Set("callback", fmt.Sprintf("jQuery112305975330320237164_%d", time.Now().UnixMilli()))
	params.Set("sortColumns", "REPORTDATE")
	params.Set("sortTypes", "-1")
	params.Set("pageSize", "50")
	params.Set("pageNumber", "1")
	params.Set("columns", "ALL")
	params.Set("filter", fmt.Sprintf("(SECURITY_CODE=\"%s\")", stockCode))
	params.Set("reportName", "RPT_LICO_FN_CPD")

	requestURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// 发送请求
	resp, err := e.makePerformanceRequest(requestURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch performance reports: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 解析JSONP响应
	bodyStr := string(body)

	// 提取JSON部分（去掉JSONP包装）
	start := strings.Index(bodyStr, "(") + 1
	end := strings.LastIndex(bodyStr, ")")
	if start <= 0 || end <= start {
		return nil, fmt.Errorf("invalid JSONP response format")
	}

	jsonStr := bodyStr[start:end]

	// 解析JSON响应
	var response struct {
		Result struct {
			Data []map[string]interface{} `json:"data"`
		} `json:"result"`
		Success bool   `json:"success"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &response); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API returned error: %s", response.Message)
	}

	// 转换数据
	var reports []model.PerformanceReport
	for _, item := range response.Result.Data {
		report, err := e.convertToPerformanceReport(tsCode, item)
		if err != nil {
			e.logger.Warnf("Failed to convert performance report data: %v", err)
			continue
		}
		reports = append(reports, *report)
	}

	e.logger.Infof("Fetched %d performance reports for %s", len(reports), tsCode)
	return reports, nil
}

// GetLatestPerformanceReport 获取最新业绩报表数据
func (e *EastMoneyCollector) GetLatestPerformanceReport(tsCode string) (*model.PerformanceReport, error) {
	reports, err := e.GetPerformanceReports(tsCode)
	if err != nil {
		return nil, err
	}

	if len(reports) == 0 {
		return nil, fmt.Errorf("no performance reports found for %s", tsCode)
	}

	// 返回最新的业绩报表数据（假设数据已按日期排序）
	latest := reports[0]
	for _, report := range reports {
		if report.ReportDate > latest.ReportDate {
			latest = report
		}
	}

	return &latest, nil
}

// makePerformanceRequest 发送业绩报表请求
func (e *EastMoneyCollector) makePerformanceRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// 添加业绩报表专用请求头
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Referer", "https://data.eastmoney.com/bbsj/yjbb/")
	req.Header.Set("Sec-Fetch-Dest", "script")
	req.Header.Set("Sec-Fetch-Mode", "no-cors")
	req.Header.Set("Sec-Fetch-Site", "same-site")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36")
	req.Header.Set("sec-ch-ua", `"Not;A=Brand";v="99", "Google Chrome";v="139", "Chromium";v="139"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"macOS"`)

	e.logger.Debugf("Making performance request: %s", url)

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return resp, nil
}

// convertToPerformanceReport 转换业绩报表数据
func (e *EastMoneyCollector) convertToPerformanceReport(tsCode string, data map[string]interface{}) (*model.PerformanceReport, error) {
	report := &model.PerformanceReport{
		TsCode: tsCode,
	}

	// 解析报告期
	if reportDateStr, ok := data["REPORTDATE"].(string); ok {
		if reportDate, success := parseTimeToInt(reportDateStr); success {
			report.ReportDate = reportDate
		}
	}

	// 解析数值字段 - 根据调整后的PerformanceReport模型和实际API响应字段名称
	report.EPS = parseFloat(data["BASIC_EPS"])              // 每股收益
	report.WeightEPS = parseFloat(data["DEDUCT_BASIC_EPS"]) // 扣非每股收益

	report.Revenue = parseFloat(data["TOTAL_OPERATE_INCOME"])     // 营业总收入
	report.RevenueQoQ = limitGrowthRate(parseFloat(data["YSHZ"])) // 营业收入同比增长
	report.RevenueYoY = limitGrowthRate(parseFloat(data["YSTZ"])) // 营业收入环比增长

	report.NetProfit = parseFloat(data["PARENT_NETPROFIT"])          // 净利润
	report.NetProfitQoQ = limitGrowthRate(parseFloat(data["SJLHZ"])) // 净利润同比增长
	report.NetProfitYoY = limitGrowthRate(parseFloat(data["SJLTZ"])) // 净利润环比增长

	report.BVPS = parseFloat(data["BPS"])                             // 每股净资产
	report.GrossMargin = limitGrowthRate(parseFloat(data["XSMLL"]))   // 销售毛利率
	report.DividendYield = limitGrowthRate(parseFloat(data["ZXGXL"])) // 股息率

	// 解析公告日期 - 根据实际API响应字段名称
	if noticeDateStr, ok := data["NOTICE_DATE"].(string); ok {
		if noticeDate, success := parseTimeString(noticeDateStr); success {
			report.LatestAnnouncementDate = &noticeDate
		}
	}

	// 使用更新日期作为首次公告日期的替代
	if updateDateStr, ok := data["UPDATE_DATE"].(string); ok {
		if updateDate, success := parseTimeString(updateDateStr); success {
			report.FirstAnnouncementDate = &updateDate
		}
	}

	return report, nil
}

// limitGrowthRate 限制增长率在-9999到9999之间
func limitGrowthRate(value float64) float64 {
	const maxGrowthRate = 9999.0
	const minGrowthRate = -9999.0

	if value > maxGrowthRate {
		return maxGrowthRate
	}
	if value < minGrowthRate {
		return minGrowthRate
	}
	return value
}

// GetShareholderCounts 获取股东户数数据
func (e *EastMoneyCollector) GetShareholderCounts(tsCode string) ([]model.ShareholderCount, error) {
	e.logger.Infof("Fetching shareholder counts for %s from EastMoney", tsCode)

	// 验证股票代码格式
	if !isValidTsCode(tsCode) {
		return nil, fmt.Errorf("invalid tsCode format: %s", tsCode)
	}

	// 提取股票代码（去掉交易所后缀）
	stockCode := strings.Split(tsCode, ".")[0]

	// 构建股东户数API URL
	baseURL := "https://datacenter-web.eastmoney.com/api/data/v1/get"
	params := url.Values{}
	params.Set("callback", fmt.Sprintf("jQuery1123014159649525581786_%d", time.Now().UnixMilli()))
	params.Set("sortColumns", "END_DATE")
	params.Set("sortTypes", "-1")
	params.Set("pageSize", "50")
	params.Set("pageNumber", "1")
	params.Set("reportName", "RPT_HOLDERNUM_DET")
	params.Set("columns", "SECURITY_CODE,SECURITY_NAME_ABBR,CHANGE_SHARES,CHANGE_REASON,END_DATE,INTERVAL_CHRATE,AVG_MARKET_CAP,AVG_HOLD_NUM,TOTAL_MARKET_CAP,TOTAL_A_SHARES,HOLD_NOTICE_DATE,HOLDER_NUM,PRE_HOLDER_NUM,HOLDER_NUM_CHANGE,HOLDER_NUM_RATIO,END_DATE,PRE_END_DATE")
	params.Set("quoteColumns", "f2,f3")
	params.Set("quoteType", "0")
	params.Set("filter", fmt.Sprintf("(SECURITY_CODE=\"%s\")", stockCode))
	params.Set("source", "WEB")
	params.Set("client", "WEB")

	requestURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// 发送请求
	resp, err := e.makeRequest(requestURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch shareholder counts: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 解析JSONP响应
	bodyStr := string(body)

	// 提取JSON部分（去掉JSONP包装）
	start := strings.Index(bodyStr, "(") + 1
	end := strings.LastIndex(bodyStr, ")")
	if start <= 0 || end <= start {
		return nil, fmt.Errorf("invalid JSONP response format")
	}

	jsonStr := bodyStr[start:end]

	// 解析JSON响应
	var response struct {
		Result struct {
			Data []map[string]interface{} `json:"data"`
		} `json:"result"`
		Success bool   `json:"success"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &response); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API returned error: %s", response.Message)
	}

	// 转换数据
	var counts []model.ShareholderCount
	for _, item := range response.Result.Data {
		count, err := e.convertToShareholderCount(tsCode, item)
		if err != nil {
			e.logger.Warnf("Failed to convert shareholder count data: %v", err)
			continue
		}
		counts = append(counts, *count)
	}

	e.logger.Infof("Fetched %d shareholder count records for %s", len(counts), tsCode)
	return counts, nil
}

// GetLatestShareholderCount 获取最新股东户数数据
func (e *EastMoneyCollector) GetLatestShareholderCount(tsCode string) (*model.ShareholderCount, error) {
	counts, err := e.GetShareholderCounts(tsCode)
	if err != nil {
		return nil, err
	}

	if len(counts) == 0 {
		return nil, fmt.Errorf("no shareholder count data found for %s", tsCode)
	}

	// 返回最新的股东户数数据（假设数据已按日期排序）
	latest := counts[0]
	for _, count := range counts {
		if count.EndDate > latest.EndDate {
			latest = count
		}
	}

	return &latest, nil
}

// GetDailyKLine 获取日K线数据
func (e *EastMoneyCollector) GetDailyKLine(tsCode string, startDate, endDate time.Time) ([]model.DailyData, error) {
	klines, err := e.fetchKLineRawData(tsCode, KLineTypeDaily)
	if err != nil {
		return nil, err
	}

	result := make([]model.DailyData, 0, len(klines))
	for _, kline := range klines {
		if data, err := e.parser.ParseToDaily(tsCode, kline); err == nil {
			// 根据时间范围过滤数据
			if e.isInDateRange(data.TradeDate, startDate, endDate) {
				result = append(result, *data)
			}
		} else {
			e.logger.Warnf("Failed to parse daily K-line data: %v", err)
		}
	}

	e.logger.Infof("Fetched %d daily K-line records for %s (filtered from %d total)", len(result), tsCode, len(klines))
	return result, nil
}

// GetWeeklyKLine 获取周K线数据
func (e *EastMoneyCollector) GetWeeklyKLine(tsCode string, startDate, endDate time.Time) ([]model.WeeklyData, error) {
	klines, err := e.fetchKLineRawData(tsCode, KLineTypeWeekly)
	if err != nil {
		return nil, err
	}

	result := make([]model.WeeklyData, 0, len(klines))
	for _, kline := range klines {
		if data, err := e.parser.ParseToWeekly(tsCode, kline); err == nil {
			// 根据时间范围过滤数据
			if e.isInDateRange(data.TradeDate, startDate, endDate) {
				result = append(result, *data)
			}
		} else {
			e.logger.Warnf("Failed to parse weekly K-line data: %v", err)
		}
	}

	e.logger.Infof("Fetched %d weekly K-line records for %s (filtered from %d total)", len(result), tsCode, len(klines))
	return result, nil
}

// GetMonthlyKLine 获取月K线数据
func (e *EastMoneyCollector) GetMonthlyKLine(tsCode string, startDate, endDate time.Time) ([]model.MonthlyData, error) {
	klines, err := e.fetchKLineRawData(tsCode, KLineTypeMonthly)
	if err != nil {
		return nil, err
	}

	result := make([]model.MonthlyData, 0, len(klines))
	for _, kline := range klines {
		if data, err := e.parser.ParseToMonthly(tsCode, kline); err == nil {
			// 根据时间范围过滤数据
			if e.isInDateRange(data.TradeDate, startDate, endDate) {
				result = append(result, *data)
			}
		} else {
			e.logger.Warnf("Failed to parse monthly K-line data: %v", err)
		}
	}

	e.logger.Infof("Fetched %d monthly K-line records for %s (filtered from %d total)", len(result), tsCode, len(klines))
	return result, nil
}

// GetYearlyKLine 获取年K线数据
func (e *EastMoneyCollector) GetYearlyKLine(tsCode string, startDate, endDate time.Time) ([]model.YearlyData, error) {
	klines, err := e.fetchKLineRawData(tsCode, KLineTypeYearly)
	if err != nil {
		return nil, err
	}

	result := make([]model.YearlyData, 0, len(klines))
	for _, kline := range klines {
		if data, err := e.parser.ParseToYearly(tsCode, kline); err == nil {
			// 根据时间范围过滤数据
			if e.isInDateRange(data.TradeDate, startDate, endDate) {
				result = append(result, *data)
			}
		} else {
			e.logger.Warnf("Failed to parse yearly K-line data: %v", err)
		}
	}

	e.logger.Infof("Fetched %d yearly K-line records for %s (filtered from %d total)", len(result), tsCode, len(klines))
	return result, nil
}

// fetchKLineRawData 获取原始K线数据
func (e *EastMoneyCollector) fetchKLineRawData(tsCode string, klineType KLineType) (
	[]string, error) {
	e.logger.Debugf("Fetching K-line data for %s, type: %s", tsCode, klineType)

	// 构建请求URL
	requestURL, err := e.buildKLineURL(tsCode, klineType)
	if err != nil {
		return nil, fmt.Errorf("build URL failed: %w", err)
	}

	// 发送请求并解析响应
	response, err := e.sendKLineRequest(requestURL)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if response.RC != 0 {
		return nil, fmt.Errorf("API error: rc=%d", response.RC)
	}

	return response.Data.Klines, nil
}

// buildKLineURL 构建K线请求URL
func (e *EastMoneyCollector) buildKLineURL(tsCode, klineType string) (string, error) {
	symbol, market, err := e.parseStockCode(tsCode)
	if err != nil {
		return "", err
	}

	// 构建secid
	secid := e.buildSecID(symbol, market)
	if secid == "" {
		return "", fmt.Errorf("unsupported market: %s", market)
	}

	// 构建请求参数
	params := url.Values{
		"fields1": {"f1,f2,f3,f4,f5,f6,f7,f8,f9,f10,f11,f12,f13"},
		"fields2": {"f51,f52,f53,f54,f55,f56,f57,f58,f59,f60,f61"},
		"beg":     {"0"},
		"end":     {"20500101"},
		"ut":      {"fa5fd1943c7b386f172d6893dbfba10b"},
		"rtntype": {"6"},
		"secid":   {secid},
		"klt":     {klineType},
		"fqt":     {"1"},
		"cb":      {fmt.Sprintf("jsonp%d", time.Now().UnixMilli())},
	}

	return "https://push2his.eastmoney.com/api/qt/stock/kline/get?" + params.Encode(), nil
}

// buildSecID 构建证券ID
func (e *EastMoneyCollector) buildSecID(symbol, market string) string {
	switch market {
	case "SH":
		return "1." + symbol
	case "SZ":
		return "0." + symbol
	default:
		return ""
	}
}

// sendKLineRequest 发送K线请求并解析响应
func (e *EastMoneyCollector) sendKLineRequest(requestURL string) (*KLineResponse, error) {
	resp, err := e.makeRequest(requestURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 解析JSONP响应
	jsonData, err := e.extractJSONFromJSONP(string(body))
	if err != nil {
		return nil, err
	}

	var response KLineResponse
	if err := json.Unmarshal([]byte(jsonData), &response); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &response, nil
}

// extractJSONFromJSONP 从JSONP响应中提取JSON数据
func (e *EastMoneyCollector) extractJSONFromJSONP(body string) (string, error) {
	re := regexp.MustCompile(`jsonp\d+\((.*)\)`)
	matches := re.FindStringSubmatch(body)
	if len(matches) < 2 {
		return "", fmt.Errorf("invalid JSONP response format")
	}
	return matches[1], nil
}

// isInDateRange 检查交易日期是否在指定的时间范围内
func (e *EastMoneyCollector) isInDateRange(tradeDate int, startDate, endDate time.Time) bool {
	// 将int类型的交易日期转换为time.Time
	tradeDateStr := fmt.Sprintf("%d", tradeDate)
	date, err := time.Parse("20060102", tradeDateStr)
	if err != nil {
		e.logger.Warnf("Failed to parse trade date %d: %v", tradeDate, err)
		return false
	}

	// 检查是否在时间范围内
	if !startDate.IsZero() && date.Before(startDate) {
		return false
	}
	if !endDate.IsZero() && date.After(endDate) {
		return false
	}

	return true
}

// convertToShareholderCount 转换股东户数数据
func (e *EastMoneyCollector) convertToShareholderCount(tsCode string, data map[string]interface{}) (*model.ShareholderCount, error) {
	count := &model.ShareholderCount{
		TsCode: tsCode,
	}

	// 解析基本信息
	if securityCode, ok := data["SECURITY_CODE"].(string); ok {
		count.SecurityCode = securityCode
	}

	if securityName, ok := data["SECURITY_NAME_ABBR"].(string); ok {
		count.SecurityName = securityName
	}

	// 解析统计截止日期
	if endDateStr, ok := data["END_DATE"].(string); ok {
		if endDate, success := parseTimeToInt(endDateStr); success {
			count.EndDate = endDate
		}
	}

	// 解析股东户数相关数据
	count.HolderNum = int64(parseFloat(data["HOLDER_NUM"]))
	count.PreHolderNum = int64(parseFloat(data["PRE_HOLDER_NUM"]))
	count.HolderNumChange = int64(parseFloat(data["HOLDER_NUM_CHANGE"]))

	// 解析股东户数变化比例，限制在合理范围内
	holderNumRatio := parseFloat(data["HOLDER_NUM_RATIO"])
	if holderNumRatio > 999999 || holderNumRatio < -999999 {
		e.logger.Warnf("Holder num ratio out of range: %f, setting to 0", holderNumRatio)
		holderNumRatio = 0
	}
	count.HolderNumRatio = holderNumRatio

	// 解析市值相关数据
	count.AvgMarketCap = parseFloat(data["AVG_MARKET_CAP"])
	count.AvgHoldNum = parseFloat(data["AVG_HOLD_NUM"])
	count.TotalMarketCap = parseFloat(data["TOTAL_MARKET_CAP"])
	count.TotalAShares = int64(parseFloat(data["TOTAL_A_SHARES"]))

	// 解析其他数据
	count.IntervalChrate = parseFloat(data["INTERVAL_CHRATE"])
	count.ChangeShares = int64(parseFloat(data["CHANGE_SHARES"]))

	if changeReason, ok := data["CHANGE_REASON"].(string); ok {
		count.ChangeReason = changeReason
	}

	// 解析公告日期
	if holdNoticeDateStr, ok := data["HOLD_NOTICE_DATE"].(string); ok {
		if holdNoticeDate, success := parseTimeString(holdNoticeDateStr); success {
			count.HoldNoticeDate = &holdNoticeDate
		}
		// 如果解析失败，保持为 nil（GORM会将其存储为NULL）
	}

	// 解析上期截止日期
	if preEndDateStr, ok := data["PRE_END_DATE"].(string); ok {
		if preEndDate, success := parseTimeToInt(preEndDateStr); success {
			count.PreEndDate = &preEndDate
		}
		// 如果解析失败，保持为 nil（GORM会将其存储为NULL）
	}

	// 设置创建和更新时间
	now := time.Now()
	count.CreatedAt = now
	count.UpdatedAt = now

	return count, nil
}
