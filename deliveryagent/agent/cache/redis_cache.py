"""
Redis 缓存实现 — 四层缓存的核心引擎

架构决策
========
1. 为什么同时支持 sync/async？
   - async: LLM 缓存、Rate Limiter、Session 缓存（都在 async event loop 中运行）
   - sync:  RAG 缓存（现有方法 retrieve() 是同步的，暂时不重构为 async）
   - 两个 client 共享同一套连接配置，但连接池独立（这是生产中的常见模式）

2. 连接池管理
   - async 用 redis.asyncio.ConnectionPool（被 asyncio 管理）
   - sync  用 redis.ConnectionPool（在同步代码中共享）
   - 都支持自动重连（retry_on_timeout=True）

3. 序列化
   - 所有 value 通过 JSON 序列化，保证跨语言兼容
   - 对 Python 对象自动处理 datetime、set 等特殊类型

4. 命名空间
   - 所有 key 自动拼接 key_prefix，避免多应用共用 Redis 时冲突
   - 最终 key 格式: {prefix}{namespace}:{原始 key}
   - 例如: jdagent:llm:550e8400...

面试亮点：
  - 滑动窗口限流（Redis Sorted Set）
  - 缓存穿透防护
  - 连接池复用
  - 优雅降级（Redis 不可用时静默降级，不阻塞业务）
"""


# ═══════════════════════════════════════════════════════════
# Lua 脚本：原子滑动窗口限流
# ═══════════════════════════════════════════════════════════
#
# 为什么需要这个脚本？
# 原始的 check() 方法用 3-4 条 Redis 命令实现限流：
#   ZREMRANGEBYSCORE → ZCARD → [ZADD → EXPIRE]
# 这些命令不是原子的：并发请求可能同时读到 ZCARD=0，全部放行。
#
# Lua 脚本把全部逻辑包在一个脚本里。
# Redis 保证 Lua 脚本执行期间不会有其他命令插入——这是真正的原子性。
#
# 参数：
#   KEYS[1] — 限流 key（已加前缀）
#   ARGV[1] — 当前时间戳（毫秒）
#   ARGV[2] — 窗口大小（毫秒）
#   ARGV[3] — 限制次数
#   ARGV[4] — 当前请求的 member ID
#
# 返回值：
#   [1, remaining] — 放行，remaining 是剩余次数
#   [0, current]   — 拒绝，current 是当前计数
# ═══════════════════════════════════════════════════════════

_RATE_LIMIT_SCRIPT = """
redis.replicate_commands()

local key = KEYS[1]
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])
local member = ARGV[4]

-- 1. 清理窗口外的过期数据
redis.call('ZREMRANGEBYSCORE', key, 0, now - window)

-- 2. 统计当前窗口内请求数
local current = redis.call('ZCARD', key)

-- 3. 判断是否超限
if current >= limit then
    return {0, current}
end

-- 4. 记录本次请求 + 设 TTL
redis.call('ZADD', key, now, member)
redis.call('EXPIRE', key, math.floor(window / 1000))

return {1, current + 1}
"""


# ═══════════════════════════════════════════════════════════
# Lua 脚本：分布式锁（SET NX + 看门狗）
# ═══════════════════════════════════════════════════════════
#
# 分布式锁用于"缓存击穿"防护：
# 当缓存 miss 时，只有拿到锁的请求才去调用 LLM API，
# 其他请求等待并轮询缓存。
#
# 为什么不用 Python 的 asyncio.Lock？
# asyncio.Lock 只在一个进程内有效。
# 如果部署了多个进程（或多台机器），一个进程的锁挡不住其他进程的请求。
# Redis 分布式锁对所有进程可见。
#
# 锁释放：
# 1. 正常释放：拿到锁的请求完成工作后主动 DEL
# 2. 异常释放：锁有 TTL（5 秒），持有者崩溃后自动过期
# ═══════════════════════════════════════════════════════════

_LOCK_SCRIPT = """
-- KEYS[1] = 锁 key
-- ARGV[1] = 锁 value（唯一标识，防止误删别人的锁）
-- ARGV[2] = TTL（秒）

local existed = redis.call('SET', KEYS[1], ARGV[1], 'NX', 'EX', ARGV[2])
if existed then
    return 1
end
return 0
"""

