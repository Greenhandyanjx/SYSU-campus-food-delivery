"""
MQ Consumer — RabbitMQ 异步消费者（Worker）

职责：
1. 长连接 RabbitMQ，持续消费 3 个队列的消息
2. 反序列化消息 → 按 routing_key 分发到对应的 handler
3. 自动重连 + 任务重试

启动方式（由 Orchestrator 在后台启动）：
    consumer = MQConsumer(config, agent_loop)
    asyncio.create_task(consumer.start())

消息确认机制（面试常考）：
- ACK（确认）：消费者处理成功后，告诉 Broker 可以删除消息了
- NACK（拒绝）：处理失败，Broker 可以选择重新入队（requeue）或丢弃
- 如果消费者连接断开（崩溃），所有未 ACK 的消息自动重新入队

重试策略：
- 临时失败 → nack + requeue，最多重试 max_retries 次
  （基于 RabbitMQ 的 redelivered 标志——消息被重新投递时 Broker 置此标志）
- 永久失败 → ack + 日志告警（不进死信，避免队列积压）

对比 nanobot 的设计：
nanobot 内部没有独立的 MQ 系统——它的"异步"通过 asyncio.create_task 实现。
这里引入 RabbitMQ 后，后台任务可以跨进程（甚至跨机器）执行，
为未来拆分为独立 Worker 进程做准备。
"""
from __future__ import annotations

import asyncio
import json
from typing import TYPE_CHECKING, Any

from loguru import logger

if TYPE_CHECKING:
    import aio_pika

    from agent.loop import AgentLoop


