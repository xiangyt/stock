package indicator

import (
	"math"
	"stock/internal/model"
)

// ComplexIndicatorResult 复杂指标计算结果
type ComplexIndicatorResult struct {
	// 基础线
	Overbought  []float64 `json:"overbought"`   // 超买线
	Oversold    []float64 `json:"oversold"`     // 超卖线
	MinValue    []float64 `json:"min_value"`    // 最小值
	MaxValue    []float64 `json:"max_value"`    // 最大值
	WaveLine    []float64 `json:"wave_line"`    // 波动线
	AverageLine []float64 `json:"average_line"` // 平均线

	// 状态信息
	Info         []bool `json:"info"`          // 信息
	Strengthen   []bool `json:"strengthen"`    // 走强
	Weaken       []bool `json:"weaken"`        // 走弱
	VolumeSignal []bool `json:"volume_signal"` // 量

	// 子指标
	RSI5            []float64 `json:"rsi5"`
	ADX             []float64 `json:"adx"`
	WR10            []float64 `json:"wr10"`
	BestBuy         []float64 `json:"best_buy"`
	RiskCoefficient []float64 `json:"risk_coefficient"`

	// 信号点
	Signals struct {
		ExtremeBottom  []int `json:"extreme_bottom"`  // 极底 √
		Rise           []int `json:"rise"`            // 升 √
		Top            []int `json:"top"`             // 顶 ×
		Down           []int `json:"down"`            // 下 ×
		BuildPosition  []int `json:"build_position"`  // 建仓 ？
		Escape         []int `json:"escape"`          // 逃 ？
		Bottom         []int `json:"bottom"`          // 见底 √
		AbsoluteBottom []int `json:"absolute_bottom"` // 绝底 √
		SeeRise        []int `json:"see_rise"`        // 见涨 √
		MustRise       []int `json:"must_rise"`       // 必涨 ×
		BottomFishing  []int `json:"bottom_fishing"`  // 抄底 缺失
		GoldenCross    []int `json:"golden_cross"`    // 金叉 √
	} `json:"signals"`
}

// CalculateComplexIndicator 计算复杂指标
func CalculateComplexIndicator(data []model.DailyData) *ComplexIndicatorResult {
	if len(data) < 38 { // 需要足够的数据计算各种指标
		return nil
	}

	result := &ComplexIndicatorResult{}

	// 提取价格和成交量数据
	opens := make([]float64, len(data))
	highs := make([]float64, len(data))
	lows := make([]float64, len(data))
	closes := make([]float64, len(data))
	volumes := make([]int64, len(data))

	for i, d := range data {
		opens[i] = d.Open
		highs[i] = d.High
		lows[i] = d.Low
		closes[i] = d.Close
		volumes[i] = d.Volume
	}

	// 1. 计算基础线
	calculateBasicLines(result, opens, highs, lows, closes, volumes)

	// 2. 计算RSI和ADX相关指标
	calculateRSIAndADX(result, highs, lows, closes)

	// 3. 计算最佳买入指标
	calculateBestBuyIndicator(result, highs, lows, closes)

	// 4. 计算风险系数和抄底信号
	calculateRiskCoefficient(result, highs, lows, closes)

	// 5. 计算MACD和KDJ金叉
	calculateGoldenCrossSignals(result, highs, lows, closes, data)

	// 6. 计算各种交易信号（完整版）
	calculateAllSignalsComplete(result, data)

	// 7. 计算绝底、见涨、必涨信号
	calculateSpecialSignals(result, data)

	return result
}

