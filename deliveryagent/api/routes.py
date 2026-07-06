"""
API Routes: FastAPI 路由定义
=========================================================
JD-Agent 的 RESTful API 层，提供核心端点：
  - POST /api/v1/chat          → 聊天（SSE 流式/非流式），session_key 格式 u:{user}:{id}
  - GET  /api/v1/health        → 健康检查
  - GET  /api/v1/history       → 获取指定会话的历史消息
  - GET  /api/v1/sessions      → 按用户列出会话（按天分组 + 最后一条预览）
  - POST /api/v1/proactive/lunch → 主动午餐推荐（时间驱动）
  - GET  /api/v1/profile       → 获取用户偏好画像

架构说明
--------
- AgentOrchestrator 全局单例，API 首次启动时自动初始化
- 会话 key 格式: u:{username}:{session_id}，按用户隔离
- 存储: JSONL 文件（sessions/u_{username}_{session_id}.jsonl）
"""
import base64
import json
import os
import uuid
from datetime import datetime
from pathlib import Path
from typing import Optional

import httpx
from fastapi import APIRouter, HTTPException, Header, Query
from fastapi.responses import StreamingResponse
from loguru import logger
from pydantic import BaseModel, Field

from agent.orchestrator import AgentOrchestrator
from agent.providers.openai_provider import OpenAIProvider
from agent.tools.tool_definitions import register_all_tools
from delivery_config import DELIVERY_BACKEND_URL, HTTP_TIMEOUT
from utils.config import load_database_config


# ─────────────────────────────────────────────────────────
# 辅助函数
# ─────────────────────────────────────────────────────────


def _get_username_from_jwt(jwt_token: str) -> str:
    """从 JWT token payload 中提取 username（不验证签名，信任 Go 后端）"""
    if not jwt_token:
        return ""
    try:
        payload_b64 = jwt_token.split(".")[1]
        padding = 4 - len(payload_b64) % 4
        if padding != 4:
            payload_b64 += "=" * padding
        payload = json.loads(base64.urlsafe_b64decode(payload_b64))
        return payload.get("username", "")
    except Exception:
        return ""


def _build_session_key(username: str, session_id: str) -> str:
    """构建用户隔离的会话 key: u:{username}:{session_id}"""
    user = username or "anonymous"
    return f"u:{user}:{session_id}"


def _extract_jwt_from_header(authorization: str) -> str:
    """从 Authorization 头提取 JWT token"""
    if not authorization:
        return ""
    if authorization.startswith("Bearer "):
        return authorization[len("Bearer "):]
    return authorization


# ─────────────────────────────────────────────────────────
# Pydantic Schemas（请求/响应模型）
# ─────────────────────────────────────────────────────────


class ChatRequest(BaseModel):
    """聊天请求体"""
    message: str = Field(..., description="用户输入的消息", min_length=1)
    session_id: Optional[str] = Field(None, description="会话 ID（不传则自动生成）")
    stream: bool = Field(False, description="是否使用 SSE 流式返回")


class ChatResponse(BaseModel):
    """聊天响应体（非流式）"""
    response: str = Field(..., description="AI 回复文本")
    session_id: str = Field(..., description="会话 ID")
    tools_used: list[str] = Field(default_factory=list, description="本轮用到的工具列表")


class HealthResponse(BaseModel):
    """健康检查响应"""
    status: str = "ok"
    version: str = "1.0.0"
    tools_count: int = 0
    skills_count: int = 0
    skills_names: list[str] = []
    sessions_count: int = 0
    pg_available: bool = False


class HistoryMessage(BaseModel):
    """单条历史消息"""
    role: str = Field(..., description="角色: user / assistant / tool")
    content: Optional[str] = Field(None, description="消息内容")
    timestamp: Optional[str] = Field(None, description="消息时间（ISO 格式）")
    tool_calls: Optional[list[dict]] = Field(None, description="工具调用信息")


class HistoryResponse(BaseModel):
    """历史消息响应"""
    history: list[HistoryMessage] = Field(default_factory=list, description="消息列表")
    total: int = 0


class SessionItem(BaseModel):
    """会话列表中的一项"""
    session_id: str = Field(..., description="会话 ID（传给 chat 接口用）")
    session_key: str = Field(..., description="完整的会话 key")
    date: str = Field(..., description="日期（YYYY-MM-DD）")
    last_message: str = Field("", description="最后一条消息的摘要")
    updated_at: str = Field("", description="最后更新时间")
    message_count: int = Field(0, description="会话中的消息总数")


