# 骑手端API接口文档

## 概述
本文档详细描述了骑手端前端所需的所有API接口，包括已实现和待实现的接口。供后端开发人员参考实现。

## 基础信息

### 基础路径
- 基础URL: `/rider`
- 认证方式: JWT Token (Header中携带)
- 数据格式: JSON
- 时间格式: ISO 8601

### 通用响应格式
```json
{
  "code": 1,           // 1:成功, 0:失败
  "message": "success", // 响应消息
  "data": {}           // 响应数据
}
```

---

## 1. 骑手信息相关

### 1.1 获取骑手个人信息
- **接口**: `GET /rider/info`
- **描述**: 获取当前登录骑手的个人信息
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "id": "rider001",
    "name": "李骑手",
    "avatar": "https://example.com/avatar.jpg",
    "phone": "13800138000",
    "rating": 4.8,
    "completedOrders": 1250,
    "isOnline": true
  }
}
```

### 1.2 更新骑手在线状态
- **接口**: `POST /rider/status`
- **描述**: 更新骑手的在线状态
- **请求体**:
```json
{
  "isOnline": true
}
```
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "success": true
  }
}
```

---

## 2. 订单管理相关

### 2.1 获取新订单列表
- **接口**: `GET /rider/orders/new`
- **描述**: 获取待接单的新订单列表
- **响应示例**:
```json
{
  "code": 1,
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
      "createdAt": "2024-01-01T10:00:00Z"
    }
  ]
}
```

### 2.2 骑手接单
- **接口**: `POST /rider/orders/{orderId}/accept`
- **描述**: 骑手接受指定订单
- **路径参数**: `orderId` - 订单ID
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "success": true,
    "pickupCode": "A123"
  }
}
```

### 2.3 获取待取货订单列表
- **接口**: `GET /rider/orders/pickup`
- **描述**: 获取已接单但待取货的订单列表
- **响应示例**:
```json
{
  "code": 1,
  "data": [
    {
      "id": "order001",
      "restaurant": "麦当劳",
      "pickupAddress": "珠海市香洲区唐家湾大学路1号",
      "pickupCode": "A123",
      "shopPhone": "13800138001",
      "remainingTime": 300
    }
  ]
}
```

### 2.4 确认取货
- **接口**: `POST /rider/orders/{orderId}/pickup`
- **描述**: 确认已取货
- **路径参数**: `orderId` - 订单ID
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "success": true
  }
}
```

### 2.5 获取配送中订单列表
- **接口**: `GET /rider/orders/delivering`
- **描述**: 获取配送中的订单列表
- **响应示例**:
```json
{
  "code": 1,
  "data": [
    {
      "id": "order001",
      "customer": "张同学",
      "customerPhone": "13800138002",
      "customerAvatar": "https://example.com/customer.jpg",
      "deliveryAddress": "珠海市香洲区中山大学珠海校区榕园",
      "remainingTime": 600
    }
  ]
}
```

### 2.6 完成配送
- **接口**: `POST /rider/orders/{orderId}/complete`
- **描述**: 完成订单配送
- **路径参数**: `orderId` - 订单ID
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "success": true,
    "actualFee": 6.5
  }
}
```

### 2.7 获取订单详情
- **接口**: `GET /rider/orders/{orderId}`
- **描述**: 获取指定订单的详细信息
- **路径参数**: `orderId` - 订单ID
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "id": "order001",
    "items": [
      {
        "name": "巨无霸套餐",
        "quantity": 2,
        "price": 35.0
      }
    ],
    "customerInfo": {
      "name": "张同学",
      "phone": "13800138002",
      "avatar": "https://example.com/customer.jpg"
    },
    "shopInfo": {
      "name": "麦当劳",
      "phone": "13800138001",
      "address": "珠海市香洲区唐家湾大学路1号"
    },
    "total": 70.0,
    "status": "delivering",
    "timeline": [
      {
        "time": "2024-01-01T10:00:00Z",
        "status": "接单",
        "description": "骑手已接单"
      }
    ]
  }
}
```

### 2.8 获取历史订单列表
- **接口**: `GET /rider/orders/history`
- **描述**: 获取历史订单列表
- **查询参数**:
  - `page`: 页码，默认1
  - `size`: 每页数量，默认20
  - `status`: 订单状态筛选
  - `date`: 日期筛选
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "items": [
      {
        "id": "order001",
        "restaurant": "麦当劳",
        "customer": "张同学",
        "fee": 6.5,
        "status": "completed",
        "completedAt": "2024-01-01T12:00:00Z"
      }
    ],
    "total": 100
  }
}
```

---

## 3. 配送状态更新相关

### 3.1 开始配送
- **接口**: `PUT /rider/orders/{orderId}/start`
- **描述**: 开始配送（取货后出发）
- **路径参数**: `orderId` - 订单ID
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "success": true,
    "startTime": "2024-01-01T11:00:00Z"
  }
}
```

