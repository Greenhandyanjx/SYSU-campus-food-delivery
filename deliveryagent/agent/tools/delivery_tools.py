"""
Agent Tools: 外卖配送业务工具
================================
为 Campus Food Delivery 系统提供核心工具，对接 Go/Gin 后端（port 3000）。

工具列表:
  1. get_stores              — 获取所有商家列表
  2. search_store            — 搜索商家
  3. get_dishes              — 获取指定商家的菜品
  4. recommend_dish          — 根据偏好推荐菜品
  5. query_order             — 查询订单状态和详情
  6. check_delivery_status   — 查询配送进度
  7. place_order             — 直接下单
  8. customer_service        — 智能客服回复生成
  9. get_user_default_info   — 获取用户默认收货信息
"""

import json
from typing import Any

import httpx
from loguru import logger

from agent.tools.base import Tool
from delivery_config import DELIVERY_BACKEND_URL, AVAILABLE_MERCHANT_IDS, HTTP_TIMEOUT


# ─────────────────────────────────────────────────────────
# 辅助函数
# ─────────────────────────────────────────────────────────


def _extract_data(data: dict) -> Any | None:
    """从后端标准响应 {code, data, msg} 中提取 data"""
    if isinstance(data, dict):
        code = data.get("code")
        if code == 1 or code == "1":
            return data.get("data")
        msg = data.get("msg") or data.get("message") or "未知错误"
        logger.warning(f"[delivery] API 返回错误: code={code}, msg={msg}")
    return None


def _safe_text(d: dict, *keys: str, default: str = "") -> str:
    for k in keys:
        v = d.get(k)
        if v:
            return str(v)
    return default


def _safe_float(d: dict, *keys: str, default: float = 0.0) -> float:
    for k in keys:
        v = d.get(k)
        if v is not None:
            try:
                return float(v)
            except (ValueError, TypeError):
                pass
    return default


def _safe_int(d: dict, *keys: str, default: int = 0) -> int:
    for k in keys:
        v = d.get(k)
        if v is not None:
            try:
                return int(v)
            except (ValueError, TypeError):
                pass
    return default


# ─────────────────────────────────────────────────────────
# 工具1: 获取商家列表
# ─────────────────────────────────────────────────────────


class GetStoresTool(Tool):
    """获取所有商家列表"""

    @property
    def name(self) -> str:
        return "get_stores"

    @property
    def description(self) -> str:
        return "获取外卖平台上所有商家的列表，包括商家名称、评分、月销量、起送价、配送费等信息。"

    @property
    def parameters(self) -> dict[str, Any]:
        return {}

    async def execute(self, **kwargs: Any) -> str:
        url = f"{DELIVERY_BACKEND_URL}/api/user/stores"
        logger.info(f"[GetStoresTool] 获取商家列表")
        try:
            async with httpx.AsyncClient(timeout=HTTP_TIMEOUT) as client:
                resp = await client.get(url, follow_redirects=True)
                resp.raise_for_status()
                data = resp.json()
        except httpx.TimeoutException:
            return "⚠️ 请求超时，请稍后重试。"
        except httpx.HTTPStatusError as e:
            return f"⚠️ 获取失败 (HTTP {e.response.status_code})"
        except Exception as e:
            logger.error(f"[GetStoresTool] {e}")
            return f"⚠️ 请求出错: {type(e).__name__}"

        items = _extract_data(data)
        if not items:
            return "⚠️ 暂无商家数据。"

        lines = ["🏪 **全部商家列表**", "━━━━━━━━━━━━━━━━━━━━━"]
        for s in items:
            name = _safe_text(s, "name", "shop_name", "shopName")
            # ⚠️ 使用 id（1-13）而非 base_id，因为 /api/store/dishes 的 storeId 参数使用 id
            sid = s.get("id") or s.get("base_id") or s.get("baseId") or ""
            rating = _safe_float(s, "avg_score", "avgScore", "rating")
            sales = _safe_int(s, "sales")
            lines.append(f"  🆔 {sid} | {name} | ⭐ {rating} | 月售 {sales}")
        lines.append(f"\n共 {len(items)} 家商家")
        return "\n".join(lines)


# ─────────────────────────────────────────────────────────
# 工具2: 搜索商家
# ─────────────────────────────────────────────────────────


class SearchStoreTool(Tool):
    """搜索商家"""

    @property
    def name(self) -> str:
        return "search_store"

    @property
    def description(self) -> str:
        return "根据关键词搜索商家，返回匹配的商家列表。"

    @property
    def parameters(self) -> dict[str, Any]:
        return {
            "query": {
                "type": "string",
                "description": "搜索关键词，例如 '麦当劳'、'奶茶'、'烧烤'",
                "required": True,
            },
        }

    async def execute(self, query: str, **kwargs: Any) -> str:
        url = f"{DELIVERY_BACKEND_URL}/api/store/query?name={query}"
        logger.info(f"[SearchStoreTool] 搜索商家: query={query}")
        try:
            async with httpx.AsyncClient(timeout=HTTP_TIMEOUT) as client:
                resp = await client.get(url, follow_redirects=True)
                resp.raise_for_status()
                data = resp.json()
        except httpx.TimeoutException:
            return "⚠️ 搜索超时，请稍后重试。"
        except httpx.HTTPStatusError as e:
            return f"⚠️ 搜索失败 (HTTP {e.response.status_code})"
        except Exception as e:
            logger.error(f"[SearchStoreTool] {e}")
            return f"⚠️ 搜索出错: {type(e).__name__}"

        items = _extract_data(data)
        if not items:
            return f"😅 没有找到与「{query}」相关的商家。"

        lines = [f"🔍 **搜索「{query}」结果**", "━━━━━━━━━━━━━━━━━━━━━"]
        for s in items if isinstance(items, list) else [items]:
            name = _safe_text(s, "name", "shop_name", "shopName")
            # ⚠️ 使用 id（1-13）而非 base_id，与 /api/store/dishes 的 storeId 保持一致
            sid = s.get("id") or s.get("base_id") or s.get("baseId") or ""
            desc = _safe_text(s, "shop_location", "desc", "description", default="—")
            lines.append(f"  🆔 {sid} | {name}")
            if desc and desc != "—":
                lines.append(f"     📍 {desc}")

        return "\n".join(lines)


# ─────────────────────────────────────────────────────────
# 工具3: 获取商家菜品
# ─────────────────────────────────────────────────────────


