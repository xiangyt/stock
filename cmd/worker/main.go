package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"stock/internal/collector"
	"stock/internal/config"
	"stock/internal/database"
	"stock/internal/logger"
	"stock/internal/model"
	"stock/internal/notification"
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
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 初始化全局日志器
	logger.InitGlobalLogger(cfg.Log)

	// 初始化数据库连接
	dbManager, err := database.NewDatabase(&cfg.Database, logger.GetGlobalLogger())
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	db := dbManager.DB

	// 初始化服务
	services, err := initServicesWithDB(cfg, db)
	if err != nil {
		logger.Fatalf("Failed to initialize services: %v", err)
	}

	logger.Info("Worker starting...")

	// 创建定时任务调度器
	c := cron.New(cron.WithSeconds())

	// 设置定时任务
	setupCronJobs(c, services)

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

const maxConcurrent = 100 // 最大并发量
func setupCronJobs(c *cron.Cron, services *service.Services) {

	c.AddFunc("0 10 16 * * *", func() {
		if !isWorkingDay(time.Now()) {
			return
		}
		// 除权、退市股票处理 - 第一优先级
		_ = collectSkipStock(services)
	})

	c.AddFunc("0 10 17 * * *", func() {
		if !work {
			return
		}
		// 从数据库获取所有活跃股票列表
		stocks, err := services.DataService.GetAllStocks()
		if err != nil {
			return
		}
		var list = make([]*model.Stock, 0, len(stocks))
		for _, stock := range stocks {
			if !strings.HasPrefix(stock.Name, "XD") {
				list = append(list, stock)
			}
		}
		// 更新本周K线数据
		_ = collectThisWeekKLineData(services, list)
		// 更新本月K线数据
		_ = collectThisMonthlyKLineData(services, list)
		// 更新本年K线数据
		_ = collectThisYearlyKLineData(services, list)
	})

	c.AddFunc("0 10 22 * * *", func() {
		if !work {
			return
		}
		_ = collectAndPersistPerformanceReports(services)
	})

	c.AddFunc("0 10 23 * * *", func() {
		if !work {
			return
		}
		_ = collectAndPersistShareholderCounts(services)
	})

	logger.Info("定时任务配置完成！")
}

// initServicesWithDB 初始化带数据库连接的服务
func initServicesWithDB(cfg *config.Config, db *gorm.DB) (*service.Services, error) {
	logger.Info("初始化服务...")

	// 创建基础服务
	services, err := service.NewServices(cfg, logger.GetGlobalLogger())
	if err != nil {
		return nil, fmt.Errorf("创建基础服务失败: %v", err)
	}

	// 初始化需要数据库连接的服务
	services.DataService = service.GetDataService(db, logger.GetGlobalLogger())

	// 为PerformanceService创建必要的依赖
	performanceRepo := repository.NewPerformance(db)
	stockRepo := repository.NewStock(db)
	eastMoneyCollector := collector.GetCollectorFactory(logger.GetGlobalLogger()).GetEastMoneyCollector()
	services.PerformanceService = service.NewPerformanceService(performanceRepo, stockRepo, eastMoneyCollector)

	// 为ShareholderService创建必要的依赖
	shareholderRepo := repository.NewShareholder(db)
	services.ShareholderService = service.NewShareholderService(shareholderRepo, eastMoneyCollector)

	services.IndicatorService = service.GetIndicatorService(db)

	logger.Info("所有服务初始化完成")
	return services, nil
}

var work = true // 今天是否工作日