// 计算基础线
func calculateBasicLines(result *ComplexIndicatorResult, opens, highs, lows, closes []float64, volumes []int64) {
	n := len(closes)

	// 超买超卖线 (常数)
	result.Overbought = make([]float64, n)
	result.Oversold = make([]float64, n)
	for i := range result.Overbought {
		result.Overbought[i] = 3.2
		result.Oversold[i] = 0.5
	}

	// 最小值和最大值
	result.MinValue = LLV(lows, 10)
	result.MaxValue = HHV(highs, 25)

	// 波动线: EMA((CLOSE-最小值)/(最大值-最小值)*4,4)
	waveRaw := make([]float64, n)
	for i := range waveRaw {
		if result.MaxValue[i] != result.MinValue[i] {
			waveRaw[i] = (closes[i] - result.MinValue[i]) / (result.MaxValue[i] - result.MinValue[i]) * 4
		} else {
			waveRaw[i] = 0
		}
	}
	result.WaveLine = EMA(waveRaw, 4)

	// 平均线: EMA(波动线,3)
	result.AverageLine = EMA(result.WaveLine, 3)

	// 信息: 平均线>=REF(平均线,1)
	result.Info = make([]bool, n)
	refAverage := REF(result.AverageLine, 1)
	for i := range result.Info {
		if i > 0 {
			result.Info[i] = result.AverageLine[i] >= refAverage[i]
		}
	}

	// 走强: CLOSE>MA(CLOSE,20) AND CLOSE>MA(CLOSE,5)
	ma5 := MA(closes, 5)
	ma20 := MA(closes, 20)
	result.Strengthen = make([]bool, n)
	for i := range result.Strengthen {
		result.Strengthen[i] = closes[i] > ma20[i] && closes[i] > ma5[i]
	}

	// 走弱: CLOSE<MA(CLOSE,10) AND CLOSE<MA(CLOSE,5)
	ma10 := MA(closes, 10)
	result.Weaken = make([]bool, n)
	for i := range result.Weaken {
		result.Weaken[i] = closes[i] < ma10[i] && closes[i] < ma5[i]
	}

	// 量: VOL>MA(VOL,5)
	volumeFloats := make([]float64, n)
	for i, v := range volumes {
		volumeFloats[i] = float64(v)
	}
	maVol5 := MA(volumeFloats, 5)
	result.VolumeSignal = make([]bool, n)
	for i := range result.VolumeSignal {
		result.VolumeSignal[i] = volumeFloats[i] > maVol5[i]
	}
}

// 计算RSI和ADX指标
func calculateRSIAndADX(result *ComplexIndicatorResult, highs, lows, closes []float64) {
	n := len(closes)

	// RSI5计算
	lc := REF(closes, 1)
	priceChange := make([]float64, n)
	absPriceChange := make([]float64, n)
	for i := range priceChange {
		if i > 0 {
			priceChange[i] = math.Max(closes[i]-lc[i], 0)
			absPriceChange[i] = math.Abs(closes[i] - lc[i])
		}
	}

	smaGain := SMA(priceChange, 5, 1)
	smaLoss := SMA(absPriceChange, 5, 1)

	result.RSI5 = make([]float64, n)
	for i := range result.RSI5 {
		if smaLoss[i] != 0 {
			result.RSI5[i] = (smaGain[i] / smaLoss[i]) * 100
		} else {
			result.RSI5[i] = 50
		}
	}

	// ADX计算 (简化版)
	result.ADX = make([]float64, n)
	for i := range result.ADX {
		if i < 14 {
			result.ADX[i] = 0
		} else {
			// 简化的ADX计算
			tr := 0.0
			for j := i - 13; j <= i; j++ {
				tr1 := highs[j] - lows[j]
				tr2 := math.Abs(highs[j] - closes[j-1])
				tr3 := math.Abs(lows[j] - closes[j-1])
				tr += math.Max(tr1, math.Max(tr2, tr3))
			}
			result.ADX[i] = tr / 14
		}
	}

	// WR10计算
	result.WR10 = make([]float64, n)
	hhv10 := HHV(highs, 10)
	llv10 := LLV(lows, 10)
	for i := range result.WR10 {
		if hhv10[i] != llv10[i] {
			result.WR10[i] = (100 * (hhv10[i] - closes[i])) / (hhv10[i] - llv10[i])
		} else {
			result.WR10[i] = 50
		}
	}
}

// 计算最佳买入指标
func calculateBestBuyIndicator(result *ComplexIndicatorResult, highs, lows, closes []float64) {
	n := len(closes)

	// AV: (RSI5 + ADX)
	av := make([]float64, n)
	for i := range av {
		av[i] = result.RSI5[i] + result.ADX[i]
	}

	// ZCJL: (RSI5 - WR10)
	zcjl := make([]float64, n)
	for i := range zcjl {
		zcjl[i] = result.RSI5[i] - result.WR10[i]
	}

	// 最佳买入: (AV + ZCJL)
	result.BestBuy = make([]float64, n)
	for i := range result.BestBuy {
		result.BestBuy[i] = av[i] + zcjl[i]
	}
}

