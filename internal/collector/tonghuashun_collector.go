package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"stock/internal/logger"
	"stock/internal/model"

	"golang.org/x/time/rate"
)

// K线类型常量 - 同花顺采集器使用
const (
	THSKLineTypeDaily     = "01" // 日K线
	THSKLineTypeWeekly    = "11" // 周K线
	THSKLineTypeMonthly   = "21" // 月K线
	THSKLineTypeQuarterly = "91" // 季K线
	THSKLineTypeYearly    = "81" // 年K线
)

// TongHuaShunCollector 同花顺数据采集器
type TongHuaShunCollector struct {
	BaseCollector
	client         *http.Client
	logger         *logger.Logger
	limiter        *rate.Limiter // 限流器
	userAgentGen   *UserAgentGenerator
	cookieGen      *CookieGenerator
	currentUA      string
	currentCookie  string
	lastUpdateTime time.Time
}

// newTongHuaShunCollector 创建同花顺采集器
func newTongHuaShunCollector(logger *logger.Logger) *TongHuaShunCollector {
	config := CollectorConfig{
		Name:      "tonghuashun",
		BaseURL:   "https://d.10jqka.com.cn",
		Timeout:   10 * time.Second,
		RateLimit: 100, // 每秒1个请求
		Headers: map[string]string{
			"Accept":           "*/*",
			"Accept-Language":  "zh-CN,zh;q=0.9,en;q=0.8",
			"Connection":       "keep-alive",
			"sec-ch-ua-mobile": "?0",
			"sec-fetch-dest":   "script",
			"sec-fetch-mode":   "no-cors",
			"sec-fetch-site":   "same-site",
			"cache-control":    "no-cache",
			"pragma":           "no-cache",
		},
	}

	// 创建限流器，每秒允许 RateLimit 个请求，突发容量为 RateLimit*2
	limiter := rate.NewLimiter(rate.Limit(config.RateLimit), config.RateLimit*2)

	// 创建随机生成器
	userAgentGen := NewUserAgentGenerator()
	cookieGen := NewCookieGenerator()

	collector := &TongHuaShunCollector{
		BaseCollector: BaseCollector{
			Config:    config,
			Connected: false,
		},
		client: &http.Client{
			Timeout: config.Timeout,
		},
		logger:       logger,
		limiter:      limiter,
		userAgentGen: userAgentGen,
		cookieGen:    cookieGen,
	}

	// 初始化随机User-Agent和Cookie
	collector.updateUserAgentAndCookie()

	return collector
}

// Connect 连接数据源
func (t *TongHuaShunCollector) Connect() error {
	t.Connected = true
	t.logger.Infof("Successfully connected to %s", t.Config.Name)
	return nil
}

// Disconnect 断开连接
func (t *TongHuaShunCollector) Disconnect() error {
	t.Connected = false
	t.logger.Infof("Disconnected from %s", t.Config.Name)
	return nil
}

// makeRequest 发送HTTP请求（带限流）
func (t *TongHuaShunCollector) makeRequest(url, refer string) (*http.Response, error) {
	return t.makeRequestWithContext(context.Background(), url, refer)
}

// makeRequestWithContext 发送HTTP请求（带限流和上下文）
func (t *TongHuaShunCollector) makeRequestWithContext(ctx context.Context, url, refer string) (*http.Response, error) {
	// 应用限流
	if err := t.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait failed: %v", err)
	}

	// 检查是否需要更新User-Agent和Cookie（每1分钟更新一次）
	if time.Since(t.lastUpdateTime) > 1*time.Minute {
		t.updateUserAgentAndCookie()
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// 添加基础请求头
	for key, value := range t.Config.Headers {
		// 跳过User-Agent和Cookie，使用随机生成的
		if key == "User-Agent" || key == "cookie" {
			continue
		}
		req.Header.Set(key, value)
	}

	// 设置随机生成的User-Agent和Cookie
	req.Header.Set("User-Agent", t.currentUA)
	req.Header.Set("Cookie", t.currentCookie)
	req.Header.Set("sec-ch-ua", t.userAgentGen.GenerateSecChUa(t.currentUA))
	req.Header.Set("sec-ch-ua-platform", t.getPlatformFromUA(t.currentUA))
	req.Header.Set("Referer", refer)

	t.logger.Debugf("Making rate-limited request to TongHuaShun: %s", url)

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return resp, nil
}

// GetStockList 获取股票列表
func (t *TongHuaShunCollector) GetStockList() ([]model.Stock, error) {
	t.logger.Info("TongHuaShun GetStockList - 开始获取股票列表")

	var allStocks []model.Stock
	maxPages := 103 // 限制最大页数，避免无限循环

	for page := 1; page <= maxPages; page++ {
		stocks, hasMore, err := t.getStockListPage(page)
		if err != nil {
			t.logger.Errorf("获取第%d页股票列表失败: %v", page, err)
			// 如果是第一页就失败，返回错误；否则继续处理已获取的数据
			if page == 1 {
				return nil, fmt.Errorf("获取股票列表失败: %v", err)
			}
			break
		}

		allStocks = append(allStocks, stocks...)
		t.logger.Infof("已获取第%d页，本页%d只股票，累计%d只股票", page, len(stocks), len(allStocks))

		if !hasMore || len(stocks) == 0 {
			break
		}

		// 添加延迟，避免请求过快
		time.Sleep(1 * time.Second)
	}

	t.logger.Infof("TongHuaShun GetStockList 完成，共获取%d只股票", len(allStocks))
	return allStocks, nil
}

