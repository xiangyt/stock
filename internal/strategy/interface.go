package strategy

import (
	"time"

	"stock/internal/model"
)

// Strategy 选股策略接口
type Strategy interface {
	// GetName 获取策略名称
	GetName() string

	// GetDescription 获取策略描述
	GetDescription() string

	// GetType 获取策略类型
	GetType() StrategyType

	// GetParameters 获取策略参数
	GetParameters() map[string]interface{}

	// SetParameters 设置策略参数
	SetParameters(params map[string]interface{}) error

	// Execute 执行选股策略
	Execute(stocks []model.Stock, date time.Time) ([]StrategyResult, error)

	// Validate 验证策略参数
	Validate() error
}

// StrategyType 策略类型
type StrategyType string

const (
	StrategyTypeTechnical    StrategyType = "technical"    // 技术分析策略
	StrategyTypeFundamental  StrategyType = "fundamental"  // 基本面分析策略
	StrategyTypeQuantitative StrategyType = "quantitative" // 量化策略
	StrategyTypeComposite    StrategyType = "composite"    // 复合策略
)

// StrategyResult 策略执行结果
type StrategyResult struct {
	TsCode    string                 `json:"ts_code"`   // 股票代码
	Score     float64                `json:"score"`     // 评分 (0-100)
	Rank      int                    `json:"rank"`      // 排名
	Signal    SignalType             `json:"signal"`    // 信号类型
	Reason    string                 `json:"reason"`    // 选股原因
	Details   map[string]interface{} `json:"details"`   // 详细信息
	Timestamp time.Time              `json:"timestamp"` // 计算时间
}

// SignalType 信号类型
type SignalType string

const (
	SignalBuy    SignalType = "buy"    // 买入信号
	SignalSell   SignalType = "sell"   // 卖出信号
	SignalHold   SignalType = "hold"   // 持有信号
	SignalIgnore SignalType = "ignore" // 忽略信号
)

// StrategyConfig 策略配置
type StrategyConfig struct {
	Name        string                 `json:"name"`        // 策略名称
	Type        StrategyType           `json:"type"`        // 策略类型
	Description string                 `json:"description"` // 策略描述
	Parameters  map[string]interface{} `json:"parameters"`  // 策略参数
	IsActive    bool                   `json:"is_active"`   // 是否启用
	CreatedBy   string                 `json:"created_by"`  // 创建者
}

// StrategyFactory 策略工厂接口
type StrategyFactory interface {
	// CreateStrategy 创建策略实例
	CreateStrategy(config StrategyConfig) (Strategy, error)

	// GetAvailableStrategies 获取可用策略列表
	GetAvailableStrategies() []string

	// GetStrategyInfo 获取策略信息
	GetStrategyInfo(name string) (*StrategyInfo, error)
}

// StrategyInfo 策略信息
type StrategyInfo struct {
	Name          string                 `json:"name"`           // 策略名称
	Type          StrategyType           `json:"type"`           // 策略类型
	Description   string                 `json:"description"`    // 策略描述
	Parameters    []ParameterInfo        `json:"parameters"`     // 参数信息
	DefaultParams map[string]interface{} `json:"default_params"` // 默认参数
}

// ParameterInfo 参数信息
type ParameterInfo struct {
	Name        string      `json:"name"`        // 参数名称
	Type        string      `json:"type"`        // 参数类型
	Description string      `json:"description"` // 参数描述
	Required    bool        `json:"required"`    // 是否必需
	Default     interface{} `json:"default"`     // 默认值
	Min         interface{} `json:"min"`         // 最小值
	Max         interface{} `json:"max"`         // 最大值
}
