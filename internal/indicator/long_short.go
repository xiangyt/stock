package indicator

import (
	"math"
	"stock/internal/model"
)

/*
指标特点说明：

动态支撑阻力系统：
阻力线：基于动态价格范围计算的阻力位
支撑线：基于动态价格范围计算的支撑位
中线：支撑和阻力的中间位置

趋势判断系统：
趋势线：基于55周期价格范围的归一化指标
多级别判断：从超卖到超买的完整趋势分级

买卖信号系统：
准备信号：趋势进入极端区域时的预警
确认信号：趋势反转时的具体买卖点
星级信号：重要的买入(★买多)和卖出(★沽空)信号

主要信号解读：
趋势线 < 11：超卖区域，准备买入
趋势线 > 89：超买区域，准备卖出
★买多：明确的买入信号
★沽空：明确的卖出信号
等待底分型/顶分型：需要等待形态确认

这个指标适合用于判断中长期趋势的转折点和关键支撑阻力位。
*/

// SupportResistanceResult 支撑阻力趋势指标结果
type SupportResistanceResult struct {
	// 基础价格线
	Resistance []float64 `json:"resistance"`  // 阻力线
	Support    []float64 `json:"support"`     // 支撑线
	CenterLine []float64 `json:"center_line"` // 中线

	// 趋势指标
	TrendLine []float64 `json:"trend_line"` // 趋势线
	V11       []float64 `json:"v11"`        // 中间计算指标
	V12       []float64 `json:"v12"`        // 趋势变化率

	// 买卖信号
	BuySignals struct {
		PrepareBuy []int `json:"prepare_buy"` // 准备买入
		ReadyBuy   []int `json:"ready_buy"`   // 等待底分型
		ActualBuy  []int `json:"actual_buy"`  // 下单买入
		BuyStars   []int `json:"buy_stars"`   // ★买多信号
	} `json:"buy_signals"`

	SellSignals struct {
		PrepareSell []int `json:"prepare_sell"` // 准备卖出
		ReadySell   []int `json:"ready_sell"`   // 等待顶分型
		ActualSell  []int `json:"actual_sell"`  // 下单卖出
		SellStars   []int `json:"sell_stars"`   // ★沽空信号
	} `json:"sell_signals"`

	// 关键水平位
	TopLevel    float64 `json:"top_level"`    // 顶线 89
	BottomLevel float64 `json:"bottom_level"` // 底线 11
	MiddleLevel float64 `json:"middle_level"` // 中线 50
}

// DynamicInfo 动态信息结构（模拟DYNAINFO函数）
type DynamicInfo struct {
	Open  float64 `json:"open"`  // 今开
	High  float64 `json:"high"`  // 最高
	Low   float64 `json:"low"`   // 最低
	Close float64 `json:"close"` // 最新
}

// CalculateSupportResistance 计算支撑阻力趋势指标
func CalculateSupportResistance(data []model.DailyData, dynaInfo *DynamicInfo) *SupportResistanceResult {
	if len(data) < 55 { // 需要至少55个数据点
		return nil
	}

	result := &SupportResistanceResult{
		TopLevel:    89,
		BottomLevel: 11,
		MiddleLevel: 50,
	}

	n := len(data)

	// 提取价格数据
	highs := make([]float64, n)
	lows := make([]float64, n)
	closes := make([]float64, n)
	opens := make([]float64, n)

	for i, d := range data {
		highs[i] = d.High
		lows[i] = d.Low
		closes[i] = d.Close
		opens[i] = d.Open
	}

	// 1. 计算支撑阻力线
	calculateSupportResistanceLines(result, dynaInfo, n)

	// 2. 计算趋势线
	calculateTrendLine(result, highs, lows, closes, n)

	// 3. 计算买卖信号
	calculateBuySellSignals(result, closes, n, data)

	return result
}

// 计算支撑阻力线
func calculateSupportResistanceLines(result *SupportResistanceResult, dynaInfo *DynamicInfo, n int) {
	// H1 := MAX(DYNAINFO(3), DYNAINFO(5))
	// L1 := MIN(DYNAINFO(3), DYNAINFO(6))
	// 假设 DYNAINFO(3) 是开盘价，DYNAINFO(5) 是最高价，DYNAINFO(6) 是最低价
	h1 := math.Max(dynaInfo.Open, dynaInfo.High)
	l1 := math.Min(dynaInfo.Open, dynaInfo.Low)
	p1 := h1 - l1

	// 阻力: L1 + P1 * 7/8
	resistance := l1 + p1*7/8
	// 支撑: L1 + P1 * 0.5/8
	support := l1 + p1*0.5/8
	// 中线: (支撑 + 阻力) / 2
	center := (support + resistance) / 2

	// 将固定值扩展到整个序列
	result.Resistance = make([]float64, n)
	result.Support = make([]float64, n)
	result.CenterLine = make([]float64, n)

	for i := 0; i < n; i++ {
		result.Resistance[i] = resistance
		result.Support[i] = support
		result.CenterLine[i] = center
	}
}

