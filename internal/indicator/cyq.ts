// @ts-nocheck

/**
 * 筹码分布算法
 * @param {Array.<Array.<string>>} kdata K图数据 [time,open,close,high,low,volume,amount,amplitude,turnoverRate]
 * @param {number} [accuracyFactor=500] 精度因子
 * @param {number} [range] 计算范围
 */
export function CYQCalculator(kdata, accuracyFactor, range) {
    /**
     * K图数据[time,open,close,high,low,volume,amount,amplitude,turnoverRate]
     */
    this.klinedata = kdata;
    /**
     * 精度因子(纵轴刻度数)
     */
    this.fator = accuracyFactor || 150;
    /**
     * 计算K线条数
     */
    this.range = range;
}

/**
 * 计算分布及相关指标
 * @param {number} index 当前选中的K线的索引
 * @return {{x: Array.<number>, y: Array.<number>}}
 */
CYQCalculator.prototype.calc = function (index) {
    var maxprice = 0;
    var minprice = 0;
    var factor = this.fator;
    var start = this.range ? Math.max(0, index - this.range + 1) : 0;
    /**
     * K图数据[time,open,close,high,low,volume,amount,amplitude,turnoverRate]
     */
    var kdata = this.klinedata.slice(start, Math.max(1, index + 1));
    if (kdata.length === 0) throw 'invaild index';
    for (var i = 0; i < kdata.length; i++) {
        var elements = kdata[i];
        maxprice = !maxprice ? elements.high : Math.max(maxprice, elements.high);
        minprice = !minprice ? elements.low : Math.min(minprice, elements.low);
    }

    // 精度不小于0.01 产品逻辑
    var accuracy = Math.max(0.01, (maxprice - minprice) / (factor - 1));
    /**
     * 值域
     * @type {Array.<number>}
     */
    var yrange = [];
    for (var i = 0; i < factor; i++) {
        yrange.push((minprice + accuracy * i).toFixed(2) / 1);
    }
    /**
     * 横轴数据
     */
    var xdata = createNumberArray(factor);

    for (var i = 0; i < kdata.length; i++) {
        var eles = kdata[i];
        var open = eles.open,
            close = eles.close,
            high = eles.high,
            low = eles.low,
            avg = (open + close + high + low) / 4,
            turnoverRate = Math.min(1, eles.hsl / 100 || 0);
        var H = Math.floor((high - minprice) / accuracy),
            L = Math.ceil((low - minprice) / accuracy),
            // G点坐标, 一字板时, X为进度因子
            GPoint = [high == low ? factor - 1 : 2 / (high - low), Math.floor((avg - minprice) / accuracy)];
        // 衰减
        for (var n = 0; n < xdata.length; n++) {
            xdata[n] *= (1 - turnoverRate);
        }
        if (high == low) {
            // 一字板时，画矩形面积是三角形的2倍
            xdata[GPoint[1]] += GPoint[0] * turnoverRate / 2;
        } else {
            for (var j = L; j <= H; j++) {
                var curprice = minprice + accuracy * j;
                if (curprice <= avg) {
                    // 上半三角叠加分布分布
                    if (Math.abs(avg - low) < 1e-8) {
                        xdata[j] += GPoint[0] * turnoverRate;
                    } else {
                        xdata[j] += (curprice - low) / (avg - low) * GPoint[0] * turnoverRate;
                    }
                } else {
                    // 下半三角叠加分布分布
                    if (Math.abs(high - avg) < 1e-8) {
                        xdata[j] += GPoint[0] * turnoverRate;
                    } else {
                        xdata[j] += (high - curprice) / (high - avg) * GPoint[0] * turnoverRate;
                    }
                }
            }
        }
    }


    var currentprice = this.klinedata[index].close;
    var totalChips = 0;
    for (var i = 0; i < factor; i++) {
        var x = xdata[i].toPrecision(12) / 1;
        //if (x < 0) xdata[i] = 0;
        totalChips += x;
    }
    var result = new CYQData();
    result.x = xdata;
    result.y = yrange;
    result.benefitPart = result.getBenefitPart(currentprice);
    result.avgCost = getCostByChip(totalChips * 0.5).toFixed(2);
    result.percentChips = {
        '90': result.computePercentChips(0.9),
        '70': result.computePercentChips(0.7)
    };
    return result;

    /**
     * 获取指定筹码处的成本
     * @param {number} chip 堆叠筹码
     */
    function getCostByChip(chip) {
        var result = 0,
            sum = 0;
        for (var i = 0; i < factor; i++) {
            var x = xdata[i].toPrecision(12) / 1;
            if (sum + x > chip) {
                result = minprice + i * accuracy;
                break;
            }
            sum += x;
        }
        return result;
    }

    /**
     * 筹码分布数据
     */
    function CYQData() {
        /**
         * 筹码堆叠
         * @type {Array.<number>}
         */
        this.x = arguments[0];
        /**
         * 价格分布
         * @type {Array.<number>}
         */
        this.y = arguments[1];
        /**
         * 获利比例
         * @type {number}
         */
        this.benefitPart = arguments[2];
        /**
         * 平均成本
         * @type {number}
         */
        this.avgCost = arguments[3];
        /**
         * 百分比筹码
         * @type {{Object.<string, {{priceRange: number[], concentration: number}}>}}
         */
        this.percentChips = arguments[4];
        /**
         * 计算指定百分比的筹码
         * @param {number} percent 百分比大于0，小于1
         */
        this.computePercentChips = function (percent) {
            if (percent > 1 || percent < 0) throw 'argument "percent" out of range';
            var ps = [(1 - percent) / 2, (1 + percent) / 2];
            var pr = [getCostByChip(totalChips * ps[0]), getCostByChip(totalChips * ps[1])];
            return {
                priceRange: [pr[0].toFixed(2), pr[1].toFixed(2)],
                concentration: pr[0] + pr[1] === 0 ? 0 : (pr[1] - pr[0]) / (pr[0] + pr[1])
            };
        };
        /**
         * 获取指定价格的获利比例
         * @param {number} price 价格
         */
        this.getBenefitPart = function (price) {
            var below = 0;
            for (var i = 0; i < factor; i++) {
                var x = xdata[i].toPrecision(12) / 1;
                if (price >= minprice + i * accuracy) {
                    below += x;
                }
            }
            return totalChips == 0 ? 0 : below / totalChips;
        };
    }
}

/**
 * 构造数字型数组
 * @param {number} count 数组数量
 */
function createNumberArray(count) {
    var array = [];
    for (var i = 0; i < count; i++) {
        array.push(0);
    }
    return array;
}