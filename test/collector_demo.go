package main

import (
	"fmt"
	"log"
	"time"

	"stock/internal/collector"
	"stock/internal/config"
	"stock/internal/utils"
)

func main() {
	fmt.Println("=== 数据采集器功能测试 ===")

	// 创建日志记录器
	logConfig := config.LogConfig{
		Level:  "info",
		Format: "text",
	}
	logger := utils.NewLogger(logConfig)

	// 创建东方财富采集器
	eastMoneyCollector := collector.NewEastMoneyCollector(logger)

	// 连接数据源
	fmt.Println("\n1. 连接东方财富数据源...")
	if err := eastMoneyCollector.Connect(); err != nil {
		log.Fatalf("连接失败: %v", err)
	}
	fmt.Println("✅ 连接成功")

	// 测试获取股票列表
	fmt.Println("\n2. 获取股票列表（前10只）...")
	stocks, err := eastMoneyCollector.GetStockList()
	if err != nil {
		log.Fatalf("获取股票列表失败: %v", err)
	}

	fmt.Printf("✅ 成功获取 %d 只股票\n", len(stocks))
	fmt.Println("前10只股票:")
	for i, stock := range stocks {
		if i >= 10 {
			break
		}
		fmt.Printf("  %s - %s (%s)\n", stock.TsCode, stock.Name, stock.Area)
	}

	if len(stocks) == 0 {
		fmt.Println("❌ 没有获取到股票数据")
		return
	}

	// 测试获取股票详情
	testStock := stocks[0]
	fmt.Printf("\n3. 获取股票详情: %s...\n", testStock.TsCode)
	stockDetail, err := eastMoneyCollector.GetStockDetail(testStock.TsCode)
	if err != nil {
		fmt.Printf("❌ 获取股票详情失败: %v\n", err)
	} else {
		fmt.Printf("✅ 股票详情: %s - %s (%s %s)\n",
			stockDetail.TsCode, stockDetail.Name, stockDetail.Area, stockDetail.Industry)
	}

	// 测试获取日K线数据
	fmt.Printf("\n4. 获取日K线数据: %s (最近30天)...\n", testStock.TsCode)
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	dailyData, err := eastMoneyCollector.GetDailyKLine(testStock.TsCode, startDate, endDate)
	if err != nil {
		fmt.Printf("❌ 获取日K线数据失败: %v\n", err)
	} else {
		fmt.Printf("✅ 成功获取 %d 条日K线数据\n", len(dailyData))
		if len(dailyData) > 0 {
			latest := dailyData[len(dailyData)-1]
			fmt.Printf("  最新数据: %s 开盘:%.2f 收盘:%.2f 最高:%.2f 最低:%.2f 成交量:%d\n",
				latest.TradeDate.Format("2006-01-02"), latest.Open, latest.Close,
				latest.High, latest.Low, latest.Volume)
		}
	}

	// 测试获取财务数据
	fmt.Printf("\n5. 获取财务数据: %s...\n", testStock.TsCode)
	financialData, err := eastMoneyCollector.GetPerformanceReports(testStock.TsCode)
	if err != nil {
		fmt.Printf("❌ 获取财务数据失败: %v\n", err)
	} else {
		fmt.Printf("✅ 成功获取 %d 条财务数据\n", len(financialData))
		if len(financialData) > 0 {
			latest := financialData[0]
			fmt.Printf("  最新财务数据: %s ROE:%.2f%% ROA:%.2f%% 毛利率:%.2f%% 净利率:%.2f%%\n",
				latest.EndDate.Format("2006-01-02"), latest.ROE, latest.ROA,
				latest.GrossProfitMargin, latest.NetProfitMargin)
		}
	}

	// 测试获取实时数据
	fmt.Println("\n6. 获取实时数据（前5只股票）...")
	testCodes := make([]string, 0, 5)
	for i, stock := range stocks {
		if i >= 5 {
			break
		}
		testCodes = append(testCodes, stock.TsCode)
	}

	realtimeData, err := eastMoneyCollector.GetRealtimeData(testCodes)
	if err != nil {
		fmt.Printf("❌ 获取实时数据失败: %v\n", err)
	} else {
		fmt.Printf("✅ 成功获取 %d 条实时数据\n", len(realtimeData))
		for _, data := range realtimeData {
			fmt.Printf("  %s: %.2f\n", data.TsCode, data.Close)
		}
	}

	fmt.Println("\n=== 数据采集器测试完成 ===")
}
