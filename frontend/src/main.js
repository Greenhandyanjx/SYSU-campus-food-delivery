import '@/styles/home.scss';
import '@/styles/index.scss';
import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'

import App from './App.vue'
import router from './router'

const app = createApp(App)

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
