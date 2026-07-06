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

# 系统中可用的商家 base_id 列表（来自数据库 merchants 表）
AVAILABLE_MERCHANT_IDS = [1, 4, 9, 10, 42, 68, 69, 70, 71, 72, 73, 74, 75]

# 菜品推荐最大请求数量（每个商家）
DISHES_PAGE_SIZE = 50

# HTTP 请求超时时间（秒）
HTTP_TIMEOUT = 15.0
