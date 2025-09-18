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

	// 检查是否为测试模式
	if len(os.Args) > 1 {
		testMode := os.Args[1]
		switch testMode {
		case "test-stock":
			testStockCollection(cfg, logger)
			return
		case "test-kline":
			testKLineCollection(cfg, logger)
			return
		case "test-daily":
			testDailyKLineCollection(cfg, logger)
			return
		case "test-weekly":
			testWeeklyKLineCollection(cfg, logger)
			return
		case "test-monthly":
			testMonthlyKLineCollection(cfg, logger)
			return
		case "test-yearly":
			testYearlyKLineCollection(cfg, logger)
			return
		case "migrate":
			runMigration(cfg, logger)
			return
		case "help":
			printTestHelp()
			return
		}
	}

	logger.Info("Worker starting...")

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
	// 每天晚上10点执行主要任务
	c.AddFunc("0 0 22 * * *", func() {
		// 创建并发执行器
		executor := utils.NewConcurrentExecutor(100, logger, 30*time.Minute) // 最大100个并发，30分钟超时
		defer executor.Close()
		logger.Info("开始执行每日晚上10点定时任务...")
		executeNightlyTasks(services, executor, logger)
	})

	logger.Info("定时任务配置完成:")
	logger.Info("- 每日主要任务: 22:00 (每天)")
}

