"""
Delivery Agent: 配送系统配置
=================================
定义配送后端的连接信息、可用商家列表等配置常量。

环境变量覆盖优先级高于默认值。
"""

import os

# 外卖后端 URL（Go/Gin 后端，端口 3000）
# 可通过环境变量 DELIVERY_BACKEND_URL 覆盖
DELIVERY_BACKEND_URL = os.environ.get(
    "DELIVERY_BACKEND_URL",
    "http://localhost:3000",
)

# 系统中可用的商家 ID 列表（Merchant.id，即店铺ID，非 base_id）
# 通过 GET /api/user/stores 返回的 id 字段获取
AVAILABLE_MERCHANT_IDS = [1, 2, 3, 4, 5, 6, 8, 9, 10, 11, 12, 13]

# 菜品推荐最大请求数量（每个商家）
DISHES_PAGE_SIZE = 50

# HTTP 请求超时时间（秒）
HTTP_TIMEOUT = 15.0
