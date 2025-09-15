package strategy

import (
	"fmt"
	"time"

	"stock/internal/model"
	"stock/internal/repository"
	"stock/internal/utils"
)

// FundamentalStrategy 基本面分析策略
type FundamentalStrategy struct {
	*BaseStrategy
	financialDataRepo *repository.FinancialDataRepository
	dailyDataRepo     *repository.DailyDataRepository
}

// NewFundamentalStrategy 创建基本面分析策略
func NewFundamentalStrategy(name, description string, financialDataRepo *repository.FinancialDataRepository, dailyDataRepo *repository.DailyDataRepository, logger *utils.Logger) *FundamentalStrategy {
	return &FundamentalStrategy{
		BaseStrategy:      NewBaseStrategy(name, description, StrategyTypeFundamental, logger),
		financialDataRepo: financialDataRepo,
		dailyDataRepo:     dailyDataRepo,
	}
}

// Execute 执行基本面分析策略
func (s *FundamentalStrategy) Execute(stocks []model.Stock, date time.Time) ([]StrategyResult, error) {
	s.logger.Infof("Executing fundamental strategy: %s for %d stocks", s.GetName(), len(stocks))

	var results []StrategyResult

	// 获取策略参数
	minROE := s.GetFloatParameter("min_roe", 10.0)
	maxPE := s.GetFloatParameter("max_pe", 30.0)
	maxPB := s.GetFloatParameter("max_pb", 5.0)
	minMarketCap := s.GetFloatParameter("min_market_cap", 1000000000) // 10亿
	scoreThreshold := s.GetFloatParameter("score_threshold", 60.0)

	for _, stock := range stocks {
		if !stock.IsActive {
			continue
		}

		// 获取最新财务数据
		financialData, err := s.getLatestFinancialData(stock.TsCode, date)
		if err != nil {
			s.logger.Warnf("Failed to get financial data for %s: %v", stock.TsCode, err)
			continue
		}

		if financialData == nil {
			s.logger.Debugf("No financial data available for %s", stock.TsCode)
			continue
		}

		// 获取最新价格数据
		latestPrice, err := s.dailyDataRepo.GetLatestDailyData(stock.TsCode)
		if err != nil || latestPrice == nil {
			s.logger.Warnf("Failed to get latest price for %s: %v", stock.TsCode, err)
			continue
		}

		// 计算基本面评分
		score, signal, reason, details := s.calculateFundamentalScore(stock, *financialData, *latestPrice, minROE, maxPE, maxPB, minMarketCap)

		if score >= scoreThreshold {
			result := s.CreateResult(stock.TsCode, score, signal, reason, details)
			results = append(results, result)
		}
	}

	// 排序并返回结果
	results = s.SortResults(results)
	s.logger.Infof("Fundamental strategy completed: found %d candidates", len(results))

	return results, nil
}

// getLatestFinancialData 获取最新财务数据
func (s *FundamentalStrategy) getLatestFinancialData(tsCode string, date time.Time) (*model.FinancialData, error) {
	// 获取指定日期前的最新财务数据
	startDate := date.AddDate(-2, 0, 0) // 查找2年内的数据

	financialDataList, err := s.financialDataRepo.GetFinancialDataByTsCode(tsCode, startDate, date, 1)
	if err != nil {
		return nil, err
	}

	if len(financialDataList) == 0 {
		return nil, nil
	}

	return &financialDataList[0], nil
}

// calculateFundamentalScore 计算基本面评分
func (s *FundamentalStrategy) calculateFundamentalScore(stock model.Stock, financial model.FinancialData, price model.DailyData, minROE, maxPE, maxPB, minMarketCap float64) (float64, SignalType, string, map[string]interface{}) {
	var score float64
	var reasons []string
	details := make(map[string]interface{})

	// 估算市值（简化计算）
	estimatedShares := 1000000000.0 // 假设10亿股，实际应该从股本数据获取
	marketCap := price.Close * estimatedShares

	// 1. ROE分析 (25分)
	roeScore, roeReason := s.analyzeROE(financial.ROE, minROE)
	score += roeScore
	reasons = append(reasons, roeReason)
	details["roe_score"] = roeScore
	details["roe"] = financial.ROE

	// 2. PE比率分析 (25分)
	peScore, peReason := s.analyzePE(financial.PERatio, maxPE)
	score += peScore
	reasons = append(reasons, peReason)
	details["pe_score"] = peScore
	details["pe_ratio"] = financial.PERatio

	// 3. PB比率分析 (20分)
	pbScore, pbReason := s.analyzePB(financial.PBRatio, maxPB)
	score += pbScore
	reasons = append(reasons, pbReason)
	details["pb_score"] = pbScore
	details["pb_ratio"] = financial.PBRatio

	// 4. 盈利能力分析 (15分)
	profitScore, profitReason := s.analyzeProfitability(financial.GrossProfitMargin, financial.NetProfitMargin)
	score += profitScore
	reasons = append(reasons, profitReason)
	details["profit_score"] = profitScore

	// 5. 财务健康度分析 (15分)
	healthScore, healthReason := s.analyzeFinancialHealth(financial.DebtToAssets, financial.ROA)
	score += healthScore
	reasons = append(reasons, healthReason)
	details["health_score"] = healthScore

	// 市值过滤
	if marketCap < minMarketCap {
		score *= 0.5 // 市值过小，评分减半
		reasons = append(reasons, "市值偏小")
	}

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
	reason := fmt.Sprintf("基本面分析评分: %.1f分. %s", score, joinReasons(reasons))

	details["market_cap"] = marketCap
	details["total_score"] = score
	details["latest_price"] = price.Close

	return score, signal, reason, details
}