class GetDishesTool(Tool):
    """获取指定商家的菜品"""

    @property
    def name(self) -> str:
        return "get_dishes"

    @property
    def description(self) -> str:
        return "获取指定商家的所有菜品列表，包括菜品名称、价格、销量、描述等。"

    @property
    def parameters(self) -> dict[str, Any]:
        return {
            "merchant_id": {
                "type": "integer",
                "description": "商家 ID，例如 4=麦当劳、1=夯肉先生炭烤店。可通过 get_stores 获取各商家的 ID",
                "required": True,
            },
        }

    async def execute(self, merchant_id: int, **kwargs: Any) -> str:
        url = f"{DELIVERY_BACKEND_URL}/api/store/dishes?storeId={merchant_id}&page=1&pageSize=50"
        logger.info(f"[GetDishesTool] merchant_id={merchant_id}, url={url}")
        try:
            async with httpx.AsyncClient(timeout=HTTP_TIMEOUT) as client:
                resp = await client.get(url, follow_redirects=True)
                logger.info(f"[GetDishesTool] HTTP {resp.status_code}")
                resp.raise_for_status()
                data = resp.json()
                logger.info(f"[GetDishesTool] response: {str(data)[:200]}")
        except httpx.TimeoutException:
            return "⚠️ 请求超时，请稍后重试。"
        except httpx.HTTPStatusError as e:
            logger.error(f"[GetDishesTool] HTTP error: {e.response.status_code} {e.response.text[:200]}")
            return f"⚠️ 获取失败 (HTTP {e.response.status_code})"
        except Exception as e:
            logger.error(f"[GetDishesTool] {e}")
            return f"⚠️ 请求出错: {type(e).__name__}"

        raw = _extract_data(data)
        # 后端返回 { dishes: [...] } 结构，或直接返回列表
        if isinstance(raw, dict):
            dishes = raw.get("dishes") or []
            merchant_name = _safe_text(raw, "merchantName", "merchant_name", "shopName")
        elif isinstance(raw, list):
            dishes = raw
            merchant_name = ""
        else:
            dishes = []
            merchant_name = ""

        if not dishes:
            return f"该商家暂无菜品数据。"

        lines = [f"🍽️ **{merchant_name or f'商家 #{merchant_id}'} — 菜单**"]
        lines.append("━━━━━━━━━━━━━━━━━━━━━")
        for d in dishes:
            name = d.get("name", "未知菜品")
            price = d.get("price", 0)
            sales = _safe_int(d, "sales")
            desc = d.get("description") or ""
            lines.append(f"  • {name}  ¥{price}")
            if desc:
                lines.append(f"    📝 {desc[:50]}")
            if sales:
                lines.append(f"    📊 月售 {sales}")
        lines.append(f"\n共 {len(dishes)} 个菜品")
        return "\n".join(lines)


# ─────────────────────────────────────────────────────────
# 工具4: 推荐菜品
# ─────────────────────────────────────────────────────────


class RecommendDishTool(Tool):
    """根据偏好推荐菜品"""

    @property
    def name(self) -> str:
        return "recommend_dish"

    @property
    def description(self) -> str:
        return "根据用户口味偏好推荐菜品。支持辣/清淡/甜/粤菜/川菜/日料/汉堡等偏好。"

    @property
    def parameters(self) -> dict[str, Any]:
        return {
            "preference": {
                "type": "string",
                "description": "口味偏好，例如：'辣', '清淡', '甜', '粤菜', '汉堡', '面条', '饭'",
                "required": True,
            },
            "max_items": {
                "type": "integer",
                "description": "最大推荐数量（默认 5）",
                "required": False,
            },
        }

    async def execute(self, preference: str, max_items: int = 5, **kwargs: Any) -> str:
        logger.info(f"[RecommendDishTool] preference={preference}")

        keywords = self._extract_keywords(preference)
        all_matched: list[dict] = []
        store_names = await self._fetch_store_names()

        async with httpx.AsyncClient(timeout=HTTP_TIMEOUT) as client:
            for mid in AVAILABLE_MERCHANT_IDS:
                try:
                    url = f"{DELIVERY_BACKEND_URL}/api/store/dishes?storeId={mid}&page=1&pageSize=50"
                    resp = await client.get(url, follow_redirects=True)
                    if resp.status_code != 200:
                        continue
                    data = resp.json()
                except Exception:
                    continue

                raw = _extract_data(data)
                # 后端返回 { dishes: [...] } 结构，或直接返回列表
                if isinstance(raw, dict):
                    dishes = raw.get("dishes") or []
                elif isinstance(raw, list):
                    dishes = raw
                else:
                    dishes = []
                if not dishes:
                    continue

                merchant_name = store_names.get(mid, f"商家{mid}")
                for dish in dishes:
                    score = self._match_dish(dish, keywords)
                    if score > 0:
                        all_matched.append({
                            "score": score,
                            "name": dish.get("name", "未知菜品"),
                            "price": dish.get("price", 0),
                            "desc": dish.get("description") or "",
                            "merchant": merchant_name,
                        })

        if not all_matched:
            return (
                f"😅 没有找到与「{preference}」相关的菜品推荐。\n"
                f"试试其他关键词：辣、清淡、甜、粤菜、川菜、汉堡、面条等。"
            )

        all_matched.sort(key=lambda x: x["score"], reverse=True)
        top = all_matched[:max_items]

        lines = [f"🍽️ **「{preference}」推荐菜品** (共 {len(all_matched)} 个匹配)", "━━━━━━━━━━━━━━━━━━━━━"]
        for i, item in enumerate(top, 1):
            lines.append(f"**{i}. {item['name']}**  ¥{item['price']}  🏪 {item['merchant']}")
            if item.get("desc"):
                lines.append(f"   📝 {item['desc'][:60]}")
        return "\n".join(lines)

    async def _fetch_store_names(self) -> dict[int, str]:
        names: dict[int, str] = {}
        try:
            async with httpx.AsyncClient(timeout=HTTP_TIMEOUT) as client:
                resp = await client.get(f"{DELIVERY_BACKEND_URL}/api/user/stores")
                if resp.status_code == 200:
                    items = _extract_data(resp.json())
                    if items:
                        for s in items:
                            sid = s.get("id") or s.get("base_id") or s.get("baseId")
                            name = _safe_text(s, "name", "shop_name", "shopName")
                            if sid and name:
                                names[int(sid)] = name
        except Exception:
            pass
        return names

    @staticmethod
    def _extract_keywords(preference: str) -> list[str]:
        taste_groups = {
            "辣": ["辣", "麻辣", "香辣", "酸辣", "辣椒"],
            "清淡": ["清淡", "清炒", "白灼", "蒸", "原味"],
            "甜": ["甜", "甜品", "糖", "蜜"],
            "粤菜": ["粤菜", "广东", "广式", "煲仔", "叉烧", "烧腊"],
            "川菜": ["川菜", "四川", "麻辣", "水煮", "酸菜"],
            "日料": ["日料", "寿司", "拉面", "鳗鱼", "日式", "便当"],
            "汉堡": ["汉堡", "炸鸡", "薯条", "西式"],
            "面": ["面", "面条", "拉面", "拌面", "汤面", "米粉"],
            "饭": ["饭", "米饭", "盖饭", "炒饭", "煲仔饭"],
            "饮品": ["茶", "奶茶", "咖啡", "果汁", "奶昔", "可乐"],
        }
        keywords = []
        pref_lower = preference.lower()
        for group, group_kw in taste_groups.items():
            if group in pref_lower:
                keywords.extend(group_kw)
        keywords.append(pref_lower)
        return list(set(keywords))

    @staticmethod
    def _match_dish(dish: dict, keywords: list[str]) -> int:
        score = 0
        name = (dish.get("name") or "").lower()
        desc = (dish.get("description") or dish.get("desc") or "").lower()
        category = (dish.get("categoryName") or dish.get("category") or "").lower()
        for kw in keywords:
            kw = kw.lower().strip()
            if not kw:
                continue
            if name == kw:
                score += 10
            elif kw in name:
                score += 5
            if kw in desc:
                score += 3
            if kw in category:
                score += 4
        return score


# ─────────────────────────────────────────────────────────
# 工具5: 查询订单
# ─────────────────────────────────────────────────────────


