package indicator

import (
	"fmt"
	"math"
	"os"
	"sort"
	"stock/internal/model"
	"strings"
)

// ChipStock 计算筹码分布指标需要实现的方法
type ChipStock interface {
	IndStock
}

// ChipDistribution 筹码分布数据结构
type ChipDistribution struct {
	Price      float64 `json:"price"`      // 价格
	Percentage float64 `json:"percentage"` // 筹码分布百分比
	Volume     int64   `json:"volume"`     // 该价位的成交量
	Accumulate float64 `json:"accumulate"` // 累计筹码分布百分比
}

// ChipDistributionResult 筹码分布计算结果
type ChipDistributionResult struct {
	Symbol        string             `json:"symbol"`        // 股票代码
	TradeDate     int                `json:"trade_date"`    // 交易日期
	Distributions []ChipDistribution `json:"distributions"` // 筹码分布数据
	AvgCost       float64            `json:"avg_cost"`      // 平均成本
	Concentration float64            `json:"concentration"` // 筹码集中度(90%筹码的价格区间/平均价格)
	ProfitRatio   float64            `json:"profit_ratio"`  // 获利盘比例
	LossRatio     float64            `json:"loss_ratio"`    // 套牢盘比例
	ActiveChips   float64            `json:"active_chips"`  // 活跃筹码比例
	DeadChips     float64            `json:"dead_chips"`    // 死筹码比例
}

// ChipDistributionCalculator 筹码分布计算器
func ChipDistributionCalculator(stocks []ChipStock, args ...int) []*ChipDistributionResult {
	if len(stocks) == 0 {
		return nil
	}

	// 默认参数：换手率衰减周期90天，价格区间数量100
	var decayPeriod, priceIntervals = 90, 100
	switch len(args) {
	case 1:
		decayPeriod = args[0]
	case 2:
		decayPeriod, priceIntervals = args[0], args[1]
	}

	var results = make([]*ChipDistributionResult, 0, len(stocks))

	for i := 0; i < len(stocks); i++ {
		if i < decayPeriod-1 {
			// 数据不足，跳过
			continue
		}

		result := calculateChipDistribution(stocks, i, decayPeriod, priceIntervals)
		if result != nil {
			results = append(results, result)
		}
	}

	return results
}

// calculateChipDistribution 计算单日筹码分布
func calculateChipDistribution(stocks []ChipStock, currentIndex, decayPeriod, priceIntervals int) *ChipDistributionResult {
	currentStock := stocks[currentIndex]
	currentHigh, currentLow, _, currentClose := currentStock.Get4Price()

	// 计算价格区间
	minPrice, maxPrice := findPriceRange(stocks, currentIndex, decayPeriod)
	if maxPrice <= minPrice {
		return nil
	}

	priceStep := (maxPrice - minPrice) / float64(priceIntervals)
	if priceStep <= 0 {
		return nil
	}

	// 初始化筹码分布数组
	distributions := make([]ChipDistribution, priceIntervals)
	totalVolume := int64(0)
	totalCost := float64(0)

	// 计算每个价格区间的筹码分布
	for i := 0; i < priceIntervals; i++ {
		price := minPrice + float64(i)*priceStep
		distributions[i].Price = price
	}

	// 遍历历史数据，计算筹码分布
	for j := currentIndex - decayPeriod + 1; j <= currentIndex; j++ {
		if j < 0 {
			continue
		}

		stock := stocks[j]
		high, low, open, close := stock.Get4Price()
		volume := stock.GetVolume()

		// 计算换手率衰减因子 (距离当前日期越远，权重越小)
		daysDiff := currentIndex - j
		decayFactor := math.Exp(-float64(daysDiff) / float64(decayPeriod) * 3) // 3是衰减系数

		// 计算该日的平均价格
		avgPrice := (high + low + open + close) / 4

		// 将成交量分配到对应的价格区间
		priceIndex := int((avgPrice - minPrice) / priceStep)
		if priceIndex >= 0 && priceIndex < priceIntervals {
			adjustedVolume := int64(float64(volume) * decayFactor)
			distributions[priceIndex].Volume += adjustedVolume
			totalVolume += adjustedVolume
			totalCost += avgPrice * float64(adjustedVolume)
		}
	}

	// 计算百分比和累计百分比
	var accumulate float64 = 0
	for i := 0; i < len(distributions); i++ {
		if totalVolume > 0 {
			distributions[i].Percentage = float64(distributions[i].Volume) / float64(totalVolume) * 100
		}
		accumulate += distributions[i].Percentage
		distributions[i].Accumulate = accumulate
	}

	// 计算平均成本
	avgCost := float64(0)
	if totalVolume > 0 {
		avgCost = totalCost / float64(totalVolume)
	}

	// 计算筹码集中度 (90%筹码的价格区间)
	concentration := calculateConcentration(distributions, avgCost)

	// 计算获利盘和套牢盘比例
	profitRatio, lossRatio := calculateProfitLossRatio(distributions, currentClose)

	// 计算活跃筹码和死筹码比例
	activeChips, deadChips := calculateActiveDeadChips(distributions, currentHigh, currentLow)

	return &ChipDistributionResult{
		Symbol:        currentStock.GetSymbol(),
		TradeDate:     currentStock.GetTradeDate(),
		Distributions: distributions,
		AvgCost:       avgCost,
		Concentration: concentration,
		ProfitRatio:   profitRatio,
		LossRatio:     lossRatio,
		ActiveChips:   activeChips,
		DeadChips:     deadChips,
	}
}

