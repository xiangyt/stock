// 路由守卫
import { isLoggedIn, canAccessPath } from '@/utils/auth'
import { ElMessage } from 'element-plus'

export function setupRouterGuards(router) {
  // 前置守卫
  router.beforeEach((to, from, next) => {
    // 检查是否需要登录
    if (to.meta.requiresAuth !== false) {
      if (!isLoggedIn()) {
        ElMessage.warning('请先登录')
        next({ path: '/login', query: { redirect: to.fullPath } })
        return
      }
      
      // 检查权限
      if (to.meta.permission && !canAccessPath(to.path)) {
        ElMessage.error('您没有权限访问该页面')
        next({ path: '/403' })
        return
      }
    }
    
    // 如果已登录且访问登录页，重定向到首页
    if (to.path === '/login' && isLoggedIn()) {
      next({ path: '/' })
      return
    }
    
    next()
  })
  
  // 后置守卫
  router.afterEach((to) => {
    // 设置页面标题
    document.title = to.meta.title ? `${to.meta.title} - 股票系统` : '股票系统'
  })
}