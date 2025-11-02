<template>
  <div class="dashboard">
    <!-- 顶部统计卡片 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :xs="24" :sm="12" :md="6" v-for="stat in stats" :key="stat.title">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" :style="{ backgroundColor: stat.color }">
              <el-icon :size="24"><component :is="stat.icon" /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stat.value }}</div>
              <div class="stat-title">{{ stat.title }}</div>
              <div class="stat-change" :class="stat.changeType">
                <el-icon><component :is="stat.changeIcon" /></el-icon>
                {{ stat.change }}
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 图表区域 -->
    <el-row :gutter="20" class="charts-row">
      <!-- 市场概览 -->
      <el-col :xs="24" :lg="12">
        <el-card class="chart-card">
          <template #header>
            <div class="card-header">
              <span>市场概览</span>
              <el-button type="text" @click="refreshMarketData">
                <el-icon><Refresh /></el-icon>
              </el-button>
            </div>
          </template>
          <div class="market-overview">
            <div class="market-item" v-for="market in marketData" :key="market.name">
              <div class="market-name">{{ market.name }}</div>
              <div class="market-value" :class="market.changeType">{{ market.value }}</div>
              <div class="market-change" :class="market.changeType">{{ market.change }}</div>
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- 策略表现 -->
      <el-col :xs="24" :lg="12">
        <el-card class="chart-card">
          <template #header>
            <div class="card-header">
              <span>策略表现</span>
              <el-select v-model="selectedPeriod" size="small" style="width: 100px">
                <el-option label="7天" value="7d" />
                <el-option label="30天" value="30d" />
                <el-option label="90天" value="90d" />
              </el-select>
            </div>
          </template>
          <div class="strategy-performance">
            <div class="performance-item" v-for="strategy in strategyData" :key="strategy.name">
              <div class="strategy-info">
                <div class="strategy-name">{{ strategy.name }}</div>
                <div class="strategy-return" :class="strategy.returnType">{{ strategy.return }}</div>
              </div>
              <div class="strategy-progress">
                <el-progress 
                  :percentage="Math.abs(strategy.return)" 
                  :color="strategy.returnType === 'positive' ? '#67c23a' : '#f56c6c'"
                  :show-text="false"
                />
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 最新选股和任务状态 -->
    <el-row :gutter="20" class="content-row">
      <!-- 今日选股 -->
      <el-col :xs="24" :lg="12">
        <el-card class="content-card">
          <template #header>
            <div class="card-header">
              <span>今日选股</span>
              <el-button type="primary" size="small" @click="goToSelections">查看全部</el-button>
            </div>
          </template>
          <el-table :data="todaySelections" style="width: 100%" size="small">
            <el-table-column prop="code" label="股票代码" width="100" />
            <el-table-column prop="name" label="股票名称" />
            <el-table-column prop="strategy" label="策略" width="120" />
            <el-table-column prop="score" label="评分" width="80">
              <template #default="scope">
                <el-tag :type="getScoreType(scope.row.score)">{{ scope.row.score }}</el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>

      <!-- 系统状态 -->
      <el-col :xs="24" :lg="12">
        <el-card class="content-card">
          <template #header>
            <div class="card-header">
              <span>系统状态</span>
              <el-button type="text" @click="refreshSystemStatus">
                <el-icon><Refresh /></el-icon>
              </el-button>
            </div>
          </template>
          <div class="system-status">
            <div class="status-item" v-for="item in systemStatus" :key="item.name">
              <div class="status-info">
                <span class="status-name">{{ item.name }}</span>
                <el-tag :type="item.status === 'running' ? 'success' : 'danger'" size="small">
                  {{ item.status === 'running' ? '运行中' : '已停止' }}
                </el-tag>
              </div>
              <div class="status-detail">{{ item.detail }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 快捷操作 -->
    <el-row :gutter="20" class="actions-row">
      <el-col :span="24">
        <el-card class="actions-card">
          <template #header>
            <span>快捷操作</span>
          </template>
          <div class="quick-actions">
            <el-button 
              v-for="action in quickActions" 
              :key="action.name"
              :type="action.type"
              :icon="action.icon"
              @click="action.handler"
              class="action-btn"
            >
              {{ action.name }}
            </el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import api from '../utils/api'

export default {
  name: 'Dashboard',
  setup() {
    const router = useRouter()
    
    // 统计数据
    const stats = ref([
      {
        title: '总股票数',
        value: '5,146',
        change: '+12',
        changeType: 'positive',
        changeIcon: 'ArrowUp',
        icon: 'DataLine',
        color: '#409eff'
      },
      {
        title: '活跃策略',
        value: '8',
        change: '+2',
        changeType: 'positive',
        changeIcon: 'ArrowUp',
        icon: 'SetUp',
        color: '#67c23a'
      },
      {
        title: '今日选股',
        value: '23',
        change: '-5',
        changeType: 'negative',
        changeIcon: 'ArrowDown',
        icon: 'Select',
        color: '#e6a23c'
      },
      {
        title: '投资组合',
        value: '3',
        change: '0',
        changeType: 'neutral',
        changeIcon: 'Minus',
        icon: 'PieChart',
        color: '#f56c6c'
      }
    ])

    // 市场数据
    const marketData = ref([
      { name: '上证指数', value: '3,245.67', change: '+1.23%', changeType: 'positive' },
      { name: '深证成指', value: '12,456.78', change: '-0.45%', changeType: 'negative' },
      { name: '创业板指', value: '2,678.90', change: '+0.78%', changeType: 'positive' },
      { name: '科创50', value: '1,234.56', change: '+2.15%', changeType: 'positive' }
    ])

    // 策略表现数据
    const selectedPeriod = ref('30d')
    const strategyData = ref([
      { name: '价值投资策略', return: 15.6, returnType: 'positive' },
      { name: '成长股策略', return: 8.9, returnType: 'positive' },
      { name: '技术分析策略', return: -2.3, returnType: 'negative' },
      { name: '量化策略', return: 12.1, returnType: 'positive' }
    ])

    // 今日选股
    const todaySelections = ref([
      { code: '000001', name: '平安银行', strategy: '价值投资', score: 85 },
      { code: '000002', name: '万科A', strategy: '成长股', score: 78 },
      { code: '600036', name: '招商银行', strategy: '价值投资', score: 92 },
      { code: '000858', name: '五粮液', strategy: '消费股', score: 88 }
    ])

    // 系统状态
    const systemStatus = ref([
      { name: '数据采集服务', status: 'running', detail: '最后更新: 2分钟前' },
      { name: '策略执行引擎', status: 'running', detail: '运行中 - 8个策略' },
      { name: '通知服务', status: 'running', detail: '已发送 15 条消息' },
      { name: '数据库连接', status: 'running', detail: '连接正常' }
    ])

    // 快捷操作
    const quickActions = ref([
      { name: '创建策略', type: 'primary', icon: 'Plus', handler: () => router.push('/strategies/create') },
      { name: '数据同步', type: 'success', icon: 'Refresh', handler: () => syncData() },
      { name: '发送通知', type: 'warning', icon: 'Bell', handler: () => sendNotification() },
      { name: '生成报告', type: 'info', icon: 'Document', handler: () => generateReport() }
    ])

    // 方法
    const refreshMarketData = async () => {
      try {
        // TODO: 调用API获取市场数据
        console.log('刷新市场数据')
      } catch (error) {
        console.error('刷新市场数据失败:', error)
      }
    }

    const refreshSystemStatus = async () => {
      try {
        // TODO: 调用API获取系统状态
        console.log('刷新系统状态')
      } catch (error) {
        console.error('刷新系统状态失败:', error)
      }
    }

    const goToSelections = () => {
      router.push('/selections/today')
    }

    const getScoreType = (score) => {
      if (score >= 90) return 'success'
      if (score >= 80) return 'warning'
      return 'info'
    }

    const syncData = () => {
      console.log('同步数据')
      router.push('/collectors/sync')
    }

    const sendNotification = () => {
      console.log('发送通知')
      router.push('/notifications/robots')
    }

    const generateReport = () => {
      console.log('生成报告')
      router.push('/reports/strategy')
    }

    // 初始化数据
    onMounted(() => {
      // TODO: 加载仪表板数据
    })

    return {
      stats,
      marketData,
      selectedPeriod,
      strategyData,
      todaySelections,
      systemStatus,
      quickActions,
      refreshMarketData,
      refreshSystemStatus,
      goToSelections,
      getScoreType,
      syncData,
      sendNotification,
      generateReport
    }
  }
}
</script>

<style lang="scss" scoped>
.dashboard {
  .stats-row {
    margin-bottom: 20px;
  }

  .stat-card {
    .stat-content {
      display: flex;
      align-items: center;

      .stat-icon {
        width: 60px;
        height: 60px;
        border-radius: 12px;
        display: flex;
        align-items: center;
        justify-content: center;
        color: white;
        margin-right: 16px;
      }

      .stat-info {
        flex: 1;

        .stat-value {
          font-size: 24px;
          font-weight: bold;
          color: #303133;
          margin-bottom: 4px;
        }

        .stat-title {
          font-size: 14px;
          color: #909399;
          margin-bottom: 4px;
        }

        .stat-change {
          font-size: 12px;
          display: flex;
          align-items: center;

          &.positive {
            color: #67c23a;
          }

          &.negative {
            color: #f56c6c;
          }

          &.neutral {
            color: #909399;
          }

          .el-icon {
            margin-right: 2px;
          }
        }
      }
    }
  }

  .charts-row,
  .content-row,
  .actions-row {
    margin-bottom: 20px;
  }

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .market-overview {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 16px;

    .market-item {
      text-align: center;
      padding: 16px;
      border: 1px solid #ebeef5;
      border-radius: 8px;

      .market-name {
        font-size: 14px;
        color: #909399;
        margin-bottom: 8px;
      }

      .market-value {
        font-size: 18px;
        font-weight: bold;
        margin-bottom: 4px;

        &.positive {
          color: #f56c6c;
        }

        &.negative {
          color: #67c23a;
        }
      }

      .market-change {
        font-size: 12px;

        &.positive {
          color: #f56c6c;
        }

        &.negative {
          color: #67c23a;
        }
      }
    }
  }

  .strategy-performance {
    .performance-item {
      display: flex;
      align-items: center;
      margin-bottom: 16px;

      .strategy-info {
        width: 150px;
        margin-right: 16px;

        .strategy-name {
          font-size: 14px;
          color: #303133;
          margin-bottom: 4px;
        }

        .strategy-return {
          font-size: 16px;
          font-weight: bold;

          &.positive {
            color: #67c23a;
          }

          &.negative {
            color: #f56c6c;
          }

          &::after {
            content: '%';
          }
        }
      }

      .strategy-progress {
        flex: 1;
      }
    }
  }

  .system-status {
    .status-item {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 12px 0;
      border-bottom: 1px solid #ebeef5;

      &:last-child {
        border-bottom: none;
      }

      .status-info {
        display: flex;
        align-items: center;

        .status-name {
          margin-right: 12px;
          font-size: 14px;
          color: #303133;
        }
      }

      .status-detail {
        font-size: 12px;
        color: #909399;
      }
    }
  }

  .quick-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;

    .action-btn {
      min-width: 120px;
    }
  }
}

@media (max-width: 768px) {
  .dashboard {
    .market-overview {
      grid-template-columns: 1fr;
    }

    .quick-actions {
      .action-btn {
        flex: 1;
        min-width: auto;
      }
    }
  }
}
</style>