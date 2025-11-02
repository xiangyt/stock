<template>
  <ComingSoon 
    title="策略管理"
    description="智能选股策略管理功能正在开发中"
    icon="SetUp"
    :features="[
      '策略创建和编辑',
      '策略回测分析',
      '策略性能监控',
      '策略参数优化',
      '策略组合管理'
    ]"
  />
</template>

<script>
import ComingSoon from '../../components/ComingSoon.vue'

export default {
  name: 'StrategyList',
  components: {
    ComingSoon
  }
}
</script>