class SessionsResponse(BaseModel):
    """会话列表响应，按日期分组"""
    groups: dict[str, list[SessionItem]] = Field(default_factory=dict, description="按日期分组的会话列表")


# ─────────────────────────────────────────────────────────
# AgentOrchestrator 全局单例
# ─────────────────────────────────────────────────────────

_orchestrator: AgentOrchestrator | None = None


async def get_orchestrator() -> AgentOrchestrator:
    """
    获取（并在首次时初始化）全局 AgentOrchestrator 单例。

    初始化流程：
      1. 确定 Provider（优先 DEEPSEEK_API_KEY / OPENAI_API_KEY，
         否则回退到通义千问）
      2. 创建 Orchestrator
      3. 初始化内部组件（AgentLoop、SessionManager、Memory 等）
      4. 注册内置工具
    """
    global _orchestrator
    if _orchestrator is not None:
        return _orchestrator

    # ── 1. 确定 LLM Provider ──
    api_key = (
        os.environ.get("DEEPSEEK_API_KEY")
        or os.environ.get("OPENAI_API_KEY")
        or os.environ.get("DEFAULT_API_KEY")
    )
    api_base = (
        os.environ.get("DEEPSEEK_BASE_URL")
        or os.environ.get("OPENAI_API_BASE")
        or "https://api.deepseek.com"
    )
    model = os.environ.get("DEFAULT_MODEL", "deepseek-chat")

    workspace = os.environ.get(
        "JD_AGENT_WORKSPACE",
        str(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
    )
    skills_dir = os.environ.get(
        "JD_AGENT_SKILLS_DIR",
        os.path.join(workspace, "skills")
    )

    if api_key:
        logger.info(f"[API] 使用 OpenAIProvider: model={model}, base={api_base}")
        provider = OpenAIProvider(api_key=api_key, api_base=api_base, model=model)
    else:
        logger.info("[API] 未配置 API Key，尝试使用通义千问")
        from agent.providers.tongyi_provider import TongyiProvider
        provider = TongyiProvider(model="qwen-plus")

    # ── 2. 尝试加载 PostgreSQL 配置（没有配置则继续用 JSONL）──
    pg_config = None
    try:
        pg_config = load_database_config()
        if pg_config:
            logger.info(f"[API] PostgreSQL 已配置，DSN={pg_config.get('dsn', '')[:30]}...")
    except Exception as e:
        logger.info(f"[API] PostgreSQL 未配置，使用 JSONL 存储: {e}")

    # ── 3. 创建 Orchestrator（自动初始化技能系统）──
    orchestrator = AgentOrchestrator(
        provider=provider,
        workspace=workspace,
        skills_dir=skills_dir,  # 技能自动发现目录
        pg_config=pg_config,    # PostgreSQL 配置（可选）
    )
    await orchestrator.initialize()

    # ── 4. 注册内置业务工具 ──
    registry = orchestrator.get_tool_registry()
    if registry is not None:
        register_all_tools(registry)
        logger.info(f"[API] 已注册 {len(registry)} 个业务工具")

    _orchestrator = orchestrator
    logger.info("[API] AgentOrchestrator 初始化完成")
    return orchestrator


# ─────────────────────────────────────────────────────────
# Router 定义
# ─────────────────────────────────────────────────────────

router = APIRouter()


# ─── Health ──────────────────────────────────────────────

@router.get("/health", response_model=HealthResponse)
async def health_check():
    """
    GET /api/v1/health — 健康检查端点

    返回服务状态、版本号、已注册工具数、活跃会话数。
    """
    orch = await get_orchestrator()
    tools = orch.get_all_tools()
    # 会话数（优先查 PG，回退到 JSONL 文件扫描）
    sessions_count = 0
    pg_available = False
    try:
        if orch.loop and orch.loop.sessions.pg_available:
            sessions_list = await orch.loop.sessions.alist_sessions()
            sessions_count = len(sessions_list)
            pg_available = True
        else:
            sessions_dir = orch.workspace / "sessions"
            if sessions_dir.exists():
                sessions_count = len(list(sessions_dir.glob("*.jsonl")))
    except Exception:
        pass
    # 技能系统信息
    skills_count = 0
    skills_names = []
    if orch.loop:
        skills_count = orch.loop.skill_manager.count
        skills_names = [s.name for s in orch.loop.skill_manager.get_all()]

    return HealthResponse(
        status="ok",
        version="1.0.0",
        tools_count=len(tools),
        skills_count=skills_count,
        skills_names=skills_names,
        sessions_count=sessions_count,
        pg_available=pg_available,
    )


# ─── Chat ────────────────────────────────────────────────

@router.post("/chat")
async def chat(
    request: ChatRequest,
    authorization: str = Header(""),
):
    """
    POST /api/v1/chat — 聊天接口

    支持两种模式：
      1. 非流式（stream=false）→ 返回 JSON { response, session_id, tools_used }
      2. 流式   （stream=true） → 返回 SSE 流，每行一个 data: chunk
         - 最终会收到 data: [DONE] 标记结束
         - 使用 &#x60;Accept: text/event-stream&#x60; 消费

    参数:
      - message:    用户消息（必填）
      - session_id: 会话 ID（选填，不传则自动生成）
      - stream:     是否流式（默认 false）

    认证:
      - Authorization: Bearer <jwt_token>（从登录接口获取，选填）
      - 不传则只能访问无需登录的接口
    """
    orch = await get_orchestrator()
    session_id = request.session_id or str(uuid.uuid4())[:8]

    # 从 Authorization 头提取 JWT token
    jwt_token = ""
    if authorization:
        if authorization.startswith("Bearer "):
            jwt_token = authorization[len("Bearer "):]
        else:
            jwt_token = authorization

    # 从 JWT 提取用户名，构建用户隔离的 session key
    username = _get_username_from_jwt(jwt_token)
    session_key = _build_session_key(username, session_id)

    # ── 非流式模式 ──
    if not request.stream:
        try:
            response_text = await orch.chat_async(
                request.message,
                session_key=session_key,
                jwt_token=jwt_token,
            )
            return ChatResponse(
                response=response_text,
                session_id=session_id,
                tools_used=[],  # 非流式下暂时不追踪工具列表
            )
        except Exception as e:
            logger.error(f"[API] 非流式聊天失败: {e}")
            raise HTTPException(status_code=500, detail=str(e))

    # ── 流式模式 (SSE) ──
    async def event_stream():
        try:
            async for chunk in orch.chat_stream_async(
                request.message,
                session_key=session_key,
                jwt_token=jwt_token,
            ):
                encoded = json.dumps({"text": chunk}, ensure_ascii=False)
                yield f"data: {encoded}\n\n"
        except Exception as e:
            logger.error(f"[API] 流式聊天失败: {e}")
            yield f"data: {json.dumps({'error': str(e)})}\n\n"
        finally:
            yield "data: [DONE]\n\n"

    return StreamingResponse(
        event_stream(),
        media_type="text/event-stream",
        headers={
            "Cache-Control": "no-cache",
            "Connection": "keep-alive",
            "X-Accel-Buffering": "no",
        },
    )


# ─── History ─────────────────────────────────────────────

@router.get("/history", response_model=HistoryResponse, response_model_exclude_none=True)
async def get_history(session_key: Optional[str] = Query(None, description="会话标识")):
    """
    GET /api/v1/history — 获取会话历史

    参数:
      - session_key: 会话标识（选填，默认返回 "cli:direct" 会话）
    """
    orch = await get_orchestrator()
    if orch.loop is None:
        return HistoryResponse(history=[], total=0)

    key = session_key or "cli:direct"
    try:
        session = await orch.loop.sessions.aget_or_create(key)
        history = session.get_history(max_messages=200)
        # 返回完整消息列表（包含 tool_calls），由前端自行过滤展示
        # 相比之前跳过 tool_calls 消息的策略，前端需要完整上下文
        cleaned = []
        for msg in history:
            entry: dict = {"role": msg.get("role", "")}
            if msg.get("content"):
                entry["content"] = msg["content"]
            if msg.get("timestamp"):
                ts = msg["timestamp"]
                if hasattr(ts, "isoformat"):
                    ts = ts.isoformat()
                elif isinstance(ts, (int, float)):
                    ts = datetime.fromtimestamp(ts).isoformat()
                elif isinstance(ts, str):
                    try:
                        num = float(ts)
                        ts = datetime.fromtimestamp(num).isoformat()
                    except (ValueError, OverflowError):
                        pass  # keep as-is (might be ISO string)
                entry["timestamp"] = str(ts)
            if msg.get("tool_calls"):
                # 兼容 PG JSONB 反序列化的类型差异
                tc = msg["tool_calls"]
                if isinstance(tc, str):
                    tc = json.loads(tc)
                if isinstance(tc, list):
                    entry["tool_calls"] = tc
            cleaned.append(entry)
        return HistoryResponse(history=cleaned, total=len(cleaned))
    except Exception as e:
        logger.error(f"[API] 获取历史失败: {e}")
        return HistoryResponse(history=[], total=0)


# ─── Sessions ─────────────────────────────────────────────


@router.get("/sessions", response_model=SessionsResponse)
async def list_sessions(
    authorization: str = Header(""),
):
    """
    GET /api/v1/sessions — 获取当前用户的会话列表（按天分组）

    从 JSONL 文件扫描当前用户的所有会话，读取元数据和最后一条消息，
    按日期分组返回，日期从新到旧排序。
    """
    jwt_token = ""
    if authorization:
        if authorization.startswith("Bearer "):
            jwt_token = authorization[len("Bearer "):]
        else:
            jwt_token = authorization

    username = _get_username_from_jwt(jwt_token)
    if not username:
        return SessionsResponse(groups={})

    orch = await get_orchestrator()

    # 优先走 PG，回退到 JSONL（SessionManager 内部处理）
    sessions_list: list[dict] = []
    if orch.loop and orch.loop.sessions:
        sessions_list = await orch.loop.sessions.alist_user_sessions(username)
    else:
        sessions_list = await _fallback_list_sessions(orch, username)

    # 构建 SessionItem 并按日期分组
    all_items: list[SessionItem] = []
    for s in sessions_list:
        raw_key = s.get("key", "")
        updated = s.get("updated_at", "") or ""
        date_str = updated[:10] if len(updated) >= 10 else (s.get("created_at", "") or "")[:10]
        if not date_str:
            date_str = "未知日期"

        # extract session_id from key: u:{user}:{session_id}
        parts = raw_key.split(":", 2)
        sid = parts[2] if len(parts) == 3 else raw_key

        all_items.append(SessionItem(
            session_id=sid,
            session_key=raw_key,
            date=date_str,
            last_message=(s.get("last_message", "") or "")[:120],
            updated_at=updated,
            message_count=s.get("message_count", 0) or 0,
        ))

    # 按日期分组
    groups: dict[str, list[SessionItem]] = {}
    for item in all_items:
        groups.setdefault(item.date, []).append(item)

    return SessionsResponse(groups=groups)


# ─── Proactive Lunch Recommendation ───────────────────────


@router.post("/proactive/lunch")
async def proactive_lunch(
    authorization: str = Header(""),
):
    """
    POST /api/v1/proactive/lunch — 主动午餐推荐（时间驱动）

    根据用户偏好和历史订单，用 LLM 生成个性化午餐推荐。
    前端应在 10:00-11:00 之间调用此接口。

    认证: Bearer <jwt_token>（必需）
    返回: { "message": "推荐文本" }
    """
    orch = await get_orchestrator()
    jwt_token = _extract_jwt_from_header(authorization)
    username = _get_username_from_jwt(jwt_token)
    if not username:
        raise HTTPException(status_code=401, detail="Unauthorized")

    # 1. 获取用户画像
    profile = {}
    if orch.loop and orch.loop.user_profiles:
        profile = orch.loop.user_profiles.get(username)

    # 2. 获取最近订单
    recent_orders_text = ""
    try:
        async with httpx.AsyncClient(timeout=HTTP_TIMEOUT) as client:
            resp = await client.get(
                f"{DELIVERY_BACKEND_URL}/api/user/order/list",
                headers={"Authorization": f"Bearer {jwt_token}"},
                params={"page": 1, "pageSize": 5},
                follow_redirects=True,
            )
            if resp.status_code == 200:
                data = resp.json()
                orders = data.get("data") if isinstance(data, dict) else data
                if isinstance(orders, list) and orders:
                    recent_orders_text = _format_orders_for_prompt(orders)
    except Exception as e:
        logger.warning(f"[Proactive] 获取订单失败: {e}")

    # 3. 获取商家列表
    stores_text = ""
    try:
        async with httpx.AsyncClient(timeout=HTTP_TIMEOUT) as client:
            resp = await client.get(
                f"{DELIVERY_BACKEND_URL}/api/user/stores",
                follow_redirects=True,
            )
            if resp.status_code == 200:
                data = resp.json()
                stores = data.get("data") if isinstance(data, dict) else data
                if isinstance(stores, list) and stores:
                    lines = []
                    for s in stores[:8]:
                        name = s.get("name") or s.get("shop_name") or ""
                        sid = s.get("id") or s.get("base_id") or ""
                        rating = s.get("avg_score") or s.get("rating") or ""
                        lines.append(f"  - ID {sid}: {name} (评分: {rating})")
                    stores_text = "\n".join(lines)
    except Exception as e:
        logger.warning(f"[Proactive] 获取商家失败: {e}")

    # 4. 构造 Prompt + 调用 LLM
    sys_prompt = "你是一位贴心的外卖助手，擅长根据用户的饮食偏好和点餐历史推荐合适的午餐。\n\n请用温暖亲切的语气推荐，要求：\n1. 推荐 1-2 家店铺及其具体菜品\n2. 说明推荐理由（基于用户的偏好或历史订单）\n3. 提及价格\n4. 最后询问用户是否要下单\n5. 控制在 150 字以内"

    has_profile = any(v for v in profile.values() if isinstance(v, list) and v) or (profile.get("notes") or "").strip()
    profile_text = json.dumps(profile, ensure_ascii=True) if has_profile else "暂无（新用户）"
    user_prompt = (
        f"当前时间：午餐时间（约 10:30）\n\n"
        f"## 用户偏好\n{profile_text}\n\n"
        f"## 最近订单\n{recent_orders_text or '暂无'}\n\n"
        f"## 可用商家\n{stores_text or '暂无'}\n\n请推荐午餐。"
    )

    try:
        if orch.loop and orch.loop.provider:
            response = await orch.loop.provider.chat_with_retry(
                messages=[
                    {"role": "system", "content": sys_prompt},
                    {"role": "user", "content": user_prompt},
                ],
                model=orch.loop.model,
            )
            message = response.content or "推荐生成失败。"
        else:
            message = "😋 午餐时间到！看看今天有什么好吃的吧～"
    except Exception as e:
        logger.error(f"[Proactive] LLM 调用失败: {e}")
        message = "😋 午餐时间到！今天可以看看有哪些好吃的～打开商家列表挑选一下吧！"

    return {"message": message}


# ─── User Profile ─────────────────────────────────────────


@router.get("/profile")
async def get_profile(
    authorization: str = Header(""),
):
    """
    GET /api/v1/profile — 获取当前用户的偏好画像

    返回用户存储的饮食偏好、忌口等个性化信息。
    """
    jwt_token = _extract_jwt_from_header(authorization)
    username = _get_username_from_jwt(jwt_token)
    if not username:
        raise HTTPException(status_code=401, detail="Unauthorized")

    orch = await get_orchestrator()
    if orch.loop and orch.loop.user_profiles:
        profile = orch.loop.user_profiles.get(username)
        return {"profile": profile}
    return {"profile": {}}


# ── 辅助：格式化订单信息为 LLM 可读文本 ──


def _format_orders_for_prompt(orders: list) -> str:
    """将订单列表简化为 LLM 可读的文本摘要。"""
    lines = []
    for o in orders[:5]:
        merchant = o.get("merchantName") or o.get("merchant_name") or o.get("shopName") or ""
        dishes = o.get("dishes") or o.get("items") or []
        dish_names = []
        for d in (dishes if isinstance(dishes, list) else []):
            name = d.get("name") or d.get("dishName") or ""
            qty = d.get("qty") or d.get("num") or 1
            if name:
                dish_names.append(f"{name}x{qty}")
        total = o.get("totalPrice") or o.get("total_price") or 0
        lines.append(f"  - [{merchant}] {'、'.join(dish_names)} (¥{total})")
    return "\n".join(lines) if lines else ""


async def _fallback_list_sessions(orch: AgentOrchestrator, username: str) -> list[dict]:
    """JSONL 文件扫描回退方案"""
    sessions_dir = orch.workspace / "sessions"
    prefix = f"u_{username}_"
    results: list[dict] = []

    for fpath in sorted(sessions_dir.glob(f"{prefix}*.jsonl"), reverse=True):
        try:
            with open(fpath, encoding="utf-8") as f:
                first = f.readline().strip()
                meta = json.loads(first) if first else {}
                preview = ""
                msg_count = 0
                for line in f:
                    line = line.strip()
                    if not line:
                        continue
                    try:
                        msg = json.loads(line)
                        if msg.get("role") == "user":
                            msg_count += 1
                        if msg.get("role") in ("user", "assistant"):
                            content = msg.get("content", "")
                            if content:
                                preview = content[:120]
                    except json.JSONDecodeError:
                        continue
            results.append({
                "key": meta.get("key", ""),
                "created_at": meta.get("created_at", ""),
                "updated_at": meta.get("updated_at", ""),
                "last_message": preview,
                "message_count": msg_count,
            })
        except Exception as e:
            logger.debug(f"[API] 读取会话文件失败 {fpath.name}: {e}")
            continue
    return results
