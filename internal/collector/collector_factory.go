package collector

import (
	"fmt"
	"sync"
	"time"

	"stock/internal/logger"
	"stock/internal/model"
)

// CollectorType 采集器类型
type CollectorType string

const (
	CollectorTypeEastMoney   CollectorType = "eastmoney"
	CollectorTypeTongHuaShun CollectorType = "tonghuashun"
	CollectorTypeHTTP        CollectorType = "http"
	CollectorTypeTushare     CollectorType = "tushare"
	CollectorTypeAKShare     CollectorType = "akshare"
)

// 单例实例存储
var (
	eastMoneyCollectorInstance *EastMoneyCollector
	eastMoneyCollectorOnce     sync.Once

	tongHuaShunCollectorInstance *TongHuaShunCollector
	tongHuaShunCollectorOnce     sync.Once

	httpCollectorInstances = make(map[string]*HTTPCollector)
	httpCollectorMutex     sync.RWMutex

	tushareCollectorInstance *HTTPCollector
	tushareCollectorOnce     sync.Once

	akshareCollectorInstance *HTTPCollector
	akshareCollectorOnce     sync.Once

	factoryInstance *CollectorFactory
	factoryOnce     sync.Once
)

// CollectorFactory 采集器工厂
type CollectorFactory struct {
	logger     *logger.Logger
	collectors map[string]DataCollector
	mu         sync.RWMutex
}

// GetCollectorFactory 获取采集器工厂单例
func GetCollectorFactory(log *logger.Logger) *CollectorFactory {
	factoryOnce.Do(func() {
		factoryInstance = &CollectorFactory{
			logger:     log,
			collectors: make(map[string]DataCollector),
		}
	})
	return factoryInstance
}

// GetEastMoneyCollector 获取东方财富采集器单例
func (f *CollectorFactory) GetEastMoneyCollector() *EastMoneyCollector {
	eastMoneyCollectorOnce.Do(func() {
		eastMoneyCollectorInstance = newEastMoneyCollector(f.logger)
	})
	return eastMoneyCollectorInstance
}

// GetTongHuaShunCollector 获取同花顺采集器单例
func (f *CollectorFactory) GetTongHuaShunCollector() *TongHuaShunCollector {
	tongHuaShunCollectorOnce.Do(func() {
		tongHuaShunCollectorInstance = newTongHuaShunCollector(f.logger)
	})
	return tongHuaShunCollectorInstance
}

// GetHTTPCollector 获取HTTP采集器单例（根据配置名称区分）
func (f *CollectorFactory) GetHTTPCollector(config CollectorConfig) *HTTPCollector {
	httpCollectorMutex.Lock()
	defer httpCollectorMutex.Unlock()

	if instance, exists := httpCollectorInstances[config.Name]; exists {
		return instance
	}

	instance := NewHTTPCollector(config, f.logger)
	httpCollectorInstances[config.Name] = instance
	return instance
}

// GetTushareCollector 获取Tushare采集器单例
func (f *CollectorFactory) GetTushareCollector(config ...CollectorConfig) *HTTPCollector {
	tushareCollectorOnce.Do(func() {
		cfg := CollectorConfig{
			Name:      "tushare",
			BaseURL:   "https://api.tushare.pro",
			Timeout:   30 * time.Second,
			RateLimit: 200, // Tushare限制每分钟200次
		}
		if len(config) > 0 {
			cfg = config[0]
		}
		tushareCollectorInstance = NewHTTPCollector(cfg, f.logger)
	})
	return tushareCollectorInstance
}

// GetAKShareCollector 获取AKShare采集器单例
func (f *CollectorFactory) GetAKShareCollector(config ...CollectorConfig) *HTTPCollector {
	akshareCollectorOnce.Do(func() {
		cfg := CollectorConfig{
			Name:      "akshare",
			BaseURL:   "https://api.akshare.xyz",
			Timeout:   30 * time.Second,
			RateLimit: 100,
		}
		if len(config) > 0 {
			cfg = config[0]
		}
		akshareCollectorInstance = NewHTTPCollector(cfg, f.logger)
	})
	return akshareCollectorInstance
}

