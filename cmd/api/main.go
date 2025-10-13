package main

import (
	"fmt"
	"log"
	"os"
	"stock/internal/indicator"
	"time"

	"stock/internal/collector"
	"stock/internal/config"
	"stock/internal/logger"
)

func main() {
	// 初始化配置
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 初始化全局日志器
	logger.InitGlobalLogger(cfg.Log)

	// 创建数据采集器
	collector := collector.GetCollectorFactory(logger.GetGlobalLogger()).GetTongHuaShunCollector()
	if err := collector.Connect(); err != nil {
		log.Fatalf("Failed to connect to data source: %v", err)
	}

	if daily, err := collector.GetDailyKLine("001208.SZ", time.Time{}, time.Time{}); err != nil {
		log.Fatalf("GetDailyKLine err: %s", err.Error())
	} else {
		//var list = make([]indicator.IndStock, 0, len(daily))
		//for _, data := range daily {
		//	data := data
		//	list = append(list, data)
		//}
		//result := indicator.RedThree(list)
		//if result != nil {
		//	fmt.Printf("买入信号出现在: %v\n", result.Signals.BuySignals)
		//	fmt.Printf("大底信号出现在: %v\n", result.Signals.DaDiSignals)
		//	fmt.Printf("警告信号出现在: %v\n", result.Signals.WarningSignals)
		//}

		//result := indicator.CalculateComplexIndicator(daily)
		//if result != nil {
		//	fmt.Printf("极底信号出现次数: %v\n", result.Signals.ExtremeBottom)
		//	fmt.Printf("金叉信号出现次数: %v\n", result.Signals.GoldenCross)
		//	fmt.Printf("抄底信号出现次数: %v\n", result.Signals.BottomFishing)
		//}

		// 计算时间控制指标
		//result := indicator.CalculateTimeControlIndicator(daily)
		//
		//if result != nil {
		//	fmt.Printf("红色柱状图数量: %d\n", len(result.RedSticks))
		//	fmt.Printf("绿色柱状图数量: %d\n", len(result.GreenSticks))
		//
		//	// 打印最新的指标值
		//	if len(result.A) > 0 {
		//		fmt.Printf("最新A值: %.2f\n", result.A[len(result.A)-1])
		//		fmt.Printf("最新A1值: %.2f\n", result.A1[len(result.A1)-1])
		//	}
		//}

		dynaInfo := &indicator.DynamicInfo{
			Open:  daily[len(daily)-1].Open,
			High:  daily[len(daily)-1].High,
			Low:   daily[len(daily)-1].Low,
			Close: daily[len(daily)-1].Close,
		}

		result := indicator.CalculateSupportResistance(daily, dynaInfo)

		if result != nil {
			summary := result.GetSignalSummary()
			fmt.Printf("当前趋势级别: %s\n", summary["trend_level"])
			fmt.Printf("相对于中线位置: %s\n", summary["position_vs_center"])
			fmt.Printf("买入信号数量: %d\n", summary["buy_signals_count"])
			fmt.Printf("卖出信号数量: %d\n", summary["sell_signals_count"])

			// 打印支撑阻力位
			if len(result.Support) > 0 {
				fmt.Printf("支撑位: %.3f\n", result.Support[0])
				fmt.Printf("阻力位: %.3f\n", result.Resistance[0])
				fmt.Printf("中线: %.3f\n", result.CenterLine[0])
			}
		}
	}
}