// getStockListPage 获取指定页的股票列表
func (t *TongHuaShunCollector) getStockListPage(page int) ([]model.Stock, bool, error) {
	// 使用提供的同花顺API端点
	url := fmt.Sprintf("https://data.10jqka.com.cn/funds/ggzjl/field/zdf/order/desc/page/%d/ajax/1/free/1/", page)

	// 创建请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, false, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头 - 完全按照curl请求设置
	req.Header.Set("Accept", "text/html, */*; q=0.01")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Priority", "u=1, i")
	req.Header.Set("Referer", "https://data.10jqka.com.cn/funds/ggzjl/")
	req.Header.Set("Sec-Ch-Ua", `"Not;A=Brand";v="99", "Google Chrome";v="139", "Chromium";v="139"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"macOS"`)
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	// 动态生成 hexin-v token
	hexinV := GenerateWencaiToken()

	// 设置Cookie - 基于curl请求中的格式
	timestamp := time.Now().Unix()
	cookieValue := fmt.Sprintf("Hm_lvt_722143063e4892925903024537075d0d=%d; HMACCOUNT=17C55F0F7B5ABE69; Hm_lvt_929f8b362150b1f77b477230541dbbc2=%d; Hm_lvt_78c58f01938e4d85eaf619eae71b4ed1=%d; Hm_lvt_69929b9dce4c22a060bd22d703b2a280=%d; spversion=20130314; Hm_lvt_60bad21af9c824a4a0530d5dbf4357ca=%d; Hm_lvt_f79b64788a4e377c608617fba4c736e2=%d; historystock=600930%%7C*%%7C001208%%7C*%%7C001201%%7C*%%7C300111; log=; Hm_lpvt_f79b64788a4e377c608617fba4c736e2=%d; Hm_lpvt_60bad21af9c824a4a0530d5dbf4357ca=%d; Hm_lpvt_722143063e4892925903024537075d0d=%d; Hm_lpvt_78c58f01938e4d85eaf619eae71b4ed1=%d; Hm_lpvt_929f8b362150b1f77b477230541dbbc2=%d; Hm_lpvt_69929b9dce4c22a060bd22d703b2a280=%d; v=%s",
		timestamp, timestamp, timestamp, timestamp, timestamp, timestamp,
		timestamp, timestamp, timestamp, timestamp, timestamp, timestamp, hexinV)

	req.Header.Set("Cookie", cookieValue)
	req.Header.Set("Hexin-V", hexinV)

	// 发送请求
	resp, err := t.client.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, false, fmt.Errorf("HTTP状态码: %d", resp.StatusCode)
	}

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, false, fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析HTML响应
	stocks, hasMore, err := t.parseStockListHTML(string(body))
	if err != nil {
		return nil, false, fmt.Errorf("解析HTML失败: %v", err)
	}

	return stocks, hasMore, nil
}

// parseStockListHTML 解析股票列表HTML
func (t *TongHuaShunCollector) parseStockListHTML(html string) ([]model.Stock, bool, error) {
	// 查找所有股票行
	lines := strings.Split(html, "\n")
	var stocks = make([]model.Stock, 0, len(lines))
	for i, line := range lines {
		line = strings.TrimSpace(line)

		// 查找包含股票代码的行
		if strings.Contains(line, "stockCode") && strings.Contains(line, "linkToGghq") {
			stock, err := t.parseStockRow(line, lines, i)
			if err != nil {
				t.logger.Warnf("解析股票行失败: %v, 行内容: %s", err, line)
				continue
			}
			if stock != nil {
				stocks = append(stocks, *stock)
			}
		}
	}

	// 简单判断是否还有更多页面（如果当前页有数据，假设还有更多）
	hasMore := len(stocks) > 0

	return stocks, hasMore, nil
}

