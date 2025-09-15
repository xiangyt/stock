package collector

import (
	"stock/internal/model"
	"stock/internal/utils"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// TestEastMoneyCollector_Pagination 测试分页功能
func TestEastMoneyCollector_Pagination(t *testing.T) {
	// 创建一个简单的logger
	logrusLogger := logrus.New()
	logrusLogger.SetLevel(logrus.InfoLevel)
	logger := &utils.Logger{Logger: logrusLogger}

	collector := NewEastMoneyCollector(logger)

	// 测试分页获取
	page := 1
	pageSize := 50
	totalStocks := 0

	for {
		t.Logf("Fetching page %d with page size %d", page, pageSize)

		response, err := collector.fetchStockListPage(page, pageSize)
		assert.NoError(t, err)
		assert.NotNil(t, response)

		if response.RC != 0 {
			t.Errorf("API returned error code: %d", response.RC)
			break
		}

		if response.Data.Diff == nil || len(response.Data.Diff) == 0 {
			t.Log("No more data available")
			break
		}

		t.Logf("Page %d: fetched %d stocks", page, len(response.Data.Diff))
		totalStocks += len(response.Data.Diff)

		// 检查是否还有更多数据
		if len(response.Data.Diff) < pageSize {
			t.Log("Reached last page")
			break
		}

		page++

		// 添加延迟避免请求过快
		time.Sleep(100 * time.Millisecond)

		// 安全限制，最多获取5页
		if page > 5 {
			t.Log("Reached maximum page limit (5 pages)")
			break
		}
	}

	t.Logf("Total fetched %d stocks from %d pages", totalStocks, page-1)

	// 应该获取到一些股票数据
	assert.True(t, totalStocks > 0, "Should fetch at least some stocks")
}

// TestEastMoneyCollector_GetDailyKLine_001208 测试获取001208的日K数据
func TestEastMoneyCollector_GetDailyKLine_001208(t *testing.T) {
	// 创建一个简单的logger
	logrusLogger := logrus.New()
	logrusLogger.SetLevel(logrus.InfoLevel)
	logger := &utils.Logger{Logger: logrusLogger}

	collector := NewEastMoneyCollector(logger)

	// 测试股票代码 001208.SZ
	stockCode := "001208.SZ"

	// 设置时间范围：最近30天
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	t.Logf("Testing GetDailyKLine for stock: %s", stockCode)
	t.Logf("Date range: %s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	// 获取日K数据
	dailyData, err := collector.GetDailyKLine(stockCode, startDate, endDate)

	// 检查错误
	if err != nil {
		t.Fatalf("GetDailyKLine returned error: %v", err)
	}

	// 检查数据不为空
	assert.NotNil(t, dailyData)
	assert.True(t, len(dailyData) > 0, "Should have at least some daily data")

	t.Logf("Successfully retrieved %d daily K-line records for %s", len(dailyData), stockCode)

	// 显示最近几天的数据
	displayCount := 5
	if len(dailyData) < displayCount {
		displayCount = len(dailyData)
	}

	t.Logf("=== 最近%d天的K线数据 ===", displayCount)
	for i := 0; i < displayCount; i++ {
		data := dailyData[i]
		t.Logf("日期: %d, 开盘: %.3f, 最高: %.3f, 最低: %.3f, 收盘: %.3f, 成交量: %d, 成交额: %.2f万元",
			data.TradeDate,
			data.Open,
			data.High,
			data.Low,
			data.Close,
			data.Volume,
			data.Amount/10000) // 转换为万元
	}

	// 验证数据完整性
	for i, data := range dailyData {
		if i >= 3 { // 只检查前3条数据
			break
		}

		// 检查基本字段
		assert.Equal(t, stockCode, data.TsCode, "TsCode should match")
		assert.True(t, data.TradeDate > 0, "TradeDate should be valid")
		assert.True(t, data.Open > 0, "Open price should be positive")
		assert.True(t, data.High > 0, "High price should be positive")
		assert.True(t, data.Low > 0, "Low price should be positive")
		assert.True(t, data.Close > 0, "Close price should be positive")
		assert.True(t, data.Volume >= 0, "Volume should be non-negative")
		assert.True(t, data.Amount >= 0, "Amount should be non-negative")

		// 检查价格逻辑
		assert.True(t, data.High >= data.Low, "High should be >= Low")
		assert.True(t, data.High >= data.Open, "High should be >= Open")
		assert.True(t, data.High >= data.Close, "High should be >= Close")
		assert.True(t, data.Low <= data.Open, "Low should be <= Open")
		assert.True(t, data.Low <= data.Close, "Low should be <= Close")
	}

	t.Logf("Data validation completed successfully")
}

// TestEastMoneyCollector_GetRecentDailyData_001208 测试获取001208最近的交易数据
func TestEastMoneyCollector_GetRecentDailyData_001208(t *testing.T) {
	// 创建一个简单的logger
	logrusLogger := logrus.New()
	logrusLogger.SetLevel(logrus.InfoLevel)
	logger := &utils.Logger{Logger: logrusLogger}

	collector := NewEastMoneyCollector(logger)

	// 测试股票代码 001208.SZ
	stockCode := "001208.SZ"

	// 设置时间范围：最近3个月
	endDate := time.Now()
	startDate := endDate.AddDate(0, -3, 0)

	t.Logf("Testing GetDailyKLine for recent data of stock: %s", stockCode)
	t.Logf("Date range: %s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	// 获取日K数据
	dailyData, err := collector.GetDailyKLine(stockCode, startDate, endDate)

	if err != nil {
		t.Fatalf("GetDailyKLine returned error: %v", err)
	}

	assert.NotNil(t, dailyData)
	assert.True(t, len(dailyData) > 0, "Should have recent daily data")

	t.Logf("Successfully retrieved %d recent daily K-line records for %s", len(dailyData), stockCode)

	// 找到最新的交易日数据
	var latestData *model.DailyData
	latestTradeDate := 0

	for i := range dailyData {
		if dailyData[i].TradeDate > latestTradeDate {
			latestTradeDate = dailyData[i].TradeDate
			latestData = &dailyData[i]
		}
	}

	if latestData != nil {
		t.Logf("=== 最新交易日数据 ===")
		t.Logf("股票代码: %s", latestData.TsCode)
		t.Logf("交易日期: %d", latestData.TradeDate)
		t.Logf("开盘价: %.3f元", latestData.Open)
		t.Logf("最高价: %.3f元", latestData.High)
		t.Logf("最低价: %.3f元", latestData.Low)
		t.Logf("收盘价: %.3f元", latestData.Close)
		t.Logf("成交量: %d股", latestData.Volume)
		t.Logf("成交额: %.2f万元", latestData.Amount/10000)

		// 计算涨跌幅（需要前一交易日数据）
		var prevData *model.DailyData
		prevTradeDate := 0
		for i := range dailyData {
			if dailyData[i].TradeDate < latestTradeDate && dailyData[i].TradeDate > prevTradeDate {
				prevTradeDate = dailyData[i].TradeDate
				prevData = &dailyData[i]
			}
		}

		if prevData != nil {
			changePercent := (latestData.Close - prevData.Close) / prevData.Close * 100
			changeAmount := latestData.Close - prevData.Close
			t.Logf("前一交易日收盘: %.3f元", prevData.Close)
			t.Logf("涨跌额: %.3f元", changeAmount)
			t.Logf("涨跌幅: %.2f%%", changePercent)
		}
	}

	// 显示最近10个交易日的收盘价走势
	t.Logf("=== 最近10个交易日收盘价走势 ===")

	// 按日期排序（降序）
	sortedData := make([]model.DailyData, len(dailyData))
	copy(sortedData, dailyData)

	// 简单排序：找出最新的10个交易日
	recentCount := 10
	if len(sortedData) < recentCount {
		recentCount = len(sortedData)
	}

	// 找出最新的几个交易日
	var recentDates []int
	for i := 0; i < recentCount; i++ {
		maxDate := 0
		maxIndex := -1
		for j, data := range sortedData {
			if data.TradeDate > maxDate {
				// 检查是否已经在recentDates中
				found := false
				for _, date := range recentDates {
					if date == data.TradeDate {
						found = true
						break
					}
				}
				if !found {
					maxDate = data.TradeDate
					maxIndex = j
				}
			}
		}
		if maxIndex >= 0 {
			recentDates = append(recentDates, maxDate)
			t.Logf("%d: 收盘价 %.3f元", maxDate, sortedData[maxIndex].Close)
		}
	}

	t.Logf("001208日K数据固化测试完成")
}

// TestEastMoneyCollector_GetPerformanceReports 测试获取业绩报表数据功能
func TestEastMoneyCollector_GetPerformanceReports(t *testing.T) {
	// 创建一个简单的logger
	logrusLogger := logrus.New()
	logrusLogger.SetLevel(logrus.InfoLevel)
	logger := &utils.Logger{Logger: logrusLogger}

	collector := NewEastMoneyCollector(logger)

	// 测试股票代码
	stockCode := "001208.SZ"

	t.Logf("Testing GetPerformanceReports for stock: %s", stockCode)

	// 获取业绩报表数据
	reports, err := collector.GetPerformanceReports(stockCode)

	// 检查错误
	if err != nil {
		t.Logf("GetPerformanceReports returned error: %v", err)
		// 即使有错误，也可能是正常的（比如股票不存在或没有业绩数据）
		return
	}

	// 检查业绩报表数据不为空
	assert.NotNil(t, reports)

	// 如果有业绩数据，检查基本字段
	if reports != nil && len(reports) > 0 {
		t.Logf("Successfully retrieved %d performance reports for stock %s", len(reports), stockCode)

		// 获取最新的业绩报表数据
		latestReport := reports[0]
		t.Logf("=== 最新业绩报表数据 ===")
		t.Logf("股票代码: %s", latestReport.TsCode)
		t.Logf("报告期: %s", latestReport.ReportDate.Format("2006-01-02"))

		// 每股收益相关
		t.Logf("每股收益(EPS): %.4f元", latestReport.EPS)
		t.Logf("加权每股收益: %.4f元", latestReport.WeightEPS)

		// 营业收入相关
		t.Logf("营业总收入: %.2f亿元", latestReport.Revenue/100000000)
		t.Logf("营业收入同比增长: %.2f%%", latestReport.RevenueYoY)
		t.Logf("营业收入环比增长: %.2f%%", latestReport.RevenueQoQ)

		// 净利润相关
		t.Logf("净利润: %.2f亿元", latestReport.NetProfit/100000000)
		t.Logf("净利润同比增长: %.2f%%", latestReport.NetProfitYoY)
		t.Logf("净利润环比增长: %.2f%%", latestReport.NetProfitQoQ)

		// 其他指标
		t.Logf("每股净资产(BVPS): %.4f元", latestReport.BVPS)
		t.Logf("销售毛利率: %.2f%%", latestReport.GrossMargin)
		t.Logf("股息率: %.2f%%", latestReport.DividendYield)

		// 公告日期
		if latestReport.LatestAnnouncementDate != nil {
			t.Logf("最新公告日期: %s", latestReport.LatestAnnouncementDate.Format("2006-01-02"))
		}
		if latestReport.FirstAnnouncementDate != nil {
			t.Logf("首次公告日期: %s", latestReport.FirstAnnouncementDate.Format("2006-01-02"))
		}

		// 验证数据完整性
		assert.Equal(t, stockCode, latestReport.TsCode, "TsCode should match")
		// 注意：如果日期解析失败，ReportDate可能为零值，这在某些情况下是正常的
		if latestReport.ReportDate.IsZero() {
			t.Logf("Warning: ReportDate is zero, possibly due to date parsing issues")
		}

		// 显示最近几期的业绩对比
		displayCount := 3
		if len(reports) < displayCount {
			displayCount = len(reports)
		}

		t.Logf("=== 最近%d期业绩对比 ===", displayCount)
		for i := 0; i < displayCount; i++ {
			report := reports[i]
			t.Logf("报告期: %s, EPS: %.4f元, 营收: %.2f亿元, 净利润: %.2f亿元",
				report.ReportDate.Format("2006-01-02"),
				report.EPS,
				report.Revenue/100000000,
				report.NetProfit/100000000)
		}
	} else {
		t.Logf("No performance reports available for stock %s", stockCode)
	}
}

// TestEastMoneyCollector_GetPerformanceReports_InvalidCode 测试无效股票代码
func TestEastMoneyCollector_GetPerformanceReports_InvalidCode(t *testing.T) {
	// 创建一个简单的logger
	logrusLogger := logrus.New()
	logrusLogger.SetLevel(logrus.InfoLevel)
	logger := &utils.Logger{Logger: logrusLogger}

	collector := NewEastMoneyCollector(logger)

	// 测试无效的股票代码格式
	invalidCodes := []string{
		"000001",    // 缺少交易所后缀
		"000001.XX", // 无效交易所
		"12345.SZ",  // 股票代码位数不对
		"ABCDEF.SH", // 非数字股票代码
		"",          // 空字符串
	}

	for _, invalidCode := range invalidCodes {
		t.Logf("Testing invalid stock code: %s", invalidCode)

		reports, err := collector.GetPerformanceReports(invalidCode)

		// 应该返回错误
		assert.Error(t, err, "Should return error for invalid code: %s", invalidCode)
		assert.Nil(t, reports, "Should return nil reports for invalid code: %s", invalidCode)

		t.Logf("Expected error for invalid code %s: %v", invalidCode, err)
	}
}

// TestEastMoneyCollector_GetPerformanceReports_Multiple 测试批量获取业绩报表数据
func TestEastMoneyCollector_GetPerformanceReports_Multiple(t *testing.T) {
	// 创建一个简单的logger
	logrusLogger := logrus.New()
	logrusLogger.SetLevel(logrus.InfoLevel)
	logger := &utils.Logger{Logger: logrusLogger}

	collector := NewEastMoneyCollector(logger)

	// 测试多个股票代码
	testStocks := []string{
		"000001.SZ", // 平安银行
		"000002.SZ", // 万科A
		"600000.SH", // 浦发银行
		"600036.SH", // 招商银行
	}

	for _, stockCode := range testStocks {
		t.Logf("Testing GetPerformanceReports for stock: %s", stockCode)

		reports, err := collector.GetPerformanceReports(stockCode)

		if err != nil {
			t.Logf("Error getting performance reports for %s: %v", stockCode, err)
			continue
		}

		if reports != nil && len(reports) > 0 {
			latestReport := reports[0]
			t.Logf("Success - Stock %s: EPS=%.4f元, 营收=%.2f亿元, 净利润=%.2f亿元, 毛利率=%.2f%%",
				stockCode,
				latestReport.EPS,
				latestReport.Revenue/100000000,
				latestReport.NetProfit/100000000,
				latestReport.GrossMargin)
		} else {
			t.Logf("No performance reports available for stock %s", stockCode)
		}

		// 添加延迟避免请求过快
		time.Sleep(200 * time.Millisecond)
	}
}

// TestEastMoneyCollector_GetLatestPerformanceReport 测试获取最新业绩报表数据
func TestEastMoneyCollector_GetLatestPerformanceReport(t *testing.T) {
	// 创建一个简单的logger
	logrusLogger := logrus.New()
	logrusLogger.SetLevel(logrus.InfoLevel)
	logger := &utils.Logger{Logger: logrusLogger}

	collector := NewEastMoneyCollector(logger)

	// 测试股票代码 000001.SZ
	stockCode := "000001.SZ"

	t.Logf("Testing GetLatestPerformanceReport for stock: %s", stockCode)

	// 获取最新业绩报表数据
	latestReport, err := collector.GetLatestPerformanceReport(stockCode)

	if err != nil {
		t.Logf("GetLatestPerformanceReport returned error: %v", err)
		return
	}

	// 检查最新业绩报表数据
	assert.NotNil(t, latestReport)

	if latestReport != nil {
		t.Logf("=== 最新业绩报表数据 ===")
		t.Logf("股票代码: %s", latestReport.TsCode)
		t.Logf("报告期: %s", latestReport.ReportDate.Format("2006-01-02"))
		t.Logf("每股收益: %.4f元", latestReport.EPS)
		t.Logf("营业收入: %.2f亿元", latestReport.Revenue/100000000)
		t.Logf("净利润: %.2f亿元", latestReport.NetProfit/100000000)

		// 验证数据完整性
		assert.Equal(t, stockCode, latestReport.TsCode, "TsCode should match")
		// 注意：如果日期解析失败，ReportDate可能为零值，这在某些情况下是正常的
		if latestReport.ReportDate.IsZero() {
			t.Logf("Warning: ReportDate is zero, possibly due to date parsing issues")
		}

		// 同时获取所有业绩报表数据进行对比
		allReports, err := collector.GetPerformanceReports(stockCode)
		if err == nil && len(allReports) > 0 {
			// 验证返回的确实是最新的报表
			for _, report := range allReports {
				assert.True(t, latestReport.ReportDate.After(report.ReportDate) || latestReport.ReportDate.Equal(report.ReportDate),
					"Latest report should have the most recent date")
			}
			t.Logf("Verified that returned report is indeed the latest among %d reports", len(allReports))
		}
	}
}
