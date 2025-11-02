<template>
  <ComingSoon 
    title="投资组合"
    description="投资组合管理功能正在开发中"
    icon="PieChart"
    :features="[
      '组合创建和管理',
      '持仓股票管理',
      '组合业绩分析',
      '风险评估报告',
      '资产配置建议'
    ]"
  />
</template>

<script>
import ComingSoon from '../../components/ComingSoon.vue'

export default {
  name: 'PortfolioList',
  components: {
    ComingSoon
  }
}
</script>