package service

import (
	"stock/internal/notification"
	"sync"

	"stock/internal/config"
	"stock/internal/logger"
)

// Services 服务集合
type Services struct {
	Database           *DatabaseService
	DataService        *DataService
	PerformanceService *PerformanceService
	ShareholderService *ShareholderService
	IndicatorService   *IndicatorService
	NotifyManger       *notification.Manager
}

// NewServices 创建服务集合 (使用单例模式)
func NewServices(cfg *config.Config, logger *logger.Logger) (*Services, error) {
	// 初始化数据库服务 (使用单例)
	dbService := GetDatabaseService(cfg, logger)
	// 注意：DataService、PerformanceService和ShareholderService需要数据库连接，这里先设为nil
	// 在实际使用时需要通过InitServicesWithDB来初始化
	notify, err := notification.NewFactory(logger).CreateManager(&cfg.Notify)
	if err != nil {
		return nil, err
	}
	return &Services{
		Database:           dbService,
		NotifyManger:       notify,
		DataService:        nil, // 需要数据库连接后初始化
		PerformanceService: nil, // 需要数据库连接后初始化
		ShareholderService: nil, // 需要数据库连接后初始化
		IndicatorService:   nil, // 需要数据库连接后初始化
	}, nil
}

// InitServicesWithDB 使用数据库连接初始化服务
func (s *Services) InitServicesWithDB(db interface{}, logger *logger.Logger) error {
	// 这里暂时返回nil，具体实现需要根据实际的数据库和repository结构来完成
	logger.Info("Services initialized with database connection")
	return nil
}

// DatabaseService 数据库服务
type DatabaseService struct {
	config *config.Config
	logger *logger.Logger
}

var (
	databaseServiceInstance *DatabaseService
	databaseServiceOnce     sync.Once
)

// GetDatabaseService 获取数据库服务单例
func GetDatabaseService(cfg *config.Config, logger *logger.Logger) *DatabaseService {
	databaseServiceOnce.Do(func() {
		databaseServiceInstance = &DatabaseService{
			config: cfg,
			logger: logger,
		}
	})
	return databaseServiceInstance
}

// NewDatabaseService 创建数据库服务 (保持向后兼容)
func NewDatabaseService(cfg *config.Config, logger *logger.Logger) *DatabaseService {
	return GetDatabaseService(cfg, logger)
}

// InitDB 初始化数据库
func (s *DatabaseService) InitDB() error {
	s.logger.Info("Initializing database...")
	// TODO: 实现数据库初始化逻辑
	return nil
}

// Migrate 数据库迁移
func (s *DatabaseService) Migrate() error {
	s.logger.Info("Running database migration...")
	// TODO: 实现数据库迁移逻辑
	return nil
}

// DataCollectorService 数据采集服务
type DataCollectorService struct {
	cfg    *config.Config
	logger *logger.Logger
}

var (
	dataCollectorServiceInstance *DataCollectorService
	dataCollectorServiceOnce     sync.Once
)

// GetDataCollectorService 获取数据采集服务单例
func GetDataCollectorService(cfg *config.Config, logger *logger.Logger) *DataCollectorService {
	dataCollectorServiceOnce.Do(func() {
		dataCollectorServiceInstance = &DataCollectorService{
			cfg:    cfg,
			logger: logger,
		}
	})
	return dataCollectorServiceInstance
}

// NewDataCollectorService 创建数据采集服务 (保持向后兼容)
func NewDataCollectorService(cfg *config.Config, logger *logger.Logger) *DataCollectorService {
	return GetDataCollectorService(cfg, logger)
}

// TechnicalAnalyzerService 技术分析服务
type TechnicalAnalyzerService struct {
	cfg    *config.Config
	logger *logger.Logger
}

var (
	technicalAnalyzerServiceInstance *TechnicalAnalyzerService
	technicalAnalyzerServiceOnce     sync.Once
)

// GetTechnicalAnalyzerService 获取技术分析服务单例
func GetTechnicalAnalyzerService(cfg *config.Config, logger *logger.Logger) *TechnicalAnalyzerService {
	technicalAnalyzerServiceOnce.Do(func() {
		technicalAnalyzerServiceInstance = &TechnicalAnalyzerService{
			cfg:    cfg,
			logger: logger,
		}
	})
	return technicalAnalyzerServiceInstance
}

// NewTechnicalAnalyzerService 创建技术分析服务 (保持向后兼容)
func NewTechnicalAnalyzerService(cfg *config.Config, logger *logger.Logger) *TechnicalAnalyzerService {
	return GetTechnicalAnalyzerService(cfg, logger)
}

// CalculateAllIndicators 计算所有技术指标
func (s *TechnicalAnalyzerService) CalculateAllIndicators() error {
	s.logger.Info("Calculating all technical indicators...")
	// TODO: 实现技术指标计算逻辑
	return nil
}

// StrategyEngineService 策略引擎服务
type StrategyEngineService struct {
	cfg    *config.Config
	logger *logger.Logger
}

var (
	strategyEngineServiceInstance *StrategyEngineService
	strategyEngineServiceOnce     sync.Once
)

// GetStrategyEngineService 获取策略引擎服务单例
func GetStrategyEngineService(cfg *config.Config, logger *logger.Logger) *StrategyEngineService {
	strategyEngineServiceOnce.Do(func() {
		strategyEngineServiceInstance = &StrategyEngineService{
			cfg:    cfg,
			logger: logger,
		}
	})
	return strategyEngineServiceInstance
}

// NewStrategyEngineService 创建策略引擎服务 (保持向后兼容)
func NewStrategyEngineService(cfg *config.Config, logger *logger.Logger) *StrategyEngineService {
	return GetStrategyEngineService(cfg, logger)
}

// SelectionResult 选股结果
type SelectionResult struct {
	Stock  Stock   `json:"stock"`
	Score  float64 `json:"score"`
	Reason string  `json:"reason"`
}

// Stock 股票信息
type Stock struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Industry string `json:"industry"`
}

// ExecuteStrategy 执行选股策略
func (s *StrategyEngineService) ExecuteStrategy(strategy string, limit int) ([]SelectionResult, error) {
	s.logger.Infof("Executing strategy: %s (limit: %d)", strategy, limit)

	// TODO: 实现真实的选股逻辑
	// 这里返回一些示例数据
	results := []SelectionResult{
		{
			Stock:  Stock{Code: "000001.SZ", Name: "平安银行", Industry: "银行"},
			Score:  85.6,
			Reason: "技术指标良好，RSI处于合理区间",
		},
		{
			Stock:  Stock{Code: "000002.SZ", Name: "万科A", Industry: "房地产开发"},
			Score:  78.9,
			Reason: "基本面稳健，估值合理",
		},
	}

	return results, nil
}

// BacktestEngineService 回测引擎服务
type BacktestEngineService struct {
	cfg    *config.Config
	logger *logger.Logger
}

var (
	backtestEngineServiceInstance *BacktestEngineService
	backtestEngineServiceOnce     sync.Once
)

// GetBacktestEngineService 获取回测引擎服务单例
func GetBacktestEngineService(cfg *config.Config, logger *logger.Logger) *BacktestEngineService {
	backtestEngineServiceOnce.Do(func() {
		backtestEngineServiceInstance = &BacktestEngineService{
			cfg:    cfg,
			logger: logger,
		}
	})
	return backtestEngineServiceInstance
}

// NewBacktestEngineService 创建回测引擎服务 (保持向后兼容)
func NewBacktestEngineService(cfg *config.Config, logger *logger.Logger) *BacktestEngineService {
	return GetBacktestEngineService(cfg, logger)
}
