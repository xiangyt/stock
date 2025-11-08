package indicator

import (
	"fmt"
	"math"
	"stock/internal/model"
)

// KLineData K线数据接口，对应TS中的kdata数组元素
type KLineData interface {
	GetOpen() float64         // 开盘价
	GetClose() float64        // 收盘价
	GetHigh() float64         // 最高价
	GetLow() float64          // 最低价
	GetVolume() int64         // 成交量
	GetTurnoverRate() float64 // 换手率(hsl)，百分比形式
}

// DailyDataAdapter 为model.DailyData提供KLineData接口适配
type DailyDataAdapter struct {
	model.DailyData
	hsl float64
}

func (d *DailyDataAdapter) GetOpen() float64         { return d.Open }
func (d *DailyDataAdapter) GetClose() float64        { return d.Close }
func (d *DailyDataAdapter) GetHigh() float64         { return d.High }
func (d *DailyDataAdapter) GetLow() float64          { return d.Low }
func (d *DailyDataAdapter) GetVolume() int64         { return d.Volume }
func (d *DailyDataAdapter) GetTurnoverRate() float64 { return d.hsl }

// NewDailyDataAdapter 创建DailyData适配器
func NewDailyDataAdapter(data model.DailyData) *DailyDataAdapter {
	return &DailyDataAdapter{
		DailyData: data,
		hsl:       math.Round(float64(data.Volume)/243277980*100) / 100,
	}
}

// CYQTSCalculator 基于TypeScript算法的筹码分布计算器
type CYQTSCalculator struct {
	klineData      []KLineData // K线数据
	accuracyFactor int         // 精度因子(纵轴刻度数)
	rangeLimit     int         // 计算K线条数限制
}

// CYQTSData 筹码分布数据结果
type CYQTSData struct {
	X            []float64              `json:"x"`             // 筹码堆叠数据
	Y            []float64              `json:"y"`             // 价格分布数据
	BenefitPart  float64                `json:"benefit_part"`  // 获利比例
	AvgCost      float64                `json:"avg_cost"`      // 平均成本
	PercentChips map[string]PercentChip `json:"percent_chips"` // 百分比筹码
}

// PercentChip 百分比筹码数据
type PercentChip struct {
	PriceRange    [2]float64 `json:"price_range"`   // 价格区间 [下限, 上限]
	Concentration float64    `json:"concentration"` // 集中度
}

// NewCYQTSCalculator 创建新的筹码分布计算器
// kdata: K线数据数组
// accuracyFactor: 精度因子(纵轴刻度数)，默认150
// rangeLimit: 计算范围，nil表示使用全部数据
func NewCYQTSCalculator(kdata []KLineData, accuracyFactor int, rangeLimit int) *CYQTSCalculator {
	if accuracyFactor <= 0 {
		accuracyFactor = 150
	}

	return &CYQTSCalculator{
		klineData:      kdata,
		accuracyFactor: accuracyFactor,
		rangeLimit:     rangeLimit,
	}
}