// CreateCollector 创建指定类型的采集器（使用单例）
func (f *CollectorFactory) CreateCollector(collectorType CollectorType, config ...CollectorConfig) (DataCollector, error) {
	switch collectorType {
	case CollectorTypeEastMoney:
		return f.GetEastMoneyCollector(), nil

	case CollectorTypeTongHuaShun:
		return f.GetTongHuaShunCollector(), nil

	case CollectorTypeHTTP:
		if len(config) == 0 {
			return nil, fmt.Errorf("HTTP collector requires configuration")
		}
		return f.GetHTTPCollector(config[0]), nil

	case CollectorTypeTushare:
		return f.GetTushareCollector(config...), nil

	case CollectorTypeAKShare:
		return f.GetAKShareCollector(config...), nil

	default:
		return nil, fmt.Errorf("unsupported collector type: %s", collectorType)
	}
}

// CreateDefaultCollectors 创建默认的采集器集合（使用单例）
func (f *CollectorFactory) CreateDefaultCollectors() map[string]DataCollector {
	collectors := make(map[string]DataCollector)

	// 创建东方财富采集器（单例）
	collectors["eastmoney"] = f.GetEastMoneyCollector()

	// 创建同花顺采集器（单例）
	collectors["tonghuashun"] = f.GetTongHuaShunCollector()

	// 创建Tushare采集器（单例）
	collectors["tushare"] = f.GetTushareCollector()

	// 创建AKShare采集器（单例）
	collectors["akshare"] = f.GetAKShareCollector()

	f.logger.Infof("Created %d default collectors using singleton pattern", len(collectors))
	return collectors
}

// GetAllCollectors 获取所有已创建的采集器实例
func (f *CollectorFactory) GetAllCollectors() map[string]DataCollector {
	collectors := make(map[string]DataCollector)

	// 东方财富采集器
	if eastMoneyCollectorInstance != nil {
		collectors["eastmoney"] = eastMoneyCollectorInstance
	}

	// 同花顺采集器
	if tongHuaShunCollectorInstance != nil {
		collectors["tonghuashun"] = tongHuaShunCollectorInstance
	}

	// Tushare采集器
	if tushareCollectorInstance != nil {
		collectors["tushare"] = tushareCollectorInstance
	}

	// AKShare采集器
	if akshareCollectorInstance != nil {
		collectors["akshare"] = akshareCollectorInstance
	}

	// HTTP采集器实例
	httpCollectorMutex.RLock()
	for name, instance := range httpCollectorInstances {
		collectors[name] = instance
	}
	httpCollectorMutex.RUnlock()

	return collectors
}

// ResetCollectors 重置所有采集器实例（主要用于测试）
func (f *CollectorFactory) ResetCollectors() {
	eastMoneyCollectorOnce = sync.Once{}
	eastMoneyCollectorInstance = nil

	tongHuaShunCollectorOnce = sync.Once{}
	tongHuaShunCollectorInstance = nil

	tushareCollectorOnce = sync.Once{}
	tushareCollectorInstance = nil

	akshareCollectorOnce = sync.Once{}
	akshareCollectorInstance = nil

	httpCollectorMutex.Lock()
	httpCollectorInstances = make(map[string]*HTTPCollector)
	httpCollectorMutex.Unlock()

	f.logger.Info("All collector instances have been reset")
}

// GetSupportedCollectors 获取支持的采集器类型列表
func (f *CollectorFactory) GetSupportedCollectors() []CollectorType {
	return []CollectorType{
		CollectorTypeEastMoney,
		CollectorTypeTongHuaShun,
		CollectorTypeHTTP,
		CollectorTypeTushare,
		CollectorTypeAKShare,
	}
}

// ===== CollectorManager兼容方法 =====

// RegisterCollector 注册采集器
func (f *CollectorFactory) RegisterCollector(name string, collector DataCollector) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.collectors[name] = collector
	f.logger.Infof("Registered collector: %s", name)
}

// GetCollector 获取采集器
func (f *CollectorFactory) GetCollector(name string) (DataCollector, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	collector, exists := f.collectors[name]
	if !exists {
		// 尝试创建默认采集器
		switch name {
		case "eastmoney":
			collector = f.GetEastMoneyCollector()
			f.collectors[name] = collector
			return collector, nil
		case "tonghuashun":
			collector = f.GetTongHuaShunCollector()
			f.collectors[name] = collector
			return collector, nil
		case "tushare":
			collector = f.GetTushareCollector()
			f.collectors[name] = collector
			return collector, nil
		case "akshare":
			collector = f.GetAKShareCollector()
			f.collectors[name] = collector
			return collector, nil
		default:
			return nil, fmt.Errorf("collector not found: %s", name)
		}
	}

	return collector, nil
}

