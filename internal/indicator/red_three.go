package indicator

import (
	"math"
	"stock/internal/model"
)

// RedThreeStock 红三角抄底指标

// IndicatorResult 指标计算结果
type IndicatorResult struct {
	VAR1   []float64
	VAR2   []float64
	VAR3   []float64
	VAR4   []float64
	QUSHI  []float64
	JIANDI []float64
	MAIRU  []float64
	VAR5   []float64
	DADI   []float64
	// 为了方便使用，也返回一些关键信号
	Signals struct {
		BuySignals     []int // 买入信号的位置索引
		DaDiSignals    []int // 大底信号的位置索引
		WarningSignals []int // 注意警告信号的位置索引
	}
}

// REF - 引用前N周期的数据
func REF(data []float64, n int) []float64 {
	result := make([]float64, len(data))
	for i := range data {
		if i < n {
			result[i] = 0 // 对于前n个数据，没有足够的历史数据，返回0
		} else {
			result[i] = data[i-n]
		}
	}
	return result
}

// SMA - 平滑移动平均
func SMA(x []float64, n, m int) []float64 {
	result := make([]float64, len(x))
	if len(x) == 0 {
		return result
	}

	// 第一个值
	result[0] = x[0]

	for i := 1; i < len(x); i++ {
		// SMA公式: (M * 当前值 + (N - M) * 前一周期的SMA) / N
		if i < n {
			// 对于前n个周期，使用简单平均
			sum := 0.0
			for j := 0; j <= i; j++ {
				sum += x[j]
			}
			result[i] = sum / float64(i+1)
		} else {
			result[i] = (float64(m)*x[i] + float64(n-m)*result[i-1]) / float64(n)
		}
	}
	return result
}

// EMA - 指数移动平均
func EMA(data []float64, n int) []float64 {
	result := make([]float64, len(data))
	if len(data) == 0 {
		return result
	}

	// 平滑系数
	alpha := 2.0 / (float64(n) + 1.0)

	// 第一个值
	result[0] = data[0]

	for i := 1; i < len(data); i++ {
		result[i] = alpha*data[i] + (1-alpha)*result[i-1]
	}
	return result
}

// LLV - N周期内的最小值
func LLV(data []float64, n int) []float64 {
	result := make([]float64, len(data))
	for i := range data {
		if i < n-1 {
			// 对于前n-1个数据，计算从开始到当前的最小值
			minVal := data[0]
			for j := 0; j <= i; j++ {
				if data[j] < minVal {
					minVal = data[j]
				}
			}
			result[i] = minVal
		} else {
			// 计算最近n个周期的最小值
			minVal := data[i-n+1]
			for j := i - n + 2; j <= i; j++ {
				if data[j] < minVal {
					minVal = data[j]
				}
			}
			result[i] = minVal
		}
	}
	return result
}

// HHV - N周期内的最大值
func HHV(data []float64, n int) []float64 {
	result := make([]float64, len(data))
	for i := range data {
		if i < n-1 {
			// 对于前n-1个数据，计算从开始到当前的最大值
			maxVal := data[0]
			for j := 0; j <= i; j++ {
				if data[j] > maxVal {
					maxVal = data[j]
				}
			}
			result[i] = maxVal
		} else {
			// 计算最近n个周期的最大值
			maxVal := data[i-n+1]
			for j := i - n + 2; j <= i; j++ {
				if data[j] > maxVal {
					maxVal = data[j]
				}
			}
			result[i] = maxVal
		}
	}
	return result
}

// MA - 简单移动平均
func MA(data []float64, n int) []float64 {
	result := make([]float64, len(data))
	for i := range data {
		if i < n-1 {
			// 对于前n-1个数据，计算从开始到当前的平均值
			sum := 0.0
			for j := 0; j <= i; j++ {
				sum += data[j]
			}
			result[i] = sum / float64(i+1)
		} else {
			// 计算最近n个周期的平均值
			sum := 0.0
			for j := i - n + 1; j <= i; j++ {
				sum += data[j]
			}
			result[i] = sum / float64(n)
		}
	}
	return result
}

