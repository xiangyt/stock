package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"stock/internal/collector"
	"stock/internal/config"
	"stock/internal/database"
	"stock/internal/model"
	"stock/internal/repository"
	"stock/internal/service"
	"stock/internal/utils"
)

func main() {
	fmt.Println("=== 股东户数采集系统测试 ===")

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	logger := utils.NewLogger(cfg.Log)

	// 初始化数据库
	database, err := database.NewDatabase(&cfg.Database, logger)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 初始化仓库
	shareholderRepo := repository.NewShareholderRepository(database.DB)
	stockRepo := repository.NewStockRepository(database.DB, logger)

	// 初始化采集器
	shareholderCollector := collector.NewShareholderCollector(logger)

	// 初始化服务
	shareholderService := service.NewShareholderService(shareholderRepo, stockRepo, shareholderCollector, logger)

	ctx := context.Background()

	// 测试股票代码列表
	testStocks := []string{
		"000001.SZ", // 平安银行
		"000002.SZ", // 万科A
		"600000.SH", // 浦发银行
		"600036.SH", // 招商银行
		"000858.SZ", // 五粮液
	}

	fmt.Println("\n1. 测试采集器直接获取数据")
	fmt.Println("================================")

	for _, tsCode := range testStocks[:2] { // 只测试前两个
		fmt.Printf("\n测试股票: %s\n", tsCode)

		// 测试获取最新数据
		count, err := shareholderCollector.GetLatestShareholderCount(tsCode)
		if err != nil {
			fmt.Printf("  获取最新数据失败: %v\n", err)
			continue
		}

		if count != nil {
			fmt.Printf("  最新数据: %s, 股东户数: %d, 户均市值: %.2f\n",
				count.EndDate.Format("2006-01-02"), count.HolderNum, count.AvgMarketCap)
		} else {
			fmt.Printf("  未获取到数据\n")
		}

		// 添加延迟避免请求过频
		time.Sleep(2 * time.Second)
	}

	fmt.Println("\n2. 测试服务层数据同步")
	fmt.Println("================================")

	for _, tsCode := range testStocks[:1] { // 只测试一个
		fmt.Printf("\n同步股票: %s\n", tsCode)

		err := shareholderService.SyncShareholderCounts(ctx, tsCode)
		if err != nil {
			fmt.Printf("  同步失败: %v\n", err)
			continue
		}

		fmt.Printf("  同步成功\n")

		// 查询同步后的数据
		counts, err := shareholderService.GetShareholderCounts(ctx, tsCode)
		if err != nil {
			fmt.Printf("  查询失败: %v\n", err)
			continue
		}

		fmt.Printf("  数据库中共有 %d 条记录\n", len(counts))

		if len(counts) > 0 {
			latest := counts[0]
			fmt.Printf("  最新记录: %s, 股东户数: %d, 户均市值: %.2f\n",
				latest.EndDate.Format("2006-01-02"), latest.HolderNum, latest.AvgMarketCap)
		}
	}

	fmt.Println("\n3. 测试统计功能")
	fmt.Println("================================")

	// 获取统计信息
	stats, err := shareholderService.GetStatistics(ctx)
	if err != nil {
		fmt.Printf("获取统计信息失败: %v\n", err)
	} else {
		fmt.Printf("总记录数: %v\n", stats["total_records"])
		fmt.Printf("股票数量: %v\n", stats["total_stocks"])
		fmt.Printf("最新更新时间: %v\n", stats["latest_update"])
	}

	// 获取股东户数排行榜
	fmt.Println("\n股东户数排行榜 (前5名):")
	topHolders, err := shareholderService.GetTopByHolderNum(ctx, 5, false)
	if err != nil {
		fmt.Printf("获取排行榜失败: %v\n", err)
	} else {
		for i, count := range topHolders {
			fmt.Printf("%d. %s (%s): %d户, %s\n",
				i+1, count.TsCode, count.SecurityName, count.HolderNum,
				count.EndDate.Format("2006-01-02"))
		}
	}

	// 获取户均市值排行榜
	fmt.Println("\n户均市值排行榜 (前5名):")
	topMarketCap, err := shareholderService.GetTopByAvgMarketCap(ctx, 5, false)
	if err != nil {
		fmt.Printf("获取排行榜失败: %v\n", err)
	} else {
		for i, count := range topMarketCap {
			fmt.Printf("%d. %s (%s): %.2f元, %s\n",
				i+1, count.TsCode, count.SecurityName, count.AvgMarketCap,
				count.EndDate.Format("2006-01-02"))
		}
	}

	fmt.Println("\n4. 测试数据验证")
	fmt.Println("================================")

	// 测试数据验证
	testCount := &model.ShareholderCount{
		TsCode:       "000001.SZ",
		SecurityCode: "000001",
		SecurityName: "平安银行",
		EndDate:      time.Now(),
		HolderNum:    100000,
	}

	err = shareholderService.ValidateShareholderCount(testCount)
	if err != nil {
		fmt.Printf("数据验证失败: %v\n", err)
	} else {
		fmt.Printf("数据验证通过\n")
	}

	fmt.Println("\n=== 测试完成 ===")
}