### 3.2 到达取餐点
- **接口**: `PUT /rider/orders/{orderId}/arrive-pickup`
- **描述**: 到达取餐点
- **路径参数**: `orderId` - 订单ID
- **请求体**:
```json
{
  "latitude": 22.3000,
  "longitude": 113.5000,
  "code": "A123"
}
```
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "success": true,
    "arrivedAt": "2024-01-01T10:30:00Z"
  }
}
```

### 3.3 更新配送状态
- **接口**: `PUT /rider/orders/{orderId}/status`
- **描述**: 更新配送状态和位置
- **路径参数**: `orderId` - 订单ID
- **请求体**:
```json
{
  "status": "delivering",
  "latitude": 22.3000,
  "longitude": 113.5000,
  "note": "配送中"
}
```

### 3.4 获取配送路线
- **接口**: `GET /rider/orders/{orderId}/route`
- **描述**: 获取订单配送路线
- **路径参数**: `orderId` - 订单ID
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "route": [
      {
        "lat": 22.3000,
        "lng": 113.5000
      }
    ],
    "distance": 1200,
    "estimatedTime": 15
  }
}
```

### 3.5 更新骑手位置
- **接口**: `POST /rider/location`
- **描述**: 实时更新骑手位置
- **请求体**:
```json
{
  "latitude": 22.3000,
  "longitude": 113.5000,
  "address": "珠海市香洲区中山大学珠海校区"
}
```

---

## 4. 收入统计相关

### 4.1 获取收入统计
- **接口**: `GET /rider/income/stats`
- **描述**: 获取收入统计数据
- **查询参数**:
  - `period`: 统计周期 (today|week|month)
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "dailyIncome": 185.5,
    "weeklyIncome": 1280,
    "monthlyIncome": 5200,
    "completedOrders": 68
  }
}
```

### 4.2 获取收入明细
- **接口**: `GET /rider/income/details`
- **描述**: 获取收入明细列表
- **查询参数**:
  - `page`: 页码，默认1
  - `size`: 每页数量，默认20
  - `type`: 收入类型筛选
  - `startDate`: 开始日期
  - `endDate`: 结束日期
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "items": [
      {
        "id": "income001",
        "orderId": "order001",
        "amount": 6.5,
        "type": "delivery_fee",
        "time": "2024-01-01T12:00:00Z",
        "remark": "配送收入"
      }
    ],
    "total": 200
  }
}
```

### 4.3 获取收入历史
- **接口**: `GET /rider/income/history`
- **描述**: 获取收入历史记录
- **查询参数**:
  - `page`: 页码，默认1
  - `size`: 每页数量，默认20
  - `startDate`: 开始日期
  - `endDate`: 结束日期

---

## 5. 钱包相关

### 5.1 获取钱包信息
- **接口**: `GET /rider/wallet`
- **描述**: 获取骑手钱包信息
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "balance": 2580.5,
    "frozenAmount": 120,
    "totalIncome": 15680
  }
}
```

### 5.2 提现申请
- **接口**: `POST /rider/wallet/withdraw`
- **描述**: 申请提现
- **请求体**:
```json
{
  "amount": 500,
  "account": "6222021234567890"
}
```
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "success": true,
    "withdrawId": "w1"
  }
}
```

### 5.3 获取提现记录
- **接口**: `GET /rider/wallet/withdraw/history`
- **描述**: 获取提现记录列表
- **响应示例**:
```json
{
  "code": 1,
  "data": [
    {
      "id": "w1",
      "amount": 500,
      "status": "processing",
      "appliedAt": "2024-01-01T10:00:00Z",
      "processedAt": null
    }
  ]
}
```

---

## 6. 工作统计相关

### 6.1 获取工作数据统计
- **接口**: `GET /rider/stats/work`
- **描述**: 获取工作数据统计
- **查询参数**:
  - `period`: 统计周期 (today|week|month)
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "totalOrders": 68,
    "totalIncome": 1280.5,
    "completionRate": 95.8
  }
}
```

### 6.2 获取本周统计数据
- **接口**: `GET /rider/stats/weekly`
- **描述**: 获取本周统计数据
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "weekIncome": 1280,
    "weekOrders": 68,
    "onlineHours": 45,
    "avgRating": 4.8
  }
}
```

### 6.3 获取月统计数据
- **接口**: `GET /rider/stats/monthly`
- **描述**: 获取月统计数据
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "monthOrders": 280,
    "monthIncome": 5200,
    "onlineDays": 25
  }
}
```

---

## 7. 配送记录相关

### 7.1 获取配送记录
- **接口**: `GET /rider/delivery/records`
- **描述**: 获取配送记录列表
- **查询参数**:
  - `page`: 页码，默认1
  - `size`: 每页数量，默认20
  - `status`: 状态筛选
  - `startDate`: 开始日期
  - `endDate`: 结束日期
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "items": [
      {
        "id": "delivery001",
        "orderNo": "D20240101001",
        "distance": 1.2,
        "duration": 25,
        "completedAt": "2024-01-01T12:00:00Z"
      }
    ],
    "total": 150
  }
}
```

---

## 8. 评价相关