_UNLOCK_SCRIPT = """
-- KEYS[1] = 锁 key
-- ARGV[1] = 锁 value

-- 只删除属于自己的锁（value 匹配才删）
-- 防止因执行超时导致锁被其他进程误删
if redis.call('GET', KEYS[1]) == ARGV[1] then
    return redis.call('DEL', KEYS[1])
end
return 0
"""

import hashlib
import json
import os
import time
from pathlib import Path
from typing import Any, Optional

from loguru import logger

from agent.cache.backend import CacheBackend


# ─── 工具：序列化辅助 ──────────────────────────────

class _EnhancedEncoder(json.JSONEncoder):
    """
    增强型 JSON 编码器，支持更多 Python 类型。

    标准 JSON 不支持的 datetime、set、bytes 等类型，
    在这里统一转为字符串，避免序列化崩溃。
    """
    def default(self, obj: Any) -> str:
        # datetime → ISO 格式字符串
        if hasattr(obj, "isoformat"):
            return obj.isoformat()
        # set → list
        if isinstance(obj, set):
            return list(obj)
        # bytes → utf-8 字符串
        if isinstance(obj, bytes):
            return obj.decode("utf-8", errors="replace")
        # 其他未知类型 → repr（至少不崩溃）
        return repr(obj)


def _json_dumps(obj: Any) -> str:
    """带增强编码的 JSON 序列化（确保不抛异常）"""
    return json.dumps(obj, cls=_EnhancedEncoder, ensure_ascii=False)


def _json_loads(data: str | bytes) -> Any:
    """JSON 反序列化（失败返回 None 而不是抛异常）"""
    try:
        return json.loads(data)
    except (json.JSONDecodeError, TypeError, ValueError):
        return None


# ─── 配置加载 ──────────────────────────────────────

def get_redis_config(config_path: str | Path | None = None) -> dict:
    """
    从 YAML 文件加载 Redis 配置。

    优先级（从高到低）：
    1. 环境变量 REDIS_URL（用于 Docker/K8s 环境注入）
    2. YAML 配置文件
    3. 默认值（localhost:6379）

    返回:
        {
            "url": "redis://...",
            "pool_max_size": 20,
            "pool_timeout": 5.0,
            ...
        }
    """
    if config_path is None:
        # 默认搜索路径：项目根目录下的 config/cache.yml
        for candidate in [
            Path.cwd() / "config" / "cache.yml",
            Path(__file__).parent.parent.parent / "config" / "cache.yml",
        ]:
            if candidate.exists():
                config_path = candidate
                break

    config: dict = {"url": "redis://localhost:6379/0"}

    if config_path and Path(config_path).exists():
        try:
            import yaml
            raw = yaml.safe_load(Path(config_path).read_text(encoding="utf-8"))
            if raw and "redis" in raw:
                config.update(raw["redis"])
            # 同时加载 TTL 和限流配置，供调用方使用
            if raw:
                config["ttl"] = raw.get("ttl", {})
                config["rate_limit"] = raw.get("rate_limit", {})
                config["rag_cache"] = raw.get("rag_cache", {})
        except Exception as e:
            logger.warning(f"[Redis] 配置文件加载失败: {e}，使用默认配置")

    # 环境变量优先级最高
    env_url = os.environ.get("REDIS_URL")
    if env_url:
        config["url"] = env_url

    return config


# ─── 全局缓存实例（懒加载单例） ────────────────────

_redis_cache_instance: Optional["RedisCache"] = None


def get_redis_cache(config_path: str | Path | None = None) -> "RedisCache":
    """
    获取全局唯一的 RedisCache 实例（单例工厂）。

    为什么用单例？
    一个进程只需要一个 Redis 连接池。多实例会导致连接数翻倍。

    延迟初始化：首次调用时才创建连接池，避免启动时未配置 Redis 导致报错。
    """
    global _redis_cache_instance
    if _redis_cache_instance is None:
        redis_config = get_redis_config(config_path)
        _redis_cache_instance = RedisCache(redis_config)
        logger.info(f"[Redis] 全局缓存实例已创建: {redis_config.get('url', 'unknown')}")
    return _redis_cache_instance


# ═══════════════════════════════════════════════════
# Cache Key 构造工具
# ═══════════════════════════════════════════════════