// executeNightlyTasks 执行每日晚上10点的主要任务
func executeNightlyTasks(services *service.Services, executor *utils.ConcurrentExecutor, logger *utils.Logger) {
	ctx := context.Background()

	// 创建任务列表
	tasks := []utils.Task{
		// 股票基础信息采集 - 第一优先级
		&utils.SimpleTask{
			ID:          "stock-info-collection",
			Description: "股票基础信息采集",
			Func: func(ctx context.Context) error {
				return collectAndPersistStockInfo(services, executor, logger)
			},
		},

		// 日K线数据采集 - 第二优先级
		&utils.SimpleTask{
			ID:          "daily-kline-collection",
			Description: "日K线数据采集",
			Func: func(ctx context.Context) error {
				return collectAndPersistDailyKLineData(services, executor, logger)
			},
		},

		// 周K线数据采集 - 第三优先级
		&utils.SimpleTask{
			ID:          "weekly-kline-collection",
			Description: "周K线数据采集",
			Func: func(ctx context.Context) error {
				return collectAndPersistWeeklyKLineData(services, executor, logger)
			},
		},

		// 月K线数据采集 - 第四优先级
		&utils.SimpleTask{
			ID:          "monthly-kline-collection",
			Description: "月K线数据采集",
			Func: func(ctx context.Context) error {
				return collectAndPersistMonthlyKLineData(services, executor, logger)
			},
		},

		// 年K线数据采集 - 第五优先级
		&utils.SimpleTask{
			ID:          "yearly-kline-collection",
			Description: "年K线数据采集",
			Func: func(ctx context.Context) error {
				return collectAndPersistYearlyKLineData(services, executor, logger)
			},
		},

		// 业绩报表数据采集 - 第六优先级
		&utils.SimpleTask{
			ID:          "performance-reports-collection",
			Description: "业绩报表数据采集",
			Func: func(ctx context.Context) error {
				return collectAndPersistPerformanceReports(services, executor, logger)
			},
		},
	}

	// 并发执行所有任务
	results, stats := executor.ExecuteBatch(ctx, tasks)

	// 记录执行结果
	logger.Infof("每日晚上任务执行完成: 总任务=%d, 成功=%d, 失败=%d, 总耗时=%v",
		stats.TotalTasks, stats.SuccessTasks, stats.FailedTasks,
		stats.EndTime.Sub(stats.StartTime))

	// 记录失败的任务
	for _, result := range results {
		if !result.Success {
			logger.Errorf("任务执行失败: %s - %v", result.TaskID, result.Error)
		}
	}
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

// collectAndPersistStockInfo 采集并保存股票基础信息
func collectAndPersistStockInfo(services *service.Services, executor *utils.ConcurrentExecutor, logger *utils.Logger) error {
	logger.Info("开始采集股票基础信息...")

	// 检查DataService是否已初始化
	if services.DataService == nil {
		return fmt.Errorf("DataService未初始化，请先初始化数据库连接")
	}

	// 使用并发执行器来执行股票同步任务
	ctx := context.Background()
	task := &utils.SimpleTask{
		ID:          "sync-stock-list",
		Description: "同步股票基础信息",
		Func: func(ctx context.Context) error {
			return services.DataService.SyncStockList()
		},
	}

	result := executor.Execute(ctx, task)
	if !result.Success {
		return fmt.Errorf("股票信息同步失败: %v", result.Error)
	}

	logger.Infof("股票基础信息采集完成，耗时: %v", result.Duration)
	return nil
}

// collectAndPersistDailyKLineData 采集并保存日K线数据
func collectAndPersistDailyKLineData(services *service.Services, executor *utils.ConcurrentExecutor, logger *utils.Logger) error {
	logger.Info("开始采集日K线数据...")

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
	inactiveCount := 0

	for _, result := range results {
		if result.Success {
			successCount++
		} else {
			logger.Errorf("股票日K线采集失败: %v", result.Error)
			// 检查是否为非活跃股票错误
			if result.Error != nil && result.Error.Error() == "stock_inactive" {
				inactiveCount++
			}
		}
	}

	logger.Infof("日K线数据采集完成 - 总数: %d, 成功: %d, 失败: %d, 非活跃: %d, 总耗时: %v, 平均耗时: %v",
		stats.TotalTasks, successCount, stats.FailedTasks, inactiveCount, stats.TotalDuration, stats.AverageDuration)

	return nil
}

// syncStockDailyKLine 同步单只股票的日K线数据
func syncStockDailyKLine(services *service.Services, stock *model.Stock, logger *utils.Logger) error {
	// 获取该股票最新的日K线数据
	latestData, err := services.DataService.GetLatestPrice(stock.TsCode)
	if err != nil && err.Error() != "record not found" {
		return fmt.Errorf("获取最新日K线数据失败: %v", err)
	}

	var startDate time.Time

	if latestData == nil {
		// 数据库中没有数据，进行全量同步（获取近2年数据）
		startDate = time.Now().AddDate(-2, 0, 0)
		logger.Infof("股票 %s 进行全量日K线同步，起始日期: %s", stock.TsCode, startDate.Format("2006-01-02"))
	} else {
		// 将TradeDate从int转换为time.Time进行比较
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
			return fmt.Errorf("stock_inactive")
		}
	}

	// 实现真正的数据同步逻辑
	endDate := time.Now()
	logger.Debugf("股票 %s 需要同步日K线数据，时间范围: %s 到 %s",
		stock.TsCode, "1990-01-01", endDate.Format("2006-01-02"))

	// 调用DataService进行数据同步
	syncCount, err := services.DataService.SyncDailyData(stock.TsCode, startDate, endDate)
	if err != nil {
		return fmt.Errorf("同步日K线数据失败: %v", err)
	}

	logger.Debugf("股票 %s 日K线数据同步完成，共同步 %d 条记录", stock.TsCode, syncCount)
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
func collectAndPersistWeeklyKLineData(services *service.Services, executor *utils.ConcurrentExecutor, logger *utils.Logger) error {
	logger.Info("开始采集周K线数据...")

	ctx := context.Background()

	// 从数据库获取所有活跃股票列表
	stocks, err := services.DataService.GetAllStocks()
	if err != nil {
		return fmt.Errorf("获取股票列表失败: %v", err)
	}

	logger.Infof("从数据库获取到 %d 只股票，开始采集周K线数据", len(stocks))

	// 创建任务列表
	tasks := make([]utils.Task, len(stocks))
	for i, stock := range stocks {
		stock := stock // 避免闭包问题
		tasks[i] = &utils.SimpleTask{
			ID:          fmt.Sprintf("weekly_kline_%s", stock.TsCode),
			Description: fmt.Sprintf("采集股票 %s 的周K线数据", stock.TsCode),
			Func: func(ctx context.Context) error {
				return syncStockWeeklyKLine(services, stock, logger)
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
		}
	}

	logger.Infof("周K线数据采集完成 - 总数: %d, 成功: %d, 失败: %d, 总耗时: %v",
		stats.TotalTasks, successCount, stats.FailedTasks, stats.TotalDuration)

	return nil
}

// collectAndPersistMonthlyKLineData 采集并保存月K线数据
func collectAndPersistMonthlyKLineData(services *service.Services, executor *utils.ConcurrentExecutor, logger *utils.Logger) error {
	logger.Info("开始采集月K线数据...")

	ctx := context.Background()
	stocks, err := services.DataService.GetAllStocks()
	if err != nil {
		return fmt.Errorf("获取股票列表失败: %v", err)
	}

	logger.Infof("从数据库获取到 %d 只股票，开始采集月K线数据", len(stocks))

	tasks := make([]utils.Task, len(stocks))
	for i, stock := range stocks {
		stock := stock
		tasks[i] = &utils.SimpleTask{
			ID:          fmt.Sprintf("monthly_kline_%s", stock.TsCode),
			Description: fmt.Sprintf("采集股票 %s 的月K线数据", stock.TsCode),
			Func: func(ctx context.Context) error {
				return syncStockMonthlyKLine(services, stock, logger)
			},
		}
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
func collectAndPersistYearlyKLineData(services *service.Services, executor *utils.ConcurrentExecutor, logger *utils.Logger) error {
	logger.Info("开始采集年K线数据...")

	ctx := context.Background()
	stocks, err := services.DataService.GetAllStocks()
	if err != nil {
		return fmt.Errorf("获取股票列表失败: %v", err)
	}

	logger.Infof("从数据库获取到 %d 只股票，开始采集年K线数据", len(stocks))

	tasks := make([]utils.Task, len(stocks))
	for i, stock := range stocks {
		stock := stock
		tasks[i] = &utils.SimpleTask{
			ID:          fmt.Sprintf("yearly_kline_%s", stock.TsCode),
			Description: fmt.Sprintf("采集股票 %s 的年K线数据", stock.TsCode),
			Func: func(ctx context.Context) error {
				return syncStockYearlyKLine(services, stock, logger)
			},
		}
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
	// 第一步：检查股票是否活跃，不活跃直接跳过
	if !stock.IsActive {
		logger.Debugf("股票 %s 不活跃，跳过周K线数据同步", stock.TsCode)
		return fmt.Errorf("stock_inactive")
	}

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

// isSameWeek 判断两个交易日期是否在同一周
func isSameWeek(tradeDate1, tradeDate2 int) bool {
	date1Str := fmt.Sprintf("%d", tradeDate1)
	date2Str := fmt.Sprintf("%d", tradeDate2)

	date1, err1 := time.Parse("20060102", date1Str)
	date2, err2 := time.Parse("20060102", date2Str)

	if err1 != nil || err2 != nil {
		return false
	}

	// 获取两个日期所在周的周一
	year1, week1 := date1.ISOWeek()
	year2, week2 := date2.ISOWeek()

	return year1 == year2 && week1 == week2
}

// isSameMonth 判断两个交易日期是否在同一月
func isSameMonth(tradeDate1, tradeDate2 int) bool {
	date1Str := fmt.Sprintf("%d", tradeDate1)
	date2Str := fmt.Sprintf("%d", tradeDate2)

	date1, err1 := time.Parse("20060102", date1Str)
	date2, err2 := time.Parse("20060102", date2Str)

	if err1 != nil || err2 != nil {
		return false
	}

	// 比较年份和月份
	return date1.Year() == date2.Year() && date1.Month() == date2.Month()
}

// isSameYear 判断两个交易日期是否在同一年
func isSameYear(tradeDate1, tradeDate2 int) bool {
	date1Str := fmt.Sprintf("%d", tradeDate1)
	date2Str := fmt.Sprintf("%d", tradeDate2)

	date1, err1 := time.Parse("20060102", date1Str)
	date2, err2 := time.Parse("20060102", date2Str)

	if err1 != nil || err2 != nil {
		return false
	}

	// 比较年份
	return date1.Year() == date2.Year()
}

// syncStockMonthlyKLine 同步单只股票的月K线数据
func syncStockMonthlyKLine(services *service.Services, stock *model.Stock, logger *utils.Logger) error {
	// 第一步：检查股票是否活跃，不活跃直接跳过
	if !stock.IsActive {
		logger.Debugf("股票 %s 不活跃，跳过月K线数据同步", stock.TsCode)
		return fmt.Errorf("stock_inactive")
	}

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

// syncStockPerformanceReports 同步单只股票的业绩报表数据
func syncStockPerformanceReports(stock *model.Stock, services *service.Services, logger *utils.Logger) error {
	ctx := context.Background()
	return services.PerformanceService.SyncPerformanceReports(ctx, stock.TsCode)
}

// collectAndPersistPerformanceReports 采集并保存业绩报表数据
func collectAndPersistPerformanceReports(services *service.Services, executor *utils.ConcurrentExecutor, logger *utils.Logger) error {
	logger.Info("开始采集业绩报表数据...")

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

// testStockCollection 测试股票基本数据采集
func testStockCollection(cfg *config.Config, logger *utils.Logger) {
	logger.Info("开始测试股票基本数据采集...")

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

	// 测试股票基本信息采集
	logger.Info("测试股票基本信息采集...")
	start := time.Now()
	if err := collectStockBasicInfo(services, logger); err != nil {
		logger.Errorf("股票基本信息采集失败: %v", err)
		return
	}
	logger.Infof("股票基本信息采集完成，耗时: %v", time.Since(start))
	logger.Info("股票基本数据采集测试完成!")
}

// testKLineCollection 测试所有K线数据采集
func testKLineCollection(cfg *config.Config, logger *utils.Logger) {
	logger.Info("开始测试所有K线数据采集...")

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

	// 获取前5只活跃股票进行测试
	stocks, err := getTestStocks(services, logger, 5)
	if err != nil {
		logger.Errorf("获取测试股票失败: %v", err)
		return
	}

	for _, stock := range stocks {
		logger.Infof("测试股票 %s (%s) 的K线数据采集...", stock.TsCode, stock.Name)

		// 测试日K线
		start := time.Now()
		if err := syncStockDailyKLine(services, stock, logger); err != nil {
			logger.Errorf("股票 %s 日K线采集失败: %v", stock.TsCode, err)
			continue
		}
		logger.Infof("股票 %s 日K线采集完成，耗时: %v", stock.TsCode, time.Since(start))

		// 测试周K线
		start = time.Now()
		if err := syncStockWeeklyKLine(services, stock, logger); err != nil {
			logger.Errorf("股票 %s 周K线采集失败: %v", stock.TsCode, err)
			continue
		}
		logger.Infof("股票 %s 周K线采集完成，耗时: %v", stock.TsCode, time.Since(start))

		// 测试月K线
		start = time.Now()
		if err := syncStockMonthlyKLine(services, stock, logger); err != nil {
			logger.Errorf("股票 %s 月K线采集失败: %v", stock.TsCode, err)
			continue
		}
		logger.Infof("股票 %s 月K线采集完成，耗时: %v", stock.TsCode, time.Since(start))

		// 测试年K线
		start = time.Now()
		if err := syncStockYearlyKLine(services, stock, logger); err != nil {
			logger.Errorf("股票 %s 年K线采集失败: %v", stock.TsCode, err)
			continue
		}
		logger.Infof("股票 %s 年K线采集完成，耗时: %v", stock.TsCode, time.Since(start))
	}

	logger.Info("所有K线数据采集测试完成!")
}

// testDailyKLineCollection 测试日K线数据采集
func testDailyKLineCollection(cfg *config.Config, logger *utils.Logger) {
	logger.Info("开始测试日K线数据采集...")

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

	// 获取前10只活跃股票进行测试
	stocks, err := getTestStocks(services, logger, 10)
	if err != nil {
		logger.Errorf("获取测试股票失败: %v", err)
		return
	}

	logger.Infof("开始采集 %d 只股票的日K线数据...", len(stocks))
	start := time.Now()

	for i, stock := range stocks {
		logger.Infof("采集股票 %s (%s) 日K线数据 [%d/%d]", stock.TsCode, stock.Name, i+1, len(stocks))
		if err := syncStockDailyKLine(services, stock, logger); err != nil {
			logger.Errorf("股票 %s 日K线采集失败: %v", stock.TsCode, err)
		}
	}

	logger.Infof("日K线数据采集测试完成，总耗时: %v", time.Since(start))
}

// testWeeklyKLineCollection 测试周K线数据采集
func testWeeklyKLineCollection(cfg *config.Config, logger *utils.Logger) {
	logger.Info("开始测试周K线数据采集...")

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

	// 获取前10只活跃股票进行测试
	stocks, err := getTestStocks(services, logger, 10)
	if err != nil {
		logger.Errorf("获取测试股票失败: %v", err)
		return
	}

	logger.Infof("开始采集 %d 只股票的周K线数据...", len(stocks))
	start := time.Now()

	for i, stock := range stocks {
		logger.Infof("采集股票 %s (%s) 周K线数据 [%d/%d]", stock.TsCode, stock.Name, i+1, len(stocks))
		if err := syncStockWeeklyKLine(services, stock, logger); err != nil {
			logger.Errorf("股票 %s 周K线采集失败: %v", stock.TsCode, err)
		}
	}

	logger.Infof("周K线数据采集测试完成，总耗时: %v", time.Since(start))
}

// testMonthlyKLineCollection 测试月K线数据采集
func testMonthlyKLineCollection(cfg *config.Config, logger *utils.Logger) {
	logger.Info("开始测试月K线数据采集...")

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

	// 获取前10只活跃股票进行测试
	stocks, err := getTestStocks(services, logger, 10)
	if err != nil {
		logger.Errorf("获取测试股票失败: %v", err)
		return
	}

	logger.Infof("开始采集 %d 只股票的月K线数据...", len(stocks))
	start := time.Now()

	for i, stock := range stocks {
		logger.Infof("采集股票 %s (%s) 月K线数据 [%d/%d]", stock.TsCode, stock.Name, i+1, len(stocks))
		if err := syncStockMonthlyKLine(services, stock, logger); err != nil {
			logger.Errorf("股票 %s 月K线采集失败: %v", stock.TsCode, err)
		}
	}

	logger.Infof("月K线数据采集测试完成，总耗时: %v", time.Since(start))
}

// testYearlyKLineCollection 测试年K线数据采集
func testYearlyKLineCollection(cfg *config.Config, logger *utils.Logger) {
	logger.Info("开始测试年K线数据采集...")

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

	// 获取前10只活跃股票进行测试
	stocks, err := getTestStocks(services, logger, 10)
	if err != nil {
		logger.Errorf("获取测试股票失败: %v", err)
		return
	}

	logger.Infof("开始采集 %d 只股票的年K线数据...", len(stocks))
	start := time.Now()

	for i, stock := range stocks {
		logger.Infof("采集股票 %s (%s) 年K线数据 [%d/%d]", stock.TsCode, stock.Name, i+1, len(stocks))
		if err := syncStockYearlyKLine(services, stock, logger); err != nil {
			logger.Errorf("股票 %s 年K线采集失败: %v", stock.TsCode, err)
		}
	}

	logger.Infof("年K线数据采集测试完成，总耗时: %v", time.Since(start))
}

// getTestStocks 获取用于测试的股票列表
func getTestStocks(services *service.Services, logger *utils.Logger, limit int) ([]*model.Stock, error) {
	stocks, err := services.DataService.GetAllStocks()
	if err != nil {
		return nil, fmt.Errorf("获取股票失败: %v", err)
	}

	if len(stocks) == 0 {
		return nil, fmt.Errorf("没有找到股票")
	}

	// 过滤活跃股票
	var activeStocks []*model.Stock
	for _, stock := range stocks {
		if stock.IsActive {
			activeStocks = append(activeStocks, stock)
		}
	}

	if len(activeStocks) == 0 {
		return nil, fmt.Errorf("没有找到活跃股票")
	}

	// 限制测试股票数量
	if len(activeStocks) > limit {
		activeStocks = activeStocks[:limit]
	}

	logger.Infof("获取到 %d 只测试股票", len(activeStocks))
	return activeStocks, nil
}

// testPerformanceReports 测试业绩报表采集
func testPerformanceReports(services *service.Services, logger *utils.Logger, limit int) error {
	stocks, err := getTestStocks(services, logger, limit)
	if err != nil {
		return err
	}

	logger.Infof("开始采集 %d 只股票的业绩报表...", len(stocks))

	for i, stock := range stocks {
		logger.Infof("采集股票 %s (%s) 业绩报表 [%d/%d]", stock.TsCode, stock.Name, i+1, len(stocks))
		if err := syncStockPerformanceReports(stock, services, logger); err != nil {
			logger.Errorf("股票 %s 业绩报表采集失败: %v", stock.TsCode, err)
		}
	}

	return nil
}
