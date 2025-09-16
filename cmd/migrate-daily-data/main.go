package main

import (
	"fmt"
	"log"

	"stock/internal/config"
	"stock/internal/model"
	"stock/internal/service"
	"stock/internal/utils"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	logger := utils.NewLogger(cfg.Log)

	// 初始化数据库
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Fatalf("Failed to initialize database: %v", err)
	}

	// 创建日K线管理器
	dailyManager := service.NewDailyKLineManager(db, logger)

	// 检查原表是否存在数据
	var originalCount int64
	if err := db.Model(&model.DailyData{}).Count(&originalCount).Error; err != nil {
		logger.Fatalf("Failed to count original data: %v", err)
	}

	if originalCount == 0 {
		logger.Info("No data found in original daily_data table, migration not needed")
		return
	}

	logger.Infof("Found %d records in original daily_data table, starting migration...", originalCount)

	// 分批迁移数据
	batchSize := 1000
	offset := 0
	totalMigrated := 0

	for {
		var batch []model.DailyData
		if err := db.Offset(offset).Limit(batchSize).Find(&batch).Error; err != nil {
			logger.Fatalf("Failed to fetch batch data: %v", err)
		}

		if len(batch) == 0 {
			break
		}

		// 使用DailyKLineManager保存数据到对应的交易所表
		if err := dailyManager.UpsertDailyData(batch); err != nil {
			logger.Errorf("Failed to migrate batch starting at offset %d: %v", offset, err)
			continue
		}

		totalMigrated += len(batch)
		offset += batchSize

		logger.Infof("Migrated %d/%d records (%.2f%%)",
			totalMigrated, originalCount, float64(totalMigrated)/float64(originalCount)*100)
	}

	// 验证迁移结果
	stats, err := dailyManager.GetAllExchangeStats()
	if err != nil {
		logger.Errorf("Failed to get migration stats: %v", err)
	} else {
		logger.Infof("Migration completed successfully: %+v", stats)
	}

	// 询问是否备份原表
	fmt.Print("Migration completed. Do you want to rename the original daily_data table to daily_data_backup? (y/N): ")
	var response string
	fmt.Scanln(&response)

	if response == "y" || response == "Y" {
		// 重命名原表为备份表
		if err := db.Exec("RENAME TABLE daily_data TO daily_data_backup").Error; err != nil {
			logger.Errorf("Failed to rename original table: %v", err)
		} else {
			logger.Info("Original table renamed to daily_data_backup")
		}
	}

	logger.Info("Daily data migration completed successfully!")
}
