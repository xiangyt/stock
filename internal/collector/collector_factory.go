package collector

import (
	"fmt"
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

// CollectorFactory 采集器工厂
type CollectorFactory struct {
	logger *logger.Logger
}

// NewCollectorFactory 创建采集器工厂
func NewCollectorFactory(logger *logger.Logger) *CollectorFactory {
	return &CollectorFactory{
		logger: logger,
	}
}

// CreateCollector 创建指定类型的采集器
func (f *CollectorFactory) CreateCollector(collectorType CollectorType, config ...CollectorConfig) (DataCollector, error) {
	switch collectorType {
	case CollectorTypeEastMoney:
		return NewEastMoneyCollector(f.logger), nil

	case CollectorTypeHTTP:
		if len(config) == 0 {
			return nil, fmt.Errorf("HTTP collector requires configuration")
		}
		return NewHTTPCollector(config[0], f.logger), nil

	case CollectorTypeTushare:
		// 创建Tushare采集器配置
		cfg := CollectorConfig{
			Name:      "tushare",
			BaseURL:   "https://api.tushare.pro",
			Timeout:   30 * time.Second,
			RateLimit: 200, // Tushare限制每分钟200次
		}
		if len(config) > 0 {
			cfg = config[0]
		}
		return NewHTTPCollector(cfg, f.logger), nil

	case CollectorTypeAKShare:
		// 创建AKShare采集器配置
		cfg := CollectorConfig{
			Name:      "akshare",
			BaseURL:   "https://api.akshare.xyz",
			Timeout:   30 * time.Second,
			RateLimit: 100,
		}
		if len(config) > 0 {
			cfg = config[0]
		}
		return NewHTTPCollector(cfg, f.logger), nil

	default:
		return nil, fmt.Errorf("unsupported collector type: %s", collectorType)
	}
}

// CreateDefaultCollectors 创建默认的采集器集合
func (f *CollectorFactory) CreateDefaultCollectors() map[string]DataCollector {
	collectors := make(map[string]DataCollector)

	// 创建东方财富采集器
	if eastMoney, err := f.CreateCollector(CollectorTypeEastMoney); err == nil {
		collectors["eastmoney"] = eastMoney
	} else {
		f.logger.Errorf("Failed to create EastMoney collector: %v", err)
	}

	// 可以添加更多默认采集器
	// if tushare, err := f.CreateCollector(CollectorTypeTushare); err == nil {
	//     collectors["tushare"] = tushare
	// }

	return collectors
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
