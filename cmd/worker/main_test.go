package main

import (
	"log"
	"stock/internal/config"
	"stock/internal/database"
	"stock/internal/model"
	"stock/internal/utils"
	"testing"
)

func TestStockCollection(t *testing.T) {
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

	collectStockBasicInfo(services, logger)
}

func TestDailyKLineCollection(t *testing.T) {
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

	collectAndPersistDailyKLineData(services, logger)
}

func TestSyncStockDailyKLine(t *testing.T) {
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

	syncStockDailyKLine(services, &model.Stock{TsCode: "000026.SZ"}, logger)
}
