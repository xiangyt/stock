package main

import (
	"log"
	"stock/internal/config"
	"stock/internal/database"
	"stock/internal/logger"
	"stock/internal/model"
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
}