### 8.1 获取用户评价
- **接口**: `GET /rider/reviews`
- **描述**: 获取用户评价列表
- **查询参数**:
  - `page`: 页码，默认1
  - `size`: 每页数量，默认20
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "items": [
      {
        "id": "review001",
        "orderId": "order001",
        "rating": 5,
        "comment": "配送很快，服务态度好",
        "createdAt": "2024-01-01T12:00:00Z"
      }
    ],
    "avgRating": 4.8
  }
}
```

---

## 9. 排行榜相关

### 9.1 获取排行榜数据
- **接口**: `GET /rider/ranking/{type}`
- **描述**: 获取排行榜数据
- **路径参数**: `type` - 排行榜类型 (income|orders|rating|efficiency)
- **响应示例**:
```json
{
  "code": 1,
  "data": [
    {
      "rank": 1,
      "name": "王骑手",
      "avatar": "https://example.com/avatar1.jpg",
      "value": 5800,
      "isSelf": false
    },
    {
      "rank": 5,
      "name": "李骑手",
      "avatar": "https://example.com/avatar2.jpg",
      "value": 5200,
      "isSelf": true
    }
  ]
}
```

---

## 10. 通知相关

### 10.1 获取通知列表
- **接口**: `GET /rider/notifications`
- **描述**: 获取通知列表
- **查询参数**:
  - `page`: 页码，默认1
  - `size`: 每页数量，默认20
  - `read`: 是否已读筛选
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "items": [
      {
        "id": "notify001",
        "title": "新订单提醒",
        "content": "您有一个新的订单待接单",
        "type": "order",
        "isRead": false,
        "createdAt": "2024-01-01T10:00:00Z"
      }
    ],
    "unreadCount": 3
  }
}
```

### 10.2 标记通知已读
- **接口**: `PUT /rider/notifications/{id}/read`
- **描述**: 标记指定通知为已读
- **路径参数**: `id` - 通知ID
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "success": true
  }
}
```

### 10.3 获取系统消息
- **接口**: `GET /rider/messages/system`
- **描述**: 获取系统消息列表
- **查询参数**:
  - `page`: 页码，默认1
  - `size`: 每页数量，默认20
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "items": [
      {
        "id": "msg001",
        "title": "系统升级通知",
        "content": "系统将于今晚进行升级维护",
        "type": "system",
        "publishedAt": "2024-01-01T09:00:00Z"
      }
    ]
  }
}
```

---

## 11. 认证相关

### 11.1 获取骑手认证信息
- **接口**: `GET /rider/verification`
- **描述**: 获取骑手实名认证信息
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "id": "verify001",
    "realName": "张三",
    "idCard": "440106199001011234",
    "phone": "13800138000",
    "status": "approved",
    "submitTime": "2024-01-01T10:00:00Z"
  }
}
```

### 11.2 提交实名认证
- **接口**: `POST /rider/verification`
- **描述**: 提交实名认证申请
- **请求体**:
```json
{
  "realName": "张三",
  "idCard": "440106199001011234",
  "idCardFront": "https://example.com/id-front.jpg",
  "idCardBack": "https://example.com/id-back.jpg",
  "healthCert": "https://example.com/health-cert.jpg"
}
```
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "success": true,
    "verificationId": "verify001"
  }
}
```

---

## 12. 设置相关

### 12.1 获取工作设置
- **接口**: `GET /rider/settings/work`
- **描述**: 获取工作设置信息
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "autoAccept": false,
    "deliveryRange": 5000,
    "workTime": {
      "start": "08:00",
      "end": "22:00"
    },
    "restTime": {
      "enabled": true,
      "start": "12:00",
      "end": "14:00"
    },
    "maxOrders": 10
  }
}
```

### 12.2 更新工作设置
- **接口**: `PUT /rider/settings/work`
- **描述**: 更新工作设置
- **请求体**:
```json
{
  "autoAccept": true,
  "deliveryRange": 8000,
  "workTime": {
    "start": "09:00",
    "end": "21:00"
  },
  "restTime": {
    "enabled": false,
    "start": "12:00",
    "end": "14:00"
  },
  "maxOrders": 8
}
```

### 12.3 获取账户设置
- **接口**: `GET /rider/settings/account`
- **描述**: 获取账户设置信息
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "phone": "13800138000",
    "email": "rider@example.com",
    "wechat": "rider_wechat",
    "alipay": "rider_alipay",
    "bankCard": "6222021234567890"
  }
}
```

### 12.4 更新账户设置
- **接口**: `PUT /rider/settings/account`
- **描述**: 更新账户设置
- **请求体**:
```json
{
  "phone": "13800138001",
  "email": "new_rider@example.com",
  "wechat": "new_rider_wechat",
  "alipay": "new_rider_alipay",
  "bankCard": "6222021234567891"
}
```

