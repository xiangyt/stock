package main

import (
	"fmt"
	"log"
	"os"

	"stock/internal/config"
	"stock/internal/database"
	"stock/internal/model"
	"stock/internal/utils"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "migrate",
		Short: "数据库迁移工具",
		Long:  "用于执行数据库迁移和表结构管理的命令行工具。",
	}

	var upCmd = &cobra.Command{
		Use:   "up",
		Short: "执行数据库迁移",
		Long:  "创建或更新数据库表结构。",
		Run: func(cmd *cobra.Command, args []string) {
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

			fmt.Println("开始执行数据库迁移...")

			// 分步骤迁移，避免外键约束问题
			// 1. 先创建基础表（没有外键的表）
			err = db.DB.AutoMigrate(&model.Stock{})
			if err != nil {
				fmt.Printf("迁移 Stock 表失败: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("✓ Stock 表迁移完成")

			// 2. 创建有外键关系的表
			err = db.DB.AutoMigrate(&model.DailyData{})
			if err != nil {
				fmt.Printf("迁移 DailyData 表失败: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("✓ DailyData 表迁移完成")

			err = db.DB.AutoMigrate(&model.PerformanceReport{})
			if err != nil {
				fmt.Printf("迁移 PerformanceReport 表失败: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("✓ PerformanceReport 表迁移完成")

			err = db.DB.AutoMigrate(&model.ShareholderCount{})
			if err != nil {
				fmt.Printf("迁移 ShareholderCount 表失败: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("✓ ShareholderCount 表迁移完成")

			fmt.Println("数据库迁移完成！")
		},
	}

	rootCmd.AddCommand(upCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("命令执行失败: %v\n", err)
		os.Exit(1)
	}
}
