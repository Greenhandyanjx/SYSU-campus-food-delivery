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

# 系统中可用的商家 ID 列表（来自后端 /api/user/stores 返回的 id 字段）
# storeId=1=夯肉先生, 2=竹林上禾便当, 3=coco都可, 4=麦当劳,
# 5=港都热炒, 6=中山大学学一食堂, 7=金拱门, 8=瑞幸咖啡,
# 9=一点点, 10=兰州拉面, 11=杨国福麻辣烫, 12=啫啫煲仔饭, 13=沙县小吃
AVAILABLE_MERCHANT_IDS = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13]

# 菜品推荐最大请求数量（每个商家）
DISHES_PAGE_SIZE = 50

# HTTP 请求超时时间（秒）
HTTP_TIMEOUT = 15.0
