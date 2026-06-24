"""
JD-Agent: 自实现AI智能体框架
参考 nanobot (openclaw 的轻量级 Python 替代) 架构设计

核心组件：
- AgentLoop: ReAct 循环引擎
- MessageBus: 异步消息总线（解耦通道与Agent）
- LLMProvider: 统一的 LLM Provider 抽象接口
- ToolRegistry: 工具注册与调度
- SessionManager: 会话管理与持久化（JSONL）
- MemoryStore: 三层记忆（MEMORY.md + HISTORY.md + 向量存储）
- ContextBuilder: System Prompt 构建器
"""

from agent.core.types import (
    LLMResponse,
    ToolCallRequest,
    InboundMessage,
    OutboundMessage,
    GenerationSettings,
    SkillMatch,
)
from agent.core.llm_provider import LLMProvider
from agent.bus.queue import MessageBus
from agent.loop import AgentLoop
from agent.orchestrator import AgentOrchestrator
from agent.tools.base import Tool
from agent.tools.registry import ToolRegistry
from agent.session.manager import SessionManager, Session
from agent.memory.memory_store import MemoryStore, MemoryConsolidator
from agent.memory.context_builder import ContextBuilder
from agent.skills.base import Skill, SkillAbility, SkillAsset
from agent.skills.manager import SkillManager
from agent.skills.loader import SkillLoader, SkillDiscoveryResult

__all__ = [
    # 类型
    "LLMResponse",
    "ToolCallRequest",
    "InboundMessage",
    "OutboundMessage",
    "GenerationSettings",
    "SkillMatch",
    # 核心
    "LLMProvider",
    "MessageBus",
    "AgentLoop",
    "AgentOrchestrator",
    # 工具
    "Tool",
    "ToolRegistry",
    # 会话
    "SessionManager",
    "Session",
    # 记忆
    "MemoryStore",
    "MemoryConsolidator",
    "ContextBuilder",
    # 技能
    "Skill",
    "SkillAbility",
    "SkillAsset",
    "SkillManager",
    "SkillLoader",
    "SkillDiscoveryResult",
]
