package main

import (
	"context"
	"fmt"
	"log"

	"stock/internal/collector"
	"stock/internal/config"
	"stock/internal/database"
	"stock/internal/logger"
	"stock/internal/model"
	"stock/internal/notification"
	"stock/internal/repository"
	"stock/internal/service"
	"stock/internal/utils"
	"testing"
	"time"
)

func TestStockCollection(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

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

	collectAllStockBasicInfo(services)
}

// collectAllStockBasicInfo 采集股票基础信息
func collectAllStockBasicInfo(services *service.Services) error {
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

func TestDailyKLineCollection(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	log := logger.NewLogger(cfg.Log)

	// 初始化数据库连接
	dbManager, err := database.NewDatabase(&cfg.Database, log)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	db := dbManager.DB

	// 初始化服务
	services, err := initServicesWithDB(cfg, db)
	if err != nil {
		logger.Fatalf("Failed to initialize services: %v", err)
	}

	collectAndPersistDailyKLineData(services)
}

// collectAndPersistDailyKLineData 采集并保存日K线数据
func collectAndPersistDailyKLineData(services *service.Services) error {
	logger.Info("开始采集日K线数据...")

	executor := utils.NewConcurrentExecutor(maxConcurrent, 45*time.Minute) // 最大100个并发，30分钟超时
	defer executor.Close()
	ctx := context.Background()

	// 从数据库获取所有股票列表
	stocks, err := services.DataService.GetAllStocks()
	if err != nil {
		return fmt.Errorf("获取股票列表失败: %v", err)
	}

	logger.Infof("从数据库获取到 %d 只股票，开始采集日K线数据", len(stocks))

	// 创建任务列表
	tasks := make([]utils.Task, 0, len(stocks))
	for _, stock := range stocks {
		stock := stock // 避免闭包问题
		if !stock.IsActive {
			logger.Debugf("股票 %s 不活跃，跳过日K线数据同步", stock.TsCode)
			continue
		}
		tasks = append(tasks, &utils.SimpleTask{
			ID:          fmt.Sprintf("daily_kline_%s", stock.TsCode),
			Description: fmt.Sprintf("采集股票 %s 的日K线数据", stock.TsCode),
			Func: func(ctx context.Context) error {
				return syncStockDailyKLine(services, stock)
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
			logger.Errorf("股票日K线采集失败: %v", result.Error)
		}
	}

	completionMsg := fmt.Sprintf("📊 日K线数据采集完成\n总数: %d\n成功: %d\n失败: %d\n总耗时: %v\n平均耗时: %v",
		stats.TotalTasks, successCount, stats.FailedTasks, stats.EndTime.Sub(stats.StartTime), stats.AverageDuration)

	logger.Infof("日K线数据采集完成 - 总数: %d, 成功: %d, 失败: %d, 总耗时: %v, 平均耗时: %v",
		stats.TotalTasks, successCount, stats.FailedTasks, stats.EndTime.Sub(stats.StartTime), stats.AverageDuration)

	// 同步日志信息给机器人
	services.NotifyManger.SendToAllBots(context.Background(), &notification.Message{
		Content: completionMsg,
		MsgType: notification.MessageTypeText,
	})
	return nil
}

func TestDailyKLineCollection1(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	log := logger.NewLogger(cfg.Log)

	// 初始化数据库连接
	dbManager, err := database.NewDatabase(&cfg.Database, log)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	db := dbManager.DB

	// 初始化服务
	services, err := initServicesWithDB(cfg, db)
	if err != nil {
		logger.Fatalf("Failed to initialize services: %v", err)
		return
	}

	dd := repository.NewDailyData(db)
	// 从数据库获取所有股票列表
	stocks, err := services.DataService.GetAllStocks()
	if err != nil {
		logger.Errorf("获取股票列表失败: %v", err)
		return
	}
	//stocks = []*model.Stock{
	//	{
	//		TsCode:    "001208.SZ",
	//		Symbol:    "001208",
	//		Name:      "华菱线缆",
	//		IsActive:  true,
	//		CreatedAt: time.Now(),
	//		UpdatedAt: time.Now(),
	//	},
	//}
	executor := utils.NewConcurrentExecutor(100, 45*time.Minute) // 最大100个并发，30分钟超时
	defer executor.Close()
	ctx := context.Background()

	logger.Infof("从数据库获取到 %d 只股票，开始采集日K线数据", len(stocks))
	c := collector.GetCollectorFactory(log).GetTongHuaShunCollector()
	// 创建任务列表
	tasks := make([]utils.Task, 0, len(stocks))
	for _, stock := range stocks {
		stock := stock // 避免闭包问题
		if !stock.IsActive {
			logger.Debugf("股票 %s 不活跃，跳过日K线数据同步", stock.TsCode)
			continue
		}
		tasks = append(tasks, &utils.SimpleTask{
			ID:          fmt.Sprintf("daily_kline_%s", stock.TsCode),
			Description: fmt.Sprintf("采集股票 %s 的日K线数据", stock.TsCode),
			Func: func(ctx context.Context) error {

				st := time.Date(1990, 1, 1, 0, 0, 0, 0, time.Local)
				// 获取日K线数据
				klineData, err := c.GetDailyKLine(stock.TsCode, st, time.Now())
				if err != nil {
					return fmt.Errorf("获取日K线数据失败: %v", err)
				}

				if len(klineData) == 0 {
					logger.Debugf("股票 %s 在指定时间范围内没有日K线数据", stock.TsCode)
					return nil
				}

				// 批量保存数据
				if err := dd.UpsertDailyData(klineData); err != nil {
					return fmt.Errorf("保存日K线数据失败: %v", err)
				}
				return nil
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
			logger.Errorf("股票日K线采集失败: %v", result.Error)
		}
	}
	logger.Infof("日K线数据采集完成 - 总数: %d, 成功: %d, 失败: %d, 总耗时: %v, 平均耗时: %v",
		stats.TotalTasks, successCount, stats.FailedTasks, stats.EndTime.Sub(stats.StartTime), stats.AverageDuration)

}

func TestSyncStockDailyKLine(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

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

	syncStockDailyKLine(services, &model.Stock{TsCode: "000026.SZ"})
}

func TestWeeklyKLineCollection(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

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

	collectAndPersistWeeklyKLineData(services)
	//syncStockYearlyKLine(services, &model.Stock{TsCode: "001208.SZ"})
}

// collectAndPersistWeeklyKLineData 采集并保存周K线数据
func collectAndPersistWeeklyKLineData(services *service.Services) error {
	logger.Info("开始采集周K线数据...")
	executor := utils.NewConcurrentExecutor(maxConcurrent, 45*time.Minute) // 最大100个并发，30分钟超时
	defer executor.Close()
	ctx := context.Background()

	// 从数据库获取所有活跃股票列表
	stocks, err := services.DataService.GetAllStocks()
	if err != nil {
		return fmt.Errorf("获取股票列表失败: %v", err)
	}

	logger.Infof("从数据库获取到 %d 只股票，开始采集周K线数据", len(stocks))

	// 创建任务列表
	tasks := make([]utils.Task, 0, len(stocks))
	for _, stock := range stocks {
		stock := stock // 避免闭包问题
		if !stock.IsActive {
			logger.Debugf("股票 %s 不活跃，跳过周K线数据同步", stock.TsCode)
			continue
		}
		tasks = append(tasks, &utils.SimpleTask{
			ID:          fmt.Sprintf("weekly_kline_%s", stock.TsCode),
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
		}
	}

	weeklyMsg := fmt.Sprintf("📊 周K线数据采集完成\n总数: %d\n成功: %d\n失败: %d\n总耗时: %v",
		stats.TotalTasks, successCount, stats.FailedTasks, stats.EndTime.Sub(stats.StartTime))

	logger.Infof("周K线数据采集完成 - 总数: %d, 成功: %d, 失败: %d, 总耗时: %v",
		stats.TotalTasks, successCount, stats.FailedTasks, stats.EndTime.Sub(stats.StartTime))

	// 同步日志信息给机器人
	services.NotifyManger.SendToAllBots(context.Background(), &notification.Message{
		Content: weeklyMsg,
		MsgType: notification.MessageTypeText,
	})

	return nil
}

func TestMonthlyKLineCollection(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

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

	collectAndPersistMonthlyKLineData(services)
	//syncStockYearlyKLine(services, &model.Stock{TsCode: "001208.SZ"})

}

// collectAndPersistMonthlyKLineData 采集并保存月K线数据
func collectAndPersistMonthlyKLineData(services *service.Services) error {
	logger.Info("开始采集月K线数据...")
	executor := utils.NewConcurrentExecutor(maxConcurrent, 45*time.Minute) // 最大100个并发，30分钟超时
	defer executor.Close()
	ctx := context.Background()
	stocks, err := services.DataService.GetAllStocks()
	if err != nil {
		return fmt.Errorf("获取股票列表失败: %v", err)
	}

	logger.Infof("从数据库获取到 %d 只股票，开始采集月K线数据", len(stocks))

	tasks := make([]utils.Task, 0, len(stocks))
	for _, stock := range stocks {
		stock := stock
		if !stock.IsActive {
			logger.Debugf("股票 %s 不活跃，跳过月K线数据同步", stock.TsCode)
			continue
		}
		tasks = append(tasks, &utils.SimpleTask{
			ID:          fmt.Sprintf("monthly_kline_%s", stock.TsCode),
			Description: fmt.Sprintf("采集股票 %s 的月K线数据", stock.TsCode),
			Func: func(ctx context.Context) error {
				return syncStockMonthlyKLine(services, stock)
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

	monthlyMsg := fmt.Sprintf("📊 月K线数据采集完成\n总数: %d\n成功: %d\n失败: %d\n总耗时: %v",
		stats.TotalTasks, successCount, stats.FailedTasks, stats.EndTime.Sub(stats.StartTime))

	logger.Infof("月K线数据采集完成 - 总数: %d, 成功: %d, 失败: %d, 总耗时: %v",
		stats.TotalTasks, successCount, stats.FailedTasks, stats.EndTime.Sub(stats.StartTime))

	// 同步日志信息给机器人
	services.NotifyManger.SendToAllBots(context.Background(), &notification.Message{
		Content: monthlyMsg,
		MsgType: notification.MessageTypeText,
	})

	return nil
}

func TestYearlyKLineCollection(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

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

	collectAndPersistYearlyKLineData(services)
	//syncStockYearlyKLine(services, &model.Stock{TsCode: "001208.SZ"})

}

// collectAndPersistYearlyKLineData 采集并保存年K线数据
func collectAndPersistYearlyKLineData(services *service.Services) error {
	logger.Info("开始采集年K线数据...")
	executor := utils.NewConcurrentExecutor(maxConcurrent, 45*time.Minute) // 最大100个并发，30分钟超时
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
				return syncStockYearlyKLine(services, stock)
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
		stats.TotalTasks, successCount, stats.FailedTasks, stats.EndTime.Sub(stats.StartTime))
	monthlyMsg := fmt.Sprintf("📊 年K线数据采集完成\n总数: %d\n成功: %d\n失败: %d\n总耗时: %v",
		stats.TotalTasks, successCount, stats.FailedTasks, stats.EndTime.Sub(stats.StartTime))
	// 同步日志信息给机器人
	services.NotifyManger.SendToAllBots(context.Background(), &notification.Message{
		Content: monthlyMsg,
		MsgType: notification.MessageTypeText,
	})
	return nil
}

func TestPerformanceReportsCollection(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

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

	collectAndPersistPerformanceReports(services)
	//syncStockYearlyKLine(services, &model.Stock{TsCode: "001208.SZ"})

}

func TestShareholderCountsCollection(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

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

	collectAndPersistShareholderCounts(services)
	//syncStockYearlyKLine(services, &model.Stock{TsCode: "001208.SZ"})

}

func TestCalculateKDJ(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

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

	err = services.IndicatorService.CalculateKDJByPeriod(model.Stock{
		TsCode: "001208.SZ",
		Symbol: "001208",
	}, model.TechnicalIndicatorPeriodDaily)
	//syncStockYearlyKLine(services, &model.Stock{TsCode: "001208.SZ"})
	if err != nil {
		t.Log(err)
	}
}
