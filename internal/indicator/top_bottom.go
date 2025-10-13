package indicator

import (
	"math"
	"stock/internal/model"
	"stock/internal/utils"
	"time"
)

// TimeControlIndicatorResult 时间控制指标结果
type TimeControlIndicatorResult struct {
	AS1 []float64 `json:"as1"` // 时间控制因子
	AS2 []float64 `json:"as2"`
	AD2 []float64 `json:"ad2"`
	AS3 []float64 `json:"as3"`
	AD3 []float64 `json:"ad3"`
	AS4 []float64 `json:"as4"`
	AD4 []float64 `json:"ad4"`
	AS5 []float64 `json:"as5"`
	AD5 []float64 `json:"ad5"`
	AS6 []float64 `json:"as6"`
	AD6 []float64 `json:"ad6"`
	AS7 []float64 `json:"as7"`
	AD7 []float64 `json:"ad7"`
	AS8 []float64 `json:"as8"`
	AD8 []float64 `json:"ad8"`
	A   []float64 `json:"a"`  // 主要指标线A (红色)
	A1  []float64 `json:"a1"` // 次要指标线A1 (绿色)

	// 图形化结果
	RedSticks   []StickData `json:"red_sticks"`   // 红色柱状图
	GreenSticks []StickData `json:"green_sticks"` // 绿色柱状图
}