def _llm_cache_key(messages: list[dict]) -> str:
    """
    构造 LLM 缓存键。

    策略：取最近 3 条非 tool 消息做 MD5。
    （跳过 tool = 跳过工具执行结果，因为工具结果含动态数据
    如时间戳/订单号，会严重降低缓存命中率）

    为什么是 3 条？
    影响 LLM 回复的输入变量：system prompt + 最后一轮 user query + 上一轮 assistant reply。
    用完整历史做 key 会导致任何微小变化都 miss。

    为什么用 MD5？
    content 可能很长（含上下文），直接拼接做 key 浪费内存。
    MD5 碰撞概率可忽略（缓存 key 不需要密码学安全）。
    """
    # 从后往前扫描，只收集 system/user/assistant，跳过 tool
    relevant = []
    for m in reversed(messages):
        if m.get("role") in ("system", "user", "assistant"):
            relevant.append(m)
            if len(relevant) >= 3:
                break
    relevant.reverse()

    if not relevant:
        relevant = messages[-3:] if len(messages) >= 3 else messages

    raw = "|".join(
        f"{m.get('role', '')}:{m.get('content', '')}" for m in relevant
    )
    return hashlib.md5(raw.encode("utf-8")).hexdigest()


def _rag_cache_key(query: str, top_k: int) -> str:
    """
    构造 RAG 缓存键。

    用 query 全文 + top_k 做 MD5。
    注意：query 是经过 QueryRewriter 改写后的 query（改写后更精炼），
    所以"原始 query 不同 → 改写后可能相同 → 缓存命中"，这恰好是期望行为。
    """
    raw = f"{query}:k{top_k}"
    return hashlib.md5(raw.encode("utf-8")).hexdigest()


# ═══════════════════════════════════════════════════
# RedisCache 实现
# ═══════════════════════════════════════════════════

