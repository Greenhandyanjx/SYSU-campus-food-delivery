
此文档基于前端代码中 `src/api` 的调用约定整理，供后端参考。文档按功能模块罗列：认证、商家统计、订单、菜品/套餐等。

通用约定
- Base URL: 按前端 axios 实例，默认 `http://localhost:3000/api`。
- 授权: 前端在 `request` 拦截器中会读取 `localStorage.getItem('token')` 并设置请求头 `Authorization: Bearer <token>`。
- 响应格式: 前端现有代码通常以业务 `code` 字段判断成功（成功常为 `1`），因此推荐返回结构：
  {
    "code": "1", // 或数字 1 表示成功
    "msg": "描述信息",
    "data": ...
  }

错误处理建议：
- 对于鉴权失败可返回 HTTP 401 或返回业务 code（如 code: '401'）。
- 导出/下载接口返回二进制流（Content-Type: application/octet-stream 或 对应类型），并设置 Content-Disposition: attachment; filename="..."

示例：
```
POST /login
Content-Type: application/json
{
  "username": "alice",
  "password": "123456",
  "role": "merchant",
  "code": "0"
}

返回（示例）:
{
  "code": "1",
  "msg": "登录成功",
  "data": { "token": "xxx.yyy.zzz", "user": { "username": "alice", "role": "merchant" } }
}
```

--------------------------------------------------------------------------------

1) 认证相关

1.1 登录
- URL: POST /login
- 请求体 (JSON): { username, password, role, code }
- 返回: { code, msg, data: { token?, user? } }
- 前端行为: 若 code === '1' 则视为成功，登录后前端会把 username 存入 localStorage（键: 'username'），并期望后端返回 token 用于后续鉴权。

1.2 注册
- URL: POST /register
- 请求体 (JSON): { username, password, role?, code? }
- 返回: { code, msg, data }

1.3 修改密码（前端已创建调用）
- URL: POST /change-password
- 请求体 (JSON): { username?, oldPassword, newPassword }
  - username 可选：若后端通过 token 能识别用户，则无需传 username；否则需要。
- 返回: { code, msg }
  - 成功示例: { code: '1', msg: '密码修改成功' }
  - 失败示例: { code: '0', msg: '原密码不正确' }

--------------------------------------------------------------------------------

2) 商家端 - 概览 / 统计

注：以下接口在 `src/api/index.ts` 中有引用。

2.1 Overview & business data
- GET /merchant/businessData
- GET /merchant/orderData
- GET /merchant/overviewDishes
- GET /merchant/setMealStatistics

2.2 统计详情
- GET /merchant/statistics/overview?begin=YYYY-MM-DD&end=YYYY-MM-DD
- GET /merchant/statistics/turnover?begin=...&end=...
- GET /merchant/statistics/user?begin=...&end=...
- GET /merchant/statistics/order?begin=...&end=...
- GET /merchant/statistics/top?limit=10
- GET /merchant/statistics/export  // 返回文件流（前端设置 responseType: 'blob'）

示例返回（分页/列表）:
```
{
  "code": "1",
  "msg": "ok",
  "data": {
    "records": [...],
    "total": 123
  }
}
```

--------------------------------------------------------------------------------

3) 订单管理（`src/api/merchant/order.ts`）

3.1 查询 / 列表
- GET /merchant/orders/status?status=2&page=1&pageSize=10  // 查询不同状态订单（status）
- GET /merchant/orders/page?page=1&pageSize=10&status=2     // 分页查询

3.2 详情
- GET /merchant/order/detail?orderId=xxx

3.3 订单操作（POST 请求，body 为 JSON）
- POST /merchant/order/accept    { id }
- POST /merchant/order/reject    { id, rejectionReason }
- POST /merchant/order/cancel    { id, cancelReason }
- POST /merchant/order/delivery  { id, status }
- POST /merchant/order/complete  { id, status }

返回示例: { code: '1', msg: '成功', data: ... }

--------------------------------------------------------------------------------

4) 菜品与套餐（`src/api/merchant/dish.ts`, `src/api/merchant/setMeal.ts`）

4.1 菜品
- GET  /merchant/dishes/page        // params 支持 pagination、categoryId、name 等
- GET  /merchant/dish/list         // 根据条件查询菜品列表
- GET  /merchant/dish/query?id=xxx // 查询单个菜品
- GET  /merchant/dish/categories   // 获取菜品分类
- POST /merchant/dish/add          // 新增菜品（body: 菜品信息 JSON）
- POST /merchant/dish/edit         // 编辑菜品（body: 菜品信息 JSON）
- POST /merchant/dish/delete       // 删除（body: { id })
- POST /merchant/dish/status       // 上/下架（body: { id, status })

4.2 套餐（Setmeal）
- GET  /merchant/meal/page         // 套餐分页
- GET  /merchant/meal/query?id=xxx // 查询单个套餐
- POST /merchant/meal/add          // 新增套餐
- POST /merchant/meal/edit         // 编辑套餐
- POST /merchant/meal/delete       // 删除套餐 (body: { id })
- POST /merchant/meal/status       // 启用/禁用套餐

4.3 通用下载
- GET /merchant/common/download?key=...  // 前端调用为文件下载

--------------------------------------------------------------------------------

5) 类别（category）
- GET /merchant/category/list  // 查询类别（可能按 type 分类）

--------------------------------------------------------------------------------

6) 返回格式与错误码建议（再次强调）
- 统一返回格式 (建议):
  {
    code: '1' | '0' | '401' | string,
    msg: '提示信息',
    data: any
  }
- 业务成功: code === '1'
- 业务失败: code !== '1'，并返回可读 msg

--------------------------------------------------------------------------------

7) 安全与校验建议
- 对敏感接口（登录、注册、修改密码）建议做密码强度校验与限频；
- 所有请求应通过 HTTPS；
- 导出/下载接口需设置合适的 Content-Type 与 Content-Disposition。

--------------------------------------------------------------------------------

8) 后端接口实现注意事项（给后端的实用提示）
- 登录接口建议返回 token，并返回用户基本信息（username / role / avatarUrl 可选）；
- 修改密码：如果采用 token 鉴权，后端可以允许用户仅传 newPassword（通过 token 确认身份），或同时要求 oldPassword 做二次确认；
- 分页接口返回统一字段：records（数组） + total（总数），可附加 page/pageSize 等信息。

