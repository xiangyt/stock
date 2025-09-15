<template>
  <div class="analysis">
    <el-card>
      <template #header>
        <span>数据分析</span>
      </template>
      
      <div class="analysis-content">
        <el-row :gutter="20">
          <el-col :span="8">
            <el-card class="stat-card">
              <div class="stat-item">
                <div class="stat-icon">
                  <el-icon><DataLine /></el-icon>
                </div>
                <div class="stat-info">
                  <div class="stat-value">{{ totalStocks }}</div>
                  <div class="stat-label">总股票数</div>
                </div>
              </div>
            </el-card>
          </el-col>
          <el-col :span="8">
            <el-card class="stat-card">
              <div class="stat-item">
                <div class="stat-icon">
                  <el-icon><TrendCharts /></el-icon>
                </div>
                <div class="stat-info">
                  <div class="stat-value">{{ activeStocks }}</div>
                  <div class="stat-label">正常交易</div>
                </div>
              </div>
            </el-card>
          </el-col>
          <el-col :span="8">
            <el-card class="stat-card">
              <div class="stat-item">
                <div class="stat-icon">
                  <el-icon><PieChart /></el-icon>
                </div>
                <div class="stat-info">
                  <div class="stat-value">{{ industries.length }}</div>
                  <div class="stat-label">行业数量</div>
                </div>
              </div>
            </el-card>
          </el-col>
        </el-row>

        <!-- 行业分布图 -->
        <el-card class="chart-card">
          <template #header>
            <span>行业分布</span>
          </template>
          <div v-if="industries.length > 0" class="chart-container">
            <v-chart :option="industryOption" style="height: 400px;" />
          </div>
          <div v-else class="no-data">暂无数据</div>
        </el-card>

        <!-- 市场分布图 -->
        <el-card class="chart-card">
          <template #header>
            <span>市场分布</span>
          </template>
          <div v-if="markets.length > 0" class="chart-container">
            <v-chart :option="marketOption" style="height: 300px;" />
          </div>
          <div v-else class="no-data">暂无数据</div>
        </el-card>
      </div>
    </el-card>
  </div>
</template>

<script>
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { PieChart, BarChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import api from '../utils/api'

use([
  CanvasRenderer,
  PieChart,
  BarChart,
  GridComponent,
  TooltipComponent,
  LegendComponent
])

export default {
  name: 'Analysis',
  components: {
    VChart
  },
  setup() {
    const stocks = ref([])
    const loading = ref(false)

    // 统计数据
    const totalStocks = computed(() => stocks.value.length)
    const activeStocks = computed(() => stocks.value.filter(s => s.list_status === 'L').length)
    
    // 行业分布
    const industries = computed(() => {
      const industryMap = {}
      stocks.value.forEach(stock => {
        const industry = stock.industry || '其他'
        industryMap[industry] = (industryMap[industry] || 0) + 1
      })
      return Object.entries(industryMap)
        .map(([name, value]) => ({ name, value }))
        .sort((a, b) => b.value - a.value)
        .slice(0, 10) // 只显示前10个行业
    })

    // 市场分布
    const markets = computed(() => {
      const marketMap = {}
      stocks.value.forEach(stock => {
        const market = stock.market || '其他'
        marketMap[market] = (marketMap[market] || 0) + 1
      })
      return Object.entries(marketMap).map(([name, value]) => ({ name, value }))
    })

    // 行业分布图配置
    const industryOption = computed(() => ({
      tooltip: {
        trigger: 'item',
        formatter: '{a} <br/>{b}: {c} ({d}%)'
      },
      legend: {
        orient: 'vertical',
        left: 'left'
      },
      series: [
        {
          name: '行业分布',
          type: 'pie',
          radius: '50%',
          data: industries.value,
          emphasis: {
            itemStyle: {
              shadowBlur: 10,
              shadowOffsetX: 0,
              shadowColor: 'rgba(0, 0, 0, 0.5)'
            }
          }
        }
      ]
    }))

    // 市场分布图配置
    const marketOption = computed(() => ({
      tooltip: {
        trigger: 'axis',
        axisPointer: {
          type: 'shadow'
        }
      },
      grid: {
        left: '3%',
        right: '4%',
        bottom: '3%',
        containLabel: true
      },
      xAxis: {
        type: 'category',
        data: markets.value.map(m => m.name)
      },
      yAxis: {
        type: 'value'
      },
      series: [
        {
          name: '股票数量',
          type: 'bar',
          data: markets.value.map(m => m.value),
          itemStyle: {
            color: '#409eff'
          }
        }
      ]
    }))

    // 获取股票数据
    const fetchData = async () => {
      loading.value = true
      try {
        const response = await api.getStockList({ page: 1, size: 1000 })
        stocks.value = response.data.stocks || []
      } catch (error) {
        ElMessage.error('获取数据失败: ' + error.message)
      } finally {
        loading.value = false
      }
    }

    onMounted(() => {
      fetchData()
    })

    return {
      loading,
      totalStocks,
      activeStocks,
      industries,
      markets,
      industryOption,
      marketOption
    }
  }
}
</script>

<style lang="scss" scoped>
.analysis {
  max-width: 1200px;
  margin: 0 auto;

  .analysis-content {
    .stat-card {
      .stat-item {
        display: flex;
        align-items: center;
        padding: 20px;

        .stat-icon {
          font-size: 48px;
          color: #409eff;
          margin-right: 20px;
        }

        .stat-info {
          .stat-value {
            font-size: 32px;
            font-weight: bold;
            color: #303133;
            margin-bottom: 5px;
          }

          .stat-label {
            font-size: 14px;
            color: #909399;
          }
        }
      }
    }

    .chart-card {
      margin-top: 20px;

      .chart-container {
        width: 100%;
      }

      .no-data {
        text-align: center;
        padding: 50px;
        color: #909399;
      }
    }
  }
}
</style>