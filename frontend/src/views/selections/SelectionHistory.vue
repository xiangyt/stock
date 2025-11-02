<template>
  <ComingSoon 
    title="历史回顾"
    description="选股历史分析功能正在开发中"
    icon="Clock"
    :features="[
      '历史选股记录',
      '选股效果分析',
      '收益率统计',
      '选股趋势图表',
      '策略效果对比'
    ]"
  />
</template>

<script>
import ComingSoon from '../../components/ComingSoon.vue'

export default {
  name: 'SelectionHistory',
  components: {
    ComingSoon
  }
}
</script>