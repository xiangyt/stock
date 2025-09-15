<template>
  <div class="home">
    <!-- 欢迎区域 -->
    <el-card class="welcome-card">
      <div class="welcome-content">
        <h1>
          <el-icon><TrendCharts /></el-icon>
          欢迎使用智能选股系统
        </h1>
        <p>基于Go语言后端 + Vue3前端的高性能股票分析工具</p>
        <div class="action-buttons">
          <el-button type="primary" size="large" @click="$router.push('/stocks')">
            <el-icon><DataLine /></el-icon>
            查看股票列表
          </el-button>
          <el-button type="success" size="large" @click="$router.push('/analysis')">
            <el-icon><PieChart /></el-icon>
            数据分析
          </el-button>
        </div>
      </div>
    </el-card>

    <!-- 功能介绍 -->
    <div class="features">
      <el-row :gutter="20">
        <el-col :span="8">
          <el-card class="feature-card">
            <div class="feature-icon">
              <el-icon><DataLine /></el-icon>
            </div>
            <h3>实时数据</h3>
            <p>获取A股市场实时行情数据，包括股价、成交量等关键指标</p>
          </el-card>
        </el-col>
        <el-col :span="8">
          <el-card class="feature-card">
            <div class="feature-icon">
              <el-icon><PieChart /></el-icon>
            </div>
            <h3>技术分析</h3>
            <p>提供多种技术指标分析，包括MA、MACD、RSI等经典指标</p>
          </el-card>
        </el-col>
        <el-col :span="8">
          <el-card class="feature-card">
            <div class="feature-icon">
              <el-icon><TrendCharts /></el-icon>
            </div>
            <h3>智能选股</h3>
            <p>基于多维度分析的智能选股策略，帮助发现投资机会</p>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- API测试区域 -->
    <el-card class="api-test-card">
      <template #header>
        <div class="card-header">
          <span>API接口测试</span>
          <el-button type="primary" @click="testAllAPIs">测试所有接口</el-button>
        </div>
      </template>
      
      <div class="api-tests">
        <el-row :gutter="20">
          <el-col :span="12">
            <div class="api-item">
              <h4>股票列表</h4>
              <el-button @click="testStockList" :loading="loading.stockList">
                测试接口
              </el-button>
              <div v-if="results.stockList" class="result">
                <el-tag type="success">成功获取 {{ results.stockList.total }} 只股票</el-tag>
              </div>
            </div>
          </el-col>
          <el-col :span="12">
            <div class="api-item">
              <h4>股票详情</h4>
              <el-input v-model="testCode" placeholder="输入股票代码" style="width: 150px; margin-right: 10px;" />
              <el-button @click="testStockDetail" :loading="loading.stockDetail">
                测试接口
              </el-button>
              <div v-if="results.stockDetail" class="result">
                <el-tag type="success">{{ results.stockDetail.name }}</el-tag>
              </div>
            </div>
          </el-col>
        </el-row>
      </div>
    </el-card>
  </div>
</template>

<script>
import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import api from '../utils/api'

export default {
  name: 'Home',
  setup() {
    const testCode = ref('000001.SZ')
    const loading = reactive({
      stockList: false,
      stockDetail: false
    })
    const results = reactive({
      stockList: null,
      stockDetail: null
    })

    const testStockList = async () => {
      loading.stockList = true
      try {
        const response = await api.getStockList({ page: 1, size: 10 })
        results.stockList = response.data
        ElMessage.success('股票列表接口测试成功')
      } catch (error) {
        ElMessage.error('股票列表接口测试失败: ' + error.message)
      } finally {
        loading.stockList = false
      }
    }

    const testStockDetail = async () => {
      if (!testCode.value) {
        ElMessage.warning('请输入股票代码')
        return
      }
      loading.stockDetail = true
      try {
        const response = await api.getStockDetail(testCode.value)
        results.stockDetail = response.data
        ElMessage.success('股票详情接口测试成功')
      } catch (error) {
        ElMessage.error('股票详情接口测试失败: ' + error.message)
      } finally {
        loading.stockDetail = false
      }
    }

    const testAllAPIs = async () => {
      await testStockList()
      await testStockDetail()
    }

    return {
      testCode,
      loading,
      results,
      testStockList,
      testStockDetail,
      testAllAPIs
    }
  }
}
</script>

<style lang="scss" scoped>
.home {
  max-width: 1200px;
  margin: 0 auto;
}

.welcome-card {
  margin-bottom: 30px;
  text-align: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;

  .welcome-content {
    padding: 40px 20px;

    h1 {
      font-size: 2.5em;
      margin-bottom: 20px;
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 15px;

      .el-icon {
        font-size: 1.2em;
      }
    }

    p {
      font-size: 1.2em;
      margin-bottom: 30px;
      opacity: 0.9;
    }

    .action-buttons {
      display: flex;
      gap: 20px;
      justify-content: center;

      .el-button {
        padding: 15px 30px;
        font-size: 16px;
      }
    }
  }
}

.features {
  margin-bottom: 30px;

  .feature-card {
    text-align: center;
    height: 200px;
    display: flex;
    flex-direction: column;
    justify-content: center;

    .feature-icon {
      font-size: 48px;
      color: #409eff;
      margin-bottom: 15px;
    }

    h3 {
      margin-bottom: 10px;
      color: #303133;
    }

    p {
      color: #606266;
      line-height: 1.6;
    }
  }
}

.api-test-card {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .api-tests {
    .api-item {
      padding: 20px;
      border: 1px solid #ebeef5;
      border-radius: 8px;
      margin-bottom: 15px;

      h4 {
        margin-bottom: 15px;
        color: #303133;
      }

      .result {
        margin-top: 10px;
      }
    }
  }
}
</style>