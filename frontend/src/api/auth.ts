// src/api/auth.ts
import request from "./merchant/request";

export const loginApi = (data: {
  username: string;
  password: string;
  role: string;
  code: string;
}) => {
  return request({
    url: "/login", // 根据实际后端路由调整
    method: "POST",
    data,
  });
};
