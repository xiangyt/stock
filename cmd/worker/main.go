package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"stock/internal/collector"
	"stock/internal/config"
	"stock/internal/database"
	"stock/internal/model"
	"stock/internal/repository"
	"stock/internal/service"
	"stock/internal/utils"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

func main() {
	// 初始化配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	logger := utils.NewLogger(cfg.Log)

	// 初始化数据库连接
	dbManager, err := database.NewDatabase(&cfg.Database, logger)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	db := dbManager.DB

	// 初始化服务
	services, err := initServicesWithDB(cfg, db, logger)
	if err != nil {
		logger.Fatalf("Failed to initialize services: %v", err)
	}

	// 检查是否为测试模式
	if len(os.Args) > 1 {
		testMode := os.Args[1]
		switch testMode {
		case "migrate":
			runMigration(cfg, logger)
			return
		case "help":
			printTestHelp()
			return
		}
	}

	logger.Info("Worker starting...")

	// 创建定时任务调度器
	c := cron.New(cron.WithSeconds())

	// 设置定时任务
	setupCronJobs(c, services, logger)

	// 启动调度器
	c.Start()
	logger.Info("Cron scheduler started")

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Worker shutting down...")

	// 停止调度器
	ctx := c.Stop()
	<-ctx.Done()

	logger.Info("Worker exited")
}

func setupCronJobs(c *cron.Cron, services *service.Services, logger *utils.Logger) {
	// 每天早上8点执行主要任务
	c.AddFunc("0 0 8 * * *", func() {
		logger.Info("开始执行每日早上8点定时任务...")
		services.DataService.SyncStockList()
	})

	// 每天晚上10点执行主要任务
	c.AddFunc("0 0 22 * * *", func() {
		logger.Info("开始执行每日晚上10点定时任务...")
		executeNightlyTasks(services, logger)
	})

	logger.Info("定时任务配置完成！")
}

// executeNightlyTasks 执行每日晚上10点的主要任务
func executeNightlyTasks(services *service.Services, logger *utils.Logger) {
	// 日K线数据采集 - 第一优先级
	_ = collectAndPersistDailyKLineData(services, logger)
	// 周K线数据采集 - 第二优先级
	_ = collectAndPersistWeeklyKLineData(services, logger)
	// 月K线数据采集 - 第三优先级
	_ = collectAndPersistMonthlyKLineData(services, logger)
	// 年K线数据采集 - 第四优先级
	_ = collectAndPersistYearlyKLineData(services, logger)
	// 业绩报表数据采集 - 第五优先级
	_ = collectAndPersistPerformanceReports(services, logger)
	// 股东人数数据采集 - 第六优先级
	_ = collectAndPersistShareholderCounts(services, logger)

}

// initServicesWithDB 初始化带数据库连接的服务
func initServicesWithDB(cfg *config.Config, db *gorm.DB, logger *utils.Logger) (*service.Services, error) {
	logger.Info("初始化服务...")

	// 创建基础服务
	services, err := service.NewServices(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("创建基础服务失败: %v", err)
	}

	// 初始化需要数据库连接的服务
	services.DataService = service.NewDataService(db, logger)

	// 为PerformanceService创建必要的依赖
	performanceRepo := repository.NewPerformanceRepository(db)
	stockRepo := repository.NewStockRepository(db, logger)
	eastMoneyCollector := collector.NewEastMoneyCollector(logger)
	services.PerformanceService = service.NewPerformanceService(performanceRepo, stockRepo, eastMoneyCollector, logger)

	// 为ShareholderService创建必要的依赖
	shareholderRepo := repository.NewShareholderRepository(db)
	services.ShareholderService = service.NewShareholderService(shareholderRepo, eastMoneyCollector)

	logger.Info("所有服务初始化完成")
	return services, nil
}

