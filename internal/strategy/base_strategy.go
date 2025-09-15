package strategy

import (
	"fmt"
	"time"

	"stock/internal/utils"
)

// BaseStrategy 基础策略结构
type BaseStrategy struct {
	name         string
	description  string
	strategyType StrategyType
	parameters   map[string]interface{}
	logger       *utils.Logger
}

// NewBaseStrategy 创建基础策略
func NewBaseStrategy(name, description string, strategyType StrategyType, logger *utils.Logger) *BaseStrategy {
	return &BaseStrategy{
		name:         name,
		description:  description,
		strategyType: strategyType,
		parameters:   make(map[string]interface{}),
		logger:       logger,
	}
}

// GetName 获取策略名称
func (s *BaseStrategy) GetName() string {
	return s.name
}

// GetDescription 获取策略描述
func (s *BaseStrategy) GetDescription() string {
	return s.description
}

// GetType 获取策略类型
func (s *BaseStrategy) GetType() StrategyType {
	return s.strategyType
}

// GetParameters 获取策略参数
func (s *BaseStrategy) GetParameters() map[string]interface{} {
	return s.parameters
}

// SetParameters 设置策略参数
func (s *BaseStrategy) SetParameters(params map[string]interface{}) error {
	if params == nil {
		return fmt.Errorf("parameters cannot be nil")
	}
	s.parameters = params
	return nil
}

// GetParameter 获取单个参数
func (s *BaseStrategy) GetParameter(key string, defaultValue interface{}) interface{} {
	if value, exists := s.parameters[key]; exists {
		return value
	}
	return defaultValue
}

// GetFloatParameter 获取浮点数参数
func (s *BaseStrategy) GetFloatParameter(key string, defaultValue float64) float64 {
	if value, exists := s.parameters[key]; exists {
		switch v := value.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case string:
			if f, err := parseFloat(v); err == nil {
				return f
			}
		}
	}
	return defaultValue
}

// GetIntParameter 获取整数参数
func (s *BaseStrategy) GetIntParameter(key string, defaultValue int) int {
	if value, exists := s.parameters[key]; exists {
		switch v := value.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case string:
			if i, err := parseInt(v); err == nil {
				return i
			}
		}
	}
	return defaultValue
}

// GetStringParameter 获取字符串参数
func (s *BaseStrategy) GetStringParameter(key string, defaultValue string) string {
	if value, exists := s.parameters[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return defaultValue
}

// GetBoolParameter 获取布尔参数
func (s *BaseStrategy) GetBoolParameter(key string, defaultValue bool) bool {
	if value, exists := s.parameters[key]; exists {
		if b, ok := value.(bool); ok {
			return b
		}
	}
	return defaultValue
}

// CreateResult 创建策略结果
func (s *BaseStrategy) CreateResult(tsCode string, score float64, signal SignalType, reason string, details map[string]interface{}) StrategyResult {
	return StrategyResult{
		TsCode:    tsCode,
		Score:     score,
		Signal:    signal,
		Reason:    reason,
		Details:   details,
		Timestamp: time.Now(),
	}
}

// SortResults 按评分排序结果
func (s *BaseStrategy) SortResults(results []StrategyResult) []StrategyResult {
	// 使用简单的冒泡排序，按评分降序排列
	n := len(results)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if results[j].Score < results[j+1].Score {
				results[j], results[j+1] = results[j+1], results[j]
			}
		}
	}

	// 设置排名
	for i := range results {
		results[i].Rank = i + 1
	}

	return results
}

// FilterByScore 按评分过滤结果
func (s *BaseStrategy) FilterByScore(results []StrategyResult, minScore float64) []StrategyResult {
	var filtered []StrategyResult
	for _, result := range results {
		if result.Score >= minScore {
			filtered = append(filtered, result)
		}
	}
	return filtered
}

// FilterBySignal 按信号类型过滤结果
func (s *BaseStrategy) FilterBySignal(results []StrategyResult, signal SignalType) []StrategyResult {
	var filtered []StrategyResult
	for _, result := range results {
		if result.Signal == signal {
			filtered = append(filtered, result)
		}
	}
	return filtered
}

// LimitResults 限制结果数量
func (s *BaseStrategy) LimitResults(results []StrategyResult, limit int) []StrategyResult {
	if limit <= 0 || limit >= len(results) {
		return results
	}
	return results[:limit]
}

// 辅助函数
func parseFloat(s interface{}) (float64, error) {
	switch v := s.(type) {
	case string:
		return parseFloatFromString(v)
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	default:
		return 0, fmt.Errorf("cannot parse %v to float64", s)
	}
}

func parseInt(s interface{}) (int, error) {
	switch v := s.(type) {
	case string:
		return parseIntFromString(v)
	case int:
		return v, nil
	case float64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("cannot parse %v to int", s)
	}
}

func parseFloatFromString(s string) (float64, error) {
	// 简单的字符串到浮点数转换
	// 在实际项目中应该使用 strconv.ParseFloat
	return 0, fmt.Errorf("string parsing not implemented")
}

func parseIntFromString(s string) (int, error) {
	// 简单的字符串到整数转换
	// 在实际项目中应该使用 strconv.Atoi
	return 0, fmt.Errorf("string parsing not implemented")
}
