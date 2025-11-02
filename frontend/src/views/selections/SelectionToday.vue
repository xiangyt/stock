<template>
  <ComingSoon 
    title="今日选股"
    description="今日选股结果展示功能正在开发中"
    icon="Calendar"
    :features="[
      '今日选股列表',
      '实时价格更新',
      '选股评分排序',
      '快速操作面板',
      '选股提醒设置'
    ]"
  />
</template>

<script>
import ComingSoon from '../../components/ComingSoon.vue'

export default {
  name: 'SelectionToday',
  components: {
    ComingSoon
  }
}
</script>