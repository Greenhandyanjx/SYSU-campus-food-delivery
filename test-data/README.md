# 骑手端测试数据说明

本文档包含了为骑手端功能测试设计的完整测试数据集。

## 文件说明

### 1. rider_test_data.sql
SQL格式的数据库插入脚本，包含：
- **骑手基本信息** (riders表) - 5个测试骑手，包含实名认证信息
- **骑手详细档案** (rider_profiles表) - 包含评分、在线状态、位置等
- **骑手钱包信息** (rider_wallets表) - 余额、冻结金额、总收入
- **订单数据** (orders表) - 包含各种状态的订单：
  - 待接单订单 (status=1)
  - 已接单待取货 (status=2)
  - 取货中 (status=3)
  - 配送中 (status=4)
  - 已完成 (status=5)
- **收入记录** (rider_income_records表) - 配送费、奖励等

### 2. api_test_data.json
JSON格式的API测试数据，包含：
- 各种API接口的响应格式示例
- 测试场景数据
- 可直接用于Postman等工具测试

## 如何使用

### 数据库数据导入
```bash
# 导入SQL数据到数据库
mysql -u username -p database_name < test-data/rider_test_data.sql
```

### API测试示例
使用以下骑手账号进行测试：

#### 测试骑手1：李骑手 (在线)
- **手机号**: 13800138001
- **初始状态**: 在线
- **钱包余额**: 2580.50元
- **评分**: 4.8

#### 测试骑手2：王骑手 (离线)
- **手机号**: 13800138002
- **初始状态**: 离线
- **钱包余额**: 1825.30元
- **评分**: 4.6

## 测试流程

### 1. 骑手登录测试
```
POST /api/rider/login
{
  "phone": "13800138001",
  "password": "123456"
}
```

### 2. 获取骑手信息测试
```
GET /api/rider/info
Headers: Authorization: Bearer {token}
```

### 3. 查看新订单测试
```
GET /api/rider/orders/new
Headers: Authorization: Bearer {token}
```
预期返回3个待接单订单

### 4. 接单测试
```
POST /api/rider/orders/order001/accept
Headers: Authorization: Bearer {token}
```

### 5. 确认取货测试
```
POST /api/rider/orders/order004/pickup
Headers: Authorization: Bearer {token}
```

### 6. 完成配送测试
```
POST /api/rider/orders/order006/complete
Headers: Authorization: Bearer {token}
```

## 订单状态说明

- **status = 1**: 待接单 - 新创建的订单，等待骑手接单
- **status = 2**: 已接单 - 骑手已接单，待取货
- **status = 3**: 取货中 - 骑手已到达商家，正在取货
- **status = 4**: 配送中 - 骑手已取货，正在配送
- **status = 5**: 已完成 - 订单已配送完成

## 时间线字段说明

- **accepted_at**: 接单时间
- **pickup_at**: 取货时间
- **deliver_at**: 开始配送时间
- **finish_at**: 完成配送时间

## 地理位置

测试数据基于中山大学珠海校区的实际位置：
- **纬度范围**: 22.36-22.38
- **经度范围**: 113.53-113.55

## 收入类型

- **order**: 订单配送费
- **bonus**: 奖励收入
- **adjustment**: 调整款项

## 注意事项

1. 测试数据包含敏感信息，仅用于测试环境
2. 手机号和身份证号为测试专用，非真实数据
3. 位置信息基于中山大学珠海校区，可根据实际情况调整
4. 建议在测试前备份数据库
5. 测试完成后可清理测试数据

## 常见问题

### Q: 如何模拟订单超时？
A: 修改orders表中的expectedtime字段为过去时间

### Q: 如何测试提现功能？
A: 使用rider_wallets表中rider_id=1的骑手，余额充足

### Q: 如何测试不同评分场景？
A: 修改rider_profiles表中的rating字段值

### Q: 如何模拟骑手位置变化？
A: 更新rider_profiles表中的latitude和longitude字段