// findPriceRange 找到指定周期内的价格范围
func findPriceRange(stocks []ChipStock, currentIndex, period int) (float64, float64) {
	if currentIndex < 0 || currentIndex >= len(stocks) {
		return 0, 0
	}

	startIndex := currentIndex - period + 1
	if startIndex < 0 {
		startIndex = 0
	}

	high, low, _, _ := stocks[startIndex].Get4Price()
	minPrice, maxPrice := low, high

	for i := startIndex; i <= currentIndex; i++ {
		h, l, _, _ := stocks[i].Get4Price()
		if l < minPrice {
			minPrice = l
		}
		if h > maxPrice {
			maxPrice = h
		}
	}

	// 扩展价格范围5%，确保覆盖所有可能的价格
	priceRange := maxPrice - minPrice
	minPrice -= priceRange * 0.05
	maxPrice += priceRange * 0.05

	return minPrice, maxPrice
}

// calculateConcentration 计算筹码集中度
func calculateConcentration(distributions []ChipDistribution, avgCost float64) float64 {
	if len(distributions) == 0 || avgCost <= 0 {
		return 0
	}

	// 按筹码分布百分比排序
	sorted := make([]ChipDistribution, len(distributions))
	copy(sorted, distributions)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Percentage > sorted[j].Percentage
	})

	// 找到包含90%筹码的价格区间
	var accumulate float64 = 0
	var minPrice, maxPrice float64 = math.MaxFloat64, 0

	for _, dist := range sorted {
		if accumulate < 90 {
			accumulate += dist.Percentage
			if dist.Price < minPrice {
				minPrice = dist.Price
			}
			if dist.Price > maxPrice {
				maxPrice = dist.Price
			}
		} else {
			break
		}
	}

	if maxPrice > minPrice && avgCost > 0 {
		return (maxPrice - minPrice) / avgCost * 100
	}
	return 0
}

// calculateProfitLossRatio 计算获利盘和套牢盘比例
func calculateProfitLossRatio(distributions []ChipDistribution, currentPrice float64) (float64, float64) {
	var profitVolume, lossVolume, totalVolume float64 = 0, 0, 0

	for _, dist := range distributions {
		volume := float64(dist.Volume)
		totalVolume += volume

		if dist.Price < currentPrice {
			profitVolume += volume // 成本价低于当前价，获利
		} else {
			lossVolume += volume // 成本价高于当前价，套牢
		}
	}

	profitRatio := float64(0)
	lossRatio := float64(0)

	if totalVolume > 0 {
		profitRatio = profitVolume / totalVolume * 100
		lossRatio = lossVolume / totalVolume * 100
	}

	return profitRatio, lossRatio
}

