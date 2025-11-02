package main

import (
	"log"
	"net/http"

	"stock/internal/api"
	"stock/internal/collector"
	"stock/internal/config"
	"stock/internal/database"
	"stock/internal/logger"
	"stock/internal/model"

	"github.com/gin-gonic/gin"
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
	if err := db.AutoMigrate(
		&model.Stock{},
		&model.DailyData{},
		&model.Task{},
		&model.Strategy{},
		&model.PerformanceReport{},

		&model.User{},
		&model.Role{},
		&model.Permission{},
		&model.RolePermission{},
		&model.UserLoginLog{},
		&model.UserOperationLog{},
		&model.JWTBlacklist{},
	); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	utilsLogger.Info("Database migration completed successfully")

	// 创建采集器工厂
	collectorFactory := collector.GetCollectorFactory(utilsLogger)

	// 获取东方财富采集器
	eastMoneyCollector := collectorFactory.GetEastMoneyCollector()
	if err := eastMoneyCollector.Connect(); err != nil {
		log.Fatalf("Failed to connect to data source: %v", err)
	}

	// 注册采集器
	collectorFactory.RegisterCollector("eastmoney", eastMoneyCollector)

	// 创建API处理器（传入数据库连接）
	apiHandler := api.NewHandler(collectorFactory, logrusLogger, db)

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

	// 使用模块化路由
	api.SetupRoutes(router, apiHandler)

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
