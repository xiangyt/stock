import axios from 'axios'
import { ElMessage } from 'element-plus'

// 创建axios实例
const api = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器
api.interceptors.request.use(
  config => {
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
      return data
    } else {
      ElMessage.error(data.message || '请求失败')
      return Promise.reject(new Error(data.message || '请求失败'))
    }
  },
  error => {
    let message = '网络错误'
    if (error.response) {
      message = error.response.data?.message || `请求失败 (${error.response.status})`
    } else if (error.request) {
      message = '网络连接失败'
    }
    ElMessage.error(message)
    return Promise.reject(error)
  }
)

// API方法
export default {
  // 获取股票列表
  getStockList(params = {}) {
    return api.get('/stocks', { params })
  },

  // 获取股票详情
  getStockDetail(code) {
    return api.get(`/stocks/${code}`)
  },

  // 获取K线数据
  getKLineData(code, params = {}) {
    return api.get(`/stocks/${code}/kline`, { params })
  },

  // 获取财务数据
  getFinancialData(code) {
    return api.get(`/stocks/${code}/financial`)
  },

  // 获取实时数据
  getRealtimeData(codes) {
    return api.get('/realtime', {
      params: { codes: Array.isArray(codes) ? codes.join(',') : codes }
    })
  }
}