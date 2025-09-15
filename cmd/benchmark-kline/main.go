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
		Level:  "warn", // 减少日志输出以获得更准确的性能测试
		Format: "text",
	}
	logger := utils.NewLogger(logConfig)

	// 创建东方财富采集器
	eastMoneyCollector := collector.NewEastMoneyCollector(logger)

	// 连接采集器
	if err := eastMoneyCollector.Connect(); err != nil {
		log.Fatalf("Failed to connect to EastMoney collector: %v", err)
	}

	// 测试股票代码列表
	testCodes := []string{"001208.SZ", "000001.SZ", "600000.SH"}
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)

	fmt.Println("=== K-line Data Collection Performance Benchmark ===")
	fmt.Printf("Testing %d stocks from %s to %s\n\n", len(testCodes), startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	// 基准测试周K线数据
	fmt.Println("--- Weekly K-line Benchmark ---")
	weeklyStart := time.Now()
	totalWeeklyRecords := 0
	for _, tsCode := range testCodes {
		weeklyData, err := eastMoneyCollector.GetWeeklyKLine(tsCode, startDate, endDate)
		if err != nil {
			fmt.Printf("Error fetching weekly data for %s: %v\n", tsCode, err)
			continue
		}
		totalWeeklyRecords += len(weeklyData)
		fmt.Printf("%s: %d records\n", tsCode, len(weeklyData))
	}
	weeklyDuration := time.Since(weeklyStart)
	fmt.Printf("Total weekly records: %d\n", totalWeeklyRecords)
	fmt.Printf("Weekly K-line fetch time: %v\n", weeklyDuration)
	fmt.Printf("Average time per stock: %v\n", weeklyDuration/time.Duration(len(testCodes)))
	fmt.Printf("Records per second: %.2f\n\n", float64(totalWeeklyRecords)/weeklyDuration.Seconds())

	// 基准测试月K线数据
	fmt.Println("--- Monthly K-line Benchmark ---")
	monthlyStart := time.Now()
	totalMonthlyRecords := 0
	for _, tsCode := range testCodes {
		monthlyData, err := eastMoneyCollector.GetMonthlyKLine(tsCode, startDate, endDate)
		if err != nil {
			fmt.Printf("Error fetching monthly data for %s: %v\n", tsCode, err)
			continue
		}
		totalMonthlyRecords += len(monthlyData)
		fmt.Printf("%s: %d records\n", tsCode, len(monthlyData))
	}
	monthlyDuration := time.Since(monthlyStart)
	fmt.Printf("Total monthly records: %d\n", totalMonthlyRecords)
	fmt.Printf("Monthly K-line fetch time: %v\n", monthlyDuration)
	fmt.Printf("Average time per stock: %v\n", monthlyDuration/time.Duration(len(testCodes)))
	fmt.Printf("Records per second: %.2f\n\n", float64(totalMonthlyRecords)/monthlyDuration.Seconds())

	// 基准测试年K线数据
	fmt.Println("--- Yearly K-line Benchmark ---")
	yearlyStart := time.Now()
	totalYearlyRecords := 0
	for _, tsCode := range testCodes {
		yearlyData, err := eastMoneyCollector.GetYearlyKLine(tsCode, startDate, endDate)
		if err != nil {
			fmt.Printf("Error fetching yearly data for %s: %v\n", tsCode, err)
			continue
		}
		totalYearlyRecords += len(yearlyData)
		fmt.Printf("%s: %d records\n", tsCode, len(yearlyData))
	}
	yearlyDuration := time.Since(yearlyStart)
	fmt.Printf("Total yearly records: %d\n", totalYearlyRecords)
	fmt.Printf("Yearly K-line fetch time: %v\n", yearlyDuration)
	fmt.Printf("Average time per stock: %v\n", yearlyDuration/time.Duration(len(testCodes)))
	fmt.Printf("Records per second: %.2f\n\n", float64(totalYearlyRecords)/yearlyDuration.Seconds())

	// 总体统计
	totalDuration := weeklyDuration + monthlyDuration + yearlyDuration
	totalRecords := totalWeeklyRecords + totalMonthlyRecords + totalYearlyRecords
	fmt.Println("=== Overall Performance Summary ===")
	fmt.Printf("Total execution time: %v\n", totalDuration)
	fmt.Printf("Total records fetched: %d\n", totalRecords)
	fmt.Printf("Overall records per second: %.2f\n", float64(totalRecords)/totalDuration.Seconds())
	fmt.Printf("Average time per API call: %v\n", totalDuration/time.Duration(len(testCodes)*3))

	fmt.Println("\nBenchmark completed!")
}
