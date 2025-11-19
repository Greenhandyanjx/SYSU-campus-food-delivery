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
const service = axios.create({
  baseURL: "http://localhost:3000/api", // 这里视情况改，比如你的后端前缀
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