// MAX - 返回两个数中较大的一个
func MAX(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// IF - 条件判断函数
func IF(condition []bool, trueVal, falseVal float64) []float64 {
	result := make([]float64, len(condition))
	for i, cond := range condition {
		if cond {
			result[i] = trueVal
		} else {
			result[i] = falseVal
		}
	}
	return result
}

// CROSS - 判断A是否上穿B
func CROSS(A, B []float64) []bool {
	result := make([]bool, len(A))
	for i := 1; i < len(A); i++ {
		// A上穿B: 前一天A<=B，今天A>B
		result[i] = (A[i-1] <= B[i-1]) && (A[i] > B[i])
	}
	return result
}

// ABS - 绝对值
func ABS(data []float64) []float64 {
	result := make([]float64, len(data))
	for i, val := range data {
		result[i] = math.Abs(val)
	}
	return result
}

// RedThree 计算主要指标
func RedThree(stocks []model.DailyData) *IndicatorResult {
	if len(stocks) < 33 { // 需要至少33个数据点来计算VAR4
		return nil
	}

	result := &IndicatorResult{}

	// 提取价格数据
	opens := make([]float64, len(stocks))
	highs := make([]float64, len(stocks))
	lows := make([]float64, len(stocks))
	closes := make([]float64, len(stocks))

	for i, d := range stocks {
		highs[i], lows[i], opens[i], closes[i] = d.Get4Price()
	}

	// 1. 计算VAR1: REF((LOW+OPEN+CLOSE+HIGH)/4,1)
	priceAvg := make([]float64, len(stocks))
	for i := range stocks {
		priceAvg[i] = (lows[i] + opens[i] + closes[i] + highs[i]) / 4
	}
	result.VAR1 = REF(priceAvg, 1)

	// 2. 计算VAR2: SMA(ABS(LOW-VAR1),13,1)/SMA(MAX(LOW-VAR1,0),10,1)
	lowMinusVAR1 := make([]float64, len(stocks))
	for i := range stocks {
		lowMinusVAR1[i] = lows[i] - result.VAR1[i]
	}

	absLowMinusVAR1 := ABS(lowMinusVAR1)
	maxLowMinusVAR1 := make([]float64, len(stocks))
	for i := range stocks {
		maxLowMinusVAR1[i] = MAX(lowMinusVAR1[i], 0)
	}

	numerator := SMA(absLowMinusVAR1, 13, 1)
	denominator := SMA(maxLowMinusVAR1, 10, 1)

	result.VAR2 = make([]float64, len(stocks))
	for i := range stocks {
		if denominator[i] != 0 {
			result.VAR2[i] = numerator[i] / denominator[i]
		} else {
			result.VAR2[i] = 0
		}
	}

	// 3. 计算VAR3: EMA(VAR2,10)
	result.VAR3 = EMA(result.VAR2, 10)

	// 4. 计算VAR4: LLV(LOW,33)
	result.VAR4 = LLV(lows, 33)

	// 5. 计算QUSHI: 3*SMA(A,5,1)-2*SMA(SMA(A,5,1),9,1)
	// 其中 A = (CLOSE-LLV(LOW,27))/(HHV(HIGH,27)-LLV(LOW,27))*100
	llv27 := LLV(lows, 27)
	hhv27 := HHV(highs, 27)

	A := make([]float64, len(stocks))
	for i := range stocks {
		denominatorA := hhv27[i] - llv27[i]
		if denominatorA != 0 {
			A[i] = (closes[i] - llv27[i]) / denominatorA * 100
		} else {
			A[i] = 0
		}
	}

	smaA5 := SMA(A, 5, 1)
	smaSmaA59 := SMA(smaA5, 9, 1)

	result.QUSHI = make([]float64, len(stocks))
	for i := range stocks {
		result.QUSHI[i] = 3*smaA5[i] - 2*smaSmaA59[i]
	}

	// 6. JIANDI: 1 (常数)
	result.JIANDI = make([]float64, len(stocks))
	for i := range stocks {
		result.JIANDI[i] = 1
	}

	// 7. 计算MAIRU: IF(CROSS(QUSHI,JIANDI),100,0)
	crossSignal := CROSS(result.QUSHI, result.JIANDI)
	result.MAIRU = IF(crossSignal, 100, 0)

	// 8. 计算VAR5: EMA(IF(LOW<=VAR4,VAR3,0),3)
	ifValue := make([]float64, len(stocks))
	for i := range stocks {
		if lows[i] <= result.VAR4[i] {
			ifValue[i] = result.VAR3[i]
		} else {
			ifValue[i] = 0
		}
	}
	result.VAR5 = EMA(ifValue, 3)

	// 9. 计算DADI: IF((MA(C,5)-C)/C>0.04 AND (MA(C,10)-MA(C,5))/MA(C,5)>0.04,100,0)
	ma5 := MA(closes, 5)
	ma10 := MA(closes, 10)

	condition1 := make([]bool, len(stocks))
	condition2 := make([]bool, len(stocks))
	for i := range stocks {
		if closes[i] != 0 {
			condition1[i] = (ma5[i]-closes[i])/closes[i] > 0.04
		} else {
			condition1[i] = false
		}

		if ma5[i] != 0 {
			condition2[i] = (ma10[i]-ma5[i])/ma5[i] > 0.04
		} else {
			condition2[i] = false
		}
	}

	dadiCondition := make([]bool, len(stocks))
	for i := range stocks {
		dadiCondition[i] = condition1[i] && condition2[i]
	}
	result.DADI = IF(dadiCondition, 100, 0)

	// 收集信号
	for i, val := range result.MAIRU {
		if val == 100 {
			result.Signals.BuySignals = append(result.Signals.BuySignals, stocks[i].GetTradeDate())
		}
	}

	for i, val := range result.DADI {
		if val == 100 {
			result.Signals.DaDiSignals = append(result.Signals.DaDiSignals, stocks[i].GetTradeDate())
		}
	}

	// 计算警告信号: CROSS(100, QUSHI)
	hundred := make([]float64, len(stocks))
	for i := range stocks {
		hundred[i] = 100
	}
	warningCross := CROSS(hundred, result.QUSHI)
	for i, val := range warningCross {
		if val {
			result.Signals.WarningSignals = append(result.Signals.WarningSignals, stocks[i].GetTradeDate())
		}
	}

	return result
}
