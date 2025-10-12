package indicator

import "stock/internal/model"

// IndStock 可以计算指标的股票数据
type IndStock interface {
	Get4Price() (float64, float64, float64, float64) // 获取最高、最低、开盘、收盘价
	GetSymbol() string                               // 获取股票代码
	GetTradeDate() int                               // 获取交易日期
}

// MACDStock 计算macd指标需要实现的方法
type MACDStock interface {
	IndStock
}

// MACD 计算MACD
func MACD(stocks []MACDStock, last *model.TechnicalIndicator, args ...int) []*model.TechnicalIndicator {
	var a1, a2, a3 float64 = 12, 26, 9
	switch len(args) {
	case 1:
		a1 = float64(args[1])
	case 2:
		a1, a2 = float64(args[1]), float64(args[2])
	case 3:
		a1, a2, a3 = float64(args[1]), float64(args[2]), float64(args[3])
	}

	var ins = make([]*model.TechnicalIndicator, 0, len(stocks))
	for i := 0; i < len(stocks); i++ {
		stock := stocks[i]
		_, _, _, end := stock.Get4Price()
		if i == 0 && last == nil {
			last = &model.TechnicalIndicator{
				Symbol:    stock.GetSymbol(),
				TradeDate: stock.GetTradeDate(),
				Macd:      0,
				MacdEma1:  end,
				MacdEma2:  end,
				MacdDif:   0,
				MacdDea:   0,
			}
		} else {
			var ema1, ema2 float64
			//12日EMA EMA(12) = 2/(12+1) * 今日收盘价(12) + 11/(12+1) * 昨日EMA(12)
			//26日EMA EMA(26) = 2/(26+1) * 今日收盘价(26) + 25/(26+1) * 昨日EMA(26)
			ema1 = end*2/(a1+1) + last.MacdEma1*(a1-1)/(a1+1)
			ema2 = end*2/(a2+1) + last.MacdEma2*(a2-1)/(a2+1)

			//DIFF = EMA(12) - EMA(26)
			//DEA = 2/(9+1) * 今日DIFF + 8/(9+1) * 昨日DEA
			//MACD柱线 = 2 * (DIFF-DEA)
			diff := ema1 - ema2
			dea := diff*2/(a3+1) + last.MacdDea*(a3-1)/(a3+1)
			last = &model.TechnicalIndicator{
				Symbol:    stock.GetSymbol(),
				TradeDate: stock.GetTradeDate(),
				Macd:      2 * (diff - dea),
				MacdEma1:  ema1,
				MacdEma2:  ema2,
				MacdDif:   diff,
				MacdDea:   dea,
			}
		}
		ins = append(ins, last)
	}
	return ins
}