// collectStockBasicInfo 采集股票基础信息
func collectStockBasicInfo(services *service.Services, logger *utils.Logger) error {
	logger.Info("开始采集股票基础信息...")

	// 检查DataService是否已初始化
	if services.DataService == nil {
		return fmt.Errorf("DataService未初始化，请先初始化数据库连接")
	}

	// 同步股票基础信息
	err := services.DataService.SyncStockList()
	if err != nil {
		return fmt.Errorf("股票信息同步失败: %v", err)
	}

	logger.Info("股票基础信息采集完成")
	return nil
}

// runMigration 运行数据库迁移
func runMigration(cfg *config.Config, logger *utils.Logger) {
	logger.Info("开始数据库迁移...")

	// 初始化数据库连接
	dbManager, err := database.NewDatabase(&cfg.Database, logger)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbManager.Close()

	// 执行自动迁移
	if err := dbManager.AutoMigrate(); err != nil {
		logger.Fatalf("数据库迁移失败: %v", err)
	}

	logger.Info("数据库迁移完成!")
}

// collectAndPersistDailyKLineData 采集并保存日K线数据
func collectAndPersistDailyKLineData(services *service.Services, logger *utils.Logger) error {
	logger.Info("开始采集日K线数据...")

	executor := utils.NewConcurrentExecutor(100, logger, 30*time.Minute) // 最大100个并发，30分钟超时
	defer executor.Close()
	ctx := context.Background()

	// 从数据库获取所有股票列表
	stocks, err := services.DataService.GetAllStocks()
	if err != nil {
		return fmt.Errorf("获取股票列表失败: %v", err)
	}

	logger.Infof("从数据库获取到 %d 只股票，开始采集日K线数据", len(stocks))

	// 创建任务列表
	tasks := make([]utils.Task, len(stocks))
	for i, stock := range stocks {
		stock := stock // 避免闭包问题
		tasks[i] = &utils.SimpleTask{
			ID:          fmt.Sprintf("daily_kline_%s", stock.TsCode),
			Description: fmt.Sprintf("采集股票 %s 的日K线数据", stock.TsCode),
			Func: func(ctx context.Context) error {
				return syncStockDailyKLine(services, stock, logger)
			},
		}
	}

	// 执行任务
	results, stats := executor.ExecuteBatch(ctx, tasks)

	// 统计结果
	successCount := 0

	for _, result := range results {
		if result.Success {
			successCount++
		} else {
			logger.Errorf("股票日K线采集失败: %v", result.Error)
		}
	}

	logger.Infof("日K线数据采集完成 - 总数: %d, 成功: %d, 失败: %d, 总耗时: %v, 平均耗时: %v",
		stats.TotalTasks, successCount, stats.FailedTasks, stats.TotalDuration, stats.AverageDuration)

	return nil
}

