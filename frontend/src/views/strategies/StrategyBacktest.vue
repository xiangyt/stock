<template>
  <ComingSoon 
    title="回测中心"
    description="策略回测分析功能正在开发中"
    icon="TrendCharts"
    :features="[
      '历史数据回测',
      '收益率分析',
      '风险指标计算',
      '回测报告生成',
      '参数敏感性分析'
    ]"
  />
</template>

<script>
import ComingSoon from '../../components/ComingSoon.vue'

export default {
  name: 'StrategyBacktest',
  components: {
    ComingSoon
  }
}
</script>