package main

import (
	"fmt"
	"stock/internal/collector"
	"stock/internal/logger"
	"strings"
)

func main() {
	// 初始化日志
	log := logger.NewLogger(logger.LogConfig{
		Level:  "info",
		Format: "text",
	})

	// 创建采集器工厂
	factory := collector.GetCollectorFactory(log)

	// 获取同花顺采集器
	ths := factory.GetTongHuaShunCollector()

	// 连接数据源
	if err := ths.Connect(); err != nil {
		log.Fatalf("Failed to connect to TongHuaShun: %v", err)
	}
	defer ths.Disconnect()

	// 测试股票代码列表
	testStocks := []string{
		"601899.SH", // 紫金矿业
		"000001.SZ", // 平安银行
		//"600036.SH", // 招商银行
		//"000002.SZ", // 万科A
		//"300750.SZ", // 宁德时代
	}

	fmt.Println("=== 同花顺多时间周期数据采集测试 ===")
	fmt.Println()

	for i, tsCode := range testStocks {
		fmt.Printf("%d. 测试股票: %s\n", i+1, tsCode)
		fmt.Println("   " + strings.Repeat("=", 50))

		// 1. 获取当日数据
		fmt.Printf("   📅 当日数据:\n")
		todayData, name, err := ths.GetTodayData(tsCode)
		if err != nil {
			fmt.Printf("      ❌ 获取失败: %v\n", err)
		} else {
			fmt.Printf("      ✅ 获取成功 - %s\n", name)
			fmt.Printf("         交易日期: %d\n", todayData.TradeDate)
			fmt.Printf("         开盘价: %.2f, 最高价: %.2f, 最低价: %.2f, 收盘价: %.2f\n",
				todayData.Open, todayData.High, todayData.Low, todayData.Close)
			fmt.Printf("         成交量: %d, 成交额: %.2f\n", todayData.Volume, todayData.Amount)
		}
		fmt.Println()

		// 2. 获取本周数据
		fmt.Printf("   📊 本周数据:\n")
		weekData, err := ths.GetThisWeekData(tsCode)
		if err != nil {
			fmt.Printf("      ❌ 获取失败: %v\n", err)
		} else {
			fmt.Printf("      ✅ 获取成功\n")
			fmt.Printf("         交易日期: %d\n", weekData.TradeDate)
			fmt.Printf("         开盘价: %.2f, 最高价: %.2f, 最低价: %.2f, 收盘价: %.2f\n",
				weekData.Open, weekData.High, weekData.Low, weekData.Close)
			fmt.Printf("         成交量: %d, 成交额: %.2f\n", weekData.Volume, weekData.Amount)
		}
		fmt.Println()

		// 3. 获取本月数据
		fmt.Printf("   📈 本月数据:\n")
		monthData, err := ths.GetThisMonthData(tsCode)
		if err != nil {
			fmt.Printf("      ❌ 获取失败: %v\n", err)
		} else {
			fmt.Printf("      ✅ 获取成功\n")
			fmt.Printf("         交易日期: %d\n", monthData.TradeDate)
			fmt.Printf("         开盘价: %.2f, 最高价: %.2f, 最低价: %.2f, 收盘价: %.2f\n",
				monthData.Open, monthData.High, monthData.Low, monthData.Close)
			fmt.Printf("         成交量: %d, 成交额: %.2f\n", monthData.Volume, monthData.Amount)
		}
		fmt.Println()

		// 4. 获取本季数据
		fmt.Printf("   📊 本季数据:\n")
		quarterData, err := ths.GetThisQuarterData(tsCode)
		if err != nil {
			fmt.Printf("      ❌ 获取失败: %v\n", err)
		} else {
			fmt.Printf("      ✅ 获取成功\n")
			fmt.Printf("         交易日期: %d\n", quarterData.TradeDate)
			fmt.Printf("         开盘价: %.2f, 最高价: %.2f, 最低价: %.2f, 收盘价: %.2f\n",
				quarterData.Open, quarterData.High, quarterData.Low, quarterData.Close)
			fmt.Printf("         成交量: %d, 成交额: %.2f\n", quarterData.Volume, quarterData.Amount)
		}
		fmt.Println()

		// 5. 获取本年数据
		fmt.Printf("   📈 本年数据:\n")
		yearData, err := ths.GetThisYearData(tsCode)
		if err != nil {
			fmt.Printf("      ❌ 获取失败: %v\n", err)
		} else {
			fmt.Printf("      ✅ 获取成功\n")
			fmt.Printf("         交易日期: %d\n", yearData.TradeDate)
			fmt.Printf("         开盘价: %.2f, 最高价: %.2f, 最低价: %.2f, 收盘价: %.2f\n",
				yearData.Open, yearData.High, yearData.Low, yearData.Close)
			fmt.Printf("         成交量: %d, 成交额: %.2f\n", yearData.Volume, yearData.Amount)
		}

		fmt.Println()
		fmt.Println("   " + strings.Repeat("=", 50))
		fmt.Println()
	}

	fmt.Println("=== 测试完成 ===")
}
