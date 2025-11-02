<template>
  <ComingSoon 
    title="系统监控"
    description="系统监控功能正在开发中"
    icon="Monitor"
    :features="[
      '系统性能监控',
      'API调用统计',
      '错误日志分析',
      '资源使用情况',
      '告警通知设置'
    ]"
  />
</template>

<script>
import ComingSoon from '../../components/ComingSoon.vue'

export default {
  name: 'SystemMonitoring',
  components: {
    ComingSoon
  }
}
</script>