import '@/styles/home.scss';
import '@/styles/index.scss';
import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
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

  // 改进错误处理，区分不同类型的错误
  if (err.name === 'AxiosError') {
    console.error('Axios网络请求错误:', {
      message: err.message,
      code: err.code,
      config: err.config,
      response: err.response
    })

    // 网络错误不显示全局错误，让组件内部处理
    return false
  }

  // 其他Vue错误仍然显示
  console.error('Vue运行时错误:', err)
}

app.mount('#app')
