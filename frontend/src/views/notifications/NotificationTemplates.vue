<template>
  <ComingSoon 
    title="消息模板"
    description="消息模板管理功能正在开发中"
    icon="Document"
    :features="[
      '模板创建和编辑',
      '变量占位符支持',
      '模板预览功能',
      '模板分类管理',
      '模板使用统计'
    ]"
  />
</template>

<script>
import ComingSoon from '../../components/ComingSoon.vue'

export default {
  name: 'NotificationTemplates',
  components: {
    ComingSoon
  }
}
</script>