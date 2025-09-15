package collector

import (
	"stock/internal/model"
	"time"
)

// DataCollector 数据采集器接口
type DataCollector interface {
	// 连接数据源
	Connect() error

	// 断开连接
	Disconnect() error

	// 获取股票列表
	GetStockList() ([]model.Stock, error)

	// 获取股票详情
	GetStockDetail(tsCode string) (*model.Stock, error)

	// 获取股票历史数据
	GetStockData(tsCode string, startDate, endDate time.Time) ([]model.DailyData, error)

	// 获取日K线数据
	GetDailyKLine(tsCode string, startDate, endDate time.Time) ([]model.DailyData, error)

	// 获取实时数据
	GetRealtimeData(tsCodes []string) ([]model.DailyData, error)

	// 获取业绩报表数据
	GetPerformanceReports(tsCode string) ([]model.PerformanceReport, error)

	// 获取最新业绩报表数据
	GetLatestPerformanceReport(tsCode string) (*model.PerformanceReport, error)

	// 检查连接状态
	IsConnected() bool

	// 获取数据源名称
	GetName() string
}

// CollectorConfig 采集器配置
type CollectorConfig struct {
	Name      string            `json:"name"`
	Enabled   bool              `json:"enabled"`
	BaseURL   string            `json:"base_url"`
	Token     string            `json:"token"`
	Headers   map[string]string `json:"headers"`
	Timeout   time.Duration     `json:"timeout"`
	RateLimit int               `json:"rate_limit"` // 每秒请求数限制
}

// BaseCollector 基础采集器
type BaseCollector struct {
	Config    CollectorConfig
	Connected bool
}

// GetName 获取采集器名称
func (b *BaseCollector) GetName() string {
	return b.Config.Name
}

// IsConnected 检查连接状态
func (b *BaseCollector) IsConnected() bool {
	return b.Connected
}

// Connect 连接数据源
func (b *BaseCollector) Connect() error {
	b.Connected = true
	return nil
}

// Disconnect 断开连接
func (b *BaseCollector) Disconnect() error {
	b.Connected = false
	return nil
}
