# 骑手端 API 接口文档

本文档描述了校园外卖系统骑手端的API接口规范，供后端开发人员参考实现。

## 基础信息

- **Base URL**: `http://localhost:3000/api`
- **认证方式**: Bearer Token (在请求头中添加 `Authorization: Bearer {token}`)
- **数据格式**: JSON
- **字符编码**: UTF-8

## 通用响应格式

所有API接口都遵循统一的响应格式：

```json
{
  "code": 1,
  "msg": "操作成功",
  "data": {
    // 具体数据内容
  }
}
```

- `code`: 状态码，1表示成功，0表示失败
- `msg`: 响应消息
- `data`: 返回的数据内容

## 状态码说明

| 状态码 | 说明 |
|--------|------|
| 1 | 成功 |
| 0 | 失败 |
| 401 | 未认证/Token过期 |
| 403 | 权限不足 |
| 404 | 资源不存在 |
| 500 | 服务器错误 |

## 1. 骑手信息相关接口

### 1.1 获取骑手个人信息

**接口地址**: `GET /rider/info`

**请求参数**: 无

**响应示例**:
```json
{
  "code": 1,
  "msg": "获取成功",
  "data": {
    "id": "rider001",
    "name": "李骑手",
    "avatar": "https://example.com/avatar.jpg",
    "phone": "13800138000",
    "rating": 4.8,
    "completedOrders": 1250,
    "isOnline": true,
    "workDays": 186,
    "totalIncome": 15680.50
  }
}
```

### 1.2 更新骑手在线状态

**接口地址**: `POST /rider/status`

**请求参数**:
```json
{
  "isOnline": true
}
```

**响应示例**:
```json
{
  "code": 1,
  "msg": "状态更新成功",
  "data": {
    "success": true
  }
}
```

### 1.3 更新骑手位置

**接口地址**: `POST /rider/location`

**请求参数**:
```json
{
  "latitude": 22.3689,
  "longitude": 113.5432,
  "address": "中山大学珠海校区榕园"
}
```

**响应示例**:
```json
{
  "code": 1,
  "msg": "位置更新成功",
  "data": {
    "success": true
  }
}
```

## 2. 订单相关接口

### 2.1 获取新订单列表（待接单）

**接口地址**: `GET /rider/orders/new`

**请求参数**: 无

**响应示例**:
```json
{
  "code": 1,
  "msg": "获取成功",
  "data": [
    {
      "id": "order001",
      "restaurant": "麦当劳",
      "pickupAddress": "珠海市香洲区唐家湾大学路1号",
      "customer": "张同学",
      "deliveryAddress": "珠海市香洲区中山大学珠海校区榕园",
      "distance": 1.2,
      "estimatedFee": 6.5,
      "estimatedTime": 20,
      "createdAt": "2024-11-17T14:30:00Z"
    }
  ]
}
```

### 2.2 骑手接单

**接口地址**: `POST /rider/orders/{orderId}/accept`

**请求参数**: 路径参数 `orderId`

**响应示例**:
```json
{
  "code": 1,
  "msg": "接单成功",
  "data": {
    "success": true,
    "pickupCode": "A123"
  }
}
```

### 2.3 获取待取货订单列表

**接口地址**: `GET /rider/orders/pickup`

**请求参数**: 无

**响应示例**:
```json
{
  "code": 1,
  "msg": "获取成功",
  "data": [
    {
      "id": "order002",
      "restaurant": "肯德基",
      "pickupAddress": "珠海市香洲区唐家湾大学路101号",
      "pickupCode": "B456",
      "shopPhone": "13788888888",
      "remainingTime": 600000
    }
  ]
}
```

### 2.4 确认取货

**接口地址**: `POST /rider/orders/{orderId}/pickup`

**请求参数**: 路径参数 `orderId`

**响应示例**:
```json
{
  "code": 1,
  "msg": "取货确认成功",
  "data": {
    "success": true
  }
}
```

### 2.5 获取配送中订单列表

**接口地址**: `GET /rider/orders/delivering`

**请求参数**: 无

**响应示例**:
```json
{
  "code": 1,
  "msg": "获取成功",
  "data": [
    {
      "id": "order003",
      "customer": "王同学",
      "customerPhone": "13666666666",
      "customerAvatar": "https://example.com/avatar.jpg",
      "deliveryAddress": "珠海市香洲区中山大学珠海校区荔园",
      "remainingTime": 1500000
    }
  ]
}
```

### 2.6 完成配送

**接口地址**: `POST /rider/orders/{orderId}/complete`

**请求参数**: 路径参数 `orderId`

**响应示例**:
```json
{
  "code": 1,
  "msg": "配送完成",
  "data": {
    "success": true,
    "actualFee": 6.5
  }
}
```

### 2.7 获取订单详情

**接口地址**: `GET /rider/orders/{orderId}`

**请求参数**: 路径参数 `orderId`