class RedisCache(CacheBackend):
    """
    Redis 缓存实现 — 同时支持同步和异步接口。

    设计：
    - async 方法 → 使用 redis.asyncio（用于 LLM 缓存、限流器、Session 缓存）
    - sync 方法  → 使用 redis.Redis（用于 RAG 缓存，现有 sync 调用方）

    两种 client 共享同一 URL 和连接池配置，
    但维护独立的连接池（async/sync 的 ConnectionPool 不兼容）。

    连接在第一次操作时懒初始化，不是 __init__ 时立即创建。
    这允许在未配置 Redis 时平稳降级。

    降级策略：
    - Redis 连接失败 → 不抛异常，记 warn 日志
    - 后续操作静默跳过（self._available = False）
    - 业务代码无需 try/catch Redis 异常
    """

    def __init__(self, config: dict | None = None):
        """
        Args:
            config: 配置字典（通常由 get_redis_config() 返回）
                关键字段:
                - url: 连接字符串
                - pool_max_size: 连接池大小（默认 20）
                - pool_timeout: 连接超时（默认 5.0）
                - socket_timeout: socket 超时（默认 3.0）
                - key_prefix: key 前缀（默认 "jdagent:"）
        """
        self.config = config or get_redis_config()
        self.key_prefix: str = self.config.get("key_prefix", "jdagent:")

        # TTL 配置（秒）
        ttl_cfg = self.config.get("ttl", {})
        self.ttl_llm: int = ttl_cfg.get("llm_response", 3600)
        self.ttl_rag: int = ttl_cfg.get("rag_result", 300)
        self.ttl_session: int = ttl_cfg.get("session_cache", 1800)

        # 限流配置
        rl_cfg = self.config.get("rate_limit", {})
        self.rate_limit_default: int = rl_cfg.get("default_limit", 30)
        self.rate_limit_window: int = rl_cfg.get("default_window", 60)
        self.rate_limit_llm: int = rl_cfg.get("llm_limit", 20)
        self.rate_limit_llm_window: int = rl_cfg.get("llm_window", 60)

        # 连接池（懒加载）
        self._async_pool = None
        self._redis_async = None   # import redis.asyncio 的引用，供作用域外使用
        self._sync_client = None
        # 是否可用（首次连接失败后设为 False，后续操作跳过）
        self._available = True
        self._reported_unavailable = False

    # ─── 内部：异步连接管理 ─────────────────────

    async def _get_async_conn(self):
        """
        获取异步 Redis 连接（懒初始化）。

        要点：
        - 连接池是全局共享的，不重复创建
        - 连接失败时 self._available = False，后续操作静默跳过
        - 不抛异常到上层（缓存失败不应拖垮业务）
        """
        if not self._available:
            return None

        if self._async_pool is None:
            try:
                import redis.asyncio as redis_async
            except ImportError:
                logger.warning("[Redis] redis 包未安装，缓存降级")
                self._available = False
                return None

            try:
                pool = redis_async.ConnectionPool.from_url(
                    self.config.get("url", "redis://localhost:6379/0"),
                    max_connections=self.config.get("pool_max_size", 20),
                    socket_connect_timeout=self.config.get("pool_timeout", 5.0),
                    socket_timeout=self.config.get("socket_timeout", 3.0),
                    retry_on_timeout=self.config.get("retry_on_timeout", True),
                    decode_responses=True,
                )
                # 真正尝试建立一条连接来验证 Redis 是否可达
                # 避免池建成功但连不上时，每条操作都报错
                try:
                    r = redis_async.Redis(connection_pool=pool)
                    await r.ping()
                except Exception:
                    # ping 不通 → Redis 没启动，不要浪费连接池
                    await pool.disconnect()
                    raise
                self._async_pool = pool
                self._redis_async = redis_async  # 存引用供后续使用
                logger.info("[Redis] 异步连接池已创建")
            except Exception as e:
                logger.warning(f"[Redis] 异步连接池创建失败: {e}（缓存降级）")
                self._available = False
                return None

        try:
            return self._redis_async.Redis(connection_pool=self._async_pool)
        except Exception as e:
            if not self._reported_unavailable:
                logger.warning(f"[Redis] 连接失败: {e}")
                self._reported_unavailable = True
                self._available = False
            return None

    # ─── 同步连接管理（用于 RAG 缓存） ─────────

    def _get_sync_client(self):
        """获取同步 Redis 客户端（懒初始化）。"""
        if not self._available:
            # 如果异步已标记不可用，同步大概率也不可用
            return None

        if self._sync_client is None:
            try:
                import redis as sync_redis
                self._sync_client = sync_redis.Redis.from_url(
                    self.config.get("url", "redis://localhost:6379/0"),
                    max_connections=self.config.get("pool_max_size", 20) // 2,
                    socket_timeout=self.config.get("socket_timeout", 3.0),
                    retry_on_timeout=self.config.get("retry_on_timeout", True),
                    decode_responses=True,
                )
                # 测试连接
                self._sync_client.ping()
                logger.info("[Redis] 同步客户端已创建")
            except Exception as e:
                logger.warning(f"[Redis] 同步客户端创建失败: {e}")
                self._available = False
                return None

        return self._sync_client

    # ─── 异步接口（CacheBackend 实现） ──────────

    def _full_key(self, key: str) -> str:
        """为 key 添加前缀，实现命名空间隔离。"""
        return f"{self.key_prefix}{key}"

    async def get(self, key: str) -> Optional[Any]:
        """获取缓存。返回 Python 对象（反序列化后），不存在返回 None。"""
        conn = await self._get_async_conn()
        if conn is None:
            return None
        try:
            raw = await conn.get(self._full_key(key))
            if raw is None:
                return None
            return _json_loads(raw)
        except Exception as e:
            logger.debug(f"[Redis] get 失败: {e}")
            return None

    async def set(self, key: str, value: Any, ttl: int = 0) -> None:
        """
        设置缓存。

        Args:
            key: 缓存键
            value: 任意 Python 对象（可 JSON 序列化即可）
            ttl: 过期秒数。0 表示使用默认 TTL（由 namespace 决定）
        """
        conn = await self._get_async_conn()
        if conn is None:
            return
        try:
            data = _json_dumps(value)
            if ttl > 0:
                await conn.setex(self._full_key(key), ttl, data)
            else:
                await conn.set(self._full_key(key), data)
        except Exception as e:
            logger.debug(f"[Redis] set 失败: {e}")

    async def delete(self, key: str) -> None:
        conn = await self._get_async_conn()
        if conn is None:
            return
        try:
            await conn.delete(self._full_key(key))
        except Exception as e:
            logger.debug(f"[Redis] delete 失败: {e}")

    async def exists(self, key: str) -> bool:
        conn = await self._get_async_conn()
        if conn is None:
            return False
        try:
            return bool(await conn.exists(self._full_key(key)))
        except Exception:
            return False

    async def clear_prefix(self, prefix: str) -> int:
        """
        按前缀批量删除。

        实现方式（面试常考）：
        - SCAN 迭代扫描（不是 KEYS——KEYS 会阻塞 Redis 主线程）
        - 每次 SCAN 100 条，避免单次返回过多
        - DEL 批量删除

        为什么不用 KEYS？KEYS 在 key 数量很多时会阻塞 Redis 几秒到几十秒。
        SCAN 是游标式迭代，每次返回少量数据，不阻塞。
        """
        conn = await self._get_async_conn()
        if conn is None:
            return 0
        try:
            pattern = f"{self._full_key(prefix)}*"
            cursor = 0
            deleted = 0
            while True:
                cursor, keys = await conn.scan(cursor=cursor, match=pattern, count=100)
                if keys:
                    deleted += await conn.delete(*keys)
                if cursor == 0:
                    break
            return deleted
        except Exception as e:
            logger.debug(f"[Redis] clear_prefix 失败: {e}")
            return 0

    # ─── 同步接口（供 RAG 等 sync 调用方） ──────

    def get_sync(self, key: str) -> Optional[Any]:
        """同步版 get，用于非 async 上下文。"""
        client = self._get_sync_client()
        if client is None:
            return None
        try:
            raw = client.get(self._full_key(key))
            if raw is None:
                return None
            return _json_loads(raw)
        except Exception as e:
            logger.debug(f"[Redis] get_sync 失败: {e}")
            return None

    def set_sync(self, key: str, value: Any, ttl: int = 0) -> None:
        """同步版 set。"""
        client = self._get_sync_client()
        if client is None:
            return
        try:
            data = _json_dumps(value)
            if ttl > 0:
                client.setex(self._full_key(key), ttl, data)
            else:
                client.set(self._full_key(key), data)
        except Exception as e:
            logger.debug(f"[Redis] set_sync 失败: {e}")

    def delete_sync(self, key: str) -> None:
        """同步版 delete。"""
        client = self._get_sync_client()
        if client is None:
            return
        try:
            client.delete(self._full_key(key))
        except Exception as e:
            logger.debug(f"[Redis] delete_sync 失败: {e}")

    def clear_prefix_sync(self, prefix: str) -> int:
        """
        同步版按前缀批量删除（SCAN + DEL）。
        用于 RAG 知识库更新时清除所有缓存。
        """
        client = self._get_sync_client()
        if client is None:
            return 0
        try:
            pattern = f"{self._full_key(prefix)}*"
            cursor = 0
            deleted = 0
            while True:
                cursor, keys = client.scan(cursor=cursor, match=pattern, count=100)
                if keys:
                    deleted += client.delete(*keys)
                if cursor == 0:
                    break
            return deleted
        except Exception as e:
            logger.debug(f"[Redis] clear_prefix_sync 失败: {e}")
            return 0

    # ─── LLM 缓存专用方法 ──────────────────────

    def _build_llm_cache_key(self, session_key: str, messages: list[dict]) -> str:
        """
        构造 LLM 缓存的完整 Redis key（含 session 前缀）。

        格式: llm:{session_key}:{md5}
        session_key 前缀确保不同用户的 LLM 缓存隔离。
        invalidate_llm_cache(session_key) 用 clear_prefix("llm:{session_key}:")
        只清除当前用户的缓存，不影响其他用户。
        """
        return f"llm:{session_key}:{_llm_cache_key(messages)}"

    async def get_llm_cache(self, session_key: str, messages: list[dict]) -> Optional[dict]:
        """
        查询 LLM 缓存。

        参数:
            session_key: 会话标识（用于 key 隔离）
            messages: 消息列表（用于生成缓存 key）

        缓存命中时返回 {content, tool_calls, finish_reason}，
        miss 返回 None。

        面试知识点：缓存穿透防护
        - 如果 LLM 返回空结果（content is None），也缓存一个空占位
        - 下次即使"没有结果"也不会重新调用 API
        """
        key = self._build_llm_cache_key(session_key, messages)
        return await self.get(key)

    async def set_llm_cache(self, session_key: str, messages: list[dict], result: dict) -> None:
        """
        写入 LLM 缓存。
        即使 result 为空（content=None, tool_calls=[]），也缓存空结构。
        这叫"空结果缓存"——防止缓存穿透。
        """
        key = self._build_llm_cache_key(session_key, messages)
        await self.set(key, result, ttl=self.ttl_llm)

    async def invalidate_llm_cache(self, session_key: str) -> int:
        """
        使某个 session 的 LLM 缓存失效（仅当前 session，不影响其他用户）。

        在 session 收到新消息后调用，因为新消息意味着对话状态已变，
        之前的 LLM 缓存结果不再适用。

        通过 clear_prefix 模糊删除该 session 的所有 llm 缓存。
        key 格式为 llm:{session_key}:{md5}，所以用 llm:{session_key}: 做前缀匹配。
        """
        return await self.clear_prefix(f"llm:{session_key}:")

    # ─── RAG 缓存专用方法（同步版） ────────────

    def _build_rag_cache_key(self, query: str, top_k: int) -> str:
        return f"rag:{_rag_cache_key(query, top_k)}"

    def get_rag_cache(self, query: str, top_k: int = 5) -> Optional[str]:
        """
        查询 RAG 缓存（同步）。

        返回格式化后的文档文本（string），
        miss 返回 None。

        面试知识点：为什么 RAG 缓存 TTL 短（5 分钟）？
        - RAG 知识库可能随时更新（如新增文档）
        - TTL 短 = 数据新鲜度好
        - 5 分钟足够覆盖"同个问题短时间内重复问"的场景
        """
        key = self._build_rag_cache_key(query, top_k)
        return self.get_sync(key)

    def set_rag_cache(self, query: str, top_k: int, result: str) -> None:
        """写入 RAG 缓存。"""
        key = self._build_rag_cache_key(query, top_k)
        self.set_sync(key, result, ttl=self.ttl_rag)

    # ─── Session 缓存专用方法 ──────────────────

    def _build_session_cache_key(self, session_key: str) -> str:
        return f"session:{session_key}"

    async def get_session_cache(self, session_key: str) -> Optional[dict]:
        """
        获取 Session 缓存（L2 缓存）。

        缓存内容：
        - messages: list[dict]（最近 50 条）
        - metadata: dict
        - last_consolidated: int
        - cached_at: str（缓存时间戳，用于调试）

        为什么只存最近 50 条？
        - 完整 session 可能很长（几百条消息）
        - 但 LLM 只需要最近 N 条（get_history 默认 max_messages=500）
        - 50 条足够满足绝大多数请求
        - 超过 50 条的从 PG 完整加载后补齐
        """
        key = self._build_session_cache_key(session_key)
        return await self.get(key)

    async def set_session_cache(self, session_key: str, session_data: dict) -> None:
        """
        写入 Session 缓存。

        session_data 是序列化后的 session 字典，结构:
        {
            "key": str,
            "messages": list[dict],  # 最近 50 条
            "metadata": dict,
            "last_consolidated": int,
            "updated_at": str,
        }
        """
        key = self._build_session_cache_key(session_key)
        await self.set(key, session_data, ttl=self.ttl_session)

    async def delete_session_cache(self, session_key: str) -> None:
        """删除 Session 缓存。"""
        key = self._build_session_cache_key(session_key)
        await self.delete(key)

    # ─── 资源清理 ──────────────────────────────

    async def close(self) -> None:
        """关闭所有连接。在进程退出时调用。"""
        if self._async_pool:
            await self._async_pool.disconnect()
            self._async_pool = None
        if self._sync_client:
            self._sync_client.close()
            self._sync_client = None
        logger.info("[Redis] 连接已关闭")

    # ─── 分布式锁（缓存击穿防护） ──────────────

    async def lock(self, key: str, ttl: int = 5) -> bool:
        """
        尝试获取分布式锁（SET NX EX）。

        只会在 key 不存在时设置成功——即"拿到锁"。
        ttl 是锁的自动过期时间，防止持有者崩溃后死锁。

        Args:
            key: 锁标识（会被自动加前缀）
            ttl: 锁持有时间（秒），默认 5 秒

        Returns:
            True = 拿到锁，False = 别人拿着
        """
        conn = await self._get_async_conn()
        if conn is None:
            return True  # 单机模式，不需要分布式锁
        try:
            result = await conn.eval(
                _LOCK_SCRIPT,
                1,
                self._full_key(f"lock:{key}"),
                self._lock_value,
                ttl,
            )
            return bool(result)
        except Exception as e:
            logger.debug(f"[Redis] lock 失败，放行: {e}")
            return True

    async def unlock(self, key: str) -> None:
        """释放分布式锁（只删自己的锁）。"""
        conn = await self._get_async_conn()
        if conn is None:
            return
        try:
            await conn.eval(
                _UNLOCK_SCRIPT,
                1,
                self._full_key(f"lock:{key}"),
                self._lock_value,
            )
        except Exception as e:
            logger.debug(f"[Redis] unlock 失败: {e}")

    @property
    def _lock_value(self) -> str:
        """
        锁持有者标识（hostname:pid）。

        为什么需要这个？
        假设进程 A 拿到锁，执行超时（超过 TTL），
        锁自动释放。进程 B 拿到同一把锁。
        这时 A 执行完了，如果直接 DEL 会删掉 B 的锁。

        用 value 匹配确保"谁的孩子谁抱走"。
        """
        if not hasattr(self, "_lock_value_cache"):
            import platform
            self._lock_value_cache = f"{platform.node()}:{os.getpid()}"
        return self._lock_value_cache


