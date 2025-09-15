package strategy

import (
	"fmt"
	"time"

	"stock/internal/model"
	"stock/internal/repository"
	"stock/internal/utils"
)

// TechnicalStrategy 技术分析策略
type TechnicalStrategy struct {
	*BaseStrategy
	dailyDataRepo *repository.DailyDataRepository
}

// NewTechnicalStrategy 创建技术分析策略
func NewTechnicalStrategy(name, description string, dailyDataRepo *repository.DailyDataRepository, logger *utils.Logger) *TechnicalStrategy {
	return &TechnicalStrategy{
		BaseStrategy:  NewBaseStrategy(name, description, StrategyTypeTechnical, logger),
		dailyDataRepo: dailyDataRepo,
	}
}

// Execute 执行技术分析策略
func (s *TechnicalStrategy) Execute(stocks []model.Stock, date time.Time) ([]StrategyResult, error) {
	s.logger.Infof("Executing technical strategy: %s for %d stocks", s.GetName(), len(stocks))

	var results []StrategyResult

	// 获取策略参数
	lookbackDays := s.GetIntParameter("lookback_days", 20)
	minVolume := s.GetFloatParameter("min_volume", 1000000)
	scoreThreshold := s.GetFloatParameter("score_threshold", 60.0)

	for _, stock := range stocks {
		if !stock.IsActive {
			continue
		}

		// 获取历史数据
		endDate := date
		startDate := date.AddDate(0, 0, -lookbackDays-10) // 多获取一些数据用于计算指标

		dailyData, err := s.dailyDataRepo.GetDailyDataByTsCode(stock.TsCode, startDate, endDate, 0)
		if err != nil {
			s.logger.Warnf("Failed to get daily data for %s: %v", stock.TsCode, err)
			continue
		}

		if len(dailyData) < lookbackDays {
			s.logger.Debugf("Insufficient data for %s: got %d, need %d", stock.TsCode, len(dailyData), lookbackDays)
			continue
		}

		// 计算技术指标和评分
		score, signal, reason, details := s.calculateTechnicalScore(stock, dailyData, minVolume)

		if score >= scoreThreshold {
			result := s.CreateResult(stock.TsCode, score, signal, reason, details)
			results = append(results, result)
		}
	}

	// 排序并返回结果
	results = s.SortResults(results)
	s.logger.Infof("Technical strategy completed: found %d candidates", len(results))

	return results, nil
}

// calculateTechnicalScore 计算技术分析评分
func (s *TechnicalStrategy) calculateTechnicalScore(stock model.Stock, dailyData []model.DailyData, minVolume float64) (float64, SignalType, string, map[string]interface{}) {
	if len(dailyData) == 0 {
		return 0, SignalIgnore, "No data available", nil
	}

	// 按日期排序（最新的在前）
	sortedData := make([]model.DailyData, len(dailyData))
	copy(sortedData, dailyData)

	// 简单排序（实际项目中应该使用更高效的排序）
	for i := 0; i < len(sortedData)-1; i++ {
		for j := i + 1; j < len(sortedData); j++ {
			if sortedData[i].TradeDate.Before(sortedData[j].TradeDate) {
				sortedData[i], sortedData[j] = sortedData[j], sortedData[i]
			}
		}
	}

	latest := sortedData[0]

	// 检查成交量
	if latest.Volume < int64(minVolume) {
		return 0, SignalIgnore, "Volume too low", map[string]interface{}{
			"volume":     latest.Volume,
			"min_volume": minVolume,
		}
	}

	var score float64
	var reasons []string
	details := make(map[string]interface{})

	// 1. 价格趋势分析 (30分)
	trendScore, trendReason := s.analyzePriceTrend(sortedData)
	score += trendScore
	reasons = append(reasons, trendReason)
	details["trend_score"] = trendScore

	// 2. 成交量分析 (20分)
	volumeScore, volumeReason := s.analyzeVolume(sortedData)
	score += volumeScore
	reasons = append(reasons, volumeReason)
	details["volume_score"] = volumeScore

	// 3. 移动平均线分析 (25分)
	maScore, maReason := s.analyzeMovingAverage(sortedData)
	score += maScore
	reasons = append(reasons, maReason)
	details["ma_score"] = maScore

	// 4. 相对强弱指标 (25分)
	rsiScore, rsiReason := s.analyzeRSI(sortedData)
	score += rsiScore
	reasons = append(reasons, rsiReason)
	details["rsi_score"] = rsiScore

	// 确定信号类型
	var signal SignalType
	if score >= 80 {
		signal = SignalBuy
	} else if score >= 60 {
		signal = SignalHold
	} else if score <= 30 {
		signal = SignalSell
	} else {
		signal = SignalIgnore
	}

	// 合并原因
	reason := fmt.Sprintf("技术分析评分: %.1f分. %s", score, joinReasons(reasons))

	details["latest_price"] = latest.Close
	details["latest_volume"] = latest.Volume
	details["total_score"] = score

	return score, signal, reason, details
}

