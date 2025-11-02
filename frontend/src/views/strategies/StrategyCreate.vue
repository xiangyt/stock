<template>
  <ComingSoon 
    title="创建策略"
    description="策略创建向导功能正在开发中"
    icon="Plus"
    :features="[
      '可视化策略构建器',
      '技术指标配置',
      '风险控制设置',
      '回测验证工具',
      '策略模板库'
    ]"
  />
</template>

<script>
import ComingSoon from '../../components/ComingSoon.vue'

export default {
  name: 'StrategyCreate',
  components: {
    ComingSoon
  }
}
</script>