### 12.5 获取通知设置
- **接口**: `GET /rider/settings/notification`
- **描述**: 获取通知设置信息
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "orderNotification": true,
    "systemNotification": true,
    "soundEnabled": true,
    "vibrationEnabled": false
  }
}
```

### 12.6 更新通知设置
- **接口**: `PUT /rider/settings/notification`
- **描述**: 更新通知设置
- **请求体**:
```json
{
  "orderNotification": false,
  "systemNotification": true,
  "soundEnabled": false,
  "vibrationEnabled": true
}
```

---

## 13. 热力图相关

### 13.1 获取配送热力图数据
- **接口**: `GET /rider/heatmap`
- **描述**: 获取配送热力图数据
- **查询参数**:
  - `start`: 开始日期
  - `end`: 结束日期
- **响应示例**:
```json
{
  "code": 1,
  "data": [
    {
      "date": "2024-01-01",
      "areas": [
        {
          "lat": 22.3000,
          "lng": 113.5000,
          "count": 25
        }
      ]
    }
  ]
}
```

---

## 14. 异常报告相关

### 14.1 异常情况报告
- **接口**: `POST /rider/orders/{orderId}/issue`
- **描述**: 报告订单配送过程中的异常情况
- **路径参数**: `orderId` - 订单ID
- **请求体**:
```json
{
  "type": "customer_not_contact",
  "description": "无法联系到客户",
  "images": [
    "https://example.com/issue1.jpg",
    "https://example.com/issue2.jpg"
  ],
  "timestamp": 1704110400
}
```
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "success": true,
    "issueId": "issue001"
  }
}
```

---

## 15. 个人资料相关 (新增)

### 15.1 获取个人资料详情
- **接口**: `GET /rider/profile`
- **描述**: 获取骑手个人详细信息
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "id": "rider001",
    "name": "李骑手",
    "avatar": "https://example.com/avatar.jpg",
    "phone": "13800138000",
    "gender": "male",
    "birthday": "1995-06-15",
    "emergencyContact": "张三",
    "emergencyPhone": "13900139000",
    "address": "珠海市香洲区中山大学珠海校区",
    "vehicle": "electric_bike",
    "vehicleNumber": "粤B12345",
    "registrationDate": "2023-01-01",
    "rating": 4.8,
    "completedOrders": 1250,
    "totalIncome": 58000
  }
}
```

### 15.2 更新个人资料
- **接口**: `PUT /rider/profile`
- **描述**: 更新骑手个人基本信息
- **请求体**:
```json
{
  "name": "李骑手",
  "avatar": "https://example.com/new-avatar.jpg",
  "gender": "male",
  "birthday": "1995-06-15",
  "emergencyContact": "张三",
  "emergencyPhone": "13900139000",
  "address": "珠海市香洲区中山大学珠海校区",
  "vehicle": "electric_bike",
  "vehicleNumber": "粤B12345"
}
```
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "success": true
  }
}
```

### 15.3 上传头像
- **接口**: `POST /rider/profile/avatar`
- **描述**: 上传骑手头像
- **请求体**: `multipart/form-data`
  - `avatar`: 图片文件
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "avatarUrl": "https://example.com/avatar.jpg"
  }
}
```

---

## 16. 账户安全相关 (新增)

### 16.1 获取安全设置
- **接口**: `GET /rider/profile/security`
- **描述**: 获取账户安全设置信息
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "phone": "13800138000",
    "phoneVerified": true,
    "email": "rider@example.com",
    "emailVerified": false,
    "loginPassword": true,
    "paymentPassword": true,
    "loginHistory": [
      {
        "time": "2024-01-01T10:00:00Z",
        "device": "iPhone 15",
        "location": "珠海市香洲区"
      }
    ],
    "securityScore": 85
  }
}
```

### 16.2 修改登录密码
- **接口**: `PUT /rider/profile/password`
- **描述**: 修改登录密码
- **请求体**:
```json
{
  "oldPassword": "oldpass123",
  "newPassword": "newpass456"
}
```
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "success": true
  }
}
```

### 16.3 设置支付密码
- **接口**: `POST /rider/profile/payment-password`
- **描述**: 设置支付密码
- **请求体**:
```json
{
  "paymentPassword": "123456"
}
```

### 16.4 更改手机号
- **接口**: `PUT /rider/profile/phone`
- **描述**: 更改绑定的手机号
- **请求体**:
```json
{
  "phone": "13900139000",
  "verificationCode": "123456"
}
```

### 16.5 绑定邮箱
- **接口**: `POST /rider/profile/email`
- **描述**: 绑定邮箱地址
- **请求体**:
```json
{
  "email": "rider@example.com",
  "verificationCode": "123456"
}
```

### 16.6 获取登录记录
- **接口**: `GET /rider/profile/login-history`
- **描述**: 获取登录历史记录
- **查询参数**:
  - `page`: 页码，默认1
  - `size`: 每页数量，默认20
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "items": [
      {
        "time": "2024-01-01T10:00:00Z",
        "device": "iPhone 15",
        "location": "珠海市香洲区",
        "ip": "192.168.1.100"
      }
    ],
    "total": 50
  }
}
```

---

## 17. 收款设置相关 (新增)

