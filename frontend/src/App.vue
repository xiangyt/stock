<template>
  <div id="app">
    <!-- 登录页面独立显示 -->
    <router-view v-if="$route.path === '/login'" />
    
    <!-- 主应用布局 -->
    <el-container v-else>
      <!-- 侧边栏 -->
      <el-aside :width="sidebarCollapsed ? '64px' : '240px'" class="sidebar">
        <div class="sidebar-header">
          <div class="logo" v-if="!sidebarCollapsed">
            <el-icon><TrendCharts /></el-icon>
            <span>智能选股系统</span>
          </div>
          <el-icon v-else class="logo-icon"><TrendCharts /></el-icon>
        </div>
        
        <el-menu
          :default-active="$route.path"
          :collapse="sidebarCollapsed"
          router
          class="sidebar-menu"
          background-color="transparent"
          text-color="rgba(255, 255, 255, 0.9)"
          active-text-color="#ffffff"
        >
          <!-- 动态生成菜单 -->
          <template v-for="item in menuItems" :key="item.index">
            <!-- 单菜单项 -->
            <el-menu-item v-if="item.type === 'item'" :index="item.index">
              <el-icon>
                <component :is="item.icon" />
              </el-icon>
              <template #title>{{ item.title }}</template>
            </el-menu-item>
            
            <!-- 子菜单项 -->
            <el-sub-menu v-else-if="item.type === 'submenu'" :index="item.index">
              <template #title>
                <el-icon>
                  <component :is="item.icon" />
                </el-icon>
                <span>{{ item.title }}</span>
              </template>
              <el-menu-item 
                v-for="child in item.children" 
                :key="child.index" 
                :index="child.index"
              >
                {{ child.title }}
              </el-menu-item>
            </el-sub-menu>
          </template>
        </el-menu>
      </el-aside>

      <el-container>
        <!-- 顶部导航 -->
        <el-header class="header">
          <div class="header-left">
            <el-button 
              type="text" 
              @click="toggleSidebar"
              class="sidebar-toggle"
            >
              <el-icon><Fold v-if="!sidebarCollapsed" /><Expand v-else /></el-icon>
            </el-button>
            <el-breadcrumb separator="/" class="breadcrumb">
              <el-breadcrumb-item v-for="item in breadcrumbs" :key="item.path" :to="item.path">
                {{ item.name }}
              </el-breadcrumb-item>
            </el-breadcrumb>
          </div>
          
          <div class="header-right">
            <el-dropdown>
              <span class="user-info">
                <el-avatar :size="32" :src="userAvatar">
                  <el-icon><User /></el-icon>
                </el-avatar>
                <span class="username">{{ username }}</span>
                <el-icon class="el-icon--right"><ArrowDown /></el-icon>
              </span>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item @click="showProfile">个人资料</el-dropdown-item>
                  <el-dropdown-item @click="logout" divided>退出登录</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </el-header>

        <!-- 主要内容区域 -->
        <el-main class="main-content">
          <router-view />
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<script>
import { ref, computed, watch, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getUser, hasPermission, isLoggedIn, logout as authLogout } from './utils/auth'
import { getUserMenus } from './utils/menu'