// 计算风险系数
func calculateRiskCoefficient(result *ComplexIndicatorResult, highs, lows, closes []float64) {
	n := len(closes)

	// 相对强弱指标计算
	lc := REF(closes, 1)
	rsi1 := calculateRSI(closes, lc, 3)
	rsi2 := calculateRSI(closes, lc, 5)
	rsi3 := calculateRSI(closes, lc, 8)

	relativeStrength := make([]float64, n)
	for i := range relativeStrength {
		relativeStrength[i] = 0.5*rsi1[i] + 0.31*rsi2[i] + 0.19*rsi3[i]
	}

	// 短线波段计算
	wave1 := calculateStochastic(closes, highs, lows, 8, 3)
	wave2 := calculateStochastic(closes, highs, lows, 8, 5)
	wave3 := calculateStochastic(closes, highs, lows, 8, 8)

	shortWave := make([]float64, n)
	for i := range shortWave {
		shortWave[i] = 0.5*wave1[i] + 0.31*wave2[i] + 0.19*wave3[i]
	}

	// 风险系数
	result.RiskCoefficient = make([]float64, n)
	for i := range result.RiskCoefficient {
		result.RiskCoefficient[i] = 0.5*relativeStrength[i] + 0.5*shortWave[i]
	}
}

// 计算金叉信号
func calculateGoldenCrossSignals(result *ComplexIndicatorResult, highs, lows, closes []float64, stocks []model.DailyData) {
	n := len(closes)

	// MACD金叉
	ema12 := EMA(closes, 12)
	ema26 := EMA(closes, 26)
	dif := make([]float64, n)
	for i := range dif {
		dif[i] = (ema12[i] - ema26[i]) * 100
	}
	dea := EMA(dif, 9)

	// KDJ金叉
	llv9 := LLV(lows, 9)
	hhv9 := HHV(highs, 9)
	rsv := make([]float64, n)
	for i := range rsv {
		if hhv9[i] != llv9[i] {
			rsv[i] = (closes[i] - llv9[i]) / (hhv9[i] - llv9[i]) * 100
		} else {
			rsv[i] = 50
		}
	}
	k := SMA(rsv, 9, 3)
	d := SMA(k, 9, 3)

	// 金叉信号
	for i := 1; i < n; i++ {
		// MACD金叉且KDJ金叉
		macdGolden := dif[i] > dea[i] && dif[i-1] <= dea[i-1]
		kdjGolden := k[i] > d[i] && k[i-1] <= d[i-1]

		if macdGolden && kdjGolden {
			result.Signals.GoldenCross = append(result.Signals.GoldenCross, stocks[i].GetTradeDate())
		}
	}
}

// 计算特殊信号：绝底、见涨、必涨
func calculateSpecialSignals(result *ComplexIndicatorResult, data []model.DailyData) {
	n := len(data)
	opens := make([]float64, n)
	highs := make([]float64, n)
	lows := make([]float64, n)
	closes := make([]float64, n)
	volumes := make([]int64, n)

	for i, d := range data {
		opens[i] = d.Open
		highs[i] = d.High
		lows[i] = d.Low
		closes[i] = d.Close
		volumes[i] = d.Volume
	}

	// 计算绝底信号: DRAWTEXT(C-O>=0 AND O/L>1.05 AND L<=LLV(L,20),1,'绝底')
	calculateAbsoluteBottomSignals(result, opens, highs, lows, closes, n, data)

	// 计算见涨信号: HXN相关逻辑
	calculateSeeRiseSignals(result, opens, highs, lows, closes, volumes, n, data)

	// 计算必涨信号: DRAWTEXT(CROSS(LOW,BZTD),1.25,'必涨')
	calculateMustRiseSignals(result, opens, highs, lows, closes, n, data)
}

// 计算绝底信号
func calculateAbsoluteBottomSignals(result *ComplexIndicatorResult, opens, highs, lows, closes []float64, n int,
	data []model.DailyData) {
	// 绝底条件: C-O>=0 AND O/L>1.05 AND L<=LLV(L,20)
	llv20 := LLV(lows, 20)

	for i := 0; i < n; i++ {
		condition1 := closes[i] >= opens[i] // C-O>=0
		condition2 := false
		if lows[i] != 0 {
			condition2 = opens[i]/lows[i] > 1.05 // O/L>1.05
		}
		condition3 := lows[i] <= llv20[i] // L<=LLV(L,20)

		if condition1 && condition2 && condition3 {
			result.Signals.AbsoluteBottom = append(result.Signals.AbsoluteBottom, data[i].GetTradeDate())
		}
	}
}