// calculateActiveDeadChips 计算活跃筹码和死筹码比例
func calculateActiveDeadChips(distributions []ChipDistribution, currentHigh, currentLow float64) (float64, float64) {
	var activeVolume, deadVolume, totalVolume float64 = 0, 0, 0

	// 当前交易区间
	priceRange := currentHigh - currentLow
	activeRangeMin := currentLow - priceRange*0.1  // 扩展10%
	activeRangeMax := currentHigh + priceRange*0.1 // 扩展10%

	for _, dist := range distributions {
		volume := float64(dist.Volume)
		totalVolume += volume

		if dist.Price >= activeRangeMin && dist.Price <= activeRangeMax {
			activeVolume += volume // 在活跃交易区间内
		} else {
			deadVolume += volume // 在活跃交易区间外
		}
	}

	activeRatio := float64(0)
	deadRatio := float64(0)

	if totalVolume > 0 {
		activeRatio = activeVolume / totalVolume * 100
		deadRatio = deadVolume / totalVolume * 100
	}

	return activeRatio, deadRatio
}

// ConvertToTechnicalIndicator 将筹码分布结果转换为技术指标格式
func (cdr *ChipDistributionResult) ConvertToTechnicalIndicator() *model.TechnicalIndicator {
	if cdr == nil {
		return nil
	}

	return &model.TechnicalIndicator{
		Symbol:            cdr.Symbol,
		TradeDate:         cdr.TradeDate,
		ChipAvgCost:       cdr.AvgCost,
		ChipConcentration: cdr.Concentration,
		ChipProfitRatio:   cdr.ProfitRatio,
		ChipLossRatio:     cdr.LossRatio,
		ChipActiveRatio:   cdr.ActiveChips,
		ChipDeadRatio:     cdr.DeadChips,
	}
}

// CalculateChipDistribution 计算筹码分布指标，返回技术指标格式
// 参数：stocks - 股票数据切片，args - 可选参数[衰减周期, 价格区间数量]
// 默认参数：衰减周期90天，价格区间数量100
func CalculateChipDistribution(stocks []ChipStock, args ...int) []*model.TechnicalIndicator {
	results := ChipDistributionCalculator(stocks, args...)
	if len(results) == 0 {
		return nil
	}

	indicators := make([]*model.TechnicalIndicator, 0, len(results))
	for _, result := range results {
		if indicator := result.ConvertToTechnicalIndicator(); indicator != nil {
			indicators = append(indicators, indicator)
		}
	}

	return indicators
}

// ChipDistributionChart 筹码分布图表生成器
type ChipDistributionChart struct {
	Width  int // 图表宽度（字符数）
	Height int // 图表高度（行数）
}

// NewChipDistributionChart 创建新的筹码分布图表生成器
func NewChipDistributionChart(width, height int) *ChipDistributionChart {
	if width <= 0 {
		width = 60
	}
	if height <= 0 {
		height = 20
	}
	return &ChipDistributionChart{
		Width:  width,
		Height: height,
	}
}

