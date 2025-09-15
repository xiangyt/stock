<template>
  <div class="stock-detail">
    <el-card v-loading="loading">
      <template #header>
        <div class="card-header">
          <span>{{ stockDetail?.name }} ({{ code }})</span>
          <el-button @click="$router.back()">
            <el-icon><ArrowLeft /></el-icon>
            返回
          </el-button>
        </div>
      </template>

      <div v-if="stockDetail" class="detail-content">
        <!-- 基本信息 -->
        <el-row :gutter="20">
          <el-col :span="12">
            <el-descriptions title="基本信息" :column="2" border>
              <el-descriptions-item label="股票代码">{{ stockDetail.ts_code }}</el-descriptions-item>
              <el-descriptions-item label="股票名称">{{ stockDetail.name }}</el-descriptions-item>
              <el-descriptions-item label="所属行业">{{ stockDetail.industry }}</el-descriptions-item>
              <el-descriptions-item label="上市市场">
                <el-tag :type="stockDetail.market === '上海' ? 'primary' : 'success'">
                  {{ stockDetail.market }}
                </el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="上市日期">{{ stockDetail.list_date }}</el-descriptions-item>
              <el-descriptions-item label="状态">
                <el-tag :type="stockDetail.list_status === 'L' ? 'success' : 'warning'">
                  {{ stockDetail.list_status === 'L' ? '正常' : '其他' }}
                </el-tag>
              </el-descriptions-item>
            </el-descriptions>
          </el-col>
          <el-col :span="12">
            <div class="action-buttons">
              <el-button type="primary" @click="loadKLineData">
                <el-icon><TrendCharts /></el-icon>
                查看K线图
              </el-button>
              <el-button type="warning" @click="refreshKLineData" :loading="refreshing">
                <el-icon><RefreshRight /></el-icon>
                刷新K线数据
              </el-button>
              <el-button type="info" @click="checkDataStatus">
                <el-icon><InfoFilled /></el-icon>
                数据状态
              </el-button>
              <el-button type="success" @click="loadFinancialData">
                <el-icon><Document /></el-icon>
                财务数据
              </el-button>
              <el-button type="info" @click="loadRealtimeData">
                <el-icon><Refresh /></el-icon>
                实时数据
              </el-button>
            </div>
          </el-col>
        </el-row>

        <!-- K线图 -->
        <el-card v-if="showKLine" class="chart-card">
          <template #header>
            <span>K线图 (最近30天)</span>
          </template>
          <div v-if="klineData.length > 0" class="chart-container">
            <v-chart :option="klineOption" style="height: 400px;" />
          </div>
          <div v-else class="no-data">暂无K线数据</div>
        </el-card>

        <!-- 财务数据 -->
        <el-card v-if="showFinancial" class="financial-card">
          <template #header>
            <span>财务数据</span>
          </template>
          <el-table :data="financialData" stripe style="width: 100%">
            <el-table-column prop="report_date" label="报告期" width="120" />
            <el-table-column prop="roe" label="ROE(%)" width="100" />
            <el-table-column prop="roa" label="ROA(%)" width="100" />
            <el-table-column prop="gross_profit_margin" label="毛利率(%)" width="120" />
            <el-table-column prop="net_profit_margin" label="净利率(%)" width="120" />
            <el-table-column prop="debt_ratio" label="负债率(%)" width="120" />
          </el-table>
        </el-card>

        <!-- 实时数据 -->
        <el-card v-if="showRealtime" class="realtime-card">
          <template #header>
            <span>实时数据</span>
          </template>
          <div v-if="realtimeData" class="realtime-info">
            <el-row :gutter="20">
              <el-col :span="6">
                <div class="info-item">
                  <div class="label">当前价格</div>
                  <div class="value price">¥{{ realtimeData.close }}</div>
                </div>
              </el-col>
              <el-col :span="6">
                <div class="info-item">
                  <div class="label">开盘价</div>
                  <div class="value">¥{{ realtimeData.open }}</div>
                </div>
              </el-col>
              <el-col :span="6">
                <div class="info-item">
                  <div class="label">最高价</div>
                  <div class="value">¥{{ realtimeData.high }}</div>
                </div>
              </el-col>
              <el-col :span="6">
                <div class="info-item">
                  <div class="label">最低价</div>
                  <div class="value">¥{{ realtimeData.low }}</div>
                </div>
              </el-col>
            </el-row>
          </div>
        </el-card>
      </div>
    </el-card>
  </div>
