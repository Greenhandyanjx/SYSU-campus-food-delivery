"""
MQLLMProvider — 通过 RabbitMQ 对 LLM API 调用进行速率控制

设计模式：代理模式（Proxy Pattern）
透明替换原始 LLMProvider，AgentLoop 无感知。AgentLoop 照常调
self.provider.chat_with_retry()，只是底层改走 MQ 而已。

为什么需要这个？
没有 MQ 时：用户请求 → ReAct 循环 → 调 LLM API（可能多个请求
同时调，API 费用失控）
有 MQ 后：  用户请求 → ReAct 循环 → publish 到 llm_request 队列
            → Worker 以 prefetch=1 逐个消费（天然全局限流）

调用流程：
AgentLoop._run_agent_loop()
  → self.provider.chat_with_retry()     ← MQLLMProvider 透明拦截
    ├─ MQ 可用  → 发消息 + 等回复（限流）
    └─ MQ 不可用 → 降级直连

RPC over MQ 模式（面试必问）：
┌─────────────────────────────────────────────────────────────────┐
│ Producer（MQLLMProvider）                                        │
│   1. 创建临时回复队列（exclusive + auto_delete = 用完即焚）          │
│   2. 发消息到 llm_request 队列（带上 reply_to + correlation_id）   │
│   3. await Future（等回复）                                       │
└─────────────────────────┬───────────────────────────────────────┘
                          │
                    RabbitMQ Exchange → Queue
                          │
┌─────────────────────────▼───────────────────────────────────────┐
│ Worker（LLMRequestWorker）                                       │
│   4. 接收请求 → 调 LLM API                                       │
│   5. 发回复到 reply_to 队列（带相同 correlation_id）               │
│   6. 默认 Exchange 路由回 Producer 的临时队列                     │
└─────────────────────────────────────────────────────────────────┘

几个关键设计决策（面试常考）：
1. 临时队列 vs 固定队列
   - 每个请求创建专属队列 → 完全隔离，一个超时不影响其他
   - auto_delete → 连接断开自动清理，不泄漏
   - 代价：每次多一次 declare_queue 调用（微秒级，可忽略）

2. correlationId 的作用
   - 防止"请求A 收到 请求B 的回复"
   - Producer 用 Future dict 匹配，不匹配的消息直接丢弃

3. 超时降级
   - 60s 没收到回复 → TimeoutError → 外层 catch 后直连
   - 防止 MQ/Worker 卡死导致整个请求挂起

4. 流式不走 MQ（但走令牌桶限流）
   - MQ 的流式支持有限（需要 STREAM 插件或大量消息堆积）
   - 流式场景直连 LLM API，但通过令牌桶控制速率
   - 配合 AgentLoop._dispatch 的 concurrency gate（默认 3 并发）两层级联削峰
"""
from __future__ import annotations

import asyncio
import json
import os
import time
import uuid
from typing import Any

from loguru import logger


# ─── 令牌桶：流式请求速率限制 ─────────────────────────

class _TokenBucket:
    """异步令牌桶——控制流式 LLM 请求速率。

    与 asyncio.Semaphore 的区别：
    - Semaphore 控制**并发数**（同时 N 个）
    - TokenBucket 控制**速率**（每秒 N 个），突发时排队等待

    默认值：rate=5/s, capacity=3（允许小突发，但不允许持续高压）
    """

    def __init__(self, rate: float = 5.0, capacity: int = 3):
        self._rate = rate
        self._capacity = capacity
        self._tokens = float(capacity)
        self._last_refill = 0.0
        self._lock = asyncio.Lock()

    async def acquire(self, timeout: float = 30.0) -> None:
        """消费一个令牌，不足则阻塞直到 timeout 秒后超时。"""
        deadline = time.monotonic() + timeout
        while True:
            async with self._lock:
                now = time.monotonic()
                if self._last_refill == 0.0:
                    self._last_refill = now
                elapsed = now - self._last_refill
                self._tokens = min(self._capacity, self._tokens + elapsed * self._rate)
                self._last_refill = now
                if self._tokens >= 1.0:
                    self._tokens -= 1.0
                    return
            remaining = deadline - time.monotonic()
            if remaining <= 0:
                logger.warning("[TokenBucket] 流式请求等待令牌超时")
                return  # 超时也放行，不让单次超时杀死请求
            await asyncio.sleep(min(0.05, remaining))


