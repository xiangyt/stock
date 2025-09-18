package service

import (
	"stock/internal/config"
	"stock/internal/utils"
	"sync"
)

// Services 服务集合
type Services struct {
	Database           *DatabaseService
	DataCollector      *DataCollectorService
	TechnicalAnalyzer  *TechnicalAnalyzerService
	StrategyEngine     *StrategyEngineService
	BacktestEngine     *BacktestEngineService
	DataService        *DataService
	PerformanceService *PerformanceService
}

// NewServices 创建服务集合 (使用单例模式)
func NewServices(cfg *config.Config, logger *utils.Logger) (*Services, error) {
	// 初始化数据库服务 (使用单例)
	dbService := GetDatabaseService(cfg, logger)

	// 初始化数据采集服务 (使用单例)
	dataCollectorService := GetDataCollectorService(cfg, logger)

	// 初始化技术分析服务 (使用单例)
	technicalAnalyzerService := GetTechnicalAnalyzerService(cfg, logger)

	// 初始化策略引擎服务 (使用单例)
	strategyEngineService := GetStrategyEngineService(cfg, logger)

	// 初始化回测引擎服务 (使用单例)
	backtestEngineService := GetBacktestEngineService(cfg, logger)

	// 注意：DataService和PerformanceService需要数据库连接，这里先设为nil
	// 在实际使用时需要通过InitServicesWithDB来初始化
	return &Services{
		Database:           dbService,
		DataCollector:      dataCollectorService,
		TechnicalAnalyzer:  technicalAnalyzerService,
		StrategyEngine:     strategyEngineService,
		BacktestEngine:     backtestEngineService,
		DataService:        nil, // 需要数据库连接后初始化
		PerformanceService: nil, // 需要数据库连接后初始化
	}, nil
}

// InitServicesWithDB 使用数据库连接初始化服务
func (s *Services) InitServicesWithDB(db interface{}, logger *utils.Logger) error {
	// 这里暂时返回nil，具体实现需要根据实际的数据库和repository结构来完成
	logger.Info("Services initialized with database connection")
	return nil
}

// DatabaseService 数据库服务
type DatabaseService struct {
	config *config.Config
	logger *utils.Logger
}

var (
	databaseServiceInstance *DatabaseService
	databaseServiceOnce     sync.Once
)

// GetDatabaseService 获取数据库服务单例
func GetDatabaseService(cfg *config.Config, logger *utils.Logger) *DatabaseService {
	databaseServiceOnce.Do(func() {
		databaseServiceInstance = &DatabaseService{
			config: cfg,
			logger: logger,
		}
	})
	return databaseServiceInstance
}

// NewDatabaseService 创建数据库服务 (保持向后兼容)
func NewDatabaseService(cfg *config.Config, logger *utils.Logger) *DatabaseService {
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
	logger *utils.Logger
}

var (
	dataCollectorServiceInstance *DataCollectorService
	dataCollectorServiceOnce     sync.Once
)

// GetDataCollectorService 获取数据采集服务单例
func GetDataCollectorService(cfg *config.Config, logger *utils.Logger) *DataCollectorService {
	dataCollectorServiceOnce.Do(func() {
		dataCollectorServiceInstance = &DataCollectorService{
			cfg:    cfg,
			logger: logger,
		}
	})
	return dataCollectorServiceInstance
}

// NewDataCollectorService 创建数据采集服务 (保持向后兼容)
func NewDataCollectorService(cfg *config.Config, logger *utils.Logger) *DataCollectorService {
	return GetDataCollectorService(cfg, logger)
}

// UpdateAllData 更新所有数据
func (s *DataCollectorService) UpdateAllData(source string) error {
	s.logger.Infof("Updating all data from %s...", source)
	// TODO: 实现数据更新逻辑
	return nil
}

// UpdateDailyData 更新日线数据
func (s *DataCollectorService) UpdateDailyData() error {
	s.logger.Info("Updating daily data...")
	// TODO: 实现日线数据更新逻辑
	return nil
}

// UpdateFinancialData 更新财务数据
func (s *DataCollectorService) UpdateFinancialData() error {
	s.logger.Info("Updating financial data...")
	// TODO: 实现财务数据更新逻辑
	return nil
}

// UpdateRealtimeData 更新实时数据
func (s *DataCollectorService) UpdateRealtimeData() error {
	s.logger.Debug("Updating realtime data...")
	// TODO: 实现实时数据更新逻辑
	return nil
}

// TechnicalAnalyzerService 技术分析服务
type TechnicalAnalyzerService struct {
	cfg    *config.Config
	logger *utils.Logger
}

var (
	technicalAnalyzerServiceInstance *TechnicalAnalyzerService
	technicalAnalyzerServiceOnce     sync.Once
)

// GetTechnicalAnalyzerService 获取技术分析服务单例
func GetTechnicalAnalyzerService(cfg *config.Config, logger *utils.Logger) *TechnicalAnalyzerService {
	technicalAnalyzerServiceOnce.Do(func() {
		technicalAnalyzerServiceInstance = &TechnicalAnalyzerService{
			cfg:    cfg,
			logger: logger,
		}
	})
	return technicalAnalyzerServiceInstance
}

// NewTechnicalAnalyzerService 创建技术分析服务 (保持向后兼容)
func NewTechnicalAnalyzerService(cfg *config.Config, logger *utils.Logger) *TechnicalAnalyzerService {
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
	logger *utils.Logger
}

var (
	strategyEngineServiceInstance *StrategyEngineService
	strategyEngineServiceOnce     sync.Once
)

// GetStrategyEngineService 获取策略引擎服务单例
func GetStrategyEngineService(cfg *config.Config, logger *utils.Logger) *StrategyEngineService {
	strategyEngineServiceOnce.Do(func() {
		strategyEngineServiceInstance = &StrategyEngineService{
			cfg:    cfg,
			logger: logger,
		}
	})
	return strategyEngineServiceInstance
}

// NewStrategyEngineService 创建策略引擎服务 (保持向后兼容)
func NewStrategyEngineService(cfg *config.Config, logger *utils.Logger) *StrategyEngineService {
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
	logger *utils.Logger
}

var (
	backtestEngineServiceInstance *BacktestEngineService
	backtestEngineServiceOnce     sync.Once
)

// GetBacktestEngineService 获取回测引擎服务单例
func GetBacktestEngineService(cfg *config.Config, logger *utils.Logger) *BacktestEngineService {
	backtestEngineServiceOnce.Do(func() {
		backtestEngineServiceInstance = &BacktestEngineService{
			cfg:    cfg,
			logger: logger,
		}
	})
	return backtestEngineServiceInstance
}

// NewBacktestEngineService 创建回测引擎服务 (保持向后兼容)
func NewBacktestEngineService(cfg *config.Config, logger *utils.Logger) *BacktestEngineService {
	return GetBacktestEngineService(cfg, logger)
}
