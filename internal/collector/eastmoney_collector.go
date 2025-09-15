package collector

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"stock/internal/model"
	"stock/internal/utils"
)

// EastMoneyCollector 东方财富数据采集器
type EastMoneyCollector struct {
	BaseCollector
	client *http.Client
	logger *utils.Logger
}

// NewEastMoneyCollector 创建东方财富采集器
func NewEastMoneyCollector(logger *utils.Logger) *EastMoneyCollector {
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
			"sec-ch-ua-platform": `"macOS"`,
		},
	}

	return &EastMoneyCollector{
		BaseCollector: BaseCollector{
			Config:    config,
			Connected: false,
		},
		client: &http.Client{
			Timeout: config.Timeout,
		},
		logger: logger,
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
	_, err := e.fetchStockListPage(1, 10)
	return err
}

// makeRequest 发送HTTP请求
func (e *EastMoneyCollector) makeRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// 添加请求头
	for key, value := range e.Config.Headers {
		req.Header.Set(key, value)
	}

	e.logger.Debugf("Making request: %s", url)

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
	params.Set("cb", fmt.Sprintf("jQuery112307402025327315181_%d", time.Now().UnixMilli()))
	params.Set("fid", "f62")
	params.Set("po", "1")
	params.Set("pz", strconv.Itoa(pageSize))
	params.Set("pn", strconv.Itoa(page))
	params.Set("np", "1")
	params.Set("fltt", "2")
	params.Set("invt", "2")
	params.Set("ut", "8dec03ba335b81bf4ebdf7b29ec27d15")

	// 市场筛选参数 - 所有A股
	params.Set("fs", "m:0+t:6,m:0+t:80,m:1+t:2,m:0+t:7,m:1+t:3")

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
		time.DateTime, // "2006-01-02 15:04:05"
		time.Layout,   // "01/02 03:04:05PM '06 -0700"
		time.DateOnly, // "2006-01-02"
	}

	for _, format := range formats {
		if parsedTime, err := time.ParseInLocation(format, timeStr, time.Local); err == nil {
			return parsedTime, true
		}
	}

	return time.Time{}, false
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
				TsCode:   tsCode,
				Symbol:   item.F12,
				Name:     item.F14,
				Market:   market,
				IsActive: true,
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

// GetDailyKLine 获取日K线数据
func (e *EastMoneyCollector) GetDailyKLine(tsCode string, startDate, endDate time.Time) ([]model.DailyData, error) {
	e.logger.Infof("Fetching daily K-line data for %s from %s to %s", tsCode, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	// 从tsCode解析股票代码和市场
	parts := strings.Split(tsCode, ".")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid tsCode format: %s", tsCode)
	}

	symbol := parts[0]
	market := parts[1]

	// 构建请求URL - 使用你提供的新API
	baseURL := "https://push2his.eastmoney.com/api/qt/stock/kline/get"
	params := url.Values{}
	params.Set("ut", "fa5fd1943c7b386f172d6893dbfba10b")
	params.Set("rtntype", "6")
	params.Set("cb", fmt.Sprintf("jsonp%d", time.Now().UnixMilli()))

	// 设置股票代码
	var secid string
	if market == "SH" {
		secid = fmt.Sprintf("1.%s", symbol)
	} else if market == "SZ" {
		secid = fmt.Sprintf("0.%s", symbol)
	} else {
		return nil, fmt.Errorf("unsupported market: %s", market)
	}
	params.Set("secid", secid)

	// 设置字段和K线类型
	params.Set("fields1", "f1,f2,f3,f4,f5,f6,f7,f8,f9,f10,f11,f12,f13")
	params.Set("fields2", "f51,f52,f53,f54,f55,f56,f57,f58,f59,f60,f61")
	params.Set("klt", "101") // 日K线
	params.Set("fqt", "1")   // 前复权
	// 设置时间范围 - 使用你提供的格式
	params.Set("beg", "0")        // 从最早开始
	params.Set("end", "20500101") // 到未来日期

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
			Code   string   `json:"code"`
			Market int      `json:"market"`
			Name   string   `json:"name"`
			Klines []string `json:"klines"`
		} `json:"data"`
	}

	if err := json.Unmarshal([]byte(jsonData), &response); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	if response.RC != 0 {
		return nil, fmt.Errorf("API error: rc=%d", response.RC)
	}

	var dailyData []model.DailyData

	for _, kline := range response.Data.Klines {
		// K线数据格式：日期,开盘,收盘,最高,最低,成交量,成交额,振幅,涨跌幅,涨跌额,换手率
		fields := strings.Split(kline, ",")
		if len(fields) < 11 {
			continue
		}

		// 解析日期并转换为int格式 (YYYYMMDD)
		tradeDate, err := time.Parse("2006-01-02", fields[0])
		if err != nil {
			e.logger.Warnf("Failed to parse date %s: %v", fields[0], err)
			continue
		}

		// 转换为YYYYMMDD格式的int
		tradeDateInt := tradeDate.Year()*10000 + int(tradeDate.Month())*100 + tradeDate.Day()

		data := model.DailyData{
			TsCode:    tsCode,
			TradeDate: tradeDateInt,
			Open:      parseFloat(fields[1]),
			Close:     parseFloat(fields[2]),
			High:      parseFloat(fields[3]),
			Low:       parseFloat(fields[4]),
			Volume:    int64(parseFloat(fields[5])),
			Amount:    parseFloat(fields[6]),
			CreatedAt: time.Now().Unix(),
		}

		// 注意：DailyData模型中暂未包含涨跌幅和涨跌额字段
		// 如需要可以在模型中添加这些字段

		dailyData = append(dailyData, data)
	}

	e.logger.Infof("Fetched %d daily K-line records for %s", len(dailyData), tsCode)
	return dailyData, nil
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
		if report.ReportDate.After(latest.ReportDate) {
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
		if reportDate, success := parseTimeString(reportDateStr); success {
			report.ReportDate = reportDate
		}
	}

	// 解析数值字段 - 根据调整后的PerformanceReport模型和实际API响应字段名称
	report.EPS = parseFloat(data["BASIC_EPS"])              // 每股收益
	report.WeightEPS = parseFloat(data["DEDUCT_BASIC_EPS"]) // 扣非每股收益

	report.Revenue = parseFloat(data["TOTAL_OPERATE_INCOME"]) // 营业总收入
	report.RevenueQoQ = parseFloat(data["YSHZ"])              // 营业收入同比增长
	report.RevenueYoY = parseFloat(data["YSTZ"])              // 营业收入环比增长

	report.NetProfit = parseFloat(data["PARENT_NETPROFIT"]) // 净利润
	report.NetProfitQoQ = parseFloat(data["SJLHZ"])         // 净利润同比增长
	report.NetProfitYoY = parseFloat(data["SJLTZ"])         // 净利润环比增长

	report.BVPS = parseFloat(data["BPS"])            // 每股净资产
	report.GrossMargin = parseFloat(data["XSMLL"])   // 销售毛利率
	report.DividendYield = parseFloat(data["ZXGXL"]) // 股息率

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
