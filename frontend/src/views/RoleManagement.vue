<template>
  <div class="role-management">
    <div class="page-header">
      <h2>角色管理</h2>
      <el-button type="primary" @click="showCreateDialog = true">
        <el-icon><Plus /></el-icon>
        新增角色
      </el-button>
    </div>

    <!-- 搜索栏 -->
    <div class="search-bar">
      <el-row :gutter="20">
        <el-col :span="8">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索角色名称或编码"
            clearable
            @keyup.enter="loadRoles"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </el-col>
        <el-col :span="4">
          <el-button type="primary" @click="loadRoles">搜索</el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-col>
      </el-row>
    </div>

    <!-- 角色列表 -->
    <div class="role-table">
      <el-table 
        :data="roles" 
        v-loading="loading"
        stripe
        style="width: 100%; min-width: 1200px"
      >
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="role_name" label="角色名称" width="150" />
        <el-table-column prop="role_code" label="角色编码" width="150" />
        <el-table-column prop="description" label="描述" min-width="200" />
        <el-table-column label="系统角色" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.is_system ? 'warning' : 'info'">
              {{ scope.row.is_system ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="scope">
            <el-tag 
              :type="scope.row.status === 1 ? 'success' : 'danger'"
            >
              {{ scope.row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="sort_order" label="排序" width="80" />
        <el-table-column prop="created_at" label="创建时间" width="200">
          <template #default="scope">
            {{ formatDateTime(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="scope">
            <el-button 
              type="primary" 
              size="small" 
              @click="editRole(scope.row)"
              :disabled="scope.row.is_system"
            >
              编辑
            </el-button>
            <el-button 
              type="info" 
              size="small" 
              @click="managePermissions(scope.row)"
              :disabled="scope.row.is_system"
            >
              权限
            </el-button>
            <el-button 
              :type="scope.row.status === 1 ? 'warning' : 'success'"
              size="small" 
              @click="toggleRoleStatus(scope.row)"
              :disabled="scope.row.is_system"
            >
              {{ scope.row.status === 1 ? '禁用' : '启用' }}
            </el-button>
            <el-button 
              type="danger" 
              size="small" 
              @click="deleteRole(scope.row)"
              :disabled="scope.row.is_system"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="loadRoles"
          @current-change="loadRoles"

        />
      </div>
    </div>

    <!-- 创建/编辑角色对话框 -->
    <el-dialog
      v-model="showCreateDialog"
      :title="editingRole ? '编辑角色' : '新增角色'"
      width="600px"
    >
      <el-form
        ref="roleFormRef"
        :model="roleForm"
        :rules="roleFormRules"
        label-width="100px"
      >
        <el-form-item label="角色名称" prop="role_name">
          <el-input 
            v-model="roleForm.role_name" 
            placeholder="请输入角色名称"
          />
        </el-form-item>
        <el-form-item label="角色编码" prop="role_code">
          <el-input 
            v-model="roleForm.role_code" 
            :disabled="editingRole && editingRole.is_system"
            placeholder="请输入角色编码"
          />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input 
            v-model="roleForm.description" 
            type="textarea"
            :rows="3"
            placeholder="请输入角色描述"
          />
        </el-form-item>
        <el-form-item label="排序" prop="sort_order">
          <el-input-number 
            v-model="roleForm.sort_order" 
            :min="0"
            :max="999"
            placeholder="排序值"
          />
        </el-form-item>
        <el-form-item label="状态" prop="status" v-if="editingRole">
          <el-select v-model="roleForm.status" placeholder="请选择状态">
            <el-option label="启用" :value="1" />
            <el-option label="禁用" :value="0" />
          </el-select>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="saveRole" :loading="saving">
          {{ editingRole ? '更新' : '创建' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 权限管理对话框 -->
    <el-dialog
      v-model="showPermissionDialog"
      title="权限管理"
      width="60%"
      :close-on-click-modal="false"
      top="8vh"
    >
      <div v-if="currentRole" class="permission-dialog">
        <div class="role-info">
          <h4>
            <el-icon><UserFilled /></el-icon>
            角色：{{ currentRole.role_name }}
          </h4>
          <p class="role-desc">{{ currentRole.description || '暂无描述' }}</p>
        </div>
        
        <div class="permission-content">
          <div class="permission-header">
            <span class="permission-title">
              <el-icon><Key /></el-icon>
              权限配置
            </span>
            <div class="permission-actions">
              <el-button size="small" @click="expandAll">全部展开</el-button>
              <el-button size="small" @click="collapseAll">全部收起</el-button>
              <el-button size="small" type="primary" @click="checkAll">全选</el-button>
              <el-button size="small" @click="uncheckAll">全不选</el-button>
            </div>
          </div>
          
          <div class="permission-tree-container" v-loading="loadingPermissions">
            <el-tree
              ref="permissionTreeRef"
              :data="menuPermissionTree"
              :props="menuTreeProps"
              show-checkbox
              node-key="id"
              :default-checked-keys="rolePermissions"
              :check-strictly="false"
              :expand-on-click-node="false"
              :key="`tree-${currentRole?.id}-${rolePermissions.length}`"
              class="permission-tree"
            >
              <template #default="{ node, data }">
                <span class="tree-node-label">{{ data.label }}</span>
              </template>
            </el-tree>
          </div>
        </div>
      </div>
      
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="showPermissionDialog = false">取消</el-button>
          <el-button type="primary" @click="savePermissions" :loading="savingPermissions">
            <el-icon><Check /></el-icon>
            保存权限
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  Plus, 
  Search, 
  UserFilled, 
  Key, 
  Check,
  Folder,
  Document,
  Switch,
  Link,
  Setting,
  House,
  TrendCharts,
  Operation,
  Briefcase,
  Download,
  Bell,
  DataAnalysis,
  Monitor
} from '@element-plus/icons-vue'
import { roleAPI, permissionAPI } from '@/utils/api'
import { MENU_STRUCTURE } from '@/config/menu'

// 响应式数据
const loading = ref(false)
const saving = ref(false)
const savingPermissions = ref(false)
const loadingPermissions = ref(false)
const roles = ref([])
const permissions = ref([])
const permissionTree = ref([])
const menuPermissionTree = ref([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(10)
const searchKeyword = ref('')
const showCreateDialog = ref(false)
const showPermissionDialog = ref(false)
const editingRole = ref(null)
const currentRole = ref(null)
const rolePermissions = ref([])

// 角色表单
const roleForm = reactive({
  role_name: '',
  role_code: '',
  description: '',
  sort_order: 0,
  status: 1
})

// 表单验证规则
const roleFormRules = {
  role_name: [
    { required: true, message: '请输入角色名称', trigger: 'blur' },
    { min: 2, max: 50, message: '角色名称长度在 2 到 50 个字符', trigger: 'blur' }
  ],
  role_code: [
    { required: true, message: '请输入角色编码', trigger: 'blur' },
    { min: 2, max: 50, message: '角色编码长度在 2 到 50 个字符', trigger: 'blur' },
    { pattern: /^[a-zA-Z][a-zA-Z0-9_]*$/, message: '角色编码只能包含字母、数字和下划线，且以字母开头', trigger: 'blur' }
  ]
}

// 树形控件属性
const treeProps = {
  children: 'children',
  label: 'permission_name'
}

// 菜单树形控件属性
const menuTreeProps = {
  children: 'children',
  label: 'label'
}

const roleFormRef = ref()
const permissionTreeRef = ref()

// 加载角色列表
const loadRoles = async () => {
  loading.value = true
  try {
    const response = await roleAPI.getRoleList({
      page: currentPage.value,
      page_size: pageSize.value,
      keyword: searchKeyword.value
    })
    
    // 处理嵌套的数据结构
    const data = response.data || response
    roles.value = data.roles || []
    total.value = data.total || 0
  } catch (error) {
    console.error('获取角色列表失败:', error)
    ElMessage.error('获取角色列表失败')
  } finally {
    loading.value = false
  }
}

// 加载权限列表
const loadPermissions = async () => {
  try {
    const response = await permissionAPI.getPermissionList()
    // 处理嵌套的数据结构
    const data = response.data || response
    permissions.value = data.permissions || []
    buildPermissionTree()
  } catch (error) {
    console.error('获取权限列表失败:', error)
  }
}

// 构建权限树
const buildPermissionTree = () => {
  const tree = []
  const map = {}
  
  // 创建映射
  permissions.value.forEach(permission => {
    map[permission.id] = { ...permission, children: [] }
  })
  
  // 构建树结构
  permissions.value.forEach(permission => {
    if (permission.parent_id) {
      if (map[permission.parent_id]) {
        map[permission.parent_id].children.push(map[permission.id])
      }
    } else {
      tree.push(map[permission.id])
    }
  })
  
  permissionTree.value = tree
  buildMenuPermissionTree()
}

// 构建菜单权限树 - 基于统一的菜单配置与后端权限的交集
const buildMenuPermissionTree = () => {
  console.log('开始构建权限树')
  console.log('数据库权限列表:', permissions.value)
  
  if (!permissions.value || permissions.value.length === 0) {
    console.warn('权限数据为空，使用前端配置构建权限树')
    // 使用统一的菜单结构构建空权限树
    menuPermissionTree.value = MENU_STRUCTURE.map(item => ({
      id: item.id,
      label: item.name,
      children: item.children?.map(child => ({
        id: child.id,
        label: child.name
      })) || []
    }))
    return
  }
  
  // 创建权限编码到ID的映射
  const permissionCodeMap = {}
  permissions.value.forEach(permission => {
    const code = permission.permission_code || permission.code
    if (code) {
      permissionCodeMap[code] = permission.id
    }
  })
  
  console.log('权限编码映射:', permissionCodeMap)
  
  // 构建权限树 - 只包含数据库中存在的权限
  const buildTreeNode = (menuItem) => {
    const permissionId = permissionCodeMap[menuItem.permission]
    if (!permissionId) {
      // 如果数据库中没有对应权限，跳过此节点
      return null
    }
    
    const node = {
      id: permissionId.toString(),
      label: menuItem.name,
      children: []
    }
    
    // 递归处理子节点
    if (menuItem.children && menuItem.children.length > 0) {
      menuItem.children.forEach(childItem => {
        const childNode = buildTreeNode(childItem)
        if (childNode) {
          node.children.push(childNode)
        }
      })
    }
    
    return node
  }
  
  const treeData = []
  MENU_STRUCTURE.forEach(menuItem => {
    const node = buildTreeNode(menuItem)
    if (node) {
      treeData.push(node)
    }
  })
  
  console.log('构建的权限树数据:', treeData)
  menuPermissionTree.value = treeData
}

// 根据权限编码获取模块名称
const getModuleName = (permissionCode) => {
  if (permissionCode.startsWith('dashboard:')) return '仪表板'
  if (permissionCode.startsWith('stocks:')) return '股票数据'
  if (permissionCode.startsWith('strategies:')) return '策略管理'
  if (permissionCode.startsWith('selections:')) return '选股结果'
  if (permissionCode.startsWith('portfolios:')) return '投资组合'
  if (permissionCode.startsWith('collectors:')) return '数据采集'
  if (permissionCode.startsWith('notifications:')) return '通知管理'
  if (permissionCode.startsWith('reports:')) return '报表分析'
  if (permissionCode.startsWith('system:')) return '系统管理'
  return null
}

// 过滤重复权限，只保留标准格式的权限编码
const filterDuplicatePermissions = (permissions) => {
  const filtered = []
  const seen = new Set()
  
  // 定义标准权限编码的优先级
  const standardCodes = {
    '用户管理': 'system:users',
    '角色管理': 'system:roles', 
    '系统监控': 'system:monitoring',
    '系统设置': 'system:settings'
  }
  
  permissions.forEach(permission => {
    const key = permission.permission_name
    
    // 如果是重复的权限名称，只保留标准编码格式的
    if (standardCodes[key]) {
      if (permission.permission_code === standardCodes[key] && !seen.has(key)) {
        filtered.push(permission)
        seen.add(key)
      }
    } else {
      // 对于非重复的权限，直接添加
      if (!seen.has(key)) {
        filtered.push(permission)
        seen.add(key)
      }
    }
  })
  
  return filtered
}

// 获取节点图标
const getNodeIcon = (data) => {
  if (data.icon) return data.icon
  
  const iconMap = {
    'dashboard:view': 'House',
    'stocks:module': 'TrendCharts',
    'strategies:module': 'Operation',
    'selections:module': 'Search',
    'portfolios:module': 'Briefcase',
    'collectors:module': 'Download',
    'notifications:module': 'Bell',
    'reports:module': 'DataAnalysis',
    'system:module': 'Setting'
  }
  
  return iconMap[data.permission_code] || (data.resource_type === 'module' ? 'Folder' : 'Document')
}

// 获取节点图标样式类
const getNodeIconClass = (data) => {
  if (data.resource_type === 'module') return 'module-icon'
  return 'menu-icon'
}

// 获取资源类型颜色
const getResourceTypeColor = (type) => {
  const colors = {
    module: 'warning',
    menu: 'primary',
    button: 'success',
    api: 'info'
  }
  return colors[type] || 'info'
}

// 获取资源类型标签
const getResourceTypeLabel = (type) => {
  const labels = {
    module: '模块',
    menu: '菜单',
    button: '按钮',
    api: 'API'
  }
  return labels[type] || type
}

// 展开所有节点
const expandAll = () => {
  if (permissionTreeRef.value) {
    const allKeys = []
    const collectKeys = (nodes) => {
      nodes.forEach(node => {
        allKeys.push(node.id)
        if (node.children && node.children.length > 0) {
          collectKeys(node.children)
        }
      })
    }
    collectKeys(menuPermissionTree.value)
    
    // 展开所有节点
    allKeys.forEach(key => {
      permissionTreeRef.value.store.nodesMap[key]?.expand()
    })
  }
}

// 收起所有节点
const collapseAll = () => {
  if (permissionTreeRef.value) {
    const allKeys = []
    const collectKeys = (nodes) => {
      nodes.forEach(node => {
        allKeys.push(node.id)
        if (node.children && node.children.length > 0) {
          collectKeys(node.children)
        }
      })
    }
    collectKeys(menuPermissionTree.value)
    
    // 收起所有节点
    allKeys.forEach(key => {
      permissionTreeRef.value.store.nodesMap[key]?.collapse()
    })
  }
}

// 全选
const checkAll = () => {
  if (permissionTreeRef.value) {
    const allKeys = []
    const collectKeys = (nodes) => {
      nodes.forEach(node => {
        allKeys.push(node.id)
        if (node.children && node.children.length > 0) {
          collectKeys(node.children)
        }
      })
    }
    collectKeys(menuPermissionTree.value)
    permissionTreeRef.value.setCheckedKeys(allKeys)
  }
}

// 全不选
const uncheckAll = () => {
  if (permissionTreeRef.value) {
    permissionTreeRef.value.setCheckedKeys([])
  }
}

// 重置搜索
const resetSearch = () => {
  searchKeyword.value = ''
  currentPage.value = 1
  loadRoles()
}

// 编辑角色
const editRole = (role) => {
  editingRole.value = role
  Object.assign(roleForm, {
    role_name: role.role_name,
    role_code: role.role_code,
    description: role.description || '',
    sort_order: role.sort_order || 0,
    status: role.status
  })
  showCreateDialog.value = true
}

// 保存角色
const saveRole = async () => {
  if (!roleFormRef.value) return
  
  try {
    await roleFormRef.value.validate()
    saving.value = true
    
    // 构建请求数据，确保字段格式正确
    const requestData = {
      role_name: roleForm.role_name,
      role_code: roleForm.role_code,
      description: roleForm.description || '',
      sort_order: parseInt(roleForm.sort_order) || 0,
      status: parseInt(roleForm.status)
    }
    
    console.log('发送角色数据:', requestData)
    
    let response
    if (editingRole.value) {
      // 更新角色
      response = await roleAPI.updateRole(editingRole.value.id, requestData)
    } else {
      // 创建角色
      response = await roleAPI.createRole(requestData)
    }
    
    console.log('角色保存响应:', response)
    
    ElMessage.success(editingRole.value ? '角色更新成功' : '角色创建成功')
    showCreateDialog.value = false
    resetForm()
    loadRoles()
  } catch (error) {
    console.error('保存角色失败:', error)
    ElMessage.error('操作失败: ' + (error.response?.data?.message || error.message))
  } finally {
    saving.value = false
  }
}

// 切换角色状态
const toggleRoleStatus = async (role) => {
  const action = role.status === 1 ? '禁用' : '启用'
  try {
    await ElMessageBox.confirm(
      `确定要${action}角色 "${role.role_name}" 吗？`,
      '确认操作',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    const newStatus = role.status === 1 ? 0 : 1
    const requestData = { status: parseInt(newStatus) }
    
    console.log(`${action}角色请求数据:`, requestData)
    
    const response = await roleAPI.updateRole(role.id, requestData)
    
    console.log(`${action}角色响应:`, response)
    
    ElMessage.success(`${action}成功`)
    loadRoles()
  } catch (error) {
    if (error !== 'cancel') {
      console.error(`${action}角色失败:`, error)
      ElMessage.error(`${action}失败: ` + (error.response?.data?.message || error.message))
    }
  }
}

// 删除角色
const deleteRole = async (role) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除角色 "${role.role_name}" 吗？此操作不可恢复！`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    const response = await roleAPI.deleteRole(role.id)
    
    ElMessage.success('删除成功')
    loadRoles()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除角色失败:', error)
      ElMessage.error('删除失败')
    }
  }
}

// 管理权限
const managePermissions = async (role) => {
  currentRole.value = role
  loadingPermissions.value = true
  
  try {
    // 先加载所有权限数据
    await loadPermissions()
    
    // 获取角色权限
    const response = await roleAPI.getRolePermissions(role.id)
    console.log('角色权限响应:', response)
    
    // 确保正确提取权限ID数组
    let permissionIds = []
    if (response.permissions && Array.isArray(response.permissions)) {
      permissionIds = response.permissions.map(p => p.id)
    } else if (response.data && response.data.permissions && Array.isArray(response.data.permissions)) {
      permissionIds = response.data.permissions.map(p => p.id)
    } else if (Array.isArray(response)) {
      permissionIds = response.map(p => p.id || p)
    }
    
    // 转换为字符串数组，因为树组件的node-key是字符串
    rolePermissions.value = permissionIds.map(id => id.toString())
    console.log('角色权限ID列表:', rolePermissions.value)
    console.log('菜单权限树:', menuPermissionTree.value)
    
    // 延迟显示对话框，确保数据完全加载
    setTimeout(() => {
      showPermissionDialog.value = true
      // 在对话框显示后设置选中状态
      setTimeout(() => {
        if (permissionTreeRef.value && rolePermissions.value.length > 0) {
          console.log('设置默认选中权限:', rolePermissions.value)
          permissionTreeRef.value.setCheckedKeys(rolePermissions.value)
        }
      }, 100)
    }, 50)
  } catch (error) {
    console.error('获取角色权限失败:', error)
    ElMessage.error('获取角色权限失败')
  } finally {
    loadingPermissions.value = false
  }
}

// 保存权限
const savePermissions = async () => {
  if (!currentRole.value || !permissionTreeRef.value) return
  
  try {
    savingPermissions.value = true
    
    const checkedKeys = permissionTreeRef.value.getCheckedKeys()
    const halfCheckedKeys = permissionTreeRef.value.getHalfCheckedKeys()
    const allKeys = [...checkedKeys, ...halfCheckedKeys]
    
    // 确保权限ID是数字类型，不是字符串
    const permissionIds = allKeys.map(id => parseInt(id, 10)).filter(id => !isNaN(id))
    
    console.log('发送的权限ID:', permissionIds)
    
    const response = await roleAPI.updateRolePermissions(currentRole.value.id, {
      permission_ids: permissionIds
    })
    
    console.log('权限保存响应:', response)
    
    // 更宽松的成功判断逻辑
    const isSuccess = response.code === 0 || 
                     response.data?.code === 0 || 
                     response.status === 200 ||
                     (response.status >= 200 && response.status < 300) ||
                     !response.error
    
    if (isSuccess) {
      ElMessage.success('权限更新成功')
      showPermissionDialog.value = false
    } else {
      ElMessage.error(response.message || response.data?.message || '权限更新失败')
    }
  } catch (error) {
    console.error('保存权限失败:', error)
    ElMessage.error('保存权限失败')
  } finally {
    savingPermissions.value = false
  }
}

// 重置表单
const resetForm = () => {
  editingRole.value = null
  Object.assign(roleForm, {
    role_name: '',
    role_code: '',
    description: '',
    sort_order: 0,
    status: 1
  })
  if (roleFormRef.value) {
    roleFormRef.value.resetFields()
  }
}

// 格式化日期时间
const formatDateTime = (dateTime) => {
  if (!dateTime) return '-'
  return new Date(dateTime).toLocaleString('zh-CN')
}

// 组件挂载时加载数据
onMounted(() => {
  loadRoles()
  loadPermissions()
})
</script>

<style scoped>
.role-management {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
  color: #303133;
}

.search-bar {
  margin-bottom: 20px;
  padding: 20px;
  background: #f5f7fa;
  border-radius: 4px;
}

.role-table {
  background: white;
  border-radius: 4px;
  overflow: hidden;
}

.pagination {
  padding: 20px;
  text-align: right;
  background: white;
  border-top: 1px solid #ebeef5;
}

/* 表格基础样式 */
:deep(.el-table) {
  border: 1px solid #ebeef5;
  border-radius: 4px;
  overflow: hidden;
}

/* 表头样式 - 加黑加粗 */
:deep(.el-table .el-table__header-wrapper) {
  background: #f8f9fa;
}

:deep(.el-table .el-table__header th) {
  background: #f8f9fa !important;
  color: #303133 !important;
  font-weight: 700 !important;
  font-size: 14px;
  border-bottom: 2px solid #e4e7ed;
  padding: 12px 0;
  text-align: left;
}

:deep(.el-table .el-table__header th .cell) {
  font-weight: 700;
  color: #303133;
}

/* 数据行样式 */
:deep(.el-table .el-table__body tr td) {
  background: #ffffff;
  border-bottom: 1px solid #f0f0f0;
  padding: 12px 0;
  transition: background-color 0.25s ease;
}

:deep(.el-table .el-table__body tr:hover td) {
  background: #f5f7fa !important;
}

:deep(.el-table .el-table__body tr.el-table__row--striped td) {
  background: #fafbfc;
}

:deep(.el-table .el-table__body tr.el-table__row--striped:hover td) {
  background: #f5f7fa !important;
}

/* 右侧固定列样式 */
:deep(.el-table .el-table__fixed-right) {
  background: #ffffff;
  box-shadow: -2px 0 8px rgba(0, 0, 0, 0.1);
  z-index: 3;
}

:deep(.el-table .el-table__fixed-right .el-table__fixed-header-wrapper) {
  background: #f8f9fa;
  z-index: 4;
}

:deep(.el-table .el-table__fixed-right .el-table__header th) {
  background: #f8f9fa !important;
  color: #303133 !important;
  font-weight: 700 !important;
  font-size: 14px;
  border-bottom: 2px solid #e4e7ed;
  padding: 12px 0;
  text-align: left;
}

:deep(.el-table .el-table__fixed-right .el-table__body tr td) {
  background: #ffffff;
  border-bottom: 1px solid #f0f0f0;
  padding: 12px 0;
  transition: background-color 0.25s ease;
}

:deep(.el-table .el-table__fixed-right .el-table__body tr:hover td) {
  background: #f5f7fa !important;
}

:deep(.el-table .el-table__fixed-right .el-table__body tr.el-table__row--striped td) {
  background: #fafbfc;
}

:deep(.el-table .el-table__fixed-right .el-table__body tr.el-table__row--striped:hover td) {
  background: #f5f7fa !important;
}

:deep(.el-table .el-table__fixed-right-patch) {
  background: #f8f9fa;
  z-index: 4;
}

/* 左侧滚动区域样式同步 */
:deep(.el-table .el-table__body-wrapper) {
  overflow-x: auto;
  overflow-y: hidden;
  scroll-behavior: smooth;
}

:deep(.el-table .el-table__header-wrapper) {
  overflow: hidden;
  position: relative;
}

/* 优化滚动时表头和数据行的同步性能 */
:deep(.el-table .el-table__header),
:deep(.el-table .el-table__body) {
  transition: none;
  will-change: transform;
}

/* 减少滚动时的重绘和回流 */
:deep(.el-table .el-table__header-wrapper),
:deep(.el-table .el-table__body-wrapper) {
  transform: translateZ(0);
  backface-visibility: hidden;
  perspective: 1000px;
}

/* 表头固定优化 */
:deep(.el-table .el-table__header) {
  position: relative;
  z-index: 2;
  transform: translate3d(0, 0, 0);
}

/* 数据行滚动优化 */
:deep(.el-table .el-table__body) {
  position: relative;
  z-index: 1;
  transform: translate3d(0, 0, 0);
}

/* 单元格内容样式 */
:deep(.el-table td .cell),
:deep(.el-table th .cell) {
  padding: 0 12px;
  word-break: break-word;
  line-height: 1.5;
}

/* 按钮组样式优化 */
:deep(.el-table .el-button + .el-button) {
  margin-left: 8px;
}

:deep(.el-table .el-button--small) {
  padding: 5px 12px;
  font-size: 12px;
}

:deep(.el-dialog__body) {
  padding: 20px;
}

:deep(.el-tree) {
  margin-top: 10px;
}

/* 权限管理对话框样式 */
.permission-dialog {
  padding: 0;
}

.role-info {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 20px;
  border-radius: 8px;
  margin-bottom: 20px;
}

.role-info h4 {
  margin: 0 0 8px 0;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 18px;
}

.role-desc {
  margin: 0;
  opacity: 0.9;
  font-size: 14px;
}

.permission-content {
  background: #f8f9fa;
  border-radius: 8px;
  overflow: hidden;
}

.permission-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  background: white;
  border-bottom: 1px solid #e9ecef;
}

.permission-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  color: #495057;
}

.permission-actions {
  display: flex;
  gap: 8px;
}

.permission-tree-container {
  background: white;
  max-height: 400px;
  overflow-y: auto;
  padding: 16px;
}

.permission-tree {
  background: transparent;
}

.tree-node-label {
  font-weight: 500;
  color: #303133;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 0 0 0;
}

:deep(.permission-tree .el-tree-node__content) {
  height: auto;
  padding: 8px 0;
}

:deep(.permission-tree .el-tree-node__expand-icon) {
  color: #409eff;
}

:deep(.permission-tree .el-checkbox) {
  margin-right: 8px;
}

:deep(.permission-tree .el-tree-node__label) {
  flex: 1;
}
</style>