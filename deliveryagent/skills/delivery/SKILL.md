---
name: delivery
description: "外卖配送助手：查询商家、菜品、订单状态、配送进度、客服售后。Use when: user asks about food ordering, stores, menus, orders, delivery status, or customer service."
version: "1.0.0"
categories: ["外卖", "生活服务", "客服"]
metadata:
  emoji: "🍽️"
---

# Delivery Skill

提供外卖平台相关的全部能力，包括商家浏览、菜品查询、订单追踪和客服售后。

## When to Use

✅ **USE this skill when:**

- "有哪些商家？"
- "我要点外卖"
- "推荐几个菜"
- "查一下我的订单"
- "我的订单到哪了？"
- "配送怎么还没到？"
- "我想退款"
- "麦当劳有什么吃的？"

## When NOT to Use

❌ **DON'T use this skill when:**

- 需要天气信息 → 用 weather skill
- 需要日期计算 → 用 datetime skill

## Abilities

### get_stores
获取所有商家列表，包括名称、评分、月销量。

**参数：** 无

### search_store
根据关键词搜索商家。

**参数：**
- `query` (string, 必填): 搜索关键词

### get_dishes
获取指定商家的菜品列表。

**参数：**
- `merchant_id` (integer, 必填): 商家 ID

### recommend_dish
根据口味偏好推荐菜品。

**参数：**
- `preference` (string, 必填): 口味偏好
- `max_items` (integer, 可选): 最大推荐数量

### query_order
查询订单状态和详情。

**参数：**
- `order_id` (string, 必填): 订单 ID

### check_delivery_status
查询配送进度和骑手位置。

**参数：**
- `order_id` (string, 必填): 订单 ID

### customer_service
生成客服回复，处理投诉和售后。

**参数：**
- `user_query` (string, 必填): 用户问题
- `order_id` (string, 可选): 订单 ID

### get_user_default_info
获取当前用户的默认收货信息（姓名、电话、地址），下单前调用可自动填充。

**参数：**
- `user_id` (string, 必填): 用户 ID