### 17.1 获取收款设置
- **接口**: `GET /rider/profile/payment-settings`
- **描述**: 获取收款账户设置
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "defaultMethod": "alipay",
    "accounts": [
      {
        "id": "acc001",
        "type": "alipay",
        "account": "rider@alipay.com",
        "name": "李骑手",
        "isDefault": true,
        "isVerified": true
      },
      {
        "id": "acc002",
        "type": "wechat",
        "account": "rider_wechat",
        "name": "李骑手",
        "isDefault": false,
        "isVerified": true
      },
      {
        "id": "acc003",
        "type": "bank",
        "account": "6222021234567890",
        "name": "李骑手",
        "bankName": "工商银行",
        "isDefault": false,
        "isVerified": false
      }
    ]
  }
}
```

### 17.2 添加收款账户
- **接口**: `POST /rider/profile/payment-accounts`
- **描述**: 添加新的收款账户
- **请求体**:
```json
{
  "type": "bank",
  "account": "6222021234567890",
  "name": "李骑手",
  "bankName": "工商银行"
}
```

### 17.3 更新收款账户
- **接口**: `PUT /rider/profile/payment-accounts/{accountId}`
- **描述**: 更新收款账户信息
- **路径参数**: `accountId` - 账户ID
- **请求体**:
```json
{
  "name": "李骑手",
  "isDefault": true
}
```

### 17.4 删除收款账户
- **接口**: `DELETE /rider/profile/payment-accounts/{accountId}`
- **描述**: 删除收款账户
- **路径参数**: `accountId` - 账户ID

### 17.5 设置默认收款方式
- **接口**: `PUT /rider/profile/payment-settings/default`
- **描述**: 设置默认收款方式
- **请求体**:
```json
{
  "accountId": "acc001"
}
```

---

## 18. 工作偏好设置相关 (新增)

### 18.1 获取工作偏好设置
- **接口**: `GET /rider/profile/work-preferences`
- **描述**: 获取工作偏好和配送设置
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "workMode": "full_time",
    "autoAccept": true,
    "deliveryRange": 5000,
    "preferredAreas": ["唐家湾", "香洲区"],
    "avoidAreas": ["斗门区"],
    "workTime": {
      "start": "08:00",
      "end": "22:00"
    },
    "restTime": {
      "enabled": true,
      "start": "12:00",
      "end": "14:00"
    },
    "deliveryMode": "balanced",
    "maxOrders": 8,
    "weekendWork": true,
    "holidayWork": false,
    "nightDelivery": true
  }
}
```

### 18.2 更新工作偏好设置
- **接口**: `PUT /rider/profile/work-preferences`
- **描述**: 更新工作偏好设置
- **请求体**:
```json
{
  "workMode": "part_time",
  "autoAccept": false,
  "deliveryRange": 8000,
  "preferredAreas": ["唐家湾", "香洲区", "高新区"],
  "avoidAreas": [],
  "workTime": {
    "start": "09:00",
    "end": "21:00"
  },
  "restTime": {
    "enabled": false,
    "start": "12:00",
    "end": "14:00"
  },
  "deliveryMode": "efficient",
  "maxOrders": 6,
  "weekendWork": true,
  "holidayWork": true,
  "nightDelivery": false
}
```

### 18.3 获取常用地址
- **接口**: `GET /rider/profile/favorite-locations`
- **描述**: 获取常用地址列表
- **响应示例**:
```json
{
  "code": 1,
  "data": [
    {
      "id": "loc001",
      "name": "家",
      "address": "珠海市香洲区中山大学珠海榕园",
      "latitude": 22.3000,
      "longitude": 113.5000
    },
    {
      "id": "loc002",
      "name": "唐家湾商圈",
      "address": "珠海市香洲区唐家湾",
      "latitude": 22.2800,
      "longitude": 113.5200
    }
  ]
}
```

### 18.4 添加常用地址
- **接口**: `POST /rider/profile/favorite-locations`
- **描述**: 添加常用地址
- **请求体**:
```json
{
  "name": "公司",
  "address": "珠海市香洲区高新区科技创新海岸",
  "latitude": 22.3200,
  "longitude": 113.4800
}
```

---

## 19. 消息通知设置相关 (新增)