**响应示例**:
```json
{
  "code": 1,
  "msg": "获取成功",
  "data": {
    "id": "order003",
    "orderNo": "RD20241117001",
    "status": "completed",
    "items": [
      {
        "name": "巨无霸汉堡",
        "quantity": 1,
        "price": 25.00
      }
    ],
    "customerInfo": {
      "name": "王同学",
      "phone": "13666666666",
      "address": "珠海市香洲区中山大学珠海校区荔园"
    },
    "shopInfo": {
      "name": "肯德基",
      "phone": "13788888888",
      "address": "珠海市香洲区唐家湾大学路101号"
    },
    "total": 25.00,
    "deliveryFee": 6.5,
    "timeline": [
      {
        "time": "2024-11-17T14:30:00Z",
        "status": "订单创建",
        "description": "用户下单成功"
      },
      {
        "time": "2024-11-17T14:35:00Z",
        "status": "骑手接单",
        "description": "李骑手已接单"
      }
    ]
  }
}
```

### 2.8 获取历史订单列表

**接口地址**: `GET /rider/orders/history`

**请求参数**:
- `page`: 页码，默认1
- `size`: 每页数量，默认20
- `status`: 订单状态筛选（可选）
- `date`: 日期筛选（可选）

**响应示例**:
```json
{
  "code": 1,
  "msg": "获取成功",
  "data": {
    "items": [
      {
        "id": "order004",
        "orderNo": "RD20241117002",
        "status": "completed",
        "restaurant": "星巴克",
        "customer": "陈同学",
        "deliveryAddress": "珠海市香洲区中山大学珠海校区翰林",
        "distance": 1.5,
        "deliveryTime": 22,
        "fee": 7.0,
        "bonus": 1.0,
        "completedAt": "2024-11-17T15:20:00Z"
      }
    ],
    "total": 68,
    "page": 1,
    "size": 20
  }
}
```

### 2.9 获取配送路线

**接口地址**: `GET /rider/orders/{orderId}/route`

**请求参数**: 路径参数 `orderId`

**响应示例**:
```json
{
  "code": 1,
  "msg": "获取成功",
  "data": {
    "route": [
      {
        "latitude": 22.3689,
        "longitude": 113.5432
      },
      {
        "latitude": 22.3701,
        "longitude": 113.5456
      }
    ],
    "distance": 1200,
    "estimatedTime": 15
  }
}
```

## 3. 收入统计相关接口

### 3.1 获取收入统计

**接口地址**: `GET /rider/income/stats`

**请求参数**:
- `period`: 时间周期（today|week|month，可选）

**响应示例**:
```json
{
  "code": 1,
  "msg": "获取成功",
  "data": {
    "dailyIncome": 185.5,
    "weeklyIncome": 1280.0,
    "monthlyIncome": 5200.0,
    "completedOrders": 68,
    "estimatedIncome": 45.0,
    "orderIncome": 1200.0,
    "bonusIncome": 80.0,
    "chartData": [
      {
        "date": "11-11",
        "orders": 8,
        "total": 142.50
      }
    ]
  }
}
```

### 3.2 获取收入明细

**接口地址**: `GET /rider/income/history`

**请求参数**:
- `page`: 页码，默认1
- `size`: 每页数量，默认20
- `startDate`: 开始日期（可选）
- `endDate`: 结束日期（可选）

**响应示例**:
```json
{
  "code": 1,
  "msg": "获取成功",
  "data": {
    "items": [
      {
        "id": "income001",
        "orderId": "order001",
        "amount": 6.5,
        "type": "order",
        "time": "2024-11-17T14:30:00Z",
        "remark": "订单RD20241117001配送费"
      },
      {
        "id": "income002",
        "amount": 2.0,
        "type": "bonus",
        "time": "2024-11-17T13:45:00Z",
        "remark": "准时配送奖励"
      }
    ],
    "total": 150,
    "page": 1,
    "size": 20
  }
}
```

### 3.3 获取本周统计数据

**接口地址**: `GET /rider/stats/weekly`

**请求参数**: 无

**响应示例**:
```json
{
  "code": 1,
  "msg": "获取成功",
  "data": {
    "weekIncome": 1280.0,
    "weekOrders": 68,
    "onlineHours": 45,
    "avgRating": 4.8,
    "avgDeliveryTime": 18,
    "avgDistance": 1.2,
    "onTimeRate": 95,
    "positiveRate": 98,
    "acceptRate": 85,
    "deliveryEfficiency": 88,
    "customerSatisfaction": 96
  }
}
```

## 4. 钱包相关接口

### 4.1 获取钱包信息

**接口地址**: `GET /rider/wallet`

**请求参数**: 无

**响应示例**:
```json
{
  "code": 1,
  "msg": "获取成功",
  "data": {
    "balance": 2580.5,
    "frozenAmount": 120.0,
    "totalIncome": 15680.0,
    "withdrawnAmount": 13000.0
  }
}
```

### 4.2 提现申请

**接口地址**: `POST /rider/wallet/withdraw`

**请求参数**:
```json
{
  "amount": 500.0,
  "account": "alipay_138****8000"
}
```

