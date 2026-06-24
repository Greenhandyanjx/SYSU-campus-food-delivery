# Delivery Agent — 外卖配送 AI 助手

## 概述

Delivery Agent 是基于 [JD-Agent](../jd_agent) 框架构建的外卖配送 AI 助手，
与 [SYSU Campus Food Delivery](https://github.com/SYSU-campus-food-delivery) 系统集成，
提供智能化的外卖配送服务体验。

## 配送专用工具

本 Agent 新增了四个配送业务工具，对接校园外卖配送后端（Go 后端）：

### 1. `query_order` — 查询订单状态和详情

- **描述**: 根据订单 ID 获取订单的完整信息
- **API**: `GET /api/merchant/order/detail?id={order_id}`（需认证）
- **参数**: `order_id` (必填), `user_id` (可选)
- **返回**: 订单状态、菜品列表、金额、收货地址、骑手信息等

### 2. `recommend_dish` — 根据用户偏好推荐菜品

- **描述**: 遍历所有商家，根据关键词匹配菜品名称/描述/标签
- **API**: `GET /api/store/dishes?merchant_id={mid}&page=1&pageSize=50`（无需认证）
- **参数**: `preference` (必填，如 "辣"/"清淡"/"甜"/"粤菜"), `max_items` (可选，默认5)
- **返回**: 相关度排序的推荐菜品列表

### 3. `check_delivery_status` — 查询配送进度和骑手位置

- **描述**: 获取配送状态、预计送达时间、骑手实时位置
- **API**: `GET /api/order/status?id={order_id}`（无需认证）
- **参数**: `order_id` (必填)
- **返回**: 配送状态、进度条、骑手信息等

### 4. `customer_service` — 生成智能客服回复

- **描述**: 模板化客服回复生成器，处理常见场景
- **场景覆盖**: 订单丢失、配送延迟、送错商品、退款、质量问题、价格问题
- **特点**: 不调用外部 API，本地智能匹配 + 共情回复

## 配置

### 环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `DELIVERY_BACKEND_URL` | `http://localhost:8080` | 配送后端地址 |
| `AVAILABLE_MERCHANT_IDS` | `[1..12]` | 可遍历的商家 ID 列表 |

### 配置文件

`delivery_config.py` 包含所有配送相关的配置常量。

## 前端集成

配送 Agent 在前端以浮动聊天面板形式集成：

- **`AgentPanel.vue`** — 浮动聊天面板组件（FAB 按钮 + 滑出式面板）
- **`agent.ts`** — API 封装（支持 SSE 流式响应）
- **路由** — 可通过 `/user/agent` 访问，或通过首页浮动按钮唤起

## 运行

```bash
# 1. 启动配送后端（Go）
cd SYSU-campus-food-delivery/backend
go run main.go

# 2. 启动 Agent API
cd delivery-agent
copy .env.example .env    # 编辑 .env 填入 API Key
python -m api.app

# 3. 启动前端
cd SYSU-campus-food-delivery/frontend
npm run dev
```

## 技术栈

- **Agent 框架**: Python (FastAPI + JD-Agent)
- **配送后端**: Go (Gin)
- **前端**: Vue 3 (Composition API + TypeScript)
- **通信**: REST API + Server-Sent Events (SSE)
