// src/api/auth.ts
import request from "./merchant/request";

/**
 * Auth 接口说明（登录/注册/修改密码）
 * 下面接口为前端使用的认证相关接口，后端需按下面示例实现或返回可兼容的 JSON。
 */

/**
 * loginApi(data)
 * 功能：用户登录（支持 user/rider/merchant 根据 role 字段区分）
 * 请求：POST /login
 * 请求体示例：{ username: 'alice', password: 'xxx', role: 'user' }
 * 返回示例：{ code:1, data: { token:'jwt-token', username:'alice', role:'user' } }
 * 说明：前端会把 token 写入 localStorage，后续请求会通过拦截器加入 Authorization 头。
 */
export const loginApi = (data: {
  username: string;
  password: string;
  role: string;
  code?: string;
}) => {
  return request({
    url: "/login", // 根据实际后端路由调整
    method: "POST",
    data,
  });
};

/**
 * registerApi(data)
 * 功能：用户注册（可选 role）
 * 请求：POST /register
 * 请求体示例：{ username, password, role?: 'user' }
 * 返回示例：{ code:1, data:{ success:true, userId:'u1' } }
 */
export const registerApi = (data: { username: string; password: string; role?: string; code?: string }) => {
  return request({
    url: "/register",
    method: "POST",
    data,
  });
};

/**
 * changePassword(data)
 * 功能：修改密码（个人信息页）
 * 请求：POST /change-password
 * 请求体示例：{ username?, oldPassword, newPassword }
 * 返回示例：{ code:1, data: { success:true } }
 * 说明：后端需校验 oldPassword 并更新为 newPassword，建议返回标准错误码以便前端提示。
 */
export const changePassword = (data: { username?: string; oldPassword: string; newPassword: string }) => {
  return request({
    url: "/change-password",
    method: "POST",
    data,
  });
};
