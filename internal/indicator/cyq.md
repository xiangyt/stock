# 筹码分布图(CYQ)Canvas算法分析

## 概述
筹码分布图(Chip Distribution / Cost-Yield Quantification)是一种技术分析工具，用于显示在不同价位上的持仓成本分布。它通过横向柱状图的形式，叠加在K线图的右侧。

## 核心算法原理

### 1. 数据结构
```typescript
interface ChipData {
  price: number;        // 价格
  volume: number;       // 成交量
  percentage: number;   // 占总筹码的百分比
  profit: number;       // 盈利比例 (当前价-成本价)/成本价
}

interface CYQConfig {
  priceMin: number;     // 价格区间最小值
  priceMax: number;     // 价格区间最大值
  priceStep: number;    // 价格步长
  decay: number;        // 衰减系数 (通常0.9-0.98)
  period: number;       // 计算周期 (如90天)
}
```

### 2. 筹码计算算法

#### 2.1 换手率衰减模型
筹码分布基于一个核心假设：随着时间推移，旧筹码会逐渐被新筹码替换。

```typescript
function calculateChipDistribution(
  klineData: KLineData[],
  config: CYQConfig
): ChipData[] {
  const { decay, period, priceStep } = config;
  
  // 初始化价格区间桶
  const priceBuckets = new Map<number, number>();
  
  // 反向遍历K线数据(从最新到最旧)
  for (let i = 0; i < period && i < klineData.length; i++) {
    const kline = klineData[klineData.length - 1 - i];
    const weight = Math.pow(decay, i); // 衰减权重
    
    // 将成交量分配到价格区间
    const avgPrice = (kline.high + kline.low) / 2;
    const priceKey = Math.floor(avgPrice / priceStep) * priceStep;
    
    const currentVolume = priceBuckets.get(priceKey) || 0;
    priceBuckets.set(priceKey, currentVolume + kline.volume * weight);
  }
  
  // 转换为筹码数据数组
  const totalChips = Array.from(priceBuckets.values())
    .reduce((sum, vol) => sum + vol, 0);
    
  return Array.from(priceBuckets.entries()).map(([price, volume]) => ({
    price,
    volume,
    percentage: (volume / totalChips) * 100,
    profit: 0 // 稍后计算
  }));
}
```

#### 2.2 成交量分布算法
更精确的实现会将单根K线的成交量分散到高低价之间：

```typescript
function distributeVolumeInRange(
  low: number,
  high: number,
  volume: number,
  weight: number,
  priceBuckets: Map<number, number>,
  priceStep: number
) {
  const levels = Math.ceil((high - low) / priceStep);
  const volumePerLevel = volume / levels;
  
  for (let price = low; price <= high; price += priceStep) {
    const priceKey = Math.floor(price / priceStep) * priceStep;
    const current = priceBuckets.get(priceKey) || 0;
    priceBuckets.set(priceKey, current + volumePerLevel * weight);
  }
}
```

### 3. Canvas绘制算法

#### 3.1 坐标映射
```typescript
interface ChartDimensions {
  width: number;
  height: number;
  paddingTop: number;
  paddingBottom: number;
}

function mapPriceToY(
  price: number,
  priceMin: number,
  priceMax: number,
  dims: ChartDimensions
): number {
  const priceRange = priceMax - priceMin;
  const chartHeight = dims.height - dims.paddingTop - dims.paddingBottom;
  
  return dims.paddingTop + 
         (priceMax - price) / priceRange * chartHeight;
}

function mapVolumeToWidth(
  volume: number,
  maxVolume: number,
  maxWidth: number
): number {
  return (volume / maxVolume) * maxWidth;
}
```

#### 3.2 渐变色计算
筹码分布图使用渐变色表示盈亏状态：
- 蓝色/紫色：盈利筹码
- 橙色/红色：亏损筹码

```typescript
function getChipColor(
  currentPrice: number,
  chipPrice: number,
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  width: number,
  height: number
): CanvasGradient {
  const profit = (currentPrice - chipPrice) / chipPrice;
  
  const gradient = ctx.createLinearGradient(x, y, x + width, y);
  
  if (profit > 0) {
    // 盈利筹码 - 蓝色系
    gradient.addColorStop(0, 'rgba(59, 130, 246, 0.6)');  // 蓝色
    gradient.addColorStop(1, 'rgba(139, 92, 246, 0.3)');  // 紫色
  } else {
    // 亏损筹码 - 橙红色系
    gradient.addColorStop(0, 'rgba(251, 146, 60, 0.6)'); // 橙色
    gradient.addColorStop(1, 'rgba(239, 68, 68, 0.3)');  // 红色
  }
  
  return gradient;
}
```

#### 3.3 绘制主流程
```typescript
function drawCYQ(
  ctx: CanvasRenderingContext2D,
  chipData: ChipData[],
  currentPrice: number,
  dims: ChartDimensions,
  priceRange: { min: number; max: number }
) {
  ctx.clearRect(0, 0, dims.width, dims.height);
  
  const maxVolume = Math.max(...chipData.map(d => d.volume));
  const maxBarWidth = dims.width * 0.8; // 留20%空间
  
  // 按价格从低到高排序
  chipData.sort((a, b) => a.price - b.price);
  
  chipData.forEach(chip => {
    const y = mapPriceToY(chip.price, priceRange.min, priceRange.max, dims);
    const barWidth = mapVolumeToWidth(chip.volume, maxVolume, maxBarWidth);
    const barHeight = 3; // 柱子高度
    
    // 绘制渐变柱状图
    ctx.fillStyle = getChipColor(
      currentPrice,
      chip.price,
      ctx,
      0,
      y - barHeight / 2,
      barWidth,
      barHeight
    );
    
    ctx.fillRect(0, y - barHeight / 2, barWidth, barHeight);
    
    // 可选：绘制边框
    ctx.strokeStyle = 'rgba(255, 255, 255, 0.2)';
    ctx.lineWidth = 0.5;
    ctx.strokeRect(0, y - barHeight / 2, barWidth, barHeight);
  });
  
  // 绘制当前价格指示线
  drawCurrentPriceLine(ctx, currentPrice, priceRange, dims);
}
```