// parseStockRow 解析单个股票行
func (t *TongHuaShunCollector) parseStockRow(line string, allLines []string, lineIndex int) (*model.Stock, error) {
	// 提取股票代码
	codeStart := strings.Index(line, "stockCode\">")
	if codeStart == -1 {
		return nil, fmt.Errorf("未找到股票代码")
	}
	codeStart += len("stockCode\">")
	codeEnd := strings.Index(line[codeStart:], "</a>")
	if codeEnd == -1 {
		return nil, fmt.Errorf("未找到股票代码结束标记")
	}
	stockCode := line[codeStart : codeStart+codeEnd]

	// 尝试多种方法提取股票名称
	stockName := t.extractStockName(line, allLines, lineIndex, stockCode)

	// 确定市场和构建ts_code
	var market string
	var tsCode string

	if len(stockCode) == 6 {
		// 根据股票代码前缀判断市场
		switch {
		case strings.HasPrefix(stockCode, "60") || strings.HasPrefix(stockCode, "68"):
			market = "SH" // 上海证券交易所
			tsCode = stockCode + ".SH"
		case strings.HasPrefix(stockCode, "00") || strings.HasPrefix(stockCode, "30"):
			market = "SZ" // 深圳证券交易所
			tsCode = stockCode + ".SZ"
		default:
			market = "SZ" // 默认深圳
			tsCode = stockCode + ".SZ"
		}
	} else {
		return nil, fmt.Errorf("无效的股票代码格式: %s", stockCode)
	}

	// 创建股票对象
	stock := &model.Stock{
		TsCode:    tsCode,
		Symbol:    stockCode,
		Name:      stockName,
		Market:    market,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return stock, nil
}

// extractStockName 提取股票名称的多种方法
func (t *TongHuaShunCollector) extractStockName(line string, allLines []string, lineIndex int, stockCode string) string {
	// 方法1：查找 title 属性
	if name := t.extractFromTitle(line); name != "" {
		return name
	}

	// 方法2：查找后续行的 title 属性
	for j := lineIndex + 1; j < len(allLines) && j < lineIndex+5; j++ {
		if name := t.extractFromTitle(allLines[j]); name != "" {
			return name
		}
	}

	// 方法3：查找 alt 属性
	if name := t.extractFromAlt(line); name != "" {
		return name
	}

	// 方法4：查找后续行的 alt 属性
	for j := lineIndex + 1; j < len(allLines) && j < lineIndex+5; j++ {
		if name := t.extractFromAlt(allLines[j]); name != "" {
			return name
		}
	}

	// 方法5：查找包含中文的纯文本行
	for j := lineIndex + 1; j < len(allLines) && j < lineIndex+10; j++ {
		nextLine := strings.TrimSpace(allLines[j])
		if t.isValidStockName(nextLine) {
			return nextLine
		}
	}

	// 方法6：查找 data-name 或其他数据属性
	if name := t.extractFromDataAttributes(line); name != "" {
		return name
	}

	// 方法7：查找后续行的数据属性
	for j := lineIndex + 1; j < len(allLines) && j < lineIndex+5; j++ {
		if name := t.extractFromDataAttributes(allLines[j]); name != "" {
			return name
		}
	}

	// 最后使用股票代码作为名称
	return stockCode
}

// extractFromTitle 从 title 属性提取名称
func (t *TongHuaShunCollector) extractFromTitle(line string) string {
	patterns := []string{"title=\"", "title='"}
	for _, pattern := range patterns {
		if nameStart := strings.Index(line, pattern); nameStart != -1 {
			nameStart += len(pattern)
			endChar := pattern[len(pattern)-1:]
			if nameEnd := strings.Index(line[nameStart:], endChar); nameEnd != -1 {
				name := strings.TrimSpace(line[nameStart : nameStart+nameEnd])
				if t.isValidStockName(name) {
					return name
				}
			}
		}
	}
	return ""
}

// extractFromAlt 从 alt 属性提取名称
func (t *TongHuaShunCollector) extractFromAlt(line string) string {
	patterns := []string{"alt=\"", "alt='"}
	for _, pattern := range patterns {
		if nameStart := strings.Index(line, pattern); nameStart != -1 {
			nameStart += len(pattern)
			endChar := pattern[len(pattern)-1:]
			if nameEnd := strings.Index(line[nameStart:], endChar); nameEnd != -1 {
				name := strings.TrimSpace(line[nameStart : nameStart+nameEnd])
				if t.isValidStockName(name) {
					return name
				}
			}
		}
	}
	return ""
}

// extractFromDataAttributes 从数据属性提取名称
func (t *TongHuaShunCollector) extractFromDataAttributes(line string) string {
	patterns := []string{"data-name=\"", "data-title=\"", "data-stock-name=\""}
	for _, pattern := range patterns {
		if nameStart := strings.Index(line, pattern); nameStart != -1 {
			nameStart += len(pattern)
			if nameEnd := strings.Index(line[nameStart:], "\""); nameEnd != -1 {
				name := strings.TrimSpace(line[nameStart : nameStart+nameEnd])
				if t.isValidStockName(name) {
					return name
				}
			}
		}
	}
	return ""
}

// isValidStockName 判断是否为有效的股票名称
func (t *TongHuaShunCollector) isValidStockName(name string) bool {
	if len(name) == 0 || len(name) > 20 {
		return false
	}

	// 排除HTML标签
	if strings.Contains(name, "<") || strings.Contains(name, ">") {
		return false
	}

	// 排除明显不是股票名称的内容
	excludePatterns := []string{"http", "www", "javascript", "function", "var ", "return", "null", "undefined"}
	for _, pattern := range excludePatterns {
		if strings.Contains(strings.ToLower(name), pattern) {
			return false
		}
	}

	// 必须包含中文字符或者是常见的英文股票名称格式
	return containsChinese(name) || t.isEnglishStockName(name)
}

// isEnglishStockName 判断是否为英文股票名称
func (t *TongHuaShunCollector) isEnglishStockName(name string) bool {
	// 简单判断：包含字母且长度合理
	hasLetter := false
	for _, r := range name {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
			hasLetter = true
			break
		}
	}
	return hasLetter && len(name) >= 2 && len(name) <= 15
}

// containsChinese 检查字符串是否包含中文字符
func containsChinese(s string) bool {
	for _, r := range s {
		if r >= 0x4e00 && r <= 0x9fff {
			return true
		}
	}
	return false
}

// GetStockDetail 获取股票详情 - 空实现
func (t *TongHuaShunCollector) GetStockDetail(tsCode string) (*model.Stock, error) {
	t.logger.Infof("TongHuaShun GetStockDetail for %s - 功能暂未实现", tsCode)
	return nil, fmt.Errorf("TongHuaShun GetStockDetail not implemented yet")
}

// GetStockData 获取股票历史数据 - 空实现
func (t *TongHuaShunCollector) GetStockData(tsCode string, startDate, endDate time.Time) ([]model.DailyData, error) {
	t.logger.Infof("TongHuaShun GetStockData for %s - 功能暂未实现", tsCode)
	return []model.DailyData{}, fmt.Errorf("TongHuaShun GetStockData not implemented yet")
}

// GetDailyKLine 获取日K线数据
func (t *TongHuaShunCollector) GetDailyKLine(tsCode string, startDate, endDate time.Time) ([]model.DailyData, error) {
	t.logger.Infof("TongHuaShun GetDailyKLine for %s", tsCode)

	// 使用通用方法获取K线数据
	rawData, err := t.getKLineData(tsCode, THSKLineTypeDaily, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch daily K-line data: %w", err)
	}

	// 转换为日K线数据格式
	var dailyData []model.DailyData
	for _, item := range rawData {
		daily := model.DailyData{
			TsCode:    item.TsCode,
			TradeDate: item.TradeDate,
			Open:      item.Open,
			High:      item.High,
			Low:       item.Low,
			Close:     item.Close,
			Volume:    item.Volume,
			Amount:    item.Amount,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		}
		dailyData = append(dailyData, daily)
	}

	return dailyData, nil
}

// GetWeeklyKLine 获取周K线数据
func (t *TongHuaShunCollector) GetWeeklyKLine(tsCode string, startDate, endDate time.Time) ([]model.WeeklyData, error) {
	t.logger.Infof("TongHuaShun GetWeeklyKLine for %s", tsCode)

	// 使用通用方法获取K线数据
	rawData, err := t.getKLineData(tsCode, THSKLineTypeWeekly, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weekly K-line data: %w", err)
	}

	// 转换为周K线数据格式
	var weeklyData []model.WeeklyData
	for _, item := range rawData {
		weekly := model.WeeklyData{
			TsCode:    item.TsCode,
			TradeDate: item.TradeDate,
			Open:      item.Open,
			High:      item.High,
			Low:       item.Low,
			Close:     item.Close,
			Volume:    item.Volume,
			Amount:    item.Amount,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		}
		weeklyData = append(weeklyData, weekly)
	}

	return weeklyData, nil
}