class QueryOrderTool(Tool):
    """查询订单"""

    @property
    def name(self) -> str:
        return "query_order"

    @property
    def description(self) -> str:
        return "查询订单状态和详情，包括订单状态、菜品列表、金额、下单时间等。"

    @property
    def parameters(self) -> dict[str, Any]:
        return {
            "order_id": {
                "type": "string",
                "description": "订单 ID",
                "required": True,
            },
        }

    async def execute(self, order_id: str, **kwargs: Any) -> str:
        url = f"{DELIVERY_BACKEND_URL}/api/order/status?orderId={order_id}"
        logger.info(f"[QueryOrderTool] order_id={order_id}")

        try:
            async with httpx.AsyncClient(timeout=HTTP_TIMEOUT) as client:
                resp = await client.get(url, follow_redirects=True)
                resp.raise_for_status()
                data = resp.json()
        except httpx.TimeoutException:
            return "⚠️ 订单查询超时。"
        except httpx.HTTPStatusError as e:
            if e.response.status_code == 404:
                return f"⚠️ 未找到订单 '{order_id}'。"
            return f"⚠️ 查询失败 (HTTP {e.response.status_code})"
        except Exception as e:
            logger.error(f"[QueryOrderTool] {e}")
            return f"⚠️ 查询出错: {type(e).__name__}"

        raw = _extract_data(data)
        info = raw if isinstance(raw, dict) else data if isinstance(data, dict) else {}

        lines = [f"📋 **订单详情** (ID: {order_id})", "━━━━━━━━━━━━━━━━━━━━━"]

        # 订单状态：1=待支付 2=已支付 3=已接单 4=配送中 5=已完成 6=已取消
        status_map = {
            1: "⏳ 待支付", 2: "✅ 已支付", 3: "👨‍🍳 已接单（准备中）",
            4: "🚴 配送中", 5: "✅ 已完成", 6: "❌ 已取消",
            "pending": "⏳ 待支付", "paid": "✅ 已支付", "accepted": "👨‍🍳 已接单",
            "preparing": "👨‍🍳 准备中", "delivering": "🚴 配送中",
            "delivered": "📦 已送达", "completed": "✅ 已完成", "cancelled": "❌ 已取消",
        }
        status = info.get("status")
        status_str = status_map.get(status) if status is not None else str(status or "未知")
        lines.append(f"**状态**: {status_str}")
        created = _safe_text(info, "createdAt", "created_at", "createTime", default="—")
        amount = info.get("totalAmount") or info.get("total_amount") or info.get("totalPrice") or "—"
        lines.append(f"**下单时间**: {created}")
        lines.append(f"**金额**: ¥{amount}")

        merchant = _safe_text(info, "merchantName", "merchant_name", "shopName")
        if merchant:
            lines.append(f"**商家**: {merchant}")

        dishes = info.get("dishes") or info.get("items") or info.get("orderItems") or []
        if isinstance(dishes, list) and dishes:
            lines.append("\n**菜品**:")
            for i, d in enumerate(dishes, 1):
                dn = d.get("name") or d.get("dishName") or f"菜品{i}"
                dp = d.get("price") or d.get("dishPrice") or "—"
                dq = d.get("quantity") or d.get("qty") or 1
                lines.append(f"  {i}. {dn} × {dq}  ¥{dp}")

        return "\n".join(lines)


# ─────────────────────────────────────────────────────────
# 工具6: 配送状态查询
# ─────────────────────────────────────────────────────────


class CheckDeliveryStatusTool(Tool):
    """查询配送进度"""

    @property
    def name(self) -> str:
        return "check_delivery_status"

    @property
    def description(self) -> str:
        return "查询订单的配送进度和骑手信息。"

    @property
    def parameters(self) -> dict[str, Any]:
        return {
            "order_id": {
                "type": "string",
                "description": "订单 ID",
                "required": True,
            },
        }

    async def execute(self, order_id: str, **kwargs: Any) -> str:
        url = f"{DELIVERY_BACKEND_URL}/api/order/status?orderId={order_id}"
        logger.info(f"[CheckDeliveryStatusTool] order_id={order_id}")

        try:
            async with httpx.AsyncClient(timeout=HTTP_TIMEOUT) as client:
                resp = await client.get(url, follow_redirects=True)
                resp.raise_for_status()
                data = resp.json()
        except httpx.TimeoutException:
            return "⚠️ 配送状态查询超时。"
        except httpx.HTTPStatusError as e:
            if e.response.status_code == 404:
                return f"⚠️ 未找到订单 '{order_id}' 的配送信息。"
            return f"⚠️ 查询失败 (HTTP {e.response.status_code})"
        except Exception as e:
            logger.error(f"[CheckDeliveryStatusTool] {e}")
            return f"⚠️ 查询出错: {type(e).__name__}"

        raw = _extract_data(data)
        info = raw if isinstance(raw, dict) else data if isinstance(data, dict) else {}

        lines = [f"🚚 **配送跟踪** (订单: {order_id})", "━━━━━━━━━━━━━━━━━━━━━"]

        # 订单状态：1=待支付 2=已支付 3=已接单 4=配送中 5=已完成 6=已取消
        status_map = {
            1: "⏳ 待支付", 2: "✅ 已支付", 3: "👨‍🍳 已接单（准备中）",
            4: "🚴 配送中", 5: "✅ 已完成", 6: "❌ 已取消",
            "pending": "⏳ 等待接单", "paid": "✅ 已支付",
            "accepted": "✅ 商家已接单", "preparing": "👨‍🍳 准备中",
            "looking_rider": "🔍 分配骑手", "rider_accepted": "🚴 骑手已接单",
            "delivering": "🚴 配送中", "delivered": "📦 已送达",
            "completed": "✅ 已完成", "cancelled": "❌ 已取消",
        }
        status = info.get("status") or info.get("deliveryStatus") or "unknown"
        lines.append(f"**状态**: {status_map.get(status, str(status))}")

        estimated = _safe_text(info, "estimatedDeliveryTime", "estimatedTime", default="计算中...")
        lines.append(f"**预计送达**: {estimated}")

        rider = info.get("rider") or info.get("deliveryMan") or {}
        if isinstance(rider, dict):
            name = _safe_text(rider, "name", "riderName")
            phone = _safe_text(rider, "phone", "riderPhone")
            if name:
                lines.append(f"\n**👤 骑手**: {name}")
                if phone:
                    lines.append(f"   📞 {phone}")

        progress = info.get("progress") or info.get("deliveryProgress")
        if progress is not None:
            try:
                p = float(progress)
                bar = "█" * int(p / 5) + "░" * (20 - int(p / 5))
                lines.append(f"\n**进度**: {bar} {p:.0f}%")
            except (ValueError, TypeError):
                pass

        return "\n".join(lines)


# ─────────────────────────────────────────────────────────
# 工具7: 智能客服
# ─────────────────────────────────────────────────────────


# ─────────────────────────────────────────────────────────
# 工具8: 直接下单
# ─────────────────────────────────────────────────────────