// analyzePriceTrend 分析价格趋势
func (s *TechnicalStrategy) analyzePriceTrend(data []model.DailyData) (float64, string) {
	if len(data) < 5 {
		return 0, "数据不足"
	}

	// 计算短期趋势（5日）
	shortTrend := (data[0].Close - data[4].Close) / data[4].Close * 100

	// 计算中期趋势（10日）
	var midTrend float64
	if len(data) >= 10 {
		midTrend = (data[0].Close - data[9].Close) / data[9].Close * 100
	}

	var score float64
	var reason string

	if shortTrend > 5 && midTrend > 3 {
		score = 30
		reason = "强势上涨趋势"
	} else if shortTrend > 2 && midTrend > 0 {
		score = 20
		reason = "温和上涨趋势"
	} else if shortTrend > -2 && midTrend > -2 {
		score = 10
		reason = "横盘整理"
	} else if shortTrend < -5 || midTrend < -5 {
		score = 0
		reason = "下跌趋势"
	} else {
		score = 5
		reason = "趋势不明"
	}

	return score, reason
}

// analyzeVolume 分析成交量
func (s *TechnicalStrategy) analyzeVolume(data []model.DailyData) (float64, string) {
	if len(data) < 5 {
		return 0, "数据不足"
	}

	// 计算平均成交量
	var avgVolume int64
	for i := 1; i < 6 && i < len(data); i++ {
		avgVolume += data[i].Volume
	}
	avgVolume /= 5

	currentVolume := data[0].Volume
	volumeRatio := float64(currentVolume) / float64(avgVolume)

	var score float64
	var reason string

	if volumeRatio > 2.0 {
		score = 20
		reason = "成交量大幅放大"
	} else if volumeRatio > 1.5 {
		score = 15
		reason = "成交量明显放大"
	} else if volumeRatio > 1.2 {
		score = 10
		reason = "成交量温和放大"
	} else if volumeRatio < 0.5 {
		score = 0
		reason = "成交量萎缩"
	} else {
		score = 5
		reason = "成交量正常"
	}

	return score, reason
}

// analyzeMovingAverage 分析移动平均线
func (s *TechnicalStrategy) analyzeMovingAverage(data []model.DailyData) (float64, string) {
	if len(data) < 20 {
		return 0, "数据不足"
	}

	// 计算5日和20日移动平均线
	var ma5, ma20 float64

	for i := 0; i < 5; i++ {
		ma5 += data[i].Close
	}
	ma5 /= 5

	for i := 0; i < 20; i++ {
		ma20 += data[i].Close
	}
	ma20 /= 20

	currentPrice := data[0].Close

	var score float64
	var reason string

	if currentPrice > ma5 && ma5 > ma20 {
		score = 25
		reason = "价格站上均线，多头排列"
	} else if currentPrice > ma5 {
		score = 15
		reason = "价格站上短期均线"
	} else if currentPrice > ma20 {
		score = 10
		reason = "价格站上长期均线"
	} else if currentPrice < ma5 && ma5 < ma20 {
		score = 0
		reason = "价格跌破均线，空头排列"
	} else {
		score = 5
		reason = "均线纠缠"
	}

	return score, reason
}

// analyzeRSI 分析相对强弱指标
func (s *TechnicalStrategy) analyzeRSI(data []model.DailyData) (float64, string) {
	if len(data) < 14 {
		return 0, "数据不足"
	}

	// 简化的RSI计算
	var gains, losses float64
	for i := 1; i < 14; i++ {
		change := data[i-1].Close - data[i].Close
		if change > 0 {
			gains += change
		} else {
			losses += -change
		}
	}

	if losses == 0 {
		return 25, "RSI超买区间"
	}

	rs := gains / losses
	rsi := 100 - (100 / (1 + rs))

	var score float64
	var reason string

	if rsi > 70 {
		score = 5
		reason = "RSI超买"
	} else if rsi > 50 {
		score = 20
		reason = "RSI强势区间"
	} else if rsi > 30 {
		score = 15
		reason = "RSI中性区间"
	} else {
		score = 25
		reason = "RSI超卖，可能反弹"
	}

	return score, reason
}

// Validate 验证策略参数
func (s *TechnicalStrategy) Validate() error {
	lookbackDays := s.GetIntParameter("lookback_days", 20)
	if lookbackDays < 5 || lookbackDays > 100 {
		return fmt.Errorf("lookback_days must be between 5 and 100")
	}

	minVolume := s.GetFloatParameter("min_volume", 1000000)
	if minVolume < 0 {
		return fmt.Errorf("min_volume must be non-negative")
	}

	scoreThreshold := s.GetFloatParameter("score_threshold", 60.0)
	if scoreThreshold < 0 || scoreThreshold > 100 {
		return fmt.Errorf("score_threshold must be between 0 and 100")
	}

	return nil
}

// 辅助函数
func joinReasons(reasons []string) string {
	if len(reasons) == 0 {
		return ""
	}

	result := reasons[0]
	for i := 1; i < len(reasons); i++ {
		result += "; " + reasons[i]
	}
	return result
}