class MQConsumer:
    """RabbitMQ 消费者（后台 Worker）。

    架构说明：
    1. 使用 aio_pika.connect_robust() 建立长连接（自动处理断线重连）
    2. 使用 channel.set_qos(prefetch_count=1) —— 每次只取 1 条消息
       原因：任务之间无依赖，但 Dream 和 Consolidation 都是 LLM 调用密集型，
       prefetch=1 防止一个 Worker 拉取过多消息导致处理延迟。
    3. 每个 routing_key 对应独立的 handler 函数

    自动重连：
    - connect_robust() 在连接断开后自动重连
    - 重连后需要重新声明 Exchange + Queue + Binding
      所以整个声明逻辑在 while 循环中（断开 → 重连 → 重新声明）
    """

    def __init__(self, config: dict, agent_loop: "AgentLoop"):
        """
        Args:
            config: MQ 配置字典（由 get_mq_config() 加载）
            agent_loop: AgentLoop 实例引用——worker 通过它访问
                       ChatMemory、MemoryConsolidator、SessionManager 等组件
        """
        self.config = config
        self.agent_loop = agent_loop
        self._running = False

        # 从配置读取
        mq_cfg = config.get("rabbitmq", config)
        self._url = mq_cfg.get("url", "amqp://guest:guest@localhost:5672/")
        self._exchange_name = mq_cfg.get("exchange", "agent.tasks")
        self._exchange_type = mq_cfg.get("exchange_type", "direct")
        self._reconnect_interval = mq_cfg.get("reconnect_interval", 3.0)
        self._max_retries = mq_cfg.get("retry", {}).get("max_retries", 3)

        # 延迟导入 Handler（避免 TYPE_CHECKING 的循环依赖在运行时生效）
        self._handler_map = None

    async def start(self) -> None:
        """启动消费者（长连接，持续阻塞直到被取消）。

        此方法运行一个无限循环：
        while running:
            连接 RabbitMQ → 声明 Exchange + Queue → 消费消息
            连接断开 → 自动重连

        被调用者通过 cancel() 停止。
        """
        from agent.mq.tasks import HANDLER_MAP, ROUTING_KEY_MAP

        self._handler_map = HANDLER_MAP
        self._routing_key_map = ROUTING_KEY_MAP
        self._running = True

        logger.info("[MQ Consumer] 启动 Worker...")

        while self._running:
            try:
                await self._consume_loop()
            except asyncio.CancelledError:
                logger.info("[MQ Consumer] 收到取消信号，退出")
                break
            except Exception as e:
                logger.warning(
                    f"[MQ Consumer] 连接断开 ({e})，"
                    f"{self._reconnect_interval}s 后重连..."
                )
                await asyncio.sleep(self._reconnect_interval)

        logger.info("[MQ Consumer] Worker 已停止")

    async def _consume_loop(self) -> None:
        """消费循环：连接 → 声明 → 消费。

        使用 connect_robust() 建立自动重连连接。
        async with connection: 确保退出时连接被正确关闭。

        在 connection 的上下文中创建 channel 和 consumer，
        如果连接断开，consumer 自动停止；重连后需要重新调用此方法。
        """
        import aio_pika

        # connect_robust：自动重连连接
        # 连接断开后会自动尝试重连（默认重试间隔由 aio-pika 内部管理）
        connection = await aio_pika.connect_robust(
            self._url,
            timeout=self.config.get("connection_timeout", 5.0),
        )

        async with connection:
            logger.info(f"[MQ Consumer] 已连接到 {self._url}")

            # 创建 channel
            # 一个 connection 可以创建多个 channel（类比：一根光纤多个信道）
            channel = await connection.channel()

            # QoS：prefetch_count=1
            # 每次只给这个 consumer 发 1 条消息，处理完再发下一条
            # 避免某个 queue 消息堆积时，一个 consumer 拉取过多消息
            await channel.set_qos(prefetch_count=1)

            # 声明 Exchange（direct 类型）
            exchange = await channel.declare_exchange(
                name=self._exchange_name,
                type=aio_pika.ExchangeType(self._exchange_type),
                durable=True,       # Exchange 持久化（Broker 重启不丢失）
            )

            # 声明队列并绑定
            queues_cfg = self.config.get("queues", {})
            for queue_name, q_cfg in queues_cfg.items():
                # 声明队列
                # durable=True：队列持久化（重启不丢失）
                queue = await channel.declare_queue(
                    name=f"agent.{queue_name}",
                    durable=q_cfg.get("durable", True),
                )

                # 绑定队列到 Exchange
                # binding 建立 Exchange → Queue 的路由关系
                routing_key = q_cfg.get("routing_key", queue_name)
                await queue.bind(
                    exchange=exchange,
                    routing_key=routing_key,
                )

                # 启动消费
                # 当消息到达队列时，回调函数 _on_message_factory(routing_key) 被调用
                await queue.consume(
                    callback=self._on_message_factory(routing_key),
                    # no_ack=False：手动 ACK（默认）
                    # 消费者处理完消息后必须显式调用 message.ack()
                    # 如果连接断开，未 ACK 的消息自动重新入队
                )

                prefetch = q_cfg.get("prefetch", 1)
                logger.info(
                    f"[MQ Consumer] 队列 agent.{queue_name} 已就绪: "
                    f"routing_key={routing_key}, prefetch={prefetch}"
                )

            logger.info(f"[MQ Consumer] 开始消费 {len(queues_cfg)} 个队列")

            # 阻塞等待，直到连接断开
            # 注意：这里不能使用 asyncio.Future() 直接等待，
            # 因为 robust 连接断开后会自动重连，不会抛出异常。
            # 使用 self._running 作为信号量，在 cancel 时退出。
            while self._running:
                await asyncio.sleep(1)

    def _on_message_factory(self, routing_key: str) -> Any:
        """创建消息回调函数（闭包绑定 routing_key）。

        每个队列的 callback 都需要知道自己的 routing_key，
        使用闭包而不是在每个消息体里查 routing_key 效率更高。

        回调的职责：
        1. 反序列化消息体（JSON → dict）
        2. 查找对应的 handler
        3. 执行 handler
        4. ACK（成功）或 NACK（失败+重试）
        """
        async def _on_message(message: "aio_pika.IncomingMessage") -> None:
            """单条消息的回调处理。

            message.process() 上下文管理器：
            - 正常退出 → 自动 ACK
            - 异常退出 → 自动 NACK + requeue
            - 具有幂等性保护（ignore_processed=True 避免重复处理）

            对于可恢复错误（如网络超时），我们用 nack(requeue=True)
            让消息重新入队，后续可以重新消费。
            对于不可恢复错误（如数据格式错误），ack 丢弃。
            """
            task_type = self._routing_key_map.get(routing_key)
            if task_type is None:
                logger.warning(f"[MQ Consumer] 未知 routing_key: {routing_key}，丢弃消息")
                await message.ack()
                return

            handler = self._handler_map.get(task_type)
            if handler is None:
                logger.warning(f"[MQ Consumer] 无 handler 处理: {task_type}，丢弃消息")
                await message.ack()
                return

            # 反序列化
            try:
                body = json.loads(message.body.decode("utf-8"))
            except (json.JSONDecodeError, UnicodeDecodeError) as e:
                logger.error(f"[MQ Consumer] 消息解码失败: {e}，丢弃消息")
                await message.ack()
                return

            # 执行 handler
            try:
                await handler(body, self.agent_loop)
                # handler 成功 → ACK（RabbitMQ 删除此消息）
                await message.ack()
                logger.debug(f"[MQ Consumer] 任务完成: {routing_key}")

            except Exception as e:
                logger.error(f"[MQ Consumer] 任务失败: routing_key={routing_key}: {e}")

                # 用 redelivered 标志判断重试次数
                # redelivered=True 意味着这是至少第二次投递（上一次 nack/超时后重新入队）
                if message.redelivered:
                    # 已经重试过一次了，放弃（避免无限重试循环）
                    logger.warning(f"[MQ Consumer] 消息已重试仍失败，丢弃: {routing_key}")
                    await message.ack()
                else:
                    # 首次失败，nack + requeue（让其他 Worker 或后续消费重试）
                    # requeue=True：消息重新进入队列头部
                    # requeue=False：消息被丢弃（或进入死信队列）
                    logger.info(f"[MQ Consumer] 消息 requeue: {routing_key}")
                    await message.nack(requeue=True)

        return _on_message

    def stop(self) -> None:
        """通知消费者停止（设置标志位，等待当前消息处理完退出）。"""
        self._running = False
        logger.info("[MQ Consumer] 收到停止信号")
