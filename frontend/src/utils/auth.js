// 前端权限管理工具
import { ElMessage } from 'element-plus'
import api from './api'

// Token管理
export const TOKEN_KEY = 'stock_token'
export const USER_KEY = 'stock_user'
export const PERMISSIONS_KEY = 'stock_permissions'

// 获取Token
export function getToken() {
  return localStorage.getItem(TOKEN_KEY)
}

// 设置Token
export function setToken(token) {
  localStorage.setItem(TOKEN_KEY, token)
}

// 移除Token
export function removeToken() {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(USER_KEY)
  localStorage.removeItem(PERMISSIONS_KEY)
}

// 获取用户信息
export function getUser() {
  const userStr = localStorage.getItem(USER_KEY)
  return userStr ? JSON.parse(userStr) : null
}

// 设置用户信息
export function setUser(user) {
  localStorage.setItem(USER_KEY, JSON.stringify(user))
}

// 获取用户权限
export function getPermissions() {
  const permissionsStr = localStorage.getItem(PERMISSIONS_KEY)
  return permissionsStr ? JSON.parse(permissionsStr) : []
}

// 设置用户权限
export function setPermissions(permissions) {
  localStorage.setItem(PERMISSIONS_KEY, JSON.stringify(permissions))
}

// 检查是否已登录
export function isLoggedIn() {
  return !!getToken()
}

// 检查用户是否有指定权限
export function hasPermission(permissionCode) {
  if (!permissionCode) return true
  
  const permissions = getPermissions()
  return permissions.some(p => p.permission_code === permissionCode)
}

// 检查用户是否有任一权限
export function hasAnyPermission(permissionCodes) {
  if (!permissionCodes || permissionCodes.length === 0) return true
  
  return permissionCodes.some(code => hasPermission(code))
}

// 检查用户是否有所有权限
export function hasAllPermissions(permissionCodes) {
  if (!permissionCodes || permissionCodes.length === 0) return true
  
  return permissionCodes.every(code => hasPermission(code))
}

// 检查用户角色
export function hasRole(roleCode) {
  const user = getUser()
  return user && user.role_code === roleCode
}

// 登录
export async function login(credentials) {
  try {
    // API拦截器已经处理了响应，直接获取data
    const data = await api.post('/auth/login', credentials)
    
    // 检查响应数据是否有效
    if (!data || !data.token || !data.user) {
      throw new Error('登录响应数据无效')
    }
    
    const { token, user, expires_at } = data
    
    // 保存认证信息
    setToken(token)
    setUser(user)
    
    // 获取用户权限
    await loadUserPermissions()
    
    return { success: true, data: data }
  } catch (error) {
    console.error('登录失败:', error)
    
    // 确保清除任何可能的残留数据
    removeToken()
    
    return { 
      success: false, 
      message: error.response?.data?.message || error.message || '登录失败，请稍后重试' 
    }
  }
}

// 登出
export async function logout() {
  try {
    // 调用后端登出接口
    await api.post('/auth/logout')
  } catch (error) {
    console.error('登出请求失败:', error)
  } finally {
    // 清除本地存储
    removeToken()
    // 重定向到登录页
    window.location.href = '/login'
  }
}

// 加载用户权限
export async function loadUserPermissions() {
  try {
    // API拦截器已经处理了响应，直接获取data
    const data = await api.get('/permissions/user/menu')
    
    setPermissions(data.permissions)
    return data.permissions
  } catch (error) {
    console.error('获取用户权限失败:', error)
    return []
  }
}

// 刷新用户信息和权限
export async function refreshUserInfo() {
  try {
    const response = await api.get('/auth/profile')
    
    if (response.data.code === 0) {
      setUser(response.data.data)
      await loadUserPermissions()
      return true
    } else {
      console.error('刷新用户信息失败:', response.data.message)
      return false
    }
  } catch (error) {
    console.error('刷新用户信息失败:', error)
    return false
  }
}

// 权限指令 - 用于Vue模板
export const permissionDirective = {
  mounted(el, binding) {
    const { value } = binding
    
    if (value && !hasPermission(value)) {
      el.parentNode && el.parentNode.removeChild(el)
    }
  },
  updated(el, binding) {
    const { value } = binding
    
    if (value && !hasPermission(value)) {
      el.parentNode && el.parentNode.removeChild(el)
    }
  }
}

// 角色指令 - 用于Vue模板
export const roleDirective = {
  mounted(el, binding) {
    const { value } = binding
    
    if (value && !hasRole(value)) {
      el.parentNode && el.parentNode.removeChild(el)
    }
  },
  updated(el, binding) {
    const { value } = binding
    
    if (value && !hasRole(value)) {
      el.parentNode && el.parentNode.removeChild(el)
    }
  }
}

// 检查Token是否即将过期（提前5分钟提醒）
export function isTokenExpiringSoon() {
  const user = getUser()
  if (!user || !user.expires_at) return false
  
  const expiresAt = new Date(user.expires_at)
  const now = new Date()
  const fiveMinutes = 5 * 60 * 1000
  
  return (expiresAt.getTime() - now.getTime()) < fiveMinutes
}

// 自动刷新Token
export async function autoRefreshToken() {
  if (isTokenExpiringSoon()) {
    try {
      const response = await api.post('/auth/refresh')
      
      if (response.data.code === 0) {
        const { token, expires_at } = response.data.data
        setToken(token)
        
        const user = getUser()
        if (user) {
          user.expires_at = expires_at
          setUser(user)
        }
        
        return true
      }
    } catch (error) {
      console.error('Token刷新失败:', error)
      // Token刷新失败，强制登出
      logout()
      return false
    }
  }
  
  return true
}

// 检查用户是否可以访问指定路径
export function canAccessPath(path) {
  if (!isLoggedIn()) {
    return false
  }
  
  // 公共路径，无需权限
  const publicPaths = ['/dashboard', '/profile', '/']
  if (publicPaths.includes(path)) {
    return true
  }
  
  // 根据路径判断所需权限
  const pathPermissions = {
    '/users': 'user:list',
    '/roles': 'role:list', 
    '/permissions': 'permission:list',
    '/stocks': 'stock:list',
    '/strategies': 'strategy:list'
  }
  
  const requiredPermission = pathPermissions[path]
  if (!requiredPermission) {
    return true // 未定义权限的路径默认允许访问
  }
  
  return hasPermission(requiredPermission)
}