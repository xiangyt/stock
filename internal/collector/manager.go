package collector

import (
	"fmt"
	"sync"
	"time"

	"stock/internal/model"
	"stock/internal/utils"
)

// CollectorManager 采集器管理器
type CollectorManager struct {
	collectors map[string]DataCollector
	logger     *utils.Logger
	mu         sync.RWMutex
}

// NewCollectorManager 创建采集器管理器
func NewCollectorManager(logger *utils.Logger) *CollectorManager {
	return &CollectorManager{
		collectors: make(map[string]DataCollector),
		logger:     logger,
	}
}

// RegisterCollector 注册采集器
func (m *CollectorManager) RegisterCollector(name string, collector DataCollector) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.collectors[name] = collector
	m.logger.Infof("Registered collector: %s", name)
}

// GetCollector 获取采集器
func (m *CollectorManager) GetCollector(name string) (DataCollector, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	collector, exists := m.collectors[name]
	if !exists {
		return nil, fmt.Errorf("collector not found: %s", name)
	}

	return collector, nil
}

// GetAvailableCollectors 获取可用的采集器列表
func (m *CollectorManager) GetAvailableCollectors() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.collectors))
	for name, collector := range m.collectors {
		if collector.IsConnected() {
			names = append(names, name)
		}
	}

	return names
}

// ConnectAll 连接所有采集器
func (m *CollectorManager) ConnectAll() error {
	m.mu.RLock()
	collectors := make(map[string]DataCollector)
	for name, collector := range m.collectors {
		collectors[name] = collector
	}
	m.mu.RUnlock()

	var errors []error
	for name, collector := range collectors {
		if err := collector.Connect(); err != nil {
			m.logger.Errorf("Failed to connect collector %s: %v", name, err)
			errors = append(errors, fmt.Errorf("collector %s: %v", name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to connect some collectors: %v", errors)
	}

	return nil
}

// DisconnectAll 断开所有采集器
func (m *CollectorManager) DisconnectAll() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for name, collector := range m.collectors {
		if err := collector.Disconnect(); err != nil {
			m.logger.Errorf("Failed to disconnect collector %s: %v", name, err)
		}
	}
}

// GetStockListFromSource 从指定数据源获取股票列表
func (m *CollectorManager) GetStockListFromSource(sourceName string) ([]model.Stock, error) {
	collector, err := m.GetCollector(sourceName)
	if err != nil {
		return nil, err
	}

	return collector.GetStockList()
}

// GetStockDataFromSource 从指定数据源获取股票数据
func (m *CollectorManager) GetStockDataFromSource(sourceName, tsCode string, startDate, endDate time.Time) ([]model.DailyData, error) {
	collector, err := m.GetCollector(sourceName)
	if err != nil {
		return nil, err
	}

	return collector.GetStockData(tsCode, startDate, endDate)
}

// GetRealtimeDataFromSource 从指定数据源获取实时数据
func (m *CollectorManager) GetRealtimeDataFromSource(sourceName string, tsCodes []string) ([]model.DailyData, error) {
	collector, err := m.GetCollector(sourceName)
	if err != nil {
		return nil, err
	}

	return collector.GetRealtimeData(tsCodes)
}

// GetPerformanceReportsFromSource 从指定数据源获取业绩报表数据
func (m *CollectorManager) GetPerformanceReportsFromSource(sourceName, tsCode string) ([]model.PerformanceReport, error) {
	collector, err := m.GetCollector(sourceName)
	if err != nil {
		return nil, err
	}

	return collector.GetPerformanceReports(tsCode)
}

// GetStockListWithFallback 获取股票列表（支持备用数据源）
func (m *CollectorManager) GetStockListWithFallback(primarySource string, fallbackSources []string) ([]model.Stock, error) {
	// 尝试主数据源
	if stocks, err := m.GetStockListFromSource(primarySource); err == nil {
		return stocks, nil
	} else {
		m.logger.Warnf("Primary source %s failed: %v", primarySource, err)
	}

	// 尝试备用数据源
	for _, source := range fallbackSources {
		if stocks, err := m.GetStockListFromSource(source); err == nil {
			m.logger.Infof("Using fallback source: %s", source)
			return stocks, nil
		} else {
			m.logger.Warnf("Fallback source %s failed: %v", source, err)
		}
	}

	return nil, fmt.Errorf("all data sources failed")
}

// GetStockDataWithFallback 获取股票数据（支持备用数据源）
func (m *CollectorManager) GetStockDataWithFallback(primarySource string, fallbackSources []string, tsCode string, startDate, endDate time.Time) ([]model.DailyData, error) {
	// 尝试主数据源
	if data, err := m.GetStockDataFromSource(primarySource, tsCode, startDate, endDate); err == nil {
		return data, nil
	} else {
		m.logger.Warnf("Primary source %s failed for %s: %v", primarySource, tsCode, err)
	}

	// 尝试备用数据源
	for _, source := range fallbackSources {
		if data, err := m.GetStockDataFromSource(source, tsCode, startDate, endDate); err == nil {
			m.logger.Infof("Using fallback source %s for %s", source, tsCode)
			return data, nil
		} else {
			m.logger.Warnf("Fallback source %s failed for %s: %v", source, tsCode, err)
		}
	}

	return nil, fmt.Errorf("all data sources failed for %s", tsCode)
}