export default {
  name: 'App',
  setup() {
    const route = useRoute()
    const router = useRouter()
    
    const sidebarCollapsed = ref(false)
    const username = ref('管理员')
    const userAvatar = ref('')
    const menuItems = ref([])
    const userRole = ref('')

    // 初始化用户信息和菜单
    const initializeUserData = async () => {
      try {
        if (isLoggedIn()) {
          const user = getUser()
          if (user) {
            username.value = user.username || user.real_name || '管理员'
            userRole.value = user.role_code || ''
          }
          
          // 获取用户菜单
          const menus = await getUserMenus()
          menuItems.value = generateMenuItems(menus)
        } else {
          // 未登录，跳转到登录页
          router.push('/login')
        }
      } catch (error) {
        console.error('初始化权限失败:', error)
        // 如果权限初始化失败，跳转到登录页
        router.push('/login')
      }
    }

    // 监听路由变化，当从登录页跳转到其他页面时重新初始化
    watch(() => route.path, async (newPath, oldPath) => {
      if (oldPath === '/login' && newPath !== '/login' && isLoggedIn()) {
        await initializeUserData()
      }
    })

    // 初始化权限和菜单
    onMounted(initializeUserData)

    // 生成动态菜单项
    const generateMenuItems = (menus = []) => {
      const items = []

      // 如果有菜单数据，使用API返回的菜单
      if (menus && menus.length > 0) {
        menus.forEach(menu => {
          if (menu.children && menu.children.length > 0) {
            // 有子菜单
            items.push({
              type: 'submenu',
              index: menu.id.toString(),
              icon: menu.icon || 'Document',
              title: menu.name,
              children: menu.children.map(child => ({
                index: child.path,
                title: child.name
              }))
            })
          } else {
            // 单个菜单项
            items.push({
              type: 'item',
              index: menu.path,
              icon: menu.icon || 'Document',
              title: menu.name
            })
          }
        })
      } else {
        // 默认菜单（当API不可用时）
        const defaultMenus = []
        
        // 仪表板菜单 - 检查权限
        if (hasPermission('dashboard:view')) {
          defaultMenus.push({
            type: 'item',
            index: '/',
            icon: 'Odometer',
            title: '仪表板'
          })
        }
        
        // 根据权限添加菜单
        if (hasPermission('stocks:list')) {
          defaultMenus.push({
            type: 'submenu',
            index: 'stocks',
            icon: 'DataLine',
            title: '股票数据',
            children: [
              { index: '/stocks', title: '股票列表' },
              { index: '/stocks/watchlist', title: '自选股' },
              { index: '/stocks/realtime', title: '实时行情' }
            ].filter(item => {
              const permissionMap = {
                '/stocks': 'stocks:list',
                '/stocks/watchlist': 'stocks:watchlist',
                '/stocks/realtime': 'stocks:realtime'
              }
              return hasPermission(permissionMap[item.index])
            })
          })
        }
        
        if (hasPermission('strategies:list')) {
          defaultMenus.push({
            type: 'submenu',
            index: 'strategies',
            icon: 'SetUp',
            title: '策略管理',
            children: [
              { index: '/strategies', title: '策略列表' },
              { index: '/strategies/create', title: '创建策略' },
              { index: '/strategies/backtest', title: '回测中心' }
            ].filter(item => {
              const permissionMap = {
                '/strategies': 'strategies:list',
                '/strategies/create': 'strategies:create',
                '/strategies/backtest': 'strategies:backtest'
              }
              return hasPermission(permissionMap[item.index])
            })
          })
        }
        
        if (hasPermission('system:users')) {
          defaultMenus.push({
            type: 'submenu',
            index: 'system',
            icon: 'Setting',
            title: '系统管理',
            children: [
              { index: '/system/users', title: '用户管理' },
              { index: '/system/roles', title: '角色管理' },
              { index: '/system/monitoring', title: '系统监控' }
            ].filter(item => {
              const permissionMap = {
                '/system/users': 'system:users',
                '/system/roles': 'system:roles',
                '/system/monitoring': 'system:monitoring'
              }
              return hasPermission(permissionMap[item.index])
            })
          })
        }
        
        items.push(...defaultMenus)
      }

      return items
    }

    const toggleSidebar = () => {
      sidebarCollapsed.value = !sidebarCollapsed.value
    }

    // 面包屑导航
    const breadcrumbs = computed(() => {
      const pathHierarchy = {
        '/': [
          { name: '仪表板', path: '/' }
        ],
        '/stocks': [
          { name: '股票数据', path: '/stocks' },
          { name: '股票列表', path: '/stocks' }
        ],
        '/stocks/watchlist': [
          { name: '股票数据', path: '/stocks' },
          { name: '自选股', path: '/stocks/watchlist' }
        ],
        '/stocks/realtime': [
          { name: '股票数据', path: '/stocks' },
          { name: '实时行情', path: '/stocks/realtime' }
        ],
        '/strategies': [
          { name: '策略管理', path: '/strategies' },
          { name: '策略列表', path: '/strategies' }
        ],
        '/strategies/create': [
          { name: '策略管理', path: '/strategies' },
          { name: '创建策略', path: '/strategies/create' }
        ],
        '/strategies/backtest': [
          { name: '策略管理', path: '/strategies' },
          { name: '回测中心', path: '/strategies/backtest' }
        ],
        '/selections': [
          { name: '选股结果', path: '/selections' },
          { name: '选股记录', path: '/selections' }
        ],
        '/selections/today': [
          { name: '选股结果', path: '/selections' },
          { name: '今日选股', path: '/selections/today' }
        ],
        '/selections/history': [
          { name: '选股结果', path: '/selections' },
          { name: '历史回顾', path: '/selections/history' }
        ],
        '/portfolios': [
          { name: '投资组合', path: '/portfolios' },
          { name: '组合管理', path: '/portfolios' }
        ],
        '/portfolios/performance': [
          { name: '投资组合', path: '/portfolios' },
          { name: '业绩分析', path: '/portfolios/performance' }
        ],
        '/collectors': [
          { name: '数据采集', path: '/collectors' },
          { name: '采集器管理', path: '/collectors' }
        ],
        '/collectors/tasks': [
          { name: '数据采集', path: '/collectors' },
          { name: '采集任务', path: '/collectors/tasks' }
        ],
        '/collectors/sync': [
          { name: '数据采集', path: '/collectors' },
          { name: '数据同步', path: '/collectors/sync' }
        ],
        '/notifications': [
          { name: '通知管理', path: '/notifications' }
        ],
        '/notifications/robots': [
          { name: '通知管理', path: '/notifications' },
          { name: '机器人配置', path: '/notifications/robots' }
        ],
        '/notifications/templates': [
          { name: '通知管理', path: '/notifications' },
          { name: '消息模板', path: '/notifications/templates' }
        ],
        '/notifications/logs': [
          { name: '通知管理', path: '/notifications' },
          { name: '发送日志', path: '/notifications/logs' }
        ],
        '/reports': [
          { name: '报表分析', path: '/reports' }
        ],
        '/reports/strategy': [
          { name: '报表分析', path: '/reports' },
          { name: '策略报告', path: '/reports/strategy' }
        ],
        '/reports/market': [
          { name: '报表分析', path: '/reports' },
          { name: '市场分析', path: '/reports/market' }
        ],
        '/reports/risk': [
          { name: '报表分析', path: '/reports' },
          { name: '风险分析', path: '/reports/risk' }
        ],
        '/system': [
          { name: '系统管理', path: '/system' }
        ],
        '/system/users': [
          { name: '系统管理', path: '/system' },
          { name: '用户管理', path: '/system/users' }
        ],
        '/system/roles': [
          { name: '系统管理', path: '/system' },
          { name: '角色管理', path: '/system/roles' }
        ],
        '/system/monitoring': [
          { name: '系统管理', path: '/system' },
          { name: '系统监控', path: '/system/monitoring' }
        ],
        '/system/settings': [
          { name: '系统管理', path: '/system' },
          { name: '系统设置', path: '/system/settings' }
        ]
      }
      
      const currentPath = route.path
      const hierarchy = pathHierarchy[currentPath]
      
      if (hierarchy) {
        return hierarchy
      }
      
      // 如果没有找到精确匹配，尝试匹配父路径
      const segments = currentPath.split('/').filter(Boolean)
      if (segments.length > 0) {
        const parentPath = '/' + segments[0]
        const parentHierarchy = pathHierarchy[parentPath]
        if (parentHierarchy) {
          return [
            ...parentHierarchy,
            { name: '未知页面', path: currentPath }
          ]
        }
      }
      
      return [{ name: '未知页面', path: currentPath }]
    })

    const showProfile = () => {
      // TODO: 显示个人资料对话框
      console.log('显示个人资料')
    }

    const logout = async () => {
      try {
        await authLogout()
      } catch (error) {
        console.error('退出登录失败:', error)
      }
    }

    return {
      sidebarCollapsed,
      username,
      userAvatar,
      breadcrumbs,
      menuItems,
      userRole,
      toggleSidebar,
      showProfile,
      logout
    }
  }
}
</script>