### 19.1 获取通知设置
- **接口**: `GET /rider/profile/notification-settings`
- **描述**: 获取消息通知设置
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "orderNotifications": {
      "newOrder": true,
      "orderAccepted": true,
      "orderPicked": true,
      "orderDelivered": true,
      "orderCancelled": true
    },
    "systemNotifications": {
      "systemMessages": true,
      "promotionMessages": false,
      "updateNotifications": true
    },
    "soundSettings": {
      "enabled": true,
      "volume": 80,
      "newOrderSound": "default",
      "systemSound": "default"
    },
    "vibrationSettings": {
      "enabled": true,
      "newOrderVibration": true,
      "systemVibration": false
    },
    "doNotDisturb": {
      "enabled": true,
      "startTime": "22:00",
      "endTime": "08:00",
      "allowEmergency": true
    }
  }
}
```

### 19.2 更新通知设置
- **接口**: `PUT /rider/profile/notification-settings`
- **描述**: 更新消息通知设置
- **请求体**:
```json
{
  "orderNotifications": {
    "newOrder": true,
    "orderAccepted": false,
    "orderPicked": true,
    "orderDelivered": true,
    "orderCancelled": true
  },
  "systemNotifications": {
    "systemMessages": true,
    "promotionMessages": true,
    "updateNotifications": false
  },
  "soundSettings": {
    "enabled": false,
    "volume": 50,
    "newOrderSound": "gentle",
    "systemSound": "default"
  },
  "vibrationSettings": {
    "enabled": true,
    "newOrderVibration": false,
    "systemVibration": true
  },
  "doNotDisturb": {
    "enabled": false,
    "startTime": "22:00",
    "endTime": "08:00",
    "allowEmergency": true
  }
}
```

### 19.3 测试通知
- **接口**: `POST /rider/profile/test-notification`
- **描述**: 发送测试通知
- **请求体**:
```json
{
  "type": "new_order"
}
```

---

## 20. 地图设置相关 (新增)

### 20.1 获取地图设置
- **接口**: `GET /rider/profile/map-settings`
- **描述**: 获取地图导航设置
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "defaultProvider": "baidu",
    "navigationSettings": {
      "avoidTolls": false,
      "avoidHighways": false,
      "voiceNavigation": true,
      "voiceType": "female",
      "language": "zh-CN"
    },
    "displaySettings": {
      "trafficLayer": true,
      "satelliteLayer": false,
      "deliveryZones": true,
      "restaurants": true
    },
    "offlineMaps": [
      {
        "id": "offline001",
        "city": "珠海市",
        "size": 125.6,
        "downloadedAt": "2024-01-01T10:00:00Z",
        "version": "1.2.0"
      }
    ]
  }
}
```

### 20.2 更新地图设置
- **接口**: `PUT /rider/profile/map-settings`
- **描述**: 更新地图导航设置
- **请求体**:
```json
{
  "defaultProvider": "amap",
  "navigationSettings": {
    "avoidTolls": true,
    "avoidHighways": true,
    "voiceNavigation": true,
    "voiceType": "male",
    "language": "zh-CN"
  },
  "displaySettings": {
    "trafficLayer": false,
    "satelliteLayer": true,
    "deliveryZones": true,
    "restaurants": false
  }
}
```

### 20.3 下载离线地图
- **接口**: `POST /rider/profile/offline-maps/download`
- **描述**: 下载离线地图
- **请求体**:
```json
{
  "city": "广州市"
}
```

### 20.4 删除离线地图
- **接口**: `DELETE /rider/profile/offline-maps/{mapId}`
- **描述**: 删除离线地图
- **路径参数**: `mapId` - 地图ID

### 20.5 获取可用离线地图
- **接口**: `GET /rider/profile/offline-maps/available`
- **描述**: 获取可下载的离线地图列表
- **响应示例**:
```json
{
  "code": 1,
  "data": [
    {
      "city": "深圳市",
      "size": 156.8,
      "version": "1.3.0",
      "downloaded": false
    },
    {
      "city": "广州市",
      "size": 189.2,
      "version": "1.3.0",
      "downloaded": true
    }
  ]
}
```

---

## 21. 帮助中心相关 (新增)

### 21.1 获取帮助文档分类
- **接口**: `GET /rider/help/categories`
- **描述**: 获取帮助文档分类列表
- **响应示例**:
```json
{
  "code": 1,
  "data": [
    {
      "id": "cat001",
      "name": "订单配送",
      "icon": "delivery",
      "articleCount": 25
    },
    {
      "id": "cat002",
      "name": "账户设置",
      "icon": "account",
      "articleCount": 15
    }
  ]
}
```

### 21.2 获取帮助文档列表
- **接口**: `GET /rider/help/articles`
- **描述**: 获取帮助文档列表
- **查询参数**:
  - `category`: 分类ID
  - `page`: 页码，默认1
  - `size`: 每页数量，默认20
  - `search`: 搜索关键词
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "items": [
      {
        "id": "article001",
        "title": "如何接单",
        "category": "订单配送",
        "summary": "详细介绍接单流程和注意事项",
        "views": 1250,
        "helpful": 89,
        "createdAt": "2024-01-01T10:00:00Z"
      }
    ],
    "total": 50
  }
}
```

### 21.3 获取帮助文档详情
- **接口**: `GET /rider/help/articles/{articleId}`
- **描述**: 获取帮助文档详情
- **路径参数**: `articleId` - 文档ID
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "id": "article001",
    "title": "如何接单",
    "content": "详细的接单说明内容...",
    "category": "订单配送",
    "views": 1250,
    "helpful": 89,
    "createdAt": "2024-01-01T10:00:00Z",
    "relatedArticles": [
      {
        "id": "article002",
        "title": "订单取消规则"
      }
    ]
  }
}
```

