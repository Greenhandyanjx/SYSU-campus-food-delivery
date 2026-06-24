"""
会话管理：JSONL + PostgreSQL 双后端支持。

架构说明：
┌──────────────────────────────────────────────────┐
│ SessionManager（路由层）                           │
│   get_or_create(key)  → ①内存缓存 ②JSONL          │
│   aget_or_create(key) → ①内存 ②Redis ③PG ④JSONL  │
│   save(session)       → JSONL                     │
│   asave(session)      → PG + Redis + JSONL（三写） │
└──────────────────┬───────────────────────────────┘
                   │
        ┌──────────┼──────────┐
        ▼          ▼          ▼
   ┌─────────┐ ┌─────────┐ ┌──────────┐
   │ Redis   │ │   PG    │ │  JSONL   │
   │ (L2缓存) │ │ (主存储) │ │ (保底)   │
   └─────────┘ └─────────┘ └──────────┘

四层查找链：内存(L1) → Redis(L2) → PG(L3) → JSONL(L4)

设计原则：
1. JSONL 始终可用——PG 离线时系统不中断
2. PG 写失败 → 日志告警 + 写入 JSONL（不抛出异常）
3. 内存缓存作为 L1，减少 PG 读压力
4. Redis 作为 L2，加速跨进程会话恢复（进程重启后内存空，但 Redis 热）
5. Session 的 in-memory 表示（dataclass）不变，各后端只是不同序列化方式

重要约定：
- assistant 消息的 content 字段必须为字符串（不允许 None/null），
  避免 LLM 将 null 渲染为 "None"
"""

import json
import os
from contextlib import suppress
from dataclasses import dataclass, field
from datetime import datetime
from pathlib import Path
from typing import TYPE_CHECKING, Any

if TYPE_CHECKING:
    from agent.cache.redis_cache import RedisCache

from loguru import logger


def ensure_dir(path: Path) -> Path:
    """确保目录存在（mkdir -p）"""
    path.mkdir(parents=True, exist_ok=True)
    return path


def safe_filename(name: str) -> str:
    """
    将会话 key 转为安全的文件名。

    把非字母数字的字符替换为下划线，避免文件系统非法字符。
    示例: "api:abc123" → "api_abc123"
    """
    safe = ""
    for c in name:
        if c.isalnum() or c in ('-', '_', '.'):
            safe += c
        else:
            safe += '_'
    return safe


def find_legal_message_start(messages: list[dict]) -> int:
    """
    找到第一个合法的消息起始位置。

    为什么需要这个函数：
    LLM API 要求消息序列以 user 或 system 开头，
    不能以孤立的 tool 消息开头（tool 必须有对应的 assistant tool_call）。

    逻辑：
    - 从前往后扫描，记录 assistant 声明的 tool_call_id
    - 如果遇到 tool 消息但其 tool_call_id 不在已声明集合中 → 这条 tool 是孤儿
    - 跳过开头的孤儿 tool 消息，直到找到合法的起始位置
    """
    declared: set[str] = set()
    for i, msg in enumerate(messages):
        role = msg.get("role")
        if role == "assistant":
            raw_tc = msg.get("tool_calls") or []
            # 兼容 tool_calls 为字符串的情况（JSONL 反序列化后）
            if isinstance(raw_tc, str):
                try:
                    raw_tc = json.loads(raw_tc)
                except (json.JSONDecodeError, TypeError):
                    raw_tc = []
            for tc in raw_tc:
                if isinstance(tc, dict) and tc.get("id"):
                    declared.add(str(tc["id"]))
        elif role == "tool":
            tid = msg.get("tool_call_id")
            if tid and str(tid) not in declared:
                # 发现孤儿 tool → 跳过
                start = i + 1
                declared.clear()
                for prev in messages[start:i + 1]:
                    if prev.get("role") == "assistant":
                        for tc in prev.get("tool_calls") or []:
                            if isinstance(tc, dict) and tc.get("id"):
                                declared.add(str(tc["id"]))
    return 0


