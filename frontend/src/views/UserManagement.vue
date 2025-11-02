<template>
  <div class="user-management">
    <div class="page-header">
      <h2>用户管理</h2>
      <el-button type="primary" @click="createUser">
        <el-icon><Plus /></el-icon>
        新增用户
      </el-button>
    </div>

    <!-- 搜索栏 -->
    <div class="search-bar">
      <el-row :gutter="20">
        <el-col :span="8">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索用户名、邮箱或姓名"
            clearable
            @keyup.enter="loadUsers"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </el-col>
        <el-col :span="4">
          <el-button type="primary" @click="loadUsers">搜索</el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-col>
      </el-row>
    </div>

    <!-- 用户列表 -->
    <div class="user-table">
      <el-table 
        :data="users" 
        v-loading="loading"
        stripe
        style="width: 100%; min-width: 1200px"
      >
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="username" label="用户名" width="120" />
        <el-table-column prop="email" label="邮箱" width="200" />
        <el-table-column prop="real_name" label="真实姓名" width="120" />
        <el-table-column prop="phone" label="手机号" width="130" />
        <el-table-column label="角色" width="120">
          <template #default="scope">
            <el-tag v-if="scope.row.role">{{ scope.row.role.role_name }}</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="scope">
            <el-tag 
              :type="getUserStatusType(scope.row.status)"
            >
              {{ getUserStatusText(scope.row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="last_login_at" label="最后登录" width="160">
          <template #default="scope">
            {{ formatDateTime(scope.row.last_login_at) }}
          </template>
        </el-table-column>
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
              @click="editUser(scope.row)"
            >
              编辑
            </el-button>
            <el-button 
              :type="scope.row.status === 1 ? 'warning' : 'success'"
              size="small" 
              @click="toggleUserStatus(scope.row)"
            >
              {{ scope.row.status === 1 ? '禁用' : '启用' }}
            </el-button>
            <el-button 
              type="danger" 
              size="small" 
              @click="deleteUser(scope.row)"
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

          @size-change="loadUsers"
          @current-change="loadUsers"
        />
      </div>
    </div>

    <!-- 创建/编辑用户对话框 -->
    <el-dialog
      v-model="showCreateDialog"
      :title="editingUser ? '编辑用户' : '新增用户'"
      width="600px"
      @close="handleDialogClose"
    >
      <el-form
        ref="userFormRef"
        :model="userForm"
        :rules="userFormRules"
        label-width="100px"
      >
        <el-form-item label="用户名" prop="username">
          <el-input 
            v-model="userForm.username" 
            :disabled="editingUser"
            placeholder="请输入用户名"
          />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input 
            v-model="userForm.email" 
            placeholder="请输入邮箱地址"
          />
        </el-form-item>
        <el-form-item label="密码" prop="password" v-if="!editingUser">
          <el-input 
            v-model="userForm.password" 
            type="password"
            placeholder="请输入密码"
            show-password
          />
        </el-form-item>
        <el-form-item label="真实姓名" prop="real_name">
          <el-input 
            v-model="userForm.real_name" 
            placeholder="请输入真实姓名"
          />
        </el-form-item>
        <el-form-item label="手机号" prop="phone">
          <el-input 
            v-model="userForm.phone" 
            placeholder="请输入手机号"
          />
        </el-form-item>
        <el-form-item label="角色" prop="role_id">
          <el-select 
            v-model="userForm.role_id" 
            placeholder="请选择角色"
            style="width: 100%"
          >
            <el-option
              v-for="role in roles"
              :key="role.id"
              :label="role.role_name"
              :value="role.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="状态" prop="status" v-if="editingUser">
          <el-select v-model="userForm.status" placeholder="请选择状态">
            <el-option label="正常" :value="1" />
            <el-option label="禁用" :value="0" />
            <el-option label="锁定" :value="2" />
          </el-select>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="saveUser" :loading="saving">
          {{ editingUser ? '更新' : '创建' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search } from '@element-plus/icons-vue'
import { userAPI, roleAPI } from '@/utils/api'

// 响应式数据
const loading = ref(false)
const saving = ref(false)
const users = ref([])
const roles = ref([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(10)
const searchKeyword = ref('')
const showCreateDialog = ref(false)
const editingUser = ref(null)

// 用户表单
const userForm = reactive({
  username: '',
  email: '',
  password: '',
  real_name: '',
  phone: '',
  role_id: null,
  status: 1
})

// 表单验证规则
const userFormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 50, message: '用户名长度在 3 到 50 个字符', trigger: 'blur' }
  ],
  email: [
    { required: true, message: '请输入邮箱地址', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱地址', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于 6 个字符', trigger: 'blur' }
  ]
}

const userFormRef = ref()

// 加载用户列表
const loadUsers = async () => {
  loading.value = true
  try {
    const response = await userAPI.getUserList({
      page: currentPage.value,
      page_size: pageSize.value,
      keyword: searchKeyword.value
    })
    
    // 处理嵌套的数据结构
    const data = response.data || response
    users.value = data.users || []
    total.value = data.total || 0
  } catch (error) {
    console.error('获取用户列表失败:', error)
    ElMessage.error('获取用户列表失败')
  } finally {
    loading.value = false
  }
}

// 加载角色列表
const loadRoles = async () => {
  try {
    const response = await roleAPI.getRoleList()
    // 处理嵌套的数据结构
    const data = response.data || response
    // 只显示启用状态的角色
    roles.value = (data.roles || []).filter(role => role.status === 1)
  } catch (error) {
    console.error('获取角色列表失败:', error)
  }
}

// 重置搜索
const resetSearch = () => {
  searchKeyword.value = ''
  currentPage.value = 1
  loadUsers()
}

// 创建用户
const createUser = async () => {
  editingUser.value = null
  resetForm()
  // 重新获取最新的角色列表（只包含启用的角色）
  await loadRoles()
  showCreateDialog.value = true
}

// 编辑用户
const editUser = async (user) => {
  editingUser.value = user
  
  // 重新获取最新的角色列表（只包含启用的角色）
  await loadRoles()
  
  // 检查用户当前角色是否在启用的角色列表中
  const isRoleEnabled = roles.value.some(role => role.id === user.role_id)
  
  Object.assign(userForm, {
    username: user.username,
    email: user.email,
    password: '',
    real_name: user.real_name || '',
    phone: user.phone || '',
    role_id: isRoleEnabled ? user.role_id : null, // 如果角色被禁用，设为null
    status: user.status
  })
  
  showCreateDialog.value = true
}

// 保存用户
const saveUser = async () => {
  if (!userFormRef.value) return
  
  try {
    await userFormRef.value.validate()
    saving.value = true
    
    let response
    if (editingUser.value) {
      // 更新用户
      response = await userAPI.updateUser(editingUser.value.id, userForm)
    } else {
      // 创建用户
      response = await userAPI.createUser(userForm)
    }
    
    ElMessage.success(editingUser.value ? '用户更新成功' : '用户创建成功')
    showCreateDialog.value = false
    editingUser.value = null
    resetForm()
    loadUsers()
  } catch (error) {
    console.error('保存用户失败:', error)
    ElMessage.error('操作失败')
  } finally {
    saving.value = false
  }
}

// 切换用户状态
const toggleUserStatus = async (user) => {
  const action = user.status === 1 ? '禁用' : '启用'
  try {
    await ElMessageBox.confirm(
      `确定要${action}用户 "${user.username}" 吗？`,
      '确认操作',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    const newStatus = user.status === 1 ? 0 : 1
    const response = await userAPI.updateUser(user.id, { status: newStatus })
    
    ElMessage.success(`${action}成功`)
    loadUsers()
  } catch (error) {
    if (error !== 'cancel') {
      console.error(`${action}用户失败:`, error)
      ElMessage.error(`${action}失败`)
    }
  }
}

// 删除用户
const deleteUser = async (user) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除用户 "${user.username}" 吗？此操作不可恢复！`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    const response = await userAPI.deleteUser(user.id)
    
    ElMessage.success('删除成功')
    loadUsers()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除用户失败:', error)
      ElMessage.error('删除失败')
    }
  }
}

// 重置表单
const resetForm = () => {
  // 重置表单数据
  Object.assign(userForm, {
    username: '',
    email: '',
    password: '',
    real_name: '',
    phone: '',
    role_id: null,
    status: 1
  })
  
  // 清除表单验证状态
  if (userFormRef.value) {
    userFormRef.value.resetFields()
    userFormRef.value.clearValidate()
  }
}

// 处理对话框关闭
const handleDialogClose = () => {
  editingUser.value = null
  resetForm()
}

// 获取用户状态类型
const getUserStatusType = (status) => {
  switch (status) {
    case 1: return 'success'
    case 0: return 'danger'
    case 2: return 'warning'
    default: return 'info'
  }
}

// 获取用户状态文本
const getUserStatusText = (status) => {
  switch (status) {
    case 1: return '正常'
    case 0: return '禁用'
    case 2: return '锁定'
    default: return '未知'
  }
}

// 格式化日期时间
const formatDateTime = (dateTime) => {
  if (!dateTime) return '-'
  return new Date(dateTime).toLocaleString('zh-CN')
}

// 组件挂载时加载数据
onMounted(() => {
  loadUsers()
  loadRoles()
})
</script>

<style scoped>
.user-management {
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

.user-table {
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
</style>