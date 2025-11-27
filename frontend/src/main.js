import '@/styles/home.scss';
import '@/styles/index.scss';
import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import 'element-plus/dist/index.css'

import App from './App.vue'
import router from './router'
import Vant from 'vant'
import 'vant/lib/index.css'
const app = createApp(App)
app.use(Vant)
app.use(ElementPlus)
app.use(router)

// 注册 Element Plus 图标（可选）
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}
app.config.errorHandler = (err, vm, info) => {
  console.error('Vue 全局错误捕获:', err, info, vm)
}

app.mount('#app')
