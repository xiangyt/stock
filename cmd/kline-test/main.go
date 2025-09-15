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
	// 创建日志器
	logConfig := config.LogConfig{
		Level:  "info",
		Format: "text",
	}
	logger := utils.NewLogger(logConfig)

	// 创建东方财富采集器
	eastMoneyCollector := collector.NewEastMoneyCollector(logger)

	// 连接采集器
	if err := eastMoneyCollector.Connect(); err != nil {
		log.Fatalf("Failed to connect to EastMoney collector: %v", err)
	}

	// 测试股票代码
	tsCode := "001208.SZ"
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)

	fmt.Printf("Testing K-line data collection for %s\n", tsCode)
	fmt.Printf("Date range: %s to %s\n\n", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	// 测试周K线数据
	fmt.Println("=== Testing Weekly K-line Data ===")
	weeklyData, err := eastMoneyCollector.GetWeeklyKLine(tsCode, startDate, endDate)
	if err != nil {
		fmt.Printf("Error fetching weekly data: %v\n", err)
	} else {
		fmt.Printf("Fetched %d weekly K-line records\n", len(weeklyData))
		if len(weeklyData) > 0 {
			fmt.Printf("First record: %+v\n", weeklyData[0])
			fmt.Printf("Last record: %+v\n", weeklyData[len(weeklyData)-1])
		}
	}
	fmt.Println()

	// 测试月K线数据
	fmt.Println("=== Testing Monthly K-line Data ===")
	monthlyData, err := eastMoneyCollector.GetMonthlyKLine(tsCode, startDate, endDate)
	if err != nil {
		fmt.Printf("Error fetching monthly data: %v\n", err)
	} else {
		fmt.Printf("Fetched %d monthly K-line records\n", len(monthlyData))
		if len(monthlyData) > 0 {
			fmt.Printf("First record: %+v\n", monthlyData[0])
			fmt.Printf("Last record: %+v\n", monthlyData[len(monthlyData)-1])
		}
	}
	fmt.Println()

	// 测试年K线数据
	fmt.Println("=== Testing Yearly K-line Data ===")
	yearlyData, err := eastMoneyCollector.GetYearlyKLine(tsCode, startDate, endDate)
	if err != nil {
		fmt.Printf("Error fetching yearly data: %v\n", err)
	} else {
		fmt.Printf("Fetched %d yearly K-line records\n", len(yearlyData))
		if len(yearlyData) > 0 {
			fmt.Printf("First record: %+v\n", yearlyData[0])
			fmt.Printf("Last record: %+v\n", yearlyData[len(yearlyData)-1])
		}
	}

	fmt.Println("\nK-line data collection test completed!")
}