// GetMonthlyKLine 获取月K线数据
func (t *TongHuaShunCollector) GetMonthlyKLine(tsCode string, startDate, endDate time.Time) ([]model.MonthlyData, error) {
	t.logger.Infof("TongHuaShun GetMonthlyKLine for %s", tsCode)

	// 使用通用方法获取K线数据
	rawData, err := t.getKLineData(tsCode, THSKLineTypeMonthly, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch monthly K-line data: %w", err)
	}

	// 转换为月K线数据格式
	var monthlyData []model.MonthlyData
	for _, item := range rawData {
		monthly := model.MonthlyData{
			TsCode:    item.TsCode,
			TradeDate: item.TradeDate,
			Open:      item.Open,
			High:      item.High,
			Low:       item.Low,
			Close:     item.Close,
			Volume:    item.Volume,
			Amount:    item.Amount,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		}
		monthlyData = append(monthlyData, monthly)
	}

	return monthlyData, nil
}

// GetQuarterlyKLine 获取季K线数据
func (t *TongHuaShunCollector) GetQuarterlyKLine(tsCode string, startDate, endDate time.Time) ([]model.QuarterlyData, error) {
	t.logger.Infof("TongHuaShun GetQuarterlyKLine for %s", tsCode)

	// 使用通用方法获取K线数据
	rawData, err := t.getKLineData(tsCode, THSKLineTypeQuarterly, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch quarterly K-line data: %w", err)
	}

	// 转换为季K线数据格式
	var quarterlyData []model.QuarterlyData
	for _, item := range rawData {
		quarterly := model.QuarterlyData{
			TsCode:    item.TsCode,
			TradeDate: item.TradeDate,
			Open:      item.Open,
			High:      item.High,
			Low:       item.Low,
			Close:     item.Close,
			Volume:    item.Volume,
			Amount:    item.Amount,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		}
		quarterlyData = append(quarterlyData, quarterly)
	}

	return quarterlyData, nil
}

// GetYearlyKLine 获取年K线数据
func (t *TongHuaShunCollector) GetYearlyKLine(tsCode string, startDate, endDate time.Time) ([]model.YearlyData, error) {
	t.logger.Infof("TongHuaShun GetYearlyKLine for %s", tsCode)

	// 使用通用方法获取K线数据
	rawData, err := t.getKLineData(tsCode, THSKLineTypeYearly, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch yearly K-line data: %w", err)
	}

	// 转换为年K线数据格式
	var yearlyData []model.YearlyData
	for _, item := range rawData {
		yearly := model.YearlyData{
			TsCode:    item.TsCode,
			TradeDate: item.TradeDate,
			Open:      item.Open,
			High:      item.High,
			Low:       item.Low,
			Close:     item.Close,
			Volume:    item.Volume,
			Amount:    item.Amount,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		}
		yearlyData = append(yearlyData, yearly)
	}

	return yearlyData, nil
}

// GetRealtimeData 获取实时数据 - 空实现
func (t *TongHuaShunCollector) GetRealtimeData(tsCodes []string) ([]model.DailyData, error) {
	t.logger.Infof("TongHuaShun GetRealtimeData for %d stocks - 功能暂未实现", len(tsCodes))
	return []model.DailyData{}, fmt.Errorf("TongHuaShun GetRealtimeData not implemented yet")
}

// GetTodayData 获取当日数据
func (t *TongHuaShunCollector) GetTodayData(tsCode string) (*model.DailyData, string, error) {
	t.logger.Infof("TongHuaShun GetTodayData for %s", tsCode)

	// 解析股票代码
	symbol, market, err := t.parseStockCode(tsCode)
	if err != nil {
		return nil, "", fmt.Errorf("invalid tsCode format: %s", tsCode)
	}

	// 构建同花顺股票代码格式
	thsCode := t.buildTHSStockCode(symbol, market)
	if thsCode == "" {
		return nil, "", fmt.Errorf("unsupported market for TongHuaShun: %s", market)
	}

	// 构建请求URL - 基于提供的curl命令
	requestURL := fmt.Sprintf("https://d.10jqka.com.cn/v6/line/%s/%s/defer/today.js", thsCode, THSKLineTypeDaily)

	// 发送请求
	resp, err := t.makeTodayDataRequest(requestURL)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch today data: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response body: %w", err)
	}

	// 解析响应数据
	todayData, name, err := t.parseTodayDataResponse(tsCode, thsCode, string(body))
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse today data response: %w", err)
	}

	return todayData, name, nil
}

// makeTodayDataRequest 发送当日数据请求
func (t *TongHuaShunCollector) makeTodayDataRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// 设置请求头 - 完全按照curl请求设置
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("If-Modified-Since", "Sun, 05 Oct 2025 06:43:51 GMT")
	req.Header.Set("Referer", "https://stockpage.10jqka.com.cn/")
	req.Header.Set("Sec-Fetch-Dest", "script")
	req.Header.Set("Sec-Fetch-Mode", "no-cors")
	req.Header.Set("Sec-Fetch-Site", "same-site")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36")
	req.Header.Set("sec-ch-ua", `"Not;A=Brand";v="99", "Google Chrome";v="139", "Chromium";v="139"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"macOS"`)

	// 设置Cookie - 基于curl请求中的格式
	timestamp := time.Now().Unix()
	cookieValue := fmt.Sprintf("Hm_lvt_722143063e4892925903024537075d0d=%d; "+
		"HMACCOUNT=17C55F0F7B5ABE69; Hm_lvt_929f8b362150b1f77b477230541dbbc2=%d; "+
		"Hm_lvt_78c58f01938e4d85eaf619eae71b4ed1=%d; Hm_lvt_69929b9dce4c22a060bd22d703b2a280=%d; "+
		"spversion=20130314; historystock=601899%%7C*%%7C001208%%7C*%%7C600930%%7C*%%7C001201; H"+
		"m_lpvt_929f8b362150b1f77b477230541dbbc2=%d; Hm_lpvt_69929b9dce4c22a060bd22d703b2a280=%d; "+
		"Hm_lpvt_722143063e4892925903024537075d0d=%d; Hm_ck_%d=42; "+
		"Hm_lpvt_78c58f01938e4d85eaf619eae71b4ed1=%d; v=%s", timestamp, timestamp, timestamp, timestamp,
		timestamp, timestamp, timestamp, timestamp-1, timestamp, GenerateWencaiToken())

	req.Header.Set("Cookie", cookieValue)

	// 应用限流
	if err := t.limiter.Wait(context.Background()); err != nil {
		return nil, fmt.Errorf("rate limit wait failed: %v", err)
	}

	t.logger.Debugf("Making today data request to TongHuaShun: %s", url)

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return resp, nil
}

