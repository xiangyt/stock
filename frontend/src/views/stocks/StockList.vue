<template>
  <div class="stock-list">
    <!-- 搜索和筛选区域 -->
    <el-card class="search-card">
      <el-row :gutter="20">
        <el-col :xs="24" :sm="12" :md="8">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索股票代码或名称"
            clearable
            @input="handleSearch"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </el-col>
        <el-col :xs="24" :sm="12" :md="6">
          <el-select v-model="selectedMarket" placeholder="选择市场" clearable @change="handleFilter">
            <el-option label="全部" value="" />
            <el-option label="上海证券交易所" value="SH" />
            <el-option label="深圳证券交易所" value="SZ" />
          </el-select>
        </el-col>
        <el-col :xs="24" :sm="12" :md="6">
          <el-select v-model="selectedIndustry" placeholder="选择行业" clearable @change="handleFilter">
            <el-option label="全部" value="" />
            <el-option label="银行" value="银行" />
            <el-option label="房地产" value="房地产" />
            <el-option label="医药生物" value="医药生物" />
            <el-option label="电子" value="电子" />
          </el-select>
        </el-col>
        <el-col :xs="24" :sm="12" :md="4">
          <el-button type="primary" @click="refreshData" :loading="loading">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </el-col>
      </el-row>
    </el-card>

    <!-- 股票列表 -->
    <el-card class="table-card">
      <template #header>
        <div class="card-header">
          <span>股票列表 (共 {{ total }} 只)</span>
          <div class="header-actions">
            <el-button size="small" @click="exportData">
              <el-icon><Download /></el-icon>
              导出
            </el-button>
            <el-button type="primary" size="small" @click="syncAllStocks">
              <el-icon><Refresh /></el-icon>
              同步数据
            </el-button>
          </div>
        </div>
      </template>

      <el-table
        :data="stockList"
        v-loading="loading"
        stripe
        style="width: 100%"
        @row-click="handleRowClick"
        class="stock-table"
      >
        <el-table-column prop="ts_code" label="股票代码" width="120" fixed="left">
          <template #default="scope">
            <el-link type="primary" @click.stop="goToDetail(scope.row.ts_code)">
              {{ scope.row.ts_code }}
            </el-link>
          </template>
        </el-table-column>
        
        <el-table-column prop="name" label="股票名称" width="150" show-overflow-tooltip />
        
        <el-table-column prop="market" label="市场" width="80">
          <template #default="scope">
            <el-tag :type="scope.row.market === 'SH' ? 'danger' : 'success'" size="small">
              {{ scope.row.market }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column prop="industry" label="行业" width="120" show-overflow-tooltip />
        
        <el-table-column prop="area" label="地区" width="100" show-overflow-tooltip />
        
        <el-table-column label="实时价格" width="100">
          <template #default="scope">
            <span v-if="realtimeData[scope.row.ts_code]" class="price">
              {{ realtimeData[scope.row.ts_code].price }}
            </span>
            <span v-else class="no-data">--</span>
          </template>
        </el-table-column>
        
        <el-table-column label="涨跌幅" width="100">
          <template #default="scope">
            <span 
              v-if="realtimeData[scope.row.ts_code]" 
              :class="getPriceChangeClass(realtimeData[scope.row.ts_code].change)"
              class="change"
            >
              {{ realtimeData[scope.row.ts_code].change }}%
            </span>
            <span v-else class="no-data">--</span>
          </template>
        </el-table-column>
        
        <el-table-column prop="list_date" label="上市日期" width="120">
          <template #default="scope">
            {{ formatDate(scope.row.list_date) }}
          </template>
        </el-table-column>
        
        <el-table-column label="状态" width="80">
          <template #default="scope">
            <el-tag :type="scope.row.is_active ? 'success' : 'info'" size="small">
              {{ scope.row.is_active ? '正常' : '停牌' }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="scope">
            <el-button size="small" type="text" @click.stop="addToWatchlist(scope.row)">
              <el-icon><Star /></el-icon>
              关注
            </el-button>
            <el-button size="small" type="text" @click.stop="goToDetail(scope.row.ts_code)">
              <el-icon><View /></el-icon>
              详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="[20, 50, 100, 200]"
        :total="total"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </el-card>
  </div>
</template>

<script>
import { ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '../../utils/api'

export default {
  name: 'StockList',
  setup() {
    const router = useRouter()
    
    // 响应式数据
    const loading = ref(false)
    const stockList = ref([])
    const realtimeData = ref({})
    const total = ref(0)
    const currentPage = ref(1)
    const pageSize = ref(20)
    
    // 搜索和筛选
    const searchKeyword = ref('')
    const selectedMarket = ref('')
    const selectedIndustry = ref('')
    
    // 方法
    const loadStockList = async () => {
      loading.value = true
      try {
        const response = await api.get('/api/v1/stocks', {
          params: {
            page: currentPage.value,
            size: pageSize.value,
            keyword: searchKeyword.value,
            market: selectedMarket.value,
            industry: selectedIndustry.value
          }
        })
        
        if (response.data.code === 0) {
          stockList.value = response.data.data.stocks || []
          total.value = response.data.data.total || 0
          
          // 加载实时数据
          await loadRealtimeData()
        } else {
          ElMessage.error(response.data.message || '加载股票列表失败')
        }
      } catch (error) {
        console.error('加载股票列表失败:', error)
        ElMessage.error('加载股票列表失败')
      } finally {
        loading.value = false
      }
    }
    
    const loadRealtimeData = async () => {
      if (stockList.value.length === 0) return
      
      try {
        const codes = stockList.value.slice(0, 10).map(stock => stock.ts_code).join(',')
        const response = await api.get('/api/v1/realtime', {
          params: { codes }
        })
        
        if (response.data.code === 0) {
          const realtime = response.data.data.realtime || []
          const realtimeMap = {}
          realtime.forEach(item => {
            realtimeMap[item.code] = {
              price: item.price,
              change: item.change_percent
            }
          })
          realtimeData.value = realtimeMap
        }
      } catch (error) {
        console.error('加载实时数据失败:', error)
      }
    }
    
    const handleSearch = () => {
      currentPage.value = 1
      loadStockList()
    }
    
    const handleFilter = () => {
      currentPage.value = 1
      loadStockList()
    }
    
    const handleSizeChange = (size) => {
      pageSize.value = size
      currentPage.value = 1
      loadStockList()
    }
    
    const handleCurrentChange = (page) => {
      currentPage.value = page
      loadStockList()
    }
    
    const handleRowClick = (row) => {
      goToDetail(row.ts_code)
    }
    
    const goToDetail = (code) => {
      router.push(`/stock/${code}`)
    }
    
    const addToWatchlist = async (stock) => {
      try {
        await api.post(`/api/v1/stocks/${stock.ts_code}/watchlist`)
        ElMessage.success(`已将 ${stock.name} 添加到自选股`)
      } catch (error) {
        console.error('添加自选股失败:', error)
        ElMessage.error('添加自选股失败')
      }
    }
    
    const refreshData = () => {
      loadStockList()
    }
    
    const syncAllStocks = async () => {
      try {
        await ElMessageBox.confirm('同步全量股票数据可能需要较长时间，是否继续？', '确认同步', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        })
        
        const response = await api.post('/api/v1/tasks/sync/stocks')
        if (response.data.code === 0) {
          ElMessage.success('数据同步任务已启动')
        } else {
          ElMessage.error(response.data.message || '启动同步任务失败')
        }
      } catch (error) {
        if (error !== 'cancel') {
          console.error('同步数据失败:', error)
          ElMessage.error('同步数据失败')
        }
      }
    }
    
    const exportData = () => {
      // TODO: 实现数据导出功能
      ElMessage.info('导出功能开发中...')
    }
    
    const formatDate = (dateStr) => {
      if (!dateStr) return '--'
      return new Date(dateStr).toLocaleDateString()
    }
    
    const getPriceChangeClass = (change) => {
      const changeNum = parseFloat(change)
      if (changeNum > 0) return 'price-up'
      if (changeNum < 0) return 'price-down'
      return 'price-flat'
    }
    
    // 监听搜索关键词变化
    watch(searchKeyword, () => {
      if (searchKeyword.value === '') {
        handleSearch()
      }
    })
    
    // 初始化
    onMounted(() => {
      loadStockList()
    })
    
    return {
      loading,
      stockList,
      realtimeData,
      total,
      currentPage,
      pageSize,
      searchKeyword,
      selectedMarket,
      selectedIndustry,
      loadStockList,
      handleSearch,
      handleFilter,
      handleSizeChange,
      handleCurrentChange,
      handleRowClick,
      goToDetail,
      addToWatchlist,
      refreshData,
      syncAllStocks,
      exportData,
      formatDate,
      getPriceChangeClass
    }
  }
}
</script>

<style lang="scss" scoped>
.stock-list {
  .search-card {
    margin-bottom: 20px;
  }

  .table-card {
    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;

      .header-actions {
        display: flex;
        gap: 8px;
      }
    }
  }

  .stock-table {
    .price {
      font-weight: bold;
    }

    .change {
      font-weight: bold;

      &.price-up {
        color: #f56c6c;
      }

      &.price-down {
        color: #67c23a;
      }

      &.price-flat {
        color: #909399;
      }
    }

    .no-data {
      color: #c0c4cc;
    }
  }
}

@media (max-width: 768px) {
  .stock-list {
    .card-header {
      flex-direction: column;
      align-items: flex-start;
      gap: 12px;

      .header-actions {
        width: 100%;
        justify-content: flex-end;
      }
    }
  }
}
</style>