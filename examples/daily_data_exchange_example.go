package main

import (
	"fmt"
	"log"
	"time"

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
	manager := service.NewDailyKLineManager(db, logger)

	// 示例1: 测试不同交易所的股票代码识别
	fmt.Println("=== 交易所识别示例 ===")
	testCodes := []string{"600000.SH", "000001.SZ", "300001", "688001"}
	for _, code := range testCodes {
		data := model.DailyData{TsCode: code}
		tableName := data.TableName()
		fmt.Printf("股票代码: %s -> 表名: %s\n", code, tableName)
	}

	// 示例2: 保存日K线数据
	fmt.Println("\n=== 保存数据示例 ===")
	sampleData := []model.DailyData{
		{
			TsCode:    "600000.SH",
			TradeDate: 20240101,
			Open:      10.0,
			High:      11.0,
			Low:       9.5,
			Close:     10.5,
			Volume:    1000000,
			Amount:    10500000.0,
			CreatedAt: time.Now().Unix(),
		},
		{
			TsCode:    "000001.SZ",
			TradeDate: 20240101,
			Open:      15.0,
			High:      16.0,
			Low:       14.5,
			Close:     15.5,
			Volume:    2000000,
			Amount:    31000000.0,
			CreatedAt: time.Now().Unix(),
		},
	}

	err = manager.SaveDailyData(sampleData)
	if err != nil {
		logger.Errorf("保存数据失败: %v", err)
	} else {
		fmt.Printf("成功保存数据: %+v \n", sampleData)
	}

	// 示例3: 查询数据
	fmt.Println("\n=== 查询数据示例 ===")
	for _, data := range sampleData {
		latest, err := manager.GetLatestDailyData(data.TsCode)
		if err != nil {
			logger.Errorf("查询数据失败: %v", err)
			continue
		}
		if latest != nil {
			fmt.Printf("最新数据 %s: 日期=%d, 收盘价=%.2f\n",
				latest.TsCode, latest.TradeDate, latest.Close)
		}
	}

	// 示例4: 获取统计信息
	fmt.Println("\n=== 统计信息示例 ===")
	stats, err := manager.GetAllExchangeStats()
	if err != nil {
		logger.Errorf("获取统计信息失败: %v", err)
	} else {
		fmt.Printf("上海交易所数据量: %v\n", stats["sh_count"])
		fmt.Printf("深圳交易所数据量: %v\n", stats["sz_count"])
		fmt.Printf("总数据量: %v\n", stats["total_count"])
	}

	// 示例5: 按时间范围查询
	fmt.Println("\n=== 时间范围查询示例 ===")
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)

	dataList, err := manager.GetDailyData("600000.SH", startDate, endDate, 10)
	if err != nil {
		logger.Errorf("查询时间范围数据失败: %v", err)
	} else {
		fmt.Printf("查询到 %d 条数据\n", len(dataList))
		for i, item := range dataList {
			if i < 3 { // 只显示前3条
				fmt.Printf("  %s: 日期=%d, 收盘价=%.2f\n",
					item.TsCode, item.TradeDate, item.Close)
			}
		}
	}

	fmt.Println("\n=== 示例完成 ===")
}
