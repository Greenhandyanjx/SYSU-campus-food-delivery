"""
Provider: 通义千问 LLM Provider
参考 nanobot 的 providers/base.py 设计

实现了 LLMProvider 抽象基类，封装通义千问 API：
1. 标准的 OpenAI-compatible 调用方式
2. 推理模式支持（reasoning_effort）
3. 流式支持
4. 错误码判断与重试
"""

import json
from typing import Any, Callable

from loguru import logger

from agent.core.types import LLMResponse, ToolCallRequest
from agent.core.llm_provider import LLMProvider


class TongyiProvider(LLMProvider):
    """
    通义千问 Provider (DashScope SDK)
    
    使用 langchain_community.chat_models.tongyi.ChatTongyi 作为底层调用，
    但会构建兼容 OpenAI 格式的请求以确保统一的响应解析。
    """

    def __init__(
        self,
        api_key: str | None = None,
        api_base: str | None = None,
        model: str = "qwen-plus",
    ):
        super().__init__(api_key=api_key, api_base=api_base)
        self._model = model

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
        """调用通义千问模型"""
        model_name = model or self._model

        try:
            # 使用 ChatTongyi（同步调用）
            from langchain_community.chat_models.tongyi import ChatTongyi

            llm = ChatTongyi(
                model=model_name,
                api_key=self.api_key,
                temperature=temperature,
                max_tokens=max_tokens,
            )

            # 处理 tool_choice
            invoke_kwargs = {}
            if tools:
                invoke_kwargs["tools"] = tools
            if tool_choice:
                invoke_kwargs["tool_choice"] = tool_choice

            # 通义千问的 ChatTongyi 调用
            response = await llm.ainvoke(messages, **invoke_kwargs)

            return self._parse_response(response)

        except Exception as e:
            error_str = str(e)
            logger.error(f"[TongyiProvider] 调用失败: {error_str[:200]}")

            # 判断是否是瞬态错误
            if self._is_transient_error(error_str):
                return LLMResponse(content=error_str, finish_reason="error")

            # 非瞬态错误（参数错误、鉴权失败等），不重试
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
        """
        流式调用通义千问模型。
        
        注意：通义千问的 stream 模式只返回文本内容，不返回 tool_calls。
        所以如果同时有 tools，需要先非流式拿到 tool_calls，然后流式文本。
        """
        model_name = model or self._model

        try:
            from langchain_community.chat_models.tongyi import ChatTongyi

            llm = ChatTongyi(
                model=model_name,
                api_key=self.api_key,
                temperature=temperature,
                max_tokens=max_tokens,
                streaming=True,
            )

            invoke_kwargs = {}
            if tools:
                invoke_kwargs["tools"] = tools
            if tool_choice:
                invoke_kwargs["tool_choice"] = tool_choice

            # 流式响应
            collected_content = ""
            tool_call_acc = None

            async for chunk in llm.astream(messages, **invoke_kwargs):
                delta = chunk.content or ""
                if delta:
                    collected_content += delta
                    if on_content_delta:
                        await on_content_delta(delta)

                # 检查是否有 tool_calls
                if hasattr(chunk, "tool_call_chunks") and chunk.tool_call_chunks:
                    if tool_call_acc is None:
                        tool_call_acc = {"id": "", "function": {"name": "", "arguments": ""}}
                    for tcc in chunk.tool_call_chunks:
                        if tcc.id:
                            tool_call_acc["id"] += tcc.id
                        if tcc.name:
                            tool_call_acc["function"]["name"] += tcc.name
                        if tcc.args:
                            tool_call_acc["function"]["arguments"] += tcc.args

            # 解析响应
            if tool_call_acc and tool_call_acc.get("function", {}).get("name"):
                return self._parse_response_with_tool_call(collected_content, tool_call_acc)
            else:
                return LLMResponse(content=collected_content or None)

        except Exception as e:
            error_str = str(e)
            logger.error(f"[TongyiProvider] 流式调用失败: {error_str[:200]}")
            return LLMResponse(
                content=error_str if self._is_transient_error(error_str) else f"模型调用失败: {error_str}",
                finish_reason="error",
            )

    # ─── 响应解析 ─────────────────────────────

    def _parse_response(self, response) -> LLMResponse:
        """解析 ChatTongyi 的同步响应"""
        content = getattr(response, "content", None) or ""
        additional_kwargs = getattr(response, "additional_kwargs", {}) or {}

        # 检查 tool_calls
        tool_calls_raw = None
        if hasattr(response, "tool_calls") and response.tool_calls:
            tool_calls_raw = response.tool_calls
        elif "tool_calls" in additional_kwargs:
            tool_calls_raw = additional_kwargs["tool_calls"]

        # token 用量
        usage = {}
        if hasattr(response, "usage_metadata") and response.usage_metadata:
            usage = {
                "prompt_tokens": response.usage_metadata.get("input_tokens", 0),
                "completion_tokens": response.usage_metadata.get("output_tokens", 0),
                "total_tokens": response.usage_metadata.get("total_tokens", 0),
            }

        if tool_calls_raw:
            tool_calls = self._parse_tool_calls(tool_calls_raw)
            return LLMResponse(
                content=content or None,
                tool_calls=tool_calls,
                finish_reason="tool_calls",
                usage=usage,
            )

        return LLMResponse(content=content or None, usage=usage)

    def _parse_response_with_tool_call(
        self, content: str, tool_call_acc: dict
    ) -> LLMResponse:
        """从流式累积的 tool_call 数据构建响应"""
        try:
            args = json.loads(tool_call_acc["function"]["arguments"])
        except (json.JSONDecodeError, KeyError):
            args = {}

        tc = ToolCallRequest(
            id=tool_call_acc.get("id", ""),
            name=tool_call_acc["function"].get("name", ""),
            arguments=args,
        )
        return LLMResponse(
            content=content or None,
            tool_calls=[tc],
            finish_reason="tool_calls",
        )

    @staticmethod
    def _parse_tool_calls(tool_calls_raw: list) -> list[ToolCallRequest]:
        """将原始 tool_calls 解析为 ToolCallRequest 列表"""
        result = []
        for tc in tool_calls_raw:
            # LangChain 的 AIMessage.tool_calls 格式
            if isinstance(tc, dict):
                tid = tc.get("id", "")
                tname = tc.get("name", tc.get("function", {}).get("name", ""))
                targs = tc.get("args", tc.get("function", {}).get("arguments", {}))
                if isinstance(targs, str):
                    try:
                        targs = json.loads(targs)
                    except json.JSONDecodeError:
                        targs = {}
                result.append(ToolCallRequest(id=tid, name=tname, arguments=targs))
            else:
                # 假设是 LangChain ToolCall 对象
                result.append(ToolCallRequest(
                    id=getattr(tc, "id", ""),
                    name=getattr(tc, "name", ""),
                    arguments=getattr(tc, "args", {}),
                ))
        return result
