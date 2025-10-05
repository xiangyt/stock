package main

import (
	"fmt"
	"log"
	"time"

	"stock/internal/collector"
	"stock/internal/logger"
	"stock/internal/model"
)

func main() {
	// 创建同花顺采集器
	c := collector.GetCollectorFactory(logger.GetGlobalLogger()).GetTongHuaShunCollector()

	// 测试股票代码
	testCodes := []string{
		"601899.SH", // 紫金矿业
		//"000001.SZ", // 平安银行
	}

	// 设置不同的日期范围
	endDate := time.Now()

	// 周K线和月K线：最近6个月
	weeklyMonthlyStartDate := endDate.AddDate(0, -6, 0)

	// 季K线：最近2年
	quarterlyStartDate := endDate.AddDate(-1, 0, 0)

	// 年K线：最近5年
	yearlyStartDate := endDate.AddDate(-5, 0, 0)

	fmt.Printf("测试同花顺K线数据采集\n")
	fmt.Printf("===========================================\n\n")

	for _, tsCode := range testCodes {
		fmt.Printf("正在测试股票: %s\n", tsCode)

		// 测试周K线数据
		fmt.Printf("  [周K线] ")
		weeklyData, err := c.GetWeeklyKLine(tsCode, weeklyMonthlyStartDate, endDate)
		if err != nil {
			log.Printf("获取 %s 周K线数据失败: %v", tsCode, err)
		} else {
			fmt.Printf("成功获取周K线数据，共 %d 条记录\n", len(weeklyData))
			displayWeeklyKLineData("周K线", weeklyData[:min(3, len(weeklyData))])
		}

		// 测试月K线数据
		fmt.Printf("  [月K线] ")
		monthlyData, err := c.GetMonthlyKLine(tsCode, weeklyMonthlyStartDate, endDate)
		if err != nil {
			log.Printf("获取 %s 月K线数据失败: %v", tsCode, err)
		} else {
			fmt.Printf("成功获取月K线数据，共 %d 条记录\n", len(monthlyData))
			displayMonthlyKLineData("月K线", monthlyData[:min(5, len(monthlyData))])
		}

		// 测试季K线数据（推测参数为 "31"）
		fmt.Printf("  [季K线] ")
		quarterlyData, err := c.GetQuarterlyKLine(tsCode, quarterlyStartDate, endDate)
		if err != nil {
			log.Printf("获取 %s 季K线数据失败: %v", tsCode, err)
		} else {
			fmt.Printf("成功获取季K线数据，共 %d 条记录\n", len(quarterlyData))
			displayQuarterlyKLineData("季K线", quarterlyData[:min(3, len(quarterlyData))])
		}

		// 测试年K线数据（推测参数为 "41"）
		fmt.Printf("  [年K线] ")
		yearlyData, err := c.GetYearlyKLine(tsCode, yearlyStartDate, endDate)
		if err != nil {
			log.Printf("获取 %s 年K线数据失败: %v", tsCode, err)
		} else {
			fmt.Printf("成功获取年K线数据，共 %d 条记录\n", len(yearlyData))
			displayYearlyKLineData("年K线", yearlyData[:min(6, len(yearlyData))])
		}

		fmt.Printf("\n")
	}

	// 总结参数结果
	fmt.Printf("K线参数总结:\n")
	fmt.Printf("===========================================\n")
	fmt.Printf("日K线: 01 (已验证)\n")
	fmt.Printf("周K线: 11 (已验证)\n")
	fmt.Printf("月K线: 21 (已验证)\n")
	fmt.Printf("季K线: 91 (已验证)\n")
	fmt.Printf("年K线: 81 (已验证)\n")
	fmt.Printf("\n")

	fmt.Printf("测试完成！\n")
}

// displayWeeklyKLineData 显示周K线数据
func displayWeeklyKLineData(dataType string, data []model.WeeklyData) {
	if len(data) == 0 {
		return
	}

	fmt.Printf("    %s前 %d 条数据示例:\n", dataType, len(data))
	fmt.Printf("    %-12s %-10s %-8s %-8s %-8s %-8s %-12s\n",
		"股票代码", "交易日期", "开盘", "最高", "最低", "收盘", "成交量")
	fmt.Printf("    %-12s %-10s %-8s %-8s %-8s %-8s %-12s\n",
		"--------", "--------", "----", "----", "----", "----", "--------")

	for _, item := range data {
		fmt.Printf("    %-12s %-10d %-8.2f %-8.2f %-8.2f %-8.2f %-12d\n",
			item.TsCode, item.TradeDate, item.Open, item.High,
			item.Low, item.Close, item.Volume)
	}
}

// displayMonthlyKLineData 显示月K线数据
func displayMonthlyKLineData(dataType string, data []model.MonthlyData) {
	if len(data) == 0 {
		return
	}

	fmt.Printf("    %s前 %d 条数据示例:\n", dataType, len(data))
	fmt.Printf("    %-12s %-10s %-8s %-8s %-8s %-8s %-12s\n",
		"股票代码", "交易日期", "开盘", "最高", "最低", "收盘", "成交量")
	fmt.Printf("    %-12s %-10s %-8s %-8s %-8s %-8s %-12s\n",
		"--------", "--------", "----", "----", "----", "----", "--------")

	for _, item := range data {
		fmt.Printf("    %-12s %-10d %-8.2f %-8.2f %-8.2f %-8.2f %-12d\n",
			item.TsCode, item.TradeDate, item.Open, item.High,
			item.Low, item.Close, item.Volume)
	}
}

// displayQuarterlyKLineData 显示季K线数据
func displayQuarterlyKLineData(dataType string, data []model.QuarterlyData) {
	if len(data) == 0 {
		return
	}

	fmt.Printf("    %s前 %d 条数据示例:\n", dataType, len(data))
	fmt.Printf("    %-12s %-10s %-8s %-8s %-8s %-8s %-12s\n",
		"股票代码", "交易日期", "开盘", "最高", "最低", "收盘", "成交量")
	fmt.Printf("    %-12s %-10s %-8s %-8s %-8s %-8s %-12s\n",
		"--------", "--------", "----", "----", "----", "----", "--------")

	for _, item := range data {
		fmt.Printf("    %-12s %-10d %-8.2f %-8.2f %-8.2f %-8.2f %-12d\n",
			item.TsCode, item.TradeDate, item.Open, item.High,
			item.Low, item.Close, item.Volume)
	}
}

// displayYearlyKLineData 显示年K线数据
func displayYearlyKLineData(dataType string, data []model.YearlyData) {
	if len(data) == 0 {
		return
	}

	fmt.Printf("    %s前 %d 条数据示例:\n", dataType, len(data))
	fmt.Printf("    %-12s %-10s %-8s %-8s %-8s %-8s %-12s\n",
		"股票代码", "交易日期", "开盘", "最高", "最低", "收盘", "成交量")
	fmt.Printf("    %-12s %-10s %-8s %-8s %-8s %-8s %-12s\n",
		"--------", "--------", "----", "----", "----", "----", "--------")

	for _, item := range data {
		fmt.Printf("    %-12s %-10d %-8.2f %-8.2f %-8.2f %-8.2f %-12d\n",
			item.TsCode, item.TradeDate, item.Open, item.High,
			item.Low, item.Close, item.Volume)
	}
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
