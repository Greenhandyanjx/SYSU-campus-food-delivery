"""
任务定义 — RabbitMQ 后台任务类型枚举 + Handler 路由表

设计说明：
- TaskType 枚举值对应 RabbitMQ 的 routing_key，用于消息路由
- HANDLER_MAP 将 TaskType → 处理函数，由 Consumer 在消费时调度
- 每个 handler 接收 (data: dict, loop: AgentLoop) 参数

两类后台任务：
1. Dream 归档      → 从 history.jsonl 提取原子事实，增量更新 MEMORY.md
2. 记忆 Consolidation → 压缩旧会话摘要到 history.jsonl

异常处理策略：
- 临时失败（网络超时等）：nack + requeue → 最多重试 max_retries 次
- 永久失败（数据错误）：ack + 记录 error 日志（不阻塞队列）
"""
from __future__ import annotations

from enum import Enum
from typing import TYPE_CHECKING, Any

from loguru import logger

if TYPE_CHECKING:
    from agent.loop import AgentLoop


class TaskType(str, Enum):
    """后台任务类型枚举。

    值 = RabbitMQ routing_key，用于消息路由。
    命名采用短小精悍的风格（对标 Celery Task 命名）。
    """
    DREAM = "dream"
    MEMORY_CONSOLIDATION = "consolidation"


# ─── Handlers ──────────────────────────────────────────


async def handle_dream(data: dict, loop: "AgentLoop") -> None:
    """执行 Dream 归档：从 history.jsonl 提取事实写入 MEMORY.md

    为什么放后台：
    Dream 需要调用 LLM 做事实提取+去重，耗时约 3-10 秒，
    用户不需要等待这个结果。
    """
    result = loop.chat_memory.run_dream()
    added = result.get("added", 0)
    replaced = result.get("replaced", 0)
    if added > 0 or replaced > 0:
        logger.info(f"[MQ Dream] 归档完成: +{added} 条新增, 替换 {replaced} 条")
    else:
        logger.debug(f"[MQ Dream] 无新内容: {result}")


async def handle_consolidation(data: dict, loop: "AgentLoop") -> None:
    """执行记忆 Consolidation：压缩旧会话摘要到 history.jsonl

    流程：
    1. 从存储重新加载 session（避免共享内存状态）
    2. 执行 maybe_consolidate（内部判断阈值）
    3. 如果做了 consolidation，截断已归档的消息
    4. 保存 session + 清除内存缓存（下次请求重新加载）

    注意：
    consolidation 需要在同一进程的共享 Session 对象中小心处理。
    这里从存储层重新加载，避免与主进程的 in-memory session 冲突。
    """
    session_key = data.get("session_key")
    if not session_key:
        logger.warning("[MQ Consolidation] 缺少 session_key，跳过")
        return

    # 从存储重新加载 session（不与主进程共享对象引用）
    # 这样可以安全地修改 last_consolidated 和截断 messages
    session = await loop.sessions.aget_or_create(session_key)

    await loop.memory_consolidator.maybe_consolidate(session)

    # 如果做了 consolidation，截断已归档的消息
    if session.last_consolidated > 0:
        session.messages[:session.last_consolidated] = []
        session.last_consolidated = 0

    # 保存回存储（会覆盖 JSONL/PG）
    await loop.sessions.asave(session)

    # 清除内存缓存，确保下一次主进程请求加载最新数据
    # 注意：当前正在处理请求的主进程持有的是旧 Session 对象，
    # 清除缓存后，下一次 get_or_create 会从存储重新加载。
    loop.sessions._cache.pop(session_key, None)

    logger.debug(f"[MQ Consolidation] session={session_key} 处理完成")


# ─── 路由表 ────────────────────────────────────────────

HANDLER_MAP: dict[TaskType, Any] = {
    TaskType.DREAM: handle_dream,
    TaskType.MEMORY_CONSOLIDATION: handle_consolidation,
}

# routing_key → TaskType 映射（避免每次查枚举）
ROUTING_KEY_MAP: dict[str, TaskType] = {
    "dream": TaskType.DREAM,
    "consolidation": TaskType.MEMORY_CONSOLIDATION,
}