# ═══════════════════════════════════════════════════
# RedisRateLimiter — 滑动窗口限流器
# ═══════════════════════════════════════════════════
#
# 算法：滑动窗口（Sliding Window Log）
# ────────────────────────────────
# 用 Redis Sorted Set 实现，member=雪花ID，score=时间戳（毫秒）。
#
# 每次请求：
# 1. ZREMRANGEBYSCORE 0 (now - window) — 清理窗口外的过期数据
# 2. ZCARD — 统计窗口内请求数
# 3. 未超限 → ZADD 当前时间戳 → EXPIRE window（让 key 自动过期）
# 4. 超限 → 拒绝请求
#
# 为什么用 Sorted Set 而不是 INCR？
# - INCR + EXPIRE 是"固定窗口"，边界处有突刺问题：
#   窗口前半段没流量，后半段涌进来，下一秒前半段又没流量了
# - Sorted Set 是"滑动窗口"，每个请求都有自己的时间戳，边界平滑
# - 代价是内存略多（每个请求存一个 member + score），但限流 key 数量有限，可忽略
#
# 面试常考点：固定窗口 vs 滑动窗口
# - 固定窗口（INCR + EXPIRE）：实现简单，但窗口边界处允许 2 倍流量
# - 滑动窗口（Sorted Set）：更精确，但内存消耗稍高
# ═══════════════════════════════════════════════════