// 计算趋势线
func calculateTrendLine(result *SupportResistanceResult, highs, lows, closes []float64, n int) {
	// V11 := 3*SMA((C-LLV(L,55))/(HHV(H,55)-LLV(L,55))*100,5,1)-2*SMA(SMA((C-LLV(L,55))/(HHV(H,55)-LLV(L,55))*100,5,1),3,1)

	// 计算LLV(L,55)和HHV(H,55)
	llv55 := LLV(lows, 55)
	hhv55 := HHV(highs, 55)

	// 计算 (C-LLV(L,55))/(HHV(H,55)-LLV(L,55))*100
	rawValue := make([]float64, n)
	for i := range rawValue {
		if hhv55[i] != llv55[i] {
			rawValue[i] = (closes[i] - llv55[i]) / (hhv55[i] - llv55[i]) * 100
		} else {
			rawValue[i] = 50 // 默认值
		}
	}

	// 第一层SMA
	sma1 := SMA(rawValue, 5, 1)
	// 第二层SMA
	sma2 := SMA(sma1, 3, 1)

	// 计算V11
	result.V11 = make([]float64, n)
	for i := range result.V11 {
		result.V11[i] = 3*sma1[i] - 2*sma2[i]
	}

	// 趋势线: EMA(V11,3)
	result.TrendLine = EMA(result.V11, 3)

	// V12:=(趋势线-REF(趋势线,1))/REF(趋势线,1)*100
	result.V12 = make([]float64, n)
	refTrendLine := REF(result.TrendLine, 1)
	for i := range result.V12 {
		if i > 0 && refTrendLine[i] != 0 {
			result.V12[i] = (result.TrendLine[i] - refTrendLine[i]) / refTrendLine[i] * 100
		}
	}
}

// 计算买卖信号
func calculateBuySellSignals(result *SupportResistanceResult, closes []float64, n int, data []model.DailyData) {
	// 准备买入: 趋势线<11
	for i, trend := range result.TrendLine {
		if trend < result.BottomLevel {
			result.BuySignals.PrepareBuy = append(result.BuySignals.PrepareBuy, data[i].GetTradeDate())
		}
	}

	// 等待底分型: AA:=(趋势线<11) AND FILTER((趋势线<=11),15) AND C<中线
	for i := 14; i < n; i++ {
		trend := result.TrendLine[i]
		center := result.CenterLine[i]

		// 检查趋势线是否连续15周期<=11
		filterCondition := true
		for j := i - 14; j <= i; j++ {
			if result.TrendLine[j] > result.BottomLevel {
				filterCondition = false
				break
			}
		}

		if trend < result.BottomLevel && filterCondition && closes[i] < center {
			result.BuySignals.ReadyBuy = append(result.BuySignals.ReadyBuy, data[i].GetTradeDate())
		}
	}

	// 买入条件判断
	refTrendLine := REF(result.TrendLine, 1)
	for i := 1; i < n; i++ {
		trend := result.TrendLine[i]
		prevTrend := refTrendLine[i]
		center := result.CenterLine[i]

		// BB0: REF(趋势线,1)<11 AND CROSS(趋势线,11) AND C<中线
		if prevTrend < result.BottomLevel &&
			prevTrend <= result.BottomLevel && trend > result.BottomLevel &&
			closes[i] < center {
			result.BuySignals.BuyStars = append(result.BuySignals.BuyStars, data[i].GetTradeDate())
		}

		// BB1-BB5 条件
		bb1 := prevTrend < result.BottomLevel && prevTrend > 6 && trend > result.BottomLevel
		bb2 := prevTrend < 6 && prevTrend > 3 && trend > 6
		bb3 := prevTrend < 3 && prevTrend > 1 && trend > 3
		bb4 := prevTrend < 1 && prevTrend > 0 && trend > 1
		bb5 := prevTrend < 0 && trend > 0

		bb := bb1 || bb2 || bb3 || bb4 || bb5

		if bb && closes[i] < center {
			result.BuySignals.ActualBuy = append(result.BuySignals.ActualBuy, data[i].GetTradeDate())
		}
	}

	// 准备卖出: 趋势线>89
	for i, trend := range result.TrendLine {
		if trend > result.TopLevel {
			result.SellSignals.PrepareSell = append(result.SellSignals.PrepareSell, data[i].GetTradeDate())
		}
	}

	// 等待顶分型: CC:=(趋势线>89) AND FILTER((趋势线>89),15) AND C>中线
	for i := 14; i < n; i++ {
		trend := result.TrendLine[i]
		center := result.CenterLine[i]

		// 检查趋势线是否连续15周期>89
		filterCondition := true
		for j := i - 14; j <= i; j++ {
			if result.TrendLine[j] <= result.TopLevel {
				filterCondition = false
				break
			}
		}

		if trend > result.TopLevel && filterCondition && closes[i] > center {
			result.SellSignals.ReadySell = append(result.SellSignals.ReadySell, data[i].GetTradeDate())
		}
	}

	// 卖出条件判断
	for i := 1; i < n; i++ {
		trend := result.TrendLine[i]
		prevTrend := refTrendLine[i]
		center := result.CenterLine[i]

		// DD0: REF(趋势线,1)>89 AND CROSS(89,趋势线) AND C>中线
		if prevTrend > result.TopLevel &&
			prevTrend >= result.TopLevel && trend < result.TopLevel &&
			closes[i] > center {
			result.SellSignals.SellStars = append(result.SellSignals.SellStars, data[i].GetTradeDate())
		}

		// DD1-DD5 条件
		dd1 := prevTrend > result.TopLevel && prevTrend < 94 && trend < result.TopLevel
		dd2 := prevTrend > 94 && prevTrend < 97 && trend < 94
		dd3 := prevTrend > 97 && prevTrend < 99 && trend < 97
		dd4 := prevTrend > 99 && prevTrend < 100 && trend < 99
		dd5 := prevTrend > 100 && trend < 100

		dd := dd1 || dd2 || dd3 || dd4 || dd5

		if dd && closes[i] > center {
			result.SellSignals.ActualSell = append(result.SellSignals.ActualSell, data[i].GetTradeDate())
		}
	}
}

