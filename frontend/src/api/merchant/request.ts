// src/api/request.ts
import axios from 'axios'

// 创建 axios 实例
const service = axios.create({
  baseURL: '/api', // 这里视情况改，比如你的后端前缀
  timeout: 5000,
})

// 请求拦截器
service.interceptors.request.use(
  (config) => {
    // 这里可以统一加 token
    const token = localStorage.getItem('token')
    if (token) config.headers.Authorization = `Bearer ${token}`
    return config
  },
  (error) => Promise.reject(error)
)

// 响应拦截器
service.interceptors.response.use(
  (response) => response,
  (error) => {
    console.error('API Error:', error)
    return Promise.reject(error)
  }
)

export default service