// GetAvailableCollectors 获取可用的采集器列表
func (f *CollectorFactory) GetAvailableCollectors() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	names := make([]string, 0, len(f.collectors))
	for name, collector := range f.collectors {
		if collector.IsConnected() {
			names = append(names, name)
		}
	}

	return names
}

// ConnectAll 连接所有采集器
func (f *CollectorFactory) ConnectAll() error {
	f.mu.RLock()
	collectors := make(map[string]DataCollector)
	for name, collector := range f.collectors {
		collectors[name] = collector
	}
	f.mu.RUnlock()

	var errors []error
	for name, collector := range collectors {
		if err := collector.Connect(); err != nil {
			f.logger.Errorf("Failed to connect collector %s: %v", name, err)
			errors = append(errors, fmt.Errorf("collector %s: %v", name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to connect some collectors: %v", errors)
	}

	return nil
}

// DisconnectAll 断开所有采集器
func (f *CollectorFactory) DisconnectAll() {
	f.mu.RLock()
	defer f.mu.RUnlock()

	for name, collector := range f.collectors {
		if err := collector.Disconnect(); err != nil {
			f.logger.Errorf("Failed to disconnect collector %s: %v", name, err)
		}
	}
}

// GetStockListFromSource 从指定数据源获取股票列表
func (f *CollectorFactory) GetStockListFromSource(sourceName string) ([]model.Stock, error) {
	collector, err := f.GetCollector(sourceName)
	if err != nil {
		return nil, err
	}

	return collector.GetStockList()
}

// GetStockDataFromSource 从指定数据源获取股票数据
func (f *CollectorFactory) GetStockDataFromSource(sourceName, tsCode string, startDate, endDate time.Time) ([]model.DailyData, error) {
	collector, err := f.GetCollector(sourceName)
	if err != nil {
		return nil, err
	}

	return collector.GetStockData(tsCode, startDate, endDate)
}

// GetRealtimeDataFromSource 从指定数据源获取实时数据
func (f *CollectorFactory) GetRealtimeDataFromSource(sourceName string, tsCodes []string) ([]model.DailyData, error) {
	collector, err := f.GetCollector(sourceName)
	if err != nil {
		return nil, err
	}

	return collector.GetRealtimeData(tsCodes)
}

// GetPerformanceReportsFromSource 从指定数据源获取业绩报表数据
func (f *CollectorFactory) GetPerformanceReportsFromSource(sourceName, tsCode string) ([]model.PerformanceReport, error) {
	collector, err := f.GetCollector(sourceName)
	if err != nil {
		return nil, err
	}

	return collector.GetPerformanceReports(tsCode)
}

// GetStockListWithFallback 获取股票列表（支持备用数据源）
func (f *CollectorFactory) GetStockListWithFallback(primarySource string, fallbackSources []string) ([]model.Stock, error) {
	// 尝试主数据源
	if stocks, err := f.GetStockListFromSource(primarySource); err == nil {
		return stocks, nil
	} else {
		f.logger.Warnf("Primary source %s failed: %v", primarySource, err)
	}

	// 尝试备用数据源
	for _, source := range fallbackSources {
		if stocks, err := f.GetStockListFromSource(source); err == nil {
			f.logger.Infof("Using fallback source: %s", source)
			return stocks, nil
		} else {
			f.logger.Warnf("Fallback source %s failed: %v", source, err)
		}
	}

	return nil, fmt.Errorf("all data sources failed")
}

// GetStockDataWithFallback 获取股票数据（支持备用数据源）
func (f *CollectorFactory) GetStockDataWithFallback(primarySource string, fallbackSources []string, tsCode string, startDate, endDate time.Time) ([]model.DailyData, error) {
	// 尝试主数据源
	if data, err := f.GetStockDataFromSource(primarySource, tsCode, startDate, endDate); err == nil {
		return data, nil
	} else {
		f.logger.Warnf("Primary source %s failed for %s: %v", primarySource, tsCode, err)
	}

	// 尝试备用数据源
	for _, source := range fallbackSources {
		if data, err := f.GetStockDataFromSource(source, tsCode, startDate, endDate); err == nil {
			f.logger.Infof("Using fallback source %s for %s", source, tsCode)
			return data, nil
		} else {
			f.logger.Warnf("Fallback source %s failed for %s: %v", source, tsCode, err)
		}
	}

	return nil, fmt.Errorf("all data sources failed for %s", tsCode)
}