**响应示例**:
```json
{
  "code": 1,
  "msg": "提现申请成功",
  "data": {
    "success": true,
    "withdrawId": "w001"
  }
}
```

### 4.3 获取提现记录

**接口地址**: `GET /rider/wallet/withdraw/history`

**请求参数**: 无

**响应示例**:
```json
{
  "code": 1,
  "msg": "获取成功",
  "data": [
    {
      "id": "w001",
      "amount": 500.0,
      "status": "success",
      "appliedAt": "2024-11-15T10:30:00Z",
      "processedAt": "2024-11-16T14:20:00Z",
      "account": "支付宝(138****8000)"
    },
    {
      "id": "w002",
      "amount": 300.0,
      "status": "processing",
      "appliedAt": "2024-11-16T14:20:00Z",
      "account": "微信(136****6666)"
    }
  ]
}
```

## 5. 错误处理

### 5.1 常见错误码

| 错误码 | 说明 | 处理建议 |
|--------|------|----------|
| 1001 | Token无效或过期 | 重新登录 |
| 1002 | 骑手信息不存在 | 检查骑手账号 |
| 1003 | 订单不存在 | 检查订单ID |
| 1004 | 订单状态错误 | 检查订单当前状态 |
| 1005 | 余额不足 | 检查账户余额 |
| 1006 | 提现金额限制 | 检查提现规则 |

### 5.2 错误响应示例

```json
{
  "code": 0,
  "msg": "订单不存在",
  "data": null
}
```

## 6. 开发注意事项

### 6.1 安全要求

1. 所有接口都需要进行身份验证
2. 敏感操作需要记录操作日志
3. 防止SQL注入和XSS攻击
4. 定期更新Token有效期

### 6.2 性能要求

1. 订单列表接口响应时间应在500ms以内
2. 支持分页查询，避免一次性返回大量数据
3. 图片资源使用CDN加速
4. 合理使用缓存机制

### 6.3 业务规则

1. 骑手只能接单距离在合理范围内的新订单
2. 取货码需要验证真实性
3. 提现申请需要审核，T+1到账
4. 评分系统需要防刷机制

### 6.4 数据库设计建议

```sql
-- 骑手表
CREATE TABLE riders (
  id VARCHAR(50) PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  phone VARCHAR(20) UNIQUE NOT NULL,
  avatar VARCHAR(255),
  rating DECIMAL(2,1) DEFAULT 5.0,
  completed_orders INT DEFAULT 0,
  is_online BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 订单表
CREATE TABLE orders (
  id VARCHAR(50) PRIMARY KEY,
  order_no VARCHAR(50) UNIQUE NOT NULL,
  rider_id VARCHAR(50),
  customer_id VARCHAR(50) NOT NULL,
  shop_id VARCHAR(50) NOT NULL,
  status ENUM('pending', 'accepted', 'pickup', 'delivering', 'completed', 'cancelled') DEFAULT 'pending',
  pickup_address TEXT NOT NULL,
  delivery_address TEXT NOT NULL,
  distance DECIMAL(5,2),
  delivery_fee DECIMAL(5,2) DEFAULT 0,
  total_amount DECIMAL(8,2) NOT NULL,
  pickup_code VARCHAR(20),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (rider_id) REFERENCES riders(id)
);

-- 收入记录表
CREATE TABLE income_records (
  id VARCHAR(50) PRIMARY KEY,
  rider_id VARCHAR(50) NOT NULL,
  order_id VARCHAR(50),
  amount DECIMAL(8,2) NOT NULL,
  type ENUM('order', 'bonus', 'penalty') NOT NULL,
  remark VARCHAR(255),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (rider_id) REFERENCES riders(id),
  FOREIGN KEY (order_id) REFERENCES orders(id)
);

-- 提现记录表
CREATE TABLE withdraw_records (
  id VARCHAR(50) PRIMARY KEY,
  rider_id VARCHAR(50) NOT NULL,
  amount DECIMAL(8,2) NOT NULL,
  account VARCHAR(100) NOT NULL,
  status ENUM('pending', 'processing', 'success', 'failed') DEFAULT 'pending',
  applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  processed_at TIMESTAMP NULL,
  remark VARCHAR(255),
  FOREIGN KEY (rider_id) REFERENCES riders(id)
);
```

## 7. 测试用例

### 7.1 订单流程测试

1. **接单流程**：
   - 骑手上线
   - 获取新订单列表
   - 接单
   - 确认取货
   - 完成配送

2. **异常情况**：
   - 重复接单
   - 超时未取货
   - 配送失败

### 7.2 数据统计测试

1. **收入统计**：
   - 日收入查询
   - 周收入查询
   - 月收入查询

2. **提现功能**：
   - 正常提现
   - 余额不足提现
   - 重复提现申请

## 8. 版本更新记录

- **v1.0.0** (2024-11-17): 初始版本，包含基础骑手功能
- 后续版本将根据需求迭代更新

---

如有疑问，请联系前端开发团队。