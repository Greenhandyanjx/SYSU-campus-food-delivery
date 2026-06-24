"""
Provider: OpenAI Provider
使用 openai Python SDK 实现 LLMProvider 抽象基类

支持：
- GPT-4o / GPT-4o-mini / o1 / o3-mini
- 推理模式（reasoning_effort）
- 流式输出
- Tool Calling
"""

import json
from typing import Any, Callable

from openai import AsyncOpenAI
from loguru import logger

from agent.core.types import LLMResponse, ToolCallRequest
from agent.core.llm_provider import LLMProvider


class OpenAIProvider(LLMProvider):
    """
    OpenAI API 兼容的 LLM Provider。
    
    适用于：
    - 官方 OpenAI (api_base="https://api.openai.com/v1")
    - 通义千问兼容模式 (api_base="https://dashscope.aliyuncs.com/compatible-mode/v1")
    - DeepSeek (api_base="https://api.deepseek.com")
    - 任何 OpenAI-compatible 的 API
    """

    def __init__(
        self,
        api_key: str | None = None,
        api_base: str | None = None,
        model: str = "gpt-4o-mini",
    ):
        super().__init__(api_key=api_key, api_base=api_base)
        self._model = model
        self._client = AsyncOpenAI(api_key=api_key or "", base_url=api_base)

    def get_default_model(self) -> str:
        return self._model

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
        """调用 OpenAI API"""
        model_name = model or self._model

        # 构建请求参数
        kwargs = dict(
            model=model_name,
            messages=messages,
            max_tokens=max_tokens,
            temperature=temperature,
        )
        if tools:
            kwargs["tools"] = tools
        if tool_choice:
            kwargs["tool_choice"] = tool_choice
        # 推理模型（o1/o3-mini 等）不支持 temperature
        if reasoning_effort:
            kwargs["reasoning_effort"] = reasoning_effort
            kwargs.pop("temperature", None)

        try:
            response = await self._client.chat.completions.create(**kwargs)
            return self._parse_response(response)

        except Exception as e:
            error_str = str(e)
            logger.error(f"[OpenAIProvider] 调用失败: {error_str[:200]}")

            if self._is_transient_error(error_str):
                return LLMResponse(content=error_str, finish_reason="error")

            return LLMResponse(
                content=f"模型调用失败: {error_str}",
                finish_reason="error",
            )

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
        """流式调用 OpenAI API"""
        model_name = model or self._model

        kwargs = dict(
            model=model_name,
            messages=messages,
            max_tokens=max_tokens,
            temperature=temperature,
            stream=True,
            stream_options={"include_usage": True},
        )
        if tools:
            kwargs["tools"] = tools
        if tool_choice:
            kwargs["tool_choice"] = tool_choice
        if reasoning_effort:
            kwargs["reasoning_effort"] = reasoning_effort
            kwargs.pop("temperature", None)

        try:
            collected_content = ""
            tool_call_acc: dict[str, Any] = {}
            usage = {}

            stream = await self._client.chat.completions.create(**kwargs)
            async for chunk in stream:
                delta = chunk.choices[0].delta if chunk.choices else None
                if delta is None:
                    # 检查 usage
                    if chunk.usage:
                        usage = {
                            "prompt_tokens": chunk.usage.prompt_tokens or 0,
                            "completion_tokens": chunk.usage.completion_tokens or 0,
                            "total_tokens": chunk.usage.total_tokens or 0,
                        }
                    continue

                # 文本内容
                if delta.content:
                    collected_content += delta.content
                    if on_content_delta:
                        await on_content_delta(delta.content)

                # Tool calls
                if delta.tool_calls:
                    for tc in delta.tool_calls:
                        idx = tc.index
                        if idx not in tool_call_acc:
                            tool_call_acc[idx] = {
                                "id": tc.id or "",
                                "function": {"name": "", "arguments": ""},
                            }
                        if tc.id:
                            tool_call_acc[idx]["id"] = tc.id
                        if tc.function:
                            if tc.function.name:
                                tool_call_acc[idx]["function"]["name"] += tc.function.name
                            if tc.function.arguments:
                                tool_call_acc[idx]["function"]["arguments"] += tc.function.arguments

            # 解析响应
            if tool_call_acc:
                tool_calls = []
                for idx in sorted(tool_call_acc.keys()):
                    tcd = tool_call_acc[idx]
                    try:
                        args = json.loads(tcd["function"]["arguments"])
                    except (json.JSONDecodeError, KeyError, TypeError):
                        args = {}
                    tool_calls.append(ToolCallRequest(
                        id=tcd.get("id", ""),
                        name=tcd["function"].get("name", ""),
                        arguments=args,
                    ))
                return LLMResponse(
                    content=collected_content or None,
                    tool_calls=tool_calls,
                    finish_reason="tool_calls",
                    usage=usage,
                )

            return LLMResponse(
                content=collected_content or None,
                usage=usage,
            )

        except Exception as e:
            error_str = str(e)
            logger.error(f"[OpenAIProvider] 流式调用失败: {error_str[:200]}")
            return LLMResponse(
                content=error_str if self._is_transient_error(error_str) else f"模型调用失败: {error_str}",
                finish_reason="error",
            )

    # ─── 响应解析 ─────────────────────────────

    @staticmethod
    def _parse_response(response) -> LLMResponse:
        """解析 OpenAI SDK 响应"""
        choice = response.choices[0] if response.choices else None
        if not choice:
            return LLMResponse(content=None, finish_reason="error")

        message = choice.message
        content = message.content or ""

        # Token 用量
        usage = {}
        if response.usage:
            usage = {
                "prompt_tokens": response.usage.prompt_tokens or 0,
                "completion_tokens": response.usage.completion_tokens or 0,
                "total_tokens": response.usage.total_tokens or 0,
            }

        # 解析 tool_calls
        if message.tool_calls:
            tool_calls = [
                ToolCallRequest(
                    id=tc.id,
                    name=tc.function.name,
                    arguments=json.loads(tc.function.arguments) if tc.function.arguments else {},
                )
                for tc in message.tool_calls
            ]
            return LLMResponse(
                content=content or None,
                tool_calls=tool_calls,
                finish_reason="tool_calls",
                usage=usage,
            )

        return LLMResponse(content=content or None, usage=usage)