// 计算见涨信号
func calculateSeeRiseSignals(result *ComplexIndicatorResult, opens, highs, lows, closes []float64, volumes []int64,
	n int, data []model.DailyData) {
	// HXN:=IF(CLOSE/REF(CLOSE,1)>1.05 AND HIGH/CLOSE<1.01 AND IF(CLOSE>REF(CLOSE,1),88,0)>0, 91, 0);
	// DRAWTEXT(HXN>90 AND VOL>REF(VOL,1) AND CLOSE>REF(CLOSE,1) AND COUNT(HXN>90,30)=1,平均线-0.25, '见涨')

	refCloses := REF(closes, 1)
	refVolumes := make([]float64, n)
	for i, vol := range volumes {
		refVolumes[i] = float64(vol)
	}
	refVolumes = REF(refVolumes, 1)

	hxn := make([]float64, n)
	for i := 1; i < n; i++ {
		condition1 := closes[i]/refCloses[i] > 1.05 // CLOSE/REF(CLOSE,1)>1.05
		condition2 := false
		if closes[i] != 0 {
			condition2 = highs[i]/closes[i] < 1.01 // HIGH/CLOSE<1.01
		}
		condition3 := closes[i] > refCloses[i] // CLOSE>REF(CLOSE,1)

		if condition1 && condition2 && condition3 {
			hxn[i] = 91
		} else {
			hxn[i] = 0
		}
	}

	// 计算COUNT(HXN>90,30)
	hxnCount := make([]int, n)
	for i := 0; i < n; i++ {
		count := 0
		start := i - 29
		if start < 0 {
			start = 0
		}
		for j := start; j <= i; j++ {
			if hxn[j] > 90 {
				count++
			}
		}
		hxnCount[i] = count
	}

	for i := 1; i < n; i++ {
		condition1 := hxn[i] > 90
		condition2 := float64(volumes[i]) > refVolumes[i] // VOL>REF(VOL,1)
		condition3 := closes[i] > refCloses[i]            // CLOSE>REF(CLOSE,1)
		condition4 := hxnCount[i] == 1                    // COUNT(HXN>90,30)=1

		if condition1 && condition2 && condition3 && condition4 {
			result.Signals.SeeRise = append(result.Signals.SeeRise, data[i].GetTradeDate())
		}
	}
}

// 计算必涨信号
func calculateMustRiseSignals(result *ComplexIndicatorResult, opens, highs, lows, closes []float64, n int,
	data []model.DailyData) {
	// HXJZ:=MA((2*CLOSE+HIGH+LOW)/4,5);
	// BZTD:=HXJZ*89/100;
	// DRAWTEXT(CROSS(LOW,BZTD),1.25,'必涨')

	// 计算HXJZ: MA((2*CLOSE+HIGH+LOW)/4,5)
	hxjzInput := make([]float64, n)
	for i := range hxjzInput {
		hxjzInput[i] = (2*closes[i] + highs[i] + lows[i]) / 4
	}
	hxjz := MA(hxjzInput, 5)

	// 计算BZTD: HXJZ*89/100
	bztd := make([]float64, n)
	for i := range bztd {
		bztd[i] = hxjz[i] * 89 / 100
	}

	// 计算CROSS(LOW,BZTD)
	for i := 1; i < n; i++ {
		// CROSS条件: 前一天LOW>=BZTD, 今天LOW<BZTD
		crossCondition := lows[i-1] >= bztd[i-1] && lows[i] < bztd[i]
		if crossCondition {
			result.Signals.MustRise = append(result.Signals.MustRise, data[i].GetTradeDate())
		}
	}
}

