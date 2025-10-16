package main

import (
	"fmt"
	"log"

	"stock/internal/config"
	"stock/internal/database"
	"stock/internal/logger"
	"stock/internal/model"
	"stock/internal/service"
	"testing"
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
	// 从数据库获取所有股票列表
	stocks, err := services.DataService.GetAllStocks()
	if err != nil {
		return fmt.Errorf("获取股票列表失败: %v", err)
	}

	return collectTodayKLineData(services, stocks)
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

	//syncStockDailyKLine(services, &model.Stock{TsCode: "001208.SZ"})
	//syncStockWeeklyKLine(services, &model.Stock{TsCode: "001208.SZ"})
	syncStockMonthlyKLine(services, &model.Stock{TsCode: "001208.SZ"})
	//syncStockYearlyKLine(services, &model.Stock{TsCode: "001208.SZ"})
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
	// 从数据库获取所有活跃股票列表
	stocks, err := services.DataService.GetAllStocks()
	if err != nil {
		return fmt.Errorf("获取股票列表失败: %v", err)
	}

	return collectThisWeeklyKLineData(services, stocks)
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
	// 从数据库获取所有活跃股票列表
	stocks, err := services.DataService.GetAllStocks()
	if err != nil {
		return fmt.Errorf("获取股票列表失败: %v", err)
	}

	return collectThisMonthlyKLineData(services, stocks)
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
	// 从数据库获取所有活跃股票列表
	stocks, err := services.DataService.GetAllStocks()
	if err != nil {
		return fmt.Errorf("获取股票列表失败: %v", err)
	}

	return collectThisYearlyKLineData(services, stocks)
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