// analyzeROE 分析净资产收益率
func (s *FundamentalStrategy) analyzeROE(roe, minROE float64) (float64, string) {
	if roe >= minROE*1.5 {
		return 25, "ROE优秀"
	} else if roe >= minROE {
		return 20, "ROE良好"
	} else if roe >= minROE*0.7 {
		return 10, "ROE一般"
	} else if roe > 0 {
		return 5, "ROE偏低"
	} else {
		return 0, "ROE为负"
	}
}

// analyzePE 分析市盈率
func (s *FundamentalStrategy) analyzePE(pe, maxPE float64) (float64, string) {
	if pe <= 0 {
		return 0, "PE为负或零"
	} else if pe <= maxPE*0.5 {
		return 25, "PE很低，估值便宜"
	} else if pe <= maxPE*0.8 {
		return 20, "PE较低，估值合理"
	} else if pe <= maxPE {
		return 15, "PE适中"
	} else if pe <= maxPE*1.5 {
		return 5, "PE偏高"
	} else {
		return 0, "PE过高"
	}
}

// analyzePB 分析市净率
func (s *FundamentalStrategy) analyzePB(pb, maxPB float64) (float64, string) {
	if pb <= 0 {
		return 0, "PB为负或零"
	} else if pb <= 1.0 {
		return 20, "PB小于1，破净股"
	} else if pb <= maxPB*0.6 {
		return 18, "PB很低"
	} else if pb <= maxPB {
		return 15, "PB合理"
	} else if pb <= maxPB*1.5 {
		return 5, "PB偏高"
	} else {
		return 0, "PB过高"
	}
}

// analyzeProfitability 分析盈利能力
func (s *FundamentalStrategy) analyzeProfitability(grossMargin, netMargin float64) (float64, string) {
	var score float64
	var reason string

	// 毛利率评分
	if grossMargin >= 50 {
		score += 8
	} else if grossMargin >= 30 {
		score += 6
	} else if grossMargin >= 20 {
		score += 4
	} else if grossMargin >= 10 {
		score += 2
	}

	// 净利率评分
	if netMargin >= 20 {
		score += 7
	} else if netMargin >= 10 {
		score += 5
	} else if netMargin >= 5 {
		score += 3
	} else if netMargin >= 0 {
		score += 1
	}

	if score >= 12 {
		reason = "盈利能力优秀"
	} else if score >= 8 {
		reason = "盈利能力良好"
	} else if score >= 4 {
		reason = "盈利能力一般"
	} else {
		reason = "盈利能力较差"
	}

	return score, reason
}

// analyzeFinancialHealth 分析财务健康度
func (s *FundamentalStrategy) analyzeFinancialHealth(debtRatio, roa float64) (float64, string) {
	var score float64
	var reason string

	// 资产负债率评分
	if debtRatio <= 30 {
		score += 8
	} else if debtRatio <= 50 {
		score += 6
	} else if debtRatio <= 70 {
		score += 3
	} else {
		score += 0
	}

	// ROA评分
	if roa >= 10 {
		score += 7
	} else if roa >= 5 {
		score += 5
	} else if roa >= 2 {
		score += 3
	} else if roa > 0 {
		score += 1
	}

	if score >= 12 {
		reason = "财务健康度优秀"
	} else if score >= 8 {
		reason = "财务健康度良好"
	} else if score >= 4 {
		reason = "财务健康度一般"
	} else {
		reason = "财务健康度较差"
	}

	return score, reason
}

// Validate 验证策略参数
func (s *FundamentalStrategy) Validate() error {
	minROE := s.GetFloatParameter("min_roe", 10.0)
	if minROE < 0 || minROE > 50 {
		return fmt.Errorf("min_roe must be between 0 and 50")
	}

	maxPE := s.GetFloatParameter("max_pe", 30.0)
	if maxPE <= 0 || maxPE > 100 {
		return fmt.Errorf("max_pe must be between 0 and 100")
	}

	maxPB := s.GetFloatParameter("max_pb", 5.0)
	if maxPB <= 0 || maxPB > 20 {
		return fmt.Errorf("max_pb must be between 0 and 20")
	}

	minMarketCap := s.GetFloatParameter("min_market_cap", 1000000000)
	if minMarketCap < 0 {
		return fmt.Errorf("min_market_cap must be non-negative")
	}

	scoreThreshold := s.GetFloatParameter("score_threshold", 60.0)
	if scoreThreshold < 0 || scoreThreshold > 100 {
		return fmt.Errorf("score_threshold must be between 0 and 100")
	}

	return nil
}