// 完整的信号计算（包含之前的所有信号）
func calculateAllSignalsComplete(result *ComplexIndicatorResult, data []model.DailyData) {
	n := len(data)
	closes := make([]float64, n)
	opens := make([]float64, n)
	lows := make([]float64, n)
	volumes := make([]int64, n)

	for i, d := range data {
		closes[i] = d.Close
		opens[i] = d.Open
		lows[i] = d.Low
		volumes[i] = d.Volume
	}

	refInfo1 := REFBool(result.Info, 1)
	refInfo2 := REFBool(result.Info, 2)
	refInfo3 := REFBool(result.Info, 3)
	refStrengthen1 := REFBool(result.Strengthen, 1)

	// 计算建仓买点信号
	calculateBuildPositionSignals(result, data)

	// 计算逃亡信号
	calculateEscapeSignals(result, data)

	// 计算见底信号
	calculateBottomSignals(result, data)

	for i := 3; i < n; i++ {
		// D: 极底信号
		dCondition := result.Info[i] && !refInfo1[i] && !refInfo2[i] && !refInfo3[i] && result.AverageLine[i] < 0.5
		if dCondition {
			result.Signals.ExtremeBottom = append(result.Signals.ExtremeBottom, data[i].GetTradeDate())
		}

		// S: 升信号
		sCondition := result.Info[i] && !refInfo1[i] && !refInfo2[i] && !refInfo3[i] &&
			result.Strengthen[i] && !refStrengthen1[i] && result.VolumeSignal[i]
		if sCondition {
			result.Signals.Rise = append(result.Signals.Rise, data[i].GetTradeDate())
		}

		// 抄底信号
		if result.RiskCoefficient[i] < 20 && i > 0 && lows[i] >= lows[i-1] && closes[i] > lows[i] {
			result.Signals.BottomFishing = append(result.Signals.BottomFishing, data[i].GetTradeDate())
		}
	}
}

// 计算建仓买点信号
func calculateBuildPositionSignals(result *ComplexIndicatorResult, data []model.DailyData) {
	n := len(data)
	closes := make([]float64, n)

	for i, d := range data {
		closes[i] = d.Close
	}

	// 最佳买入选股:=IF(CROSS(最佳买入,0),1,0)
	bestBuyCross := make([]float64, n)
	refBestBuy := REF(result.BestBuy, 1)
	for i := 1; i < n; i++ {
		if result.BestBuy[i] > 0 && refBestBuy[i] <= 0 {
			bestBuyCross[i] = 1
		}
	}

	// VAR5:=SMA(最佳买入选股,3,1)
	var5 := SMA(bestBuyCross, 3, 1)
	// VAR6:=SMA(VAR5,3,1)
	var6 := SMA(var5, 3, 1)
	// VAR7:=SMA(VAR6,3,1)
	var7 := SMA(var6, 3, 1)

	// 建仓买点:=IF(CROSS(VAR6,VAR7) AND (VAR6<40),5,0)
	refVar6 := REF(var6, 1)
	refVar7 := REF(var7, 1)

	for i := 1; i < n; i++ {
		crossCondition := var6[i] > var7[i] && refVar6[i] <= refVar7[i]
		var6Condition := var6[i] < 40

		if crossCondition && var6Condition {
			result.Signals.BuildPosition = append(result.Signals.BuildPosition, data[i].GetTradeDate())
		}
	}
}

// 计算逃亡信号
func calculateEscapeSignals(result *ComplexIndicatorResult, data []model.DailyData) {
	n := len(data)
	closes := make([]float64, n)

	for i, d := range data {
		closes[i] = d.Close
	}

	// VAR8:=REF(CLOSE,2)
	var8 := REF(closes, 2)

	// 会员:=SMA(MAX(CLOSE-VAR8,0),7,1)/SMA(ABS(CLOSE-VAR8),7,1)*100
	closeMinusVar8 := make([]float64, n)
	absCloseMinusVar8 := make([]float64, n)
	for i := range closeMinusVar8 {
		closeMinusVar8[i] = math.Max(closes[i]-var8[i], 0)
		absCloseMinusVar8[i] = math.Abs(closes[i] - var8[i])
	}

	numerator := SMA(closeMinusVar8, 7, 1)
	denominator := SMA(absCloseMinusVar8, 7, 1)

	member := make([]float64, n)
	for i := range member {
		if denominator[i] != 0 {
			member[i] = (numerator[i] / denominator[i]) * 100
		}
	}

	// 逃亡:=IF(会员< REF(会员,1) AND 会员>79,会员,0)
	refMember := REF(member, 1)
	for i := 1; i < n; i++ {
		if member[i] < refMember[i] && member[i] > 79 {
			result.Signals.Escape = append(result.Signals.Escape, data[i].GetTradeDate())
		}
	}
}