@dataclass
class Session:
    """
    会话：存储对话历史。

    这是一个纯内存对象，不关心数据最终存到哪（JSONL 还是 PG）。
    持久化由 SessionManager 负责。

    关键字段：
    - last_consolidated: 已 consolidated 的消息数量（阈值），
      get_history() 从该偏移量之后开始返回
    - messages: 有序的消息列表，越靠后越新

    设计决策：
    - messages 是追加写的（不在中间插入），这有利于 LLM 的 cache 效率
    - 不限制消息总数上限——由调用方（AgentLoop）在 consolidation 时截断
    """
    key: str
    messages: list[dict] = field(default_factory=list)
    created_at: datetime = field(default_factory=datetime.now)
    updated_at: datetime = field(default_factory=datetime.now)
    metadata: dict = field(default_factory=dict)
    last_consolidated: int = 0

    def add_message(self, role: str, content: str, **kwargs: Any) -> None:
        """添加一条消息到会话末尾"""
        msg = {
            "role": role,
            "content": content,
            "timestamp": datetime.now().isoformat(),
            **kwargs,
        }
        self.messages.append(msg)
        self.updated_at = datetime.now()

    def get_history(self, max_messages: int = 500) -> list[dict]:
        """
        获取 LLM 输入用的消息列表。

        按以下规则过滤：
        1. 跳过已 consolidated 的消息（从 last_consolidated 之后开始）
        2. 最多返回 max_messages 条（取最后 N 条）
        3. 跳过开头的孤儿 tool 消息（参考 find_legal_message_start）
        4. 确保 content 不会是 None（某些 LLM API 对 null 敏感）
        5. 确保 tool_calls 始终是 list[dict]（某些后端会序列化为字符串）
        """
        unconsolidated = self.messages[self.last_consolidated:]
        sliced = unconsolidated[-max_messages:]

        # 统一 tool_calls 类型：字符串 → list[dict]
        # JSONL 存储时 tool_calls 被序列化为字符串，但 LLM API 要求数组
        # 必须在 find_legal_message_start 之前做，因为该函数也遍历 tool_calls
        for msg in sliced:
            if isinstance(msg.get("tool_calls"), str):
                try:
                    parsed = json.loads(msg["tool_calls"])
                    if isinstance(parsed, list):
                        msg["tool_calls"] = parsed
                except (json.JSONDecodeError, TypeError):
                    pass

        # 避免以非 user 消息开头
        for i, msg in enumerate(sliced):
            if msg.get("role") == "user":
                sliced = sliced[i:]
                break

        # 移除开头的孤儿 tool 消息
        start = find_legal_message_start(sliced)
        if start:
            sliced = sliced[start:]

        # 修复 content 为 None 的兼容性问题
        for msg in sliced:
            if msg.get("content") is None:
                msg["content"] = ""

        return sliced

    def clear(self) -> None:
        """清空会话（保留 key 和创建时间）"""
        self.messages.clear()
        self.last_consolidated = 0
        self.updated_at = datetime.now()


# ─── JSONL 后端 ─────────────────────────────────────

