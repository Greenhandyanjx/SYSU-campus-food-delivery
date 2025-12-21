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
