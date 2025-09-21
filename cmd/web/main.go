package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"stock/internal/api"
	"stock/internal/collector"
	"stock/internal/config"
	"stock/internal/database"
	"stock/internal/logger"
	"stock/internal/model"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	utilsLogger := logger.NewLogger(cfg.Log)

	// 获取内部的logrus.Logger用于API handler
	logrusLogger := utilsLogger.Logger

	// 初始化数据库连接
	dbManager, err := database.NewDatabase(&cfg.Database, utilsLogger)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	db := dbManager.DB

	// 自动迁移数据库表
	if err := db.AutoMigrate(&model.Stock{}, &model.DailyData{}, &model.PerformanceReport{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 创建数据采集器
	eastMoneyCollector := collector.GetCollectorFactory(logger.GetGlobalLogger()).GetEastMoneyCollector()
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
			stocks.GET("/", apiHandler.GetStockList)                           // 获取股票列表
			stocks.GET("/:code", apiHandler.GetStockDetail)                    // 获取股票详情
			stocks.GET("/:code/kline", apiHandler.GetKLineData)                // 获取K线数据
			stocks.GET("/:code/performance", apiHandler.GetPerformanceReports) // 获取业绩报表数据
		}

		// 实时数据接口
		v1.GET("/realtime", apiHandler.GetRealtimeData) // 获取实时数据
	}

	// 静态文件服务
	router.Static("/static", "./web/static")
	router.LoadHTMLGlob("web/templates/*")

	// 首页
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "智能选股系统",
		})
	})

	utilsLogger.Info("Starting web server on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
