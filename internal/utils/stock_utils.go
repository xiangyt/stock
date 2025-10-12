package utils

import (
	"fmt"
	"strings"
	"time"
)

// ConvertToTsCode 转换股票代码格式
func ConvertToTsCode(code string) string {
	// 如果已经是标准格式（包含.），直接返回
	if strings.Contains(code, ".") {
		return strings.ToUpper(code)
	}

	// 根据代码前缀判断交易所
	if strings.HasPrefix(code, "60") || strings.HasPrefix(code, "68") || strings.HasPrefix(code, "90") {
		return code + ".SH"
	} else if strings.HasPrefix(code, "00") || strings.HasPrefix(code, "30") || strings.HasPrefix(code, "20") {
		return code + ".SZ"
	}

	// 默认返回原代码
	return code
}

// ParseTradeDate 解析交易日期
func ParseTradeDate(date int) (time.Time, error) {
	tradeDateStr := fmt.Sprintf("%d", date)
	tradeDate, err := time.Parse("20060102", tradeDateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("解析交易日期失败: %v", err)
	}
	return tradeDate, nil
}
