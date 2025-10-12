package indicator

import (
	"stock/internal/model"
)

// KDJStock 计算kdj指标需要实现的方法
type KDJStock interface {
	IndStock
	GetKDJ() (float64, float64, float64) // 获取KDJ三个值
	SetKDJ(float64, float64, float64)    // 暂存KDJ三个值
}

// KDJBase kdj指标简单结构
type KDJBase struct {
	IndStock
	Indicator model.TechnicalIndicator
}

func (b *KDJBase) GetKDJ() (float64, float64, float64) {
	return b.Indicator.KdjK, b.Indicator.KdjD, b.Indicator.KdjJ
}

func (b *KDJBase) SetKDJ(k, d, j float64) {
	b.Indicator.KdjK, b.Indicator.KdjD, b.Indicator.KdjJ = k, d, j
}

// KDJ 计算KDJ数据
func KDJ(stocks []KDJStock, args ...int) []*model.TechnicalIndicator {
	var start, n = 0, 9
	var m1, m2 float64 = 3, 3
	switch len(args) {
	case 1:
		start = args[0]
	case 2:
		start, n = args[0], args[1]
	case 3:
		start, n, m1 = args[0], args[1], float64(args[2])
	case 4:
		start, n, m1, m2 = args[0], args[1], float64(args[2]), float64(args[3])

	}
	if start < 0 {
		// 从start开始计算
		start = 0
	}

	var ins = make([]*model.TechnicalIndicator, 0, len(stocks)-start)
	for i := start; i < len(stocks); i++ {
		stock := stocks[i]
		vk, vd, vj := stock.GetKDJ()
		if !(vk == 0 && vd == 0 && vj == 0) {
			continue
		}

		if i == 0 {
			vk, vd, vj = 50, 50, 50
		} else {
			var offset int
			if i+1 > n {
				offset = i + 1 - n
			}

			high, low, _, _ := stocks[offset].Get4Price()
			for j := offset; j <= i; j++ {
				high1, low1, _, _ := stocks[j].Get4Price()
				if low1 < low {
					low = low1
				}

				if high1 > high {
					high = high1
				}
			}

			var rsv float64 = 100
			if high > low {
				_, _, _, end := stock.Get4Price()
				rsv = (end - low) / (high - low) * 100
			}
			var k, d, j float64
			vk2, vd2, _ := stocks[i-1].GetKDJ()
			k = vk2*(m1-1)/m1 + rsv/m1
			d = vd2*(m2-1)/m2 + k/m2
			j = 3*k - 2*d
			vk, vd, vj = k, d, j
		}
		stock.SetKDJ(vk, vd, vj)
		ins = append(ins, &model.TechnicalIndicator{
			Symbol:    stock.GetSymbol(),
			TradeDate: stock.GetTradeDate(),
			KdjK:      vk,
			KdjD:      vd,
			KdjJ:      vj,
		})
	}
	return ins
}
