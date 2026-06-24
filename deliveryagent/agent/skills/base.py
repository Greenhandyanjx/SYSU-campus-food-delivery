"""
Skill 系统核心：Skill 抽象基类 + 能力/资产定义
=================================================

参照 OpenClaw 的 Skill 架构设计。

在 OpenClaw 中，Skill = SKILL.md（文档声明）+ 对应代码实现。
每个 Skill 有：
  - name/description: 基本描述（用在 SKILL.md 的 yaml 头部）
  - abilities: 能力列表（一组 Tool 声明）
  - assets: 资产列表（数据文件、模板、prompt 等）
  - tool_registry: 注册的本技能下的 Tool 集合

与 Tool 的区别：
  - Tool 是"原子操作"：读文件、查天气、搜索
  - Skill 是"能力包"：一组 Tool 的组合 + 文档 + 数据资产
  - 一个 Skill 包含多个 Tool，但不一定要全部注册到 Agent
  - Skill 可以根据用户意图"智能匹配"（而不只是 Function Calling）
"""

from __future__ import annotations

import copy
from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any

from agent.tools.base import Tool


# ─────────────────────────────────────────────────────────
# 基础数据类型
# ─────────────────────────────────────────────────────────


@dataclass
class SkillAbility:
    """
    技能能力声明。
    
    一个 Skill 可以提供多个能力，每个能力对应一个 Tool。
    name 与 ToolRegistry 中的工具名对应。
    
    额外元数据用于"智能发现"——Agent 可以根据用户意图匹配技能。
    """
    name: str                           # 能力名称（= Tool.name）
    description: str                    # 能力描述（LLM 意图匹配用）
    keywords: list[str] = field(default_factory=list)  # 触发关键词
    examples: list[str] = field(default_factory=list)   # 查询示例
    tool: Tool | None = None            # 关联的 Tool 实例


@dataclass
class SkillAsset:
    """
    技能资产声明。
    
    有些 Skill 需要附带的静态资源文件：
      - 数据文件（JSON/CSV/...）
      - Prompt 模板
      - 配置文件
      - 图片/文档
    
    OpenClaw 的 skills/weather/ 下就有这一类资产。
    """
    name: str                           # 资产名称
    path: str                           # 相对路径（相对于技能目录）
    type: str = "file"                  # 资产类型: file / prompt / data / config
    description: str = ""               # 资产用途描述


# ─────────────────────────────────────────────────────────
# Skill 基类
# ─────────────────────────────────────────────────────────


class Skill(ABC):
    """
    技能抽象基类（对标 OpenClaw 的 Skill 概念）。
    
    每个 Skill = 一个 skills/ 子目录，包含：
      SKILL.md   — 声明文档（yaml 头部 + markdown 正文）
      __init__.py — 代码实现（导出 skill_instance）
    
    核心方法：
      - register_to(): 将本技能的全部 Tool 注册到 ToolRegistry
      - match_query(): 判断技能是否匹配用户意图（可选重写）
    
    内置会帮你做的：
      - 根据 abilities 自动构造 Tool 注册
      - 管理技能目录路径
    """

    def __init__(self, skill_dir: Path | None = None):
        self.skill_dir: Path | None = skill_dir
        self._registered_tools: list[Tool] = []

    # ─── 子类必须实现的属性 ───────────────────

    @property
    @abstractmethod
    def name(self) -> str:
        """技能名称（唯一标识，用作技能目录名）"""
        pass

    @property
    @abstractmethod
    def description(self) -> str:
        """技能描述（告诉 LLM 这个技能能做什么）"""
        pass

    @property
    @abstractmethod
    def version(self) -> str:
        """技能版本号"""
        pass

    # ─── 子类可选实现 ────────────────────────

    @property
    def abilities(self) -> list[SkillAbility]:
        """
        技能提供的能力列表。
        
        默认空列表。子类可以覆盖此属性来声明能力。
        
        每个能力 = 一个 Tool 声明。
        Agent 会在 ToolRegistry 中注册这些能力对应的 Tool。
        """
        return []

    @property
    def assets(self) -> list[SkillAsset]:
        """
        技能附带的资产列表。
        
        默认空列表。子类可覆盖。
        """
        return []

    @property
    def requires(self) -> dict[str, Any]:
        """
        技能运行所需的依赖/资源（对标 OpenClaw 的 requires 字段）。
        
        例如：
          {"bins": ["ffmpeg"], "env": ["API_KEY"]}
        """
        return {}

    @property
    def categories(self) -> list[str]:
        """技能分类标签（用于技能管理界面分类展示）"""
        return []

    async def initialize(self) -> None:
        """
        技能初始化钩子。
        
        在 Agent 启动时调用，可以用于：
          - 检查环境
          - 加载模型
          - 准备数据
        """
        pass

    async def match_query(self, query: str) -> float:
        """
        判断该技能与用户查询的匹配度。
        
        返回 0.0 ~ 1.0 的匹配分数。
        默认实现：基于 keywords 和描述的关键词匹配。
        
        子类可以重写此方法实现更智能的匹配逻辑。
        """
        score = 0.0
        q_lower = query.lower()

        # 1. 关键词匹配
        for ability in self.abilities:
            for kw in ability.keywords:
                if kw.lower() in q_lower:
                    score += 0.3

        # 2. 查询示例匹配（模糊匹配）
        for ability in self.abilities:
            for ex in ability.examples:
                # 检查示例中的关键短语是否出现在用户查询中
                ex_keywords = ex.lower().split()
                matches = sum(1 for kw in ex_keywords if len(kw) > 1 and kw in q_lower)
                if matches >= 2:
                    score += 0.2

        return min(score, 1.0)

    # ─── 注册方法 ────────────────────────────

    def register_to(self, tool_registry) -> None:
        """
        将本技能的所有 Tool 注册到指定 ToolRegistry。
        
        遍历 abilities，将其中声明的 Tool 逐个注册。
        """
        for ability in self.abilities:
            if ability.tool is not None:
                tool_registry.register(ability.tool)
                self._registered_tools.append(ability.tool)

    def unregister_from(self, tool_registry) -> None:
        """从 ToolRegistry 中注销本技能的所有 Tool"""
        for tool in self._registered_tools:
            tool_registry.unregister(tool.name)
        self._registered_tools.clear()

    # ─── 工具列表 ────────────────────────────

    @property
    def tools(self) -> list[Tool]:
        """获取本技能关联的所有 Tool 实例"""
        return [a.tool for a in self.abilities if a.tool is not None]

    @property
    def tool_names(self) -> list[str]:
        """获取本技能关联的所有 Tool 名称"""
        return [t.name for t in self.tools]

    # ─── 序列化 ──────────────────────────────

    def to_dict(self) -> dict[str, Any]:
        """序列化为字典（用于 API 展示）"""
        return {
            "name": self.name,
            "description": self.description,
            "version": self.version,
            "categories": self.categories,
            "abilities": [
                {
                    "name": a.name,
                    "description": a.description,
                    "keywords": a.keywords,
                    "examples": a.examples,
                }
                for a in self.abilities
            ],
            "assets": [
                {"name": a.name, "path": a.path, "type": a.type}
                for a in self.assets
            ],
            "requires": self.requires,
        }

    def __repr__(self) -> str:
        return f"<Skill '{self.name}' v{self.version} ({len(self.abilities)} abilities)>"
