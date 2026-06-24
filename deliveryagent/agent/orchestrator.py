"""
Agent Orchestrator: Agent 编排器
参考 nanobot 的 bus/queue.py + agent/loop.py 设计

编排器是系统的统一入口：
1. 初始化 MessageBus + AgentLoop + Provider + Tools
2. 提供 chat() / chat_stream() 接口给外部调用
3. 管理 Agent 生命周期

核心改变（参考 nanobot 后）：
- 使用 MessageBus 解耦输入/输出
- AgentLoop 处理所有消息
- Provider 抽象化，支持切换
- 新的三层记忆系统
"""

import asyncio
from pathlib import Path
from typing import AsyncGenerator

from loguru import logger

from agent.core.types import InboundMessage, OutboundMessage
from agent.core.llm_provider import LLMProvider
from agent.bus.queue import MessageBus
from agent.cache.redis_cache import get_redis_cache
from agent.loop import AgentLoop
from agent.mq import get_mq_config
from agent.mq.producer import MQProducer
from agent.mq.consumer import MQConsumer
from agent.mq.llm_provider import MQLLMProvider
from agent.mq.llm_worker import LLMRequestWorker
from agent.tools.registry import ToolRegistry


class AgentOrchestrator:
    """
    Agent 编排器。
    
    负责：
    1. 初始化所有组件（总线、循环、Provider、工具）
    2. 提供同步/异步的 chat() 接口
    3. 管理工具注册
    4. 管理技能系统（Skill 自动发现与注册）
    """

    _instance = None

    def __init__(
        self,
        provider: LLMProvider | None = None,
        workspace: str | None = None,
        skills_dir: str | None = None,
        pg_config: dict | None = None,
    ):
        self.bus = MessageBus()
        self.provider = provider
        self.workspace = Path(workspace or Path.cwd()).expanduser().resolve()
        self.skills_dir = Path(skills_dir).expanduser().resolve() if skills_dir else None
        self.pg_config = pg_config
        self.redis_cache = None
        self.loop: AgentLoop | None = None
        self._task: asyncio.Task | None = None

        # MQ（RabbitMQ 可选——连接失败不阻塞启动）
        self.mq_config: dict | None = None
        self.mq_producer: MQProducer | None = None
        self.mq_consumer: MQConsumer | None = None
        self._mq_task: asyncio.Task | None = None
        self._llm_worker: LLMRequestWorker | None = None
        self._llm_worker_task: asyncio.Task | None = None

    # ─── 初始化 ────────────────────────────────

    async def initialize(self) -> None:
        """
        初始化 Agent 系统。

        必须先调用此方法才能使用 Agent。
        初始化后，技能系统也会自动发现并注册。

        Redis 缓存是可选组件：如果 Redis 不可用（未配置或连接失败），
        所有缓存操作静默降级（get→None, set→noop），不会影响业务逻辑。
        """
        # 初始化 Redis 缓存
        # get_redis_cache() 会从 config/cache.yml 或环境变量 REDIS_URL 读取配置。
        # 如果 Redis 未配置，返回一个空实例（所有操作静默失败）。
        self.redis_cache = get_redis_cache()

        # ── MQ 配置加载（放在 AgentLoop 创建之前，因为要注入 provider）──
        self.mq_config = get_mq_config()

        # 构造 AgentLoop（如果 MQ 配置了，provider 被 MQLLMProvider 透明包装）
        loop_provider = self.provider
        if self.mq_config and self.provider:
            loop_provider = MQLLMProvider(self.provider, self.mq_config)
            logger.info("[Orchestrator] LLM 调用将通过 MQ 限流")

        self.loop = AgentLoop(
            bus=self.bus,
            provider=loop_provider,
            workspace=self.workspace,
            skills_dir=self.skills_dir,
            pg_config=self.pg_config,
            redis_cache=self.redis_cache,
        )
        self._task = asyncio.create_task(self.loop.run())

        # 异步初始化：PG 连接池 → 技能自动发现
        await self.loop.initialize()

        # ── MQ 后台任务 Worker 初始化 ──
        if self.mq_config:
            # ① 生产者（供 _process_message 发布 Dream/Consolidation 任务）
            self.mq_producer = MQProducer(self.mq_config)
            self.loop.mq_producer = self.mq_producer

            # ② MQConsumer：消费 Dream + Consolidation 队列
            self.mq_consumer = MQConsumer(self.mq_config, agent_loop=self.loop)
            self._mq_task = asyncio.create_task(self.mq_consumer.start())

            # ③ LLM RPC Worker：消费 llm_request 队列
            # 注意：使用原始 provider（self.provider），不是 MQLLMProvider
            # 否则 Worker 调 LLM 会再走 MQ→自己→死循环
            default_model = self.provider.get_default_model() if self.provider else "deepseek-chat"
            self._llm_worker = LLMRequestWorker(
                self.mq_config,
                provider=self.provider,
                model=default_model,
            )
            self._llm_worker_task = asyncio.create_task(self._llm_worker.start())
            logger.info("[Orchestrator] RabbitMQ + LLM Worker 已在后台启动")
        else:
            logger.info("[Orchestrator] RabbitMQ 未配置，任务将同步执行")

        logger.info("[Orchestrator] Agent 初始化完成")

    # ─── 同步接口 ──────────────────────────────

    def chat(self, user_query: str) -> str:
        """
        同步聊天接口。
        
        适用于 CLI、Streamlit 等简单调用场景。
        内部实际走异步路径。
        """
        return asyncio.run(self.chat_async(user_query))

    async def chat_async(self, user_query: str, session_key: str = "cli:direct", jwt_token: str = "") -> str:
        """
        异步聊天接口（非流式）。

        构造消息 → 直接处理 → 返回结果。
        """
        if self.loop is None:
            await self.initialize()

        metadata = {}
        if jwt_token:
            metadata["jwt_token"] = jwt_token

        response = await self.loop.process_direct(
            content=user_query,
            session_key=session_key,
            channel="cli",
            chat_id="direct",
            metadata=metadata,
        )
        return response.content if response else ""

    async def chat_stream_async(self, user_query: str, session_key: str = "cli:direct", jwt_token: str = "") -> AsyncGenerator[str, None]:
        """
        异步流式聊天接口。

        走 _dispatch 路径（含 per-session lock + concurrency gate）：
          1. 构造 InboundMessage（统一 session_key 推导 channel:chat_id）
          2. 挂载 on_stream 回调（通过 asyncio.Queue 实现实时 yield）
          3. 通过 loop._dispatch() 处理全部流程（session、history、messages、ReAct、持久化）
          4. 逐块 yield 流式内容

        使用 asyncio.Queue 桥接回调 → 生成器，确保前端能实时收到每一块文本。
        """
        if self.loop is None:
            await self.initialize()

        # 从 session_key 推导 channel:chat_id，保持与 _dispatch 的 session_key 一致
        parts = session_key.split(":", 1)
        channel = parts[0] if len(parts) == 2 else "cli"
        chat_id = parts[1] if len(parts) == 2 else session_key

        metadata = {}
        if jwt_token:
            metadata["jwt_token"] = jwt_token

        msg = InboundMessage(
            channel=channel,
            sender_id="user",
            chat_id=chat_id,
            content=user_query,
            metadata=metadata,
        )

        # asyncio.Queue 桥接：on_stream 回调放入 → 生成器 yield 取出
        queue: asyncio.Queue[str | None] = asyncio.Queue()
        _streamed_anything = False  # 标记是否真的发过流式数据

        async def on_stream(delta: str) -> None:
            nonlocal _streamed_anything
            _streamed_anything = True
            await queue.put(delta)

        async def _run_and_signal():
            """在后台任务中运行 _dispatch，完成后发送结束信号"""
            try:
                result = await self.loop._dispatch(msg, on_stream=on_stream)
                # 仅当完全没有流式数据时，回退到完整内容
                if not _streamed_anything and result and result.content:
                    await queue.put(result.content)
            except Exception as e:
                logger.error(f"[Orchestrator] 流式处理失败: {e}")
                await queue.put(f"错误: {e}")
            finally:
                await queue.put(None)  # sentinel: 流结束

        # 在后台运行 _dispatch，当前生成器从 queue 消费
        task = asyncio.create_task(_run_and_signal())

        while True:
            chunk = await queue.get()
            if chunk is None:
                break
            yield chunk

        # 重新抛出台 dispatch 中的异常（如果有）
        await task

    # ─── 工具管理 ──────────────────────────────

    def get_tool_registry(self) -> ToolRegistry:
        """获取工具注册中心（用于注册外部工具）"""
        if self.loop:
            return self.loop.tools
        return None

    def get_all_tools(self) -> list[dict]:
        """获取所有已注册工具的 Schema"""
        if self.loop:
            return self.loop.tools.get_definitions()
        return []

    def get_tool_names(self) -> list[str]:
        """获取所有工具名称"""
        if self.loop:
            return self.loop.tools.tool_names
        return []

    # ─── 技能管理 ──────────────────────────────

    def get_all_skills(self) -> list[dict]:
        """获取所有已注册技能的字典表示"""
        if self.loop:
            return self.loop.skill_manager.to_dict()
        return []

    def get_skill_count(self) -> int:
        """获取技能数量"""
        if self.loop:
            return self.loop.skill_manager.count
        return 0

    def get_skill_names(self) -> list[str]:
        """获取所有技能名称"""
        if self.loop:
            return [s.name for s in self.loop.skill_manager.get_all()]
        return []

    # ─── 会话管理 ──────────────────────────────

    def clear_session(self) -> None:
        """清空当前会话"""
        if self.loop:
            session = self.loop.sessions.get_or_create("cli:direct")
            session.clear()

    def get_session_history(self) -> list[dict]:
        """获取当前会话历史"""
        if self.loop:
            session = self.loop.sessions.get_or_create("cli:direct")
            return session.get_history()
        return []

    # ─── 生命周期 ──────────────────────────────

    async def shutdown(self) -> None:
        """关闭 Agent（停止循环 → 关闭 PG/Redis/MQ → 取消任务）"""
        # 停止 LLM RPC Worker
        if self._llm_worker:
            self._llm_worker.stop()
        if self._llm_worker_task:
            self._llm_worker_task.cancel()
            try:
                await self._llm_worker_task
            except asyncio.CancelledError:
                pass

        # 停止 MQ 消费者
        if self.mq_consumer:
            self.mq_consumer.stop()
        if self._mq_task:
            self._mq_task.cancel()
            try:
                await self._mq_task
            except asyncio.CancelledError:
                pass
        # 关闭 MQ 生产者
        if self.mq_producer:
            await self.mq_producer.close()

        if self.loop:
            self.loop.stop()
            # 关闭 PG 连接池
            await self.loop.sessions.close_pg()
        # 关闭 Redis 连接
        if self.redis_cache:
            await self.redis_cache.close()
        if self._task:
            self._task.cancel()
            try:
                await self._task
            except asyncio.CancelledError:
                pass
        logger.info("[Orchestrator] Agent 已关闭")
