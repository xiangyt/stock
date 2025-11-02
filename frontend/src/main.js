import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import store from './store'

// Element Plus
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import zhCn from 'element-plus/dist/locale/zh-cn.mjs'

// 权限指令和组件
import { permissionDirective, roleDirective } from '@/utils/auth'
import PermissionButton from '@/components/PermissionButton.vue'
import PermissionWrapper from '@/components/PermissionWrapper.vue'

// 样式
import '@/styles/index.css'

const app = createApp(App)

// 注册Element Plus图标
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

// 注册权限指令
app.directive('permission', permissionDirective)
app.directive('role', roleDirective)

// 注册权限组件
app.component('PermissionButton', PermissionButton)
app.component('PermissionWrapper', PermissionWrapper)

app.use(store)
app.use(router)
app.use(ElementPlus, {
  locale: zhCn,
})

app.mount('#app')