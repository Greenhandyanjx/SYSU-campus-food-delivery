"""
MQ Producer — 异步 RabbitMQ 生产者

职责：
1. 将任务 JSON 序列化后发布到指定的 Exchange + Routing Key
2. 连接失败时静默降级（记录日志，不抛异常）
3. 适配当前代码库的"优雅降级"风格（参考 Redis 缓存）

Connection 策略：
使用短生命周期连接——每次 publish 创建新连接。
原因：producer 的调用频率低（每处理一次用户请求才发 1-2 条），
不值得维护长连接池。如果未来消息量增大，可改为长连接 + 自动重连。

关于消息可靠性的 3 个层次（面试常考）：
1. Message Persistence（消息持久化）
   - delivery_mode=2（PERSISTENT）：消息写入磁盘，RabbitMQ 重启不丢
   - 配合队列 durable=True，即使 Broker 崩溃也能恢复
2. Publisher Confirm（发布者确认）
   - 生产者等 Broker 返回 confirm（确保消息到达 Exchange）
   - 类似 TCP ACK，但由 AMQP 协议层保证
3. Consumer ACK（消费者确认）
   - 消费者处理完消息后发送 ACK，Broker 才删除消息
   - 如果消费者崩溃（连接断开），未 ACK 的消息自动重新入队
"""
from __future__ import annotations

import json
from typing import Any

from loguru import logger

from agent.mq.tasks import TaskType


class MQProducer:
    """RabbitMQ 生产者封装。

    用法：
        producer = MQProducer(config)
        await producer.publish("dream", {"session_key": "..."})

    设计目标：
    - 简单可靠：每个 publish 独立连接，用完即关
    - 优雅降级：RabbitMQ 不可用时静默跳过（不阻塞业务）
    - 消息持久化：确保 RabbitMQ 重启后消息不丢失
    """

    def __init__(self, config: dict):
        """
        Args:
            config: MQ 配置字典（由 get_mq_config() 加载）
                关键字段:
                - url: amqp 连接字符串
                - exchange: Exchange 名称
                - exchange_type: "direct"
        """
        self.config = config
        self._rabbitmq_url = config.get("url", "amqp://guest:guest@localhost:5672/")
        self._exchange_name = config.get("exchange", "agent.tasks")
        self._exchange_type = config.get("exchange_type", "direct")

    async def publish(self, routing_key: str, data: dict) -> bool:
        """发布一条任务消息到 RabbitMQ。

        Args:
            routing_key: 路由键（"dream" / "consolidation"）
            data: 任务数据（JSON 可序列化）。注意：不要包含敏感信息。

        Returns:
            bool: 是否成功投递
                True  = 投递成功（消息已进入 RabbitMQ）
                False = 投递失败（MQ 不可用，静默降级）
        """
        try:
            import aio_pika

            # 建立连接（带超时）
            connection = await aio_pika.connect_robust(
                self._rabbitmq_url,
                timeout=self.config.get("connection_timeout", 5.0),
            )

            async with connection:
                # 创建 channel（AMQP 通道——可以理解为"连接内的轻量级子连接"）
                channel = await connection.channel()

                # 声明 Exchange：确保 Exchange 存在（不存在则自动创建）
                # durable=True：Exchange 持久化（Broker 重启不丢失）
                exchange = await channel.declare_exchange(
                    name=self._exchange_name,
                    type=aio_pika.ExchangeType(self._exchange_type),
                    durable=True,
                )

                # 构造消息体（JSON bytes）
                # delivery_mode=PERSISTENT：写入磁盘，Broker 崩溃不丢消息
                message = aio_pika.Message(
                    body=json.dumps(data, ensure_ascii=False).encode("utf-8"),
                    delivery_mode=aio_pika.DeliveryMode.PERSISTENT,
                    content_type="application/json",
                    content_encoding="utf-8",
                )

                # 发布消息到 Exchange
                # Exchange 根据 routing_key 将消息路由到绑定的 Queue
                await exchange.publish(
                    message=message,
                    routing_key=routing_key,
                    # mandatory=True 意味着如果消息无法路由到任何队列，
                    # RabbitMQ 会将消息返回给 producer（通过 Basic.Return）
                    # 注意：需要 channel 开启 confirm_mode 才能收到返回
                )

            logger.debug(f"[MQ] 消息已发布: routing_key={routing_key}, data={data}")
            return True

        except ImportError:
            logger.warning("[MQ] aio-pika 未安装，消息投递跳过")
            return False
        except Exception as e:
            # 优雅降级：连接失败只记日志，不阻塞业务
            logger.warning(f"[MQ] 消息发布失败 (routing_key={routing_key}): {e}")
            return False

    async def close(self) -> None:
        """关闭生产者（预留接口）。

        当前使用短连接模式，每个 publish 自动关闭连接。
        如果将来改为长连接，需要在此处清理资源。
        """
        pass