// parseTodayDataResponse 解析当日数据响应
func (t *TongHuaShunCollector) parseTodayDataResponse(tsCode, thsCode, res string) (*model.DailyData, string, error) {
	// 同花顺返回的是JavaScript格式，需要提取数据部分
	// 实际格式: quotebridge_v6_line_hs_601899_11_defer_today({"hs_601899": {...}})

	// 构建回调函数名（注意这里可能是11而不是01）
	callbackName := fmt.Sprintf("quotebridge_v6_line_%s_", thsCode)

	// 找到回调函数的开始位置
	startIdx := strings.Index(res, callbackName)
	if startIdx == -1 {
		return nil, "", fmt.Errorf("callback function not found in response")
	}

	// 找到参数开始的位置
	parenIdx := strings.Index(res[startIdx:], "(")
	if parenIdx == -1 {
		return nil, "", fmt.Errorf("callback parameters not found")
	}

	// 提取JSON部分
	jsonStart := startIdx + parenIdx + 1
	jsonEnd := strings.LastIndex(res, ")")
	if jsonEnd == -1 || jsonEnd <= jsonStart {
		return nil, "", fmt.Errorf("invalid callback format")
	}

	jsonStr := res[jsonStart:jsonEnd]

	// 解析JSON - 实际格式是包含股票代码作为key的对象
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &response); err != nil {
		return nil, "", fmt.Errorf("failed to parse JSON: %w", err)
	}

	// 获取股票数据（key是thsCode，如"hs_601899"）
	stockData, exists := response[thsCode]
	if !exists {
		return nil, "", fmt.Errorf("stock data not found for %s", thsCode)
	}

	// 将stockData转换为map[string]interface{}
	dataMap, ok := stockData.(map[string]interface{})
	if !ok {
		return nil, "", fmt.Errorf("invalid stock data format")
	}

	// 解析各个字段
	// "1": "20250930" - 交易日期
	// "7": "27.48" - 开盘价
	// "8": "29.88" - 最高价
	// "9": "27.31" - 最低价
	// "11": "29.44" - 收盘价
	// "13": 785682260 - 成交量
	// "19": "22670644000.00" - 成交额

	tradeDateStr := t.getStringValue(dataMap, "1")
	openStr := t.getStringValue(dataMap, "7")
	highStr := t.getStringValue(dataMap, "8")
	lowStr := t.getStringValue(dataMap, "9")
	closeStr := t.getStringValue(dataMap, "11")
	volumeStr := t.getStringValue(dataMap, "13")
	amountStr := t.getStringValue(dataMap, "19")

	// 转换数据类型
	tradeDate, err := strconv.Atoi(tradeDateStr)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse trade date: %v", err)
	}

	open := t.parseFloat(openStr)
	high := t.parseFloat(highStr)
	low := t.parseFloat(lowStr)
	over := t.parseFloat(closeStr)
	volume := t.parseInt64(volumeStr)
	amount := t.parseFloat(amountStr)

	// 创建DailyData对象
	todayData := &model.DailyData{
		TsCode:    tsCode,
		TradeDate: tradeDate,
		Open:      open,
		High:      high,
		Low:       low,
		Close:     over,
		Volume:    volume,
		Amount:    amount,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return todayData, t.getStringValue(dataMap, "name"), nil
}

// getStringValue 从map中获取字符串值，支持多种类型转换
func (t *TongHuaShunCollector) getStringValue(dataMap map[string]interface{}, key string) string {
	value, exists := dataMap[key]
	if !exists || value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case float64:
		// 如果是整数，不显示小数点
		if v == float64(int64(v)) {
			return fmt.Sprintf("%.0f", v)
		}
		return fmt.Sprintf("%f", v)
	case int:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case bool:
		if v {
			return "1"
		}
		return "0"
	default:
		return fmt.Sprintf("%v", v)
	}
}

