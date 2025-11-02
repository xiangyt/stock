<template>
  <ComingSoon 
    title="选股记录"
    description="选股结果管理功能正在开发中"
    icon="Select"
    :features="[
      '选股结果展示',
      '选股历史记录',
      '选股成功率统计',
      '选股结果导出',
      '选股策略关联'
    ]"
  />
</template>

<script>
import ComingSoon from '../../components/ComingSoon.vue'

export default {
  name: 'SelectionList',
  components: {
    ComingSoon
  }
}
</script>