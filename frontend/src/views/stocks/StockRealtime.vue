<template>
  <ComingSoon 
    title="实时行情"
    description="实时股票行情数据展示功能正在开发中"
    icon="DataLine"
    :features="[
      '实时价格更新',
      '涨跌幅排行榜',
      '成交量统计',
      '市场热点分析',
      '自定义监控面板'
    ]"
  />
</template>

<script>
import ComingSoon from '../../components/ComingSoon.vue'

export default {
  name: 'StockRealtime',
  components: {
    ComingSoon
  }
}
</script>