// GetThisWeekData 获取本周数据
func (t *TongHuaShunCollector) GetThisWeekData(tsCode string) (*model.WeeklyData, error) {
	t.logger.Infof("TongHuaShun GetThisWeekData for %s", tsCode)

	// 解析股票代码
	symbol, market, err := t.parseStockCode(tsCode)
	if err != nil {
		return nil, fmt.Errorf("invalid tsCode format: %s", tsCode)
	}

	// 构建同花顺股票代码格式
	thsCode := t.buildTHSStockCode(symbol, market)
	if thsCode == "" {
		return nil, fmt.Errorf("unsupported market for TongHuaShun: %s", market)
	}

	// 构建请求URL - 使用周K线类型
	requestURL := fmt.Sprintf("https://d.10jqka.com.cn/v6/line/%s/%s/defer/today.js", thsCode, THSKLineTypeWeekly)

	// 发送请求
	resp, err := t.makeTodayDataRequest(requestURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch this week data: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 解析响应数据
	weekData, err := t.parseThisWeekDataResponse(tsCode, thsCode, string(body))
	if err != nil {
		return nil, fmt.Errorf("failed to parse this week data response: %w", err)
	}

	return weekData, nil
}

// GetThisMonthData 获取本月数据
func (t *TongHuaShunCollector) GetThisMonthData(tsCode string) (*model.MonthlyData, error) {
	t.logger.Infof("TongHuaShun GetThisMonthData for %s", tsCode)

	// 解析股票代码
	symbol, market, err := t.parseStockCode(tsCode)
	if err != nil {
		return nil, fmt.Errorf("invalid tsCode format: %s", tsCode)
	}

	// 构建同花顺股票代码格式
	thsCode := t.buildTHSStockCode(symbol, market)
	if thsCode == "" {
		return nil, fmt.Errorf("unsupported market for TongHuaShun: %s", market)
	}

	// 构建请求URL - 使用月K线类型
	requestURL := fmt.Sprintf("https://d.10jqka.com.cn/v6/line/%s/%s/defer/today.js", thsCode, THSKLineTypeMonthly)

	// 发送请求
	resp, err := t.makeTodayDataRequest(requestURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch this month data: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 解析响应数据
	monthData, err := t.parseThisMonthDataResponse(tsCode, thsCode, string(body))
	if err != nil {
		return nil, fmt.Errorf("failed to parse this month data response: %w", err)
	}

	return monthData, nil
}

// GetThisQuarterData 获取本季数据
func (t *TongHuaShunCollector) GetThisQuarterData(tsCode string) (*model.QuarterlyData, error) {
	t.logger.Infof("TongHuaShun GetThisQuarterData for %s", tsCode)

	// 解析股票代码
	symbol, market, err := t.parseStockCode(tsCode)
	if err != nil {
		return nil, fmt.Errorf("invalid tsCode format: %s", tsCode)
	}

	// 构建同花顺股票代码格式
	thsCode := t.buildTHSStockCode(symbol, market)
	if thsCode == "" {
		return nil, fmt.Errorf("unsupported market for TongHuaShun: %s", market)
	}

	// 构建请求URL - 使用季K线类型
	requestURL := fmt.Sprintf("https://d.10jqka.com.cn/v6/line/%s/%s/defer/today.js", thsCode, THSKLineTypeQuarterly)

	// 发送请求
	resp, err := t.makeTodayDataRequest(requestURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch this quarter data: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 解析响应数据
	quarterData, err := t.parseThisQuarterDataResponse(tsCode, thsCode, string(body))
	if err != nil {
		return nil, fmt.Errorf("failed to parse this quarter data response: %w", err)
	}

	return quarterData, nil
}

// GetThisYearData 获取本年数据
func (t *TongHuaShunCollector) GetThisYearData(tsCode string) (*model.YearlyData, error) {
	t.logger.Infof("TongHuaShun GetThisYearData for %s", tsCode)

	// 解析股票代码
	symbol, market, err := t.parseStockCode(tsCode)
	if err != nil {
		return nil, fmt.Errorf("invalid tsCode format: %s", tsCode)
	}

	// 构建同花顺股票代码格式
	thsCode := t.buildTHSStockCode(symbol, market)
	if thsCode == "" {
		return nil, fmt.Errorf("unsupported market for TongHuaShun: %s", market)
	}

	// 构建请求URL - 使用年K线类型
	requestURL := fmt.Sprintf("https://d.10jqka.com.cn/v6/line/%s/%s/defer/today.js", thsCode, THSKLineTypeYearly)

	// 发送请求
	resp, err := t.makeTodayDataRequest(requestURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch this year data: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 解析响应数据
	yearData, err := t.parseThisYearDataResponse(tsCode, thsCode, string(body))
	if err != nil {
		return nil, fmt.Errorf("failed to parse this year data response: %w", err)
	}

	return yearData, nil
}

// parseThisWeekDataResponse 解析本周数据响应
func (t *TongHuaShunCollector) parseThisWeekDataResponse(tsCode, thsCode, res string) (*model.WeeklyData, error) {
	// 使用通用解析方法
	dailyData, _, err := t.parseTodayDataResponse(tsCode, thsCode, res)
	if err != nil {
		return nil, err
	}

	// 转换为周K线数据格式
	weekData := &model.WeeklyData{
		TsCode:    dailyData.TsCode,
		TradeDate: dailyData.TradeDate,
		Open:      dailyData.Open,
		High:      dailyData.High,
		Low:       dailyData.Low,
		Close:     dailyData.Close,
		Volume:    dailyData.Volume,
		Amount:    dailyData.Amount,
		CreatedAt: dailyData.CreatedAt,
		UpdatedAt: dailyData.UpdatedAt,
	}

	return weekData, nil
}

// parseThisMonthDataResponse 解析本月数据响应
func (t *TongHuaShunCollector) parseThisMonthDataResponse(tsCode, thsCode, res string) (*model.MonthlyData, error) {
	// 使用通用解析方法
	dailyData, _, err := t.parseTodayDataResponse(tsCode, thsCode, res)
	if err != nil {
		return nil, err
	}

	// 转换为月K线数据格式
	monthData := &model.MonthlyData{
		TsCode:    dailyData.TsCode,
		TradeDate: dailyData.TradeDate,
		Open:      dailyData.Open,
		High:      dailyData.High,
		Low:       dailyData.Low,
		Close:     dailyData.Close,
		Volume:    dailyData.Volume,
		Amount:    dailyData.Amount,
		CreatedAt: dailyData.CreatedAt,
		UpdatedAt: dailyData.UpdatedAt,
	}

	return monthData, nil
}

// parseThisQuarterDataResponse 解析本季数据响应
func (t *TongHuaShunCollector) parseThisQuarterDataResponse(tsCode, thsCode, res string) (*model.QuarterlyData, error) {
	// 使用通用解析方法
	dailyData, _, err := t.parseTodayDataResponse(tsCode, thsCode, res)
	if err != nil {
		return nil, err
	}

	// 转换为季K线数据格式
	quarterData := &model.QuarterlyData{
		TsCode:    dailyData.TsCode,
		TradeDate: dailyData.TradeDate,
		Open:      dailyData.Open,
		High:      dailyData.High,
		Low:       dailyData.Low,
		Close:     dailyData.Close,
		Volume:    dailyData.Volume,
		Amount:    dailyData.Amount,
		CreatedAt: dailyData.CreatedAt,
		UpdatedAt: dailyData.UpdatedAt,
	}

	return quarterData, nil
}

// parseThisYearDataResponse 解析本年数据响应
func (t *TongHuaShunCollector) parseThisYearDataResponse(tsCode, thsCode, res string) (*model.YearlyData, error) {
	// 使用通用解析方法
	dailyData, _, err := t.parseTodayDataResponse(tsCode, thsCode, res)
	if err != nil {
		return nil, err
	}

	// 转换为年K线数据格式
	yearData := &model.YearlyData{
		TsCode:    dailyData.TsCode,
		TradeDate: dailyData.TradeDate,
		Open:      dailyData.Open,
		High:      dailyData.High,
		Low:       dailyData.Low,
		Close:     dailyData.Close,
		Volume:    dailyData.Volume,
		Amount:    dailyData.Amount,
		CreatedAt: dailyData.CreatedAt,
		UpdatedAt: dailyData.UpdatedAt,
	}

	return yearData, nil
}

// GetPerformanceReports 获取业绩报表数据 - 空实现
func (t *TongHuaShunCollector) GetPerformanceReports(tsCode string) ([]model.PerformanceReport, error) {
	t.logger.Infof("TongHuaShun GetPerformanceReports for %s - 功能暂未实现", tsCode)
	return []model.PerformanceReport{}, fmt.Errorf("TongHuaShun GetPerformanceReports not implemented yet")
}

// GetLatestPerformanceReport 获取最新业绩报表数据 - 空实现
func (t *TongHuaShunCollector) GetLatestPerformanceReport(tsCode string) (*model.PerformanceReport, error) {
	t.logger.Infof("TongHuaShun GetLatestPerformanceReport for %s - 功能暂未实现", tsCode)
	return nil, fmt.Errorf("TongHuaShun GetLatestPerformanceReport not implemented yet")
}

// GetShareholderCounts 获取股东户数数据 - 空实现
func (t *TongHuaShunCollector) GetShareholderCounts(tsCode string) ([]model.ShareholderCount, error) {
	t.logger.Infof("TongHuaShun GetShareholderCounts for %s - 功能暂未实现", tsCode)
	return []model.ShareholderCount{}, fmt.Errorf("TongHuaShun GetShareholderCounts not implemented yet")
}

// GetLatestShareholderCount 获取最新股东户数数据 - 空实现
func (t *TongHuaShunCollector) GetLatestShareholderCount(tsCode string) (*model.ShareholderCount, error) {
	t.logger.Infof("TongHuaShun GetLatestShareholderCount for %s - 功能暂未实现", tsCode)
	return nil, fmt.Errorf("TongHuaShun GetLatestShareholderCount not implemented yet")
}

// SetRateLimit 动态设置限流速率
func (t *TongHuaShunCollector) SetRateLimit(requestsPerSecond int) {
	if requestsPerSecond <= 0 {
		requestsPerSecond = 1 // 最小值为1
	}

	t.Config.RateLimit = requestsPerSecond
	t.limiter.SetLimit(rate.Limit(requestsPerSecond))
	t.limiter.SetBurst(requestsPerSecond * 2) // 突发容量为速率的2倍

	t.logger.Infof("TongHuaShun rate limit updated to %d requests/second", requestsPerSecond)
}

// GetRateLimit 获取当前限流速率
func (t *TongHuaShunCollector) GetRateLimit() int {
	return t.Config.RateLimit
}

// GetRateLimitStats 获取限流统计信息
func (t *TongHuaShunCollector) GetRateLimitStats() map[string]interface{} {
	return map[string]interface{}{
		"rate_limit":    t.Config.RateLimit,
		"current_limit": float64(t.limiter.Limit()),
		"burst_size":    t.limiter.Burst(),
		"tokens":        t.limiter.Tokens(), // 当前可用令牌数
	}
}

// updateUserAgentAndCookie 更新随机User-Agent和Cookie
func (t *TongHuaShunCollector) updateUserAgentAndCookie() {
	t.currentUA = t.userAgentGen.GenerateUserAgent()
	t.currentCookie = t.cookieGen.GenerateCookie()
	t.lastUpdateTime = time.Now()

	t.logger.Debugf("TongHuaShun updated User-Agent: %s", t.currentUA[:50]+"...")
	t.logger.Debugf("TongHuaShun updated Cookie length: %d characters", len(t.currentCookie))
}

// getPlatformFromUA 从User-Agent中提取平台信息
func (t *TongHuaShunCollector) getPlatformFromUA(userAgent string) string {
	if userAgent == "" {
		return `"Unknown"`
	}

	switch {
	case contains(userAgent, "Windows"):
		return `"Windows"`
	case contains(userAgent, "Macintosh") || contains(userAgent, "Mac OS X"):
		return `"macOS"`
	case contains(userAgent, "Linux"):
		return `"Linux"`
	default:
		return `"Unknown"`
	}
}

// contains 检查字符串是否包含子字符串（简化版本）
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && findSubstring(s, substr)))
}

// findSubstring 查找子字符串
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// GetCurrentUserAgent 获取当前使用的User-Agent
func (t *TongHuaShunCollector) GetCurrentUserAgent() string {
	return t.currentUA
}

// GetCurrentCookie 获取当前使用的Cookie
func (t *TongHuaShunCollector) GetCurrentCookie() string {
	return t.currentCookie
}

// ForceUpdateUserAgentAndCookie 强制更新User-Agent和Cookie
func (t *TongHuaShunCollector) ForceUpdateUserAgentAndCookie() {
	t.updateUserAgentAndCookie()
	t.logger.Info("TongHuaShun forced update of User-Agent and Cookie")
}

// parseStockCode 解析股票代码，返回代码和市场
func (t *TongHuaShunCollector) parseStockCode(tsCode string) (string, string, error) {
	parts := strings.Split(tsCode, ".")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid tsCode format: %s", tsCode)
	}
	return parts[0], parts[1], nil
}

