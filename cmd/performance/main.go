package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
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
	// 定义命令行参数
	var (
		action     = flag.String("action", "help", "操作类型: sync, sync-all, stats, top, help")
		tsCode     = flag.String("code", "", "股票代码 (例如: 000001.SZ)")
		limit      = flag.Int("limit", 10, "返回数量限制")
		orderBy    = flag.String("order", "eps", "排序字段: eps, roe, roa, gross_margin, dividend_yield, revenue, net_profit")
		configPath = flag.String("config", "configs/app.yaml", "配置文件路径")
	)
	flag.Parse()

	// 加载配置
	cfg, err := config.Load()
	_ = configPath // 暂时忽略配置路径参数
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	logger := utils.NewLogger(cfg.Log)

	// 初始化数据库
	dbManager, err := database.NewDatabase(&cfg.Database, logger)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	db := dbManager.DB

	// 初始化采集器
	eastMoneyCollector := collector.NewEastMoneyCollector(logger)
	if err := eastMoneyCollector.Connect(); err != nil {
		log.Fatalf("Failed to connect to collector: %v", err)
	}

	// 初始化仓库和服务
	performanceRepo := repository.NewPerformanceRepository(db)
	stockRepo := repository.NewStockRepository(db, logger)
	performanceService := service.NewPerformanceService(performanceRepo, stockRepo, eastMoneyCollector, logger)

	ctx := context.Background()

	// 根据操作类型执行相应功能
	switch *action {
	case "sync":
		if *tsCode == "" {
			fmt.Println("错误: 同步单只股票需要指定股票代码")
			fmt.Println("使用方法: -action=sync -code=000001.SZ")
			os.Exit(1)
		}
		syncSingleStock(ctx, performanceService, *tsCode)

	case "sync-all":
		syncAllStocks(ctx, performanceService)

	case "stats":
		showStatistics(ctx, performanceService)

	case "top":
		showTopPerformers(ctx, performanceService, *limit, *orderBy)

	case "help":
		showHelp()

	default:
		fmt.Printf("未知操作: %s\n", *action)
		showHelp()
		os.Exit(1)
	}
}

