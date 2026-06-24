"""
Delivery Agent 技能：外卖配送助手
=================================
提供商家浏览、菜品查询、订单追踪、客服售后等外卖平台能力。
所有工具对接真实后端 API（Go/Gin, port 3000）。
"""

from agent.skills.base import Skill, SkillAbility
from agent.tools.delivery_tools import (
    GetStoresTool,
    SearchStoreTool,
    GetDishesTool,
    RecommendDishTool,
    QueryOrderTool,
    CheckDeliveryStatusTool,
    CustomerServiceTool,
    GetUserDefaultInfoTool,
    GetUserProfileTool,
    GetUserCartTool,
    AddToCartTool,
    GetUserOrdersTool,
    CancelOrderTool,
    PayOrderTool,
)


class DeliverySkill(Skill):
    """
    外卖配送技能：一站式外卖助手能力包。

    包含 7 个 Tool：
      - get_stores              — 获取商家列表
      - search_store            — 搜索商家
      - get_dishes              — 获取商家菜品
      - recommend_dish          — 推荐菜品
      - query_order             — 查询订单
      - check_delivery_status   — 配送追踪
      - customer_service        — 智能客服
    """

    @property
    def name(self) -> str:
        return "delivery"

    @property
    def description(self) -> str:
        return "外卖配送助手：查询商家、菜品、订单状态、配送进度、客服售后等"

    @property
    def version(self) -> str:
        return "1.0.0"

    @property
    def categories(self) -> list[str]:
        return ["外卖", "生活服务", "客服"]

    @property
    def abilities(self) -> list[SkillAbility]:
        return [
            SkillAbility(
                name="get_stores",
                description="获取外卖平台上所有商家列表，包括名称、评分、月销量",
                keywords=["商家", "店铺", "商店", "饭店", "有什么吃的", "点外卖", "store"],
                examples=["有哪些商家？", "看看有什么店铺", "我想点外卖"],
                tool=GetStoresTool(),
            ),
            SkillAbility(
                name="search_store",
                description="根据关键词搜索商家",
                keywords=["搜索", "找", "麦当劳", "奶茶", "烧烤", "search"],
                examples=["搜索麦当劳", "找奶茶店", "附近有什么烧烤店"],
                tool=SearchStoreTool(),
            ),
            SkillAbility(
                name="get_dishes",
                description="获取指定商家的菜品列表",
                keywords=["菜单", "菜品", "有什么菜", "吃什么", "menu", "dish"],
                examples=["麦当劳有什么吃的", "看看菜单", "这家店有什么菜"],
                tool=GetDishesTool(),
            ),
            SkillAbility(
                name="recommend_dish",
                description="根据用户的口味偏好推荐菜品",
                keywords=["推荐", "想吃", "口味", "推荐菜", "推荐吃的", "recommend"],
                examples=["推荐几个菜", "我想吃辣的", "有什么清淡的推荐"],
                tool=RecommendDishTool(),
            ),
            SkillAbility(
                name="query_order",
                description="查询订单状态和详情",
                keywords=["订单", "我的订单", "查订单", "order", "订单状态", "订单详情"],
                examples=["查一下我的订单", "订单状态", "我的订单到哪了"],
                tool=QueryOrderTool(),
            ),
            SkillAbility(
                name="check_delivery_status",
                description="查询配送进度和骑手位置",
                keywords=["配送", "配送进度", "骑手", "送到哪了", "还要多久", "delivery"],
                examples=["配送到哪里了", "骑手到哪了", "还有多久送到"],
                tool=CheckDeliveryStatusTool(),
            ),
            SkillAbility(
                name="customer_service",
                description="生成智能客服回复，处理投诉、退款、售后等问题",
                keywords=["客服", "投诉", "退款", "售后", "customer service", "退货", "赔偿"],
                examples=["我要退款", "配送太慢了", "送错东西了", "我想投诉"],
                tool=CustomerServiceTool(),
            ),
            SkillAbility(
                name="get_user_default_info",
                description="获取当前用户的默认收货信息（姓名、电话、地址），下单前自动填充",
                keywords=["默认信息", "收货信息", "我的信息", "个人资料", "默认地址", "user info", "default"],
                examples=["我的收货信息是什么", "用我的默认地址下单", "我的姓名和电话"],
                tool=GetUserDefaultInfoTool(),
            ),
            SkillAbility(
                name="get_user_profile",
                description="获取当前登录用户的个人资料（昵称、手机号、积分、头像）",
                keywords=["我的信息", "个人资料", "我的资料", "profile", "我的手机号", "我的昵称"],
                examples=["我的个人资料", "我的手机号是什么", "查看我的资料"],
                tool=GetUserProfileTool(),
            ),
            SkillAbility(
                name="get_user_cart",
                description="查看当前用户的购物车内容",
                keywords=["购物车", "cart", "我的购物车", "购物车内容"],
                examples=["查看购物车", "购物车里有什么", "我的购物车"],
                tool=GetUserCartTool(),
            ),
            SkillAbility(
                name="add_to_cart",
                description="将菜品添加到购物车",
                keywords=["加购物车", "加入购物车", "add to cart", "添加到购物车"],
                examples=["帮我加个麦辣鸡腿堡到购物车", "把番茄炒蛋加入购物车"],
                tool=AddToCartTool(),
            ),
            SkillAbility(
                name="get_user_orders",
                description="获取当前用户的订单列表",
                keywords=["我的订单", "历史订单", "订单列表", "orders", "全部订单"],
                examples=["我的订单有哪些", "查看历史订单", "帮我查一下订单"],
                tool=GetUserOrdersTool(),
            ),
            SkillAbility(
                name="cancel_order",
                description="取消指定订单",
                keywords=["取消订单", "cancel", "取消"],
                examples=["取消订单", "帮我取消订单", "不想要了"],
                tool=CancelOrderTool(),
            ),
            SkillAbility(
                name="pay_order",
                description="支付指定订单",
                keywords=["支付", "付款", "pay", "付钱"],
                examples=["帮我支付", "付款", "支付订单"],
                tool=PayOrderTool(),
            ),
        ]


skill_instance = DeliverySkill()