// buildTHSStockCode 构建同花顺股票代码格式
func (t *TongHuaShunCollector) buildTHSStockCode(symbol, market string) string {
	switch market {
	case "SH":
		return "hs_" + symbol // 沪市
	case "SZ":
		return "hs_" + symbol // 深市
	default:
		return ""
	}
}

// THSKLineData 同花顺K线数据结构
type THSKLineData struct {
	TsCode    string
	TradeDate int
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    int64
	Amount    float64
	CreatedAt time.Time
	UpdatedAt time.Time
}

// getKLineData 通用K线数据获取方法
func (t *TongHuaShunCollector) getKLineData(tsCode, klineType string, startDate, endDate time.Time) ([]THSKLineData, error) {
	// 解析股票代码
	symbol, market, err := t.parseStockCode(tsCode)
	if err != nil {
		return nil, fmt.Errorf("invalid tsCode format: %s", tsCode)
	}

	// 构建同花顺股票代码格式
	thsCode := t.buildTHSStockCode(symbol, market)
	if thsCode == "" {
		return nil, fmt.Errorf("unsupported market for TongHuaShun: %s", market)
	}

	// 构建请求URL
	requestURL := fmt.Sprintf("https://d.10jqka.com.cn/v6/line/%s/%s/all.js", thsCode, klineType)

	// 发送请求
	resp, err := t.makeKLineRequest(requestURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch K-line data: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 解析响应数据
	klineData, err := t.parseKLineResponse(tsCode, thsCode, klineType, string(body), startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse K-line response: %w", err)
	}

	return klineData, nil
}

// makeKLineRequest 发送K线数据请求
func (t *TongHuaShunCollector) makeKLineRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// 设置请求头 - 根据提供的curl命令
	req.Header.Set("sec-ch-ua-platform", `"macOS"`)
	req.Header.Set("Referer", "https://stockpage.10jqka.com.cn/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36")
	req.Header.Set("sec-ch-ua", `"Not;A=Brand";v="99", "Google Chrome";v="139", "Chromium";v="139"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")

	// 应用限流
	if err := t.limiter.Wait(context.Background()); err != nil {
		return nil, fmt.Errorf("rate limit wait failed: %v", err)
	}

	t.logger.Debugf("Making K-line request to TongHuaShun: %s", url)

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return resp, nil
}

// parseKLineResponse 解析同花顺K线响应数据
func (t *TongHuaShunCollector) parseKLineResponse(tsCode, thsCode, klineType, res string, startDate, endDate time.Time) ([]THSKLineData, error) {
	// 同花顺返回的是JavaScript格式，需要提取数据部分
	// 示例格式: quotebridge_v6_line_hs_001208_01_all({"data":"20240101,10.5,10.8,10.2,10.6,1000000;..."})

	// 构建回调函数名
	callbackName := fmt.Sprintf("quotebridge_v6_line_%s_%s_all(", thsCode, klineType)

	res = strings.TrimPrefix(res, callbackName)
	res = strings.TrimSuffix(res, ")")

	// 解析JSON
	var response struct {
		Start    string  `json:"start"`
		SortYear [][]int `json:"sortYear"`
		Price    string  `json:"price"`
		Volume   string  `json:"volumn"`
		Dates    string  `json:"dates"`
	}

	if err := json.Unmarshal([]byte(res), &response); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	prices := strings.Split(response.Price, ",")
	volumes := strings.Split(response.Volume, ",")
	dates := strings.Split(response.Dates, ",")

	// 解析K线数据
	var klineData = make([]THSKLineData, 0, len(dates))
	var index int
	for _, arr := range response.SortYear {
		year, num := arr[0], arr[1]
		for num > 0 && len(dates) > index && len(prices) > index*4 && len(volumes) > index {
			var data = THSKLineData{
				TsCode:    tsCode,
				Volume:    0,
				Amount:    0,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			td, err := time.ParseInLocation("20060102", fmt.Sprintf("%d%s", year, dates[index]), time.Local)
			if err != nil {
				return nil, err
			}

			// 检查日期范围
			if !startDate.IsZero() && td.Before(startDate) {
				num--
				index++
				continue
			}
			if !endDate.IsZero() && td.After(endDate) {
				num--
				index++
				continue
			}

			data.TradeDate, _ = strconv.Atoi(fmt.Sprintf("%d%s", year, dates[index]))
			low, _ := strconv.Atoi(prices[index*4])
			open, _ := strconv.Atoi(prices[index*4+1])
			high, _ := strconv.Atoi(prices[index*4+2])
			over, _ := strconv.Atoi(prices[index*4+3])
			volume, _ := strconv.ParseInt(volumes[index], 10, 64)

			data.Low = float64(low) / 100
			data.Open = float64(low+open) / 100
			data.High = float64(low+high) / 100
			data.Close = float64(low+over) / 100
			data.Volume = volume

			klineData = append(klineData, data)

			num--
			index++
		}
	}

	return klineData, nil
}

// filterDataByDateRange 根据时间范围过滤数据
func (t *TongHuaShunCollector) filterDataByDateRange(data []model.DailyData, startDate, endDate time.Time) []model.DailyData {
	var filtered []model.DailyData

	for _, item := range data {
		// 将int类型的交易日期转换为time.Time
		tradeDateStr := fmt.Sprintf("%d", item.TradeDate)
		date, err := time.Parse("20060102", tradeDateStr)
		if err != nil {
			t.logger.Warnf("Failed to parse trade date %d: %v", item.TradeDate, err)
			continue
		}

		// 检查是否在时间范围内
		if !startDate.IsZero() && date.Before(startDate) {
			continue
		}
		if !endDate.IsZero() && date.After(endDate) {
			continue
		}

		filtered = append(filtered, item)
	}

	return filtered
}

// parseFloat 安全地将字符串转换为float64
func (t *TongHuaShunCollector) parseFloat(s string) float64 {
	if s == "" || s == "-" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}

// parseInt64 安全地将字符串转换为int64
func (t *TongHuaShunCollector) parseInt64(s string) int64 {
	if s == "" || s == "-" {
		return 0
	}
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return i
}
