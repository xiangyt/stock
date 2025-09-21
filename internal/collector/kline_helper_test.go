package collector

import (
	"testing"
	"time"

	"stock/internal/model"
)

// TestKLineParser_ParseToDaily 测试日K线解析功能
func TestKLineParser_ParseToDaily(t *testing.T) {
	parser := NewKLineParser()

	tests := []struct {
		name     string
		tsCode   string
		kline    string
		expected *model.DailyData
		hasError bool
	}{
		{
			name:   "正常日K线数据解析",
			tsCode: "000001.SZ",
			kline:  "2025-09-20,10.50,10.80,11.00,10.30,1000000,10500000.00",
			expected: &model.DailyData{
				TsCode:    "000001.SZ",
				TradeDate: 20250920,
				Open:      10.50,
				High:      11.00,
				Low:       10.30,
				Close:     10.80,
				Volume:    1000000,
				Amount:    10500000.00,
			},
			hasError: false,
		},
		{
			name:   "上海股票日K线数据解析",
			tsCode: "600000.SH",
			kline:  "2025-09-21,15.20,15.45,15.60,15.10,2000000,30900000.00",
			expected: &model.DailyData{
				TsCode:    "600000.SH",
				TradeDate: 20250921,
				Open:      15.20,
				High:      15.60,
				Low:       15.10,
				Close:     15.45,
				Volume:    2000000,
				Amount:    30900000.00,
			},
			hasError: false,
		},
		{
			name:   "包含空值的日K线数据",
			tsCode: "000002.SZ",
			kline:  "2025-09-19,8.50,-,8.80,8.40,500000,4250000.00",
			expected: &model.DailyData{
				TsCode:    "000002.SZ",
				TradeDate: 20250919,
				Open:      8.50,
				High:      8.80,
				Low:       8.40,
				Close:     0, // 空值应该解析为0
				Volume:    500000,
				Amount:    4250000.00,
			},
			hasError: false,
		},
		{
			name:     "无效的K线数据格式 - 字段不足",
			tsCode:   "000003.SZ",
			kline:    "2025-09-18,12.50,12.80",
			expected: nil,
			hasError: true,
		},
		{
			name:     "无效的日期格式",
			tsCode:   "000004.SZ",
			kline:    "invalid-date,12.50,12.80,13.00,12.30,800000,10240000.00",
			expected: nil,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseToDaily(tt.tsCode, tt.kline)

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Errorf("Expected result but got nil")
				return
			}

			// 验证基本字段
			if result.TsCode != tt.expected.TsCode {
				t.Errorf("TsCode: expected %s, got %s", tt.expected.TsCode, result.TsCode)
			}

			if result.TradeDate != tt.expected.TradeDate {
				t.Errorf("TradeDate: expected %d, got %d", tt.expected.TradeDate, result.TradeDate)
			}

			if result.Open != tt.expected.Open {
				t.Errorf("Open: expected %.2f, got %.2f", tt.expected.Open, result.Open)
			}

			if result.High != tt.expected.High {
				t.Errorf("High: expected %.2f, got %.2f", tt.expected.High, result.High)
			}

			if result.Low != tt.expected.Low {
				t.Errorf("Low: expected %.2f, got %.2f", tt.expected.Low, result.Low)
			}

			if result.Close != tt.expected.Close {
				t.Errorf("Close: expected %.2f, got %.2f", tt.expected.Close, result.Close)
			}

			if result.Volume != tt.expected.Volume {
				t.Errorf("Volume: expected %d, got %d", tt.expected.Volume, result.Volume)
			}

			if result.Amount != tt.expected.Amount {
				t.Errorf("Amount: expected %.2f, got %.2f", tt.expected.Amount, result.Amount)
			}

			// 验证时间戳字段存在且合理
			if result.CreatedAt <= 0 {
				t.Errorf("CreatedAt should be positive, got %d", result.CreatedAt)
			}

			if result.UpdatedAt <= 0 {
				t.Errorf("UpdatedAt should be positive, got %d", result.UpdatedAt)
			}

			// 验证时间戳是最近的时间
			now := time.Now().Unix()
			if result.CreatedAt > now || result.CreatedAt < now-10 {
				t.Errorf("CreatedAt should be recent, got %d, now is %d", result.CreatedAt, now)
			}
		})
	}
}

// TestKLineParser_ParseToDaily_TableName 测试日K线数据的表名选择
func TestKLineParser_ParseToDaily_TableName(t *testing.T) {
	parser := NewKLineParser()

	tests := []struct {
		name          string
		tsCode        string
		expectedTable string
	}{
		{
			name:          "深圳股票应该使用 daily_data_sz 表",
			tsCode:        "000001.SZ",
			expectedTable: "daily_data_sz",
		},
		{
			name:          "上海股票应该使用 daily_data_sh 表",
			tsCode:        "600000.SH",
			expectedTable: "daily_data_sh",
		},
		{
			name:          "创业板股票应该使用 daily_data_sz 表",
			tsCode:        "300001.SZ",
			expectedTable: "daily_data_sz",
		},
	}

	kline := "2025-09-20,10.50,10.80,11.00,10.30,1000000,10500000.00"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseToDaily(tt.tsCode, kline)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			tableName := result.TableName()
			if tableName != tt.expectedTable {
				t.Errorf("Expected table name %s, got %s", tt.expectedTable, tableName)
			}
		})
	}
}

// BenchmarkKLineParser_ParseToDaily 基准测试日K线解析性能
func BenchmarkKLineParser_ParseToDaily(b *testing.B) {
	parser := NewKLineParser()
	tsCode := "000001.SZ"
	kline := "2025-09-20,10.50,10.80,11.00,10.30,1000000,10500000.00"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.ParseToDaily(tsCode, kline)
		if err != nil {
			b.Errorf("Unexpected error: %v", err)
		}
	}
}

// TestKLineParser_AllParsers 测试所有解析器的一致性
func TestKLineParser_AllParsers(t *testing.T) {
	parser := NewKLineParser()
	tsCode := "000001.SZ"
	kline := "2025-09-20,10.50,10.80,11.00,10.30,1000000,10500000.00"

	// 测试所有解析器都能正常工作
	daily, err := parser.ParseToDaily(tsCode, kline)
	if err != nil {
		t.Errorf("ParseToDaily failed: %v", err)
	}

	weekly, err := parser.ParseToWeekly(tsCode, kline)
	if err != nil {
		t.Errorf("ParseToWeekly failed: %v", err)
	}

	monthly, err := parser.ParseToMonthly(tsCode, kline)
	if err != nil {
		t.Errorf("ParseToMonthly failed: %v", err)
	}

	yearly, err := parser.ParseToYearly(tsCode, kline)
	if err != nil {
		t.Errorf("ParseToYearly failed: %v", err)
	}

	// 验证基本数据一致性
	if daily.TsCode != weekly.TsCode || daily.TsCode != monthly.TsCode || daily.TsCode != yearly.TsCode {
		t.Errorf("TsCode inconsistency across parsers")
	}

	if daily.TradeDate != weekly.TradeDate || daily.TradeDate != monthly.TradeDate || daily.TradeDate != yearly.TradeDate {
		t.Errorf("TradeDate inconsistency across parsers")
	}

	t.Logf("All parsers working correctly for %s", tsCode)
}