#### 3.4 当前价格线
```typescript
function drawCurrentPriceLine(
  ctx: CanvasRenderingContext2D,
  currentPrice: number,
  priceRange: { min: number; max: number },
  dims: ChartDimensions
) {
  const y = mapPriceToY(currentPrice, priceRange.min, priceRange.max, dims);
  
  ctx.save();
  ctx.strokeStyle = '#000';
  ctx.lineWidth = 1;
  ctx.setLineDash([5, 3]);
  
  ctx.beginPath();
  ctx.moveTo(0, y);
  ctx.lineTo(dims.width, y);
  ctx.stroke();
  
  ctx.restore();
}
```

### 4. 性能优化

#### 4.1 数据缓存
```typescript
class CYQCalculator {
  private cache: Map<string, ChipData[]> = new Map();
  
  calculate(klineData: KLineData[], config: CYQConfig): ChipData[] {
    const cacheKey = this.getCacheKey(klineData, config);
    
    if (this.cache.has(cacheKey)) {
      return this.cache.get(cacheKey)!;
    }
    
    const result = calculateChipDistribution(klineData, config);
    this.cache.set(cacheKey, result);
    
    return result;
  }
  
  private getCacheKey(data: KLineData[], config: CYQConfig): string {
    const lastDataHash = data[data.length - 1]?.timestamp || 0;
    return `${lastDataHash}_${config.period}_${config.decay}`;
  }
}
```

#### 4.2 离屏Canvas
```typescript
class CYQRenderer {
  private offscreenCanvas: HTMLCanvasElement;
  private offscreenCtx: CanvasRenderingContext2D;
  
  constructor(width: number, height: number) {
    this.offscreenCanvas = document.createElement('canvas');
    this.offscreenCanvas.width = width;
    this.offscreenCanvas.height = height;
    this.offscreenCtx = this.offscreenCanvas.getContext('2d')!;
  }
  
  render(chipData: ChipData[], currentPrice: number, dims: ChartDimensions) {
    // 在离屏canvas上绘制
    drawCYQ(this.offscreenCtx, chipData, currentPrice, dims, priceRange);
    
    // 复制到主canvas
    return this.offscreenCanvas;
  }
}
```

## 5. 关键参数说明

### 衰减系数(decay)
- **范围**: 0.9 ~ 0.98
- **含义**: 每天筹码的保留率
- **影响**:
    - 接近1.0: 历史筹码保留更久，分布更平滑
    - 接近0.9: 近期筹码权重更大，分布更集中

### 计算周期(period)
- **常用值**: 60天、90天、120天
- **影响**: 周期越长，计算越全面但速度越慢

### 价格步长(priceStep)
- **计算方式**: 通常为 (最高价 - 最低价) / 100
- **影响**: 步长越小，分布越细腻但计算量越大

## 6. 实际应用技巧

### 6.1 盈亏比计算
```typescript
function calculateProfitRatio(chipData: ChipData[], currentPrice: number) {
  let profitChips = 0;
  let lossChips = 0;
  
  chipData.forEach(chip => {
    if (chip.price < currentPrice) {
      profitChips += chip.volume;
    } else {
      lossChips += chip.volume;
    }
  });
  
  return {
    profitRatio: profitChips / (profitChips + lossChips) * 100,
    lossRatio: lossChips / (profitChips + lossChips) * 100
  };
}
```

### 6.2 集中度计算
```typescript
function calculateConcentration(chipData: ChipData[], range: number) {
  // 计算价格区间内的筹码集中度
  chipData.sort((a, b) => b.volume - a.volume);
  
  const topChips = chipData.slice(0, Math.floor(chipData.length * range));
  const totalVolume = chipData.reduce((sum, c) => sum + c.volume, 0);
  const topVolume = topChips.reduce((sum, c) => sum + c.volume, 0);
  
  return (topVolume / totalVolume) * 100;
}
```

## 7. 与K线图的集成

```typescript
class StockChart {
  private cyqCanvas: HTMLCanvasElement;
  private klineCanvas: HTMLCanvasElement;
  
  syncScroll(scrollY: number) {
    // 同步K线图和筹码分布图的垂直滚动
    this.cyqCanvas.style.transform = `translateY(${scrollY}px)`;
  }
  
  syncPriceScale(priceMin: number, priceMax: number) {
    // 确保两个图表使用相同的价格刻度
    this.redrawCYQ(priceMin, priceMax);
  }
}
```

## 总结

筹码分布图的核心在于：
1. **换手率衰减模型** - 模拟筹码随时间的自然流转
2. **成交量分配** - 将历史成交量合理分配到各价格区间
3. **渐变色渲染** - 直观展示盈亏状态
4. **性能优化** - 缓存和离屏渲染确保流畅度

这个算法结合了统计学、金融学和计算机图形学，是股票技术分析中的重要工具。
