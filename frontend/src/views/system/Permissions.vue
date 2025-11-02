<template>
  <div class="permission-management">
    <div class="page-header">
      <h2>权限管理</h2>
      <div class="header-actions">
        <el-button type="primary" @click="showCreateDialog = true">
          <el-icon><Plus /></el-icon>
          新增权限
        </el-button>
        <el-button type="success" @click="resetPermissions">
          <el-icon><Refresh /></el-icon>
          重置权限
        </el-button>
      </div>
    </div>

    <!-- 搜索栏 -->
    <div class="search-bar">
      <el-row :gutter="20">
        <el-col :span="8">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索权限名称或编码"
            clearable
            @keyup.enter="filterPermissions"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </el-col>
        <el-col :span="6">
          <el-select v-model="filterType" placeholder="权限类型" clearable>
            <el-option label="菜单" value="menu" />
            <el-option label="模块" value="module" />
            <el-option label="按钮" value="button" />
            <el-option label="API" value="api" />
          </el-select>
        </el-col>
        <el-col :span="4">
          <el-button type="primary" @click="filterPermissions">搜索</el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-col>
      </el-row>
    </div>

    <!-- 权限树形表格 -->
    <div class="permission-tree">
      <el-table
        :data="filteredPermissions"
        v-loading="loading"
        row-key="id"
        :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
        stripe
        style="width: 100%"
        :default-expand-all="false"
      >
        <el-table-column prop="permission_name" label="权限名称" min-width="200">
          <template #default="scope">
            <div class="permission-name">
              <el-icon v-if="scope.row.resource_type === 'module'" class="module-icon">
                <Folder />
              </el-icon>
              <el-icon v-else-if="scope.row.resource_type === 'menu'" class="menu-icon">
                <Document />
              </el-icon>
              <el-icon v-else-if="scope.row.resource_type === 'button'" class="button-icon">
                <Switch />
              </el-icon>
              <el-icon v-else class="api-icon">
                <Link />
              </el-icon>
              <span>{{ scope.row.permission_name }}</span>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column prop="permission_code" label="权限编码" width="200">
          <template #default="scope">
            <el-tag type="info" size="small">{{ scope.row.permission_code }}</el-tag>
          </template>
        </el-table-column>
        
        <el-table-column prop="resource_type" label="类型" width="100">
          <template #default="scope">
            <el-tag 
              :type="getTypeColor(scope.row.resource_type)" 
              size="small"
            >
              {{ getTypeLabel(scope.row.resource_type) }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column prop="resource_path" label="资源路径" width="200">
          <template #default="scope">
            <code v-if="scope.row.resource_path" class="resource-path">
              {{ scope.row.resource_path }}
            </code>
            <span v-else class="no-path">-</span>
          </template>
        </el-table-column>
        
        <el-table-column label="状态" width="100">
          <template #default="scope">
            <el-tag 
              :type="scope.row.status === 1 ? 'success' : 'danger'"
              size="small"
            >
              {{ scope.row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column prop="sort_order" label="排序" width="80" />
        
        <el-table-column prop="description" label="描述" min-width="150" />
        
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="scope">
            <el-button 
              type="primary" 
              size="small" 
              @click="editPermission(scope.row)"
            >
              编辑
            </el-button>
            <el-button 
              type="success" 
              size="small" 
              @click="addChildPermission(scope.row)"
              v-if="scope.row.resource_type === 'module'"
            >
              添加子权限
            </el-button>
            <el-button 
              type="danger" 
              size="small" 
              @click="deletePermission(scope.row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 创建/编辑权限对话框 -->
    <el-dialog
      v-model="showCreateDialog"
      :title="editingPermission ? '编辑权限' : '新增权限'"
      width="700px"
    >
      <el-form
        ref="permissionFormRef"
        :model="permissionForm"
        :rules="permissionFormRules"
        label-width="120px"
      >
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="权限名称" prop="permission_name">
              <el-input 
                v-model="permissionForm.permission_name" 
                placeholder="请输入权限名称"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="权限编码" prop="permission_code">
              <el-input 
                v-model="permissionForm.permission_code" 
                placeholder="请输入权限编码"
              />
            </el-form-item>
          </el-col>
        </el-row>
        
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="权限类型" prop="resource_type">
              <el-select v-model="permissionForm.resource_type" placeholder="请选择权限类型">
                <el-option label="模块" value="module" />
                <el-option label="菜单" value="menu" />
                <el-option label="按钮" value="button" />
                <el-option label="API" value="api" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="父级权限" prop="parent_id">
              <el-tree-select
                v-model="permissionForm.parent_id"
                :data="parentPermissionOptions"
                :props="{ label: 'permission_name', value: 'id' }"
                placeholder="请选择父级权限"
                clearable
                check-strictly
              />
            </el-form-item>
          </el-col>
        </el-row>
        
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="资源路径" prop="resource_path">
              <el-input 
                v-model="permissionForm.resource_path" 
                placeholder="请输入资源路径"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="排序" prop="sort_order">
              <el-input-number 
                v-model="permissionForm.sort_order" 
                :min="0"
                :max="999"
                placeholder="排序值"
              />
            </el-form-item>
          </el-col>
        </el-row>
        
        <el-form-item label="描述" prop="description">
          <el-input 
            v-model="permissionForm.description" 
            type="textarea"
            :rows="3"
            placeholder="请输入权限描述"
          />
        </el-form-item>
        
        <el-form-item label="状态" prop="status" v-if="editingPermission">
          <el-select v-model="permissionForm.status" placeholder="请选择状态">
            <el-option label="启用" :value="1" />
            <el-option label="禁用" :value="0" />
          </el-select>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="savePermission" :loading="saving">
          {{ editingPermission ? '更新' : '创建' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search, Refresh, Folder, Document, Switch, Link } from '@element-plus/icons-vue'
import { permissionAPI } from '@/utils/api'

// 响应式数据
const loading = ref(false)
const saving = ref(false)
const permissions = ref([])
const filteredPermissions = ref([])
const searchKeyword = ref('')
const filterType = ref('')
const showCreateDialog = ref(false)
const editingPermission = ref(null)
const parentPermission = ref(null)

// 权限表单
const permissionForm = reactive({
  permission_name: '',
  permission_code: '',
  resource_type: 'menu',
  resource_path: '',
  parent_id: null,
  description: '',
  sort_order: 0,
  status: 1
})

// 表单验证规则
const permissionFormRules = {
  permission_name: [
    { required: true, message: '请输入权限名称', trigger: 'blur' },
    { min: 2, max: 50, message: '权限名称长度在 2 到 50 个字符', trigger: 'blur' }
  ],
  permission_code: [
    { required: true, message: '请输入权限编码', trigger: 'blur' },
    { min: 2, max: 100, message: '权限编码长度在 2 到 100 个字符', trigger: 'blur' },
    { pattern: /^[a-zA-Z][a-zA-Z0-9:_]*$/, message: '权限编码只能包含字母、数字、冒号和下划线，且以字母开头', trigger: 'blur' }
  ],
  resource_type: [
    { required: true, message: '请选择权限类型', trigger: 'change' }
  ]
}

const permissionFormRef = ref()

// 父级权限选项
const parentPermissionOptions = computed(() => {
  const buildOptions = (perms, level = 0) => {
    const options = []
    perms.forEach(perm => {
      if (editingPermission.value && perm.id === editingPermission.value.id) {
        return // 不能选择自己作为父级
      }
      
      options.push({
        id: perm.id,
        permission_name: '  '.repeat(level) + perm.permission_name,
        children: perm.children ? buildOptions(perm.children, level + 1) : []
      })
    })
    return options
  }
  
  return buildOptions(permissions.value)
})

// 加载权限列表
const loadPermissions = async () => {
  loading.value = true
  try {
    const response = await permissionAPI.getPermissionList()
    permissions.value = buildPermissionTree(response.permissions || [])
    filterPermissions()
  } catch (error) {
    console.error('获取权限列表失败:', error)
    ElMessage.error('获取权限列表失败')
  } finally {
    loading.value = false
  }
}

// 构建权限树
const buildPermissionTree = (permissionList) => {
  const tree = []
  const map = {}
  
  // 创建映射
  permissionList.forEach(permission => {
    map[permission.id] = { ...permission, children: [] }
  })
  
  // 构建树结构
  permissionList.forEach(permission => {
    if (permission.parent_id && map[permission.parent_id]) {
      map[permission.parent_id].children.push(map[permission.id])
    } else {
      tree.push(map[permission.id])
    }
  })
  
  return tree
}

// 过滤权限
const filterPermissions = () => {
  let filtered = [...permissions.value]
  
  if (searchKeyword.value) {
    filtered = filterByKeyword(filtered, searchKeyword.value)
  }
  
  if (filterType.value) {
    filtered = filterByType(filtered, filterType.value)
  }
  
  filteredPermissions.value = filtered
}

// 按关键词过滤
const filterByKeyword = (perms, keyword) => {
  const result = []
  
  const searchInTree = (nodes) => {
    const matched = []
    nodes.forEach(node => {
      const nameMatch = node.permission_name.toLowerCase().includes(keyword.toLowerCase())
      const codeMatch = node.permission_code.toLowerCase().includes(keyword.toLowerCase())
      
      if (nameMatch || codeMatch) {
        matched.push({ ...node, children: node.children || [] })
      } else if (node.children && node.children.length > 0) {
        const childMatches = searchInTree(node.children)
        if (childMatches.length > 0) {
          matched.push({ ...node, children: childMatches })
        }
      }
    })
    return matched
  }
  
  return searchInTree(perms)
}

// 按类型过滤
const filterByType = (perms, type) => {
  const result = []
  
  const filterTree = (nodes) => {
    const matched = []
    nodes.forEach(node => {
      if (node.resource_type === type) {
        matched.push({ ...node, children: node.children || [] })
      } else if (node.children && node.children.length > 0) {
        const childMatches = filterTree(node.children)
        if (childMatches.length > 0) {
          matched.push({ ...node, children: childMatches })
        }
      }
    })
    return matched
  }
  
  return filterTree(perms)
}

// 重置搜索
const resetSearch = () => {
  searchKeyword.value = ''
  filterType.value = ''
  filterPermissions()
}

// 获取类型颜色
const getTypeColor = (type) => {
  const colors = {
    module: 'warning',
    menu: 'primary',
    button: 'success',
    api: 'info'
  }
  return colors[type] || 'info'
}

// 获取类型标签
const getTypeLabel = (type) => {
  const labels = {
    module: '模块',
    menu: '菜单',
    button: '按钮',
    api: 'API'
  }
  return labels[type] || type
}

// 编辑权限
const editPermission = (permission) => {
  editingPermission.value = permission
  Object.assign(permissionForm, {
    permission_name: permission.permission_name,
    permission_code: permission.permission_code,
    resource_type: permission.resource_type || 'menu',
    resource_path: permission.resource_path || '',
    parent_id: permission.parent_id || null,
    description: permission.description || '',
    sort_order: permission.sort_order || 0,
    status: permission.status
  })
  showCreateDialog.value = true
}

// 添加子权限
const addChildPermission = (permission) => {
  parentPermission.value = permission
  resetForm()
  permissionForm.parent_id = permission.id
  showCreateDialog.value = true
}

// 保存权限
const savePermission = async () => {
  if (!permissionFormRef.value) return
  
  try {
    await permissionFormRef.value.validate()
    saving.value = true
    
    let response
    if (editingPermission.value) {
      // 更新权限
      response = await permissionAPI.updatePermission(editingPermission.value.id, permissionForm)
    } else {
      // 创建权限
      response = await permissionAPI.createPermission(permissionForm)
    }
    
    ElMessage.success(editingPermission.value ? '权限更新成功' : '权限创建成功')
    showCreateDialog.value = false
    resetForm()
    loadPermissions()
  } catch (error) {
    console.error('保存权限失败:', error)
    ElMessage.error('操作失败')
  } finally {
    saving.value = false
  }
}

// 删除权限
const deletePermission = async (permission) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除权限 "${permission.permission_name}" 吗？此操作不可恢复！`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await permissionAPI.deletePermission(permission.id)
    
    ElMessage.success('删除成功')
    loadPermissions()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除权限失败:', error)
      ElMessage.error('删除失败')
    }
  }
}

// 重置权限
const resetPermissions = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要重置权限吗？这将重新初始化所有系统权限！',
      '确认重置',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    loading.value = true
    await permissionAPI.resetPermissions()
    
    ElMessage.success('权限重置成功')
    loadPermissions()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('重置权限失败:', error)
      ElMessage.error('重置失败')
    }
  } finally {
    loading.value = false
  }
}

// 重置表单
const resetForm = () => {
  editingPermission.value = null
  parentPermission.value = null
  Object.assign(permissionForm, {
    permission_name: '',
    permission_code: '',
    resource_type: 'menu',
    resource_path: '',
    parent_id: null,
    description: '',
    sort_order: 0,
    status: 1
  })
  if (permissionFormRef.value) {
    permissionFormRef.value.resetFields()
  }
}

// 组件挂载时加载数据
onMounted(() => {
  loadPermissions()
})
</script>

<style scoped>
.permission-management {
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

.header-actions {
  display: flex;
  gap: 10px;
}

.search-bar {
  margin-bottom: 20px;
  padding: 20px;
  background: #f5f7fa;
  border-radius: 4px;
}

.permission-tree {
  background: white;
  border-radius: 4px;
  overflow: hidden;
}

.permission-name {
  display: flex;
  align-items: center;
  gap: 8px;
}

.module-icon {
  color: #e6a23c;
}

.menu-icon {
  color: #409eff;
}

.button-icon {
  color: #67c23a;
}

.api-icon {
  color: #909399;
}

.resource-path {
  font-family: 'Courier New', monospace;
  background: #f5f7fa;
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 12px;
  color: #606266;
}

.no-path {
  color: #c0c4cc;
}

:deep(.el-table) {
  border: 1px solid #ebeef5;
}

:deep(.el-dialog__body) {
  padding: 20px;
}

:deep(.el-tree-select) {
  width: 100%;
}

:deep(.el-table .el-table__row .el-table__cell) {
  padding: 8px 0;
}
</style>