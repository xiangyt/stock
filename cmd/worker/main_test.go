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

	collectStockBasicInfo(services)
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

	// 从数据库获取所有股票列表
	stocks, err := services.DataService.GetAllStocks()
	if err != nil {
		logger.Errorf("获取股票列表失败: %v", err)
		return
	}

	executor := utils.NewConcurrentExecutor(1000, 45*time.Minute) // 最大100个并发，30分钟超时
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

				// 获取日K线数据
				klineData, err := c.GetDailyKLine(stock.TsCode, time.Time{}, time.Now())
				if err != nil {
					return fmt.Errorf("获取日K线数据失败: %v", err)
				}

				if len(klineData) == 0 {
					logger.Debugf("股票 %s 在指定时间范围内没有日K线数据", stock.TsCode)
					return nil
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
