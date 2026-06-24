"""
JD-Agent Skill System: 技能系统

模仿 OpenClaw 的 skill 架构，为 Agent 提供可插拔的"技能"定义和执行能力。

核心区别：Skill ≠ Tool
======================
在 OpenClaw 中，Skill 是一组能力的封装，包含：
  - SKILL.md（文档 + 元数据）
  - 一个或多个 Tool 实现

而 JD-Agent 的 Tool 是单个可调用的能力。
Skill 则是一个"能力包"，一组相关联的 Tool + 文档 + 自动发现机制。

核心组件：
  - Skill          抽象基类（定义技能规范）
  - SkillManager   技能管理器（注册、发现、匹配、加载）
  - SkillLoader    SKILL.md 解析器 + 文件系统发现
  - SKILL_PATTERN  技能规范文档（对标 OpenClaw SKILL.md）

工作流程：
  1. Agent 启动时，SkillLoader 扫描 skills/ 目录
  2. 发现每个子目录中的 SKILL.md + __init__.py
  3. 解析 SKILL.md 的 metadata + 能力声明
  4. 调用 __init__.py 中的注册函数，注册 Skill 和底层 Tool
  5. SkillManager 统一管理所有技能
"""

from agent.skills.base import Skill, SkillAsset, SkillAbility
from agent.skills.manager import SkillManager
from agent.skills.loader import SkillLoader, SkillDiscoveryResult

__all__ = [
    "Skill",
    "SkillAsset",
    "SkillAbility",
    "SkillManager",
    "SkillLoader",
    "SkillDiscoveryResult",
]