// syncStockDailyKLine 同步单只股票的日K线数据
func syncStockDailyKLine(services *service.Services, stock *model.Stock, logger *utils.Logger) error {
	// 获取该股票最新的日K线数据
	latestData, err := services.DataService.GetLatestPrice(stock.TsCode)
	if err != nil {
		return fmt.Errorf("获取最新日K线数据失败: %v", err)
	}

	var startDate time.Time
	if latestData == nil {
		// 数据库中没有数据，进行全量同步
		startDate = time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
		logger.Infof("股票 %s 进行全量日K线同步，起始日期: %s", stock.TsCode, startDate.Format("2006-01-02"))
	} else {
		// 将TradeDate从int转换为time.Time进行比较
		tradeDateStr := fmt.Sprintf("%d", latestData.TradeDate)
		tradeDate, err := time.Parse("20060102", tradeDateStr)
		if err != nil {
			return fmt.Errorf("解析交易日期失败: %v", err)
		}
		startDate = tradeDate
	}

	// 实现真正的数据同步逻辑
	endDate := time.Now()
	logger.Debugf("股票 %s 需要同步日K线数据，时间范围: %s 到 %s",
		stock.TsCode, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	// 调用DataService进行数据同步
	syncCount, err := services.DataService.SyncDailyData(stock.TsCode, startDate, endDate)
	if err != nil {
		return fmt.Errorf("同步日K线数据失败: %v", err)
	}

	logger.Debugf("股票 %s 日K线数据同步完成，共同步 %d 条记录", stock.TsCode, syncCount)

	latestData, _ = services.DataService.GetLatestPrice(stock.TsCode)
	if latestData != nil { // 日k一个月没更新，可能已经退市了
		tradeDateStr := fmt.Sprintf("%d", latestData.TradeDate)
		tradeDate, err := time.Parse("20060102", tradeDateStr)
		if err != nil {
			return fmt.Errorf("解析交易日期失败: %v", err)
		}
		// 检查最新数据是否超过一个月
		oneMonthAgo := time.Now().AddDate(0, -1, 0)
		if tradeDate.Before(oneMonthAgo) {
			// 标记为非活跃股票
			if err := markStockInactive(services, stock.TsCode, logger); err != nil {
				logger.Errorf("标记股票 %s 为非活跃状态失败: %v", stock.TsCode, err)
			}
		}
	}
	return nil
}

// markStockInactive 标记股票为非活跃状态
func markStockInactive(services *service.Services, tsCode string, logger *utils.Logger) error {
	logger.Infof("标记股票 %s 为非活跃状态", tsCode)

	// 获取股票信息
	stock, err := services.DataService.GetStockInfo(tsCode)
	if err != nil {
		return fmt.Errorf("获取股票信息失败: %v", err)
	}

	// 检查股票是否已经是非活跃状态
	if !stock.IsActive {
		logger.Debugf("股票 %s 已经是非活跃状态", tsCode)
		return nil
	}

	// 更新股票状态为非活跃
	err = services.DataService.UpdateStockStatus(tsCode, false)
	if err != nil {
		return fmt.Errorf("更新股票状态失败: %v", err)
	}

	logger.Infof("成功标记股票 %s 为非活跃状态", tsCode)
	return nil
}

// collectAndPersistWeeklyKLineData 采集并保存周K线数据
func collectAndPersistWeeklyKLineData(services *service.Services, logger *utils.Logger) error {
	logger.Info("开始采集周K线数据...")
	executor := utils.NewConcurrentExecutor(100, logger, 30*time.Minute) // 最大100个并发，30分钟超时
	defer executor.Close()
	ctx := context.Background()

	// 从数据库获取所有活跃股票列表
	stocks, err := services.DataService.GetAllStocks()
	if err != nil {
		return fmt.Errorf("获取股票列表失败: %v", err)
	}

	logger.Infof("从数据库获取到 %d 只股票，开始采集周K线数据", len(stocks))

	// 创建任务列表
	tasks := make([]utils.Task, len(stocks))
	for _, stock := range stocks {
		stock := stock // 避免闭包问题
		if !stock.IsActive {
			logger.Debugf("股票 %s 不活跃，跳过年K线数据同步", stock.TsCode)
			continue
		}
		tasks = append(tasks, &utils.SimpleTask{
			ID:          fmt.Sprintf("weekly_kline_%s", stock.TsCode),
			Description: fmt.Sprintf("采集股票 %s 的周K线数据", stock.TsCode),
			Func: func(ctx context.Context) error {
				return syncStockWeeklyKLine(services, stock, logger)
			},
		})
	}

	// 执行任务
	results, stats := executor.ExecuteBatch(ctx, tasks)

	// 统计结果
	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		}
	}

	logger.Infof("周K线数据采集完成 - 总数: %d, 成功: %d, 失败: %d, 总耗时: %v",
		stats.TotalTasks, successCount, stats.FailedTasks, stats.TotalDuration)

	return nil
}

