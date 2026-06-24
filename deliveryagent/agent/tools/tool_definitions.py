"""
Agent Tools: 所有内置工具定义
================================
所有工具继承 agent.tools.base.Tool 基类，通过 ToolRegistry.register() 注册。

包含：
1. 文件/Web 工具（在 filesystem.py / web.py 中，由 AgentLoop 默认注册）
2. 配送业务工具（定义在 delivery_tools.py）
3. 以下辅助工具：用户位置、用户ID、当前月份
"""

import random
from datetime import datetime
from typing import Any

from loguru import logger

from agent.tools.base import Tool
from agent.tools.delivery_tools import (
    GetStoresTool,
    SearchStoreTool,
    GetDishesTool,
    RecommendDishTool,
    QueryOrderTool,
    CheckDeliveryStatusTool,
    PlaceOrderTool,
    CustomerServiceTool,
    GetUserDefaultInfoTool,
    GetUserProfileTool,
    GetUserCartTool,
    AddToCartTool,
    GetUserOrdersTool,
    CancelOrderTool,
    PayOrderTool,
)

# Mock 数据
CITIES = ["北京", "上海", "广州", "深圳", "杭州", "成都", "武汉", "珠海", "中山"]
USER_IDS = ["1001", "1002", "1003", "1004", "1005"]


class UserLocationTool(Tool):
    """获取用户地理位置"""

    @property
    def name(self) -> str:
        return "get_user_location"

    @property
    def description(self) -> str:
        return "获取用户的当前地理位置信息（城市、详细地址）。可用于判断配送范围、推荐附近商家。"

    @property
    def parameters(self) -> dict[str, Any]:
        return {
            "user_id": {
                "type": "string",
                "description": "用户 ID（可选，不传则返回默认位置）",
                "required": False,
            },
        }

    async def execute(self, user_id: str = "", **kwargs: Any) -> str:
        jwt_token = kwargs.get("_jwt_token", "")

        # 有 JWT 时：通过 /api/user/addresses 获取真实位置
        if jwt_token:
            try:
                from delivery_config import DELIVERY_BACKEND_URL, HTTP_TIMEOUT
                import httpx
                async with httpx.AsyncClient(timeout=HTTP_TIMEOUT) as client:
                    resp = await client.get(
                        f"{DELIVERY_BACKEND_URL}/api/user/addresses",
                        headers={"Authorization": f"Bearer {jwt_token}"},
                        follow_redirects=True,
                    )
                    if resp.status_code == 200:
                        data = resp.json()
                        addrs = data.get("data") if isinstance(data, dict) else data
                        if isinstance(addrs, list) and len(addrs) > 0:
                            addr = addrs[0]
                            province = addr.get("province") or ""
                            city = addr.get("city") or ""
                            district = addr.get("district") or ""
                            street = addr.get("street") or ""
                            detail = addr.get("detail") or ""
                            parts = [p for p in [province, city, district, street, detail] if p]
                            if parts:
                                return f"📍 {' '.join(parts)}"
            except Exception:
                pass

        # 没有 JWT 但有 user_id：尝试 noAuth 接口
        if user_id:
            try:
                from delivery_config import DELIVERY_BACKEND_URL, HTTP_TIMEOUT
                import httpx
                async with httpx.AsyncClient(timeout=HTTP_TIMEOUT) as client:
                    resp = await client.get(
                        f"{DELIVERY_BACKEND_URL}/api/user/addresses?baseUserId={user_id}",
                        follow_redirects=True,
                    )
                    if resp.status_code == 200:
                        data = resp.json()
                        addrs = data.get("data") if isinstance(data, dict) else data
                        if isinstance(addrs, list) and len(addrs) > 0:
                            addr = addrs[0]
                            detail = addr.get("detail") or addr.get("address") or ""
                            city = addr.get("city") or ""
                            if detail or city:
                                parts = [p for p in [city, detail] if p]
                                return f"📍 {', '.join(parts)}"
            except Exception:
                pass

        # 系统配送范围默认为广州（中山大学）
        return "广东省广州市 中山大学"


class UserIDTool(Tool):
    """获取用户 ID（模拟）"""

    @property
    def name(self) -> str:
        return "get_user_id"

    @property
    def description(self) -> str:
        return "获取当前登录用户的ID，用于查询用户数据"

    @property
    def parameters(self) -> dict[str, Any]:
        return {}

    async def execute(self, **kwargs: Any) -> str:
        return random.choice(USER_IDS)


class CurrentMonthTool(Tool):
    """获取当前月份"""

    @property
    def name(self) -> str:
        return "get_current_month"

    @property
    def description(self) -> str:
        return "获取当前的月份，格式为YYYY-MM"

    @property
    def parameters(self) -> dict[str, Any]:
        return {}

    async def execute(self, **kwargs: Any) -> str:
        return datetime.now().strftime("%Y-%m")


# ─── 工具注册函数 ────────────────────────────


def register_all_tools(registry) -> None:
    """
    将所有业务工具注册到指定的 ToolRegistry。

    此函数由 api/routes.py 中的 get_orchestrator() 调用。
    """
    tools = [
        # 辅助工具
        UserLocationTool(),
        UserIDTool(),
        CurrentMonthTool(),
        # 配送业务工具
        GetStoresTool(),
        SearchStoreTool(),
        GetDishesTool(),
        RecommendDishTool(),
        PlaceOrderTool(),
        QueryOrderTool(),
        CheckDeliveryStatusTool(),
        CustomerServiceTool(),
        GetUserDefaultInfoTool(),
        # 用户鉴权工具（需要 JWT）
        GetUserProfileTool(),
        GetUserCartTool(),
        AddToCartTool(),
        GetUserOrdersTool(),
        CancelOrderTool(),
        PayOrderTool(),
    ]
    for tool in tools:
        registry.register(tool)
    logger.info(f"[Tools] 注册完成: {len(tools)} 个业务工具")
