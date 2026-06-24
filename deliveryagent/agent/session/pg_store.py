"""
PostgreSQL 存储后端：会话和消息的持久化。

设计目标：
1. 替代 JSONL 作为主存储，提供 SQL 查询能力
2. 与 SessionManager 解耦，通过统一的 load/save 接口交互
3. 幂等操作：多次 save 不会产生重复数据
4. 自动建表：首次使用时检测表是否存在，不存在则自动创建

使用方式（由 SessionManager 内部调用，不直接对外）：
    backend = PgBackend(config)
    await backend.initialize()
    session = await backend.load("cli:direct")
    await backend.save(session)
    await backend.close()
"""

import json
import os
from datetime import datetime
from pathlib import Path
from typing import Any

from loguru import logger


class PgBackend:
    """
    PostgreSQL 存储后端。

    职责范围：
    - 管理 asyncpg 连接池（懒加载，首次使用时才建立连接）
    - 执行建表 DDL（幂等，仅首次连接时执行）
    - Session ↔ 关系表的序列化/反序列化
    - 事务内原子写入（sessions + messages 在同一个 PG 事务中完成）
    - 故障隔离：连接失败时抛出异常，由调用方（SessionManager）处理降级

    不是啥：
    × 不负责内存缓存（那是 SessionManager 的职责）
    × 不负责 JSONL 回退（那是 SessionManager 的职责）
    × 不负责业务逻辑（只是存储层）
    """

    def __init__(self, config: dict):
        """
        初始化 PG 后端。

        参数:
            config: 数据库配置字典，包含:
                - dsn: 连接字符串（必需）
                - pool_min_size: 连接池最小连接数（默认 1）
                - pool_max_size: 连接池最大连接数（默认 10）
                - connect_timeout: 连接超时秒数（默认 10）
        """
        self.dsn: str = config.get("dsn") or os.environ.get("DATABASE_URL", "")
        if not self.dsn:
            raise ValueError(
                "PgBackend: 缺少 dsn 连接串。"
                "请在 config/database.yml 中配置，或设置 DATABASE_URL 环境变量。"
            )
        self.pool_min_size: int = config.get("pool_min_size", 1)
        self.pool_max_size: int = config.get("pool_max_size", 10)
        self.connect_timeout: float = config.get("connect_timeout", 10.0)

        # 连接池对象（由 initialize() 创建）
        self._pool = None

        logger.debug(f"[PgBackend] 初始化完成，dsn={self._mask_dsn(self.dsn)}")

    # ─── 连接池管理 ────────────────────────────────────

    async def initialize(self) -> None:
        """
        初始化连接池 + 自动建表。

        幂等：可以多次调用，第二次开始是空操作（检测到 _pool 已存在则跳过）。
        线程安全：asyncio 单线程模型 + asyncpg 内部锁保护。
        """
        if self._pool is not None:
            return  # 已初始化，跳过

        import asyncpg

        try:
            self._pool = await asyncpg.create_pool(
                dsn=self.dsn,
                min_size=self.pool_min_size,
                max_size=self.pool_max_size,
                timeout=self.connect_timeout,
            )
            logger.info(f"[PgBackend] 连接池已创建（min={self.pool_min_size}, max={self.pool_max_size}）")
        except Exception as e:
            logger.error(f"[PgBackend] 连接池创建失败: {e}")
            raise

        # 连接成功后自动建表
        await self._ensure_tables()

    async def close(self) -> None:
        """关闭连接池，释放所有 PG 连接。"""
        if self._pool:
            await self._pool.close()
            self._pool = None
            logger.info("[PgBackend] 连接池已关闭")

    async def _get_connection(self):
        """
        从连接池获取一个连接。

        使用 _get_connection() 而不是直接访问 self._pool，
        这样可以在获取连接时做额外的检查或等待。
        """
        if self._pool is None:
            raise RuntimeError("PgBackend 未初始化，请先调用 initialize()")
        return await self._pool.acquire()

    # ─── 建表 ────────────────────────────────────────

    async def _ensure_tables(self) -> None:
        """
        确保数据库表已创建（幂等）。

        执行 db/migrations/001_init.sql 中的 DDL。
        如果文件不存在或读取出错，退而使用内嵌的 SQL（保证可独立运行）。
        """
        # 优先从文件读取 SQL（方便开发者修改 DDL）
        sql = self._read_init_sql()

        # 如果文件不存在，使用内嵌 DDL（保证无外部依赖也可运行）
        if not sql:
            sql = self._INIT_SQL
            logger.debug("[PgBackend] 使用内嵌 DDL 建表")

        async with self._pool.acquire() as conn:
            try:
                await conn.execute(sql)
                logger.info("[PgBackend] 数据库表已就绪")
            except Exception as e:
                logger.error(f"[PgBackend] 建表失败: {e}")
                raise

    @staticmethod
    def _read_init_sql() -> str | None:
        """
        从 db/migrations/001_init.sql 文件读取建表 SQL。

        搜索路径：从当前工作目录开始，向上查找 migrations 目录。
        这是为了支持从不同位置（项目根、api/ 目录）启动的情况。
        """
        # 搜索路径列表：从最可能的开始
        search_paths = [
            Path.cwd() / "db" / "migrations" / "001_init.sql",
            Path(__file__).parent.parent.parent.parent / "db" / "migrations" / "001_init.sql",
            Path(__file__).parent.parent.parent / "db" / "migrations" / "001_init.sql",
        ]
        for sp in search_paths:
            if sp.exists():
                try:
                    return sp.read_text(encoding="utf-8")
                except Exception:
                    continue
        return None

    # ─── DDL: 内嵌建表 SQL ─────────────────────────────

    _INIT_SQL = """
    CREATE TABLE IF NOT EXISTS sessions (
        key                TEXT PRIMARY KEY,
        created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        metadata           JSONB DEFAULT '{}',
        last_consolidated  INTEGER DEFAULT 0
    );

    CREATE TABLE IF NOT EXISTS messages (
        id               BIGSERIAL PRIMARY KEY,
        session_key      TEXT NOT NULL REFERENCES sessions(key) ON DELETE CASCADE,
        role             TEXT NOT NULL,
        content          TEXT NOT NULL DEFAULT '',
        timestamp        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        tool_calls       JSONB,
        tool_call_id     TEXT,
        name             TEXT
    );

    CREATE INDEX IF NOT EXISTS idx_messages_session_ts
        ON messages(session_key, timestamp, id);

    CREATE INDEX IF NOT EXISTS idx_messages_role
        ON messages(role);

    CREATE OR REPLACE FUNCTION touch_session_updated_at()
    RETURNS TRIGGER AS $$
    BEGIN
        UPDATE sessions SET updated_at = NOW()
        WHERE key = COALESCE(NEW.session_key, OLD.session_key);
        RETURN NEW;
    END;
    $$ LANGUAGE plpgsql;

    DROP TRIGGER IF EXISTS trg_messages_touch_session ON messages;

    CREATE TRIGGER trg_messages_touch_session
        AFTER INSERT OR UPDATE OR DELETE ON messages
        FOR EACH ROW
        EXECUTE FUNCTION touch_session_updated_at();
    """

    # ─── 核心 CRUD ─────────────────────────────────────

    async def load(self, key: str) -> dict | None:
        """
        从 PostgreSQL 加载一个完整会话。

        流程：
        1. 查 sessions 表获取会话元数据
        2. 查 messages 表获取该会话所有消息（按 id 排序）
        3. 组装为 {session, messages} 字典返回

        参数:
            key: 会话标识，如 "cli:direct"

        返回:
            会话数据的字典（包含 session 元数据和 messages 列表），
            或 None（会话不存在）
        """
        conn = await self._get_connection()
        try:
            # 第一步：查 sessions 表
            row = await conn.fetchrow(
                "SELECT * FROM sessions WHERE key = $1", key
            )
            if row is None:
                return None

            # 第二步：查 messages 表（按 id 排序保证顺序）
            msg_rows = await conn.fetch(
                "SELECT * FROM messages WHERE session_key = $1 ORDER BY id",
                key,
            )

            logger.debug(f"[PgBackend] 加载会话 {key}: {len(msg_rows)} 条消息")

            # 第三步：组装 Session 的字段（由 SessionManager 创建 Session 对象）
            return {
                "key": row["key"],
                "messages": [self._row_to_message(r) for r in msg_rows],
                "created_at": row["created_at"],
                "updated_at": row["updated_at"],
                "metadata": row["metadata"] or {},
                "last_consolidated": row["last_consolidated"] or 0,
            }
        except Exception as e:
            logger.warning(f"[PgBackend] 加载会话 {key} 失败: {e}")
            return None
        finally:
            # 释放连接回连接池
            await self._pool.release(conn)

    async def save(self, key: str, messages: list[dict],
                   created_at: datetime, updated_at: datetime,
                   metadata: dict, last_consolidated: int) -> bool:
        """
        将会话保存到 PostgreSQL。

        写入策略：
        - sessions 表：UPSERT（INSERT ON CONFLICT DO UPDATE）
        - messages 表：基于 PG 中的实际记录数做增量写入
        - 异常安全：两条 SQL 在同一个 PG 事务中，要么全成功要么全回滚

        增量写入逻辑：
        1. 查询 PG 中该会话的现有消息数（SELECT COUNT）
        2. 与传入的消息列表长度比较：
           a. total >= count → 只写新增部分（messages[count:]）
           b. total < count → 内存发生过 truncate，删除 PG 旧消息后重写
        3. 这样即使内存中的 session.messages 被截断过，也不会丢数据

        参数:
            key: 会话标识
            messages: 完整消息列表（包含历史和新消息）
            created_at: 创建时间
            updated_at: 最后更新时间
            metadata: 会话元数据
            last_consolidated: consolidation 偏移

        返回:
            True=保存成功, False=保存失败
        """  # noqa: E501
        import asyncio

        max_retries = 3
        for attempt in range(max_retries):
            conn = await self._get_connection()
            try:
                async with conn.transaction():
                    # ── Step 1: 写 sessions 表（UPSERT）──
                    await conn.execute("""
                        INSERT INTO sessions (key, created_at, updated_at, metadata, last_consolidated)
                        VALUES ($1, $2, $3, $4::jsonb, $5)
                        ON CONFLICT (key) DO UPDATE SET
                            updated_at = EXCLUDED.updated_at,
                            metadata = EXCLUDED.metadata,
                            last_consolidated = EXCLUDED.last_consolidated
                    """, key, created_at, updated_at,
                        json.dumps(metadata, ensure_ascii=False),
                        last_consolidated,
                    )

                    # ── Step 2: 查出 PG 中现有的消息数 ──
                    count = await conn.fetchval(
                        "SELECT COUNT(*) FROM messages WHERE session_key = $1", key
                    )
                    count = count or 0
                    total = len(messages)

                    # ── Step 3: 确定需要写入的消息 ──
                    if total >= count:
                        # 正常情况：有新增消息，只追加差异部分
                        new_msgs = messages[count:]
                    else:
                        # 异常情况：内存中的 messages 被 truncate 过
                        # （例如 consolidation 后清除了前半部分）
                        # 此时 PG 中的记录数多于内存中的记录数。
                        #
                        # 策略分支：
                        # - total > 0：部分截断，删除 PG 旧消息后重新写入剩余消息
                        # - total == 0：纯 truncation（session.messages 被清空但 PG 数据
                        #   依然有效），跳过删除，保留 PG 中已有的消息
                        if total > 0:
                            await conn.execute(
                                "DELETE FROM messages WHERE session_key = $1", key
                            )
                            new_msgs = messages
                            count = 0
                        else:
                            # total == 0：纯 truncation，PG 数据是完整的，跳过写操作
                            new_msgs = []

                    # ── Step 4: 批量写入消息 ──
                    if new_msgs:
                        for msg in new_msgs:
                            await conn.execute("""
                                INSERT INTO messages
                                    (session_key, role, content, timestamp,
                                     tool_calls, tool_call_id, name)
                                VALUES ($1, $2, $3, $4, $5::jsonb, $6, $7)
                            """,
                                key,
                                msg.get("role", ""),
                                msg.get("content") or "",
                                self._parse_timestamp(msg.get("timestamp")),
                                json.dumps(msg.get("tool_calls"), ensure_ascii=False)
                                if msg.get("tool_calls") else None,
                                msg.get("tool_call_id"),
                                msg.get("name"),
                            )
                        logger.debug(
                            f"[PgBackend] 保存会话 {key}: "
                            f"{'重写' if total < count else '新增'} "
                            f"{len(new_msgs)} 条消息"
                        )

                    return True

            except Exception as e:
                if attempt < max_retries - 1:
                    wait = 0.1 * (2 ** attempt)
                    logger.debug(f"[PgBackend] 保存会话 {key} 失败，{wait:.1f}s 后重试 ({attempt + 1}/3): {e}")
                    await asyncio.sleep(wait)
                else:
                    logger.warning(f"[PgBackend] 保存会话 {key} 失败（已重试 3 次）: {e}")
                    return False
            finally:
                await self._pool.release(conn)

        return False

    async def delete(self, key: str) -> bool:
        """
        删除会话及其所有消息（CASCADE 自动处理 messages）。

        参数:
            key: 会话标识

        返回:
            True=删除成功, False=会话不存在或删除失败
        """
        conn = await self._get_connection()
        try:
            result = await conn.execute("DELETE FROM sessions WHERE key = $1", key)
            deleted = result != "DELETE 0"
            if deleted:
                    logger.debug(f"[PgBackend] 删除会话 {key}")
            return deleted
        except Exception as e:
            logger.warning(f"[PgBackend] 删除会话 {key} 失败: {e}")
            return False
        finally:
            await self._pool.release(conn)

    async def list_keys(self) -> list[dict]:
        """
        列出所有会话的概要信息。

        用于 health 端点或会话管理 UI，不做全量消息加载。

        返回:
            [{"key": "...", "created_at": "...", "updated_at": "..."}, ...]
            按 updated_at 降序排列（最近更新的在前）
        """
        conn = await self._get_connection()
        try:
            rows = await conn.fetch(
                    "SELECT key, created_at, updated_at FROM sessions ORDER BY updated_at DESC"
            )
            return [
                    {
                    "key": r["key"],
                    "created_at": r["created_at"].isoformat(),
                    "updated_at": r["updated_at"].isoformat(),
                    }
                    for r in rows
            ]
        except Exception as e:
            logger.warning(f"[PgBackend] 列出会话失败: {e}")
            return []
        finally:
            await self._pool.release(conn)

    # ─── 工具方法 ─────────────────────────────────────

    @staticmethod
    def _row_to_message(row) -> dict:
        """
        将 messages 表的一行转为内存中的消息字典。

        关键转换：
        - timestamp: TIMESTAMPTZ → ISO 格式字符串（与 JSONL 格式一致）
        - tool_calls: JSONB → Python list/dict（JSONL 中也是 dict）
        - content: None → ""（避免 LLM 将 null 渲染为 "None"）
        """
        msg: dict[str, Any] = {
            "role": row["role"],
            "content": row["content"] or "",
        }

        # timestamp 可能有不同的类型（取决于驱动版本），统一转字符串
        ts = row["timestamp"]
        if ts is not None:
            if hasattr(ts, "isoformat"):
                    msg["timestamp"] = ts.isoformat()
            else:
                    msg["timestamp"] = str(ts)

        # 以下字段只在非空时包含，避免无关字段污染消息字典
        tool_calls = row["tool_calls"]
        if tool_calls:
            # tool_calls 从 PG 读出来是 Python list/dict（asyncpg 自动解析 JSONB）
            msg["tool_calls"] = tool_calls

        if row["tool_call_id"]:
            msg["tool_call_id"] = row["tool_call_id"]

        if row["name"]:
            msg["name"] = row["name"]

        return msg

    @staticmethod
    def _parse_timestamp(ts: Any) -> datetime | None:
        """
        将各种格式的时间戳解析为 datetime 对象。

        消息中的 timestamp 可能是：
        - ISO 格式字符串: "2024-01-15T10:30:00.123456"
        - float 时间戳: 1705293000.123456
        - 已经是 datetime 对象
        - None
        """
        if ts is None:
            return None
        if isinstance(ts, datetime):
            return ts
        if isinstance(ts, (int, float)):
            return datetime.fromtimestamp(ts)
        if isinstance(ts, str):
            try:
                    return datetime.fromisoformat(ts)
            except ValueError:
                    return None
        return None

    @staticmethod
    def _mask_dsn(dsn: str) -> str:
        """
        脱敏 DSN，用于日志输出。

        将密码替换为 ****，防止敏感信息泄露到日志。
        示例: postgresql://user:****@host:5432/db
        """
        if "@" not in dsn:
            return dsn.split("://")[0] + "://****"
        user_part = dsn.split("@")[0]
        host_part = dsn.split("@")[1]
        if ":" in user_part and user_part.count(":") >= 2:
            safe_user = user_part.rsplit(":", 1)[0] + ":****"
        else:
            safe_user = user_part
        return f"{safe_user}@{host_part}"