// Calc 计算分布及相关指标
// index: 当前选中的K线的索引
func (calc *CYQTSCalculator) Calc(index int) (*CYQTSData, error) {
	if index < 0 || index >= len(calc.klineData) {
		return nil, fmt.Errorf("invalid index: %d", index)
	}

	var maxPrice, minPrice float64
	factor := calc.accuracyFactor

	// 确定计算范围
	start := 0
	if calc.rangeLimit > 0 && index > calc.rangeLimit-1 {
		start = index - calc.rangeLimit + 1
	}

	// 获取计算范围内的K线数据
	kdata := calc.klineData[start : index+1]
	if len(kdata) == 0 {
		return nil, fmt.Errorf("invalid index")
	}

	// 计算价格区间
	for i, data := range kdata {
		high := data.GetHigh()
		low := data.GetLow()

		if i == 0 {
			maxPrice = high
			minPrice = low
		} else {
			maxPrice = math.Max(maxPrice, high)
			minPrice = math.Min(minPrice, low)
		}
	}

	// 精度不小于0.01 产品逻辑
	accuracy := math.Max(0.01, (maxPrice-minPrice)/float64(factor-1))

	// 构建价格区间数组(值域)
	yRange := make([]float64, factor)
	for i := 0; i < factor; i++ {
		price := minPrice + accuracy*float64(i)
		yRange[i] = math.Round(price*100) / 100 // 保留2位小数
	}

	// 初始化横轴数据(筹码堆叠)
	xData := make([]float64, factor)

	// 遍历K线数据计算筹码分布
	for _, data := range kdata {
		open := data.GetOpen()
		close := data.GetClose()
		high := data.GetHigh()
		low := data.GetLow()
		avg := (open + close + high + low) / 4
		turnoverRate := math.Min(1, data.GetTurnoverRate()/100) // 转换为小数

		// 计算价格索引
		H := int(math.Floor((high - minPrice) / accuracy))
		L := int(math.Ceil((low - minPrice) / accuracy))

		// G点坐标, 一字板时, X为进度因子
		var gPoint [2]float64
		if high == low {
			gPoint[0] = float64(factor - 1)
		} else {
			gPoint[0] = 2 / (high - low)
		}
		gPoint[1] = math.Floor((avg - minPrice) / accuracy)

		// 衰减现有筹码
		for n := 0; n < len(xData); n++ {
			xData[n] *= (1 - turnoverRate)
		}

		// 分配新筹码
		if high == low {
			// 一字板时，画矩形面积是三角形的2倍
			idx := int(gPoint[1])
			if idx >= 0 && idx < len(xData) {
				xData[idx] += gPoint[0] * turnoverRate / 2
			}
		} else {
			// 三角形分布
			for j := L; j <= H && j < factor; j++ {
				if j < 0 {
					continue
				}

				curPrice := minPrice + accuracy*float64(j)

				if curPrice <= avg {
					// 上半三角叠加分布
					if math.Abs(avg-low) < 1e-8 {
						xData[j] += gPoint[0] * turnoverRate
					} else {
						xData[j] += (curPrice - low) / (avg - low) * gPoint[0] * turnoverRate
					}
				} else {
					// 下半三角叠加分布
					if math.Abs(high-avg) < 1e-8 {
						xData[j] += gPoint[0] * turnoverRate
					} else {
						xData[j] += (high - curPrice) / (high - avg) * gPoint[0] * turnoverRate
					}
				}
			}
		}
	}

	// 计算总筹码
	totalChips := 0.0
	for i := 0; i < factor; i++ {
		// 精度处理，类似JavaScript的toPrecision(12)
		x := math.Round(xData[i]*1e12) / 1e12
		xData[i] = x
		totalChips += x
	}

	currentPrice := calc.klineData[index].GetClose()

	// 创建结果对象
	result := &CYQTSData{
		X: xData,
		Y: yRange,
	}

	// 计算获利比例
	result.BenefitPart = calc.getBenefitPart(currentPrice, minPrice, accuracy, factor, xData, totalChips)

	// 计算平均成本
	result.AvgCost = math.Round(calc.getCostByChip(totalChips*0.5, minPrice, accuracy, factor, xData)*100) / 100

	// 计算百分比筹码
	result.PercentChips = map[string]PercentChip{
		"90": calc.computePercentChips(0.9, totalChips, minPrice, accuracy, factor, xData),
		"70": calc.computePercentChips(0.7, totalChips, minPrice, accuracy, factor, xData),
	}

	return result, nil
}

// getCostByChip 获取指定筹码处的成本
func (calc *CYQTSCalculator) getCostByChip(chip, minPrice, accuracy float64, factor int, xData []float64) float64 {
	result := 0.0
	sum := 0.0

	for i := 0; i < factor; i++ {
		x := math.Round(xData[i]*1e12) / 1e12
		if sum+x > chip {
			result = minPrice + float64(i)*accuracy
			break
		}
		sum += x
	}

	return result
}

// getBenefitPart 获取指定价格的获利比例
func (calc *CYQTSCalculator) getBenefitPart(price, minPrice, accuracy float64, factor int, xData []float64, totalChips float64) float64 {
	below := 0.0

	for i := 0; i < factor; i++ {
		x := math.Round(xData[i]*1e12) / 1e12
		if price >= minPrice+float64(i)*accuracy {
			below += x
		}
	}

	if totalChips == 0 {
		return 0
	}
	return below / totalChips
}

// computePercentChips 计算指定百分比的筹码
func (calc *CYQTSCalculator) computePercentChips(percent, totalChips, minPrice, accuracy float64, factor int, xData []float64) PercentChip {
	if percent > 1 || percent < 0 {
		return PercentChip{}
	}

	ps := []float64{(1 - percent) / 2, (1 + percent) / 2}
	pr := []float64{
		calc.getCostByChip(totalChips*ps[0], minPrice, accuracy, factor, xData),
		calc.getCostByChip(totalChips*ps[1], minPrice, accuracy, factor, xData),
	}

	concentration := 0.0
	if pr[0]+pr[1] != 0 {
		concentration = (pr[1] - pr[0]) / (pr[0] + pr[1])
	}

	return PercentChip{
		PriceRange:    [2]float64{math.Round(pr[0]*100) / 100, math.Round(pr[1]*100) / 100},
		Concentration: concentration,
	}
}