class MQLLMProvider:
    """MQ 限流的 LLM Provider（透明代理）。

    用法：在 Orchestrator 中将此 Provider 注入 AgentLoop，AgentLoop 无感知。

    覆盖的方法：
    - chat_with_retry()    → 走 MQ（prefetch=1 限流）
    - chat_stream_with_retry() → 走直连（流式不适合 MQ）
    - get_default_model()  → 透传原始 provider
    """

    def __init__(self, fallback_provider: Any, mq_config: dict):
        """
        Args:
            fallback_provider: 原始 LLMProvider 实例（MQ 不可用时直连）
            mq_config: RabbitMQ 配置字典（由 get_mq_config() 返回）
        """
        self._fallback = fallback_provider
        self._url = mq_config.get("url", "amqp://guest:guest@localhost:5672/")
        self._exchange = mq_config.get("exchange", "agent.tasks")
        self._timeout = mq_config.get("connection_timeout", 5.0)
        self._llm_timeout = 60.0  # LLM API 调用总超时（秒）

        # 令牌桶：流式请求速率限制（环境变量可覆盖）
        # JDAGENT_STREAM_RATE=5   → 每秒最多 5 个新流式请求
        # JDAGENT_STREAM_BURST=3  → 允许突发 3 个并发
        self._stream_bucket = _TokenBucket(
            rate=float(os.environ.get("JDAGENT_STREAM_RATE", "5")),
            capacity=int(os.environ.get("JDAGENT_STREAM_BURST", "3")),
        )

    async def chat_with_retry(self, messages, tools=None, model=None, **kwargs):
        """通过 MQ 调用 LLM，失败时降级到直连。

        使用 RPC over MQ 模式（见模块文档）。

        流程：
        1. 创建临时回复队列
        2. 发布消息（含 reply_to + correlation_id）
        3. 等回复（超时抛异常，外层 catch 后直连）
        """
        try:
            return await self._call_via_mq(messages, tools, model)
        except Exception as e:
            logger.warning(f"[MQLLM] MQ 调用失败 ({type(e).__name__}: {e})，降级直连")
            return await self._fallback.chat_with_retry(
                messages, tools=tools, model=model, **kwargs,
            )

    async def chat_stream_with_retry(self, messages, tools=None, model=None,
                                      **kwargs):
        """流式调用走直连 + 令牌桶限流。

        两层限流（方案 D — 混合模式）：
        1. AgentLoop._dispatch 的 concurrency gate（默认 3 并发）
        2. 此处令牌桶（默认 5/s 速率，3 突发）

        令牌桶等不到令牌时等待 30s，超时仍放行（不让限流杀死请求）。
        """
        await self._stream_bucket.acquire(timeout=30.0)
        return await self._fallback.chat_stream_with_retry(
            messages, tools=tools, model=model, **kwargs,
        )

    def get_default_model(self) -> str:
        return self._fallback.get_default_model()

    # ─── MQ RPC 核心 ───────────────────────────────────

    async def _call_via_mq(self, messages, tools, model) -> Any:
        """通过 MQ 发布 LLM 请求并等待回复（RPC 模式）。

        使用临时队列（exclusive + auto_delete），每个请求独立队列，
        防止消息混淆。

        回复匹配：通过 correlation_id（UUID）唯一标识。
        超时：self._llm_timeout 后抛 TimeoutError。
        """
        import aio_pika

        # 短连接：每次 LLM 调用单独连接（简单可靠，不用维护连接池）
        connection = await aio_pika.connect_robust(
            self._url, timeout=self._timeout,
        )
        async with connection:
            channel = await connection.channel()

            # Step 1: 创建临时回复队列
            # exclusive=True   = 只此连接可见，其他连接不能发（安全）
            # auto_delete=True = 连接断开自动删除（不泄漏资源）
            reply_queue = await channel.declare_queue(
                None, exclusive=True, auto_delete=True,
            )

            # Step 2: 准备 Future + correlation_id
            correlation_id = str(uuid.uuid4())
            event_loop = asyncio.get_event_loop()
            future = event_loop.create_future()

            # Step 3: 消费回复队列
            # 收到消息时检查 correlation_id 是否匹配
            # no_ack=True: 自动确认回复消息（回复是轻量的，不需要手动 ack）
            async def on_reply(message: aio_pika.IncomingMessage) -> None:
                if message.correlation_id == correlation_id:
                    future.set_result(message.body)

            await reply_queue.consume(on_reply, no_ack=True)

            # Step 4: 发布 LLM 请求到 llm_request 队列
            # 关键字段：
            # - reply_to: 告诉 Worker"结果回哪个队列"
            # - correlation_id: 标识请求唯一性，防回复错乱
            request_data = {
                "messages": _safe_truncate(messages),
                "tools": tools,
                "model": model,
            }
            msg = aio_pika.Message(
                body=json.dumps(request_data, ensure_ascii=False).encode(),
                reply_to=reply_queue.name,
                correlation_id=correlation_id,
                delivery_mode=aio_pika.DeliveryMode.PERSISTENT,
                content_type="application/json",
            )
            exchange = await channel.declare_exchange(
                self._exchange,
                aio_pika.ExchangeType.DIRECT,
                durable=True,
            )
            await exchange.publish(msg, routing_key="llm_request")

            # Step 5: 等待回复（带超时）
            # 超时 → 外层 chat_with_retry 的 except 捕获 → 直连
            raw = await asyncio.wait_for(future, timeout=self._llm_timeout)
            result = json.loads(raw.decode())

            # 检查 Worker 是否返回错误
            if "error" in result:
                raise RuntimeError(f"LLM Worker 返回错误: {result['error']}")

            # 反序列化为 LLMResponse
            from agent.core.types import LLMResponse, ToolCallRequest

            tool_calls = None
            if result.get("tool_calls"):
                tool_calls = [
                    ToolCallRequest(**tc) for tc in result["tool_calls"]
                ]

            return LLMResponse(
                content=result.get("content"),
                tool_calls=tool_calls,
                finish_reason=result.get("finish_reason", "stop"),
                usage=result.get("usage"),
            )


def _safe_truncate(messages: list, max_chars: int = 80000) -> list:
    """截断长 content，防止 MQ 消息体超限。

    RabbitMQ 默认消息大小限制为 128MB。
    但 LLM 消息列表可能很大（含工具执行结果等），
    这里主动截断，避免触发 MQ 的限制。
    """
    truncated = []
    for m in messages:
        entry = dict(m)
        if isinstance(entry.get("content"), str) and len(entry["content"]) > max_chars:
            entry["content"] = entry["content"][:max_chars] + "\n...(truncated for MQ)"
        truncated.append(entry)
    return truncated