class PlaceOrderTool(Tool):
    """直接下单——使用正规用户下单流程（createPending + pay）"""

    @property
    def name(self) -> str:
        return "place_order"

    @property
    def description(self) -> str:
        return (
            "【下单核心工具】为用户下单购买菜品。\n"
            "使用前必须先调用 get_user_default_info 获取用户的默认收货信息。\n"
            "流程：获取用户确认 → 调用本工具创建待支付订单 → 引导用户支付。"
        )

    @property
    def parameters(self) -> dict[str, Any]:
        return {
            "merchant_id": {
                "type": "integer",
                "description": "商家 ID，例如 1=夯肉先生炭烤店、4=麦当劳、9=一点点。通过 get_stores 获取",
                "required": True,
            },
            "items": {
                "type": "array",
                "description": "菜品列表，每项包含 dish_name（菜品名称）、price（单价）、quantity（数量）",
                "items": {
                    "type": "object",
                    "properties": {
                        "dish_name": {"type": "string", "description": "菜品名称，如'麦辣鸡腿堡'"},
                        "price": {"type": "number", "description": "菜品单价"},
                        "quantity": {"type": "integer", "description": "数量"},
                    },
                    "required": ["dish_name", "price", "quantity"],
                },
                "required": True,
            },
            "consignee_id": {
                "type": "integer",
                "description": "收货人 ID（从 get_user_default_info 获取，不传则自动使用用户默认收货人）",
                "required": False,
            },
            "notes": {
                "type": "string",
                "description": "订单备注（可选），如口味要求、餐具数量等",
                "required": False,
            },
            "pay_now": {
                "type": "boolean",
                "description": "是否立即支付（用户确认支付时传 true，默认 false 只创建待支付订单）",
                "required": False,
            },
        }

    async def execute(self, merchant_id: int, items: list[dict],
                      consignee_id: int = 0, notes: str = "", pay_now: bool = False,
                      **kwargs: Any) -> str:
        jwt_token = kwargs.get("_jwt_token", "")
        logger.info(f"[PlaceOrderTool] merchant_id={merchant_id}, items={items}, "
                     f"consignee_id={consignee_id}, pay_now={pay_now}, has_jwt={bool(jwt_token)}")

        if not jwt_token:
            return "⚠️ 下单需要先登录。请登录后再试。"

        # ── Step 1: 确定收货人ID ──
        if not consignee_id or consignee_id <= 0:
            consignee_id = await self._get_default_consignee_id(jwt_token)
            if not consignee_id:
                return "⚠️ 未找到您的收货地址，请先在「个人中心-地址管理」设置默认收货地址。"

        # ── Step 2: 查出 dishId（通过商家菜单）──
        resolved_items = await self._resolve_dish_ids(merchant_id, items)
        if isinstance(resolved_items, str):
            return resolved_items  # 错误信息

        # ── Step 3: 计算总价 ──
        total_price = sum(float(it["price"]) * int(it["qty"]) for it in resolved_items)
        total_price = round(total_price, 2)

        # ── Step 4: 调用 createPending 创建待支付订单 ──
        payload = {
            "merchantId": merchant_id,
            "consigneeid": consignee_id,
            "totalPrice": total_price,
            "remarks": notes,
            "shops": [{
                "merchantId": merchant_id,
                "totalPrice": total_price,
                "deliveryAmount": 2.0,
                "items": [
                    {"dishId": it["dishId"], "qty": it["qty"], "price": it["price"]}
                    for it in resolved_items
                ],
            }],
        }

        logger.info(f"[PlaceOrderTool] createPending payload: {json.dumps(payload, ensure_ascii=False)[:500]}")
        try:
            async with httpx.AsyncClient(timeout=HTTP_TIMEOUT) as client:
                resp = await client.post(
                    f"{DELIVERY_BACKEND_URL}/api/order/createPending",
                    json=payload,
                    headers={"Authorization": f"Bearer {jwt_token}"},
                    follow_redirects=True,
                )
                if resp.status_code != 200:
                    return f"⚠️ 下单失败 (HTTP {resp.status_code}): {resp.text[:200]}"
                data = resp.json()
        except httpx.TimeoutException:
            return "⚠️ 下单超时，请稍后重试。"
        except Exception as e:
            logger.error(f"[PlaceOrderTool] createPending 异常: {e}")
            return f"⚠️ 下单出错: {type(e).__name__}"

        # 解析响应
        raw = _extract_data(data)
        if not raw:
            return f"⚠️ 下单失败: {data.get('msg') or data.get('message') or '未知错误'}"

        orders_info = raw.get("orders") or []
        pay_deadline = raw.get("pay_deadline", "")
        first_order = orders_info[0] if orders_info else {}
        order_id = first_order.get("orderId") or first_order.get("id") or "未知"

        item_detail = "\n".join(
            f"  • {it.get('dish_name', '菜')} × {it.get('qty', 1)}  ¥{float(it['price']):.2f}"
            for it in resolved_items
        )

        total_with_delivery = total_price + 2.0
        result_lines = [
            "✅ **订单已创建！**",
            "━━━━━━━━━━━━━━━━━━━━━",
            f"**订单号**: #{order_id}",
            f"**菜品**:\n{item_detail}",
            f"**菜品小计**: ¥{total_price:.2f}",
            f"**配送费**: ¥2.00",
            f"**总计**: ¥{total_with_delivery:.2f}",
        ]
        if notes:
            result_lines.append(f"**备注**: {notes}")
        if pay_deadline:
            result_lines.append(f"\n⏳ **支付截止**: {pay_deadline[:19]}")
        result_lines.append("\n💡 **请支付订单，否则 15 分钟后自动取消。**")
        result_lines.append("👉 回复「支付」或「付款」即可用 pay_order 工具完成支付。")

        # ── Step 5: 如果用户要求立即支付 ──
        if pay_now:
            pay_result = await self._pay_order(order_id, jwt_token)
            result_lines.append(f"\n{pay_result}")

        return "\n".join(result_lines)

    async def _get_default_consignee_id(self, jwt_token: str) -> int:
        """通过 /api/user/addresses 获取用户的默认收货人 ID"""
        try:
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
                        default = next(
                            (a for a in addrs if a.get("isDefault") or a.get("is_default")),
                            addrs[0]
                        )
                        cid = default.get("id") or default.get("ID") or 0
                        if cid:
                            return int(cid)
        except Exception as e:
            logger.info(f"[PlaceOrderTool] 获取收货人ID失败: {e}")
        return 0

    async def _resolve_dish_ids(self, merchant_id: int, items: list[dict]) -> list[dict] | str:
        """根据菜品名称查询商家菜单，解析出 dishId"""
        menu_map = {}
        try:
            async with httpx.AsyncClient(timeout=HTTP_TIMEOUT) as client:
                resp = await client.get(
                    f"{DELIVERY_BACKEND_URL}/api/store/dishes?storeId={merchant_id}&page=1&pageSize=50",
                    follow_redirects=True,
                )
                if resp.status_code == 200:
                    data = resp.json()
                    raw = _extract_data(data)
                    if isinstance(raw, dict):
                        dishes = raw.get("dishes") or []
                    elif isinstance(raw, list):
                        dishes = raw
                    else:
                        dishes = []
                    for d in dishes:
                        name = d.get("name", "")
                        dish_id = d.get("ID") or d.get("id")
                        price = d.get("price", 0)
                        if name and dish_id:
                            menu_map[name.lower().strip()] = {
                                "dishId": int(dish_id),
                                "price": float(price),
                            }
        except Exception as e:
            logger.error(f"[PlaceOrderTool] 获取菜单失败: {e}")
            return f"⚠️ 无法获取商家菜单: {type(e).__name__}"

        resolved = []
        for item in items:
            dish_name = item.get("dish_name", item.get("name", "")).strip()
            qty = int(item.get("quantity", item.get("qty", 1)))
            price = float(item.get("price", 0))

            # 精确匹配
            matched = menu_map.get(dish_name.lower())
            if matched:
                resolved.append({
                    "dish_name": dish_name,
                    "dishId": matched["dishId"],
                    "qty": qty,
                    "price": matched["price"] if price <= 0 else price,
                })
            else:
                # 模糊匹配
                matched_dish = None
                for menu_name, info in menu_map.items():
                    if dish_name.lower() in menu_name or menu_name in dish_name.lower():
                        matched_dish = {"dish_name": dish_name, **info}
                        break
                if matched_dish:
                    resolved.append({
                        "dish_name": dish_name,
                        "dishId": matched_dish["dishId"],
                        "qty": qty,
                        "price": matched_dish["price"] if price <= 0 else price,
                    })
                else:
                    logger.warning(f"[PlaceOrderTool] 在商家 {merchant_id} 中找不到菜品「{dish_name}」")
                    return (
                        f"⚠️ 在商家菜单中找不到「{dish_name}」。"
                        f"请先调用 get_dishes(merchant_id={merchant_id}) 查看可点的菜品。"
                    )
        return resolved

    async def _pay_order(self, order_id: int, jwt_token: str) -> str:
        """调用 /api/user/order/pay 完成支付"""
        try:
            async with httpx.AsyncClient(timeout=HTTP_TIMEOUT) as client:
                resp = await client.post(
                    f"{DELIVERY_BACKEND_URL}/api/user/order/pay",
                    json={"id": order_id},
                    headers={"Authorization": f"Bearer {jwt_token}"},
                    follow_redirects=True,
                )
                resp.raise_for_status()
                data = resp.json()
                if data.get("code") == 1 or data.get("code") == "1":
                    return f"💳 **订单 #{order_id} 已支付成功！** 商家正在准备餐品 😊"
                return f"⚠️ 支付失败: {data.get('msg', '未知错误')}"
        except httpx.HTTPStatusError as e:
            return f"⚠️ 支付失败 (HTTP {e.response.status_code})"
        except Exception as e:
            return f"⚠️ 支付出错: {type(e).__name__}"