<style lang="scss">
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

#app {
  font-family: 'Helvetica Neue', Helvetica, 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', '微软雅黑', Arial, sans-serif;
  height: 100vh;
}

.sidebar {
  background: linear-gradient(180deg, #1e3c72 0%, #2a5298 100%);
  box-shadow: 2px 0 8px rgba(0, 0, 0, 0.1);
  transition: all 0.3s ease;
  position: relative;
  
  &::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: linear-gradient(135deg, rgba(255, 255, 255, 0.1) 0%, rgba(255, 255, 255, 0.05) 100%);
    pointer-events: none;
  }
  
  .sidebar-header {
    height: 70px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
    background: transparent;
    position: relative;
    
    .logo {
      display: flex;
      align-items: center;
      color: #ffffff;
      font-size: 20px;
      font-weight: 600;
      text-shadow: 0 2px 8px rgba(0, 0, 0, 0.5);
      letter-spacing: 0.5px;
      transition: all 0.3s ease;
      
      &:hover {
        transform: scale(1.02);
        text-shadow: 0 3px 12px rgba(0, 0, 0, 0.6);
      }
      
      .el-icon {
        margin-right: 12px;
        font-size: 28px;
        color: #ffffff;
        filter: drop-shadow(0 2px 6px rgba(0, 0, 0, 0.4));
        animation: iconGlow 4s ease-in-out infinite alternate;
        transition: all 0.3s ease;
        background: none !important;
        border: none !important;
        box-shadow: none !important;
      }
      
      span {
        font-family: 'PingFang SC', 'Microsoft YaHei', sans-serif;
        color: #ffffff;
      }
    }
    
    .logo-icon {
      font-size: 28px;
      color: #64b5f6;
      filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.3));
    }
  }
  
  .sidebar-menu {
    border: none;
    background: transparent !important;
    padding: 8px 0;
    
    .el-menu-item,
    .el-sub-menu__title {
      height: 52px;
      line-height: 52px;
      margin: 2px 8px;
      border-radius: 12px;
      transition: all 0.3s ease;
      position: relative;
      overflow: hidden;
      
      &::before {
        content: '';
        position: absolute;
        top: 0;
        left: -100%;
        width: 100%;
        height: 100%;
        background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.1), transparent);
        transition: left 0.5s ease;
      }
      
      &:hover {
        background: rgba(255, 255, 255, 0.1) !important;
        transform: translateX(4px);
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
        
        &::before {
          left: 100%;
        }
      }
      
      .el-icon {
        margin-right: 12px;
        font-size: 18px;
        color: rgba(255, 255, 255, 0.8);
        transition: all 0.3s ease;
      }
      
      span {
        color: rgba(255, 255, 255, 0.9);
        font-weight: 500;
        transition: all 0.3s ease;
      }
    }
    
    .el-menu-item.is-active {
      background: linear-gradient(135deg, #64b5f6 0%, #42a5f5 100%) !important;
      color: white !important;
      box-shadow: 0 4px 15px rgba(100, 181, 246, 0.4);
      transform: translateX(4px);
      
      .el-icon {
        color: white !important;
        transform: scale(1.1);
      }
      
      span {
        color: white !important;
        font-weight: 600;
      }
    }
    
    .el-sub-menu {
      .el-sub-menu__title {
        &:hover {
          .el-icon {
            color: #64b5f6 !important;
            transform: scale(1.05);
          }
        }
      }
      
      .el-menu {
        background: rgba(0, 0, 0, 0.1) !important;
        
        .el-menu-item {
          margin: 2px 16px 2px 24px;
          height: 44px;
          line-height: 44px;
          font-size: 14px;
          border-radius: 10px;
          position: relative;
          overflow: hidden;
          transition: all 0.3s ease;
          
          &::before {
            content: '';
            position: absolute;
            top: 0;
            left: -100%;
            width: 100%;
            height: 100%;
            background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.15), transparent);
            transition: left 0.5s ease;
          }
          
          span {
            position: relative;
            z-index: 1;
          }
          
          &:hover {
            background: rgba(255, 255, 255, 0.1) !important;
            transform: translateX(3px);
            box-shadow: 0 3px 10px rgba(0, 0, 0, 0.1);
            
            &::before {
              left: 100%;
            }
          }
          
          &.is-active {
            background: linear-gradient(135deg, #64b5f6 0%, #42a5f5 100%) !important;
            color: white !important;
            font-weight: 600;
            transform: translateX(3px);
            box-shadow: 0 4px 12px rgba(100, 181, 246, 0.3);
            
            span {
              color: white !important;
            }
          }
        }
      }
    }
    
    // 折叠状态下的样式
    &.el-menu--collapse {
      .el-menu-item,
      .el-sub-menu__title {
        margin: 2px 4px;
        justify-content: center;
        
        .el-icon {
          margin-right: 0;
          font-size: 20px;
        }
      }
    }
  }
}