// 计算见底信号
func calculateBottomSignals(result *ComplexIndicatorResult, data []model.DailyData) {
	n := len(data)
	opens := make([]float64, n)
	highs := make([]float64, n)
	lows := make([]float64, n)
	closes := make([]float64, n)

	for i, d := range data {
		opens[i] = d.Open
		highs[i] = d.High
		lows[i] = d.Low
		closes[i] = d.Close
	}

	// DRAWTEXT(88>0 AND REF(O,1)/REF(C,1)>1.04 AND REF(L,1)<=688 AND O>REF(C,1)AND C<REF(O,1)AND C/O>=1.01,1,'见底')
	refOpens := REF(opens, 1)
	refCloses := REF(closes, 1)
	refLows := REF(lows, 1)

	for i := 1; i < n; i++ {
		condition1 := refOpens[i]/refCloses[i] > 1.04 // REF(O,1)/REF(C,1)>1.04
		condition2 := refLows[i] <= 688               // REF(L,1)<=688
		condition3 := opens[i] > refCloses[i]         // O>REF(C,1)
		condition4 := closes[i] < refOpens[i]         // C<REF(O,1)
		condition5 := false
		if opens[i] != 0 {
			condition5 = closes[i]/opens[i] >= 1.01 // C/O>=1.01
		}

		if condition1 && condition2 && condition3 && condition4 && condition5 {
			result.Signals.Bottom = append(result.Signals.Bottom, data[i].GetTradeDate())
		}
	}
}

// GetSignalStats 获取信号统计信息
func (r *ComplexIndicatorResult) GetSignalStats() map[string]interface{} {
	stats := make(map[string]interface{})

	stats["extreme_bottom_count"] = len(r.Signals.ExtremeBottom)
	stats["rise_count"] = len(r.Signals.Rise)
	stats["build_position_count"] = len(r.Signals.BuildPosition)
	stats["escape_count"] = len(r.Signals.Escape)
	stats["bottom_count"] = len(r.Signals.Bottom)
	stats["absolute_bottom_count"] = len(r.Signals.AbsoluteBottom)
	stats["see_rise_count"] = len(r.Signals.SeeRise)
	stats["must_rise_count"] = len(r.Signals.MustRise)
	stats["bottom_fishing_count"] = len(r.Signals.BottomFishing)
	stats["golden_cross_count"] = len(r.Signals.GoldenCross)

	// 最近信号位置
	if len(r.Signals.ExtremeBottom) > 0 {
		stats["last_extreme_bottom"] = r.Signals.ExtremeBottom[len(r.Signals.ExtremeBottom)-1]
	}
	if len(r.Signals.AbsoluteBottom) > 0 {
		stats["last_absolute_bottom"] = r.Signals.AbsoluteBottom[len(r.Signals.AbsoluteBottom)-1]
	}
	if len(r.Signals.SeeRise) > 0 {
		stats["last_see_rise"] = r.Signals.SeeRise[len(r.Signals.SeeRise)-1]
	}
	if len(r.Signals.MustRise) > 0 {
		stats["last_must_rise"] = r.Signals.MustRise[len(r.Signals.MustRise)-1]
	}

	return stats
}

// 辅助函数：计算RSI
func calculateRSI(closes, lc []float64, period int) []float64 {
	n := len(closes)
	gains := make([]float64, n)
	losses := make([]float64, n)

	for i := 1; i < n; i++ {
		change := closes[i] - lc[i]
		if change > 0 {
			gains[i] = change
		} else {
			losses[i] = -change
		}
	}

	avgGain := SMA(gains, period, 1)
	avgLoss := SMA(losses, period, 1)

	rsi := make([]float64, n)
	for i := range rsi {
		if avgLoss[i] != 0 {
			rsi[i] = 100 - (100 / (1 + avgGain[i]/avgLoss[i]))
		} else {
			rsi[i] = 100
		}
	}

	return rsi
}

// 辅助函数：计算随机指标
func calculateStochastic(closes, highs, lows []float64, lookback, smooth int) []float64 {
	n := len(closes)
	result := make([]float64, n)

	hhv := HHV(highs, lookback)
	llv := LLV(lows, lookback)

	for i := range result {
		if hhv[i] != llv[i] {
			result[i] = 100 * (closes[i] - llv[i]) / (hhv[i] - llv[i])
		} else {
			result[i] = 50
		}
	}

	return SMA(result, smooth, 1)
}

// 辅助函数：REF for bool slice
func REFBool(data []bool, n int) []bool {
	result := make([]bool, len(data))
	for i := range data {
		if i < n {
			result[i] = false
		} else {
			result[i] = data[i-n]
		}
	}
	return result
}