// GenerateASCIIChart 生成ASCII字符筹码分布图
func (c *ChipDistributionChart) GenerateASCIIChart(distributions []ChipDistribution, currentPrice float64) string {
	if len(distributions) == 0 {
		return "没有筹码分布数据"
	}

	// 过滤出有效的分布数据并按价格排序
	var validDist []ChipDistribution
	for _, dist := range distributions {
		if dist.Percentage > 0 {
			validDist = append(validDist, dist)
		}
	}

	if len(validDist) == 0 {
		return "没有有效的筹码分布数据"
	}

	// 按价格排序
	sort.Slice(validDist, func(i, j int) bool {
		return validDist[i].Price < validDist[j].Price
	})

	// 找到最大百分比用于缩放
	maxPercentage := 0.0
	for _, dist := range validDist {
		if dist.Percentage > maxPercentage {
			maxPercentage = dist.Percentage
		}
	}

	if maxPercentage == 0 {
		return "筹码分布百分比为零"
	}

	// 计算价格范围
	minPrice := validDist[0].Price
	maxPrice := validDist[len(validDist)-1].Price
	priceRange := maxPrice - minPrice

	var result strings.Builder
	result.WriteString("筹码分布图 (当前价格: " + fmt.Sprintf("%.2f", currentPrice) + ")\n")
	result.WriteString(strings.Repeat("=", c.Width+15) + "\n")

	// 生成图表
	for i := c.Height - 1; i >= 0; i-- {
		// 计算当前行对应的价格
		priceLevel := minPrice + (float64(i)/float64(c.Height-1))*priceRange

		// 找到最接近这个价格的筹码分布
		var closestDist *ChipDistribution
		minDiff := math.MaxFloat64
		for j := range validDist {
			diff := math.Abs(validDist[j].Price - priceLevel)
			if diff < minDiff {
				minDiff = diff
				closestDist = &validDist[j]
			}
		}

		// 计算柱状图长度
		var barLength int
		var percentage float64
		if closestDist != nil && minDiff < priceRange/float64(c.Height) {
			percentage = closestDist.Percentage
			barLength = int((percentage / maxPercentage) * float64(c.Width))
		}

		// 生成价格标签
		priceLabel := fmt.Sprintf("%7.2f", priceLevel)

		// 标记当前价格
		marker := " "
		if math.Abs(priceLevel-currentPrice) < priceRange/float64(c.Height) {
			marker = "→"
		}

		// 生成柱状图
		bar := strings.Repeat("█", barLength)
		if barLength < c.Width {
			bar += strings.Repeat(" ", c.Width-barLength)
		}

		// 添加百分比标签
		percentageLabel := ""
		if percentage > 0 {
			percentageLabel = fmt.Sprintf(" %.1f%%", percentage)
		}

		result.WriteString(fmt.Sprintf("%s%s|%s|%s\n", priceLabel, marker, bar, percentageLabel))
	}

	result.WriteString(strings.Repeat("=", c.Width+15) + "\n")
	result.WriteString(fmt.Sprintf("价格范围: %.2f - %.2f\n", minPrice, maxPrice))
	result.WriteString(fmt.Sprintf("最大筹码密度: %.1f%%\n", maxPercentage))

	return result.String()
}