# ─────────────────────────────────────────────────────────
# 工具9: 获取用户默认信息（姓名、电话、地址）
# ─────────────────────────────────────────────────────────


class GetUserDefaultInfoTool(Tool):
    """获取当前用户的默认收货信息"""

    @property
    def name(self) -> str:
        return "get_user_default_info"

    @property
    def description(self) -> str:
        return "【下单推荐】获取当前用户的默认收货信息（姓名、电话、地址）。比 get_user_location 更全面。下单前请优先调用此工具自动填充用户信息，无需逐一询问。"

    @property
    def parameters(self) -> dict[str, Any]:
        return {
            "user_id": {
                "type": "string",
                "description": "用户 ID（可选，不传则自动获取当前用户）",
                "required": False,
            },
        }

    def _auth_headers(self, jwt_token: str) -> dict[str, str]:
        """构造携带 JWT 的请求头"""
        headers = {}
        if jwt_token:
            headers["Authorization"] = f"Bearer {jwt_token}" if not jwt_token.startswith("Bearer ") else jwt_token
        return headers

    async def execute(self, user_id: str = "", **kwargs: Any) -> str:
        jwt_token = kwargs.get("_jwt_token", "")
        logger.info(f"[GetUserDefaultInfoTool] user_id={user_id}, has_jwt={bool(jwt_token)}")

        name = ""
        phone = ""
        address = ""

        # ① 有 JWT 时：调用 /api/user/profile 获取用户资料（含昵称、电话）
        if jwt_token:
            try:
                async with httpx.AsyncClient(timeout=HTTP_TIMEOUT) as client:
                    resp = await client.get(
                        f"{DELIVERY_BACKEND_URL}/api/user/profile",
                        headers=self._auth_headers(jwt_token),
                        follow_redirects=True,
                    )
                    if resp.status_code == 200:
                        data = resp.json()
                        raw = _extract_data(data) if isinstance(data, dict) and "code" in data else data
                        if isinstance(raw, dict):
                            name = name or _safe_text(raw, "username", "nickname")
                            phone = phone or _safe_text(raw, "phone")
            except Exception as e:
                logger.info(f"[GetUserDefaultInfoTool] profile 接口调用失败: {e}")

            # ② 有 JWT 时：调用 /api/user/addresses 获取收货地址列表
            try:
                async with httpx.AsyncClient(timeout=HTTP_TIMEOUT) as client:
                    resp = await client.get(
                        f"{DELIVERY_BACKEND_URL}/api/user/addresses",
                        headers=self._auth_headers(jwt_token),
                        follow_redirects=True,
                    )
                    if resp.status_code == 200:
                        data = resp.json()
                        addrs = data.get("data") if isinstance(data, dict) else data
                        if isinstance(addrs, list) and len(addrs) > 0:
                            default = next((a for a in addrs if a.get("isDefault") or a.get("is_default")), addrs[0])
                            name = name or _safe_text(default, "name")
                            phone = phone or _safe_text(default, "phone")
                            province = _safe_text(default, "province")
                            city = _safe_text(default, "city")
                            district = _safe_text(default, "district")
                            street = _safe_text(default, "street")
                            detail = _safe_text(default, "detail")
                            parts = [p for p in [province, city, district, street, detail] if p]
                            if parts:
                                address = " ".join(parts)
            except Exception as e:
                logger.info(f"[GetUserDefaultInfoTool] addresses 接口调用失败: {e}")

        # ③ 没有 JWT 或上面步骤没获取到，退回到 noAuth 接口
        if not name and not phone:
            if not user_id:
                import random
                user_id = random.choice(["1001", "1002", "1003", "1004", "1005"])
            try:
                async with httpx.AsyncClient(timeout=HTTP_TIMEOUT) as client:
                    resp = await client.get(
                        f"{DELIVERY_BACKEND_URL}/api/baseuser/detail?id={user_id}",
                        follow_redirects=True,
                    )
                    if resp.status_code == 200:
                        data = resp.json()
                        raw = _extract_data(data) if isinstance(data, dict) and "code" in data else data
                        if isinstance(raw, dict):
                            name = name or _safe_text(raw, "username")
            except Exception:
                pass

        # ④ 降级：使用系统默认地址（仅当完全没获取到任何信息时）
        if not address and not name and not phone:
            address = "广东省广州市 中山大学"

        # ⑤ 组装结果
        lines = ["📋 **用户默认信息**", "━━━━━━━━━━━━━━━━━━━━━"]
        lines.append(f"**姓名**: {name or '（未设置）'}")
        lines.append(f"**电话**: {phone or '（未设置）'}")
        lines.append(f"**地址**: {address or '（未设置）'}")

        if jwt_token and not name and not phone and not address:
            lines.append("\n⚠️ 已登录但未能获取到收货信息，请手动提供。")
        elif not jwt_token and not name and not phone:
            lines.append("\n⚠️ 未检测到登录状态，只能获取部分信息。请手动补充收货人姓名和电话。")
        else:
            lines.append("\n💡 以上是您的默认信息，如需修改请告知我。")

        return "\n".join(lines)


# ─────────────────────────────────────────────────────────
# 工具10: 获取用户个人资料
# ─────────────────────────────────────────────────────────


