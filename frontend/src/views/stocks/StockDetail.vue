<template>
  <div class="stock-detail">
    <!-- 股票基本信息 -->
    <el-card class="info-card">
      <div class="stock-header">
        <div class="stock-info">
          <h2 class="stock-name">{{ stockInfo.name }} ({{ stockInfo.ts_code }})</h2>
          <div class="stock-tags">
            <el-tag :type="stockInfo.market === 'SH' ? 'danger' : 'success'" size="small">
              {{ stockInfo.market }}
            </el-tag>
            <el-tag v-if="stockInfo.industry" type="info" size="small">
              {{ stockInfo.industry }}
            </el-tag>
            <el-tag v-if="stockInfo.area" type="warning" size="small">
              {{ stockInfo.area }}
            </el-tag>
          </div>
        </div>
        <div class="stock-actions">
          <el-button type="primary" @click="addToWatchlist">
            <el-icon><Star /></el-icon>
            加入自选
          </el-button>
          <el-button @click="refreshData" :loading="loading">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </div>

      <!-- 实时价格信息 -->
      <div class="price-info" v-if="realtimeData">
        <div class="price-main">
          <span class="current-price" :class="getPriceClass(realtimeData.change_percent)">
            ¥{{ realtimeData.price }}
          </span>
          <div class="price-change">
            <span :class="getPriceClass(realtimeData.change_percent)">
              {{ realtimeData.change_amount > 0 ? '+' : '' }}{{ realtimeData.change_amount }}
            </span>
            <span :class="getPriceClass(realtimeData.change_percent)">
              ({{ realtimeData.change_percent > 0 ? '+' : '' }}{{ realtimeData.change_percent }}%)
            </span>
          </div>
        </div>
        <div class="price-details">
          <div class="price-item">
            <span class="label">今开:</span>
            <span class="value">{{ realtimeData.open }}</span>
          </div>
          <div class="price-item">
            <span class="label">昨收:</span>
            <span class="value">{{ realtimeData.pre_close }}</span>
          </div>
          <div class="price-item">
            <span class="label">最高:</span>
            <span class="value">{{ realtimeData.high }}</span>
          </div>
          <div class="price-item">
            <span class="label">最低:</span>
            <span class="value">{{ realtimeData.low }}</span>
          </div>
          <div class="price-item">
            <span class="label">成交量:</span>
            <span class="value">{{ formatVolume(realtimeData.volume) }}</span>
          </div>
          <div class="price-item">
            <span class="label">成交额:</span>
            <span class="value">{{ formatAmount(realtimeData.amount) }}</span>
          </div>
        </div>
      </div>
    </el-card>

    <!-- K线图表和数据 -->
    <el-row :gutter="20">
      <!-- K线图 -->
      <el-col :xs="24" :lg="16">
        <el-card class="chart-card">
          <template #header>
            <div class="chart-header">
              <span>K线图</span>
              <div class="chart-controls">
                <el-radio-group v-model="selectedPeriod" size="small" @change="loadKLineData">
                  <el-radio-button label="5">5日</el-radio-button>
                  <el-radio-button label="30">30日</el-radio-button>
                  <el-radio-button label="90">90日</el-radio-button>
                  <el-radio-button label="365">1年</el-radio-button>
                </el-radio-group>
                <el-button size="small" @click="refreshKLineData">
                  <el-icon><Refresh /></el-icon>
                  刷新数据
                </el-button>
              </div>
            </div>
          </template>
          <div class="chart-container" v-loading="chartLoading">
            <div v-if="klineData.length === 0" class="no-data">
              暂无K线数据
            </div>
            <div v-else class="kline-chart" ref="chartRef" style="height: 400px;"></div>
          </div>
        </el-card>
      </el-col>

      <!-- 财务数据 -->
      <el-col :xs="24" :lg="8">
        <el-card class="financial-card">
          <template #header>
            <span>财务数据</span>
          </template>
          <div class="financial-data" v-loading="financialLoading">
            <div v-if="financialData" class="financial-items">
              <div class="financial-item">
                <span class="label">总市值:</span>
                <span class="value">{{ formatAmount(financialData.total_mv) }}</span>
              </div>
              <div class="financial-item">
                <span class="label">流通市值:</span>
                <span class="value">{{ formatAmount(financialData.circ_mv) }}</span>
              </div>
              <div class="financial-item">
                <span class="label">市盈率(TTM):</span>
                <span class="value">{{ financialData.pe_ttm || '--' }}</span>
              </div>
              <div class="financial-item">
                <span class="label">市净率:</span>
                <span class="value">{{ financialData.pb || '--' }}</span>
              </div>
              <div class="financial-item">
                <span class="label">净资产收益率:</span>
                <span class="value">{{ financialData.roe || '--' }}%</span>
              </div>
              <div class="financial-item">
                <span class="label">营业收入:</span>
                <span class="value">{{ formatAmount(financialData.revenue) }}</span>
              </div>
            </div>
            <div v-else class="no-data">
              暂无财务数据
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 业绩报表 -->
    <el-card class="performance-card">
      <template #header>
        <div class="card-header">
          <span>业绩报表</span>
          <el-button size="small" @click="loadPerformanceData">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </template>
      <el-table :data="performanceData" v-loading="performanceLoading" stripe>
        <el-table-column prop="report_date" label="报告期" width="120">
          <template #default="scope">
            {{ formatDate(scope.row.report_date) }}
          </template>
        </el-table-column>
        <el-table-column prop="revenue" label="营业收入(万元)" width="140">
          <template #default="scope">
            {{ formatNumber(scope.row.revenue) }}
          </template>
        </el-table-column>
        <el-table-column prop="net_profit" label="净利润(万元)" width="140">
          <template #default="scope">
            <span :class="scope.row.net_profit >= 0 ? 'profit-positive' : 'profit-negative'">
              {{ formatNumber(scope.row.net_profit) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="eps" label="每股收益" width="100">
          <template #default="scope">
            {{ scope.row.eps || '--' }}
          </template>
        </el-table-column>
        <el-table-column prop="roe" label="净资产收益率" width="120">
          <template #default="scope">
            {{ scope.row.roe ? scope.row.roe + '%' : '--' }}
          </template>
        </el-table-column>
        <el-table-column prop="gross_margin" label="毛利率" width="100">
          <template #default="scope">
            {{ scope.row.gross_margin ? scope.row.gross_margin + '%' : '--' }}
          </template>
        </el-table-column>
        <el-table-column prop="debt_ratio" label="资产负债率" width="120">
          <template #default="scope">
            {{ scope.row.debt_ratio ? scope.row.debt_ratio + '%' : '--' }}
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script>
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import * as echarts from 'echarts'
import api from '../../utils/api'

export default {
  name: 'StockDetail',
  props: {
    code: {
      type: String,
      required: true
    }
  },
  setup(props) {
    const route = useRoute()
    
    // 响应式数据
    const loading = ref(false)
    const chartLoading = ref(false)
    const financialLoading = ref(false)
    const performanceLoading = ref(false)
    
    const stockInfo = ref({})
    const realtimeData = ref(null)
    const klineData = ref([])
    const financialData = ref(null)
    const performanceData = ref([])
    
    const selectedPeriod = ref('30')
    const chartRef = ref(null)
    let chartInstance = null
    
    // 方法
    const loadStockInfo = async () => {
      loading.value = true
      try {
        const response = await api.get(`/api/v1/stocks/${props.code}`)
        if (response.data.code === 0) {
          stockInfo.value = response.data.data
        } else {
          ElMessage.error(response.data.message || '加载股票信息失败')
        }
      } catch (error) {
        console.error('加载股票信息失败:', error)
        ElMessage.error('加载股票信息失败')
      } finally {
        loading.value = false
      }
    }
    
    const loadRealtimeData = async () => {
      try {
        const response = await api.get('/api/v1/realtime', {
          params: { codes: props.code }
        })
        if (response.data.code === 0) {
          const realtime = response.data.data.realtime || []
          if (realtime.length > 0) {
            realtimeData.value = realtime[0]
          }
        }
      } catch (error) {
        console.error('加载实时数据失败:', error)
      }
    }
    
    const loadKLineData = async () => {
      chartLoading.value = true
      try {
        const response = await api.get(`/api/v1/stocks/${props.code}/kline`, {
          params: { days: selectedPeriod.value }
        })
        if (response.data.code === 0) {
          klineData.value = response.data.data.kline || []
          await nextTick()
          renderChart()
        } else {
          ElMessage.error(response.data.message || '加载K线数据失败')
        }
      } catch (error) {
        console.error('加载K线数据失败:', error)
        ElMessage.error('加载K线数据失败')
      } finally {
        chartLoading.value = false
      }
    }
    
    const loadFinancialData = async () => {
      financialLoading.value = true
      try {
        const response = await api.get(`/api/v1/stocks/${props.code}/financial`)
        if (response.data.code === 0) {
          financialData.value = response.data.data
        }
      } catch (error) {
        console.error('加载财务数据失败:', error)
      } finally {
        financialLoading.value = false
      }
    }
    
    const loadPerformanceData = async () => {
      performanceLoading.value = true
      try {
        const response = await api.get(`/api/v1/stocks/${props.code}/performance`)
        if (response.data.code === 0) {
          performanceData.value = response.data.data.reports || []
        }
      } catch (error) {
        console.error('加载业绩数据失败:', error)
      } finally {
        performanceLoading.value = false
      }
    }
    
    const renderChart = () => {
      if (!chartRef.value || klineData.value.length === 0) return
      
      if (!chartInstance) {
        chartInstance = echarts.init(chartRef.value)
      }
      
      const dates = klineData.value.map(item => item.trade_date)
      const values = klineData.value.map(item => [
        item.open, item.close, item.low, item.high
      ])
      
      const option = {
        title: {
          text: `${stockInfo.value.name} K线图`,
          left: 'center'
        },
        tooltip: {
          trigger: 'axis',
          axisPointer: {
            type: 'cross'
          }
        },
        xAxis: {
          type: 'category',
          data: dates,
          scale: true,
          boundaryGap: false,
          axisLine: { onZero: false },
          splitLine: { show: false },
          splitNumber: 20,
          min: 'dataMin',
          max: 'dataMax'
        },
        yAxis: {
          scale: true,
          splitArea: {
            show: true
          }
        },
        series: [
          {
            name: 'K线',
            type: 'candlestick',
            data: values,
            itemStyle: {
              color: '#ef232a',
              color0: '#14b143',
              borderColor: '#ef232a',
              borderColor0: '#14b143'
            }
          }
        ]
      }
      
      chartInstance.setOption(option)
    }
    
    const refreshKLineData = async () => {
      try {
        const response = await api.post(`/api/v1/stocks/${props.code}/kline/refresh`, {
          days: selectedPeriod.value
        })
        if (response.data.code === 0) {
          ElMessage.success('K线数据刷新成功')
          await loadKLineData()
        } else {
          ElMessage.error(response.data.message || 'K线数据刷新失败')
        }
      } catch (error) {
        console.error('刷新K线数据失败:', error)
        ElMessage.error('刷新K线数据失败')
      }
    }
    
    const addToWatchlist = async () => {
      try {
        await api.post(`/api/v1/stocks/${props.code}/watchlist`)
        ElMessage.success('已添加到自选股')
      } catch (error) {
        console.error('添加自选股失败:', error)
        ElMessage.error('添加自选股失败')
      }
    }
    
    const refreshData = () => {
      loadStockInfo()
      loadRealtimeData()
      loadKLineData()
      loadFinancialData()
      loadPerformanceData()
    }
    
    // 工具方法
    const getPriceClass = (change) => {
      const changeNum = parseFloat(change)
      if (changeNum > 0) return 'price-up'
      if (changeNum < 0) return 'price-down'
      return 'price-flat'
    }
    
    const formatVolume = (volume) => {
      if (!volume) return '--'
      const vol = parseFloat(volume)
      if (vol >= 100000000) return (vol / 100000000).toFixed(2) + '亿'
      if (vol >= 10000) return (vol / 10000).toFixed(2) + '万'
      return vol.toString()
    }
    
    const formatAmount = (amount) => {
      if (!amount) return '--'
      const amt = parseFloat(amount)
      if (amt >= 100000000) return (amt / 100000000).toFixed(2) + '亿'
      if (amt >= 10000) return (amt / 10000).toFixed(2) + '万'
      return amt.toString()
    }
    
    const formatNumber = (num) => {
      if (!num) return '--'
      return parseFloat(num).toLocaleString()
    }
    
    const formatDate = (dateStr) => {
      if (!dateStr) return '--'
      return new Date(dateStr).toLocaleDateString()
    }
    
    // 生命周期
    onMounted(() => {
      loadStockInfo()
      loadRealtimeData()
      loadKLineData()
      loadFinancialData()
      loadPerformanceData()
      
      // 定时刷新实时数据
      const timer = setInterval(loadRealtimeData, 30000)
      onUnmounted(() => {
        clearInterval(timer)
        if (chartInstance) {
          chartInstance.dispose()
        }
      })
    })
    
    return {
      loading,
      chartLoading,
      financialLoading,
      performanceLoading,
      stockInfo,
      realtimeData,
      klineData,
      financialData,
      performanceData,
      selectedPeriod,
      chartRef,
      loadKLineData,
      refreshKLineData,
      addToWatchlist,
      refreshData,
      getPriceClass,
      formatVolume,
      formatAmount,
      formatNumber,
      formatDate
    }
  }
}
</script>

<style lang="scss" scoped>
.stock-detail {
  .info-card {
    margin-bottom: 20px;
    
    .stock-header {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      margin-bottom: 20px;
      
      .stock-info {
        .stock-name {
          margin: 0 0 8px 0;
          color: #303133;
        }
        
        .stock-tags {
          display: flex;
          gap: 8px;
        }
      }
      
      .stock-actions {
        display: flex;
        gap: 8px;
      }
    }
    
    .price-info {
      .price-main {
        display: flex;
        align-items: baseline;
        gap: 16px;
        margin-bottom: 16px;
        
        .current-price {
          font-size: 32px;
          font-weight: bold;
          
          &.price-up {
            color: #f56c6c;
          }
          
          &.price-down {
            color: #67c23a;
          }
          
          &.price-flat {
            color: #303133;
          }
        }
        
        .price-change {
          display: flex;
          flex-direction: column;
          font-size: 14px;
          
          span {
            &.price-up {
              color: #f56c6c;
            }
            
            &.price-down {
              color: #67c23a;
            }
            
            &.price-flat {
              color: #303133;
            }
          }
        }
      }
      
      .price-details {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
        gap: 16px;
        
        .price-item {
          display: flex;
          justify-content: space-between;
          
          .label {
            color: #909399;
            font-size: 14px;
          }
          
          .value {
            font-weight: 500;
            color: #303133;
          }
        }
      }
    }
  }
  
  .chart-card {
    margin-bottom: 20px;
    
    .chart-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      
      .chart-controls {
        display: flex;
        gap: 12px;
        align-items: center;
      }
    }
    
    .chart-container {
      .no-data {
        display: flex;
        justify-content: center;
        align-items: center;
        height: 400px;
        color: #909399;
        font-size: 16px;
      }
    }
  }
  
  .financial-card {
    margin-bottom: 20px;
    
    .financial-data {
      .financial-items {
        .financial-item {
          display: flex;
          justify-content: space-between;
          padding: 8px 0;
          border-bottom: 1px solid #f0f0f0;
          
          &:last-child {
            border-bottom: none;
          }
          
          .label {
            color: #909399;
            font-size: 14px;
          }
          
          .value {
            font-weight: 500;
            color: #303133;
          }
        }
      }
      
      .no-data {
        text-align: center;
        color: #909399;
        padding: 40px 0;
      }
    }
  }
  
  .performance-card {
    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
    }
    
    .profit-positive {
      color: #f56c6c;
    }
    
    .profit-negative {
      color: #67c23a;
    }
  }
}

@media (max-width: 768px) {
  .stock-detail {
    .stock-header {
      flex-direction: column;
      gap: 16px;
    }
    
    .price-main {
      flex-direction: column;
      align-items: flex-start !important;
      gap: 8px !important;
    }
    
    .chart-controls {
      flex-direction: column;
      gap: 8px !important;
    }
  }
}
</style>