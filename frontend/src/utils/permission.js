import axios from 'axios'
import { MENU_STRUCTURE } from '@/config/menu'

// 权限管理工具类
class PermissionManager {
  constructor() {
    this.permissions = []
    this.userRole = ''
    this.userRoleCode = ''
    this.isInitialized = false
  }

  // 初始化权限管理器
  async init() {
    try {
      const response = await axios.get('/api/v1/permissions/user/menu')
      if (response.data.code === 0) {
        this.permissions = response.data.data.permissions || []
        this.userRole = response.data.data.user_role || ''
        this.userRoleCode = response.data.data.user_role_code || ''
        this.isInitialized = true
        console.log('权限初始化成功:', this.permissions.length, '个权限')
      } else {
        console.error('权限初始化失败:', response.data.message)
      }
    } catch (error) {
      console.error('权限初始化错误:', error)
      // 如果初始化失败，使用默认权限（所有菜单可见）
      this.isInitialized = true
    }
  }

  // 检查用户是否有某个权限
  hasPermission(permissionCode) {
    if (!this.isInitialized) {
      console.warn('权限管理器未初始化，默认允许访问')
      return true
    }

    // 超级管理员拥有所有权限
    if (this.userRoleCode === 'super_admin') {
      return true
    }

    // 检查权限列表
    return this.permissions.some(permission => 
      permission.permission_code === permissionCode && permission.status === 1
    )
  }

  // 检查用户是否有任意一个权限
  hasAnyPermission(permissionCodes) {
    if (!this.isInitialized) {
      return true
    }

    if (this.userRoleCode === 'super_admin') {
      return true
    }

    return permissionCodes.some(code => this.hasPermission(code))
  }

  // 获取用户角色信息
  getUserRole() {
    return {
      role: this.userRole,
      roleCode: this.userRoleCode
    }
  }

  // 获取用户的菜单权限
  getMenuPermissions() {
    return this.permissions.filter(permission => 
      permission.resource_type === 'menu' || permission.resource_type === 'module'
    )
  }

  // 根据权限生成菜单结构
  generateMenuStructure() {
    // 使用统一的菜单结构配置
    const filteredMenus = []
    
    MENU_STRUCTURE.forEach(menuItem => {
      // 检查主菜单权限
      if (this.hasPermission(menuItem.permission)) {
        // 过滤子菜单
        const filteredChildren = menuItem.children ? 
          menuItem.children.filter(child => this.hasPermission(child.permission)) : []
        
        // 如果有子菜单或者主菜单本身有独立路径，则保留
        if (filteredChildren.length > 0 || menuItem.path) {
          filteredMenus.push({
            ...menuItem,
            children: filteredChildren
          })
        }
      }
    })

    return filteredMenus
  }

  // 重置权限管理器
  reset() {
    this.permissions = []
    this.userRole = ''
    this.userRoleCode = ''
    this.isInitialized = false
  }
}

// 创建全局权限管理器实例
const permissionManager = new PermissionManager()

export default permissionManager
=======
  // 重置权限管理器
  reset() {
    this.permissions = []
    this.userRole = ''
    this.userRoleCode = ''
    this.isInitialized = false
  }
}

// 创建全局权限管理器实例
const permissionManager = new PermissionManager()

export default permissionManager
      dashboard: {
        name: '仪表板',
        path: '/',
        icon: 'Odometer',
        permission: 'dashboard:view',
        children: []
      },
      stocks: {
        name: '股票数据',
        path: '/stocks',
        icon: 'DataLine',
        permission: 'stocks:module',
        children: [
          { name: '股票列表', path: '/stocks', permission: 'stocks:list' },
          { name: '自选股', path: '/stocks/watchlist', permission: 'stocks:watchlist' },
          { name: '实时行情', path: '/stocks/realtime', permission: 'stocks:realtime' }
        ]
      },
      strategies: {
        name: '策略管理',
        path: '/strategies',
        icon: 'SetUp',
        permission: 'strategies:module',
        children: [
          { name: '策略列表', path: '/strategies', permission: 'strategies:list' },
          { name: '创建策略', path: '/strategies/create', permission: 'strategies:create' },
          { name: '回测中心', path: '/strategies/backtest', permission: 'strategies:backtest' }
        ]
      },
      selections: {
        name: '选股结果',
        path: '/selections',
        icon: 'Select',
        permission: 'selections:module',
        children: [
          { name: '选股记录', path: '/selections', permission: 'selections:list' },
          { name: '今日选股', path: '/selections/today', permission: 'selections:today' },
          { name: '历史回顾', path: '/selections/history', permission: 'selections:history' }
        ]
      },
      portfolios: {
        name: '投资组合',
        path: '/portfolios',
        icon: 'PieChart',
        permission: 'portfolios:module',
        children: [
          { name: '组合管理', path: '/portfolios', permission: 'portfolios:manage' },
          { name: '业绩分析', path: '/portfolios/performance', permission: 'portfolios:performance' }
        ]
      },
      collectors: {
        name: '数据采集',
        path: '/collectors',
        icon: 'Download',
        permission: 'collectors:module',
        children: [
          { name: '采集器管理', path: '/collectors', permission: 'collectors:manage' },
          { name: '采集任务', path: '/collectors/tasks', permission: 'collectors:tasks' },
          { name: '数据同步', path: '/collectors/sync', permission: 'collectors:sync' }
        ]
      },
      notifications: {
        name: '通知管理',
        path: '/notifications/robots',
        icon: 'Bell',
        permission: 'notifications:module',
        children: [
          { name: '机器人配置', path: '/notifications/robots', permission: 'notifications:robots' },
          { name: '消息模板', path: '/notifications/templates', permission: 'notifications:templates' },
          { name: '发送日志', path: '/notifications/logs', permission: 'notifications:logs' }
        ]
      },
      reports: {
        name: '报表分析',
        path: '/reports/strategy',
        icon: 'Document',
        permission: 'reports:module',
        children: [
          { name: '策略报告', path: '/reports/strategy', permission: 'reports:strategy' },
          { name: '市场分析', path: '/reports/market', permission: 'reports:market' },
          { name: '风险分析', path: '/reports/risk', permission: 'reports:risk' }
        ]
      },
      system: {
        name: '系统管理',
        path: '/system/users',
        icon: 'Setting',
        permission: 'system:module',
        children: [
          { name: '用户管理', path: '/system/users', permission: 'system:users' },
          { name: '角色管理', path: '/system/roles', permission: 'system:roles' },
          { name: '系统监控', path: '/system/monitoring', permission: 'system:monitoring' },
          { name: '系统设置', path: '/system/settings', permission: 'system:settings' }
        ]
      }
    }

    // 过滤没有权限的菜单项
    const filteredMenu = {}
    
    Object.keys(menuTree).forEach(key => {
      const menuItem = menuTree[key]
      
      // 检查主菜单权限
      if (this.hasPermission(menuItem.permission)) {
        // 过滤子菜单
        const filteredChildren = menuItem.children.filter(child => 
          this.hasPermission(child.permission)
        )
        
        // 如果有子菜单或者主菜单本身有独立路径，则保留
        if (filteredChildren.length > 0 || menuItem.path) {
          filteredMenu[key] = {
            ...menuItem,
            children: filteredChildren
          }
        }
      }
    })

    return filteredMenu
  }

  // 重置权限管理器
  reset() {
    this.permissions = []
    this.userRole = ''
    this.userRoleCode = ''
    this.isInitialized = false
  }
}

// 创建全局权限管理器实例
const permissionManager = new PermissionManager()

export default permissionManager