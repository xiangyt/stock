<template>
  <ComingSoon 
    title="业绩分析"
    description="投资组合业绩分析功能正在开发中"
    icon="TrendCharts"
    :features="[
      '收益率曲线图',
      '风险收益分析',
      '基准对比分析',
      '归因分析报告',
      '业绩评价指标'
    ]"
  />
</template>

<script>
import ComingSoon from '../../components/ComingSoon.vue'

export default {
  name: 'PortfolioPerformance',
  components: {
    ComingSoon
  }
}
</script>