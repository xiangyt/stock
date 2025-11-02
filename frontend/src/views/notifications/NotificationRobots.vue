<template>
  <div class="notification-robots">
    <!-- 搜索和操作区域 -->
    <el-card class="search-card">
      <el-row :gutter="20">
        <el-col :xs="24" :sm="12" :md="8">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索机器人名称"
            clearable
            @input="handleSearch"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </el-col>
        <el-col :xs="24" :sm="12" :md="6">
          <el-select v-model="selectedType" placeholder="机器人类型" clearable @change="handleFilter">
            <el-option label="全部" value="" />
            <el-option label="钉钉机器人" :value="1" />
            <el-option label="企微机器人" :value="2" />
          </el-select>
        </el-col>
        <el-col :xs="24" :sm="12" :md="6">
          <el-select v-model="selectedStatus" placeholder="状态" clearable @change="handleFilter">
            <el-option label="全部" value="" />
            <el-option label="启用" :value="true" />
            <el-option label="禁用" :value="false" />
          </el-select>
        </el-col>
        <el-col :xs="24" :sm="12" :md="4">
          <el-button type="primary" @click="showCreateDialog">
            <el-icon><Plus /></el-icon>
            新增机器人
          </el-button>
        </el-col>
      </el-row>
    </el-card>

    <!-- 机器人列表 -->
    <el-card class="table-card">
      <template #header>
        <div class="card-header">
          <span>机器人配置 (共 {{ total }} 个)</span>
          <el-button size="small" @click="refreshData" :loading="loading">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </template>

      <el-table
        :data="robotList"
        v-loading="loading"
        stripe
        style="width: 100%"
      >
        <el-table-column prop="robot_name" label="机器人名称" width="150" show-overflow-tooltip />
        
        <el-table-column prop="robot_type" label="类型" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.robot_type === 1 ? 'primary' : 'success'" size="small">
              {{ scope.row.robot_type === 1 ? '钉钉' : '企微' }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column prop="webhook_url" label="Webhook地址" min-width="200" show-overflow-tooltip>
          <template #default="scope">
            <span class="webhook-url">{{ maskWebhookUrl(scope.row.webhook_url) }}</span>
          </template>
        </el-table-column>
        
        <el-table-column prop="description" label="描述" min-width="150" show-overflow-tooltip />
        
        <el-table-column prop="is_active" label="状态" width="80">
          <template #default="scope">
            <el-switch
              v-model="scope.row.is_active"
              @change="toggleRobotStatus(scope.row)"
              :loading="scope.row.switching"
            />
          </template>
        </el-table-column>
        
        <el-table-column prop="created_at" label="创建时间" width="160">
          <template #default="scope">
            {{ formatDate(scope.row.created_at) }}
          </template>
        </el-table-column>
        
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="scope">
            <el-button size="small" type="text" @click="testRobot(scope.row)">
              <el-icon><MessageBox /></el-icon>
              测试
            </el-button>
            <el-button size="small" type="text" @click="editRobot(scope.row)">
              <el-icon><Edit /></el-icon>
              编辑
            </el-button>
            <el-button size="small" type="text" @click="deleteRobot(scope.row)" class="danger-btn">
              <el-icon><Delete /></el-icon>
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="[10, 20, 50, 100]"
        :total="total"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </el-card>

    <!-- 创建/编辑机器人对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑机器人' : '新增机器人'"
      width="600px"
      @close="resetForm"
    >
      <el-form
        ref="formRef"
        :model="robotForm"
        :rules="formRules"
        label-width="120px"
      >
        <el-form-item label="机器人名称" prop="robot_name">
          <el-input v-model="robotForm.robot_name" placeholder="请输入机器人名称" />
        </el-form-item>
        
        <el-form-item label="机器人类型" prop="robot_type">
          <el-radio-group v-model="robotForm.robot_type">
            <el-radio :label="1">钉钉机器人</el-radio>
            <el-radio :label="2">企微机器人</el-radio>
          </el-radio-group>
        </el-form-item>
        
        <el-form-item label="Webhook地址" prop="webhook_url">
          <el-input
            v-model="robotForm.webhook_url"
            type="textarea"
            :rows="3"
            placeholder="请输入Webhook地址"
          />
        </el-form-item>
        
        <el-form-item label="访问令牌" prop="access_token">
          <el-input
            v-model="robotForm.access_token"
            placeholder="请输入访问令牌（可选）"
            show-password
          />
        </el-form-item>
        
        <el-form-item label="签名密钥" prop="secret">
          <el-input
            v-model="robotForm.secret"
            placeholder="请输入签名密钥（可选）"
            show-password
          />
        </el-form-item>
        
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="robotForm.description"
            type="textarea"
            :rows="2"
            placeholder="请输入机器人描述"
          />
        </el-form-item>
        
        <el-form-item label="是否启用" prop="is_active">
          <el-switch v-model="robotForm.is_active" />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="submitForm" :loading="submitting">
            {{ isEdit ? '更新' : '创建' }}
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { ref, onMounted, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '../../utils/api'

export default {
  name: 'NotificationRobots',
  setup() {
    // 响应式数据
    const loading = ref(false)
    const submitting = ref(false)
    const robotList = ref([])
    const total = ref(0)
    const currentPage = ref(1)
    const pageSize = ref(20)
    
    // 搜索和筛选
    const searchKeyword = ref('')
    const selectedType = ref('')
    const selectedStatus = ref('')
    
    // 对话框相关
    const dialogVisible = ref(false)
    const isEdit = ref(false)
    const formRef = ref(null)
    
    // 表单数据
    const robotForm = reactive({
      id: null,
      robot_name: '',
      robot_type: 1,
      webhook_url: '',
      access_token: '',
      secret: '',
      description: '',
      is_active: true
    })
    
    // 表单验证规则
    const formRules = {
      robot_name: [
        { required: true, message: '请输入机器人名称', trigger: 'blur' },
        { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
      ],
      robot_type: [
        { required: true, message: '请选择机器人类型', trigger: 'change' }
      ],
      webhook_url: [
        { required: true, message: '请输入Webhook地址', trigger: 'blur' },
        { type: 'url', message: '请输入正确的URL格式', trigger: 'blur' }
      ]
    }
    
    // 方法
    const loadRobotList = async () => {
      loading.value = true
      try {
        const response = await api.get('/api/v1/notifications/robots', {
          params: {
            page: currentPage.value,
            size: pageSize.value,
            keyword: searchKeyword.value,
            type: selectedType.value,
            status: selectedStatus.value
          }
        })
        
        if (response.data.code === 0) {
          robotList.value = response.data.data.robots || []
          total.value = response.data.data.total || 0
        } else {
          ElMessage.error(response.data.message || '加载机器人列表失败')
        }
      } catch (error) {
        console.error('加载机器人列表失败:', error)
        ElMessage.error('加载机器人列表失败')
      } finally {
        loading.value = false
      }
    }
    
    const handleSearch = () => {
      currentPage.value = 1
      loadRobotList()
    }
    
    const handleFilter = () => {
      currentPage.value = 1
      loadRobotList()
    }
    
    const handleSizeChange = (size) => {
      pageSize.value = size
      currentPage.value = 1
      loadRobotList()
    }
    
    const handleCurrentChange = (page) => {
      currentPage.value = page
      loadRobotList()
    }
    
    const showCreateDialog = () => {
      isEdit.value = false
      dialogVisible.value = true
      resetForm()
    }
    
    const editRobot = (robot) => {
      isEdit.value = true
      dialogVisible.value = true
      Object.assign(robotForm, robot)
    }
    
    const resetForm = () => {
      if (formRef.value) {
        formRef.value.resetFields()
      }
      Object.assign(robotForm, {
        id: null,
        robot_name: '',
        robot_type: 1,
        webhook_url: '',
        access_token: '',
        secret: '',
        description: '',
        is_active: true
      })
    }
    
    const submitForm = async () => {
      if (!formRef.value) return
      
      try {
        await formRef.value.validate()
        submitting.value = true
        
        const url = isEdit.value 
          ? `/api/v1/notifications/robots/${robotForm.id}`
          : '/api/v1/notifications/robots'
        const method = isEdit.value ? 'put' : 'post'
        
        const response = await api[method](url, robotForm)
        
        if (response.data.code === 0) {
          ElMessage.success(isEdit.value ? '更新成功' : '创建成功')
          dialogVisible.value = false
          loadRobotList()
        } else {
          ElMessage.error(response.data.message || '操作失败')
        }
      } catch (error) {
        console.error('提交表单失败:', error)
        ElMessage.error('操作失败')
      } finally {
        submitting.value = false
      }
    }
    
    const toggleRobotStatus = async (robot) => {
      robot.switching = true
      try {
        const response = await api.put(`/api/v1/notifications/robots/${robot.id}`, {
          ...robot,
          is_active: robot.is_active
        })
        
        if (response.data.code === 0) {
          ElMessage.success(robot.is_active ? '已启用' : '已禁用')
        } else {
          robot.is_active = !robot.is_active // 回滚状态
          ElMessage.error(response.data.message || '状态切换失败')
        }
      } catch (error) {
        robot.is_active = !robot.is_active // 回滚状态
        console.error('切换状态失败:', error)
        ElMessage.error('状态切换失败')
      } finally {
        robot.switching = false
      }
    }
    
    const testRobot = async (robot) => {
      try {
        const response = await api.post(`/api/v1/notifications/robots/${robot.id}/test`)
        if (response.data.code === 0) {
          ElMessage.success('测试消息发送成功')
        } else {
          ElMessage.error(response.data.message || '测试失败')
        }
      } catch (error) {
        console.error('测试机器人失败:', error)
        ElMessage.error('测试失败')
      }
    }
    
    const deleteRobot = async (robot) => {
      try {
        await ElMessageBox.confirm(`确定要删除机器人 "${robot.robot_name}" 吗？`, '确认删除', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        })
        
        const response = await api.delete(`/api/v1/notifications/robots/${robot.id}`)
        if (response.data.code === 0) {
          ElMessage.success('删除成功')
          loadRobotList()
        } else {
          ElMessage.error(response.data.message || '删除失败')
        }
      } catch (error) {
        if (error !== 'cancel') {
          console.error('删除机器人失败:', error)
          ElMessage.error('删除失败')
        }
      }
    }
    
    const refreshData = () => {
      loadRobotList()
    }
    
    // 工具方法
    const maskWebhookUrl = (url) => {
      if (!url) return '--'
      if (url.length <= 50) return url
      return url.substring(0, 30) + '...' + url.substring(url.length - 20)
    }
    
    const formatDate = (dateStr) => {
      if (!dateStr) return '--'
      return new Date(dateStr).toLocaleString()
    }
    
    // 初始化
    onMounted(() => {
      loadRobotList()
    })
    
    return {
      loading,
      submitting,
      robotList,
      total,
      currentPage,
      pageSize,
      searchKeyword,
      selectedType,
      selectedStatus,
      dialogVisible,
      isEdit,
      formRef,
      robotForm,
      formRules,
      loadRobotList,
      handleSearch,
      handleFilter,
      handleSizeChange,
      handleCurrentChange,
      showCreateDialog,
      editRobot,
      resetForm,
      submitForm,
      toggleRobotStatus,
      testRobot,
      deleteRobot,
      refreshData,
      maskWebhookUrl,
      formatDate
    }
  }
}
</script>

<style lang="scss" scoped>
.notification-robots {
  .search-card {
    margin-bottom: 20px;
  }

  .table-card {
    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
    }
  }

  .webhook-url {
    font-family: monospace;
    font-size: 12px;
    color: #606266;
  }

  .danger-btn {
    color: #f56c6c;

    &:hover {
      color: #f78989;
    }
  }

  .dialog-footer {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
  }
}

@media (max-width: 768px) {
  .notification-robots {
    .card-header {
      flex-direction: column;
      align-items: flex-start;
      gap: 12px;
    }
  }
}
</style>