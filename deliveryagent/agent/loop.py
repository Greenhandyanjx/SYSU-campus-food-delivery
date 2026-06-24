"""
Agent Core: ReAct 循环（核心处理引擎）
参考 nanobot 的 agent/loop.py 设计

这是 Agent 的核心：Think → Act → Observe 循环。

流程：
1. 接收用户消息
2. 构建上下文（System Prompt + 历史 + 当前消息）
3. 调用 LLM
4. 如果 LLM 返回工具调用 → 执行工具 → 结果注入 → 回到步骤3
5. 如果 LLM 返回文本 → 作为最终回复返回

关键设计（参考 nanobot）：
- 线程/会话隔离：每个 session 有自己的锁，保证串行处理
- 并发控制：跨 session 可以并发，同 session 串行
- Streaming 支持：LLM 回复可以流式输出给用户
- Progress 通知：工具调用时通知用户"正在做什么"
- 自动重试：调用 LLM 失败时自动重试
- MCP 支持：支持外部 MCP 服务器连接（预留）
"""

import asyncio
import json
import os
import time
from contextlib import AsyncExitStack, nullcontext
from pathlib import Path
from typing import TYPE_CHECKING, Any, Callable

from loguru import logger

from agent.core.types import InboundMessage, OutboundMessage, LLMResponse
from agent.core.llm_provider import LLMProvider
from agent.bus.queue import MessageBus
from agent.session.manager import Session, SessionManager
from agent.tools.registry import ToolRegistry
from agent.tools.filesystem import ReadFileTool, WriteFileTool
from agent.tools.web import WebFetchTool
from agent.memory.context_builder import ContextBuilder
from agent.memory.memory_store import MemoryConsolidator
from agent.memory.dream import Dream
from agent.memory.chat_memory import ChatMemory
from agent.skills.manager import SkillManager
from agent.skills.loader import SkillLoader

if TYPE_CHECKING:
    pass


