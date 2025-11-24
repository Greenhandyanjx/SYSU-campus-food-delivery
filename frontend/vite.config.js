import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    vueDevTools(),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    },
  },
  css: {
    preprocessorOptions: {
      scss: {
        additionalData: `
        @use "@/styles/_variables.scss" as *;
        @use "@/styles/_mixins.scss" as *;
      `
      }
    }
  }
  ,
  // 开发服务器代理：将以 /api 开头的请求代理到后端服务（开发时避免 CORS）
  server: {
    proxy: {
      '/api': {
        // 后端实际运行端口（你的 curl 返回数据来自 3000），将代理指向 3000
        target: 'http://localhost:3000',
        changeOrigin: true,
        secure: false,
        // 保留 /api 前缀不做重写，后端路由以 /api 开头定义
      }
    }
  }
})