// 获取最新信号摘要
func (r *SupportResistanceResult) GetSignalSummary() map[string]interface{} {
	summary := make(map[string]interface{})

	if len(r.TrendLine) == 0 {
		return summary
	}

	latestTrend := r.TrendLine[len(r.TrendLine)-1]
	latestCenter := r.CenterLine[len(r.CenterLine)-1]

	summary["current_trend"] = latestTrend
	summary["trend_level"] = getTrendLevel(latestTrend)
	summary["position_vs_center"] = getPositionVsCenter(latestTrend, latestCenter)
	summary["buy_signals_count"] = len(r.BuySignals.ActualBuy)
	summary["sell_signals_count"] = len(r.SellSignals.ActualSell)

	// 最近的信号
	if len(r.BuySignals.ActualBuy) > 0 {
		summary["last_buy_signal"] = r.BuySignals.ActualBuy[len(r.BuySignals.ActualBuy)-1]
	}
	if len(r.SellSignals.ActualSell) > 0 {
		summary["last_sell_signal"] = r.SellSignals.ActualSell[len(r.SellSignals.ActualSell)-1]
	}

	return summary
}

// 获取趋势级别
func getTrendLevel(trend float64) string {
	switch {
	case trend < 11:
		return "超卖区域"
	case trend >= 11 && trend < 30:
		return "弱势区域"
	case trend >= 30 && trend < 70:
		return "震荡区域"
	case trend >= 70 && trend <= 89:
		return "强势区域"
	case trend > 89:
		return "超买区域"
	default:
		return "未知区域"
	}
}

// 获取相对于中线的位置
func getPositionVsCenter(trend, center float64) string {
	if trend < center {
		return "中线下方"
	} else if trend > center {
		return "中线上方"
	} else {
		return "中线附近"
	}
}

// 使用示例
func main() {
	// 假设您已经有填充好的DailyData切片
	// var dailyData []DailyData

	// 创建动态信息（模拟DYNAINFO）
	// dynaInfo := &DynamicInfo{
	//     Open:   dailyData[len(dailyData)-1].Open,
	//     High:   dailyData[len(dailyData)-1].High,
	//     Low:    dailyData[len(dailyData)-1].Low,
	//     Close:  dailyData[len(dailyData)-1].Close,
	// }

	// 计算支撑阻力趋势指标
	// result := CalculateSupportResistance(dailyData, dynaInfo)

	// 使用结果
	// if result != nil {
	//     summary := result.GetSignalSummary()
	//     fmt.Printf("当前趋势级别: %s\n", summary["trend_level"])
	//     fmt.Printf("相对于中线位置: %s\n", summary["position_vs_center"])
	//     fmt.Printf("买入信号数量: %d\n", summary["buy_signals_count"])
	//     fmt.Printf("卖出信号数量: %d\n", summary["sell_signals_count"])
	//
	//     // 打印支撑阻力位
	//     if len(result.Support) > 0 {
	//         fmt.Printf("支撑位: %.3f\n", result.Support[0])
	//         fmt.Printf("阻力位: %.3f\n", result.Resistance[0])
	//         fmt.Printf("中线: %.3f\n", result.CenterLine[0])
	//     }
	// }
}
