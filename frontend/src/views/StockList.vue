<template>
  <div class="stock-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>股票列表</span>
          <div class="header-actions">
            <el-input
              v-model="searchKeyword"
              placeholder="搜索股票代码或名称"
              style="width: 200px; margin-right: 10px;"
              clearable
              @input="handleSearch"
            >
              <template #prefix>
                <el-icon><Search /></el-icon>
              </template>
            </el-input>
            <el-button type="primary" @click="refreshData" :loading="loading">
              <el-icon><Refresh /></el-icon>
              刷新数据
            </el-button>
          </div>
        </div>
      </template>

      <!-- 股票表格 -->
      <el-table
        :data="filteredStocks"
        v-loading="loading"
        stripe
        style="width: 100%"
        @row-click="handleRowClick"
        class="stock-table"
      >
        <el-table-column prop="ts_code" label="股票代码" width="120" />
        <el-table-column prop="name" label="股票名称" width="150" />
        <el-table-column prop="industry" label="行业" width="120" />
        <el-table-column prop="market" label="市场" width="80">
          <template #default="scope">
            <el-tag :type="scope.row.market === '上海' ? 'primary' : 'success'">
              {{ scope.row.market }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="list_status" label="状态" width="80">
          <template #default="scope">
            <el-tag :type="scope.row.list_status === 'L' ? 'success' : 'warning'">
              {{ scope.row.list_status === 'L' ? '正常' : '其他' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="list_date" label="上市日期" width="120" />
        <el-table-column label="操作" width="200">
          <template #default="scope">
            <el-button
              type="primary"
              size="small"
              @click.stop="viewDetail(scope.row.ts_code)"
            >
              查看详情
            </el-button>
            <el-button
              type="success"
              size="small"
              @click.stop="viewKLine(scope.row.ts_code)"
            >
              K线图
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

    <!-- K线图对话框 -->
    <el-dialog
      v-model="klineDialogVisible"
      :title="`${selectedStock?.name} (${selectedStock?.ts_code}) K线图`"
      width="80%"
      top="5vh"
    >
      <div v-if="klineData.length > 0" class="kline-container">
        <v-chart :option="klineOption" style="height: 400px;" />
      </div>
      <div v-else v-loading="klineLoading" style="height: 400px; display: flex; align-items: center; justify-content: center;">
        <span>暂无K线数据</span>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { CandlestickChart, LineChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import api from '../utils/api'

use([
  CanvasRenderer,
  CandlestickChart,
  LineChart,
  GridComponent,
  TooltipComponent,
  LegendComponent
])

export default {
  name: 'StockList',
  components: {
    VChart
  },
  setup() {
    const router = useRouter()
    
    const stocks = ref([])
    const loading = ref(false)
    const searchKeyword = ref('')
    const currentPage = ref(1)
    const pageSize = ref(20)
    const total = ref(0)
    
    const klineDialogVisible = ref(false)
    const klineLoading = ref(false)
    const klineData = ref([])
    const selectedStock = ref(null)

    // 过滤后的股票列表
    const filteredStocks = computed(() => {
      if (!searchKeyword.value) {
        return stocks.value
      }
      return stocks.value.filter(stock => 
        stock.ts_code.toLowerCase().includes(searchKeyword.value.toLowerCase()) ||
        stock.name.includes(searchKeyword.value)
      )
    })

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
      
      return {
        tooltip: {
          trigger: 'axis',
          axisPointer: {
            type: 'cross'
          }
        },
        legend: {
          data: ['K线']
        },
        grid: {
          left: '10%',
          right: '10%',
          bottom: '15%'
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
            data: data,
            itemStyle: {
              color: '#ec0000',
              color0: '#00da3c',
              borderColor: '#8A0000',
              borderColor0: '#008F28'
            }
          }
        ]
      }
    })

    // 获取股票列表
    const fetchStocks = async () => {
      loading.value = true
      try {
        const response = await api.getStockList({
          page: currentPage.value,
          size: pageSize.value
        })
        stocks.value = response.data.stocks || []
        total.value = response.data.total || 0
      } catch (error) {
        ElMessage.error('获取股票列表失败: ' + error.message)
      } finally {
        loading.value = false
      }
    }

    // 刷新数据
    const refreshData = () => {
      fetchStocks()
    }

    // 搜索处理
    const handleSearch = () => {
      // 搜索是通过计算属性实现的，这里可以添加其他逻辑
    }

    // 分页处理
    const handleSizeChange = (val) => {
      pageSize.value = val
      fetchStocks()
    }

    const handleCurrentChange = (val) => {
      currentPage.value = val
      fetchStocks()
    }

    // 行点击处理
    const handleRowClick = (row) => {
      viewDetail(row.ts_code)
    }

    // 查看详情
    const viewDetail = (code) => {
      router.push(`/stock/${code}`)
    }

    // 查看K线图
    const viewKLine = async (code) => {
      selectedStock.value = stocks.value.find(s => s.ts_code === code)
      klineDialogVisible.value = true
      klineLoading.value = true
      
      try {
        const response = await api.getKLineData(code, { days: 30 })
        klineData.value = response.data.kline || []
      } catch (error) {
        ElMessage.error('获取K线数据失败: ' + error.message)
        klineData.value = []
      } finally {
        klineLoading.value = false
      }
    }

    onMounted(() => {
      fetchStocks()
    })

    return {
      stocks,
      loading,
      searchKeyword,
      currentPage,
      pageSize,
      total,
      filteredStocks,
      klineDialogVisible,
      klineLoading,
      klineData,
      selectedStock,
      klineOption,
      refreshData,
      handleSearch,
      handleSizeChange,
      handleCurrentChange,
      handleRowClick,
      viewDetail,
      viewKLine
    }
  }
}
</script>

<style lang="scss" scoped>
.stock-list {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .header-actions {
    display: flex;
    align-items: center;
  }

  .stock-table {
    margin-bottom: 20px;

    :deep(.el-table__row) {
      cursor: pointer;
      
      &:hover {
        background-color: #f5f7fa;
      }
    }
  }

  .pagination {
    display: flex;
    justify-content: center;
    margin-top: 20px;
  }

  .kline-container {
    width: 100%;
  }
}
</style>