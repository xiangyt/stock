// 路由权限守卫
import router from './index'
import { getToken, hasPermission, canAccessPath } from '@/utils/auth'
import { ElMessage } from 'element-plus'

// 白名单路由（不需要登录即可访问）
const whiteList = ['/login', '/register', '/404', '/403']

// 路由前置守卫
router.beforeEach(async (to, from, next) => {
  // 获取Token
  const token = getToken()
  
  if (token) {
    // 已登录
    if (to.path === '/login') {
      // 如果已登录，访问登录页则重定向到首页
      next({ path: '/' })
    } else {
      // 检查路由权限
      if (canAccessPath(to.path)) {
        next()
      } else {
        ElMessage.error('您没有权限访问该页面')
        next({ path: '/403' })
      }
    }
  } else {
    // 未登录
    if (whiteList.includes(to.path)) {
      // 在白名单中，直接放行
      next()
    } else {
      // 不在白名单中，重定向到登录页
      ElMessage.warning('请先登录')
      next({ path: '/login', query: { redirect: to.fullPath } })
    }
  }
})

// 路由后置守卫
router.afterEach((to) => {
  // 设置页面标题
  document.title = to.meta.title ? `${to.meta.title} - 股票系统` : '股票系统'
})

export default router