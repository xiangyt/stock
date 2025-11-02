import { createRouter, createWebHistory } from 'vue-router'
import { setupRouterGuards } from './guards'

// 导入页面组件
import Dashboard from '../views/Dashboard.vue'

// 股票数据模块
import StockList from '../views/stocks/StockList.vue'
import StockWatchlist from '../views/stocks/StockWatchlist.vue'
import StockRealtime from '../views/stocks/StockRealtime.vue'
import StockDetail from '../views/stocks/StockDetail.vue'

// 策略管理模块
import StrategyList from '../views/strategies/StrategyList.vue'
import StrategyCreate from '../views/strategies/StrategyCreate.vue'
import StrategyBacktest from '../views/strategies/StrategyBacktest.vue'

// 选股结果模块
import SelectionList from '../views/selections/SelectionList.vue'
import SelectionToday from '../views/selections/SelectionToday.vue'
import SelectionHistory from '../views/selections/SelectionHistory.vue'

// 投资组合模块
import PortfolioList from '../views/portfolios/PortfolioList.vue'
import PortfolioPerformance from '../views/portfolios/PortfolioPerformance.vue'

// 数据采集模块
import CollectorList from '../views/collectors/CollectorList.vue'
import CollectorTasks from '../views/collectors/CollectorTasks.vue'
import CollectorSync from '../views/collectors/CollectorSync.vue'

// 通知管理模块
import NotificationRobots from '../views/notifications/NotificationRobots.vue'
import NotificationTemplates from '../views/notifications/NotificationTemplates.vue'
import NotificationLogs from '../views/notifications/NotificationLogs.vue'

// 报表分析模块
import ReportStrategy from '../views/reports/ReportStrategy.vue'
import ReportMarket from '../views/reports/ReportMarket.vue'
import ReportRisk from '../views/reports/ReportRisk.vue'

// 系统管理模块
import SystemUsers from '../views/system/SystemUsers.vue'
import UserManagement from '../views/UserManagement.vue'
import RoleManagement from '../views/RoleManagement.vue'
import SystemMonitoring from '../views/system/SystemMonitoring.vue'
import SystemSettings from '../views/system/SystemSettings.vue'

