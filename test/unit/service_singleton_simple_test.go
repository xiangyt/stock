package unit

import (
	"testing"

	"stock/internal/config"
	"stock/internal/service"
	"stock/internal/utils"
)

func TestServiceSingletonSimple(t *testing.T) {
	// 创建测试配置和logger
	cfg := &config.Config{}
	logger := utils.NewLogger(config.LogConfig{
		Level:  "info",
		Format: "text",
	})

	t.Run("TestDatabaseServiceSingleton", func(t *testing.T) {
		// 多次获取DatabaseService实例
		service1 := service.GetDatabaseService(cfg, logger)
		service2 := service.GetDatabaseService(cfg, logger)
		service3 := service.NewDatabaseService(cfg, logger) // 测试向后兼容

		// 验证是否为同一个实例
		if service1 != service2 {
			t.Error("DatabaseService单例模式失败：service1 != service2")
		}
		if service1 != service3 {
			t.Error("DatabaseService单例模式失败：service1 != service3")
		}
		if service2 != service3 {
			t.Error("DatabaseService单例模式失败：service2 != service3")
		}

		t.Logf("DatabaseService单例模式验证成功，实例地址: %p", service1)
	})

	t.Run("TestDataCollectorServiceSingleton", func(t *testing.T) {
		// 多次获取DataCollectorService实例
		service1 := service.GetDataCollectorService(cfg, logger)
		service2 := service.GetDataCollectorService(cfg, logger)
		service3 := service.NewDataCollectorService(cfg, logger) // 测试向后兼容

		// 验证是否为同一个实例
		if service1 != service2 {
			t.Error("DataCollectorService单例模式失败：service1 != service2")
		}
		if service1 != service3 {
			t.Error("DataCollectorService单例模式失败：service1 != service3")
		}
		if service2 != service3 {
			t.Error("DataCollectorService单例模式失败：service2 != service3")
		}

		t.Logf("DataCollectorService单例模式验证成功，实例地址: %p", service1)
	})

	t.Run("TestTechnicalAnalyzerServiceSingleton", func(t *testing.T) {
		// 多次获取TechnicalAnalyzerService实例
		service1 := service.GetTechnicalAnalyzerService(cfg, logger)
		service2 := service.GetTechnicalAnalyzerService(cfg, logger)
		service3 := service.NewTechnicalAnalyzerService(cfg, logger) // 测试向后兼容

		// 验证是否为同一个实例
		if service1 != service2 {
			t.Error("TechnicalAnalyzerService单例模式失败：service1 != service2")
		}
		if service1 != service3 {
			t.Error("TechnicalAnalyzerService单例模式失败：service1 != service3")
		}
		if service2 != service3 {
			t.Error("TechnicalAnalyzerService单例模式失败：service2 != service3")
		}

		t.Logf("TechnicalAnalyzerService单例模式验证成功，实例地址: %p", service1)
	})

	t.Run("TestStrategyEngineServiceSingleton", func(t *testing.T) {
		// 多次获取StrategyEngineService实例
		service1 := service.GetStrategyEngineService(cfg, logger)
		service2 := service.GetStrategyEngineService(cfg, logger)
		service3 := service.NewStrategyEngineService(cfg, logger) // 测试向后兼容

		// 验证是否为同一个实例
		if service1 != service2 {
			t.Error("StrategyEngineService单例模式失败：service1 != service2")
		}
		if service1 != service3 {
			t.Error("StrategyEngineService单例模式失败：service1 != service3")
		}
		if service2 != service3 {
			t.Error("StrategyEngineService单例模式失败：service2 != service3")
		}

		t.Logf("StrategyEngineService单例模式验证成功，实例地址: %p", service1)
	})

	t.Run("TestBacktestEngineServiceSingleton", func(t *testing.T) {
		// 多次获取BacktestEngineService实例
		service1 := service.GetBacktestEngineService(cfg, logger)
		service2 := service.GetBacktestEngineService(cfg, logger)
		service3 := service.NewBacktestEngineService(cfg, logger) // 测试向后兼容

		// 验证是否为同一个实例
		if service1 != service2 {
			t.Error("BacktestEngineService单例模式失败：service1 != service2")
		}
		if service1 != service3 {
			t.Error("BacktestEngineService单例模式失败：service1 != service3")
		}
		if service2 != service3 {
			t.Error("BacktestEngineService单例模式失败：service2 != service3")
		}

		t.Logf("BacktestEngineService单例模式验证成功，实例地址: %p", service1)
	})
}

func TestServicesCollectionSingleton(t *testing.T) {
	// 创建测试配置和logger
	cfg := &config.Config{}
	logger := utils.NewLogger(config.LogConfig{
		Level:  "info",
		Format: "text",
	})

	t.Run("TestServicesCreation", func(t *testing.T) {
		// 多次创建Services实例
		services1, err1 := service.NewServices(cfg, logger)
		services2, err2 := service.NewServices(cfg, logger)

		if err1 != nil {
			t.Fatalf("创建Services失败: %v", err1)
		}
		if err2 != nil {
			t.Fatalf("创建Services失败: %v", err2)
		}

		// 验证内部服务是否为单例
		if services1.Database != services2.Database {
			t.Error("Services中的DatabaseService不是单例")
		}
		if services1.DataCollector != services2.DataCollector {
			t.Error("Services中的DataCollectorService不是单例")
		}
		if services1.TechnicalAnalyzer != services2.TechnicalAnalyzer {
			t.Error("Services中的TechnicalAnalyzerService不是单例")
		}
		if services1.StrategyEngine != services2.StrategyEngine {
			t.Error("Services中的StrategyEngineService不是单例")
		}
		if services1.BacktestEngine != services2.BacktestEngine {
			t.Error("Services中的BacktestEngineService不是单例")
		}

		t.Logf("Services单例模式验证成功")
		t.Logf("DatabaseService地址: %p", services1.Database)
		t.Logf("DataCollectorService地址: %p", services1.DataCollector)
		t.Logf("TechnicalAnalyzerService地址: %p", services1.TechnicalAnalyzer)
		t.Logf("StrategyEngineService地址: %p", services1.StrategyEngine)
		t.Logf("BacktestEngineService地址: %p", services1.BacktestEngine)
	})

	t.Run("TestSingletonConsistency", func(t *testing.T) {
		// 直接获取服务实例
		dbService := service.GetDatabaseService(cfg, logger)
		dataCollectorService := service.GetDataCollectorService(cfg, logger)
		technicalAnalyzerService := service.GetTechnicalAnalyzerService(cfg, logger)
		strategyEngineService := service.GetStrategyEngineService(cfg, logger)
		backtestEngineService := service.GetBacktestEngineService(cfg, logger)

		// 通过Services获取实例
		services, err := service.NewServices(cfg, logger)
		if err != nil {
			t.Fatalf("创建Services失败: %v", err)
		}

		// 验证直接获取的实例与通过Services获取的实例是否一致
		if dbService != services.Database {
			t.Error("直接获取的DatabaseService与Services中的不一致")
		}
		if dataCollectorService != services.DataCollector {
			t.Error("直接获取的DataCollectorService与Services中的不一致")
		}
		if technicalAnalyzerService != services.TechnicalAnalyzer {
			t.Error("直接获取的TechnicalAnalyzerService与Services中的不一致")
		}
		if strategyEngineService != services.StrategyEngine {
			t.Error("直接获取的StrategyEngineService与Services中的不一致")
		}
		if backtestEngineService != services.BacktestEngine {
			t.Error("直接获取的BacktestEngineService与Services中的不一致")
		}

		t.Logf("单例一致性验证成功")
	})
}
