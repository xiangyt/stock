<template>
  <div class="coming-soon">
    <el-card class="coming-soon-card">
      <div class="coming-soon-content">
        <el-icon class="coming-soon-icon" :size="80">
          <component :is="icon" />
        </el-icon>
        <h2 class="coming-soon-title">{{ title }}</h2>
        <p class="coming-soon-description">{{ description }}</p>
        <div class="coming-soon-features" v-if="features && features.length > 0">
          <h3>即将推出的功能：</h3>
          <ul>
            <li v-for="feature in features" :key="feature">{{ feature }}</li>
          </ul>
        </div>
        <el-button type="primary" @click="goBack">
          <el-icon><ArrowLeft /></el-icon>
          返回上一页
        </el-button>
      </div>
    </el-card>
  </div>
</template>

<script>
import { useRouter } from 'vue-router'

export default {
  name: 'ComingSoon',
  props: {
    title: {
      type: String,
      default: '功能开发中'
    },
    description: {
      type: String,
      default: '该功能正在紧张开发中，敬请期待！'
    },
    icon: {
      type: String,
      default: 'Tools'
    },
    features: {
      type: Array,
      default: () => []
    }
  },
  setup() {
    const router = useRouter()

    const goBack = () => {
      router.go(-1)
    }

    return {
      goBack
    }
  }
}
</script>

<style lang="scss" scoped>
.coming-soon {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 60vh;

  .coming-soon-card {
    max-width: 500px;
    width: 100%;

    .coming-soon-content {
      text-align: center;
      padding: 40px 20px;

      .coming-soon-icon {
        color: #409eff;
        margin-bottom: 20px;
      }

      .coming-soon-title {
        color: #303133;
        margin-bottom: 16px;
        font-size: 24px;
      }

      .coming-soon-description {
        color: #606266;
        margin-bottom: 30px;
        font-size: 16px;
        line-height: 1.6;
      }

      .coming-soon-features {
        text-align: left;
        margin-bottom: 30px;
        padding: 20px;
        background-color: #f8f9fa;
        border-radius: 8px;

        h3 {
          color: #303133;
          margin-bottom: 12px;
          font-size: 16px;
        }

        ul {
          margin: 0;
          padding-left: 20px;

          li {
            color: #606266;
            margin-bottom: 8px;
            line-height: 1.5;
          }
        }
      }
    }
  }
}
</style>