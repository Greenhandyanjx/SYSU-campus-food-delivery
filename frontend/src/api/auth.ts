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

export const registerApi = (data: { username: string; password: string; role?: string; code?: string }) => {
  return request({
    url: "/register",
    method: "POST",
    data,
  });
};

/**
 * 修改/保存密码（个人信息页面的保存按钮对应接口）
 * data: { username?: string, oldPassword: string, newPassword: string }
 * 注意：后端应校验 oldPassword 并更新为 newPassword。
 */
export const changePassword = (data: { username?: string; oldPassword: string; newPassword: string }) => {
  return request({
    url: "/change-password",
    method: "POST",
    data,
  });
}
