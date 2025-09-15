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
	fmt.Println("业绩报表持久化功能测试")
	fmt.Println("=" + fmt.Sprintf("%50s", "="))

	// 加载配置
	cfg, err := config.Load()
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

	// 自动迁移表结构
	if err := db.AutoMigrate(&model.PerformanceReport{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

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

	// 测试股票代码
	testStocks := []string{"001208.SZ"}

	fmt.Println("\n1. 测试业绩报表统计信息")
	testStatistics(ctx, performanceService)

	fmt.Println("\n2. 测试同步业绩报表数据")
	for _, tsCode := range testStocks {
		testSyncPerformanceReports(ctx, performanceService, tsCode)
		time.Sleep(2 * time.Second) // 避免请求过于频繁
	}

	fmt.Println("\n3. 测试获取业绩报表数据")
	for _, tsCode := range testStocks {
		testGetPerformanceReports(ctx, performanceService, tsCode)
	}

	fmt.Println("\n4. 测试获取最新业绩报表")
	for _, tsCode := range testStocks {
		testGetLatestPerformanceReport(ctx, performanceService, tsCode)
	}

	fmt.Println("\n5. 测试业绩排行榜")
	testTopPerformers(ctx, performanceService)

	fmt.Println("\n6. 最终统计信息")
	testStatistics(ctx, performanceService)

	fmt.Println("\n✅ 业绩报表持久化功能测试完成!")
}

func testStatistics(ctx context.Context, service *service.PerformanceService) {
	stats, err := service.GetStatistics(ctx)
	if err != nil {
		fmt.Printf("❌ 获取统计信息失败: %v\n", err)
		return
	}

	fmt.Printf("📊 统计信息:\n")
	fmt.Printf("   总报表数量: %v\n", stats["total_reports"])
	fmt.Printf("   股票数量: %v\n", stats["stock_count"])
	fmt.Printf("   最新报告日期: %v\n", stats["latest_report_date"])
	fmt.Printf("   最早报告日期: %v\n", stats["earliest_report_date"])
}

func testSyncPerformanceReports(ctx context.Context, service *service.PerformanceService, tsCode string) {
	fmt.Printf("🔄 同步股票 %s 的业绩报表数据...\n", tsCode)

	start := time.Now()
	err := service.SyncPerformanceReports(ctx, tsCode)
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("❌ 同步失败: %v (耗时: %v)\n", err, duration)
		return
	}

	fmt.Printf("✅ 同步成功 (耗时: %v)\n", duration)
}

func testGetPerformanceReports(ctx context.Context, service *service.PerformanceService, tsCode string) {
	fmt.Printf("📈 获取股票 %s 的业绩报表数据...\n", tsCode)

	reports, err := service.GetPerformanceReports(ctx, tsCode)
	if err != nil {
		fmt.Printf("❌ 获取失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 获取成功，共 %d 条记录\n", len(reports))
	if len(reports) > 0 {
		latest := reports[0]
		fmt.Printf("   最新报告: %s, EPS: %.4f, 营业收入: %.0f\n",
			latest.ReportDate.Format("2006-01-02"),
			latest.EPS,
			latest.Revenue)
	}
}

func testGetLatestPerformanceReport(ctx context.Context, service *service.PerformanceService, tsCode string) {
	fmt.Printf("📊 获取股票 %s 的最新业绩报表...\n", tsCode)

	report, err := service.GetLatestPerformanceReport(ctx, tsCode)
	if err != nil {
		fmt.Printf("❌ 获取失败: %v\n", err)
		return
	}

	if report == nil {
		fmt.Printf("⚠️  未找到业绩报表数据\n")
		return
	}

	fmt.Printf("✅ 获取成功\n")
	fmt.Printf("   报告日期: %s\n", report.ReportDate.Format("2006-01-02"))
	fmt.Printf("   每股收益: %.4f 元\n", report.EPS)
	fmt.Printf("   营业收入: %.0f 元\n", report.Revenue)
	fmt.Printf("   净利润: %.0f 元\n", report.NetProfit)
}

func testTopPerformers(ctx context.Context, service *service.PerformanceService) {
	fmt.Printf("🏆 获取业绩排行榜 (按EPS排序, 前5名)...\n")

	reports, err := service.GetTopPerformers(ctx, 5, "eps")
	if err != nil {
		fmt.Printf("❌ 获取失败: %v\n", err)
		return
	}

	if len(reports) == 0 {
		fmt.Printf("⚠️  暂无排行榜数据\n")
		return
	}

	fmt.Printf("✅ 获取成功，共 %d 条记录\n", len(reports))
	fmt.Printf("   排名 | 股票代码   | 报告日期   | EPS     | 营业收入\n")
	fmt.Printf("   -----|-----------|-----------|---------|----------\n")
	for i, report := range reports {
		fmt.Printf("   %2d   | %-9s | %-9s | %7.4f | %8.0f\n",
			i+1,
			report.TsCode,
			report.ReportDate.Format("2006-01-02"),
			report.EPS,
			report.Revenue)
	}
}
