"""
Agent Core: LLM Provider 抽象接口
参考 nanobot 的 providers/base.py 设计

定义了统一的 LLM Provider 接口，所有具体的模型提供商（OpenAI、通义千问等）
都要实现这个接口。关键点：
1. 统一的 chat() / chat_stream() 接口
2. 内置重试/退避机制（处理限流、超时等瞬态错误）
3. 对每个 provider 的差异进行封装
"""

import asyncio
import json
import re
from abc import ABC, abstractmethod
from datetime import datetime, timezone
from typing import Any, Callable

from loguru import logger

from agent.core.types import LLMResponse, ToolCallRequest, GenerationSettings


class LLMProvider(ABC):
    """
    抽象 LLM Provider 基类。
    
    所有具体的模型提供商（OpenAI、通义千问、DeepSeek、Anthropic 等）
    都需要继承此类并实现 chat() 和 get_default_model()。
    
    关键设计（参考 nanobot）：
    - chat_with_retry() / chat_stream_with_retry() 自动处理重试
    - transient error 自动识别（429限流、500超时等）
    - 指数退避 + Retry-After 头部解析
    """

    # 重试延迟（秒）：第1次等1s，第2次等2s，第3次及以后等4s
    _CHAT_RETRY_DELAYS = (1, 2, 4)
    
    # 瞬态错误的关键词标记
    _TRANSIENT_ERROR_MARKERS = (
        "429", "rate limit", "500", "502", "503", "504",
        "overloaded", "timeout", "timed out", "connection",
        "server error", "temporarily unavailable", "速率限制",
    )

    # 构造函数：接受 API key 和 API base URL
    def __init__(self, api_key: str | None = None, api_base: str | None = None):
        self.api_key = api_key
        self.api_base = api_base
        self.generation = GenerationSettings()

    # ─── 抽象方法：子类必须实现 ───────────────────

    @abstractmethod
    async def chat(
        self,
        messages: list[dict[str, Any]],
        tools: list[dict[str, Any]] | None = None,
        model: str | None = None,
        max_tokens: int = 4096,
        temperature: float = 0.7,
        reasoning_effort: str | None = None,
        tool_choice: str | dict[str, Any] | None = None,
    ) -> LLMResponse:
        """
        发送聊天补全请求（非流式）。
        
        参数：
            messages: 消息列表，每条有 'role' 和 'content'
            tools: 可选的工具定义列表
            model: 模型标识
            max_tokens: 最大输出 token 数
            temperature: 采样温度
            reasoning_effort: 推理模式（low/medium/high）
            tool_choice: 工具选择策略（"auto"/"required"/具体工具）
        
        返回：LLMResponse（含 content 和/或 tool_calls）
        """
        pass

    @abstractmethod
    def get_default_model(self) -> str:
        """获取此 Provider 的默认模型名"""
        pass

    # ─── 流式调用（默认回退为非流式）──────────────

    async def chat_stream(
        self,
        messages: list[dict[str, Any]],
        tools: list[dict[str, Any]] | None = None,
        model: str | None = None,
        max_tokens: int = 4096,
        temperature: float = 0.7,
        reasoning_effort: str | None = None,
        tool_choice: str | dict[str, Any] | None = None,
        on_content_delta: Callable[[str], None] | None = None,
    ) -> LLMResponse:
        """
        流式聊天补全。默认回退为非流式方式（直接返回完整内容）。
        
        on_content_delta: 每个内容片段到达时的回调
        """
        # 默认实现：非流式调用 + 一次性推送完整内容
        response = await self.chat(
            messages=messages, tools=tools, model=model,
            max_tokens=max_tokens, temperature=temperature,
            reasoning_effort=reasoning_effort, tool_choice=tool_choice,
        )
        if on_content_delta and response.content:
            await on_content_delta(response.content)
        return response

    # ─── 带重试的安全调用 ──────────────────────

    async def chat_with_retry(
        self,
        messages: list[dict[str, Any]],
        tools: list[dict[str, Any]] | None = None,
        model: str | None = None,
        max_tokens: int | None = None,
        temperature: float | None = None,
        reasoning_effort: str | None = None,
        tool_choice: str | dict[str, Any] | None = None,
    ) -> LLMResponse:
        """
        带重试的聊天调用。
        
        自动处理瞬态错误（限流、超时等），使用指数退避策略重试。
        """
        # 填充默认值
        if max_tokens is None:
            max_tokens = self.generation.max_tokens
        if temperature is None:
            temperature = self.generation.temperature
        if reasoning_effort is None:
            reasoning_effort = self.generation.reasoning_effort

        kw = dict(
            messages=messages, tools=tools, model=model,
            max_tokens=max_tokens, temperature=temperature,
            reasoning_effort=reasoning_effort, tool_choice=tool_choice,
        )
        return await self._run_with_retry(self._safe_chat, kw)

    async def chat_stream_with_retry(
        self,
        messages: list[dict[str, Any]],
        tools: list[dict[str, Any]] | None = None,
        model: str | None = None,
        max_tokens: int | None = None,
        temperature: float | None = None,
        reasoning_effort: str | None = None,
        tool_choice: str | dict[str, Any] | None = None,
        on_content_delta: Callable[[str], None] | None = None,
    ) -> LLMResponse:
        """带重试的流式聊天调用。"""
        if max_tokens is None:
            max_tokens = self.generation.max_tokens
        if temperature is None:
            temperature = self.generation.temperature
        if reasoning_effort is None:
            reasoning_effort = self.generation.reasoning_effort

        kw = dict(
            messages=messages, tools=tools, model=model,
            max_tokens=max_tokens, temperature=temperature,
            reasoning_effort=reasoning_effort, tool_choice=tool_choice,
            on_content_delta=on_content_delta,
        )
        return await self._run_with_retry(self._safe_chat_stream, kw)

    # ─── 内部：安全的 chat 封装 ─────────────────

    async def _safe_chat(self, **kwargs: Any) -> LLMResponse:
        """调用 chat() 并将异常转换为错误响应"""
        try:
            return await self.chat(**kwargs)
        except asyncio.CancelledError:
            raise
        except Exception as exc:
            return LLMResponse(content=f"调用 LLM 出错: {exc}", finish_reason="error")

    async def _safe_chat_stream(self, **kwargs: Any) -> LLMResponse:
        """调用 chat_stream() 并将异常转换为错误响应"""
        try:
            return await self.chat_stream(**kwargs)
        except asyncio.CancelledError:
            raise
        except Exception as exc:
            return LLMResponse(content=f"调用 LLM 流式接口出错: {exc}", finish_reason="error")

    # ─── 内部：重试机制 ─────────────────────────

    @classmethod
    def _is_transient_error(cls, content: str | None) -> bool:
        """判断错误是否是瞬态的（可以重试的）"""
        err = (content or "").lower()
        return any(marker in err for marker in cls._TRANSIENT_ERROR_MARKERS)

    async def _run_with_retry(
        self,
        call: Callable[..., Any],
        kw: dict[str, Any],
    ) -> LLMResponse:
        """
        带指数退避重试的执行函数。
        
        策略：
        1. 如果返回 finish_reason != "error"，直接返回
        2. 如果是瞬态错误（限流/超时），等待后重试
        3. 最多重试 3 次（配置在 _CHAT_RETRY_DELAYS 中）
        4. 每次都尝试解析 Retry-After 头部
        """
        attempt = 0
        delays = list(self._CHAT_RETRY_DELAYS)
        last_response: LLMResponse | None = None

        while True:
            attempt += 1
            response = await call(**kw)

            # 成功：直接返回
            if response.finish_reason != "error":
                return response

            last_response = response

            # 非瞬态错误：不重试
            if not self._is_transient_response(response):
                return response

            # 超过最大重试次数
            if attempt > len(delays):
                logger.warning(
                    f"LLM 请求失败，已重试 {attempt} 次，放弃: "
                    f"{(response.content or '')[:120]}"
                )
                break

            # 计算等待时间：优先使用 Retry-After，否则指数退避
            base_delay = delays[min(attempt - 1, len(delays) - 1)]
            delay = self._extract_retry_after(response) or base_delay

            logger.warning(
                f"LLM 瞬态错误（第 {attempt}/{len(delays)} 次重试），"
                f"等待 {int(round(delay))}s: {(response.content or '')[:120]}"
            )
            await asyncio.sleep(delay)

        return last_response if last_response is not None else await call(**kw)

    # ─── 内部：Retry-After 解析 ──────────────────

    @classmethod
    def _is_transient_response(cls, response: LLMResponse) -> bool:
        """判断响应表示的是瞬态错误"""
        return cls._is_transient_error(response.content)

    @classmethod
    def _extract_retry_after(cls, response: LLMResponse) -> float | None:
        """从错误响应中提取 Retry-After 时间"""
        text = (response.content or "").lower()
        patterns = (
            r"retry after\s+(\d+(?:\.\d+)?)\s*(?:s|sec|seconds)?",
            r"try again in\s+(\d+(?:\.\d+)?)\s*(?:s|sec|seconds)?",
            r"wait\s+(\d+(?:\.\d+)?)\s*(?:s|sec|seconds)?",
            r"retry_after[\"'\s:=]+(\d+(?:\.\d+)?)",
        )
        for pattern in patterns:
            match = re.search(pattern, text)
            if match:
                return max(0.1, float(match.group(1)))
        return None