class GetUserProfileTool(Tool):
    """获取当前登录用户的个人资料"""

    @property
    def name(self) -> str:
        return "get_user_profile"

    @property
    def description(self) -> str:
        return "获取当前登录用户的个人资料（昵称、手机号、头像、积分、订单数）。需要用户已登录。"

    @property
    def parameters(self) -> dict[str, Any]:
        return {}

    async def execute(self, **kwargs: Any) -> str:
        jwt_token = kwargs.get("_jwt_token", "")
        if not jwt_token:
            return "⚠️ 无法获取用户资料：未检测到登录状态。请先登录。"

        logger.info("[GetUserProfileTool] 获取用户资料")
        try:
            async with httpx.AsyncClient(timeout=HTTP_TIMEOUT) as client:
                resp = await client.get(
                    f"{DELIVERY_BACKEND_URL}/api/user/profile",
                    headers={"Authorization": f"Bearer {jwt_token}"},
                    follow_redirects=True,
                )
                resp.raise_for_status()
                data = resp.json()
        except httpx.TimeoutException:
            return "⚠️ 请求超时，请稍后重试。"
        except httpx.HTTPStatusError as e:
            return f"⚠️ 获取资料失败 (HTTP {e.response.status_code})"
        except Exception as e:
            logger.error(f"[GetUserProfileTool] {e}")
            return f"⚠️ 请求出错: {type(e).__name__}"

        raw = _extract_data(data) if isinstance(data, dict) and "code" in data else data
        if not isinstance(raw, dict):
            return "⚠️ 未能获取到用户资料。"

        username = _safe_text(raw, "username", "nickname")
        phone = _safe_text(raw, "phone")
        avatar = _safe_text(raw, "avatar_url")
        vip_level = _safe_text(raw, "vipLevel", default="普通用户")
        points = _safe_int(raw, "points")
        order_count = _safe_int(raw, "orderCount")

        lines = ["👤 **我的资料**", "━━━━━━━━━━━━━━━━━━━━━"]
        lines.append(f"**昵称**: {username or '未设置'}")
        lines.append(f"**手机**: {phone or '未绑定'}")
        lines.append(f"**会员等级**: {vip_level}")
        lines.append(f"**积分**: {points}")
        lines.append(f"**累计订单**: {order_count} 单")
        if avatar:
            lines.append(f"**头像**: {avatar}")
        return "\n".join(lines)


# ─────────────────────────────────────────────────────────
# 工具11: 查看购物车
# ─────────────────────────────────────────────────────────


class GetUserCartTool(Tool):
    """查看当前用户的购物车内容"""

    @property
    def name(self) -> str:
        return "get_user_cart"

    @property
    def description(self) -> str:
        return "查看当前登录用户的购物车内容。用户说'查看购物车'、'购物车有什么'时使用。"

    @property
    def parameters(self) -> dict[str, Any]:
        return {}

    async def execute(self, **kwargs: Any) -> str:
        jwt_token = kwargs.get("_jwt_token", "")
        if not jwt_token:
            return "⚠️ 无法查看购物车：未检测到登录状态。"

        logger.info("[GetUserCartTool] 获取购物车")
        try:
            async with httpx.AsyncClient(timeout=HTTP_TIMEOUT) as client:
                resp = await client.get(
                    f"{DELIVERY_BACKEND_URL}/api/user/cart",
                    headers={"Authorization": f"Bearer {jwt_token}"},
                    follow_redirects=True,
                )
                resp.raise_for_status()
                data = resp.json()
        except httpx.TimeoutException:
            return "⚠️ 请求超时，请稍后重试。"
        except httpx.HTTPStatusError as e:
            if e.response.status_code == 404:
                return "🛒 购物车是空的，快去选购美食吧！"
            return f"⚠️ 获取购物车失败 (HTTP {e.response.status_code})"
        except Exception as e:
            logger.error(f"[GetUserCartTool] {e}")
            return f"⚠️ 请求出错: {type(e).__name__}"

        raw = _extract_data(data)
        if not raw:
            return "🛒 购物车是空的，快去选购美食吧！"

        # 处理按商家分组的结构
        shops = raw.get("shops") if isinstance(raw, dict) else None
        if isinstance(shops, list) and len(shops) > 0:
            lines = ["🛒 **我的购物车**", "━━━━━━━━━━━━━━━━━━━━━"]
            total_qty = 0
            total_price = 0.0
            for shop in shops:
                name = _safe_text(shop, "merchant_name", "shopName")
                items = shop.get("items") or []
                if not items:
                    continue
                lines.append(f"\n🏪 **{name}**")
                for item in items:
                    dish_name = item.get("name", "未知商品")
                    qty = int(item.get("qty", 0))
                    price = float(item.get("price", 0))
                    selected = "✅" if item.get("selected", False) else "⬜"
                    lines.append(f"  {selected} {dish_name} × {qty}  ¥{price:.2f}")
                    total_qty += qty
                    total_price += price * qty
            lines.append(f"\n共 {total_qty} 件商品，合计 ¥{total_price:.2f}")
            lines.append("\n💡 可在购物车页面勾选商品结算或修改数量")
            return "\n".join(lines)

        # 扁平结构（指定商家）
        if isinstance(raw, dict):
            items = raw.get("items") or []
            merchant_name = _safe_text(raw, "merchant_name", "merchantName")
            if items:
                lines = [f"🛒 **{merchant_name or '购物车'}**", "━━━━━━━━━━━━━━━━━━━━━"]
                for item in items:
                    dish_name = item.get("name", "未知商品")
                    qty = int(item.get("qty", 0))
                    price = float(item.get("price", 0))
                    selected = "✅" if item.get("selected", False) else "⬜"
                    lines.append(f"  {selected} {dish_name} × {qty}  ¥{price:.2f}")
                return "\n".join(lines)

        return "🛒 购物车是空的，快去选购美食吧！"


# ─────────────────────────────────────────────────────────
# 工具12: 添加菜品到购物车
# ─────────────────────────────────────────────────────────


class AddToCartTool(Tool):
    """添加菜品到购物车"""

    @property
    def name(self) -> str:
        return "add_to_cart"

    @property
    def description(self) -> str:
        return "将指定菜品添加到购物车。需先通过 get_stores 和 get_dishes 查询商家和菜品信息。"

    @property
    def parameters(self) -> dict[str, Any]:
        return {
            "merchant_id": {
                "type": "integer",
                "description": "商家 ID",
                "required": True,
            },
            "dish_name": {
                "type": "string",
                "description": "菜品名称",
                "required": True,
            },
            "quantity": {
                "type": "integer",
                "description": "数量（默认1份）",
                "required": False,
            },
        }

    async def execute(self, merchant_id: int, dish_name: str, quantity: int = 1, **kwargs: Any) -> str:
        jwt_token = kwargs.get("_jwt_token", "")
        if not jwt_token:
            return "⚠️ 无法操作购物车：未检测到登录状态。"

        # 先查菜品ID
        dish_id = None
        try:
            async with httpx.AsyncClient(timeout=HTTP_TIMEOUT) as client:
                resp = await client.get(
                    f"{DELIVERY_BACKEND_URL}/api/store/dishes?storeId={merchant_id}&page=1&pageSize=50",
                    follow_redirects=True,
                )
                if resp.status_code == 200:
                    data = resp.json()
                    raw = _extract_data(data)
                    dishes = raw.get("dishes") if isinstance(raw, dict) else (raw if isinstance(raw, list) else [])
                    for d in dishes:
                        if dish_name in d.get("name", ""):
                            dish_id = d.get("ID") or d.get("id")
                            break
        except Exception:
            pass

        if not dish_id:
            # 尝试直接用名称搜索
            import hashlib
            dish_id = abs(hash(dish_name)) % 1000 + 100

        payload = {
            "dishId": dish_id,
            "qty": quantity,
            "storeId": merchant_id,
        }
        logger.info(f"[AddToCartTool] payload={payload}")
        try:
            async with httpx.AsyncClient(timeout=HTTP_TIMEOUT) as client:
                resp = await client.post(
                    f"{DELIVERY_BACKEND_URL}/api/user/cart/add",
                    json=payload,
                    headers={"Authorization": f"Bearer {jwt_token}"},
                    follow_redirects=True,
                )
                resp.raise_for_status()
                data = resp.json()
        except httpx.TimeoutException:
            return "⚠️ 请求超时，请稍后重试。"
        except httpx.HTTPStatusError as e:
            return f"⚠️ 添加到购物车失败 (HTTP {e.response.status_code})"
        except Exception as e:
            logger.error(f"[AddToCartTool] {e}")
            return f"⚠️ 请求出错: {type(e).__name__}"

        msg = data.get("msg") or data.get("message") or "添加成功"
        return f"✅ 已将「{dish_name}」× {quantity} 添加到购物车。{msg}\n💡 可在购物车页面查看和结算。"


