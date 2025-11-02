// 前端菜单管理工具
import { hasPermission } from './auth'
import api from './api'
import { MENU_STRUCTURE, MENU_ICONS } from '@/config/menu'

// 使用统一的菜单结构
const DEFAULT_MENU_STRUCTURE = MENU_STRUCTURE

// 从API获取用户菜单权限
export async function getUserMenus() {
  try {
    const response = await api.get('/permissions/user/menu')
    
    // API拦截器已经处理了响应，直接使用返回的数据
    console.log('获取到的菜单响应数据:', response)
    
    if (response && response.permissions) {
      const permissions = response.permissions
      console.log('获取到的权限数据:', permissions)
      return buildMenuFromPermissions(permissions)
    } else {
      console.warn('获取用户菜单失败，使用默认菜单')
      return filterMenuByPermissions(DEFAULT_MENU_STRUCTURE)
    }
  } catch (error) {
    console.warn('获取用户菜单失败，使用默认菜单:', error)
    return filterMenuByPermissions(DEFAULT_MENU_STRUCTURE)
  }
}

// 根据权限数据构建菜单
function buildMenuFromPermissions(permissions) {
  console.log('开始构建菜单，权限数据:', permissions)
  
  // 使用预定义的菜单结构，根据权限过滤
  const filteredMenus = DEFAULT_MENU_STRUCTURE.filter(menu => {
    // 检查一级菜单权限
    const hasMainPermission = permissions.some(p => p.permission_code === menu.permission)
    if (!hasMainPermission) {
      return false
    }
    
    // 过滤子菜单
    if (menu.children && menu.children.length > 0) {
      menu.children = menu.children.filter(child => {
        return permissions.some(p => p.permission_code === child.permission)
      })
    }
    
    return true
  })
  
  console.log('构建完成的菜单:', filteredMenus)
  return filteredMenus
}

// 根据用户权限过滤菜单
function filterMenuByPermissions(menus) {
  return menus.filter(menu => {
    // 检查当前菜单权限
    if (menu.permission && !hasPermission(menu.permission)) {
      return false
    }
    
    // 递归过滤子菜单
    if (menu.children && menu.children.length > 0) {
      menu.children = filterMenuByPermissions(menu.children)
    }
    
    return true
  })
}

// 重新导出统一配置中的工具函数
export { 
  getBreadcrumb, 
  getActiveMenu, 
  flattenMenus,
  findMenuByPath,
  findMenuByPermission
} from '@/config/menu'

// 检查路径是否有权限访问
export function canAccessPath(path) {
  const menus = flattenMenus(DEFAULT_MENU_STRUCTURE)
  const menu = menus.find(m => m.path === path)
  
  if (!menu) return true // 未定义的路径默认允许访问
  
  return !menu.permission || hasPermission(menu.permission)
}

// 导出默认菜单结构供其他组件使用
export { DEFAULT_MENU_STRUCTURE }