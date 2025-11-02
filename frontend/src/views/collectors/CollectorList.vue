<template>
  <ComingSoon 
    title="采集器管理"
    description="数据采集器管理功能正在开发中"
    icon="Download"
    :features="[
      '采集器状态监控',
      '数据源配置管理',
      '采集频率设置',
      '数据质量检查',
      '采集器性能统计'
    ]"
  />
</template>

<script>
import ComingSoon from '../../components/ComingSoon.vue'

export default {
  name: 'CollectorList',
  components: {
    ComingSoon
  }
}
</script>