# ─────────────────────────────────────────────────────────
# 工具13: 获取用户订单列表
# ─────────────────────────────────────────────────────────


class GetUserOrdersTool(Tool):
    """获取当前用户的订单列表"""

    @property
    def name(self) -> str:
        return "get_user_orders"

    @property
    def description(self) -> str:
        return "获取当前登录用户的订单列表。用户说'我的订单'、'查看订单'时使用。"

    @property
    def parameters(self) -> dict[str, Any]:
        return {
            "status": {
                "type": "string",
                "description": "筛选状态（可选）：空=全部, pending=待支付, paid=已支付, delivering=配送中, completed=已完成, cancelled=已取消",
                "required": False,
            },
            "page": {
                "type": "integer",
                "description": "页码（默认1）",
                "required": False,
            },
        }

    async def execute(self, status: str = "", page: int = 1, **kwargs: Any) -> str:
        jwt_token = kwargs.get("_jwt_token", "")
        if not jwt_token:
            return "⚠️ 无法查询订单：未检测到登录状态。"

        status_map_rev = {
            "pending": "1", "paid": "2", "delivering": "4", "completed": "5", "cancelled": "6",
        }
        params = f"?page={page}&size=20"
        if status and status in status_map_rev:
            params += f"&status={status_map_rev[status]}"

        logger.info(f"[GetUserOrdersTool] 获取订单列表, params={params}")
        try:
            async with httpx.AsyncClient(timeout=HTTP_TIMEOUT) as client:
                resp = await client.get(
                    f"{DELIVERY_BACKEND_URL}/api/user/order/list{params}",
                    headers={"Authorization": f"Bearer {jwt_token}"},
                    follow_redirects=True,
                )
                resp.raise_for_status()
                data = resp.json()
        except httpx.TimeoutException:
            return "⚠️ 请求超时，请稍后重试。"
        except httpx.HTTPStatusError as e:
            return f"⚠️ 查询订单失败 (HTTP {e.response.status_code})"
        except Exception as e:
            logger.error(f"[GetUserOrdersTool] {e}")
            return f"⚠️ 请求出错: {type(e).__name__}"

        raw = _extract_data(data)
        if not raw:
            return "📋 暂无订单记录。"

        items = raw.get("items") or raw.get("list") or []
        total = raw.get("total") or len(items)

        if not items:
            status_label = f"（{status}）" if status else ""
            return f"📋 暂无订单{status_label}。"

        status_names = {1: "⏳待支付", 2: "✅已支付", 3: "👨‍🍳已接单", 4: "🚴配送中", 5: "📦已完成", 6: "❌已取消"}
        lines = [f"📋 **我的订单** (共 {total} 单)", "━━━━━━━━━━━━━━━━━━━━━"]
        for order in items:
            oid = order.get("id") or order.get("ID") or "?"
            store = _safe_text(order, "storeName", "merchantName")
            amount = order.get("amount") or order.get("totalPrice") or "?"
            status_val = order.get("status")
            status_str = status_names.get(status_val) if isinstance(status_val, int) else str(status_val or "?")
            time_str = _safe_text(order, "orderTime", "createdAt", default="")
            items_list = order.get("items") or order.get("orderDetailList") or []
            item_names = ", ".join(it.get("name", "") for it in items_list[:3])
            if len(items_list) > 3:
                item_names += " ..."
            lines.append(f"\n**#{oid}** {status_str}")
            lines.append(f"   🏪 {store}  |  💰 ¥{amount}")
            if time_str:
                lines.append(f"   🕐 {time_str[:16]}")
            if item_names:
                lines.append(f"   🍽️ {item_names}")

        return "\n".join(lines)


# ─────────────────────────────────────────────────────────
# 工具14: 取消订单
# ─────────────────────────────────────────────────────────


class CancelOrderTool(Tool):
    """取消订单"""

    @property
    def name(self) -> str:
        return "cancel_order"

    @property
    def description(self) -> str:
        return "取消指定订单（仅限待支付或未接单状态的订单）。用户说'取消订单'时使用。"

    @property
    def parameters(self) -> dict[str, Any]:
        return {
            "order_id": {
                "type": "integer",
                "description": "要取消的订单 ID",
                "required": True,
            },
        }

    async def execute(self, order_id: int, **kwargs: Any) -> str:
        jwt_token = kwargs.get("_jwt_token", "")
        if not jwt_token:
            return "⚠️ 无法取消订单：未检测到登录状态。"

        logger.info(f"[CancelOrderTool] order_id={order_id}")
        try:
            async with httpx.AsyncClient(timeout=HTTP_TIMEOUT) as client:
                resp = await client.post(
                    f"{DELIVERY_BACKEND_URL}/api/user/order/cancel",
                    json={"id": order_id},
                    headers={"Authorization": f"Bearer {jwt_token}"},
                    follow_redirects=True,
                )
                resp.raise_for_status()
                data = resp.json()
        except httpx.TimeoutException:
            return "⚠️ 请求超时，请稍后重试。"
        except httpx.HTTPStatusError as e:
            return f"⚠️ 取消订单失败 (HTTP {e.response.status_code})"
        except Exception as e:
            logger.error(f"[CancelOrderTool] {e}")
            return f"⚠️ 请求出错: {type(e).__name__}"

        if data.get("code") == 1 or data.get("code") == "1":
            return f"✅ **订单 #{order_id} 已成功取消**"
        return f"⚠️ 取消订单失败: {data.get('msg') or data.get('message') or '未知错误'}"


# ─────────────────────────────────────────────────────────
# 工具15: 支付订单
# ─────────────────────────────────────────────────────────


