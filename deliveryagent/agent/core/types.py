"""
Agent Core: 核心类型定义
参考 nanobot 的 LLMResponse / ToolCallRequest / InboundMessage / OutboundMessage 设计

本模块定义了整个 Agent 系统共享的数据类型，避免循环引用。
"""

from dataclasses import dataclass, field
from datetime import datetime
from typing import Any, TYPE_CHECKING

if TYPE_CHECKING:
    from agent.skills.base import Skill


# ─── LLM 响应类型 ───────────────────────────────────────


@dataclass
class ToolCallRequest:
    """
    LLM 返回的工具调用请求。
    对应 nanobot 的 ToolCallRequest，用于表示模型想要调用某个工具。
    """
    id: str                                    # 工具调用唯一ID
    name: str                                  # 工具名称
    arguments: dict[str, Any]                  # 工具参数
    extra_content: dict[str, Any] | None = None
    provider_specific_fields: dict[str, Any] | None = None
    function_provider_specific_fields: dict[str, Any] | None = None

    def to_openai_tool_call(self) -> dict[str, Any]:
        """序列化为 OpenAI 格式的 tool_call payload"""
        import json
        tc = {
            "id": self.id,
            "type": "function",
            "function": {
                "name": self.name,
                "arguments": json.dumps(self.arguments, ensure_ascii=False),
            },
        }
        if self.extra_content:
            tc["extra_content"] = self.extra_content
        return tc


@dataclass
class LLMResponse:
    """
    LLM 调用响应。
    对应 nanobot 的 LLMResponse，统一封装了来自不同 Provider 的 LLM 响应。
    """
    content: str | None                        # 文本回复（无工具调用时）
    tool_calls: list[ToolCallRequest] = field(default_factory=list)  # 工具调用列表
    finish_reason: str = "stop"                # stop / tool_calls / error / length
    usage: dict[str, int] = field(default_factory=dict)  # token 用量
    reasoning_content: str | None = None       # 推理过程（DeepSeek-R1 / Kimi 等）
    thinking_blocks: list[dict] | None = None  # 思维块（Anthropic extended thinking）

    @property
    def has_tool_calls(self) -> bool:
        """是否包含工具调用"""
        return len(self.tool_calls) > 0


@dataclass(frozen=True)
class GenerationSettings:
    """
    生成参数配置。
    对应 nanobot 的 GenerationSettings，控制 LLM 的生成行为。
    """
    temperature: float = 0.7
    max_tokens: int = 4096
    reasoning_effort: str | None = None  # low / medium / high — 启用推理模式


# ─── 技能匹配类型 ────────────────────────────────


@dataclass
class SkillMatch:
    """
    技能匹配结果：用户意图与某个技能的匹配结果。
    
    在 AgentLoop 中，处理每条消息前会调用 SkillManager.match()，
    将匹配到的技能信息注入 System Prompt，帮助 LLM 了解当前可用的能力包。
    """
    skill_name: str                              # 匹配到的技能名称
    skill_description: str                       # 技能描述
    score: float = 0.0                           # 匹配分数（0.0~1.0）
    matched_abilities: list[str] = field(default_factory=list)  # 匹配到的能力名
    matched_keywords: list[str] = field(default_factory=list)   # 命中的关键词


# ─── 消息总线类型 ────────────────────────────────


@dataclass
class InboundMessage:
    """
    入站消息：从通道到达 Agent 的消息。
    对应 nanobot 的 InboundMessage。
    """
    channel: str          # 来源频道（cli / streamlit / api）
    sender_id: str        # 发送者ID
    chat_id: str          # 会话ID
    content: str          # 消息内容
    timestamp: datetime = field(default_factory=datetime.now)
    media: list[str] = field(default_factory=list)   # 媒体URL
    metadata: dict[str, Any] = field(default_factory=dict)  # 通道专用元数据

    @property
    def session_key(self) -> str:
        """会话标识符：channel:chat_id"""
        return f"{self.channel}:{self.chat_id}"


@dataclass
class OutboundMessage:
    """
    出站消息：从 Agent 发往通道的消息。
    对应 nanobot 的 OutboundMessage。
    """
    channel: str
    chat_id: str
    content: str
    reply_to: str | None = None
    media: list[str] = field(default_factory=list)
    metadata: dict[str, Any] = field(default_factory=dict)
