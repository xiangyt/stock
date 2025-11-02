<template>
  <ComingSoon 
    title="用户管理"
    description="用户管理功能正在开发中"
    icon="User"
    :features="[
      '用户账号管理',
      '角色权限配置',
      '用户操作日志',
      '登录安全控制',
      '用户数据统计'
    ]"
  />
</template>

<script>
import ComingSoon from '../../components/ComingSoon.vue'

export default {
  name: 'SystemUsers',
  components: {
    ComingSoon
  }
}
</script>