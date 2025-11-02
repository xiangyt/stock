<template>
  <ComingSoon 
    title="采集任务"
    description="数据采集任务管理功能正在开发中"
    icon="Clock"
    :features="[
      '任务创建和调度',
      '任务执行状态',
      '任务日志查看',
      '任务失败重试',
      '任务性能监控'
    ]"
  />
</template>

<script>
import ComingSoon from '../../components/ComingSoon.vue'

export default {
  name: 'CollectorTasks',
  components: {
    ComingSoon
  }
}
</script>