package collector

import (
	"fmt"
	"sync"
	"time"

	"stock/internal/logger"
)

// CollectorType 采集器类型
type CollectorType string

const (
	CollectorTypeEastMoney CollectorType = "eastmoney"
	CollectorTypeHTTP      CollectorType = "http"
	CollectorTypeTushare   CollectorType = "tushare"
	CollectorTypeAKShare   CollectorType = "akshare"
)

// 单例实例存储
var (
	eastMoneyCollectorInstance *EastMoneyCollector
	eastMoneyCollectorOnce     sync.Once

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
	logger *logger.Logger
}

// GetCollectorFactory 获取采集器工厂单例
func GetCollectorFactory(log *logger.Logger) *CollectorFactory {
	factoryOnce.Do(func() {
		factoryInstance = &CollectorFactory{
			logger: log,
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
		CollectorTypeHTTP,
		CollectorTypeTushare,
		CollectorTypeAKShare,
	}
}