// collectSkipStock 跳过异常股票 XD PT等
func collectSkipStock(services *service.Services) error {
	logger.Info("开始采集股票基础信息...")
	var err error
	var stocks []*model.Stock
	var xd, pt int
	defer func() {
		completionMsg := fmt.Sprintf("📋 同步股票基本信息完成\n总数: %d\n除权数量: %d\n退市数量: %d",
			len(stocks), xd, pt)
		if err != nil {
			completionMsg = fmt.Sprintf("📋 同步股票基本信息失败，err:%s", err.Error())
		}
		// 同步日志信息给机器人
		services.NotifyManger.SendToAllBots(context.Background(), &notification.Message{
			Content: completionMsg,
			MsgType: notification.MessageTypeText,
		})
	}()
	// 检查DataService是否已初始化
	if services.DataService == nil {
		return fmt.Errorf("DataService未初始化，请先初始化数据库连接")
	}

	// 同步股票基础信息
	stocks, work, err = services.DataService.SyncStockActiveList()
	if err != nil {
		return fmt.Errorf("股票信息同步失败: %v", err)
	}

	executor := utils.NewConcurrentExecutor(maxConcurrent, 45*time.Minute) // 最大100个并发，30分钟超时
	defer executor.Close()
	ctx := context.Background()

	// 创建任务列表
	tasks := make([]utils.Task, 0, len(stocks))
	for _, stock := range stocks {
		stock := stock
		if strings.HasPrefix(stock.Name, "XD") {
			xd++
			tasks = append(tasks, &utils.SimpleTask{
				ID:          fmt.Sprintf("daily_kline_%s", stock.TsCode),
				Description: fmt.Sprintf("采集股票 %s 的日K线数据", stock.TsCode),
				Func: func(ctx context.Context) error {
					_ = syncStockDailyKLine(services, stock)
					_ = syncStockWeeklyKLine(services, stock)
					_ = syncStockMonthlyKLine(services, stock)
					_ = syncStockYearlyKLine(services, stock)
					return nil
				},
			})
		}
		if !stock.IsActive {
			pt++
		}
	}
	// 执行任务
	_, _ = executor.ExecuteBatch(ctx, tasks)

	logger.Info("股票基础信息采集完成")
	return nil
}