### 21.4 获取常见问题
- **接口**: `GET /rider/help/faq`
- **描述**: 获取常见问题列表
- **响应示例**:
```json
{
  "code": 1,
  "data": [
    {
      "id": "faq001",
      "question": "如何申请提现？",
      "answer": "进入我的钱包页面，点击提现按钮...",
      "category": "收入结算",
      "views": 523
    },
    {
      "id": "faq002",
      "question": "订单超时怎么办？",
      "answer": "如遇特殊情况即将超时，请及时联系客服...",
      "category": "订单配送",
      "views": 412
    }
  ]
}
```

### 21.5 获取帮助视频
- **接口**: `GET /rider/help/videos`
- **描述**: 获取帮助视频列表
- **响应示例**:
```json
{
  "code": 1,
  "data": [
    {
      "id": "video001",
      "title": "新手骑手入门指南",
      "duration": 180,
      "thumbnail": "https://example.com/thumb1.jpg",
      "videoUrl": "https://example.com/video1.mp4",
      "views": 3580,
      "category": "新手指南"
    }
  ]
}
```

### 21.6 搜索帮助内容
- **接口**: `GET /rider/help/search`
- **描述**: 搜索帮助内容
- **查询参数**:
  - `keyword`: 搜索关键词
  - `type`: 搜索类型 (article|faq|video|all)
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "articles": [
      {
        "id": "article001",
        "title": "如何接单",
        "highlight": "详细说明<em>如何</em>正确接单"
      }
    ],
    "faqs": [
      {
        "id": "faq001",
        "question": "如何申请提现？",
        "highlight": "说明<em>如何</em>申请提现的流程"
      }
    ]
  }
}
```

---

## 22. 意见反馈相关 (新增)

### 22.1 提交意见反馈
- **接口**: `POST /rider/feedback`
- **描述**: 提交意见反馈
- **请求体**:
```json
{
  "type": "bug",
  "title": "页面加载异常",
  "content": "在订单页面出现加载缓慢的情况",
  "contact": "rider@example.com",
  "images": [
    "https://example.com/feedback1.jpg",
    "https://example.com/feedback2.jpg"
  ],
  "location": {
    "latitude": 22.3000,
    "longitude": 113.5000,
    "address": "珠海市香洲区中山大学珠海校区"
  },
  "deviceInfo": {
    "platform": "iOS",
    "version": "17.0",
    "appVersion": "2.1.0"
  }
}
```
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "success": true,
    "feedbackId": "fb001",
    "ticketNumber": "TK20240101001"
  }
}
```

### 22.2 获取反馈历史
- **接口**: `GET /rider/feedback/history`
- **描述**: 获取反馈历史记录
- **查询参数**:
  - `page`: 页码，默认1
  - `size`: 每页数量，默认20
  - `status`: 状态筛选
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "items": [
      {
        "id": "fb001",
        "type": "bug",
        "title": "页面加载异常",
        "status": "processing",
        "createdAt": "2024-01-01T10:00:00Z",
        "repliedAt": null,
        "ticketNumber": "TK20240101001"
      }
    ],
    "total": 15
  }
}
```

### 22.3 获取反馈详情
- **接口**: `GET /rider/feedback/{feedbackId}`
- **描述**: 获取反馈详情
- **路径参数**: `feedbackId` - 反馈ID
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "id": "fb001",
    "type": "bug",
    "title": "页面加载异常",
    "content": "在订单页面出现加载缓慢的情况",
    "status": "resolved",
    "createdAt": "2024-01-01T10:00:00Z",
    "repliedAt": "2024-01-01T15:00:00Z",
    "reply": "问题已修复，请更新到最新版本",
    "images": [
      "https://example.com/feedback1.jpg"
    ]
  }
}
```

### 22.4 获取反馈分类
- **接口**: `GET /rider/feedback/categories`
- **描述**: 获取反馈分类列表
- **响应示例**:
```json
{
  "code": 1,
  "data": [
    {
      "id": "bug",
      "name": "功能异常",
      "description": "系统功能出现异常或错误"
    },
    {
      "id": "feature",
      "name": "功能建议",
      "description": "希望增加的新功能或改进建议"
    },
    {
      "id": "ui",
      "name": "界面问题",
      "description": "界面显示或操作体验问题"
    }
  ]
}
```

---

## 23. 联系客服相关 (新增)

### 23.1 获取联系信息
- **接口**: `GET /rider/support/contact`
- **描述**: 获取客服联系信息
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "phone": "400-123-4567",
    "onlineService": {
      "available": true,
      "waitTime": 30,
      "url": "https://chat.example.com/rider"
    },
    "wechat": {
      "qrCode": "https://example.com/wechat-qr.jpg",
      "accountId": "rider_support"
    },
    "email": "rider-support@example.com",
    "workHours": {
      "phone": "周一至周日 9:00-21:00",
      "online": "7x24小时",
      "wechat": "周一至周五 9:00-18:00"
    }
  }
}
```

### 23.2 获取常见问题分类
- **接口**: `GET /rider/support/common-issues`
- **描述**: 获取常见问题快速入口
- **响应示例**:
```json
{
  "code": 1,
  "data": [
    {
      "id": "withdraw",
      "title": "如何申请提现？",
      "icon": "money",
      "viewCount": 523,
      "route": "/rider/help/withdraw"
    },
    {
      "id": "timeout",
      "title": "订单超时怎么处理？",
      "icon": "timeout",
      "viewCount": 412,
      "route": "/rider/help/timeout"
    }
  ]
}
```

### 23.3 提交客服工单
- **接口**: `POST /rider/support/tickets`
- **描述**: 提交客服工单
- **请求体**:
```json
{
  "type": "order_issue",
  "orderId": "order001",
  "title": "客户联系不上",
  "description": "配送地址无法联系到客户，电话无人接听",
  "priority": "high",
  "contact": "13800138000"
}
```
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "success": true,
    "ticketId": "tk001"
  }
}
```