.header {
  background: linear-gradient(135deg, #ffffff 0%, #f8fafc 100%);
  border-bottom: 1px solid #e2e8f0;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.08);
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 24px;
  height: 70px !important;
  backdrop-filter: blur(10px);
  
  .header-left {
    display: flex;
    align-items: center;
    
    .sidebar-toggle {
      margin-right: 24px;
      color: #64748b;
      padding: 8px;
      border-radius: 8px;
      transition: all 0.3s ease;
      
      &:hover {
        color: #3b82f6;
        background-color: rgba(59, 130, 246, 0.1);
        transform: scale(1.05);
      }
      
      .el-icon {
        font-size: 18px;
      }
    }
    
    .breadcrumb {
      font-size: 14px;
      font-weight: 500;
      
      .el-breadcrumb__item {
        .el-breadcrumb__inner {
          color: #475569;
          transition: color 0.3s ease;
          
          &:hover {
            color: #3b82f6;
          }
        }
        
        &:last-child .el-breadcrumb__inner {
          color: #1e293b;
          font-weight: 600;
        }
      }
    }
  }
  
  .header-right {
    .user-info {
      display: flex;
      align-items: center;
      cursor: pointer;
      padding: 10px 16px;
      border-radius: 12px;
      transition: all 0.3s ease;
      background: rgba(255, 255, 255, 0.7);
      border: 1px solid rgba(226, 232, 240, 0.8);
      
      &:hover {
        background: rgba(59, 130, 246, 0.05);
        border-color: rgba(59, 130, 246, 0.2);
        transform: translateY(-1px);
        box-shadow: 0 4px 12px rgba(59, 130, 246, 0.15);
      }
      
      .el-avatar {
        border: 2px solid rgba(59, 130, 246, 0.2);
        transition: all 0.3s ease;
      }
      
      &:hover .el-avatar {
        border-color: #3b82f6;
        transform: scale(1.05);
      }
      
      .username {
        margin: 0 12px;
        font-size: 14px;
        font-weight: 600;
        color: #1e293b;
        transition: color 0.3s ease;
      }
      
      .el-icon {
        color: #64748b;
        transition: all 0.3s ease;
      }
      
      &:hover .el-icon {
        color: #3b82f6;
        transform: rotate(180deg);
      }
    }
  }
}

.main-content {
  background: linear-gradient(135deg, #f1f5f9 0%, #e2e8f0 100%);
  padding: 24px;
  overflow-y: auto;
  min-height: calc(100vh - 70px);
  position: relative;
  
  &::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-image: 
      radial-gradient(circle at 20% 80%, rgba(120, 119, 198, 0.03) 0%, transparent 50%),
      radial-gradient(circle at 80% 20%, rgba(255, 119, 198, 0.03) 0%, transparent 50%),
      radial-gradient(circle at 40% 40%, rgba(120, 219, 255, 0.03) 0%, transparent 50%);
    pointer-events: none;
  }
}

// Element Plus 样式覆盖
.el-card {
  border-radius: 16px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  border: 1px solid rgba(226, 232, 240, 0.8);
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(10px);
  transition: all 0.3s ease;
  
  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 8px 30px rgba(0, 0, 0, 0.12);
  }
  
  .el-card__header {
    background: linear-gradient(135deg, rgba(59, 130, 246, 0.05) 0%, rgba(147, 51, 234, 0.05) 100%);
    border-bottom: 1px solid rgba(226, 232, 240, 0.8);
    padding: 20px 24px;
    
    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      font-weight: 600;
      color: #1e293b;
    }
  }
  
  .el-card__body {
    padding: 24px;
  }
}

