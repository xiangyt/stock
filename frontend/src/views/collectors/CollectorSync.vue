<template>
  <ComingSoon 
    title="数据同步"
    description="数据同步管理功能正在开发中"
    icon="Refresh"
    :features="[
      '全量数据同步',
      '增量数据更新',
      '同步进度监控',
      '数据一致性检查',
      '同步策略配置'
    ]"
  />
</template>

<script>
import ComingSoon from '../../components/ComingSoon.vue'

export default {
  name: 'CollectorSync',
  components: {
    ComingSoon
  }
}
</script>