package main

import (
	"fmt"
	"log"

	"stock/internal/config"
	"stock/internal/database"
	"stock/internal/utils"
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
	db, err := database.NewDatabase(&cfg.Database, logger)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("修正数据库表结构...")

	// 修改 holder_num_ratio 字段的精度
	sql := "ALTER TABLE shareholder_counts MODIFY COLUMN holder_num_ratio DECIMAL(15,4) DEFAULT 0 COMMENT '股东户数变化比例(%)'"

	if err := db.DB.Exec(sql).Error; err != nil {
		log.Fatalf("修改字段失败: %v", err)
	}

	fmt.Println("✓ holder_num_ratio 字段精度修改完成")

	// 修改其他可能有问题的字段
	modifications := []struct {
		sql         string
		description string
	}{
		{
			"ALTER TABLE shareholder_counts MODIFY COLUMN interval_chrate DECIMAL(15,4) DEFAULT 0 COMMENT '区间涨跌幅(%)'",
			"interval_chrate 字段精度修改",
		},
		{
			"ALTER TABLE shareholder_counts MODIFY COLUMN avg_market_cap DECIMAL(20,2) DEFAULT 0 COMMENT '户均市值(元)'",
			"avg_market_cap 字段精度修改",
		},
		{
			"ALTER TABLE shareholder_counts MODIFY COLUMN avg_hold_num DECIMAL(20,2) DEFAULT 0 COMMENT '户均持股数(股)'",
			"avg_hold_num 字段精度修改",
		},
	}

	for _, mod := range modifications {
		if err := db.DB.Exec(mod.sql).Error; err != nil {
			fmt.Printf("⚠️  %s 失败: %v\n", mod.description, err)
		} else {
			fmt.Printf("✓ %s 完成\n", mod.description)
		}
	}

	fmt.Println("数据库表结构修正完成！")
}
