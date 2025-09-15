package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"stock/internal/config"
	"stock/internal/service"
	"stock/internal/utils"

	"github.com/robfig/cron/v3"
)

func main() {
	// 初始化配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	logger := utils.NewLogger(cfg.Log)

	// 初始化服务
	services, err := service.NewServices(cfg, logger)
	if err != nil {
		log.Fatalf("Failed to initialize services: %v", err)
	}

	// 创建定时任务调度器
	c := cron.New(cron.WithSeconds())

	// 添加定时任务
	setupCronJobs(c, services, logger)

	// 启动调度器
	c.Start()
	logger.Info("Worker started, cron jobs scheduled")

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Worker shutting down...")

	// 停止调度器
	ctx := c.Stop()
	<-ctx.Done()

	logger.Info("Worker exited")
}

func setupCronJobs(c *cron.Cron, services *service.Services, logger *utils.Logger) {
	// 每日数据更新 - 工作日18:00
	c.AddFunc("0 0 18 * * 1-5", func() {
		logger.Info("Starting daily data update...")
		if err := services.DataCollector.UpdateDailyData(); err != nil {
			logger.Errorf("Daily data update failed: %v", err)
		} else {
			logger.Info("Daily data update completed")
		}
	})

	// 财务数据更新 - 每周六02:00
	c.AddFunc("0 0 2 * * 6", func() {
		logger.Info("Starting financial data update...")
		if err := services.DataCollector.UpdateFinancialData(); err != nil {
			logger.Errorf("Financial data update failed: %v", err)
		} else {
			logger.Info("Financial data update completed")
		}
	})

	// 实时数据更新 - 工作日每5分钟
	c.AddFunc("0 */5 * * * 1-5", func() {
		logger.Debug("Starting realtime data update...")
		if err := services.DataCollector.UpdateRealtimeData(); err != nil {
			logger.Errorf("Realtime data update failed: %v", err)
		} else {
			logger.Debug("Realtime data update completed")
		}
	})

	// 技术指标计算 - 每日19:00
	c.AddFunc("0 0 19 * * 1-5", func() {
		logger.Info("Starting technical indicators calculation...")
		if err := services.TechnicalAnalyzer.CalculateAllIndicators(); err != nil {
			logger.Errorf("Technical indicators calculation failed: %v", err)
		} else {
			logger.Info("Technical indicators calculation completed")
		}
	})

	// 选股任务 - 每日20:00
	c.AddFunc("0 0 20 * * 1-5", func() {
		logger.Info("Starting daily stock selection...")
		strategies := []string{"technical", "fundamental", "combined"}
		for _, strategy := range strategies {
			if _, err := services.StrategyEngine.ExecuteStrategy(strategy, 50); err != nil {
				logger.Errorf("Stock selection failed for strategy %s: %v", strategy, err)
			} else {
				logger.Infof("Stock selection completed for strategy %s", strategy)
			}
		}
	})

	logger.Info("Cron jobs configured:")
	logger.Info("- Daily data update: 18:00 (Mon-Fri)")
	logger.Info("- Financial data update: 02:00 (Saturday)")
	logger.Info("- Realtime data update: Every 5 minutes (Mon-Fri)")
	logger.Info("- Technical indicators: 19:00 (Mon-Fri)")
	logger.Info("- Stock selection: 20:00 (Mon-Fri)")
}
