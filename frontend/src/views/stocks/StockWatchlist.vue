<template>
  <div class="stock-watchlist">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>我的自选股</span>
          <el-button type="primary" size="small" @click="addStock">
            <el-icon><Plus /></el-icon>
            添加股票
          </el-button>
        </div>
      </template>
      
      <el-table :data="watchlist" v-loading="loading" stripe>
        <el-table-column prop="ts_code" label="股票代码" width="120">
          <template #default="scope">
            <el-link type="primary" @click="goToDetail(scope.row.ts_code)">
              {{ scope.row.ts_code }}
            </el-link>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="股票名称" />
        <el-table-column label="现价" width="100">
          <template #default="scope">
            <span class="price">{{ scope.row.price || '--' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="涨跌幅" width="100">
          <template #default="scope">
            <span :class="getPriceChangeClass(scope.row.change)">
              {{ scope.row.change || '--' }}%
            </span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120">
          <template #default="scope">
            <el-button size="small" type="text" @click="removeStock(scope.row)">
              <el-icon><Delete /></el-icon>
              移除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'

export default {
  name: 'StockWatchlist',
  setup() {
    const router = useRouter()
    const loading = ref(false)
    const watchlist = ref([
      { ts_code: '000001.SZ', name: '平安银行', price: '12.34', change: '+1.23' },
      { ts_code: '000002.SZ', name: '万科A', price: '18.56', change: '-0.45' }
    ])

    const addStock = () => {
      ElMessage.info('添加股票功能开发中...')
    }

    const removeStock = (stock) => {
      ElMessage.success(`已移除 ${stock.name}`)
    }

    const goToDetail = (code) => {
      router.push(`/stock/${code}`)
    }

    const getPriceChangeClass = (change) => {
      if (!change) return ''
      const changeNum = parseFloat(change)
      if (changeNum > 0) return 'price-up'
      if (changeNum < 0) return 'price-down'
      return 'price-flat'
    }

    return {
      loading,
      watchlist,
      addStock,
      removeStock,
      goToDetail,
      getPriceChangeClass
    }
  }
}
</script>

<style lang="scss" scoped>
.stock-watchlist {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .price {
    font-weight: bold;
  }

  .price-up {
    color: #f56c6c;
  }

  .price-down {
    color: #67c23a;
  }

  .price-flat {
    color: #909399;
  }
}
</style>