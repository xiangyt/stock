<template>
  <div class="forbidden-container">
    <div class="forbidden-content">
      <div class="error-icon">
        <el-icon size="120" color="#f56c6c">
          <WarningFilled />
        </el-icon>
      </div>
      
      <h1 class="error-title">403</h1>
      <h2 class="error-subtitle">无权限访问</h2>
      <p class="error-description">
        抱歉，您没有权限访问此页面。请联系管理员获取相应权限。
      </p>
      
      <div class="error-actions">
        <el-button type="primary" @click="goHome">
          <el-icon><House /></el-icon>
          返回首页
        </el-button>
        <el-button @click="goBack">
          <el-icon><ArrowLeft /></el-icon>
          返回上页
        </el-button>
      </div>
      
      <div class="user-info" v-if="userRole">
        <p>当前角色：{{ userRole }}</p>
        <p>如需更多权限，请联系系统管理员</p>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { WarningFilled, House, ArrowLeft } from '@element-plus/icons-vue'
import permissionManager from '../utils/permission'

export default {
  name: 'Forbidden',
  components: {
    WarningFilled,
    House,
    ArrowLeft
  },
  setup() {
    const router = useRouter()
    const userRole = ref('')
    
    onMounted(() => {
      try {
        const userInfo = permissionManager.getUserRole()
        userRole.value = userInfo.role || '未知角色'
      } catch (error) {
        console.error('获取用户角色失败:', error)
      }
    })
    
    const goHome = () => {
      router.push('/')
    }
    
    const goBack = () => {
      router.go(-1)
    }
    
    return {
      userRole,
      goHome,
      goBack
    }
  }
}
</script>

<style lang="scss" scoped>
.forbidden-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
  padding: 20px;
}

.forbidden-content {
  text-align: center;
  max-width: 500px;
  padding: 40px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.1);
}

.error-icon {
  margin-bottom: 20px;
}

.error-title {
  font-size: 72px;
  font-weight: bold;
  color: #f56c6c;
  margin: 0 0 10px 0;
  line-height: 1;
}

.error-subtitle {
  font-size: 28px;
  color: #2c3e50;
  margin: 0 0 20px 0;
  font-weight: 600;
}

.error-description {
  font-size: 16px;
  color: #7f8c8d;
  line-height: 1.6;
  margin-bottom: 30px;
}

.error-actions {
  display: flex;
  gap: 15px;
  justify-content: center;
  margin-bottom: 30px;
  
  .el-button {
    padding: 12px 24px;
    font-size: 14px;
  }
}

.user-info {
  padding-top: 20px;
  border-top: 1px solid #ebeef5;
  
  p {
    margin: 8px 0;
    font-size: 14px;
    color: #909399;
    
    &:first-child {
      font-weight: 500;
      color: #606266;
    }
  }
}

// 响应式设计
@media (max-width: 600px) {
  .forbidden-content {
    padding: 30px 20px;
  }
  
  .error-title {
    font-size: 60px;
  }
  
  .error-subtitle {
    font-size: 24px;
  }
  
  .error-actions {
    flex-direction: column;
    
    .el-button {
      width: 100%;
    }
  }
}
</style>