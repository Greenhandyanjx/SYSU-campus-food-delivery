import axios, { AxiosHeaders } from "axios";

export const http = axios.create({
  baseURL: "/api",
  timeout: 15000,
});

http.interceptors.request.use((config) => {
  const token = localStorage.getItem("token") || "";
  if (!token) return config;

  const v = token.startsWith("Bearer ") ? token : `Bearer ${token}`;

  config.headers = config.headers ?? {};
  if (config.headers instanceof AxiosHeaders) {
    config.headers.set("Authorization", v);
  } else {
    (config.headers as any).Authorization = v;
  }
  return config;
});

// 🚨 响应拦截器：检查业务code字段
http.interceptors.response.use(
  (response) => {
    console.log("📡 [HTTP响应] 数据:", {
      url: response.config.url,
      status: response.status,
      data: response.data
    });

    // 检查业务响应的code字段
    if (response.data && response.data.code === 0) {
      console.log("❌ [业务错误] 后端返回失败:", response.data.msg);
      // 创建一个错误对象，让catch块能处理
      const error = new Error(response.data.msg || "请求失败");
      (error as any).response = {
        data: response.data,
        status: response.status
      };
      throw error;
    }

    return response;
  },
  (error) => {
    console.log("❌ [HTTP错误] 请求失败:", error);
    return Promise.reject(error);
  }
);
