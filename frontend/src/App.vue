<template>
  <div id="app">
    <el-container>
      <!-- 顶部导航 -->
      <el-header class="header">
        <div class="header-content">
          <div class="logo">
            <el-icon><TrendCharts /></el-icon>
            <span>智能选股系统</span>
          </div>
          <el-menu
            :default-active="$route.path"
            mode="horizontal"
            router
            class="nav-menu"
          >
            <el-menu-item index="/">
              <el-icon><House /></el-icon>
              <span>首页</span>
            </el-menu-item>
            <el-menu-item index="/stocks">
              <el-icon><DataLine /></el-icon>
              <span>股票列表</span>
            </el-menu-item>
            <el-menu-item index="/analysis">
              <el-icon><PieChart /></el-icon>
              <span>数据分析</span>
            </el-menu-item>
          </el-menu>
        </div>
      </el-header>

      <!-- 主要内容区域 -->
      <el-main class="main-content">
        <router-view />
      </el-main>

      <!-- 底部 -->
      <el-footer class="footer">
        <div class="footer-content">
          <p>&copy; 2025 智能选股系统 - 基于Go + Vue3构建</p>
        </div>
      </el-footer>
    </el-container>
  </div>
</template>

<script>
export default {
  name: 'App'
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

.header {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 0;
  height: 60px !important;
  line-height: 60px;

  .header-content {
    display: flex;
    justify-content: space-between;
    align-items: center;
    height: 100%;
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 20px;
  }

  .logo {
    display: flex;
    align-items: center;
    font-size: 20px;
    font-weight: bold;
    color: white;

    .el-icon {
      margin-right: 8px;
      font-size: 24px;
    }
  }

  .nav-menu {
    background: transparent;
    border: none;

    .el-menu-item {
      color: rgba(255, 255, 255, 0.8);
      border-bottom: 2px solid transparent;

      &:hover,
      &.is-active {
        color: white;
        background: rgba(255, 255, 255, 0.1);
        border-bottom-color: white;
      }

      .el-icon {
        margin-right: 5px;
      }
    }
  }
}

.main-content {
  background: #f5f7fa;
  min-height: calc(100vh - 120px);
  padding: 20px;
}

.footer {
  background: #2c3e50;
  color: white;
  height: 60px !important;
  line-height: 60px;
  text-align: center;

  .footer-content {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 20px;
  }
}

// Element Plus 样式覆盖
.el-card {
  border-radius: 8px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

.el-button {
  border-radius: 6px;
}

.el-table {
  border-radius: 8px;
  overflow: hidden;
}
</style>