// collectAndPersistMonthlyKLineData 采集并保存月K线数据
func collectAndPersistMonthlyKLineData(services *service.Services, logger *utils.Logger) error {
	logger.Info("开始采集月K线数据...")
	executor := utils.NewConcurrentExecutor(100, logger, 30*time.Minute) // 最大100个并发，30分钟超时
	defer executor.Close()
	ctx := context.Background()
	stocks, err := services.DataService.GetAllStocks()
	if err != nil {
		return fmt.Errorf("获取股票列表失败: %v", err)
	}

	logger.Infof("从数据库获取到 %d 只股票，开始采集月K线数据", len(stocks))

	tasks := make([]utils.Task, len(stocks))
	for _, stock := range stocks {
		stock := stock
		if !stock.IsActive {
			logger.Debugf("股票 %s 不活跃，跳过年K线数据同步", stock.TsCode)
			continue
		}
		tasks = append(tasks, &utils.SimpleTask{
			ID:          fmt.Sprintf("monthly_kline_%s", stock.TsCode),
			Description: fmt.Sprintf("采集股票 %s 的月K线数据", stock.TsCode),
			Func: func(ctx context.Context) error {
				return syncStockMonthlyKLine(services, stock, logger)
			},
		})
	}

	results, stats := executor.ExecuteBatch(ctx, tasks)
	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		}
	}

	logger.Infof("月K线数据采集完成 - 总数: %d, 成功: %d, 失败: %d, 总耗时: %v",
		stats.TotalTasks, successCount, stats.FailedTasks, stats.TotalDuration)

	return nil
}

// collectAndPersistYearlyKLineData 采集并保存年K线数据
func collectAndPersistYearlyKLineData(services *service.Services, logger *utils.Logger) error {
	logger.Info("开始采集年K线数据...")
	executor := utils.NewConcurrentExecutor(100, logger, 30*time.Minute) // 最大100个并发，30分钟超时
	defer executor.Close()
	ctx := context.Background()
	stocks, err := services.DataService.GetAllStocks()
	if err != nil {
		return fmt.Errorf("获取股票列表失败: %v", err)
	}

	logger.Infof("从数据库获取到 %d 只股票，开始采集年K线数据", len(stocks))

	tasks := make([]utils.Task, 0, len(stocks))
	for _, stock := range stocks {
		stock := stock
		if !stock.IsActive {
			logger.Debugf("股票 %s 不活跃，跳过年K线数据同步", stock.TsCode)
			continue
		}
		tasks = append(tasks, &utils.SimpleTask{
			ID:          fmt.Sprintf("yearly_kline_%s", stock.TsCode),
			Description: fmt.Sprintf("采集股票 %s 的年K线数据", stock.TsCode),
			Func: func(ctx context.Context) error {
				return syncStockYearlyKLine(services, stock, logger)
			},
		})
	}

	results, stats := executor.ExecuteBatch(ctx, tasks)
	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		}
	}

	logger.Infof("年K线数据采集完成 - 总数: %d, 成功: %d, 失败: %d, 总耗时: %v",
		stats.TotalTasks, successCount, stats.FailedTasks, stats.TotalDuration)

	return nil
}

