import axios from 'axios'
import { ElMessage } from 'element-plus'

// 创建axios实例
const api = axios.create({
  baseURL: 'http://localhost:8080/api/v1',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器
api.interceptors.request.use(
  config => {
    // 可以在这里添加token
    const token = localStorage.getItem('stock_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  error => {
    return Promise.reject(error)
  }
)

// 响应拦截器
api.interceptors.response.use(
  response => {
    const { data } = response
    if (data.code === 0) {
      return data.data
    } else {
      ElMessage.error(data.message || '请求失败')
      return Promise.reject(new Error(data.message || '请求失败'))
    }
  },
  error => {
    let message = '网络错误'
    if (error.response) {
      switch (error.response.status) {
        case 401:
          message = '未授权，请重新登录'
          // 可以在这里处理登录跳转
          break
        case 403:
          message = '拒绝访问'
          break
        case 404:
          message = '请求地址出错'
          break
        case 408:
          message = '请求超时'
          break
        case 500:
          message = '服务器内部错误'
          break
        case 501:
          message = '服务未实现'
          break
        case 502:
          message = '网关错误'
          break
        case 503:
          message = '服务不可用'
          break
        case 504:
          message = '网关超时'
          break
        case 505:
          message = 'HTTP版本不受支持'
          break
        default:
          message = `连接错误${error.response.status}`
      }
    } else {
      message = '连接到服务器失败'
    }
    ElMessage.error(message)
    return Promise.reject(error)
  }
)

// API方法
export const stockAPI = {
  // 获取股票列表
  getStockList: (params = {}) => api.get('/stocks', { params }),
  
  // 获取股票详情
  getStockDetail: (code) => api.get(`/stocks/${code}`),
  
  // 获取K线数据
  getKLineData: (code, params = {}) => api.get(`/stocks/${code}/kline`, { params }),
  
  // 获取实时数据
  getRealtimeData: (params = {}) => api.get('/realtime', { params }),
  
  // 添加到自选股
  addToWatchlist: (code) => api.post(`/stocks/${code}/watchlist`),
  
  // 从自选股移除
  removeFromWatchlist: (code) => api.delete(`/stocks/${code}/watchlist`)
}

export const strategyAPI = {
  // 获取策略列表
  getStrategies: (params = {}) => api.get('/strategies', { params }),
  
  // 创建策略
  createStrategy: (data) => api.post('/strategies', data),
  
  // 获取策略详情
  getStrategy: (id) => api.get(`/strategies/${id}`),
  
  // 更新策略
  updateStrategy: (id, data) => api.put(`/strategies/${id}`, data),
  
  // 删除策略
  deleteStrategy: (id) => api.delete(`/strategies/${id}`),
  
  // 运行回测
  runBacktest: (id, params) => api.post(`/strategies/${id}/backtest`, params),
  
  // 执行策略
  executeStrategy: (id) => api.post(`/strategies/${id}/execute`)
}

export const selectionAPI = {
  // 获取选股记录
  getSelections: (params = {}) => api.get('/selections', { params }),
  
  // 获取今日选股
  getTodaySelections: () => api.get('/selections/today'),
  
  // 获取历史选股
  getSelectionHistory: (params = {}) => api.get('/selections/history', { params }),
  
  // 创建选股记录
  createSelection: (data) => api.post('/selections', data),
  
  // 导出选股结果
  exportSelections: (params = {}) => api.get('/selections/export', { params })
}

export const portfolioAPI = {
  // 获取投资组合列表
  getPortfolios: (params = {}) => api.get('/portfolios', { params }),
  
  // 创建投资组合
  createPortfolio: (data) => api.post('/portfolios', data),
  
  // 获取投资组合详情
  getPortfolio: (id) => api.get(`/portfolios/${id}`),
  
  // 更新投资组合
  updatePortfolio: (id, data) => api.put(`/portfolios/${id}`, data),
  
  // 删除投资组合
  deletePortfolio: (id) => api.delete(`/portfolios/${id}`),
  
  // 获取组合业绩
  getPortfolioPerformance: (id, params = {}) => api.get(`/portfolios/${id}/performance`, { params }),
  
  // 添加股票到组合
  addStockToPortfolio: (id, data) => api.post(`/portfolios/${id}/stocks`, data),
  
  // 从组合移除股票
  removeStockFromPortfolio: (id, stockCode) => api.delete(`/portfolios/${id}/stocks/${stockCode}`)
}

export const notificationAPI = {
  // 获取机器人列表
  getRobots: (params = {}) => api.get('/notifications/robots', { params }),
  
  // 创建机器人
  createRobot: (data) => api.post('/notifications/robots', data),
  
  // 获取机器人详情
  getRobot: (id) => api.get(`/notifications/robots/${id}`),
  
  // 更新机器人
  updateRobot: (id, data) => api.put(`/notifications/robots/${id}`, data),
  
  // 删除机器人
  deleteRobot: (id) => api.delete(`/notifications/robots/${id}`),
  
  // 测试机器人
  testRobot: (id) => api.post(`/notifications/robots/${id}/test`),
  
  // 发送通知
  sendNotification: (data) => api.post('/notifications/send', data),
  
  // 获取通知日志
  getNotificationLogs: (params = {}) => api.get('/notifications/logs', { params }),
  
  // 获取消息模板
  getMessageTemplates: () => api.get('/notifications/templates'),
  
  // 创建消息模板
  createMessageTemplate: (data) => api.post('/notifications/templates', data)
}

export const collectorAPI = {
  // 获取采集器列表
  getCollectors: () => api.get('/collectors'),
  
  // 获取采集器状态
  getCollectorStatus: (name) => api.get(`/collectors/${name}/status`),
  
  // 同步数据
  syncData: (source) => api.post(`/collectors/sync/${source}`),
  
  // 获取采集任务
  getCollectorTasks: (params = {}) => api.get('/collectors/tasks', { params }),
  
  // 创建采集任务
  createCollectorTask: (data) => api.post('/collectors/tasks', data),
  
  // 更新采集任务
  updateCollectorTask: (id, data) => api.put(`/collectors/tasks/${id}`, data),
  
  // 删除采集任务
  deleteCollectorTask: (id) => api.delete(`/collectors/tasks/${id}`)
}

export const reportAPI = {
  // 获取策略报告
  getStrategyReport: (params = {}) => api.get('/reports/strategy', { params }),
  
  // 获取市场分析报告
  getMarketReport: (params = {}) => api.get('/reports/market', { params }),
  
  // 获取风险分析报告
  getRiskReport: (params = {}) => api.get('/reports/risk', { params }),
  
  // 获取仪表板数据
  getDashboardData: () => api.get('/reports/dashboard')
}

export const systemAPI = {
  // 获取用户列表
  getUsers: (params = {}) => api.get('/users', { params }),
  
  // 创建用户
  createUser: (data) => api.post('/users', data),
  
  // 获取用户详情
  getUser: (id) => api.get(`/users/${id}`),
  
  // 更新用户
  updateUser: (id, data) => api.put(`/users/${id}`, data),
  
  // 删除用户
  deleteUser: (id) => api.delete(`/users/${id}`),
  
  // 获取系统状态
  getSystemStatus: () => api.get('/monitoring/status'),
  
  // 获取API统计
  getAPIStats: () => api.get('/monitoring/api-stats'),
  
  // 获取系统日志
  getSystemLogs: (params = {}) => api.get('/monitoring/logs', { params }),
  
  // 获取系统设置
  getSettings: () => api.get('/settings'),
  
  // 更新系统设置
  updateSettings: (data) => api.put('/settings', data)
}

export const authAPI = {
  // 登录
  login: (data) => api.post('/auth/login', data),
  
  // 登出
  logout: () => api.post('/auth/logout'),
  
  // 刷新token
  refreshToken: () => api.post('/auth/refresh'),
  
  // 获取用户信息
  getProfile: () => api.get('/auth/profile')
}

// 用户管理API
export const userAPI = {
  getUserList: (params = {}) => api.get('/users', { params }),
  createUser: (data) => api.post('/users', data),
  getUserDetail: (id) => api.get(`/users/${id}`),
  updateUser: (id, data) => api.put(`/users/${id}`, data),
  deleteUser: (id) => api.delete(`/users/${id}`),
  getUserLogs: (id, params = {}) => api.get(`/users/${id}/logs`, { params })
}

// 角色管理API
export const roleAPI = {
  getRoleList: (params = {}) => api.get('/roles', { params }),
  createRole: (data) => api.post('/roles', data),
  getRoleDetail: (id) => api.get(`/roles/${id}`),
  updateRole: (id, data) => api.put(`/roles/${id}`, data),
  deleteRole: (id) => api.delete(`/roles/${id}`),
  getRolePermissions: (id) => api.get(`/roles/${id}/permissions`),
  updateRolePermissions: (id, data) => api.put(`/roles/${id}/permissions`, data)
}

// 权限管理API
export const permissionAPI = {
  getPermissionList: (params = {}) => api.get('/permissions', { params }),
  createPermission: (data) => api.post('/permissions', data),
  getPermissionDetail: (id) => api.get(`/permissions/${id}`),
  updatePermission: (id, data) => api.put(`/permissions/${id}`, data),
  deletePermission: (id) => api.delete(`/permissions/${id}`)
}

export default api