// 统一的菜单配置 - 所有模块共享
// 这是系统的唯一菜单数据源，所有菜单相关功能都基于此配置

// 菜单图标映射
export const MENU_ICONS = {
  // 系统管理
  'system:module': 'Setting',
  'system:users': 'User',
  'system:roles': 'UserFilled',
  'system:monitoring': 'Monitor',
  'system:settings': 'Tools',
  
  // 股票数据
  'stocks:module': 'TrendCharts',
  'stocks:list': 'List',
  'stocks:watchlist': 'Star',
  'stocks:realtime': 'DataLine',
  
  // 策略管理
  'strategies:module': 'Operation',
  'strategies:list': 'Document',
  'strategies:create': 'Plus',
  'strategies:backtest': 'DataAnalysis',
  
  // 选股结果
  'selections:module': 'Search',
  'selections:list': 'Files',
  'selections:today': 'Calendar',
  'selections:history': 'Clock',
  
  // 投资组合
  'portfolios:module': 'Briefcase',
  'portfolios:manage': 'Management',
  'portfolios:performance': 'TrendCharts',
  
  // 数据采集
  'collectors:module': 'Download',
  'collectors:manage': 'Setting',
  'collectors:tasks': 'List',
  'collectors:sync': 'Refresh',
  
  // 通知管理
  'notifications:module': 'Bell',
  'notifications:robots': 'Robot',
  'notifications:templates': 'Document',
  'notifications:logs': 'ChatDotRound',
  
  // 报表分析
  'reports:module': 'DataAnalysis',
  'reports:strategy': 'PieChart',
  'reports:market': 'TrendCharts',
  'reports:risk': 'Warning',
  
  // 仪表板
  'dashboard:view': 'Odometer'
}

// 统一的菜单结构配置
// 所有菜单、权限、面包屑等功能都基于此配置
export const MENU_STRUCTURE = [
  {
    id: 'dashboard',
    name: '仪表板',
    path: '/',
    icon: 'Odometer',
    permission: 'dashboard:view',
    children: []
  },
  {
    id: 'stocks',
    name: '股票数据',
    path: '/stocks',
    icon: 'TrendCharts',
    permission: 'stocks:module',
    children: [
      {
        id: 'stocks-list',
        name: '股票列表',
        path: '/stocks',
        icon: 'List',
        permission: 'stocks:list'
      },
      {
        id: 'stocks-watchlist',
        name: '自选股',
        path: '/stocks/watchlist',
        icon: 'Star',
        permission: 'stocks:watchlist'
      },
      {
        id: 'stocks-realtime',
        name: '实时行情',
        path: '/stocks/realtime',
        icon: 'DataLine',
        permission: 'stocks:realtime'
      }
    ]
  },
  {
    id: 'strategies',
    name: '策略管理',
    path: '/strategies',
    icon: 'Operation',
    permission: 'strategies:module',
    children: [
      {
        id: 'strategies-list',
        name: '策略列表',
        path: '/strategies',
        icon: 'Document',
        permission: 'strategies:list'
      },
      {
        id: 'strategies-create',
        name: '创建策略',
        path: '/strategies/create',
        icon: 'Plus',
        permission: 'strategies:create'
      },
      {
        id: 'strategies-backtest',
        name: '回测中心',
        path: '/strategies/backtest',
        icon: 'DataAnalysis',
        permission: 'strategies:backtest'
      }
    ]
  },
  {
    id: 'selections',
    name: '选股结果',
    path: '/selections',
    icon: 'Search',
    permission: 'selections:module',
    children: [
      {
        id: 'selections-list',
        name: '选股记录',
        path: '/selections',
        icon: 'Files',
        permission: 'selections:list'
      },
      {
        id: 'selections-today',
        name: '今日选股',
        path: '/selections/today',
        icon: 'Calendar',
        permission: 'selections:today'
      },
      {
        id: 'selections-history',
        name: '历史回顾',
        path: '/selections/history',
        icon: 'Clock',
        permission: 'selections:history'
      }
    ]
  },
  {
    id: 'portfolios',
    name: '投资组合',
    path: '/portfolios',
    icon: 'Briefcase',
    permission: 'portfolios:module',
    children: [
      {
        id: 'portfolios-manage',
        name: '组合管理',
        path: '/portfolios',
        icon: 'Management',
        permission: 'portfolios:manage'
      },
      {
        id: 'portfolios-performance',
        name: '业绩分析',
        path: '/portfolios/performance',
        icon: 'TrendCharts',
        permission: 'portfolios:performance'
      }
    ]
  },
  {
    id: 'collectors',
    name: '数据采集',
    path: '/collectors',
    icon: 'Download',
    permission: 'collectors:module',
    children: [
      {
        id: 'collectors-manage',
        name: '采集器管理',
        path: '/collectors',
        icon: 'Setting',
        permission: 'collectors:manage'
      },
      {
        id: 'collectors-tasks',
        name: '采集任务',
        path: '/collectors/tasks',
        icon: 'List',
        permission: 'collectors:tasks'
      },
      {
        id: 'collectors-sync',
        name: '数据同步',
        path: '/collectors/sync',
        icon: 'Refresh',
        permission: 'collectors:sync'
      }
    ]
  },
  {
    id: 'notifications',
    name: '通知管理',
    path: '/notifications',
    icon: 'Bell',
    permission: 'notifications:module',
    children: [
      {
        id: 'notifications-robots',
        name: '机器人配置',
        path: '/notifications/robots',
        icon: 'Robot',
        permission: 'notifications:robots'
      },
      {
        id: 'notifications-templates',
        name: '消息模板',
        path: '/notifications/templates',
        icon: 'Document',
        permission: 'notifications:templates'
      },
      {
        id: 'notifications-logs',
        name: '发送日志',
        path: '/notifications/logs',
        icon: 'ChatDotRound',
        permission: 'notifications:logs'
      }
    ]
  },
  {
    id: 'reports',
    name: '报表分析',
    path: '/reports',
    icon: 'DataAnalysis',
    permission: 'reports:module',
    children: [
      {
        id: 'reports-strategy',
        name: '策略报告',
        path: '/reports/strategy',
        icon: 'PieChart',
        permission: 'reports:strategy'
      },
      {
        id: 'reports-market',
        name: '市场分析',
        path: '/reports/market',
        icon: 'TrendCharts',
        permission: 'reports:market'
      },
      {
        id: 'reports-risk',
        name: '风险分析',
        path: '/reports/risk',
        icon: 'Warning',
        permission: 'reports:risk'
      }
    ]
  },
  {
    id: 'system',
    name: '系统管理',
    path: '/system',
    icon: 'Setting',
    permission: 'system:module',
    children: [
      {
        id: 'system-users',
        name: '用户管理',
        path: '/system/users',
        icon: 'User',
        permission: 'system:users'
      },
      {
        id: 'system-roles',
        name: '角色管理',
        path: '/system/roles',
        icon: 'UserFilled',
        permission: 'system:roles'
      },
      {
        id: 'system-monitoring',
        name: '系统监控',
        path: '/system/monitoring',
        icon: 'Monitor',
        permission: 'system:monitoring'
      }
    ]
  }
]