class RedisRateLimiter:
    """
    滑动窗口速率限制器。

    用法：
        limiter = RedisRateLimiter(redis_cache)
        allowed, remaining = await limiter.check("session:abc", limit=30, window=60)
        if not allowed:
            return 429  # Too Many Requests

    设计：
    - 每个 key 独立计数（session 级别隔离）
    - 自动清理过期数据（每次 check 时 ZREMRANGEBYSCORE）
    - 窗口大小可配置（不同场景不同阈值）
    - Redis 不可用时放行（fail-open：缓存故障不影响业务）
    """

    def __init__(self, cache: RedisCache):
        self.cache = cache
        # 生成唯一 member 的 ID 生成器（线程安全，详见 _member_id）
        self._id_counter = 0

    def _member_id(self) -> str:
        """
        生成唯一的 member ID。

        用时间戳 + 计数器，确保同一毫秒内的不同请求 ID 也不同。
        为什么不用 UUID？UUID 太长（36 字符），Sorted Set 里浪费内存。
        """
        self._id_counter += 1
        return f"{time.monotonic_ns()}:{self._id_counter}"

    async def check(
        self,
        key: str,
        limit: int | None = None,
        window: int | None = None,
    ) -> tuple[bool, int]:
        """
        检查是否允许当前请求通过。

        Args:
            key: 限流 key（通常是 session_key）
            limit: 窗口内允许的最大请求数（默认 config.rate_limit.default_limit）
            window: 滑动窗口大小（秒）（默认 config.rate_limit.default_window）

        Returns:
            (allowed, remaining)
            - allowed: True=放行, False=限流
            - remaining: 当前窗口内剩余可用次数
        """
        conn = await self.cache._get_async_conn()
        if conn is None:
            # Redis 不可用时放行（fail-open）
            return True, -1

        limit = limit or self.cache.rate_limit_default
        window = window or self.cache.rate_limit_window

        try:
            full_key = self.cache._full_key(f"ratelimit:{key}")
            now_ms = int(time.time() * 1000)            # 当前时间（毫秒）
            window_ms = window * 1000                    # 窗口大小（毫秒）
            cutoff = now_ms - window_ms                  # 窗口左边界

            # ── Step 1: 清理窗口外的过期数据 ──
            # ZREMRANGEBYSCORE: 删除 score 在 [0, cutoff] 范围内的所有 member
            await conn.zremrangebyscore(full_key, 0, cutoff)

            # ── Step 2: 统计窗口内请求数 ──
            # ZCARD: 返回 sorted set 的 cardinality（元素数量）
            current = await conn.zcard(full_key)

            # ── Step 3: 判断是否超限 ──
            if current >= limit:
                remaining = max(0, limit - current)
                return False, remaining

            # ── Step 4: 记录本次请求 ──
            # ZADD: 添加 member，score=当前时间戳
            await conn.zadd(full_key, {self._member_id(): now_ms})
            # EXPIRE: 设置 key 的过期时间（window 秒后自动删除）
            # 这样即使忘了清理，Redis 也会自动回收
            await conn.expire(full_key, window)

            remaining = max(0, limit - current - 1)
            return True, remaining

        except Exception as e:
            logger.warning(f"[RateLimiter] 检查失败，放行: {e}")
            return True, -1  # fail-open：异常时放行

    async def check_llm(self, session_key: str) -> tuple[bool, int]:
        """
        专用方法：检查 LLM 调用是否超限。

        使用独立的 key 前缀和限制参数，
        与普通消息限流互不干扰。
        """
        return await self.check(
            key=f"llm:{session_key}",
            limit=self.cache.rate_limit_llm,
            window=self.cache.rate_limit_llm_window,
        )

    async def check_atomic(
        self,
        key: str,
        limit: int | None = None,
        window: int | None = None,
    ) -> tuple[bool, int]:
        """
        原子版限流检查（Lua 脚本实现，适合高并发场景）。

        与 check() 的区别：
        - check()：3-4 条独立 Redis 命令，非原子
        - check_atomic()：整个逻辑包在 Lua 脚本中，Redis 单线程原子执行

        适用场景：
        - check()：有 per-session lock 保护的场景（loop.py 入口的消息级限流）
        - check_atomic()：无锁保护的全局 key（如 global:llm，所有 session 共用）

        参数和返回值同 check()。
        """
        conn = await self.cache._get_async_conn()
        if conn is None:
            return True, -1

        limit = limit or self.cache.rate_limit_default
        window = window or self.cache.rate_limit_window

        try:
            full_key = self.cache._full_key(f"ratelimit:{key}")
            now_ms = int(time.time() * 1000)
            result = await conn.eval(
                _RATE_LIMIT_SCRIPT,
                1,
                full_key,
                now_ms,
                window * 1000,
                limit,
                self._member_id(),
            )
            # Lua 返回 [1, count]（放行）或 [0, count]（拒绝）
            allowed = bool(result[0])
            remaining = max(0, limit - result[1])
            return allowed, remaining
        except Exception as e:
            logger.warning(f"[RateLimiter] check_atomic 失败，放行: {e}")
            return True, -1

    async def get_remaining(self, key: str) -> int:
        """查询当前窗口剩余次数（不消耗次数）。"""
        conn = await self.cache._get_async_conn()
        if conn is None:
            return -1
        try:
            full_key = self.cache._full_key(f"ratelimit:{key}")
            count = await conn.zcard(full_key)
            return max(0, self.cache.rate_limit_default - count)
        except Exception:
            return -1
