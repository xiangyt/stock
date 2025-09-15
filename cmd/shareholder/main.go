package main

import (
	"fmt"
	"log"
	"os"

	"stock/internal/collector"
	"stock/internal/config"
	"stock/internal/database"
	"stock/internal/repository"
	"stock/internal/service"
	"stock/internal/utils"

	"github.com/spf13/cobra"
)

var (
	cfg                *config.Config
	shareholderService *service.ShareholderService
)

func init() {
	// 加载配置
	var err error
	cfg, err = config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	logger := utils.NewLogger(cfg.Log)

	// 连接数据库
	db, err := database.NewDatabase(&cfg.Database, logger)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 初始化仓库
	shareholderRepo := repository.NewShareholderRepository(db.DB)

	// 初始化采集器
	eastMoneyCollector := collector.NewEastMoneyCollector(logger)

	// 初始化服务
	shareholderService = service.NewShareholderService(shareholderRepo, eastMoneyCollector)
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "shareholder",
		Short: "股东户数数据管理工具",
		Long:  "用于管理股东户数数据的命令行工具，支持数据同步、查询功能。",
	}

	// 同步命令
	var syncCmd = &cobra.Command{
		Use:   "sync [股票代码]",
		Short: "同步股东户数数据",
		Long:  "从东方财富API同步指定股票的股东户数数据",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			tsCode := args[0]
			fmt.Printf("正在同步股票 %s 的股东户数数据...\n", tsCode)

			if err := shareholderService.SyncData(tsCode); err != nil {
				fmt.Printf("同步失败: %v\n", err)
				return
			}
			fmt.Printf("股票 %s 股东户数数据同步完成\n", tsCode)
		},
	}

	// 查询最新数据命令
	var latestCmd = &cobra.Command{
		Use:   "latest [股票代码]",
		Short: "获取最新股东户数数据",
		Long:  "获取指定股票的最新股东户数数据",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			tsCode := args[0]

			latest, err := shareholderService.GetLatest(tsCode)
			if err != nil {
				fmt.Printf("获取最新数据失败: %v\n", err)
				return
			}

			fmt.Printf("股票代码: %s\n", latest.TsCode)
			fmt.Printf("证券名称: %s\n", latest.SecurityName)
			fmt.Printf("统计截止日期: %s\n", latest.EndDate.Format("2006-01-02"))
			fmt.Printf("股东户数: %d 户\n", latest.HolderNum)
			fmt.Printf("上期股东户数: %d 户\n", latest.PreHolderNum)
			fmt.Printf("股东户数变化: %d 户\n", latest.HolderNumChange)
			fmt.Printf("户均市值: %.2f 元\n", latest.AvgMarketCap)
		},
	}

	// 查询历史数据命令
	var queryCmd = &cobra.Command{
		Use:   "query [股票代码]",
		Short: "查询股东户数历史数据",
		Long:  "查询指定股票的股东户数历史数据",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			tsCode := args[0]

			counts, err := shareholderService.GetByTsCode(tsCode)
			if err != nil {
				fmt.Printf("查询历史数据失败: %v\n", err)
				return
			}

			if len(counts) == 0 {
				fmt.Printf("未找到股票 %s 的股东户数数据\n", tsCode)
				return
			}

			fmt.Printf("股票 %s 股东户数历史数据 (共 %d 条记录):\n", tsCode, len(counts))
			fmt.Printf("%-12s %-10s %-12s %-12s %-15s\n", "统计日期", "股东户数", "户数变化", "变化比例", "户均市值")
			fmt.Println("--------------------------------------------------------------------")

			for _, count := range counts {
				fmt.Printf("%-12s %-10d %-12d %-12.2f%% %-15.2f\n",
					count.EndDate.Format("2006-01-02"),
					count.HolderNum,
					count.HolderNumChange,
					count.HolderNumRatio,
					count.AvgMarketCap)
			}
		},
	}

	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(latestCmd)
	rootCmd.AddCommand(queryCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