// 工具函数：扁平化菜单（用于搜索等功能）
export function flattenMenus(menus = MENU_STRUCTURE) {
  const flattened = []
  
  function flatten(items) {
    items.forEach(item => {
      flattened.push(item)
      if (item.children && item.children.length > 0) {
        flatten(item.children)
      }
    })
  }
  
  flatten(menus)
  return flattened
}

// 工具函数：根据路径查找菜单项
export function findMenuByPath(path, menus = MENU_STRUCTURE) {
  const flatMenus = flattenMenus(menus)
  return flatMenus.find(menu => menu.path === path)
}

// 工具函数：根据权限编码查找菜单项
export function findMenuByPermission(permission, menus = MENU_STRUCTURE) {
  const flatMenus = flattenMenus(menus)
  return flatMenus.find(menu => menu.permission === permission)
}

// 工具函数：获取面包屑导航
export function getBreadcrumb(path, menus = MENU_STRUCTURE) {
  const breadcrumb = []
  
  function findPath(items, targetPath, currentPath = []) {
    for (const item of items) {
      const newPath = [...currentPath, item]
      
      if (item.path === targetPath) {
        breadcrumb.push(...newPath)
        return true
      }
      
      if (item.children && item.children.length > 0) {
        if (findPath(item.children, targetPath, newPath)) {
          return true
        }
      }
    }
    return false
  }
  
  findPath(menus, path)
  return breadcrumb
}

// 工具函数：获取当前激活的菜单项
export function getActiveMenu(path, menus = MENU_STRUCTURE) {
  function findActiveMenu(items) {
    for (const item of items) {
      if (item.path === path) {
        return item
      }
      
      if (item.children && item.children.length > 0) {
        const found = findActiveMenu(item.children)
        if (found) return found
      }
    }
    return null
  }
  
  return findActiveMenu(menus)
}