package main

import (
	"log"
	"net/http"
	"os"

	"stock/internal/api"
	"stock/internal/collector"
	"stock/internal/config"
	"stock/internal/database"
	"stock/internal/model"
	"stock/internal/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化日志
	utilsLogger := utils.NewLogger(config.LogConfig{
		Level:  "info",
		Format: "text",
	})

	// 获取内部的logrus.Logger用于API handler
	logrusLogger := utilsLogger.Logger

	// 初始化数据库连接
	dbConfig := &config.DatabaseConfig{
		Host:     "192.168.1.238",
		Port:     3306,
		User:     "root",
		Password: "123456",
		Name:     "stock",
	}

	dbManager, err := database.NewDatabase(dbConfig, utilsLogger)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	db := dbManager.DB

	// 自动迁移数据库表
	if err := db.AutoMigrate(&model.Stock{}, &model.DailyData{}, &model.TechnicalIndicator{}, &model.Task{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 创建数据采集器
	eastMoneyCollector := collector.NewEastMoneyCollector(utilsLogger)
	if err := eastMoneyCollector.Connect(); err != nil {
		log.Fatalf("Failed to connect to data source: %v", err)
	}

	// 创建采集器管理器
	collectorManager := collector.NewCollectorManager(utilsLogger)
	collectorManager.RegisterCollector("eastmoney", eastMoneyCollector)

	// 创建API处理器（传入数据库连接）
	apiHandler := api.NewHandler(collectorManager, logrusLogger, db)

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建路由
	router := gin.Default()

	// 添加CORS中间件
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// API路由组
	v1 := router.Group("/api/v1")
	{
		// 股票相关接口
		stocks := v1.Group("/stocks")
		{
			stocks.GET("/", apiHandler.GetStockList)                                 // 获取股票列表
			stocks.POST("/sync", apiHandler.SyncAllStocksAsync)                      // 异步同步所有股票数据到数据库
			stocks.GET("/stats", apiHandler.GetStockStats)                           // 获取股票统计信息
			stocks.GET("/search", apiHandler.SearchStocks)                           // 搜索股票
			stocks.GET("/:code", apiHandler.GetStockDetail)                          // 获取股票详情
			stocks.GET("/:code/kline", apiHandler.GetKLineData)                      // 获取K线数据（只查询数据库）
			stocks.POST("/:code/kline/refresh", apiHandler.RefreshKLineData)         // 刷新K线数据（从API获取并保存）
			stocks.GET("/:code/kline/range", apiHandler.GetKLineDataRange)           // 获取K线数据范围
			stocks.GET("/:code/kline/freshness", apiHandler.CheckKLineDataFreshness) // 检查数据新鲜度
			stocks.GET("/:code/performance", apiHandler.GetPerformanceReports)       // 获取业绩报表数据
			stocks.POST("/:code/sync-daily", apiHandler.SyncSingleStockAsync)        // 异步刷新单只股票日K数据
		}

		// 异步任务相关接口
		tasks := v1.Group("/tasks")
		{
			tasks.GET("/", apiHandler.ListTasks)            // 获取任务列表
			tasks.GET("/:taskId", apiHandler.GetTaskStatus) // 获取任务状态
			tasks.DELETE("/:taskId", apiHandler.CancelTask) // 取消任务
		}

		// 实时数据接口
		v1.GET("/realtime", apiHandler.GetRealtimeData) // 获取实时数据
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	utilsLogger.Infof("Starting API server on :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
