import './assets/main.css'
import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'

import Login from './views/login/Login.vue'

const app = createApp(Login)

// 注册 Element Plus
app.use(ElementPlus)

// 注册图标（可选）
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

app.mount('#app')