// syncSingleStock 同步单只股票的业绩报表数据
func syncSingleStock(ctx context.Context, service *service.PerformanceService, tsCode string) {
	fmt.Printf("开始同步股票 %s 的业绩报表数据...\n", tsCode)

	start := time.Now()
	err := service.SyncPerformanceReports(ctx, tsCode)
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("同步失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("同步完成，耗时: %v\n", duration)

	// 显示同步后的数据
	reports, err := service.GetPerformanceReports(ctx, tsCode)
	if err != nil {
		fmt.Printf("获取数据失败: %v\n", err)
		return
	}

	fmt.Printf("共同步 %d 条业绩报表记录\n", len(reports))
	if len(reports) > 0 {
		fmt.Println("\n最新业绩报表:")
		printPerformanceReport(&reports[0])
	}
}

// syncAllStocks 同步所有股票的业绩报表数据
func syncAllStocks(ctx context.Context, service *service.PerformanceService) {
	fmt.Println("开始同步所有股票的业绩报表数据...")
	fmt.Println("警告: 这可能需要很长时间，请耐心等待...")

	start := time.Now()
	err := service.SyncAllStocksPerformanceReports(ctx)
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("同步失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("同步完成，总耗时: %v\n", duration)

	// 显示统计信息
	showStatistics(ctx, service)
}

// showStatistics 显示业绩报表统计信息
func showStatistics(ctx context.Context, service *service.PerformanceService) {
	fmt.Println("业绩报表统计信息:")
	fmt.Println(strings.Repeat("=", 50))

	stats, err := service.GetStatistics(ctx)
	if err != nil {
		fmt.Printf("获取统计信息失败: %v\n", err)
		return
	}

	fmt.Printf("总报表数量: %v\n", stats["total_reports"])
	fmt.Printf("股票数量: %v\n", stats["stock_count"])
	fmt.Printf("最新报告日期: %v\n", stats["latest_report_date"])
	fmt.Printf("最早报告日期: %v\n", stats["earliest_report_date"])
}

// showTopPerformers 显示业绩表现最好的股票
func showTopPerformers(ctx context.Context, service *service.PerformanceService, limit int, orderBy string) {
	fmt.Printf("业绩排行榜 (按 %s 排序, 前 %d 名):\n", orderBy, limit)
	fmt.Println(strings.Repeat("=", 80))

	reports, err := service.GetTopPerformers(ctx, limit, orderBy)
	if err != nil {
		fmt.Printf("获取排行榜失败: %v\n", err)
		return
	}

	if len(reports) == 0 {
		fmt.Println("暂无数据")
		return
	}

	fmt.Printf("%-12s %-10s %-8s %-12s %-12s %-8s %-8s\n",
		"股票代码", "报告日期", "EPS", "营业收入", "净利润", "毛利率", "股息率")
	fmt.Println(strings.Repeat("-", 80))

	for i, report := range reports {
		fmt.Printf("%2d. %-12s %-10s %8.4f %12.0f %12.0f %8.2f %8.2f\n",
			i+1,
			report.TsCode,
			report.ReportDate.Format("2006-01-02"),
			report.EPS,
			report.Revenue,
			report.NetProfit,
			report.GrossMargin,
			report.DividendYield,
		)
	}
}

// printPerformanceReport 打印业绩报表详情
func printPerformanceReport(report *model.PerformanceReport) {
	fmt.Printf("股票代码: %s\n", report.TsCode)
	fmt.Printf("报告日期: %s\n", report.ReportDate.Format("2006-01-02"))
	fmt.Printf("每股收益(EPS): %.4f 元\n", report.EPS)
	fmt.Printf("加权每股收益: %.4f 元\n", report.WeightEPS)
	fmt.Printf("营业收入: %.0f 元\n", report.Revenue)
	fmt.Printf("营业收入同比增长: %.2f%%\n", report.RevenueYoY)
	fmt.Printf("净利润: %.0f 元\n", report.NetProfit)
	fmt.Printf("净利润同比增长: %.2f%%\n", report.NetProfitYoY)
	fmt.Printf("每股净资产(BVPS): %.4f 元\n", report.BVPS)
	fmt.Printf("销售毛利率: %.2f%%\n", report.GrossMargin)
	fmt.Printf("股息率: %.2f%%\n", report.DividendYield)
}

// showHelp 显示帮助信息
func showHelp() {
	fmt.Println("业绩报表管理工具")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()
	fmt.Println("使用方法:")
	fmt.Println("  go run cmd/performance/main.go [选项]")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("  -action string")
	fmt.Println("        操作类型 (默认: help)")
	fmt.Println("        可选值: sync, sync-all, stats, top, help")
	fmt.Println()
	fmt.Println("  -code string")
	fmt.Println("        股票代码，用于 sync 操作 (例如: 000001.SZ)")
	fmt.Println()
	fmt.Println("  -limit int")
	fmt.Println("        返回数量限制，用于 top 操作 (默认: 10)")
	fmt.Println()
	fmt.Println("  -order string")
	fmt.Println("        排序字段，用于 top 操作 (默认: eps)")
	fmt.Println("        可选值: eps, roe, roa, gross_margin, dividend_yield, revenue, net_profit")
	fmt.Println()
	fmt.Println("  -config string")
	fmt.Println("        配置文件路径 (默认: configs/app.yaml)")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  # 同步单只股票")
	fmt.Println("  go run cmd/performance/main.go -action=sync -code=000001.SZ")
	fmt.Println()
	fmt.Println("  # 同步所有股票 (耗时较长)")
	fmt.Println("  go run cmd/performance/main.go -action=sync-all")
	fmt.Println()
	fmt.Println("  # 查看统计信息")
	fmt.Println("  go run cmd/performance/main.go -action=stats")
	fmt.Println()
	fmt.Println("  # 查看业绩排行榜")
	fmt.Println("  go run cmd/performance/main.go -action=top -limit=20 -order=roe")
}
