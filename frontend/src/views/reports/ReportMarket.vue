<template>
  <ComingSoon 
    title="市场分析"
    description="市场分析报告功能正在开发中"
    icon="PieChart"
    :features="[
      '市场趋势分析',
      '行业板块分析',
      '热点股票统计',
      '资金流向分析',
      '市场情绪指标'
    ]"
  />
</template>

<script>
import ComingSoon from '../../components/ComingSoon.vue'

export default {
  name: 'ReportMarket',
  components: {
    ComingSoon
  }
}
</script>