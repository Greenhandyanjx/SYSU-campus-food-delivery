"""
LLM Request Worker — RabbitMQ RPC Worker

专门消费 llm_request 队列，调用 LLM API，通过 Direct Reply-To 返回结果。

为什么独立于 MQConsumer：
┌──────────────────────┬──────────────────────────────┐
│ MQConsumer           │ LLMRequestWorker             │
├──────────────────────┼──────────────────────────────┤
│ Fire-and-forget 任务  │ RPC 模式（请求→回复）          │
│ Dream/Consolidation  │ 需要 reply_to 机制            │
│ 不需要发回结果         │ 需要独立的 LLM Provider        │
│ 共享 AgentLoop 资源   │ 资源隔离，崩溃不影响其他任务     │
└──────────────────────┴──────────────────────────────┘

面试重点 — Direct Reply-To 机制：
Worker 不需要提前声明回复队列。Producer 创建一个临时队列，
把队列名放在消息的 reply_to 属性里。Worker 处理完后，
直接把结果发送到 reply_to 指定的队列（使用默认 Exchange，
routing_key = reply_to 队列名）。
RabbitMQ 会自动把消息路由到 Producer 的临时队列。

限流原理：
队列设置 prefetch_count=1 → Worker 同时只处理一个 LLM 请求
→ 天然控制 LLM 调用并发度（无论多少个用户同时发请求）
"""
from __future__ import annotations

import asyncio
import json
from typing import TYPE_CHECKING, Any

from loguru import logger

if TYPE_CHECKING:
    import aio_pika


class LLMRequestWorker:
    """LLM RPC Worker。

    消费 llm_request 队列，调用 LLM API，把结果发回 reply_to 队列。

    用法：
        worker = LLMRequestWorker(mq_config, provider, model="deepseek-chat")
        asyncio.create_task(worker.start())
    """

    def __init__(self, config: dict, provider: Any, model: str):
        """
        Args:
            config: RabbitMQ 配置字典
            provider: LLMProvider 实例（必须是原始 provider，不是 MQLLMProvider）
            model: 默认模型名（请求未指定时使用）
        """
        self._url = config.get("url", "amqp://guest:guest@localhost:5672/")
        self._exchange_name = config.get("exchange", "agent.tasks")
        self._timeout = config.get("connection_timeout", 5.0)
        self._provider = provider
        self._model = model
        self._running = False
        self._channel = None  # 在 _consume_loop 中赋值

    async def start(self) -> None:
        """启动 Worker（长连接，持续消费，自动重连）。"""
        self._running = True
        logger.info("[LLM Worker] 启动...")

        while self._running:
            try:
                await self._consume_loop()
            except asyncio.CancelledError:
                logger.info("[LLM Worker] 收到取消信号")
                break
            except Exception as e:
                logger.warning(f"[LLM Worker] 连接断开 ({e})，3s 后重连...")
                await asyncio.sleep(3)

        logger.info("[LLM Worker] 已停止")

    async def _consume_loop(self) -> None:
        """消费循环：连接到 RabbitMQ → 声明队列 → 消费。

        使用 connect_robust() 自动重连。
        async with connection 确保退出时连接正确关闭。

        prefetch_count=1：这是核心限流点。
        一次只处理一个 LLM 请求，其他请求在队列里等待。
        """
        import aio_pika

        connection = await aio_pika.connect_robust(
            self._url, timeout=self._timeout,
        )
        async with connection:
            # 保存 channel 引用供 _on_request 使用
            # 注意：message.channel 返回的是 aiormq 底层 Channel，
            # 不是 aio-pika 的 RobustChannel，没有 default_exchange 属性。
            # 所以这里保存 self._channel 用于回复消息。
            channel = await connection.channel()
            self._channel = channel

            # QoS：一次只拉 1 条 → 全局限流
            # 无论多少用户同时请求，LLM 调用并行度始终为 1
            #（可以调大，但 1 = 最严格的费用控制）
            await channel.set_qos(prefetch_count=1)

            # 声明 Exchange（确保存在）
            exchange = await channel.declare_exchange(
                self._exchange_name,
                aio_pika.ExchangeType.DIRECT,
                durable=True,
            )

            # 声明队列（durable=True：RabbitMQ 重启不丢队列）
            queue = await channel.declare_queue(
                "agent.llm.request",
                durable=True,
            )

            # 绑定队列到 Exchange
            await queue.bind(exchange, routing_key="llm_request")

            # 开始消费
            await queue.consume(self._on_request)

            logger.info(f"[LLM Worker] 开始消费 agent.llm.request (prefetch=1)")

            # 阻塞直到停止
            while self._running:
                await asyncio.sleep(1)

    async def _on_request(self, message: "aio_pika.IncomingMessage") -> None:
        """处理单条 LLM 请求：调 API → 序列化 → 发回复。

        Runnable 在 aio-pika 的 channel 中执行，不会被并发调用
        （因为 prefetch_count=1）。

        回复流程：
        1. 调 self._provider.chat() 调用 LLM API
        2. 将结果序列化为 JSON
        3. 使用消息的 reply_to 作为 routing_key 发回
        4. ack 原始消息（告诉 RabbitMQ 删除它）
        """
        import aio_pika

        data = {}
        try:
            # ── 解析请求 ──
            data = json.loads(message.body.decode())
            messages = data.get("messages", [])
            tools = data.get("tools")
            model = data.get("model") or self._model

            if not messages:
                raise ValueError("请求中无 messages")

            # ── 调用 LLM API ──
            logger.debug(f"[LLM Worker] 调用 {model} (messages={len(messages)})")
            response = await self._provider.chat(
                messages, tools=tools, model=model,
            )

            # ── 序列化结果 ──
            result: dict = {
                "content": response.content,
                "finish_reason": response.finish_reason,
                "usage": response.usage,
            }
            if response.tool_calls:
                result["tool_calls"] = [
                    {
                        "id": tc.id,
                        "name": tc.name,
                        "arguments": tc.arguments,
                    }
                    for tc in response.tool_calls
                ]

            reply_body = json.dumps(result, ensure_ascii=False).encode()

        except Exception as e:
            logger.error(f"[LLM Worker] 处理失败: {type(e).__name__}: {e}")
            reply_body = json.dumps({"error": str(e)}).encode()

        # ── 发回复到 reply_to 队列 ──
        # 使用默认 Exchange（空字符串），routing_key = reply_to 队列名
        # RabbitMQ 自动路由到 Producer 的临时队列
        try:
            reply_msg = aio_pika.Message(
                body=reply_body,
                correlation_id=message.correlation_id,
                content_type="application/json",
            )
            # 使用 self._channel.default_exchange（不是 message.channel）
            # message.channel 是 aiormq 底层 Channel，不含 default_exchange
            await self._channel.default_exchange.publish(
                reply_msg, routing_key=message.reply_to,
            )
        except Exception as e:
            logger.warning(f"[LLM Worker] 回复失败 (reply_to={message.reply_to}): {e}")
            # 回复失败通常因为 Producer 已超时断开（临时队列已删）
            # 属于正常情况，日志记录即可

        # 确认原始消息（无论处理成功与否，都已尝试返回结果）
        await message.ack()

    def stop(self) -> None:
        """通知 Worker 停止。"""
        self._running = False
        logger.info("[LLM Worker] 收到停止信号")
