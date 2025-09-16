package service

import (
	"strings"
	"testing"

	"stock/internal/model"

	"github.com/stretchr/testify/assert"
)

// TestDailyData_TableName 测试DailyData的TableName方法
func TestDailyData_TableName(t *testing.T) {
	testCases := []struct {
		tsCode   string
		expected string
		desc     string
	}{
		{"000001.SZ", "daily_data_sz", "深圳股票代码"},
		{"600000.SH", "daily_data_sh", "上海股票代码"},
		{"000001", "daily_data_sz", "深圳股票代码（无后缀）"},
		{"600000", "daily_data_sh", "上海股票代码（无后缀）"},
		{"300001", "daily_data_sz", "创业板股票"},
		{"688001", "daily_data_sh", "科创板股票"},
		{"", "daily_data_sh", "空代码默认上海"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			data := model.DailyData{TsCode: tc.tsCode}
			tableName := data.TableName()
			assert.Equal(t, tc.expected, tableName, "股票代码 %s 应该使用表 %s", tc.tsCode, tc.expected)
		})
	}
}

// TestDailyData_GetExchange 测试交易所识别逻辑
func TestDailyData_GetExchange(t *testing.T) {
	testCases := []struct {
		tsCode   string
		expected string
		desc     string
	}{
		{"000001.SZ", "SZ", "深圳股票代码"},
		{"600000.SH", "SH", "上海股票代码"},
		{"000001", "SZ", "深圳股票代码（无后缀）"},
		{"600000", "SH", "上海股票代码（无后缀）"},
		{"300001", "SZ", "创业板股票"},
		{"688001", "SH", "科创板股票"},
		{"", "SH", "空代码默认上海"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			data := model.DailyData{TsCode: tc.tsCode}
			// 通过TableName间接测试交易所识别逻辑
			tableName := data.TableName()
			expectedTable := "daily_data_" + strings.ToLower(tc.expected)
			assert.Equal(t, expectedTable, tableName, "股票代码 %s 应该使用表 %s", tc.tsCode, expectedTable)
		})
	}
}
