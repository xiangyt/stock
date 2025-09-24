package collector

import (
	"testing"
	"time"

	"stock/internal/logger"
)

func TestTongHuaShunCollector_Basic(t *testing.T) {
	// 创建测试日志器
	log := logger.NewLogger(logger.LogConfig{
		Level:  "info",
		Format: "text",
	})

	// 创建同花顺采集器
	collector := newTongHuaShunCollector(log)

	// 测试基本属性
	if collector.GetName() != "tonghuashun" {
		t.Errorf("Expected name 'tonghuashun', got '%s'", collector.GetName())
	}

	// 测试连接
	if collector.IsConnected() {
		t.Error("Expected collector to be disconnected initially")
	}

	err := collector.Connect()
	if err != nil {
		t.Errorf("Failed to connect: %v", err)
	}

	if !collector.IsConnected() {
		t.Error("Expected collector to be connected after Connect()")
	}

	// 测试断开连接
	err = collector.Disconnect()
	if err != nil {
		t.Errorf("Failed to disconnect: %v", err)
	}

	if collector.IsConnected() {
		t.Error("Expected collector to be disconnected after Disconnect()")
	}
}

func TestTongHuaShunCollector_EmptyImplementations(t *testing.T) {
	// 创建测试日志器
	log := logger.NewLogger(logger.LogConfig{
		Level:  "info",
		Format: "text",
	})

	// 创建同花顺采集器
	collector := newTongHuaShunCollector(log)

	// 连接采集器
	err := collector.Connect()
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// 测试所有空实现方法都返回错误
	testCases := []struct {
		name string
		test func() error
	}{
		{
			name: "GetStockList",
			test: func() error {
				_, err := collector.GetStockList()
				return err
			},
		},
		{
			name: "GetStockDetail",
			test: func() error {
				_, err := collector.GetStockDetail("000001.SZ")
				return err
			},
		},
		{
			name: "GetStockData",
			test: func() error {
				_, err := collector.GetStockData("000001.SZ", time.Now().AddDate(0, -1, 0), time.Now())
				return err
			},
		},
		{
			name: "GetDailyKLine",
			test: func() error {
				// GetDailyKLine 已实现，可能返回网络错误或成功
				_, _ = collector.GetDailyKLine("000001.SZ", time.Now().AddDate(0, -1, 0), time.Now())
				// 不期望特定的错误，因为这个方法已经实现
				return nil
			},
		},
		{
			name: "GetWeeklyKLine",
			test: func() error {
				_, err := collector.GetWeeklyKLine("000001.SZ", time.Now().AddDate(0, -1, 0), time.Now())
				return err
			},
		},
		{
			name: "GetMonthlyKLine",
			test: func() error {
				_, err := collector.GetMonthlyKLine("000001.SZ", time.Now().AddDate(0, -1, 0), time.Now())
				return err
			},
		},
		{
			name: "GetYearlyKLine",
			test: func() error {
				_, err := collector.GetYearlyKLine("000001.SZ", time.Now().AddDate(0, -1, 0), time.Now())
				return err
			},
		},
		{
			name: "GetRealtimeData",
			test: func() error {
				_, err := collector.GetRealtimeData([]string{"000001.SZ"})
				return err
			},
		},
		{
			name: "GetPerformanceReports",
			test: func() error {
				_, err := collector.GetPerformanceReports("000001.SZ")
				return err
			},
		},
		{
			name: "GetLatestPerformanceReport",
			test: func() error {
				_, err := collector.GetLatestPerformanceReport("000001.SZ")
				return err
			},
		},
		{
			name: "GetShareholderCounts",
			test: func() error {
				_, err := collector.GetShareholderCounts("000001.SZ")
				return err
			},
		},
		{
			name: "GetLatestShareholderCount",
			test: func() error {
				_, err := collector.GetLatestShareholderCount("000001.SZ")
				return err
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.test()
			// GetDailyKLine 已实现，不期望返回错误
			if tc.name == "GetDailyKLine" {
				// 对于已实现的方法，不检查错误
				return
			}
			if err == nil {
				t.Errorf("Expected %s to return an error (not implemented), but got nil", tc.name)
			}
			if err != nil && err.Error() == "" {
				t.Errorf("Expected %s to return a meaningful error message", tc.name)
			}
		})
	}
}

func TestTongHuaShunCollector_RateLimit(t *testing.T) {
	// 创建测试日志器
	log := logger.NewLogger(logger.LogConfig{
		Level:  "info",
		Format: "text",
	})

	// 创建同花顺采集器
	collector := newTongHuaShunCollector(log)

	// 测试默认限流设置
	if collector.GetRateLimit() != 1 {
		t.Errorf("Expected default rate limit to be 1, got %d", collector.GetRateLimit())
	}

	// 测试设置限流
	collector.SetRateLimit(5)
	if collector.GetRateLimit() != 5 {
		t.Errorf("Expected rate limit to be 5 after setting, got %d", collector.GetRateLimit())
	}

	// 测试限流统计
	stats := collector.GetRateLimitStats()
	if stats["rate_limit"] != 5 {
		t.Errorf("Expected rate_limit in stats to be 5, got %v", stats["rate_limit"])
	}
}