// StickData 柱状图数据结构
type StickData struct {
	Index int     `json:"index"`
	Value float64 `json:"value"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

// 计算时间控制指标
func CalculateTimeControlIndicator(data []model.DailyData) *TimeControlIndicatorResult {
	if len(data) < 58 { // 需要至少58个数据点来计算MA(58)
		return nil
	}

	result := &TimeControlIndicatorResult{}
	n := len(data)

	// 提取价格数据
	highs := make([]float64, n)
	lows := make([]float64, n)
	closes := make([]float64, n)
	dates := make([]time.Time, n)

	for i, d := range data {
		highs[i] = d.High
		lows[i] = d.Low
		closes[i] = d.Close
		// 假设DailyData中有Time字段，如果没有需要根据实际情况调整
		// dates[i] = d.Time
		dates[i], _ = utils.ParseTradeDate(d.GetTradeDate())
	}

	// 1. 计算时间控制因子 AS1
	result.AS1 = calculateAS1(dates)

	// 2. 计算AS2和AD2
	result.AS2 = calculateAS2(lows, result.AS1)
	result.AD2 = calculateAD2(highs, result.AS1)

	// 3. 计算AS3和AD3
	result.AS3 = calculateAS3(lows, result.AS2, result.AS1)
	result.AD3 = calculateAD3(highs, result.AD2, result.AS1)

	// 4. 计算AS4和AD4
	result.AS4 = calculateAS4(closes, result.AS3, result.AS1)
	result.AD4 = calculateAD4(closes, result.AD3, result.AS1)

	// 5. 计算AS5和AD5
	result.AS5 = calculateAS5(lows, result.AS1)
	result.AD5 = calculateAD5(highs, result.AS1)

	// 6. 计算AS6和AD6
	result.AS6 = calculateAS6(result.AS4, result.AS1)
	result.AD6 = calculateAD6(result.AD4, result.AS1)

	// 7. 计算AS7和AD7
	result.AS7 = calculateAS7(closes, result.AS1)
	result.AD7 = calculateAD7(closes, result.AS1)

	// 8. 计算AS8和AD8
	result.AS8 = calculateAS8(lows, result.AS5, result.AS4, result.AS6, result.AS7, result.AS1)
	result.AD8 = calculateAD8(highs, result.AD5, result.AD4, result.AD6, result.AD7, result.AS1)

	// 9. 计算最终指标A和A1
	result.A = calculateA(result.AS8, result.AS1)
	result.A1 = calculateA1(result.AD8, result.AS1)

	// 10. 生成柱状图数据
	result.RedSticks = calculateRedSticks(result.A, data)
	result.GreenSticks = calculateGreenSticks(result.A1, data)

	return result
}

// 计算时间控制因子 AS1: IF(YEAR>=2038 AND MONTH>=1,0,1)
func calculateAS1(dates []time.Time) []float64 {
	as1 := make([]float64, len(dates))
	for i, date := range dates {
		year := date.Year()
		month := date.Month()
		if year >= 2038 && month >= 1 {
			as1[i] = 0
		} else {
			as1[i] = 1
		}
	}
	return as1
}

// 计算AS2: REF(LOW,1)*AS1
func calculateAS2(lows []float64, as1 []float64) []float64 {
	refLow := REF(lows, 1)
	as2 := make([]float64, len(lows))
	for i := range as2 {
		as2[i] = refLow[i] * as1[i]
	}
	return as2
}

// 计算AD2: REF(HIGH,1)*AS1
func calculateAD2(highs []float64, as1 []float64) []float64 {
	refHigh := REF(highs, 1)
	ad2 := make([]float64, len(highs))
	for i := range ad2 {
		ad2[i] = refHigh[i] * as1[i]
	}
	return ad2
}

// 计算AS3: SMA(ABS(LOW-AS2),3,1)/SMA(MAX(LOW-AS2,0),3,1)*100*AS1
func calculateAS3(lows, as2, as1 []float64) []float64 {
	n := len(lows)
	as3 := make([]float64, n)

	// 计算 LOW - AS2
	lowMinusAS2 := make([]float64, n)
	for i := range lowMinusAS2 {
		lowMinusAS2[i] = lows[i] - as2[i]
	}

	// 计算 ABS(LOW-AS2)
	absLowMinusAS2 := ABS(lowMinusAS2)

	// 计算 MAX(LOW-AS2,0)
	maxLowMinusAS2 := make([]float64, n)
	for i := range maxLowMinusAS2 {
		maxLowMinusAS2[i] = math.Max(lowMinusAS2[i], 0)
	}

	// 计算SMA
	numerator := SMA(absLowMinusAS2, 3, 1)
	denominator := SMA(maxLowMinusAS2, 3, 1)

	// 计算最终结果
	for i := range as3 {
		if denominator[i] != 0 {
			as3[i] = (numerator[i] / denominator[i] * 100) * as1[i]
		} else {
			as3[i] = 0
		}
	}

	return as3
}

// 计算AD3: SMA(ABS(HIGH-AD2),3,1)/SMA(MAX(HIGH-AD2,0),3,1)*100*AS1
func calculateAD3(highs, ad2, as1 []float64) []float64 {
	n := len(highs)
	ad3 := make([]float64, n)

	// 计算 HIGH - AD2
	highMinusAD2 := make([]float64, n)
	for i := range highMinusAD2 {
		highMinusAD2[i] = highs[i] - ad2[i]
	}

	// 计算 ABS(HIGH-AD2)
	absHighMinusAD2 := ABS(highMinusAD2)

	// 计算 MAX(HIGH-AD2,0)
	maxHighMinusAD2 := make([]float64, n)
	for i := range maxHighMinusAD2 {
		maxHighMinusAD2[i] = math.Max(highMinusAD2[i], 0)
	}

	// 计算SMA
	numerator := SMA(absHighMinusAD2, 3, 1)
	denominator := SMA(maxHighMinusAD2, 3, 1)

	// 计算最终结果
	for i := range ad3 {
		if denominator[i] != 0 {
			ad3[i] = (numerator[i] / denominator[i] * 100) * as1[i]
		} else {
			ad3[i] = 0
		}
	}

	return ad3
}

// 计算AS4: EMA(IF(CLOSE*1.3,AS3*10,AS3/10),3)*AS1
func calculateAS4(closes, as3, as1 []float64) []float64 {
	n := len(closes)
	as4Input := make([]float64, n)

	for i := range as4Input {
		if closes[i]*1.3 > 0 { // 条件判断
			as4Input[i] = as3[i] * 10
		} else {
			as4Input[i] = as3[i] / 10
		}
	}

	as4 := EMA(as4Input, 3)

	// 乘以AS1
	for i := range as4 {
		as4[i] = as4[i] * as1[i]
	}

	return as4
}

// 计算AD4: EMA(IF(CLOSE*1.3,AD3*10,AD3/10),3)*AS1
func calculateAD4(closes, ad3, as1 []float64) []float64 {
	n := len(closes)
	ad4Input := make([]float64, n)

	for i := range ad4Input {
		if closes[i]*1.3 > 0 { // 条件判断
			ad4Input[i] = ad3[i] * 10
		} else {
			ad4Input[i] = ad3[i] / 10
		}
	}

	ad4 := EMA(ad4Input, 3)

	// 乘以AS1
	for i := range ad4 {
		ad4[i] = ad4[i] * as1[i]
	}

	return ad4
}

// 计算AS5: LLV(LOW,30)*AS1
func calculateAS5(lows, as1 []float64) []float64 {
	llvLow := LLV(lows, 30)
	as5 := make([]float64, len(lows))
	for i := range as5 {
		as5[i] = llvLow[i] * as1[i]
	}
	return as5
}

// 计算AD5: HHV(HIGH,30)*AS1
func calculateAD5(highs, as1 []float64) []float64 {
	hhvHigh := HHV(highs, 30)
	ad5 := make([]float64, len(highs))
	for i := range ad5 {
		ad5[i] = hhvHigh[i] * as1[i]
	}
	return ad5
}

// 计算AS6: HHV(AS4,30)*AS1
func calculateAS6(as4, as1 []float64) []float64 {
	hhvAS4 := HHV(as4, 30)
	as6 := make([]float64, len(as4))
	for i := range as6 {
		as6[i] = hhvAS4[i] * as1[i]
	}
	return as6
}

// 计算AD6: LLV(AD4,30)*AS1
func calculateAD6(ad4, as1 []float64) []float64 {
	llvAD4 := LLV(ad4, 30)
	ad6 := make([]float64, len(ad4))
	for i := range ad6 {
		ad6[i] = llvAD4[i] * as1[i]
	}
	return ad6
}

// 计算AS7: IF(MA(CLOSE,58),1,0)*AS1
func calculateAS7(closes, as1 []float64) []float64 {
	ma58 := MA(closes, 58)
	as7 := make([]float64, len(closes))
	for i := range as7 {
		if ma58[i] > 0 { // 条件判断
			as7[i] = 1 * as1[i]
		} else {
			as7[i] = 0
		}
	}
	return as7
}

// 计算AD7: IF(MA(CLOSE,58),1,0)*AS1
func calculateAD7(closes, as1 []float64) []float64 {
	// AD7的计算逻辑与AS7相同
	return calculateAS7(closes, as1)
}

// 计算AS8: EMA(IF(LOW<=AS5,(AS4+AS6*2)/2,0),3)/618*AS7*AS1
func calculateAS8(lows, as5, as4, as6, as7, as1 []float64) []float64 {
	n := len(lows)
	as8Input := make([]float64, n)

	for i := range as8Input {
		if lows[i] <= as5[i] {
			as8Input[i] = (as4[i] + as6[i]*2) / 2
		} else {
			as8Input[i] = 0
		}
	}

	as8 := EMA(as8Input, 3)

	// 除以618并乘以AS7和AS1
	for i := range as8 {
		as8[i] = (as8[i] / 618) * as7[i] * as1[i]
	}

	return as8
}

// 计算AD8: EMA(IF(HIGH>=AD5,(AD4+AD6*2)/2,0),3)/618*AD7*AS1
func calculateAD8(highs, ad5, ad4, ad6, ad7, as1 []float64) []float64 {
	n := len(highs)
	ad8Input := make([]float64, n)

	for i := range ad8Input {
		if highs[i] >= ad5[i] {
			ad8Input[i] = (ad4[i] + ad6[i]*2) / 2
		} else {
			ad8Input[i] = 0
		}
	}

	ad8 := EMA(ad8Input, 3)

	// 除以618并乘以AD7和AS1
	for i := range ad8 {
		ad8[i] = (ad8[i] / 618) * ad7[i] * as1[i]
	}

	return ad8
}

// 计算A: IF(AS8>100,100,AS8)*AS1
func calculateA(as8, as1 []float64) []float64 {
	a := make([]float64, len(as8))
	for i := range a {
		value := as8[i]
		if value > 100 {
			value = 100
		}
		a[i] = value * as1[i]
	}
	return a
}

// 计算A1: IF(AD8>50,50,AD8)*AS1
func calculateA1(ad8, as1 []float64) []float64 {
	a1 := make([]float64, len(ad8))
	for i := range a1 {
		value := ad8[i]
		if value > 50 {
			value = 50
		}
		a1[i] = value * as1[i]
	}
	return a1
}

// 计算红色柱状图: STICKLINE(A>-150,0,A,8,0)
func calculateRedSticks(a []float64, data []model.DailyData) []StickData {
	var sticks []StickData
	for i, value := range a {
		if value > -150 {
			sticks = append(sticks, StickData{
				Index: data[i].GetTradeDate(),
				Value: value,
				Start: 0,
				End:   value,
			})
		}
	}
	return sticks
}

// 计算绿色柱状图: STICKLINE(A1>0,0,15*A1,8,0)
func calculateGreenSticks(a1 []float64, data []model.DailyData) []StickData {
	var sticks []StickData
	for i, value := range a1 {
		if value > 0 {
			sticks = append(sticks, StickData{
				Index: data[i].GetTradeDate(),
				Value: value,
				Start: 0,
				End:   15 * value,
			})
		}
	}
	return sticks
}