### 23.4 获取工单历史
- **接口**: `GET /rider/support/tickets`
- **描述**: 获取客服工单历史
- **查询参数**:
  - `page`: 页码，默认1
  - `size`: 每页数量，默认20
  - `status`: 状态筛选
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "items": [
      {
        "id": "tk001",
        "type": "order_issue",
        "title": "客户联系不上",
        "status": "resolved",
        "createdAt": "2024-01-01T10:00:00Z",
        "resolvedAt": "2024-01-01T12:00:00Z"
      }
    ],
    "total": 8
  }
}
```

---

## 24. 系统公告相关 (新增)

### 24.1 获取系统公告列表
- **接口**: `GET /rider/announcements`
- **描述**: 获取系统公告列表
- **查询参数**:
  - `page`: 页码，默认1
  - `size`: 每页数量，默认20
  - `type`: 公告类型
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "items": [
      {
        "id": "announce001",
        "title": "系统升级维护通知",
        "type": "system",
        "priority": "high",
        "summary": "系统将于今晚进行升级维护",
        "content": "详细内容...",
        "publishTime": "2024-01-01T18:00:00Z",
        "isRead": false,
        "isTop": true
      }
    ],
    "unreadCount": 3
  }
}
```

### 24.2 获取公告详情
- **接口**: `GET /rider/announcements/{id}`
- **描述**: 获取公告详情
- **路径参数**: `id` - 公告ID
- **响应示例**:
```json
{
  "code": 1,
  "data": {
    "id": "announce001",
    "title": "系统升级维护通知",
    "type": "system",
    "priority": "high",
    "content": "详细的公告内容...",
    "publishTime": "2024-01-01T18:00:00Z",
    "author": "系统管理员",
    "attachments": [
      {
        "name": "升级说明.pdf",
        "url": "https://example.com/upgrade.pdf"
      }
    ]
  }
}
```

### 24.3 标记公告已读
- **接口**: `PUT /rider/announcements/{id}/read`
- **描述**: 标记公告为已读
- **路径参数**: `id` - 公告ID

---

## 待实现接口说明

以下接口在当前后端可能还未完全实现，需要后端开发人员优先实现：

### 1. 实时位置更新系统
- **优先级**: 高
- **说明**: 用于实时追踪骑手位置，需要支持WebSocket长连接
- **相关接口**: `/rider/location`, `/rider/orders/{orderId}/route`

### 2. 消息推送系统
- **优先级**: 高
- **说明**: 用于推送新订单、系统通知等消息
- **相关接口**: `/rider/notifications`, `/rider/messages/system`

### 3. 热力图数据统计
- **优先级**: 中
- **说明**: 需要基于历史配送数据生成热力图
- **相关接口**: `/rider/heatmap`

### 4. 排行榜系统
- **优先级**: 中
- **说明**: 需要实时计算各种排行榜数据
- **相关接口**: `/rider/ranking/{type}`

### 5. 异常处理流程
- **优先级**: 高
- **说明**: 需要完善的异常情况处理和客服介入机制
- **相关接口**: `/rider/orders/{orderId}/issue`

---

## 错误码说明

| 错误码 | 说明 | 处理建议 |
|--------|------|----------|
| 0 | 请求失败 | 检查请求参数和网络连接 |
| 1001 | 认证失败 | 重新登录 |
| 1002 | 权限不足 | 检查用户权限 |
| 2001 | 订单不存在 | 检查订单ID |
| 2002 | 订单状态错误 | 检查当前订单状态 |
| 3001 | 余额不足 | 提示用户充值 |
| 3002 | 提现失败 | 检查提现账户信息 |
| 4001 | 参数错误 | 检查请求参数格式 |
| 5001 | 系统内部错误 | 稍后重试或联系技术支持 |

---

## 注意事项

1. **安全性**: 所有接口都需要进行用户身份验证
2. **数据一致性**: 订单状态变更需要保证事务一致性
3. **性能优化**: 列表查询接口建议使用分页
4. **错误处理**: 前端需要正确处理各种错误情况
5. **实时性**: 订单和位置相关接口需要保证实时性
6. **移动端优化**: 接口响应数据需要针对移动端优化，减少数据传输量

---

## 更新日志

- **2024-01-01**: 初始版本，包含所有基础接口文档
- **2024-01-02**: 增加热力图、排行榜等高级功能接口
- **2024-01-03**: 完善异常处理和通知相关接口