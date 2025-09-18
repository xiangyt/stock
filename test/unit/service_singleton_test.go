package unit

import (
	"testing"

	"stock/internal/config"
	"stock/internal/service"
	"stock/internal/utils"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestServiceSingleton(t *testing.T) {
	// 创建测试配置和logger
	cfg := &config.Config{}
	logger := utils.NewLogger(config.LogConfig{
		Level:  "info",
		Format: "text",
	})

	// 尝试连接测试数据库，如果失败则跳过测试
	dsn := "root:123456@tcp(localhost:3306)/stock_test?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("跳过测试，无法连接数据库: %v", err)
		return
	}

	t.Run("TestDataServiceSingleton", func(t *testing.T) {
		// 多次获取DataService实例
		service1 := service.GetDataService(db, logger)
		service2 := service.GetDataService(db, logger)
		service3 := service.NewDataService(db, logger) // 测试向后兼容

		// 验证是否为同一个实例
		if service1 != service2 {
			t.Error("DataService单例模式失败：service1 != service2")
		}
		if service1 != service3 {
			t.Error("DataService单例模式失败：service1 != service3")
		}
		if service2 != service3 {
			t.Error("DataService单例模式失败：service2 != service3")
		}

		t.Logf("DataService单例模式验证成功，实例地址: %p", service1)
	})

	t.Run("TestKLinePersistenceServiceSingleton", func(t *testing.T) {
		// 多次获取KLinePersistenceService实例
		service1 := service.GetKLinePersistenceService(db, logger)
		service2 := service.GetKLinePersistenceService(db, logger)
		service3 := service.NewKLinePersistenceService(db, logger) // 测试向后兼容

		// 验证是否为同一个实例
		if service1 != service2 {
			t.Error("KLinePersistenceService单例模式失败：service1 != service2")
		}
		if service1 != service3 {
			t.Error("KLinePersistenceService单例模式失败：service1 != service3")
		}
		if service2 != service3 {
			t.Error("KLinePersistenceService单例模式失败：service2 != service3")
		}

		t.Logf("KLinePersistenceService单例模式验证成功，实例地址: %p", service1)
	})

	t.Run("TestDailyKLineManagerSingleton", func(t *testing.T) {
		// 多次获取DailyKLineManager实例
		manager1 := service.GetDailyKLineManager(db, logger)
		manager2 := service.GetDailyKLineManager(db, logger)
		manager3 := service.NewDailyKLineManager(db, logger) // 测试向后兼容

		// 验证是否为同一个实例
		if manager1 != manager2 {
			t.Error("DailyKLineManager单例模式失败：manager1 != manager2")
		}
		if manager1 != manager3 {
			t.Error("DailyKLineManager单例模式失败：manager1 != manager3")
		}
		if manager2 != manager3 {
			t.Error("DailyKLineManager单例模式失败：manager2 != manager3")
		}

		t.Logf("DailyKLineManager单例模式验证成功，实例地址: %p", manager1)
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

func TestServicesSingleton(t *testing.T) {
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
}
