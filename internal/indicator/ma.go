package indicator

import "stock/internal/model"

// Ma 计算均线数据
func Ma(stocks []IndStock) []*model.TechnicalIndicator {
	if len(stocks) == 0 {
		return nil
	}

	closes := make([]float64, len(stocks))
	for i, s := range stocks {
		_, _, _, closes[i] = s.Get4Price()
	}

	inds := make([]*model.TechnicalIndicator, len(stocks))
	ma5 := calculateMa(closes, 5)
	ma10 := calculateMa(closes, 10)
	ma20 := calculateMa(closes, 20)
	ma60 := calculateMa(closes, 60)
	for i := range inds {
		inds[i] = &model.TechnicalIndicator{
			Symbol:    stocks[i].GetSymbol(),
			TradeDate: stocks[i].GetTradeDate(),
			Ma5:       ma5[i],
			Ma10:      ma10[i],
			Ma20:      ma20[i],
			Ma60:      ma60[i],
		}
	}
	return inds
}

// calculateMa 高性能MA计算，使用真正的滑动窗口
func calculateMa(data []float64, period int) []float64 {
	if len(data) == 0 || period <= 0 {
		return nil
	}

	result := make([]float64, len(data))
	if len(data) < period {
		return result
	}

	// 处理第一个值
	//result[0] = data[0]
	//
	//// 对于前period-1个数据
	//sum := data[0]
	//for i := 1; i < period && i < len(data); i++ {
	//	sum += data[i]
	//	result[i] = sum / float64(i+1)
	//}

	// 使用滑动窗口计算后续值
	if len(data) >= period {
		// 计算第一个完整窗口的和
		windowSum := 0.0
		for i := 0; i < period; i++ {
			windowSum += data[i]
		}
		result[period-1] = windowSum / float64(period)

		// 滑动窗口
		for i := period; i < len(data); i++ {
			windowSum = windowSum - data[i-period] + data[i]
			result[i] = windowSum / float64(period)
		}
	}

	return result
}
