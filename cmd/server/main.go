package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"stock/internal/config"
	"stock/internal/handler"
	"stock/internal/utils"

	"github.com/gin-gonic/gin"
)

// @title 智能选股系统 API
// @version 1.0
// @description 基于Go语言开发的智能选股系统
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	// 初始化配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	logger := utils.NewLogger(cfg.Log)

	// 设置Gin模式
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// 初始化处理器
	handlers := handler.NewHandlers(cfg, logger)

	// 设置路由
	setupRoutes(router, handlers)

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:           fmt.Sprintf(":%d", cfg.App.Port),
		Handler:        router,
		ReadTimeout:    cfg.Server.ReadTimeout,
		WriteTimeout:   cfg.Server.WriteTimeout,
		IdleTimeout:    cfg.Server.IdleTimeout,
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
	}

	// 启动服务器
	go func() {
		logger.Infof("Server starting on port %d", cfg.App.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Server shutting down...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exited")
}

func setupRoutes(router *gin.Engine, handlers *handler.Handlers) {
	// 健康检查
	router.GET("/health", handlers.Health)

	// API路由组
	v1 := router.Group("/api/v1")
	{
		// 股票相关路由
		stocks := v1.Group("/stocks")
		{
			stocks.GET("", handlers.Stock.GetStocks)
			stocks.GET("/:code", handlers.Stock.GetStock)
			stocks.GET("/:code/data", handlers.Stock.GetStockData)
		}

		// 分析相关路由
		analysis := v1.Group("/analysis")
		{
			analysis.GET("/technical/:code", handlers.Analysis.GetTechnicalAnalysis)
			analysis.GET("/fundamental/:code", handlers.Analysis.GetFundamentalAnalysis)
		}

		// 选股相关路由
		selection := v1.Group("/selection")
		{
			selection.GET("/strategies", handlers.Strategy.GetStrategies)
			selection.POST("/execute", handlers.Strategy.ExecuteSelection)
			selection.GET("/results", handlers.Strategy.GetResults)
		}

		// 回测相关路由
		backtest := v1.Group("/backtest")
		{
			backtest.POST("/run", handlers.Backtest.RunBacktest)
			backtest.GET("/results/:id", handlers.Backtest.GetResult)
		}
	}

	// 静态文件服务
	router.Static("/static", "./web/static")
	router.LoadHTMLGlob("web/templates/*")

	// 前端页面路由
	router.GET("/", handlers.Web.Index)
	router.GET("/analysis", handlers.Web.Analysis)
	router.GET("/strategy", handlers.Web.Strategy)
}