class AgentLoop:
    """
    Agent 循环：核心处理引擎。
    
    对应 nanobot 的 AgentLoop：
    1. 从 MessageBus 接收消息
    2. Builder 上下文
    3. 调用 LLM
    4. 执行工具调用
    5. 发送响应
    """

    # 工具结果最大长度（防止 LLM 被撑爆）
    _TOOL_RESULT_MAX_CHARS = 16_000
    # content 最大字符数（防撑爆存储/带宽）
    _CONTENT_MAX_CHARS = 100_000
    # 消息字段白名单（防止 LLM 塞入未知字段）
    _ALLOWED_MESSAGE_FIELDS = frozenset({
        "role", "content", "tool_calls", "tool_call_id", "name", "timestamp",
    })

    # ─── 消息消毒 ────────────────────────────

    @staticmethod
    def _sanitize_message(entry: dict) -> dict:
        """
        消毒：确保消息数据符合存储和 LLM API 的要求。

        处理项：
        1. content 不能为 None（LLM 将 null 渲染为 "None"）
        2. 超长 content 截断（防撑爆存储）
        3. 移除空字符 \x00（PG 拒绝写入）
        4. tool_calls 必须是 list[dict]，否则丢弃
        5. 移除 schema 未定义的字段
        """
        entry = dict(entry)

        # ── content 处理 ──
        content = entry.get("content")
        if content is None:
            entry["content"] = ""
        elif isinstance(content, str):
            content = content.replace("\x00", "")
            if len(content) > AgentLoop._CONTENT_MAX_CHARS:
                content = content[:AgentLoop._CONTENT_MAX_CHARS] + "\n... (截断)"
            entry["content"] = content
        else:
            entry["content"] = str(content)

        # ── tool_calls 校验 ──
        tc = entry.get("tool_calls")
        if tc is not None:
            if not (isinstance(tc, list) and all(isinstance(x, dict) for x in tc)):
                logger.warning(f"[Agent] 丢弃非法 tool_calls: type={type(tc).__name__}")
                entry.pop("tool_calls", None)

        # ── 字段白名单 ──
        for key in list(entry.keys()):
            if key not in AgentLoop._ALLOWED_MESSAGE_FIELDS:
                entry.pop(key, None)

        entry.setdefault("role", "user")
        return entry

    def __init__(
        self,
        bus: MessageBus,
        provider: LLMProvider,
        workspace: Path,
        model: str | None = None,
        max_iterations: int = 40,
        context_window_tokens: int = 65_536,
        skills_dir: str | Path | None = None,
        pg_config: dict | None = None,
        redis_cache=None,
    ):
        """初始化 Agent Loop。

        Args:
            bus: 消息总线（解耦通道和 Agent）
            provider: LLM Provider
            workspace: 工作区路径
            model: 模型名，默认使用 provider 的默认模型
            max_iterations: 最大 ReAct 循环轮数
            context_window_tokens: 上下文窗口大小（用于记忆压缩）
            skills_dir: 技能目录路径（可选，默认 workspace/skills）
            pg_config: PostgreSQL 配置（可选，启用后会话持久化到 PG）
            redis_cache: RedisCache 实例（可选，启用四层缓存）
        """
        self.bus = bus
        self.provider = provider
        self.workspace = workspace
        self.model = model or provider.get_default_model()
        self.max_iterations = max_iterations
        self.context_window_tokens = context_window_tokens
        self._start_time = time.time()

        # 核心组件
        self.context = ContextBuilder(workspace)
        self.sessions = SessionManager(workspace, pg_config=pg_config, redis_cache=redis_cache)
        self.tools = ToolRegistry()

        # Redis 缓存系统（可选，由 orchestrator 注入）
        self.redis_cache = redis_cache
        self.rate_limiter = None
        if redis_cache:
            from agent.cache.redis_cache import RedisRateLimiter
            self.rate_limiter = RedisRateLimiter(redis_cache)
            logger.info("[Agent] Redis 缓存 + 限流器已就绪")

        # 技能系统（对标 OpenClaw 的 skills/ 机制）
        # 扫描 skills/ 目录，自动发现 Skill → 注册底层 Tool
        skills_path = Path(skills_dir) if skills_dir else (workspace / "skills")
        self.skill_manager = SkillManager(skills_path)

        # 记忆 Consolidation（已有）
        self.memory_consolidator = MemoryConsolidator(
            workspace=workspace,
            provider=provider,
            model=self.model,
            memory_store=self.context.memory,
            context_builder=self.context,
            sessions=self.sessions,
        )

        # 新增：ChatMemory 门面和 Dream（对标 nanobot 的 Dream）
        self.chat_memory = ChatMemory(workspace=workspace)
        self.dream = self.chat_memory.dream

        # MQ Producer（由 Orchestrator 注入）
        # 用于将非关键路径任务（Dream/Consolidation）投递到 RabbitMQ
        self.mq_producer = None

        # 运行状态
        self._running = False
        self._active_tasks: dict[str, list[asyncio.Task]] = {}
        self._session_locks: dict[str, asyncio.Lock] = {}

        # 并发门控：NANOBOT_MAX_CONCURRENT_REQUESTS（<=0 不限, 默认 3）
        _max = int(os.environ.get("JDAGENT_MAX_CONCURRENT_REQUESTS", "3"))
        self._concurrency_gate: asyncio.Semaphore | None = (
            asyncio.Semaphore(_max) if _max > 0 else None
        )

        # 注册默认工具
        self._register_default_tools()

    async def initialize(self) -> None:
        """
        异步初始化 Agent 子系统。

        按顺序：
        1. 初始化 PostgreSQL 连接池（如果有配置）
        2. 发现并加载技能插件
        """
        await self.sessions.initialize_pg()
        await self.initialize_skills()

    async def initialize_skills(self) -> None:
        """
        初始化技能系统：扫描 skills/ 目录，发现并加载所有技能，
        然后将技能下的 Tool 注册到 ToolRegistry。
        """
        try:
            result = await self.skill_manager.discover()
            if result.success_count > 0:
                # 自动注册技能下的 Tool 到 ToolRegistry
                count = self.skill_manager.register_all_tools(self.tools)
                logger.info(
                    f"[Skill] 技能初始化完成: {result.success_count} 个技能, "
                    f"{count} 个工具注册"
                )
            if result.failed_count > 0:
                for f in result.failed:
                    logger.warning(f"[Skill] 技能加载失败: {f}")
        except Exception as e:
            logger.warning(f"[Skill] 技能系统初始化失败: {e}")

    def _register_default_tools(self) -> None:
        """注册 Agent 的默认工具集"""
        self.tools.register(ReadFileTool(workspace=self.workspace))
        self.tools.register(WriteFileTool(workspace=self.workspace))
        self.tools.register(WebFetchTool())

        # 注意：外部工具（如 get_weather、query_order 等）在 tool_definitions.py 中注册
        # 它们会被 orchestrator 加载后追加到这里

    # ─── 工具上下文设置 ─────────────────────────

    def _set_tool_context(self, channel: str, chat_id: str, message_id: str | None = None) -> None:
        """为需要路由信息的工具设置上下文"""
        pass  # 预留：在有 MessageTool 等时需要

    # ─── 核心：ReAct 循环 ─────────────────────

    def _inject_jwt(self, kwargs: dict, jwt_token: str) -> dict:
        """向工具参数中注入 JWT token（仅当工具需要时）"""
        if jwt_token:
            return {**kwargs, "_jwt_token": jwt_token}
        return kwargs

    async def _run_agent_loop(
        self,
        initial_messages: list[dict],
        session_key: str = "",
        on_progress: Callable[..., Any] | None = None,
        on_stream: Callable[[str], Any] | None = None,
        on_stream_end: Callable[..., Any] | None = None,
        jwt_token: str = "",
    ) -> tuple[str | None, list[str], list[dict]]:
        """
        运行 Agent ReAct 循环。

        流程：
        while 迭代次数 < max_iterations:
            1. 获取工具定义
            2. 调用 LLM（流式或非流式）
            3. 如果 LLM 返回工具调用:
                a. 执行工具（并行执行所有工具调用）
                b. 将结果注入消息列表
                c. 继续循环
            4. 如果 LLM 返回文本:
                a. 作为最终回复返回

        Args:
            initial_messages: 初始消息列表（system + history + user，已含所有上下文）
            on_progress: 进度回调（非流式模式下的文字进度）
            on_stream: 流式内容回调（每个文本块）
            on_stream_end: 流式结束回调（resuming=False = 最终回复）

        Returns:
            (final_content, tools_used, all_messages)
        """
        messages = initial_messages
        iteration = 0
        final_content = None
        tools_used: list[str] = []
        _cache_miss = False          # track iteration-1 cache miss for stampede protection
        _stampede_lock_name = None   # distributed lock name for stampede protection

        while iteration < self.max_iterations:
            iteration += 1

            # 获取工具定义（每次循环都重新获取，因为工具可能动态变更）
            tool_defs = self.tools.get_definitions()

            # ═══ ① LLM Response Cache（仅第一轮有效）═══
            # 命中缓存 = 跳过整个 ReAct 循环，直接返回历史结果。
            # 只检查第一轮的原因：后续轮次含工具结果，缓存命中率极低。
            if iteration == 1 and self.redis_cache:
                cached = await self.redis_cache.get_llm_cache(session_key, messages)
                if cached is None:
                    _cache_miss = True
                else:
                    _cache_miss = False
                if cached is not None:
                    logger.info("[Agent] LLM 缓存命中，跳过 ReAct 循环")
                    final_content = cached.get("content")
                    tc_list = cached.get("tool_calls", [])
                    for tc in tc_list:
                        tools_used.append(tc.get("name", "unknown"))
                    # 将缓存的 assistant 消息追加到消息列表
                    messages = self._add_assistant_message(messages, final_content, tc_list if tc_list else None)
                    # 如果缓存里有工具调用，重新执行工具并继续循环
                    if tc_list:
                        # 注意：缓存命中时工具结果是 NOT 缓存的（工具可能有副作用），
                        # 所以即使缓存命中，仍需重新执行工具。
                        # 只有纯文本回复可以完全跳过 LLM 调用。
                        results = await asyncio.gather(*(
                            self.tools.execute(tc["name"], self._inject_jwt(tc.get("arguments", {}), jwt_token))
                            for tc in tc_list
                        ), return_exceptions=True)
                        for tc, result in zip(tc_list, results):
                            if isinstance(result, BaseException):
                                result = f"错误: {type(result).__name__}: {result}"
                            messages = self._add_tool_result(messages, tc["id"], tc["name"], result)
                        continue  # 继续循环（让 LLM 基于工具结果生成最终回复）
                    else:
                        # 纯文本回复，直接返回
                        tool_defs = self.tools.get_definitions()
                        break

            # ═══ ② Rate Limiter：LLM 调用级别限流 ═══
            # 与入口的消息级别限流不同，这里限制的是 LLM API 调用频率。
            # 防止工具循环中过快调用 LLM（某些场景可能几秒内多次调 API）。
            # 使用全局 key（不区分 session），控制整体 API 费用。
            if self.rate_limiter:
                llm_allowed, _ = await self.rate_limiter.check_atomic(
                    key="global:llm",
                    limit=self.rate_limiter.cache.rate_limit_llm,
                    window=self.rate_limiter.cache.rate_limit_llm_window,
                )
                if not llm_allowed:
                    logger.warning("[RateLimiter] LLM 调用超限，等待")
                    await asyncio.sleep(1)
                    continue

            # ═══ ②.5 缓存击穿防护（分布式锁）═══
            # N 个请求同时缓存 miss 时，只有拿到锁的请求调 LLM API，
            # 其他进程等待 → 双检缓存（缓存击穿防护）。
            _stampede_lock_name = None
            if iteration == 1 and self.redis_cache and _cache_miss:
                _lock_cache_key = self.redis_cache._build_llm_cache_key(session_key, messages)
                _stampede_lock_name = f"stampede:{_lock_cache_key}"
                if await self.redis_cache.lock(_stampede_lock_name, ttl=10):
                    # 双检：等锁期间其他进程可能已填充缓存
                    cached = await self.redis_cache.get_llm_cache(session_key, messages)
                    if cached is not None:
                        await self.redis_cache.unlock(_stampede_lock_name)
                        _stampede_lock_name = None
                        logger.info("[Agent] 双检缓存命中（其他进程已填充），跳过 LLM 调用")
                        final_content = cached.get("content")
                        tc_list = cached.get("tool_calls", [])
                        for tc in tc_list:
                            tools_used.append(tc.get("name", "unknown"))
                        messages = self._add_assistant_message(
                            messages, final_content, tc_list if tc_list else None
                        )
                        if tc_list:
                            results = await asyncio.gather(*(
                                self.tools.execute(tc["name"], self._inject_jwt(tc.get("arguments", {}), jwt_token))
                                for tc in tc_list
                            ), return_exceptions=True)
                            for tc, result in zip(tc_list, results):
                                if isinstance(result, BaseException):
                                    result = f"错误: {type(result).__name__}: {result}"
                                messages = self._add_tool_result(messages, tc["id"], tc["name"], result)
                            continue
                        else:
                            break
                else:
                    logger.info("[Agent] 等待其他进程完成 LLM 调用（缓存击穿防护）")
                    await asyncio.sleep(0.5)
                    continue

            # === LLM 调用 ===
            if on_stream:
                # 流式模式
                response = await self.provider.chat_stream_with_retry(
                    messages=messages,
                    tools=tool_defs,
                    model=self.model,
                    on_content_delta=on_stream,
                )
            else:
                # 非流式模式
                response = await self.provider.chat_with_retry(
                    messages=messages,
                    tools=tool_defs,
                    model=self.model,
                )

            # 记录 token 用量
            usage = response.usage or {}

            # ═══ ③ LLM Response Cache 写入（第一轮）═══
            if iteration == 1 and self.redis_cache and not on_stream:
                # 缓存 LLM 响应，避免相同上下文重复调 API。
                # 缓存内容：content + tool_calls（不缓存工具结果——工具可能有副作用）
                cache_data = {
                    "content": response.content,
                    "tool_calls": [
                        {"id": tc.id, "name": tc.name, "arguments": tc.arguments}
                        for tc in (response.tool_calls or [])
                    ],
                    "finish_reason": response.finish_reason,
                }
                await self.redis_cache.set_llm_cache(session_key, messages, cache_data)

            if response.has_tool_calls:
                # ─── Act: 执行工具调用 ───
                if on_stream and on_stream_end:
                    # 通知前端：流式结束，开始执行工具（可以显示 spinner）
                    await on_stream_end(resuming=True)

                if on_progress:
                    # 发送进度通知
                    for tc in response.tool_calls:
                        await on_progress(f"[调用工具] {tc.name}({json.dumps(tc.arguments, ensure_ascii=False)[:80]}...)")

                # 将 assistant 消息（含 tool_calls）追加到消息列表
                tool_call_dicts = [
                    tc.to_openai_tool_call()
                    for tc in response.tool_calls
                ]
                messages = self._add_assistant_message(
                    messages, response.content, tool_call_dicts,
                    reasoning_content=response.reasoning_content,
                    thinking_blocks=response.thinking_blocks,
                )

                # 记录工具调用
                for tc in response.tool_calls:
                    tools_used.append(tc.name)
                    logger.info(f"[Agent] Tool call: {tc.name}({json.dumps(tc.arguments, ensure_ascii=False)[:200]})")

                # ─── Observe: 执行所有工具并收集结果 ───
                # 注意：LLM 在一次响应中发出的多个工具调用是相互独立的，
                # 所以可以并发执行。return_exceptions=True 确保所有结果都被收集。
                results = await asyncio.gather(*(
                    self.tools.execute(tc.name, self._inject_jwt(tc.arguments, jwt_token))
                    for tc in response.tool_calls
                ), return_exceptions=True)

                # 将工具结果追加到消息列表
                for tool_call, result in zip(response.tool_calls, results):
                    if isinstance(result, BaseException):
                        result = f"错误: {type(result).__name__}: {result}"
                    messages = self._add_tool_result(
                        messages, tool_call.id, tool_call.name, result
                    )

                # 释放分布式锁（工具执行路径走到这里，解锁后继续下一轮迭代）
                if _stampede_lock_name:
                    await self.redis_cache.unlock(_stampede_lock_name)
                    _stampede_lock_name = None

            else:
                # ─── LLM 给出最终回复 ───
                if on_stream and on_stream_end:
                    await on_stream_end(resuming=False)

                # 检查是否出错
                if response.finish_reason == "error":
                    logger.error(f"[Agent] LLM 返回错误: {(response.content or '')[:200]}")
                    if _stampede_lock_name:
                        await self.redis_cache.unlock(_stampede_lock_name)
                        _stampede_lock_name = None
                    final_content = response.content or "抱歉，调用 AI 模型时出错。"
                    break

                # 转换为最终回复
                messages = self._add_assistant_message(
                    messages, response.content,
                    reasoning_content=response.reasoning_content,
                    thinking_blocks=response.thinking_blocks,
                )
                if _stampede_lock_name:
                    await self.redis_cache.unlock(_stampede_lock_name)
                    _stampede_lock_name = None
                final_content = response.content
                break

        # 超过最大迭代次数
        if final_content is None and iteration >= self.max_iterations:
            logger.warning(f"[Agent] 达到最大迭代次数 {self.max_iterations}")
            final_content = (
                f"我已经尝试了 {self.max_iterations} 次工具调用仍未完成任务。"
                "您可以尝试将任务分解成更多步骤。"
            )

        return final_content, tools_used, messages

    # ─── 消息处理工具方法 ────────────────────

    @staticmethod
    def _add_assistant_message(
        messages: list[dict],
        content: str | None,
        tool_call_dicts: list[dict] | None = None,
        reasoning_content: str | None = None,
        thinking_blocks: list[dict] | None = None,
    ) -> list[dict]:
        """添加 assistant 消息到消息列表"""
        # 确保 content 不会是 None → 某些 LLM API 将 null content 渲染为 "None"
        msg = {"role": "assistant", "content": content or ""}
        if tool_call_dicts:
            msg["tool_calls"] = tool_call_dicts
        if reasoning_content:
            msg["reasoning_content"] = reasoning_content
        if thinking_blocks:
            msg["thinking_blocks"] = thinking_blocks
        messages.append(msg)
        return messages

    @staticmethod
    def _add_tool_result(
        messages: list[dict],
        tool_call_id: str,
        tool_name: str,
        result: Any,
    ) -> list[dict]:
        """添加工具执行结果到消息列表"""
        content = str(result)
        # 截断过长的结果
        if len(content) > 16000:
            content = content[:16000] + "\n... (结果过长已截断)"
        messages.append({
            "role": "tool",
            "tool_call_id": tool_call_id,
            "name": tool_name,
            "content": content,
        })
        return messages

    # ─── 主循环 ─────────────────────────────

    async def run(self) -> None:
        """
        运行 Agent 主循环。
        
        从 MessageBus 消费入站消息，为每个 session 启动独立任务。
        同 session 串行，跨 session 并发。
        """
        self._running = True
        logger.info("[Agent] Agent 循环启动")

        while self._running:
            try:
                msg = await asyncio.wait_for(self.bus.consume_inbound(), timeout=1.0)
            except asyncio.TimeoutError:
                continue
            except asyncio.CancelledError:
                raise
            except Exception as e:
                logger.warning(f"[Agent] 消费消息出错: {e}")
                continue

            # 为每个消息创建独立任务
            task = asyncio.create_task(self._dispatch(msg))
            self._active_tasks.setdefault(msg.session_key, []).append(task)
            task.add_done_callback(
                lambda t, k=msg.session_key: self._active_tasks.get(k, []) and
                self._active_tasks[k].remove(t) if t in self._active_tasks.get(k, []) else None
            )

    async def _dispatch(self, msg: InboundMessage, on_stream: Callable[[str], Any] | None = None) -> OutboundMessage | None:
        """
        分发消息到对应 session。

        使用 per-session lock 保证同 session 串行处理。
        使用 concurrency gate 控制全局并发数。

        流式模式（on_stream 不为 None）：
        - 不发布到总线（由调用者直接消费）
        - 返回响应给调用者链式处理
        """
        lock = self._session_locks.setdefault(msg.session_key, asyncio.Lock())
        gate = self._concurrency_gate or nullcontext()
        async with lock, gate:
            try:
                response = await self._process_message(msg, on_stream=on_stream)
                if response is not None and on_stream is None:
                    await self.bus.publish_outbound(response)
                return response

            except asyncio.CancelledError:
                logger.info(f"[Agent] 会话 {msg.session_key} 的任务被取消")
                raise
            except Exception:
                logger.exception(f"[Agent] 处理会话 {msg.session_key} 出错")
                error_msg = OutboundMessage(
                    channel=msg.channel,
                    chat_id=msg.chat_id,
                    content="抱歉，处理您的消息时出现错误。",
                )
                if on_stream is None:
                    await self.bus.publish_outbound(error_msg)
                return error_msg

    async def _process_message(
        self,
        msg: InboundMessage,
        on_stream: Callable[[str], Any] | None = None,
    ) -> OutboundMessage | None:
        """
        处理单条入站消息。

        完整流程：
        1. 获取或创建 session
        2. 加载历史消息（最近 30 条未 consolidated 的消息）
        3. 构建消息列表（ContextBuilder 统管身份 + 长期记忆 + 中期记忆）
        4. 运行 ReAct 循环
        5. 持久化新消息
        6. 触发记忆 consolidation
        7. 截断已 consolidated 的消息（内存有界）
        8. 返回出站消息
        """
        preview = msg.content[:80] + "..." if len(msg.content) > 80 else msg.content
        logger.info(f"[Agent] 处理来自 {msg.channel}:{msg.sender_id} 的消息: {preview}")

        # ── Step 0: Rate Limiter（限流检查）──
        # 在真正处理消息之前先检查是否超限。
        # 这样做的好处：限流时连 session 都不需要加载，最大程度节约资源。
        if self.rate_limiter:
            allowed, remaining = await self.rate_limiter.check(msg.session_key)
            if not allowed:
                logger.warning(
                    f"[RateLimiter] 会话 {msg.session_key} 请求超限，拒绝"
                )
                return OutboundMessage(
                    channel=msg.channel,
                    chat_id=msg.chat_id,
                    content="请求过于频繁，请稍后再试。"
                    f"（每 {self.rate_limiter.cache.rate_limit_window} 秒最多 "
                    f"{self.rate_limiter.cache.rate_limit_default} 次）",
                )

        # 获取或创建 session（异步：优先从 PG 加载）
        session = await self.sessions.aget_or_create(msg.session_key)

        # ── 技能匹配：根据用户意图匹配最相关的技能 ──
        skill_context = await self._match_skills(msg.content)

        # 构建初始消息（ContextBuilder 统管身份 + 长期记忆 + 中期记忆 + 技能上下文）
        history = session.get_history(max_messages=30)
        initial_messages = self.context.build_messages(
            history=history,
            current_message=msg.content,
            media=msg.media if msg.media else None,
            channel=msg.channel,
            chat_id=msg.chat_id,
            chat_memory=self.chat_memory,  # 注入中期记忆摘要
            skill_context=skill_context,   # 注入技能上下文
        )

        # 运行 ReAct 循环
        jwt_token = msg.metadata.get("jwt_token", "") if msg.metadata else ""
        final_content, tools_used, all_msgs = await self._run_agent_loop(
            initial_messages,
            session_key=msg.session_key,
            on_stream=on_stream,
            jwt_token=jwt_token,
        )

        if final_content is None:
            final_content = "处理完成，但没有生成回复。"

        # 持久化消息（异步三写：PG + Redis + JSONL）
        self._save_turn(session, all_msgs, 1 + len(history))
        await self.sessions.asave(session)

        # 使 LLM 缓存失效：新消息意味着对话上下文已变，旧缓存不再适用
        if self.redis_cache:
            await self.redis_cache.invalidate_llm_cache(msg.session_key)

        # ── 后台任务（异步执行，不阻塞用户响应）──
        if self.mq_producer:
            # MQ 模式：通过 RabbitMQ 异步投递
            # 消费者 Worker 在后台处理 Dream 归档 + 记忆 Consolidation
            await asyncio.gather(
                self.mq_producer.publish("dream", {}),
                self.mq_producer.publish("consolidation", {"session_key": msg.session_key}),
                return_exceptions=True,
            )
        else:
            # 降级模式：没有 MQ 时原地同步执行（原逻辑）
            # 触发记忆 consolidation
            await self.memory_consolidator.maybe_consolidate(session)

            # 截断已 consolidated 的消息（仅清内存，不写存储）
            # 数据已在 consolidation 前的 asave() 中持久化到 PG + JSONL + history.jsonl，
            # 不再调用 asave() 避免用空 messages 覆盖 JSONL 文件。
            if session.last_consolidated > 0:
                session.messages[:session.last_consolidated] = []
                session.last_consolidated = 0

            # 触发 Dream 归档（积攒至少 3 条未处理的 consolidation 时执行）
            try:
                if self.dream.pending_count() >= 3:
                    dream_result = self.chat_memory.run_dream()
                    if dream_result.get('added', 0) > 0 or dream_result.get('replaced', 0) > 0:
                        logger.info(f"[Dream] 归档完成: {dream_result}")
            except Exception as exc:
                logger.warning(f"[Dream] 归档失败: {exc}")

        preview_rsp = final_content[:120] + "..." if len(final_content) > 120 else final_content
        logger.info(f"[Agent] 回复 {msg.channel}:{msg.sender_id}: {preview_rsp}")

        return OutboundMessage(
            channel=msg.channel,
            chat_id=msg.chat_id,
            content=final_content,
        )

    # ─── 技能匹配 ────────────────────────────────

    async def _match_skills(self, query: str) -> str | None:
        """
        根据用户查询匹配合适的技能，返回技能上下文文本。
        
        流程：
          1. 调用 SkillManager.match() 获取 top-k 技能
          2. 调用 compute_context() 生成 LLM 能理解的上下文文本
          3. 将上下文注入到 system prompt 中
        
        如果没有匹配到任何技能，返回 None。
        """
        if self.skill_manager.count == 0:
            return None

        try:
            matches = await self.skill_manager.match(query, top_k=3)
            if not matches:
                return None

            context = self.skill_manager.compute_context(matches)
            if context:
                logger.info(
                    f"[技能匹配] 命中 {len(matches)} 个技能: "
                    + ", ".join(f"{m.skill.name}({m.score:.2f})" for m in matches)
                )
            return context
        except Exception as e:
            logger.warning(f"[技能匹配] 失败: {e}")
            return None

    # ─── 消息持久化 ────────────────────────────

    def _save_turn(self, session: Session, messages: list[dict], skip: int) -> None:
        """
        将本轮消息保存到 session。

        skip: 要跳过的消息数（即历史消息的数量）
        只保存本轮新生成的消息。
        """
        for m in messages[skip:]:
            entry = dict(m)
            role = entry.get("role")
            content = entry.get("content")

            # 跳过空的 assistant 消息（可能由 tool-call-only 消息产生）
            if role == "assistant" and not content and not entry.get("tool_calls"):
                continue

            # 截断过长的 tool result
            if role == "tool":
                if isinstance(content, str) and len(content) > self._TOOL_RESULT_MAX_CHARS:
                    entry["content"] = content[:self._TOOL_RESULT_MAX_CHARS] + "\n... (截断)"

            # 消毒：content 非空/截断/空字符、tool_calls 校验、字段白名单
            entry = self._sanitize_message(entry)

            entry.setdefault("timestamp", time.time())
            session.messages.append(entry)

        session.updated_at = __import__("datetime").datetime.now()

    # ─── 直接处理（CLI/非总线模式）────────────

    async def process_direct(
        self,
        content: str,
        session_key: str = "cli:direct",
        channel: str = "cli",
        chat_id: str = "direct",
        metadata: dict | None = None,
    ) -> OutboundMessage | None:
        """
        直接处理消息（不走 MessageBus，适用于 CLI 模式）。

        注意：session_key 的格式为 "channel:chat_id"，
        例如 "cli:direct" → channel=cli, chat_id=direct。
        如果传入的 session_key 包含冒号，会自动分解覆盖 channel/chat_id。
        """
        parts = session_key.split(":", 1)
        if len(parts) == 2:
            channel, chat_id = parts[0], parts[1]

        msg = InboundMessage(
            channel=channel,
            sender_id="user",
            chat_id=chat_id,
            content=content,
            metadata=metadata or {},
        )
        return await self._process_message(msg)

    # ─── 生命周期管理 ────────────────────────

    def stop(self) -> None:
        """停止 Agent 循环"""
        self._running = False
        logger.info("[Agent] Agent 循环停止")
