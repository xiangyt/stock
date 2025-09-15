import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'

import App from './App.vue'
import Home from './views/Home.vue'
import StockList from './views/StockList.vue'
import StockDetail from './views/StockDetail.vue'
import Analysis from './views/Analysis.vue'

const routes = [
  { path: '/', component: Home },
  { path: '/stocks', component: StockList },
  { path: '/stock/:code', component: StockDetail, props: true },
  { path: '/analysis', component: Analysis }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

const app = createApp(App)

// 注册所有图标
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

app.use(ElementPlus)
app.use(router)
app.mount('#app')