const routes = [
  {
    path: '/',
    name: 'Dashboard',
    component: Dashboard,
    meta: { 
      title: '仪表板',
      requiresAuth: true,
      permission: 'dashboard:view'
    }
  },
  
  // 股票数据路由
  {
    path: '/stocks',
    name: 'StockList',
    component: StockList,
    meta: { 
      title: '股票列表',
      requiresAuth: true,
      permission: 'stocks:list'
    }
  },
  {
    path: '/stocks/watchlist',
    name: 'StockWatchlist',
    component: StockWatchlist,
    meta: { 
      title: '自选股',
      requiresAuth: true,
      permission: 'stocks:watchlist'
    }
  },
  {
    path: '/stocks/realtime',
    name: 'StockRealtime',
    component: StockRealtime,
    meta: { 
      title: '实时行情',
      requiresAuth: true,
      permission: 'stocks:realtime'
    }
  },
  {
    path: '/stock/:code',
    name: 'StockDetail',
    component: StockDetail,
    props: true,
    meta: { 
      title: '股票详情',
      requiresAuth: true,
      permission: 'stocks:detail'
    }
  },
  
  // 策略管理路由
  {
    path: '/strategies',
    name: 'StrategyList',
    component: StrategyList,
    meta: { 
      title: '策略列表',
      requiresAuth: true,
      permission: 'strategies:list'
    }
  },
  {
    path: '/strategies/create',
    name: 'StrategyCreate',
    component: StrategyCreate,
    meta: { 
      title: '创建策略',
      requiresAuth: true,
      permission: 'strategies:create'
    }
  },
  {
    path: '/strategies/backtest',
    name: 'StrategyBacktest',
    component: StrategyBacktest,
    meta: { 
      title: '回测中心',
      requiresAuth: true,
      permission: 'strategies:backtest'
    }
  },
  
  // 选股结果路由
  {
    path: '/selections',
    name: 'SelectionList',
    component: SelectionList,
    meta: { 
      title: '选股记录',
      requiresAuth: true,
      permission: 'selections:list'
    }
  },
  {
    path: '/selections/today',
    name: 'SelectionToday',
    component: SelectionToday,
    meta: { 
      title: '今日选股',
      requiresAuth: true,
      permission: 'selections:today'
    }
  },
  {
    path: '/selections/history',
    name: 'SelectionHistory',
    component: SelectionHistory,
    meta: { 
      title: '历史回顾',
      requiresAuth: true,
      permission: 'selections:history'
    }
  },
  
  // 投资组合路由
  {
    path: '/portfolios',
    name: 'PortfolioList',
    component: PortfolioList,
    meta: { 
      title: '组合管理',
      requiresAuth: true,
      permission: 'portfolios:manage'
    }
  },
  {
    path: '/portfolios/performance',
    name: 'PortfolioPerformance',
    component: PortfolioPerformance,
    meta: { 
      title: '业绩分析',
      requiresAuth: true,
      permission: 'portfolios:performance'
    }
  },
  
  // 数据采集路由
  {
    path: '/collectors',
    name: 'CollectorList',
    component: CollectorList,
    meta: { 
      title: '采集器管理',
      requiresAuth: true,
      permission: 'collectors:manage'
    }
  },
  {
    path: '/collectors/tasks',
    name: 'CollectorTasks',
    component: CollectorTasks,
    meta: { 
      title: '采集任务',
      requiresAuth: true,
      permission: 'collectors:tasks'
    }
  },
  {
    path: '/collectors/sync',
    name: 'CollectorSync',
    component: CollectorSync,
    meta: { 
      title: '数据同步',
      requiresAuth: true,
      permission: 'collectors:sync'
    }
  },
  
  // 通知管理路由
  {
    path: '/notifications/robots',
    name: 'NotificationRobots',
    component: NotificationRobots,
    meta: { 
      title: '机器人配置',
      requiresAuth: true,
      permission: 'notifications:robots'
    }
  },
  {
    path: '/notifications/templates',
    name: 'NotificationTemplates',
    component: NotificationTemplates,
    meta: { 
      title: '消息模板',
      requiresAuth: true,
      permission: 'notifications:templates'
    }
  },
  {
    path: '/notifications/logs',
    name: 'NotificationLogs',
    component: NotificationLogs,
    meta: { 
      title: '发送日志',
      requiresAuth: true,
      permission: 'notifications:logs'
    }
  },
  
  // 报表分析路由
  {
    path: '/reports/strategy',
    name: 'ReportStrategy',
    component: ReportStrategy,
    meta: { 
      title: '策略报告',
      requiresAuth: true,
      permission: 'reports:strategy'
    }
  },
  {
    path: '/reports/market',
    name: 'ReportMarket',
    component: ReportMarket,
    meta: { 
      title: '市场分析',
      requiresAuth: true,
      permission: 'reports:market'
    }
  },
  {
    path: '/reports/risk',
    name: 'ReportRisk',
    component: ReportRisk,
    meta: { 
      title: '风险分析',
      requiresAuth: true,
      permission: 'reports:risk'
    }
  },
  
  // 系统管理路由
  {
    path: '/system/users',
    name: 'UserManagement',
    component: UserManagement,
    meta: { 
      title: '用户管理',
      requiresAuth: true,
      permission: 'system:users'
    }
  },
  {
    path: '/system/roles',
    name: 'RoleManagement',
    component: RoleManagement,
    meta: { 
      title: '角色管理',
      requiresAuth: true,
      permission: 'system:roles'
    }
  },
  {
    path: '/system/monitoring',
    name: 'SystemMonitoring',
    component: SystemMonitoring,
    meta: { 
      title: '系统监控',
      requiresAuth: true,
      permission: 'system:monitoring'
    }
  },
  {
    path: '/system/settings',
    name: 'SystemSettings',
    component: SystemSettings,
    meta: { 
      title: '系统设置',
      requiresAuth: true,
      permission: 'system:settings'
    }
  },
  
  // 登录页面（不需要权限）
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue'),
    meta: { 
      title: '登录',
      requiresAuth: false
    }
  },
  
  // 403 无权限页面
  {
    path: '/403',
    name: 'Forbidden',
    component: () => import('../views/Forbidden.vue'),
    meta: { 
      title: '无权限访问',
      requiresAuth: false
    }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 设置路由守卫
setupRouterGuards(router)

export default router