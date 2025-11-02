<template>
  <div v-if="hasAccess">
    <slot />
  </div>
  <div v-else-if="showFallback">
    <slot name="fallback">
      <div class="permission-denied">
        <el-empty description="您没有权限查看此内容" />
      </div>
    </slot>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { hasPermission, hasRole } from '@/utils/auth'

const props = defineProps({
  // 权限码
  permission: {
    type: String,
    default: ''
  },
  // 角色码
  role: {
    type: String,
    default: ''
  },
  // 权限码数组（满足任一即可）
  permissions: {
    type: Array,
    default: () => []
  },
  // 角色码数组（满足任一即可）
  roles: {
    type: Array,
    default: () => []
  },
  // 是否需要同时满足所有权限
  requireAll: {
    type: Boolean,
    default: false
  },
  // 是否显示无权限时的占位内容
  showFallback: {
    type: Boolean,
    default: false
  }
})

// 检查是否有访问权限
const hasAccess = computed(() => {
  // 如果没有设置任何权限要求，默认显示
  if (!props.permission && !props.role && props.permissions.length === 0 && props.roles.length === 0) {
    return true
  }
  
  let hasPermissionAccess = true
  let hasRoleAccess = true
  
  // 检查单个权限
  if (props.permission) {
    hasPermissionAccess = hasPermission(props.permission)
  }
  
  // 检查权限数组
  if (props.permissions.length > 0) {
    if (props.requireAll) {
      // 需要满足所有权限
      hasPermissionAccess = props.permissions.every(p => hasPermission(p))
    } else {
      // 满足任一权限即可
      hasPermissionAccess = props.permissions.some(p => hasPermission(p))
    }
  }
  
  // 检查单个角色
  if (props.role) {
    hasRoleAccess = hasRole(props.role)
  }
  
  // 检查角色数组
  if (props.roles.length > 0) {
    if (props.requireAll) {
      // 需要满足所有角色
      hasRoleAccess = props.roles.every(r => hasRole(r))
    } else {
      // 满足任一角色即可
      hasRoleAccess = props.roles.some(r => hasRole(r))
    }
  }
  
  return hasPermissionAccess && hasRoleAccess
})
</script>

<style scoped>
.permission-denied {
  padding: 20px;
  text-align: center;
}
</style>