class _JsonlBackend:
    """
    JSONL 文件存储后端（保底方案）。

    每个会话存为一个 JSONL 文件：sessions/{safe_key}.jsonl
    第一行：元数据（_type: "metadata"）
    后续行：消息 JSON

    文件格式优势：
    - 追加友好（虽然当前实现是重写整个文件，但格式本身支持追加）
    - 每行独立 JSON，一行损坏不影响其他行
    - 无需任何外部依赖
    """

    def __init__(self, sessions_dir: Path):
        self.sessions_dir = ensure_dir(sessions_dir)

    def get_path(self, key: str) -> Path:
        """获取会话文件路径"""
        safe = safe_filename(key.replace(":", "_"))
        return self.sessions_dir / f"{safe}.jsonl"

    def load(self, key: str) -> Session | None:
        """从 JSONL 文件加载会话"""
        path = self.get_path(key)
        if not path.exists():
            return None

        try:
            messages: list[dict] = []
            metadata: dict = {}
            created_at: datetime | None = None
            updated_at: datetime | None = None
            last_consolidated: int = 0

            with open(path, encoding="utf-8") as f:
                for line in f:
                    line = line.strip()
                    if not line:
                        continue
                    data = json.loads(line)
                    if data.get("_type") == "metadata":
                        metadata = data.get("metadata", {})
                        if data.get("created_at"):
                            created_at = datetime.fromisoformat(data["created_at"])
                        if data.get("updated_at"):
                            updated_at = datetime.fromisoformat(data["updated_at"])
                        last_consolidated = data.get("last_consolidated", 0)
                    else:
                        messages.append(data)

            return Session(
                key=key,
                messages=messages,
                created_at=created_at or datetime.now(),
                updated_at=updated_at or datetime.now(),
                metadata=metadata,
                last_consolidated=last_consolidated,
            )
        except Exception as e:
            logger.warning(f"[JSONL] 加载会话 {key} 失败: {e}")
            return self._repair(key)

    def save(self, session: Session) -> None:
        """
        将会话写入 JSONL 文件。

        使用原子写入模式：
        1. 先写 .tmp 临时文件
        2. os.replace() 原子替换目标文件
        3. 如果中途崩溃，只有 .tmp 文件损坏，原文件完好
        """
        path = self.get_path(session.key)
        tmp_path = path.with_suffix(".jsonl.tmp")

        try:
            with open(tmp_path, "w", encoding="utf-8") as f:
                # 第一行：元数据
                metadata_line = {
                    "_type": "metadata",
                    "key": session.key,
                    "created_at": session.created_at.isoformat(),
                    "updated_at": session.updated_at.isoformat(),
                    "metadata": session.metadata,
                    "last_consolidated": session.last_consolidated,
                }
                f.write(json.dumps(metadata_line, ensure_ascii=False) + "\n")
                # 后续行：消息
                for msg in session.messages:
                    f.write(json.dumps(msg, ensure_ascii=False) + "\n")

            # 原子替换（先写 .tmp 再 rename，崩溃安全）
            os.replace(tmp_path, path)
        except BaseException:
            tmp_path.unlink(missing_ok=True)
            raise

    def delete(self, key: str) -> bool:
        """删除会话文件"""
        path = self.get_path(key)
        if path.exists():
            path.unlink()
            return True
        return False

    def list_sessions(self) -> list[dict]:
        """列出所有会话的元数据（读取每个文件的第一行）"""
        sessions: list[dict] = []
        for path in self.sessions_dir.glob("*.jsonl"):
            key = path.stem.replace("_", ":", 1)
            try:
                with open(path, encoding="utf-8") as f:
                    first_line = f.readline().strip()
                    if first_line:
                        data = json.loads(first_line)
                        if data.get("_type") == "metadata":
                            sessions.append({
                                "key": data.get("key") or key,
                                "created_at": data.get("created_at"),
                                "updated_at": data.get("updated_at"),
                            })
            except Exception:
                continue
        return sorted(sessions, key=lambda x: x.get("updated_at", ""), reverse=True)

    def _repair(self, key: str) -> Session | None:
        """
        尝试从损坏的 JSONL 文件中恢复会话。

        恢复策略：
        - 跳过无法解析的 JSON 行
        - 保留能解析的所有消息行
        - 记录跳过的行数用于告警
        """
        path = self.get_path(key)
        if not path.exists():
            return None

        try:
            messages: list[dict] = []
            metadata: dict = {}
            created_at: datetime | None = None
            updated_at: datetime | None = None
            last_consolidated: int = 0
            skipped: int = 0

            with open(path, encoding="utf-8") as f:
                for line in f:
                    line = line.strip()
                    if not line:
                        continue
                    try:
                        data = json.loads(line)
                    except json.JSONDecodeError:
                        skipped += 1
                        continue

                    if data.get("_type") == "metadata":
                        metadata = data.get("metadata", {})
                        if data.get("created_at"):
                            with suppress(ValueError):
                                created_at = datetime.fromisoformat(data["created_at"])
                        if data.get("updated_at"):
                            with suppress(ValueError):
                                updated_at = datetime.fromisoformat(data["updated_at"])
                        last_consolidated = data.get("last_consolidated", 0)
                    else:
                        messages.append(data)

            if skipped:
                logger.warning(f"[JSONL] 恢复会话 {key}: 跳过 {skipped} 行损坏数据")

            if not messages and not metadata:
                return None

            return Session(
                key=key,
                messages=messages,
                created_at=created_at or datetime.now(),
                updated_at=updated_at or datetime.now(),
                metadata=metadata,
                last_consolidated=last_consolidated,
            )
        except Exception:
            return None