// syncStockWeeklyKLine 同步单只股票的周K线数据
func syncStockWeeklyKLine(services *service.Services, stock *model.Stock, logger *utils.Logger) error {

	// 第二步：查出该股票最新的一条周K线数据
	klinePersistence := service.GetKLinePersistenceService(services.DataService.GetDB(), logger)
	latestWeeklyData, err := klinePersistence.GetLatestWeeklyData(stock.TsCode)
	if err != nil {
		return fmt.Errorf("获取最新周K线数据失败: %v", err)
	}

	// 第三步：确定采集的起始时间
	var startDate time.Time
	if latestWeeklyData == nil {
		// 如果没有最新一条数据，默认起始时间为1990年1月1日
		startDate = time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
		logger.Debugf("股票 %s 没有历史周K线数据，从1990年1月1日开始采集", stock.TsCode)
	} else {
		// 删除最新的一条周K线数据，确保数据完整性
		tradeDateStr := fmt.Sprintf("%d", latestWeeklyData.TradeDate)
		tradeDate, err := time.Parse("20060102", tradeDateStr)
		if err != nil {
			return fmt.Errorf("解析最新周K线交易日期失败: %v", err)
		}

		// 删除最新的周K线数据
		if err := klinePersistence.DeleteData(stock.TsCode, tradeDate, "weekly"); err != nil {
			logger.Errorf("删除最新周K线数据失败: %v", err)
			return fmt.Errorf("删除最新周K线数据失败: %v", err)
		}
		logger.Debugf("已删除股票 %s 最新的周K线数据，交易日期: %d", stock.TsCode, latestWeeklyData.TradeDate)

		// 从最新一条数据的时间开始采集
		startDate = tradeDate
		logger.Debugf("股票 %s 从最新周K线数据日期 %s 开始采集", stock.TsCode, startDate.Format("2006-01-02"))
	}

	// 采集到当前时间的数据
	endDate := time.Now()

	logger.Debugf("股票 %s 需要同步周K线数据，时间范围: %s 到 %s",
		stock.TsCode, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	// 调用DataService进行数据同步
	syncCount, err := services.DataService.SyncWeeklyData(stock.TsCode, startDate, endDate)
	if err != nil {
		return fmt.Errorf("同步周K线数据失败: %v", err)
	}

	logger.Debugf("股票 %s 周K线数据同步完成，共同步 %d 条记录", stock.TsCode, syncCount)
	return nil
}

// syncStockMonthlyKLine 同步单只股票的月K线数据
func syncStockMonthlyKLine(services *service.Services, stock *model.Stock, logger *utils.Logger) error {

	// 第二步：查出该股票最新的一条月K线数据
	klinePersistence := service.GetKLinePersistenceService(services.DataService.GetDB(), logger)
	latestMonthlyData, err := klinePersistence.GetLatestMonthlyData(stock.TsCode)
	if err != nil {
		return fmt.Errorf("获取最新月K线数据失败: %v", err)
	}

	// 第三步：确定采集的起始时间
	var startDate time.Time
	if latestMonthlyData == nil {
		// 如果没有最新一条数据，默认起始时间为1990年1月1日
		startDate = time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
		logger.Debugf("股票 %s 没有历史月K线数据，从1990年1月1日开始采集", stock.TsCode)
	} else {
		// 删除最新的一条月K线数据，确保数据完整性
		tradeDateStr := fmt.Sprintf("%d", latestMonthlyData.TradeDate)
		tradeDate, err := time.Parse("20060102", tradeDateStr)
		if err != nil {
			return fmt.Errorf("解析最新月K线交易日期失败: %v", err)
		}

		// 删除最新的月K线数据
		if err := klinePersistence.DeleteData(stock.TsCode, tradeDate, "monthly"); err != nil {
			logger.Errorf("删除最新月K线数据失败: %v", err)
			return fmt.Errorf("删除最新月K线数据失败: %v", err)
		}
		logger.Debugf("已删除股票 %s 最新的月K线数据，交易日期: %d", stock.TsCode, latestMonthlyData.TradeDate)

		// 从最新一条数据的时间开始采集
		startDate = tradeDate
		logger.Debugf("股票 %s 从最新月K线数据日期 %s 开始采集", stock.TsCode, startDate.Format("2006-01-02"))
	}

	// 采集到当前时间的数据
	endDate := time.Now()

	logger.Debugf("股票 %s 需要同步月K线数据，时间范围: %s 到 %s",
		stock.TsCode, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	// 调用DataService进行数据同步
	syncCount, err := services.DataService.SyncMonthlyData(stock.TsCode, startDate, endDate)
	if err != nil {
		return fmt.Errorf("同步月K线数据失败: %v", err)
	}

	logger.Debugf("股票 %s 月K线数据同步完成，共同步 %d 条记录", stock.TsCode, syncCount)
	return nil
}

// syncStockYearlyKLine 同步单只股票的年K线数据
func syncStockYearlyKLine(services *service.Services, stock *model.Stock, logger *utils.Logger) error {
	// 第一步：检查股票是否活跃，不活跃直接跳过
	if !stock.IsActive {
		logger.Debugf("股票 %s 不活跃，跳过年K线数据同步", stock.TsCode)
		return fmt.Errorf("stock_inactive")
	}

	// 第二步：查出该股票最新的一条年K线数据
	klinePersistence := service.GetKLinePersistenceService(services.DataService.GetDB(), logger)
	latestYearlyData, err := klinePersistence.GetLatestYearlyData(stock.TsCode)
	if err != nil {
		return fmt.Errorf("获取最新年K线数据失败: %v", err)
	}

	// 第三步：确定采集的起始时间
	var startDate time.Time
	if latestYearlyData == nil {
		// 如果没有最新一条数据，默认起始时间为1990年1月1日
		startDate = time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
		logger.Debugf("股票 %s 没有历史年K线数据，从1990年1月1日开始采集", stock.TsCode)
	} else {
		// 删除最新的一条年K线数据，确保数据完整性
		tradeDateStr := fmt.Sprintf("%d", latestYearlyData.TradeDate)
		tradeDate, err := time.Parse("20060102", tradeDateStr)
		if err != nil {
			return fmt.Errorf("解析最新年K线交易日期失败: %v", err)
		}

		// 删除最新的年K线数据
		if err := klinePersistence.DeleteData(stock.TsCode, tradeDate, "yearly"); err != nil {
			logger.Errorf("删除最新年K线数据失败: %v", err)
			return fmt.Errorf("删除最新年K线数据失败: %v", err)
		}
		logger.Debugf("已删除股票 %s 最新的年K线数据，交易日期: %d", stock.TsCode, latestYearlyData.TradeDate)

		// 从最新一条数据的时间开始采集
		startDate = tradeDate
		logger.Debugf("股票 %s 从最新年K线数据日期 %s 开始采集", stock.TsCode, startDate.Format("2006-01-02"))
	}

	// 采集到当前时间的数据
	endDate := time.Now()

	logger.Debugf("股票 %s 需要同步年K线数据，时间范围: %s 到 %s",
		stock.TsCode, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	// 调用DataService进行数据同步
	syncCount, err := services.DataService.SyncYearlyData(stock.TsCode, startDate, endDate)
	if err != nil {
		return fmt.Errorf("同步年K线数据失败: %v", err)
	}

	logger.Debugf("股票 %s 年K线数据同步完成，共同步 %d 条记录", stock.TsCode, syncCount)
	return nil
}

// collectAndPersistPerformanceReports 采集并保存业绩报表数据
func collectAndPersistPerformanceReports(services *service.Services, logger *utils.Logger) error {
	logger.Info("开始采集业绩报表数据...")
	executor := utils.NewConcurrentExecutor(100, logger, 30*time.Minute) // 最大100个并发，30分钟超时
	defer executor.Close()
	ctx := context.Background()

	// 从数据库获取所有股票列表
	stocks, err := services.DataService.GetAllStocks()
	if err != nil {
		return fmt.Errorf("获取股票列表失败: %v", err)
	}

	logger.Infof("从数据库获取到 %d 只股票，开始采集业绩报表数据", len(stocks))

	// 创建并发任务列表
	var tasks []utils.Task

	for _, stock := range stocks {
		// 只处理活跃股票
		if !stock.IsActive {
			logger.Debugf("股票 %s 不活跃，跳过股东人数数据采集", stock.TsCode)
			continue
		}

		// 为每只股票创建一个采集任务
		tsCode := stock.TsCode // 捕获循环变量
		task := &utils.SimpleTask{
			ID:          fmt.Sprintf("performance-report-%s", tsCode),
			Description: fmt.Sprintf("采集股票 %s 的业绩报表", tsCode),
			Func: func(ctx context.Context) error {
				return services.PerformanceService.SyncPerformanceReports(ctx, tsCode)
			},
		}
		tasks = append(tasks, task)
	}

	// 执行任务
	results, stats := executor.ExecuteBatch(ctx, tasks)

	// 统计结果
	successCount := 0
	totalReports := 0

	for _, result := range results {
		if result.Success {
			successCount++
			// 成功的任务计数（每个任务代表一只股票的报表同步）
			totalReports++
		} else {
			logger.Errorf("业绩报表采集失败: %v", result.Error)
		}
	}

	logger.Infof("业绩报表数据采集完成 - 总数: %d, 成功: %d, 失败: %d, 同步报表: %d, 总耗时: %v, 平均耗时: %v",
		stats.TotalTasks, successCount, stats.FailedTasks, totalReports, stats.TotalDuration, stats.AverageDuration)

	return nil
}

// printTestHelp 打印测试帮助信息
func printTestHelp() {
	fmt.Println("测试模式使用说明:")
	fmt.Println("  go run cmd/worker/main.go test-stock    - 测试股票基本数据采集")
	fmt.Println("  go run cmd/worker/main.go test-kline    - 测试所有K线数据采集")
	fmt.Println("  go run cmd/worker/main.go test-daily    - 测试日K线数据采集")
	fmt.Println("  go run cmd/worker/main.go test-weekly   - 测试周K线数据采集")
	fmt.Println("  go run cmd/worker/main.go test-monthly  - 测试月K线数据采集")
	fmt.Println("  go run cmd/worker/main.go test-yearly   - 测试年K线数据采集")
	fmt.Println("  go run cmd/worker/main.go help          - 显示此帮助信息")
}

// collectAndPersistShareholderCounts 采集并保存股东人数数据
func collectAndPersistShareholderCounts(services *service.Services, logger *utils.Logger) error {
	logger.Info("开始采集股东人数数据...")

	executor := utils.NewConcurrentExecutor(50, logger, 45*time.Minute) // 最大50个并发，45分钟超时
	defer executor.Close()
	ctx := context.Background()

	// 从数据库获取所有活跃股票列表
	stocks, err := services.DataService.GetAllStocks()
	if err != nil {
		return fmt.Errorf("获取股票列表失败: %v", err)
	}

	logger.Infof("从数据库获取到 %d 只股票，开始采集股东人数数据", len(stocks))

	// 创建并发任务列表
	var tasks []utils.Task

	for _, stock := range stocks {
		// 只处理活跃股票
		if !stock.IsActive {
			logger.Debugf("股票 %s 不活跃，跳过股东人数数据采集", stock.TsCode)
			continue
		}

		// 为每只股票创建一个采集任务
		tsCode := stock.TsCode // 捕获循环变量
		task := &utils.SimpleTask{
			ID:          fmt.Sprintf("shareholder-count-%s", tsCode),
			Description: fmt.Sprintf("采集股票 %s 的股东人数", tsCode),
			Func: func(ctx context.Context) error {
				return services.ShareholderService.SyncShareholderCounts(tsCode)
			},
		}
		tasks = append(tasks, task)
	}

	if len(tasks) == 0 {
		logger.Warn("没有找到需要采集股东人数数据的活跃股票")
		return nil
	}

	// 执行任务
	results, stats := executor.ExecuteBatch(ctx, tasks)

	// 统计结果
	successCount := 0
	totalCounts := 0

	for _, result := range results {
		if result.Success {
			successCount++
			totalCounts++
		} else {
			logger.Errorf("股东人数采集失败: %v", result.Error)
		}
	}

	logger.Infof("股东人数数据采集完成 - 总数: %d, 成功: %d, 失败: %d, 同步股票: %d, 总耗时: %v, 平均耗时: %v",
		stats.TotalTasks, successCount, stats.FailedTasks, totalCounts, stats.TotalDuration, stats.AverageDuration)

	return nil
}
