<template>
  <ComingSoon 
    title="系统设置"
    description="系统设置功能正在开发中"
    icon="Setting"
    :features="[
      '系统参数配置',
      '数据源设置',
      '通知配置管理',
      '系统备份恢复',
      '版本更新管理'
    ]"
  />
</template>

<script>
import ComingSoon from '../../components/ComingSoon.vue'

export default {
  name: 'SystemSettings',
  components: {
    ComingSoon
  }
}
</script>