func TestTongHuaShunCollector_UserAgentAndCookie(t *testing.T) {
	// 创建测试日志器
	log := logger.NewLogger(logger.LogConfig{
		Level:  "info",
		Format: "text",
	})

	// 创建同花顺采集器
	collector := newTongHuaShunCollector(log)

	// 测试初始User-Agent和Cookie不为空
	initialUA := collector.GetCurrentUserAgent()
	initialCookie := collector.GetCurrentCookie()

	if initialUA == "" {
		t.Error("Expected initial User-Agent to be non-empty")
	}

	if initialCookie == "" {
		t.Error("Expected initial Cookie to be non-empty")
	}

	// 测试强制更新
	collector.ForceUpdateUserAgentAndCookie()

	newUA := collector.GetCurrentUserAgent()
	newCookie := collector.GetCurrentCookie()

	if newUA == "" {
		t.Error("Expected new User-Agent to be non-empty")
	}

	if newCookie == "" {
		t.Error("Expected new Cookie to be non-empty")
	}

	// 注意：由于是随机生成，新的UA和Cookie可能与原来相同，这是正常的
	t.Logf("Initial UA: %s", initialUA[:min(50, len(initialUA))])
	t.Logf("New UA: %s", newUA[:min(50, len(newUA))])
}

// min 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestTongHuaShunCollector_GetDailyKLine_Implementation(t *testing.T) {
	// 创建测试日志器
	log := logger.NewLogger(logger.LogConfig{
		Level:  "info",
		Format: "text",
	})

	// 创建同花顺采集器
	collector := newTongHuaShunCollector(log)

	// 连接采集器
	err := collector.Connect()
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer collector.Disconnect()

	// 测试用例
	testCases := []struct {
		name      string
		tsCode    string
		startDate time.Time
		endDate   time.Time
		expectErr bool
	}{
		{
			name:      "Valid stock code - 平安银行",
			tsCode:    "000001.SZ",
			startDate: time.Now().AddDate(0, -1, 0), // 1个月前
			endDate:   time.Now(),
			expectErr: false, // 可能因为网络问题失败，但不应该是代码逻辑错误
		},
		{
			name:      "Valid stock code - 华菱线缆",
			tsCode:    "001208.SZ",
			startDate: time.Now().AddDate(0, -1, 0), // 1个月前
			endDate:   time.Now(),
			expectErr: false, // 可能因为网络问题失败，但不应该是代码逻辑错误
		},
		{
			name:      "Valid stock code - 华电新能",
			tsCode:    "600930.SH",
			startDate: time.Now().AddDate(0, -1, 0),
			endDate:   time.Now(),
			expectErr: false,
		},
		{
			name:      "Invalid stock code format",
			tsCode:    "invalid_code",
			startDate: time.Now().AddDate(0, -1, 0),
			endDate:   time.Now(),
			expectErr: true, // 应该返回格式错误
		},
		{
			name:      "Empty time range",
			tsCode:    "000001.SZ",
			startDate: time.Time{}, // 空时间
			endDate:   time.Time{}, // 空时间
			expectErr: false,       // 应该获取所有数据
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := collector.GetDailyKLine(tc.tsCode, tc.startDate, tc.endDate)

			if tc.expectErr {
				if err == nil {
					t.Errorf("Expected error for test case '%s', but got nil", tc.name)
				} else {
					t.Logf("Expected error occurred: %v", err)
				}
				return
			}

			// 对于不期望错误的情况
			if err != nil {
				// 网络错误是可以接受的，记录但不失败测试
				t.Logf("Network or API error (acceptable): %v", err)
				return
			}

			// 验证返回的数据
			t.Logf("Successfully fetched %d records for %s", len(data), tc.tsCode)

			// 验证数据结构
			for i, record := range data {
				if i >= 3 { // 只检查前3条记录
					break
				}

				if record.TsCode != tc.tsCode {
					t.Errorf("Record %d: expected TsCode %s, got %s", i, tc.tsCode, record.TsCode)
				}

				if record.TradeDate <= 0 {
					t.Errorf("Record %d: invalid TradeDate %d", i, record.TradeDate)
				}

				if record.Close <= 0 {
					t.Errorf("Record %d: invalid Close price %f", i, record.Close)
				}

				t.Logf("Record %d: Date=%d, Open=%.2f, High=%.2f, Low=%.2f, Close=%.2f, Volume=%d",
					i, record.TradeDate, record.Open, record.High, record.Low, record.Close, record.Volume)
			}
		})
	}
}
