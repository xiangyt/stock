package collector

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"stock/internal/model"
	"stock/internal/utils"
)

// HTTPCollector HTTP数据采集器
type HTTPCollector struct {
	BaseCollector
	client *http.Client
	logger *utils.Logger
}

// NewHTTPCollector 创建HTTP采集器
func NewHTTPCollector(config CollectorConfig, logger *utils.Logger) *HTTPCollector {
	return &HTTPCollector{
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
func (h *HTTPCollector) Connect() error {
	h.logger.Infof("Connecting to %s...", h.Config.Name)

	// 测试连接
	if err := h.testConnection(); err != nil {
		h.logger.Errorf("Failed to connect to %s: %v", h.Config.Name, err)
		return err
	}

	h.Connected = true
	h.logger.Infof("Successfully connected to %s", h.Config.Name)
	return nil
}

// testConnection 测试连接
func (h *HTTPCollector) testConnection() error {
	// 发送测试请求
	req, err := http.NewRequest("GET", h.Config.BaseURL, nil)
	if err != nil {
		return err
	}

	// 添加请求头
	for key, value := range h.Config.Headers {
		req.Header.Set(key, value)
	}

	// 添加认证token
	if h.Config.Token != "" {
		req.Header.Set("Authorization", "Bearer "+h.Config.Token)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	return nil
}

// makeRequest 发送HTTP请求
func (h *HTTPCollector) makeRequest(method, url string, body io.Reader) (*http.Response, error) {
	if !h.Connected {
		return nil, fmt.Errorf("collector not connected")
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	// 添加请求头
	for key, value := range h.Config.Headers {
		req.Header.Set(key, value)
	}

	// 添加认证token
	if h.Config.Token != "" {
		req.Header.Set("Authorization", "Bearer "+h.Config.Token)
	}

	// 设置Content-Type
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	h.logger.Debugf("Making request: %s %s", method, url)

	resp, err := h.client.Do(req)
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

// GetStockList 获取股票列表
func (h *HTTPCollector) GetStockList() ([]model.Stock, error) {
	h.logger.Info("Fetching stock list...")

	url := fmt.Sprintf("%s/stock_basic", h.Config.BaseURL)
	resp, err := h.makeRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Code int `json:"code"`
		Data struct {
			Items []struct {
				TsCode   string `json:"ts_code"`
				Symbol   string `json:"symbol"`
				Name     string `json:"name"`
				Area     string `json:"area"`
				Industry string `json:"industry"`
				Market   string `json:"market"`
				ListDate string `json:"list_date"`
			} `json:"items"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("API error: code %d", result.Code)
	}

	stocks := make([]model.Stock, 0, len(result.Data.Items))
	for _, item := range result.Data.Items {
		var listDate *time.Time
		if item.ListDate != "" {
			if t, err := time.Parse("20060102", item.ListDate); err == nil {
				listDate = &t
			}
		}

		stock := model.Stock{
			TsCode:   item.TsCode,
			Symbol:   item.Symbol,
			Name:     item.Name,
			Area:     item.Area,
			Industry: item.Industry,
			Market:   item.Market,
			ListDate: listDate,
			IsActive: true,
		}
		stocks = append(stocks, stock)
	}

	h.logger.Infof("Fetched %d stocks", len(stocks))
	return stocks, nil
}

// GetStockDetail 获取股票详情
func (h *HTTPCollector) GetStockDetail(tsCode string) (*model.Stock, error) {
	h.logger.Infof("Fetching stock detail for %s", tsCode)

	url := fmt.Sprintf("%s/stock_basic?ts_code=%s", h.Config.BaseURL, tsCode)
	resp, err := h.makeRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Code int `json:"code"`
		Data struct {
			TsCode   string `json:"ts_code"`
			Symbol   string `json:"symbol"`
			Name     string `json:"name"`
			Area     string `json:"area"`
			Industry string `json:"industry"`
			Market   string `json:"market"`
			ListDate string `json:"list_date"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("API error: code %d", result.Code)
	}

	var listDate *time.Time
	if result.Data.ListDate != "" {
		if t, err := time.Parse("20060102", result.Data.ListDate); err == nil {
			listDate = &t
		}
	}

	stock := &model.Stock{
		TsCode:   result.Data.TsCode,
		Symbol:   result.Data.Symbol,
		Name:     result.Data.Name,
		Area:     result.Data.Area,
		Industry: result.Data.Industry,
		Market:   result.Data.Market,
		ListDate: listDate,
		IsActive: true,
	}

	return stock, nil
}

// GetStockData 获取股票历史数据
func (h *HTTPCollector) GetStockData(tsCode string, startDate, endDate time.Time) ([]model.DailyData, error) {
	return h.GetDailyKLine(tsCode, startDate, endDate)
}

// GetDailyKLine 获取日K线数据
func (h *HTTPCollector) GetDailyKLine(tsCode string, startDate, endDate time.Time) ([]model.DailyData, error) {
	h.logger.Infof("Fetching daily K-line data for %s from %s to %s", tsCode, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	url := fmt.Sprintf("%s/daily?ts_code=%s&start_date=%s&end_date=%s",
		h.Config.BaseURL,
		tsCode,
		startDate.Format("20060102"),
		endDate.Format("20060102"))

	resp, err := h.makeRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Code int `json:"code"`
		Data struct {
			Items []struct {
				TsCode    string  `json:"ts_code"`
				TradeDate string  `json:"trade_date"`
				Open      float64 `json:"open"`
				High      float64 `json:"high"`
				Low       float64 `json:"low"`
				Close     float64 `json:"close"`
				Volume    int64   `json:"vol"`
				Amount    float64 `json:"amount"`
			} `json:"items"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("API error: code %d", result.Code)
	}

	dailyData := make([]model.DailyData, 0, len(result.Data.Items))
	for _, item := range result.Data.Items {
		tradeDate, err := time.Parse("20060102", item.TradeDate)
		if err != nil {
			h.logger.Warnf("Invalid trade date format: %s", item.TradeDate)
			continue
		}

		// 转换为YYYYMMDD格式的int
		tradeDateInt := tradeDate.Year()*10000 + int(tradeDate.Month())*100 + tradeDate.Day()

		data := model.DailyData{
			TsCode:    item.TsCode,
			TradeDate: tradeDateInt,
			Open:      item.Open,
			High:      item.High,
			Low:       item.Low,
			Close:     item.Close,
			Volume:    item.Volume,
			Amount:    item.Amount,
		}
		dailyData = append(dailyData, data)
	}

	h.logger.Infof("Fetched %d daily data records for %s", len(dailyData), tsCode)
	return dailyData, nil
}

// GetRealtimeData 获取实时数据
func (h *HTTPCollector) GetRealtimeData(tsCodes []string) ([]model.DailyData, error) {
	h.logger.Infof("Fetching realtime data for %d stocks", len(tsCodes))

	url := fmt.Sprintf("%s/realtime?ts_codes=%s", h.Config.BaseURL, strings.Join(tsCodes, ","))
	resp, err := h.makeRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Code int `json:"code"`
		Data struct {
			Items []struct {
				TsCode string  `json:"ts_code"`
				Open   float64 `json:"open"`
				High   float64 `json:"high"`
				Low    float64 `json:"low"`
				Close  float64 `json:"close"`
				Volume int64   `json:"volume"`
				Amount float64 `json:"amount"`
			} `json:"items"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("API error: code %d", result.Code)
	}

	realtimeData := make([]model.DailyData, 0, len(result.Data.Items))
	now := time.Now()
	// 转换当前日期为YYYYMMDD格式的int
	nowDateInt := now.Year()*10000 + int(now.Month())*100 + now.Day()

	for _, item := range result.Data.Items {
		data := model.DailyData{
			TsCode:    item.TsCode,
			TradeDate: nowDateInt,
			Open:      item.Open,
			High:      item.High,
			Low:       item.Low,
			Close:     item.Close,
			Volume:    item.Volume,
			Amount:    item.Amount,
		}
		realtimeData = append(realtimeData, data)
	}

	h.logger.Infof("Fetched realtime data for %d stocks", len(realtimeData))
	return realtimeData, nil
}

// GetPerformanceReports 获取业绩报表数据
func (h *HTTPCollector) GetPerformanceReports(tsCode string) ([]model.PerformanceReport, error) {
	h.logger.Infof("Fetching performance reports for %s", tsCode)

	// HTTP采集器暂不支持业绩报表数据获取
	return nil, fmt.Errorf("performance reports not supported by HTTP collector")
}

// GetLatestPerformanceReport 获取最新业绩报表数据
func (h *HTTPCollector) GetLatestPerformanceReport(tsCode string) (*model.PerformanceReport, error) {
	reports, err := h.GetPerformanceReports(tsCode)
	if err != nil {
		return nil, err
	}

	if len(reports) == 0 {
		return nil, fmt.Errorf("no performance reports found for %s", tsCode)
	}

	// 返回最新的业绩报表数据
	latest := reports[0]
	for _, report := range reports {
		if report.ReportDate.After(latest.ReportDate) {
			latest = report
		}
	}

	return &latest, nil
}