// GenerateHTMLChart 生成HTML格式的筹码分布图
func (c *ChipDistributionChart) GenerateHTMLChart(distributions []ChipDistribution, currentPrice float64, title string) string {
	if len(distributions) == 0 {
		return "<p>没有筹码分布数据</p>"
	}

	// 将价格精确到0.01并合并相同价格的数据
	priceMap := make(map[float64]ChipDistribution)

	for _, dist := range distributions {
		// 将价格精确到0.01（四舍五入到分）
		roundedPrice := math.Round(dist.Price*10) / 10

		if existing, exists := priceMap[roundedPrice]; exists {
			// 合并数据：累加成交量和百分比
			priceMap[roundedPrice] = ChipDistribution{
				Price:      roundedPrice,
				Percentage: existing.Percentage + dist.Percentage,
				Volume:     existing.Volume + dist.Volume,
				Accumulate: 0, // 累计百分比稍后重新计算
			}
		} else {
			priceMap[roundedPrice] = ChipDistribution{
				Price:      roundedPrice,
				Percentage: dist.Percentage,
				Volume:     dist.Volume,
				Accumulate: 0,
			}
		}
	}

	// 将map转换为slice并按价格排序
	var validDist []ChipDistribution
	for _, dist := range priceMap {
		validDist = append(validDist, dist)
	}

	// 按价格排序
	sort.Slice(validDist, func(i, j int) bool {
		return validDist[i].Price < validDist[j].Price
	})

	// 重新计算累计百分比
	var accumulate float64 = 0
	for i := range validDist {
		accumulate += validDist[i].Percentage
		validDist[i].Accumulate = accumulate
	}

	// 找到最大百分比用于缩放
	maxPercentage := 0.0
	for _, dist := range validDist {
		if dist.Percentage > maxPercentage {
			maxPercentage = dist.Percentage
		}
	}

	// 如果最大百分比为0，设置为1以避免除零错误
	if maxPercentage == 0 {
		maxPercentage = 1
	}

	var html strings.Builder
	html.WriteString(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>` + title + `</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .chart-container { max-width: 800px; margin: 0 auto; }
        .chart-title { text-align: center; font-size: 18px; font-weight: bold; margin-bottom: 20px; }
        .chart-row { display: flex; align-items: center; margin: 2px 0; }
        .price-label { width: 80px; text-align: right; margin-right: 10px; font-size: 12px; }
        .current-price { color: red; font-weight: bold; }
        .bar-container { flex: 1; background-color: #f0f0f0; height: 20px; position: relative; }
        .bar { background-color: #4CAF50; height: 100%; transition: width 0.3s ease; }
        .zero-bar { background-color: #e0e0e0; height: 100%; }
        .percentage-label { width: 60px; margin-left: 10px; font-size: 12px; }
        .legend { margin-top: 20px; font-size: 12px; }
        .current-price-line { border-left: 2px solid red; }
        .zero-percentage { color: #888; }
    </style>
</head>
<body>
    <div class="chart-container">
        <div class="chart-title">` + title + ` (当前价格: ` + fmt.Sprintf("%.2f", currentPrice) + `)</div>
`)

	// 计算价格范围
	var minPrice, maxPrice float64
	if len(validDist) > 0 {
		minPrice = validDist[0].Price
		maxPrice = validDist[len(validDist)-1].Price
	}

	// 生成图表行（从高价到低价）
	for i := len(validDist) - 1; i >= 0; i-- {
		dist := validDist[i]
		barWidth := (dist.Percentage / maxPercentage) * 100

		// 判断是否接近当前价格（容差为0.02）
		isCurrentPrice := math.Abs(dist.Price-currentPrice) <= 0.02
		priceClass := ""
		barContainerClass := ""
		barClass := "bar"
		percentageClass := ""

		if isCurrentPrice {
			priceClass = " current-price"
			barContainerClass = " current-price-line"
		}

		// 如果百分比为0，使用不同的样式
		if dist.Percentage == 0 {
			barClass = "zero-bar"
			percentageClass = " zero-percentage"
			barWidth = 5 // 给零百分比一个最小宽度以便显示
		}

		html.WriteString(fmt.Sprintf(`
        <div class="chart-row">
            <div class="price-label%s">%.2f</div>
            <div class="bar-container%s">
                <div class="%s" style="width: %.1f%%"></div>
            </div>
            <div class="percentage-label%s">%.1f%%</div>
        </div>`,
			priceClass,
			dist.Price,
			barContainerClass,
			barClass,
			barWidth,
			percentageClass,
			dist.Percentage))
	}

	html.WriteString(`
        <div class="legend">
            <p><strong>说明：</strong></p>
            <p>• 横轴长度表示筹码分布密度</p>
            <p>• <span style="color: red;">红色</span>标记表示当前价格附近</p>
            <p>• <span style="color: #888;">灰色</span>表示无筹码分布的价格</p>
            <p>• 价格精确到0.01，相同价格的筹码已合并</p>
            <p>• 价格范围: ` + fmt.Sprintf("%.2f - %.2f", minPrice, maxPrice) + `</p>
            <p>• 最大筹码密度: ` + fmt.Sprintf("%.1f%%", maxPercentage) + `</p>
            <p>• 总计筹码分布条目: ` + fmt.Sprintf("%d", len(validDist)) + `</p>
        </div>
    </div>
</body>
</html>`)

	return html.String()
}

// GenerateSimpleChart 生成简单的文本筹码分布图
func GenerateSimpleChart(distributions []ChipDistribution, currentPrice float64) string {
	chart := NewChipDistributionChart(50, 15)
	return chart.GenerateASCIIChart(distributions, currentPrice)
}

// GenerateDetailedChart 生成详细的筹码分布图表
func GenerateDetailedChart(distributions []ChipDistribution, currentPrice float64, width, height int) string {
	chart := NewChipDistributionChart(width, height)
	return chart.GenerateASCIIChart(distributions, currentPrice)
}

// SaveHTMLChart 保存HTML格式的筹码分布图到文件
func SaveHTMLChart(distributions []ChipDistribution, currentPrice float64, title, filename string) error {
	chart := NewChipDistributionChart(0, 0) // HTML图表不需要指定宽高
	html := chart.GenerateHTMLChart(distributions, currentPrice, title)

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(html)
	return err
}
