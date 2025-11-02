<template>
  <ComingSoon 
    title="风险分析"
    description="风险分析报告功能正在开发中"
    icon="Warning"
    :features="[
      '投资组合风险评估',
      'VaR风险度量',
      '压力测试分析',
      '风险预警系统',
      '风险控制建议'
    ]"
  />
</template>

<script>
import ComingSoon from '../../components/ComingSoon.vue'

export default {
  name: 'ReportRisk',
  components: {
    ComingSoon
  }
}
</script>