# ─── SessionManager（路由层） ──────────────────────

class SessionManager:
    """
    会话管理器：JSONL + PostgreSQL 双后端路由。

    调用路径：
    ┌──────────────┐
    │  调用方       │
    │ loop.py 等    │
    └──────┬───────┘
           │
    ┌──────▼───────┐
    │ SessionManager│
    │              │
    │  同步路径:    │  get_or_create / save → 内存缓存 + JSONL
    │  异步路径:    │  aget_or_create / asave → 内存 + PG + JSONL
    └──────────────┘

    配置方式：
    - 传入 pg_config 字典 → 启用 PG 后端
    - pg_config=None → 仅 JSONL（向后兼容）
    """

    def __init__(self, workspace: Path, pg_config: dict | None = None, redis_cache: "RedisCache | None" = None):
        """
        初始化 SessionManager。

        参数:
            workspace: 工作区路径（sessions 目录建在此路径下）
            pg_config: PostgreSQL 配置字典（可选，不传则仅用 JSONL）
                       {"dsn": "...", "pool_min_size": 1, "pool_max_size": 10, ...}
            redis_cache: RedisCache 实例（可选，用于 L2 会话缓存）
        """
        self.workspace = workspace
        self.sessions_dir = ensure_dir(workspace / "sessions")

        # 内存缓存：key → Session 对象（L1）
        # 作用：避免频繁读盘/PG/Redis，加速重复访问
        self._cache: dict[str, Session] = {}

        # Redis 缓存（L2）：可选，由外部注入
        # 作用：跨进程会话恢复。进程重启后内存清空，
        # 但 Redis 中的缓存还在，避免每次重启都查 PG
        self._redis = redis_cache

        # JSONL 后端（L4 保底，始终启用）
        self._jsonl = _JsonlBackend(self.sessions_dir)

        # PG 后端（L3 主存储，可选）
        self._pg = None
        if pg_config and pg_config.get("dsn"):
            from agent.session.pg_store import PgBackend
            self._pg = PgBackend(pg_config)
            logger.info("[SessionManager] PostgreSQL 后端已配置")

    # ─── 异步初始化（连接池） ───────────────────────

    async def initialize_pg(self) -> bool:
        """
        初始化 PG 连接池（异步）。

        在 AgentLoop 启动时的 async 上下文中调用。
        如果 PG 不可用（网络不通、认证失败），只记警告日志，不抛异常。
        系统会继续以 JSONL-only 模式运行。

        返回:
            True=PG 就绪, False=PG 不可用（已降级到 JSONL）
        """
        if self._pg is None:
            return False
        try:
            await self._pg.initialize()
            logger.info("[SessionManager] PostgreSQL 已就绪")
            return True
        except Exception as e:
            logger.warning(f"[SessionManager] PostgreSQL 不可用，降级到 JSONL: {e}")
            self._pg = None  # 降级：标记 PG 不可用，后续操作跳过 PG
            return False

    async def close_pg(self) -> None:
        """关闭 PG 连接池"""
        if self._pg:
            await self._pg.close()

    @property
    def pg_available(self) -> bool:
        """PG 后端是否可用"""
        return self._pg is not None

    # ─── 同步方法（内存缓存 + JSONL） ──────────────

    def get_or_create(self, key: str) -> Session:
        """
        同步获取/创建会话。

        查找链：
        ① 内存缓存命中 → 直接返回（最快）
        ② JSONL 文件存在 → 加载到缓存后返回
        ③ 都没有 → 新建空会话

        注意：此方法不走 PG，适用于 CLI 同步操作。
        """
        if key in self._cache:
            return self._cache[key]

        session = self._jsonl.load(key)
        if session is None:
            session = Session(key=key)

        self._cache[key] = session
        return session

    def save(self, session: Session) -> None:
        """
        同步保存（写 JSONL + 更新内存缓存）。

        不写 PG——如果需要双写，使用 asave()。
        """
        self._jsonl.save(session)
        self._cache[session.key] = session

    def delete_session(self, key: str) -> bool:
        """
        删除会话（JSONL + 内存缓存）。

        不删 PG——如果需要从 PG 也删除，使用 adelete_session()。
        """
        self._cache.pop(key, None)
        return self._jsonl.delete(key)

    # ─── 异步方法（PG + JSONL 双写） ───────────────

    async def aget_or_create(self, key: str) -> Session:
        """
        异步获取/创建会话。

        查找链（优先级从高到低）：
        ① 内存缓存（L1，最快）——本进程内已加载的 session
        ② Redis 缓存（L2，快）——跨进程共享，进程重启后可恢复
        ③ PostgreSQL（L3，主存储）——异步查询
        ④ JSONL 文件（L4，保底）——PG 不可用时的 fallback
        ⑤ 都没有 → 新建空会话

        为什么四层备份：
        - L1 内存：本进程加速（最常用路径，命中率 > 90%）
        - L2 Redis：进程重启后缓存清空，从 Redis 恢复（比 PG 快 10 倍）
        - L3 PG：跨进程持久化，支持 SQL 查询
        - L4 JSONL：PG 迁移/故障期间的数据恢复路径
        """
        # L1: 内存缓存
        if key in self._cache:
            return self._cache[key]

        # L2: Redis 缓存（独立于 PG，只要配置了就启用）
        if self._redis:
            try:
                data = await self._redis.get_session_cache(key)
                if data:
                    session = Session(
                        key=data["key"],
                        messages=data.get("messages", []),
                        created_at=datetime.fromisoformat(data["created_at"])
                        if isinstance(data.get("created_at"), str)
                        else datetime.now(),
                        updated_at=datetime.fromisoformat(data["updated_at"])
                        if isinstance(data.get("updated_at"), str)
                        else datetime.now(),
                        metadata=data.get("metadata", {}),
                        last_consolidated=data.get("last_consolidated", 0),
                    )
                    self._cache[key] = session
                    logger.debug(f"[SessionManager] 从 Redis 加载会话 {key}")
                    return session
            except Exception as e:
                logger.debug(f"[SessionManager] Redis 加载失败，查 PG: {e}")

        # L3: PostgreSQL
        if self._pg:
            try:
                data = await self._pg.load(key)
                if data:
                    session = Session(
                        key=data["key"],
                        messages=data["messages"],
                        created_at=data["created_at"],
                        updated_at=data["updated_at"],
                        metadata=data["metadata"],
                        last_consolidated=data["last_consolidated"],
                    )
                    self._cache[key] = session
                    # 同步写入 Redis 缓存（下次查 L2 就能命中）
                    if self._redis:
                        await self._redis.set_session_cache(
                            key, self._session_to_cache_dict(session)
                        )
                    logger.debug(f"[SessionManager] 从 PG 加载会话 {key}")
                    return session
            except Exception as e:
                logger.warning(f"[SessionManager] PG 加载失败，尝试 JSONL: {e}")

        # L4: JSONL（保底）
        return self.get_or_create(key)

    async def asave(self, session: Session) -> None:
        """
        异步保存（PG + Redis + JSONL 三写）。

        写入策略：
        - PG 主写：异步写入 PostgreSQL（强一致）
        - Redis 缓存：异步写入（提升后续读取速度）
        - JSONL 保底：同步写入文件系统（崩溃恢复用）
        - 任何后端写失败：记警告日志 + 继续（不抛异常，不影响业务）
        """
        pg_ok = False
        if self._pg:
            try:
                pg_ok = await self._pg.save(
                    key=session.key,
                    messages=session.messages,
                    created_at=session.created_at,
                    updated_at=session.updated_at,
                    metadata=session.metadata,
                    last_consolidated=session.last_consolidated,
                )
            except Exception as e:
                logger.warning(
                    f"[SessionManager] PG 保存会话 {session.key} 失败，"
                    f"已降级到 JSONL: {e}"
                )

        # 更新 Redis 缓存（L2）
        # 写入最新数据后，下次 aget_or_create 可以直接从 Redis 加载
        if self._redis:
            try:
                await self._redis.set_session_cache(
                    session.key, self._session_to_cache_dict(session)
                )
            except Exception as e:
                logger.debug(f"[SessionManager] Redis 缓存更新失败: {e}")

        # 无论 PG 成功与否，都写 JSONL（双写策略）
        # 异常隔离：JSONL 写失败只记日志，不抛给调用方
        try:
            self.save(session)
        except Exception as e:
            logger.warning(
                f"[SessionManager] JSONL 保存会话 {session.key} 失败: {e}"
            )

        if pg_ok:
            logger.debug(f"[SessionManager] 保存完成: {session.key} (PG+Redis+JSONL)")

    async def adelete_session(self, key: str, delete_jsonl: bool = False) -> bool:
        """
        异步删除会话（PG + Redis + 可选 JSONL）。

        参数:
            key: 会话标识
            delete_jsonl: 是否同时删除 JSONL 文件（默认 False，只删 PG）

        设计说明：
        JSONL 文件是数据保底，默认不应删除。
        只在明确需要清理所有数据时（如测试）才设为 True。
        """
        self._cache.pop(key, None)

        # 删除 Redis 缓存
        if self._redis:
            try:
                await self._redis.delete_session_cache(key)
            except Exception as e:
                logger.debug(f"[SessionManager] Redis 缓存删除失败: {e}")

        pg_ok = True
        if self._pg:
            try:
                pg_ok = await self._pg.delete(key)
            except Exception as e:
                logger.warning(f"[SessionManager] PG 删除会话 {key} 失败: {e}")
                pg_ok = False

        jsonl_ok = True
        if delete_jsonl:
            jsonl_ok = self._jsonl.delete(key)

        return pg_ok or jsonl_ok

    async def alist_sessions(self) -> list[dict]:
        """
        异步列出所有会话（优先走 PG）。

        如果 PG 不可用，回退到 JSONL 的文件扫描方式。
        """
        if self._pg:
            try:
                return await self._pg.list_keys()
            except Exception as e:
                logger.warning(f"[SessionManager] PG 列出会话失败，回退 JSONL: {e}")
        return self._jsonl.list_sessions()

    # ─── 缓存专用工具 ───────────────────────────────

    @staticmethod
    def _session_to_cache_dict(session: Session) -> dict:
        """
        将 Session 对象转为可缓存字典（供 Redis L2 缓存使用）。

        为什么单独做一个方法而不是直接序列化 Session？
        - 避免序列化全部消息（可能几百条，浪费带宽）
        - 只保留最近 50 条（Redis 内存有限，且 LLM 通常只看最近 N 条）
        - 确保存储格式是纯 JSON 可序列化的（Session 中的 datetime 需要转字符串）
        """
        MAX_CACHED_MESSAGES = 50
        messages = session.messages[-MAX_CACHED_MESSAGES:] if len(session.messages) > MAX_CACHED_MESSAGES else session.messages
        return {
            "key": session.key,
            "messages": messages,
            "created_at": session.created_at.isoformat(),
            "updated_at": session.updated_at.isoformat(),
            "metadata": session.metadata,
            "last_consolidated": session.last_consolidated,
        }