</template>

<script>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { CandlestickChart } from 'echarts/charts'
import { GridComponent, TooltipComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import api from '../utils/api'

use([
  CanvasRenderer,
  CandlestickChart,
  GridComponent,
  TooltipComponent
])

export default {
  name: 'StockDetail',
  components: {
    VChart
  },
  setup() {
    const route = useRoute()
    const code = computed(() => route.params.code)
    
    const loading = ref(false)
    const stockDetail = ref(null)
    const klineData = ref([])
    const financialData = ref([])
    const realtimeData = ref(null)
    
    const showKLine = ref(false)
    const showFinancial = ref(false)
    const showRealtime = ref(false)
    const refreshing = ref(false)

    // K线图配置
    const klineOption = computed(() => {
      if (klineData.value.length === 0) return {}
      
      const data = klineData.value.map(item => [
        item.open,
        item.close,
        item.low,
        item.high
      ])
      
      const dates = klineData.value.map(item => item.trade_date)
      const volumes = klineData.value.map(item => item.vol || 0)
      
      return {
        animation: false,
        legend: {
          bottom: 10,
          left: 'center',
          data: ['K线', '成交量']
        },
        tooltip: {
          trigger: 'axis',
          axisPointer: {
            type: 'cross'
          },
          backgroundColor: 'rgba(245, 245, 245, 0.8)',
          borderWidth: 1,
          borderColor: '#ccc',
          textStyle: {
            color: '#000'
          },
          formatter: function (params) {
            let res = params[0].name + '<br/>'
            let klineData = params[0].data
            if (klineData && klineData.length >= 4) {
              res += '开盘: ' + klineData[0] + '<br/>'
              res += '收盘: ' + klineData[1] + '<br/>'
              res += '最低: ' + klineData[2] + '<br/>'
              res += '最高: ' + klineData[3] + '<br/>'
            }
            if (params[1] && params[1].data) {
              res += '成交量: ' + params[1].data + '<br/>'
            }
            return res
          }
        },
        axisPointer: {
          link: { xAxisIndex: 'all' },
          label: {
            backgroundColor: '#777'
          }
        },
        toolbox: {
          feature: {
            dataZoom: {
              yAxisIndex: false
            },
            brush: {
              type: ['lineX', 'clear']
            }
          }
        },
        brush: {
          xAxisIndex: 'all',
          brushLink: 'all',
          outOfBrush: {
            colorAlpha: 0.1
          }
        },
        visualMap: {
          show: false,
          seriesIndex: 1,
          dimension: 2,
          pieces: [
            {
              value: 1,
              color: '#00da3c'
            },
            {
              value: -1,
              color: '#ec0000'
            }
          ]
        },
        grid: [
          {
            left: '10%',
            right: '8%',
            height: '50%'
          },
          {
            left: '10%',
            right: '8%',
            top: '63%',
            height: '16%'
          }
        ],
        xAxis: [
          {
            type: 'category',
            data: dates,
            scale: true,
            boundaryGap: false,
            axisLine: { onZero: false },
            splitLine: { show: false },
            min: 'dataMin',
            max: 'dataMax',
            axisPointer: {
              z: 100
            }
          },
          {
            type: 'category',
            gridIndex: 1,
            data: dates,
            scale: true,
            boundaryGap: false,
            axisLine: { onZero: false },
            axisTick: { show: false },
            splitLine: { show: false },
            axisLabel: { show: false },
            min: 'dataMin',
            max: 'dataMax'
          }
        ],
        yAxis: [
          {
            scale: true,
            splitArea: {
              show: true
            }
          },
          {
            scale: true,
            gridIndex: 1,
            splitNumber: 2,
            axisLabel: { show: false },
            axisLine: { show: false },
            axisTick: { show: false },
            splitLine: { show: false }
          }
        ],
        dataZoom: [
          {
            type: 'inside',
            xAxisIndex: [0, 1],
            start: 50,
            end: 100
          },
          {
            show: true,
            xAxisIndex: [0, 1],
            type: 'slider',
            top: '85%',
            start: 50,
            end: 100
          }
        ],
        series: [
          {
            name: 'K线',
            type: 'candlestick',
            data: data,
            itemStyle: {
              color: '#FD1050',
              color0: '#0CF49B',
              borderColor: '#FD1050',
              borderColor0: '#0CF49B'
            },
            markPoint: {
              label: {
                formatter: function (param) {
                  return param != null ? Math.round(param.value) + '' : ''
                }
              },
              data: [
                {
                  name: 'Mark',
                  coord: ['2013/5/31', 2300],
                  value: 2300,
                  itemStyle: {
                    color: 'rgb(41,60,85)'
                  }
                },
                {
                  name: 'highest value',
                  type: 'max',
                  valueDim: 'highest'
                },
                {
                  name: 'lowest value',
                  type: 'min',
                  valueDim: 'lowest'
                },
                {
                  name: 'average value on close',
                  type: 'average',
                  valueDim: 'close'
                }
              ],
              tooltip: {
                formatter: function (param) {
                  return param.name + '<br>' + (param.data.coord || '')
                }
              }
            },
            markLine: {
              symbol: ['none', 'none'],
              data: [
                [
                  {
                    name: 'from lowest to highest',
                    type: 'min',
                    valueDim: 'lowest',
                    symbol: 'circle',
                    symbolSize: 10,
                    label: {
                      show: false
                    },
                    emphasis: {
                      label: {
                        show: false
                      }
                    }
                  },
                  {
                    type: 'max',
                    valueDim: 'highest',
                    symbol: 'circle',
                    symbolSize: 10,
                    label: {
                      show: false
                    },
                    emphasis: {
                      label: {
                        show: false
                      }
                    }
                  }
                ],
                {
                  name: 'min line on close',
                  type: 'min',
                  valueDim: 'close'
                },
                {
                  name: 'max line on close',
                  type: 'max',
                  valueDim: 'close'
                }
              ]
            }
          },
          {
            name: '成交量',
            type: 'bar',
            xAxisIndex: 1,
            yAxisIndex: 1,
            data: volumes.map((vol, idx) => {
              const kline = data[idx]
              return {
                value: vol,
                itemStyle: {
                  color: kline && kline[1] > kline[0] ? '#FD1050' : '#0CF49B'
                }
              }
            })
          }
        ]
      }
    })

    // 加载股票详情
    const loadStockDetail = async () => {
      loading.value = true
      try {
        const response = await api.getStockDetail(code.value)
        stockDetail.value = response.data
      } catch (error) {
        ElMessage.error('获取股票详情失败: ' + error.message)
      } finally {
        loading.value = false
      }
    }

    // 加载K线数据
    const loadKLineData = async () => {
      try {
        const response = await api.getKLineData(code.value, { days: 30 })
        klineData.value = response.data.kline || []
        showKLine.value = true
        ElMessage.success('K线数据加载成功')
      } catch (error) {
        ElMessage.error('获取K线数据失败: ' + error.message)
      }
    }

    // 加载财务数据
    const loadFinancialData = async () => {
      try {
        const response = await api.getFinancialData(code.value)
        financialData.value = response.data.financial || []
        showFinancial.value = true
        ElMessage.success('财务数据加载成功')
      } catch (error) {
        ElMessage.error('获取财务数据失败: ' + error.message)
      }
    }

    // 加载实时数据
    const loadRealtimeData = async () => {
      try {
        const response = await api.getRealtimeData(code.value)
        if (response.data.realtime && response.data.realtime.length > 0) {
          realtimeData.value = response.data.realtime[0]
          showRealtime.value = true
          ElMessage.success('实时数据加载成功')
        } else {
          ElMessage.warning('暂无实时数据')
        }
      } catch (error) {
        ElMessage.error('获取实时数据失败: ' + error.message)
      }
    }

    // 刷新K线数据
    const refreshKLineData = async () => {
      refreshing.value = true
      try {
        const response = await fetch(`http://localhost:8080/api/v1/stocks/${code.value}/kline/refresh`, {
          method: 'POST'
        })
        const result = await response.json()
        
        if (result.code === 0) {
          ElMessage.success(`数据刷新成功，共获取 ${result.data.count} 条记录`)
          // 刷新成功后重新加载K线数据
          await loadKLineData()
        } else {
          ElMessage.error('刷新失败: ' + result.message)
        }
      } catch (error) {
        ElMessage.error('刷新K线数据失败: ' + error.message)
      } finally {
        refreshing.value = false
      }
    }

    // 检查数据状态
    const checkDataStatus = async () => {
      try {
        // 获取数据范围
        const rangeResponse = await fetch(`http://localhost:8080/api/v1/stocks/${code.value}/kline/range`)
        const rangeResult = await rangeResponse.json()
        
        // 获取数据新鲜度
        const freshnessResponse = await fetch(`http://localhost:8080/api/v1/stocks/${code.value}/kline/freshness`)
        const freshnessResult = await freshnessResponse.json()
        
        if (rangeResult.code === 0 && freshnessResult.code === 0) {
          const rangeData = rangeResult.data
          const freshnessData = freshnessResult.data
          
          let message = `数据状态检查结果:\n`
          if (rangeData.count > 0) {
            message += `• 数据库中共有 ${rangeData.count} 条记录\n`
            message += `• 数据时间范围: ${rangeData.start_date} 至 ${rangeData.end_date}\n`
          } else {
            message += `• 数据库中暂无数据\n`
          }
          
          if (freshnessData.need_update) {
            message += `• 状态: 需要更新\n`
            message += `• 原因: ${freshnessData.reason || '数据过期'}`
          } else {
            message += `• 状态: 数据最新`
          }
          
          ElMessage({
            message: message,
            type: freshnessData.need_update ? 'warning' : 'success',
            duration: 5000,
            showClose: true
          })
        } else {
          ElMessage.error('检查数据状态失败')
        }
      } catch (error) {
        ElMessage.error('检查数据状态失败: ' + error.message)
      }
    }

    onMounted(() => {
      loadStockDetail()
    })

    return {
      code,
      loading,
      stockDetail,
      klineData,
      financialData,
      realtimeData,
      showKLine,
      showFinancial,
      showRealtime,
      refreshing,
      klineOption,
      loadKLineData,
      loadFinancialData,
      loadRealtimeData,
      refreshKLineData,
      checkDataStatus
    }
  }
}
</script>

<style lang="scss" scoped>
.stock-detail {
  max-width: 1200px;
  margin: 0 auto;

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .detail-content {
    .action-buttons {
      display: flex;
      flex-direction: column;
      gap: 10px;
      align-items: flex-start;

      .el-button {
        width: 150px;
      }
    }

    .chart-card,
    .financial-card,
    .realtime-card {
      margin-top: 20px;
    }

    .chart-container {
      width: 100%;
    }

    .no-data {
      text-align: center;
      padding: 50px;
      color: #909399;
    }

    .realtime-info {
      .info-item {
        text-align: center;
        padding: 20px;
        border: 1px solid #ebeef5;
        border-radius: 8px;

        .label {
          font-size: 14px;
          color: #909399;
          margin-bottom: 10px;
        }

        .value {
          font-size: 24px;
          font-weight: bold;
          color: #303133;

          &.price {
            color: #f56c6c;
          }
        }
      }
    }
  }
}
</style>