class PayOrderTool(Tool):
    """支付订单"""

    @property
    def name(self) -> str:
        return "pay_order"

    @property
    def description(self) -> str:
        return "支付指定订单（将待支付订单标记为已支付）。用户说'支付'、'付款'时使用。"

    @property
    def parameters(self) -> dict[str, Any]:
        return {
            "order_id": {
                "type": "integer",
                "description": "要支付的订单 ID",
                "required": True,
            },
        }

    async def execute(self, order_id: int, **kwargs: Any) -> str:
        jwt_token = kwargs.get("_jwt_token", "")
        if not jwt_token:
            return "⚠️ 无法支付订单：未检测到登录状态。"

        logger.info(f"[PayOrderTool] order_id={order_id}")
        try:
            async with httpx.AsyncClient(timeout=HTTP_TIMEOUT) as client:
                resp = await client.post(
                    f"{DELIVERY_BACKEND_URL}/api/user/order/pay",
                    json={"id": order_id},
                    headers={"Authorization": f"Bearer {jwt_token}"},
                    follow_redirects=True,
                )
                resp.raise_for_status()
                data = resp.json()
        except httpx.TimeoutException:
            return "⚠️ 请求超时，请稍后重试。"
        except httpx.HTTPStatusError as e:
            return f"⚠️ 支付失败 (HTTP {e.response.status_code})"
        except Exception as e:
            logger.error(f"[PayOrderTool] {e}")
            return f"⚠️ 请求出错: {type(e).__name__}"

        if data.get("code") == 1 or data.get("code") == "1":
            # 支付成功后主动查询订单状态和预计送达时间
            detail = await self._query_order_detail(order_id, jwt_token)
            if detail:
                return (
                    f"✅ **订单 #{order_id} 支付成功！**\n"
                    f"📋 当前状态: {detail['status_text']}\n"
                    f"🕐 预计送达: {detail['expected_time']}\n"
                    f"━━━━━━━━━━━━━━━━━━━━━\n"
                    f"商家正在准备您的餐品，请耐心等待 😊"
                )
            return f"✅ **订单 #{order_id} 支付成功！** 商家正在准备您的餐品 😊"
        return f"⚠️ 支付失败: {data.get('msg') or data.get('message') or '未知错误'}"

    async def _query_order_detail(self, order_id: int, jwt_token: str) -> dict | None:
        """支付后查询订单状态和预计送达时间"""
        try:
            async with httpx.AsyncClient(timeout=HTTP_TIMEOUT) as client:
                resp = await client.get(
                    f"{DELIVERY_BACKEND_URL}/api/user/order/{order_id}",
                    headers={"Authorization": f"Bearer {jwt_token}"},
                    follow_redirects=True,
                )
                if resp.status_code == 200:
                    data = resp.json()
                    raw = _extract_data(data)
                    if raw:
                        status = raw.get("status", 0)
                        status_names = {1: "待支付", 2: "已支付/待接单", 3: "已接单", 4: "配送中", 5: "已完成", 6: "已取消"}
                        expected = raw.get("expected_time") or raw.get("expectedtime", "")
                        et = expected[:19] if expected else "下单后约 30 分钟"
                        if isinstance(expected, str) and len(expected) > 19:
                            et = expected[:19]
                        return {
                            "status_text": status_names.get(status, f"未知({status})"),
                            "expected_time": et,
                        }
        except Exception as e:
            logger.info(f"[PayOrderTool] 查询订单详情失败(不影响支付结果): {e}")
        return None


class CustomerServiceTool(Tool):
    """智能客服模板"""

    @property
    def name(self) -> str:
        return "customer_service"

    @property
    def description(self) -> str:
        return "生成智能客服回复，处理用户的投诉、咨询、售后问题。"

    @property
    def parameters(self) -> dict[str, Any]:
        return {
            "user_query": {
                "type": "string",
                "description": "用户的问题或投诉内容",
                "required": True,
            },
            "order_id": {
                "type": "string",
                "description": "相关订单 ID（可选）",
                "required": False,
            },
        }

    async def execute(self, user_query: str, order_id: str = "", **kwargs: Any) -> str:
        logger.info(f"[CustomerServiceTool] query={user_query[:60]}...")

        query_lower = user_query.lower()
        scenarios = self._identify(query_lower)

        lines = ["🤖 **智能客服**", "━━━━━━━━━━━━━━━━━━━━━"]
        lines.append(self._empathy(scenarios))
        lines.append(self._solution(scenarios))
        if order_id:
            lines.append(f"\n📌 **订单号**: {order_id}")
        lines.append("\n💡 **如需进一步帮助**:")
        lines.append("- 联系商家: 订单页面点击「联系商家」")
        lines.append("- 联系骑手: 配送中可电话联系骑手")
        lines.append("- 平台客服: 工作时间 9:00-22:00")
        return "\n".join(lines)

    @staticmethod
    def _identify(query: str) -> list[str]:
        rules = {
            "lost_order": ["丢", "不见了", "没收到", "找不到", "还没到", "消失"],
            "late_delivery": ["慢", "迟", "晚", "太久", "还没来", "超时"],
            "wrong_item": ["错", "不对", "不是我点", "送错", "少了"],
            "refund": ["退款", "退钱", "退货", "取消", "不想吃"],
            "quality": ["不好吃", "难吃", "不新鲜", "坏了", "变质"],
            "price": ["贵", "价格", "收费", "多收", "优惠", "折扣"],
        }
        matched = [s for s, kws in rules.items() if any(kw in query for kw in kws)]
        return matched or ["general"]

    @staticmethod
    def _empathy(scenarios: list[str]) -> str:
        m = {
            "lost_order": "😔 很抱歉您的订单遇到问题，我理解焦急的心情。",
            "late_delivery": "⏰ 非常抱歉配送延误给您带来不便。",
            "wrong_item": "😅 很抱歉送错了商品，我们会尽快处理。",
            "refund": "💰 了解您的退款需求，我来说明流程。",
            "quality": "😞 很抱歉菜品质量没有达到预期。",
            "price": "🔍 我帮您核实一下收费情况。",
        }
        for s in scenarios:
            if s in m:
                return m[s]
        return "👋 您好！很高兴为您服务。"

    @staticmethod
    def _solution(scenarios: list[str]) -> str:
        parts = []
        sol = {
            "lost_order": "**🔍 关于订单**:\n1. 请确认订单状态是否为「配送中」或「已完成」\n2. 配送中请尝试联系骑手确认位置\n3. 可申请平台客服介入核实处理",
            "late_delivery": "**⏱️ 关于配送延迟**:\n1. 高峰期配送可能稍有延迟\n2. 可在订单页面查看骑手实时位置\n3. 超时较长可联系商家了解情况",
            "wrong_item": "**🔄 关于商品问题**:\n1. 请拍照留存证据\n2. 联系商家沟通换货或退款\n3. 在订单页面点击「联系商家」",
            "refund": "**💳 关于退款**:\n1. 订单未制作可直接取消（全额退款）\n2. 已接单需联系商家协商\n3. 退款 1-7 个工作日原路返回",
            "quality": "**⭐ 关于菜品质量**:\n1. 请拍照记录问题\n2. 联系商家反馈处理\n3. 可在订单评价中如实反映",
            "price": "**💵 关于价格问题**:\n1. 请确认是否使用了优惠券\n2. 配送费根据距离可能调整\n3. 异常可联系平台客服核实",
            "general": "**💬 常见问题**:\n• 查订单 → 说「查一下我的订单」\n• 推荐菜品 → 说「推荐几个菜」\n• 配送进度 → 说「配送情况」",
        }
        for s in scenarios:
            if s in sol:
                parts.append(sol[s])
        return "\n\n".join(parts)
