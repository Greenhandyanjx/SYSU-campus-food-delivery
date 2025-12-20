// src/api/request.ts
import axios from "axios";

/**
 * axios 实例说明
 * - baseURL: 请替换为后端正确的前缀（示例为 http://localhost:3000/api）
 * - 请求拦截器：将从 localStorage 读取 token 并注入 Authorization 头
 * - 响应拦截器：打印后端错误信息并将错误抛出，方便上层处理
 *
 * 后端开发注意：建议统一返回标准 JSON 结构，例如 { code: 1, data: ..., msg: '' }
 */

// 创建 axios 实例
// 使用相对路径 `/api`，在开发模式下 Vite 的 proxy 会将其转发到后端服务，避免硬编码主机/端口
const service = axios.create({
  baseURL: '/api',
  timeout: 5000,
});

// 请求拦截器
service.interceptors.request.use(
  (config) => {
    // 这里可以统一加 token
    const token = localStorage.getItem("token");
    if (token) {
      // token may already include the 'Bearer ' prefix (backend returns it that way).
      // Avoid duplicating 'Bearer ' (which would produce 'Bearer Bearer ...').
      config.headers.Authorization = token.startsWith('Bearer ') ? token : `Bearer ${token}`;
    }
    // 临时埋点：记录 /merchant/orders/page 请求的调用栈与时间，便于定位重复触发源
    try {
      const url = String(config.url || '')
      if (url.includes('/merchant/orders/page')) {
        const info = {
          tag: 'ORDERS_PAGE_REQUEST',
          time: new Date().toISOString(),
          url: config.url,
          params: config.params,
          method: config.method,
        }
        // Print stack to help find which code path triggered the request
        // eslint-disable-next-line no-console
        console.warn('ORDERS_PAGE_REQUEST - request', info, new Error().stack)
      }
    } catch (e) {
      // ignore
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// 响应拦截器
service.interceptors.response.use(
  (response) => response,
  (error) => {
    // 打印后端返回的具体错误，便于调试
    console.error("API Error Details:", error.response?.data);
    return Promise.reject(error);
  }
);

export default service;