// collectThisWeekKLineData 更新本周K线数据
func collectThisWeekKLineData(services *service.Services, stocks []*model.Stock) error {
	logger.Info("开始更新本周K线数据...")

	executor := utils.NewConcurrentExecutor(maxConcurrent, 45*time.Minute) // 最大100个并发，30分钟超时
	defer executor.Close()
	ctx := context.Background()

	// 创建任务列表
	tasks := make([]utils.Task, 0, len(stocks))
	for _, stock := range stocks {
		stock := stock // 避免闭包问题
		tasks = append(tasks, &utils.SimpleTask{
			ID:          fmt.Sprintf("daily_kline_%s", stock.TsCode),
			Description: fmt.Sprintf("采集股票 %s 的周K线数据", stock.TsCode),
			Func: func(ctx context.Context) error {
				return syncStockWeeklyKLine(services, stock)
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
		} else {
			logger.Errorf("股票周K线采集失败: %v", result.Error)
		}
	}

	completionMsg := fmt.Sprintf("📊 周K线数据采集完成\n总数: %d\n成功: %d\n失败: %d\n总耗时: %v\n平均耗时: %v",
		stats.TotalTasks, successCount, stats.FailedTasks, stats.EndTime.Sub(stats.StartTime), stats.AverageDuration)

	logger.Infof("周K线数据采集完成 - 总数: %d, 成功: %d, 失败: %d, 总耗时: %v, 平均耗时: %v",
		stats.TotalTasks, successCount, stats.FailedTasks, stats.EndTime.Sub(stats.StartTime), stats.AverageDuration)

	// 同步日志信息给机器人
	services.NotifyManger.SendToAllBots(context.Background(), &notification.Message{
		Content: completionMsg,
		MsgType: notification.MessageTypeText,
	})
	return nil
}

// collectThisMonthlyKLineData 更新本月K线数据
func collectThisMonthlyKLineData(services *service.Services, stocks []*model.Stock) error {
	logger.Info("开始更新本月K线数据...")

	executor := utils.NewConcurrentExecutor(maxConcurrent, 45*time.Minute) // 最大100个并发，30分钟超时
	defer executor.Close()
	ctx := context.Background()

	// 创建任务列表
	tasks := make([]utils.Task, 0, len(stocks))
	for _, stock := range stocks {
		stock := stock // 避免闭包问题
		tasks = append(tasks, &utils.SimpleTask{
			ID:          fmt.Sprintf("daily_kline_%s", stock.TsCode),
			Description: fmt.Sprintf("采集股票 %s 的月K线数据", stock.TsCode),
			Func: func(ctx context.Context) error {
				return syncStockMonthlyKLine(services, stock)
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
		} else {
			logger.Errorf("股票月K线采集失败: %v", result.Error)
		}
	}

	completionMsg := fmt.Sprintf("📊 月K线数据采集完成\n总数: %d\n成功: %d\n失败: %d\n总耗时: %v\n平均耗时: %v",
		stats.TotalTasks, successCount, stats.FailedTasks, stats.EndTime.Sub(stats.StartTime), stats.AverageDuration)

	logger.Infof("周K线数据采集完成 - 总数: %d, 成功: %d, 失败: %d, 总耗时: %v, 平均耗时: %v",
		stats.TotalTasks, successCount, stats.FailedTasks, stats.EndTime.Sub(stats.StartTime), stats.AverageDuration)

	// 同步日志信息给机器人
	services.NotifyManger.SendToAllBots(context.Background(), &notification.Message{
		Content: completionMsg,
		MsgType: notification.MessageTypeText,
	})
	return nil
}

// collectThisYearlyKLineData 更新本年K线数据
func collectThisYearlyKLineData(services *service.Services, stocks []*model.Stock) error {
	logger.Info("开始更新本年K线数据...")

	executor := utils.NewConcurrentExecutor(maxConcurrent, 45*time.Minute) // 最大100个并发，30分钟超时
	defer executor.Close()
	ctx := context.Background()

	// 创建任务列表
	tasks := make([]utils.Task, 0, len(stocks))
	for _, stock := range stocks {
		stock := stock // 避免闭包问题
		tasks = append(tasks, &utils.SimpleTask{
			ID:          fmt.Sprintf("daily_kline_%s", stock.TsCode),
			Description: fmt.Sprintf("采集股票 %s 的年K线数据", stock.TsCode),
			Func: func(ctx context.Context) error {
				return syncStockYearlyKLine(services, stock)
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
		} else {
			logger.Errorf("股票年K线采集失败: %v", result.Error)
		}
	}

	completionMsg := fmt.Sprintf("📊 年K线数据采集完成\n总数: %d\n成功: %d\n失败: %d\n总耗时: %v\n平均耗时: %v",
		stats.TotalTasks, successCount, stats.FailedTasks, stats.EndTime.Sub(stats.StartTime), stats.AverageDuration)

	logger.Infof("年K线数据采集完成 - 总数: %d, 成功: %d, 失败: %d, 总耗时: %v, 平均耗时: %v",
		stats.TotalTasks, successCount, stats.FailedTasks, stats.EndTime.Sub(stats.StartTime), stats.AverageDuration)

	// 同步日志信息给机器人
	services.NotifyManger.SendToAllBots(context.Background(), &notification.Message{
		Content: completionMsg,
		MsgType: notification.MessageTypeText,
	})
	return nil
}

// markStockInactive 标记股票为非活跃状态
func markStockInactive(services *service.Services, tsCode string) error {
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

// syncStockDailyKLine 同步单只股票的日K线数据
func syncStockDailyKLine(services *service.Services, stock *model.Stock) error {
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
		if time.Now().Format("20060102") == tradeDateStr {
			if latestData.UpdatedAt.Format(time.TimeOnly) > "16:00:00" { // 今日收盘后已经更新过一次，无需再更新
				return nil
			}
			return updateStockTodayKLine(services, stock)
		}
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
			if err := markStockInactive(services, stock.TsCode); err != nil {
				logger.Errorf("标记股票 %s 为非活跃状态失败: %v", stock.TsCode, err)
			}
		}
	}
	return nil
}

// syncStockWeeklyKLine 同步单只股票的周K线数据
func syncStockWeeklyKLine(services *service.Services, stock *model.Stock) error {

	// 第一步：查出该股票最新的一条周K线数据
	klinePersistence := service.GetKLinePersistenceService(services.DataService.GetDB(), logger.GetGlobalLogger())
	latestWeeklyData, err := klinePersistence.GetLatestWeeklyData(stock.TsCode)
	if err != nil {
		return fmt.Errorf("获取最新周K线数据失败: %v", err)
	}

	// 第二步：确定采集的起始时间
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

		if IsSameISOWeek(tradeDate, time.Now()) {
			return updateStockThisWeekKLine(services, stock)
		}

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
func syncStockMonthlyKLine(services *service.Services, stock *model.Stock) error {

	// 第二步：查出该股票最新的一条月K线数据
	klinePersistence := service.GetKLinePersistenceService(services.DataService.GetDB(), logger.GetGlobalLogger())
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
		if isLastTradeDate(latestMonthlyData.TradeDate) { // 今天的数据已经固化成功
			return nil
		}
		// 删除最新的月K线数据
		if err := klinePersistence.DeleteData(stock.TsCode, tradeDate, "monthly"); err != nil {
			logger.Errorf("删除最新月K线数据失败: %v", err)
			return fmt.Errorf("删除最新月K线数据失败: %v", err)
		}
		logger.Debugf("已删除股票 %s 最新的月K线数据，交易日期: %d", stock.TsCode, latestMonthlyData.TradeDate)

		if IsSameMonth(tradeDate, time.Now()) {
			return updateStockThisMonthKLine(services, stock)
		}
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
func syncStockYearlyKLine(services *service.Services, stock *model.Stock) error {
	// 第一步：查出该股票最新的一条年K线数据
	klinePersistence := service.GetKLinePersistenceService(services.DataService.GetDB(), logger.GetGlobalLogger())
	latestYearlyData, err := klinePersistence.GetLatestYearlyData(stock.TsCode)
	if err != nil {
		return fmt.Errorf("获取最新年K线数据失败: %v", err)
	}

	// 第二步：确定采集的起始时间
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

		if isLastTradeDate(latestYearlyData.TradeDate) { // 今天的数据已经固化成功
			return nil
		}
		// 删除最新的年K线数据
		if err := klinePersistence.DeleteData(stock.TsCode, tradeDate, "yearly"); err != nil {
			logger.Errorf("删除最新年K线数据失败: %v", err)
			return fmt.Errorf("删除最新年K线数据失败: %v", err)
		}
		logger.Debugf("已删除股票 %s 最新的年K线数据，交易日期: %d", stock.TsCode, latestYearlyData.TradeDate)

		if IsSameYear(tradeDate, time.Now()) {
			return updateStockThisYearKLine(services, stock)
		}
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

// updateStockTodayKLine 更新单只股票当日K线数据
func updateStockTodayKLine(services *service.Services, stock *model.Stock) error {
	c, err := collector.GetCollectorFactory(logger.GetGlobalLogger()).CreateCollector(collector.CollectorTypeTongHuaShun)
	if err != nil {
		return err
	}

	today, _, err := c.GetTodayData(stock.TsCode)
	if err != nil {
		return err
	}

	return services.DataService.UpsertKLineData([]model.DailyData{*today})
}

// updateStockWeeklyKLine 更新单只股票本周K线数据
func updateStockThisWeekKLine(services *service.Services, stock *model.Stock) error {
	c, err := collector.GetCollectorFactory(logger.GetGlobalLogger()).CreateCollector(collector.CollectorTypeTongHuaShun)
	if err != nil {
		return err
	}

	today, err := c.GetThisWeekData(stock.TsCode)
	if err != nil {
		return err
	}

	return services.DataService.UpsertKLineData([]model.WeeklyData{*today})
}

// updateStockMonthlyKLine 更新单只股票本月K线数据
func updateStockThisMonthKLine(services *service.Services, stock *model.Stock) error {
	c, err := collector.GetCollectorFactory(logger.GetGlobalLogger()).CreateCollector(collector.CollectorTypeTongHuaShun)
	if err != nil {
		return err
	}

	today, err := c.GetThisMonthData(stock.TsCode)
	if err != nil {
		return err
	}

	return services.DataService.UpsertKLineData([]model.MonthlyData{*today})
}

// updateStockThisYearKLine 更新单只股票本年K线数据
func updateStockThisYearKLine(services *service.Services, stock *model.Stock) error {
	c, err := collector.GetCollectorFactory(logger.GetGlobalLogger()).CreateCollector(collector.CollectorTypeTongHuaShun)
	if err != nil {
		return err
	}

	today, err := c.GetThisYearData(stock.TsCode)
	if err != nil {
		return err
	}

	return services.DataService.UpsertKLineData([]model.YearlyData{*today})
}

// collectAndPersistPerformanceReports 采集并保存业绩报表数据
func collectAndPersistPerformanceReports(services *service.Services) error {
	logger.Info("开始采集业绩报表数据...")
	executor := utils.NewConcurrentExecutor(maxConcurrent, 30*time.Minute) // 最大100个并发，30分钟超时
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
	// 一个月前
	date := time.Now().AddDate(0, -1, 0)
	for _, stock := range stocks {
		report, err := services.PerformanceService.GetLatestPerformanceReport(ctx, stock.TsCode)
		if err != nil {
			continue
		}
		if report != nil && report.UpdatedAt.After(date) { // 一个月内更新过，直接跳过
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
		if len(tasks) >= 100 { // 一天只更新200条，防止封ip
			break
		}
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

	performanceMsg := fmt.Sprintf("📈 业绩报表采集完成\n总数: %d\n成功: %d\n失败: %d\n总耗时: %v\n平均耗时: %v",
		stats.TotalTasks, successCount, stats.FailedTasks, stats.EndTime.Sub(stats.StartTime), stats.AverageDuration)

	logger.Infof("业绩报表数据采集完成 - 总数: %d, 成功: %d, 失败: %d, 同步报表: %d, 总耗时: %v, 平均耗时: %v",
		stats.TotalTasks, successCount, stats.FailedTasks, totalReports, stats.EndTime.Sub(stats.StartTime), stats.AverageDuration)

	// 同步日志信息给机器人
	services.NotifyManger.SendToAllBots(context.Background(), &notification.Message{
		Content: performanceMsg,
		MsgType: notification.MessageTypeText,
	})

	return nil
}

// collectAndPersistShareholderCounts 采集并保存股东人数数据
func collectAndPersistShareholderCounts(services *service.Services) error {
	logger.Info("开始采集股东人数数据...")

	executor := utils.NewConcurrentExecutor(maxConcurrent, 45*time.Minute) // 最大50个并发，45分钟超时
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
	date := time.Now().AddDate(0, 0, -7)
	for _, stock := range stocks {
		count, err := services.ShareholderService.GetLatestShareholderCount(stock.TsCode)
		if err != nil {
			continue
		}
		if count != nil && count.UpdatedAt.After(date) { // 7天内更新过，直接跳过
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
		if len(tasks) >= 100 { // 一天只更新100条，防止封ip
			break
		}
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

	shareholderMsg := fmt.Sprintf("👥 股东人数采集完成\n总数: %d\n成功: %d\n失败: %d\n总耗时: %v\n平均耗时: %v",
		stats.TotalTasks, successCount, stats.FailedTasks, stats.EndTime.Sub(stats.StartTime), stats.AverageDuration)

	logger.Infof("股东人数数据采集完成 - 总数: %d, 成功: %d, 失败: %d, 同步股票: %d, 总耗时: %v, 平均耗时: %v",
		stats.TotalTasks, successCount, stats.FailedTasks, totalCounts, stats.EndTime.Sub(stats.StartTime),
		stats.AverageDuration)

	// 同步日志信息给机器人
	services.NotifyManger.SendToAllBots(context.Background(), &notification.Message{
		Content: shareholderMsg,
		MsgType: notification.MessageTypeText,
	})

	return nil
}

// isLastTradeDate 是否为最近一个交易日
func isLastTradeDate(tradeDate int) bool {
	// 将输入的交易日期转换为time.Time
	tradeDateStr := fmt.Sprintf("%d", tradeDate)
	inputDate, err := time.Parse("20060102", tradeDateStr)
	if err != nil {
		logger.Errorf("解析交易日期失败: %v", err)
		return false
	}

	// 获取当前时间
	now := time.Now()

	// 如果输入日期是未来日期，返回false
	if inputDate.After(now) {
		return false
	}

	// 从今天开始往前找最近的交易日
	currentDate := now
	for {
		// 检查当前日期是否为交易日（周一到周五，排除节假日）
		if isWorkingDay(currentDate) {
			// 找到最近的交易日，比较是否与输入日期相同
			lastTradeDate := currentDate.Year()*10000 + int(currentDate.Month())*100 + currentDate.Day()
			return tradeDate == lastTradeDate
		}
		// 往前推一天
		currentDate = currentDate.AddDate(0, 0, -1)

		// 防止无限循环，最多往前找30天
		if now.Sub(currentDate).Hours() > 24*30 {
			break
		}
	}

	return false
}

// isWorkingDay 判断是否为工作日（周一到周五，简化版本，不考虑部分节假日）
func isWorkingDay(date time.Time) bool {
	weekday := date.Weekday()
	// 周一到周五为工作日
	if !(weekday >= time.Monday && weekday <= time.Friday) {
		return false
	}

	if date.Month() == time.October && date.Day() <= 7 { // 国庆
		return false
	}

	if date.Month() == time.May && date.Day() <= 4 { // 五一
		return false
	}

	return true
}

// IsSameISOWeek 判断两个时间是否在同一ISO周（周一为周开始）
func IsSameISOWeek(t1, t2 time.Time) bool {
	y1, w1 := t1.ISOWeek()
	y2, w2 := t2.ISOWeek()
	return y1 == y2 && w1 == w2
}

// IsSameMonth 判断两个时间是否在同一月
func IsSameMonth(t1, t2 time.Time) bool {
	return t1.Year() == t2.Year() && t1.Month() == t2.Month()
}

// IsSameYear 判断两个时间是否在同一年
func IsSameYear(t1, t2 time.Time) bool {
	return t1.Year() == t2.Year()
}
