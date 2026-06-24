# Delivery Agent — 外卖配送 AI 助手

## 概述

Delivery Agent 是基于 JD-Agent 框架构建的校园外卖配送 AI 助手，
与 SYSU Campus Food Delivery 系统集成，提供智能化的外卖服务体验。

## 架构

```
用户 → Streamlit UI(8501) → FastAPI(8000) → AgentOrchestrator → AgentLoop(ReAct) → Tool → Go Backend(3000)
```

## 配送工具

| 工具 | 说明 | 后端 API |
|------|------|----------|
| `get_stores` | 获取所有商家列表 | `GET /api/user/stores` |
| `search_store` | 搜索商家 | `GET /api/store/query` |
| `get_dishes` | 获取商家菜品 | `GET /api/store/dishes` |
| `recommend_dish` | 根据偏好推荐菜品 | 遍历所有商家后匹配 |
| `query_order` | 查询订单状态 | `GET /api/order/status` |
| `check_delivery_status` | 查询配送进度 | `GET /api/order/status` |
| `customer_service` | 智能客服（模板） | 无外部 API |

## 启动方式

```bash
# 1. 复制环境变量
cp .env.example .env   # 填写 DEEPSEEK_API_KEY

# 2. 启动 FastAPI 后端（必须）
uvicorn api.app:app --host 127.0.0.1 --port 8000 --reload

# 3. 启动 Streamlit 前端（可选）
streamlit run main.py --server.port 8501

# 或一键启动
start.bat
```

## 配置

- `delivery_config.py`: 配送后端连接配置
- `config/agent.yml`: Agent 行为配置
- `prompts/main_prompt.txt`: 系统提示词
- `skills/delivery/`: 配送技能定义
