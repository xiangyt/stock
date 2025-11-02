<template>
  <ComingSoon 
    title="发送日志"
    description="通知发送日志功能正在开发中"
    icon="List"
    :features="[
      '发送记录查询',
      '发送状态统计',
      '失败原因分析',
      '发送频率控制',
      '日志导出功能'
    ]"
  />
</template>

<script>
import ComingSoon from '../../components/ComingSoon.vue'

export default {
  name: 'NotificationLogs',
  components: {
    ComingSoon
  }
}
</script>