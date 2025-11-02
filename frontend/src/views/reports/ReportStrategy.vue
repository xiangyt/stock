<template>
  <ComingSoon 
    title="策略报告"
    description="策略分析报告功能正在开发中"
    icon="TrendCharts"
    :features="[
      '策略收益分析',
      '风险评估报告',
      '策略对比分析',
      '收益归因分析',
      '报告导出功能'
    ]"
  />
</template>

<script>
import ComingSoon from '../../components/ComingSoon.vue'

export default {
  name: 'ReportStrategy',
  components: {
    ComingSoon
  }
}
</script>