.el-button {
  border-radius: 10px;
  font-weight: 500;
  transition: all 0.3s ease;
  
  &.el-button--primary {
    background: linear-gradient(135deg, #3b82f6 0%, #1d4ed8 100%);
    border: none;
    
    &:hover {
      transform: translateY(-1px);
      box-shadow: 0 6px 20px rgba(59, 130, 246, 0.4);
    }
  }
  
  &.el-button--success {
    background: linear-gradient(135deg, #10b981 0%, #059669 100%);
    border: none;
    
    &:hover {
      transform: translateY(-1px);
      box-shadow: 0 6px 20px rgba(16, 185, 129, 0.4);
    }
  }
  
  &.el-button--warning {
    background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
    border: none;
    
    &:hover {
      transform: translateY(-1px);
      box-shadow: 0 6px 20px rgba(245, 158, 11, 0.4);
    }
  }
  
  &.el-button--danger {
    background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
    border: none;
    
    &:hover {
      transform: translateY(-1px);
      box-shadow: 0 6px 20px rgba(239, 68, 68, 0.4);
    }
  }
}

.el-table {
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  border: 1px solid rgba(226, 232, 240, 0.8);
  
  .el-table__header {
    background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
    
    th {
      background: transparent !important;
      color: #475569;
      font-weight: 600;
      border-bottom: 2px solid #e2e8f0;
    }
  }
  
  .el-table__row {
    transition: all 0.3s ease;
    
    &:hover {
      background-color: rgba(59, 130, 246, 0.05) !important;
    }
  }
}

.el-pagination {
  margin-top: 24px;
  text-align: right;
  
  .el-pager li {
    border-radius: 8px;
    margin: 0 2px;
    transition: all 0.3s ease;
    
    &.is-active {
      background: linear-gradient(135deg, #3b82f6 0%, #1d4ed8 100%);
      color: white;
    }
    
    &:hover:not(.is-active) {
      background-color: rgba(59, 130, 246, 0.1);
      color: #3b82f6;
    }
  }
  
  .btn-prev,
  .btn-next {
    border-radius: 8px;
    transition: all 0.3s ease;
    
    &:hover {
      background-color: rgba(59, 130, 246, 0.1);
      color: #3b82f6;
    }
  }
}

// 动画效果
@keyframes iconGlow {
  0% {
    filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.3));
  }
  100% {
    filter: drop-shadow(0 2px 8px rgba(0, 0, 0, 0.5));
  }
}

// 响应式设计
@media (max-width: 768px) {
  .sidebar {
    position: fixed;
    z-index: 1000;
    height: 100vh;
  }
  
  .header-left .breadcrumb {
    display: none;
  }
  
  .sidebar .sidebar-header .logo {
    padding: 8px 12px;
    font-size: 18px;
    
    .el-icon {
      font-size: 24px;
    }
  }
}
</style>