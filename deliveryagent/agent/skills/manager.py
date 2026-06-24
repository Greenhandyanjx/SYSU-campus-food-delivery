"""
SkillManager: 技能管理器
========================

对标 OpenClaw 的可用技能表（available_skills）机制。

在 OpenClaw 中：
  - 用户配置 plugin/skills 列表
  - 系统运行时可查询/过滤/匹配可用 skill
  
在 JD-Agent 中：
  - SkillManager 持有所有已发现的 Skill
  - 提供注册、查询、意图匹配、条件过滤
  - 供 Agent 在对话中"智能选择技能"

核心功能：
  1. register() / unregister(): 管理技能
  2. match(query) → [Skill]: 根据用户意图匹配最相关的技能
  3. get_skill_by_name():  按名查找
  4. get_all(): 获取所有技能（按分类组织）
  5. compute_context(): 将匹配的技能上下文注入 System Prompt
"""

from __future__ import annotations

from pathlib import Path
from typing import Any

from loguru import logger

from agent.skills.base import Skill
from agent.skills.loader import SkillLoader, SkillDiscoveryResult


# ─────────────────────────────────────────────────────────
# MatchResult: 技能匹配结果
# ─────────────────────────────────────────────────────────


class SkillMatchResult:
    """
    用户意图与技能的匹配结果。
    
    包含：
      - skill: 匹配到的技能
      - score: 匹配分数（0.0~1.0）
      - matched_abilities: 匹配到的具体能力
    """
    def __init__(
        self,
        skill: Skill,
        score: float = 0.0,
        matched_abilities: list[str] | None = None,
    ):
        self.skill = skill
        self.score = score
        self.matched_abilities = matched_abilities or []

    def __repr__(self) -> str:
        return (
            f"<MatchResult skill={self.skill.name} "
            f"score={self.score:.2f} abilities={self.matched_abilities}>"
        )


# ─────────────────────────────────────────────────────────
# SkillManager
# ─────────────────────────────────────────────────────────


class SkillManager:
    """
    技能管理器。
    
    职责：
      1. 发现并加载技能（通过 SkillLoader）
      2. 持有所加载的 Skill 列表
      3. 提供查询/匹配接口
      4. 管理技能到 ToolRegistry 的注册
    """

    def __init__(self, skills_dir: str | Path | None = None):
        """
        Args:
            skills_dir: 技能目录路径（可选，可通过 discover() 设置）
        """
        self._skills: dict[str, Skill] = {}  # name → Skill
        self._discovery_result: SkillDiscoveryResult | None = None
        self._skills_dir: Path | None = (
            Path(skills_dir).expanduser().resolve() if skills_dir else None
        )

    # ─── 发现与加载 ───────────────────────────

    async def discover(self, skills_dir: str | Path | None = None) -> SkillDiscoveryResult:
        """
        扫描并加载技能目录中的所有技能。
        
        如果 skills_dir 为空，使用构造时传入的目录。
        
        Returns:
            SkillDiscoveryResult（成功/失败详情）
        """
        target_dir = skills_dir or self._skills_dir
        if target_dir is None:
            raise ValueError("未指定技能目录，请在构造时传入或调用 discover() 时指定")

        self._skills_dir = Path(target_dir).expanduser().resolve()
        loader = SkillLoader(self._skills_dir)
        result = await loader.discover_all()
        self._discovery_result = result

        # 添加到内部缓存
        for skill in result.discovered:
            self._skills[skill.name] = skill

        return result

    def register(self, skill: Skill) -> None:
        """手动注册一个技能"""
        self._skills[skill.name] = skill
        logger.info(f"[SkillManager] 注册技能: {skill.name}")

    def unregister(self, name: str) -> None:
        """注销一个技能"""
        self._skills.pop(name, None)
        logger.info(f"[SkillManager] 注销技能: {name}")

    # ─── 查询 ────────────────────────────────

    def get(self, name: str) -> Skill | None:
        """按名称获取技能"""
        return self._skills.get(name)

    def has(self, name: str) -> bool:
        """检查技能是否存在"""
        return name in self._skills

    def get_all(self) -> list[Skill]:
        """获取所有已注册的技能"""
        return list(self._skills.values())

    def get_by_category(self, category: str) -> list[Skill]:
        """按分类获取技能（不区分大小写）"""
        cat_lower = category.lower()
        return [
            s for s in self._skills.values()
            if any(c.lower() == cat_lower for c in s.categories)
        ]

    def get_by_ability(self, tool_name: str) -> Skill | None:
        """根据 Tool 名称查找拥有该能力的技能"""
        for skill in self._skills.values():
            if any(a.name == tool_name for a in skill.abilities):
                return skill
        return None

    # ─── 意图匹配 ────────────────────────────

    async def match(self, query: str, top_k: int = 3) -> list[SkillMatchResult]:
        """
        根据用户查询匹配最相关的技能。
        
        遍历所有技能，调用其 match_query() 方法，返回按分数降序排列的结果。
        """
        results: list[SkillMatchResult] = []
        for skill in self._skills.values():
            score = await skill.match_query(query)
            if score > 0:
                # 找出匹配到的能力
                matched = self._find_matched_abilities(skill, query)
                results.append(SkillMatchResult(skill, score, matched))

        results.sort(key=lambda r: r.score, reverse=True)
        return results[:top_k]

    def _find_matched_abilities(self, skill: Skill, query: str) -> list[str]:
        """找出技能中匹配用户查询的能力名列表"""
        matched = []
        q_lower = query.lower()
        for ability in skill.abilities:
            # 关键词匹配
            for kw in ability.keywords:
                if kw.lower() in q_lower:
                    matched.append(ability.name)
                    break
        return matched

    # ─── 上下文注入 ──────────────────────────

    def compute_context(self, matched: list[SkillMatchResult]) -> str:
        """
        将匹配到的技能信息转为 LLM 可以理解的上下文文本。
        
        在 System Prompt 中插入类似：
          [可用技能]
          - weather: 查询天气（get_weather）
            触发词: 天气, 温度, 下雨,...
          - rag: 知识库检索（rag_query）
            触发词: 知识, 文档, 查询,...
        """
        if not matched:
            return ""

        parts = ["\n[当前可用的技能包]"]
        for mr in matched:
            skill = mr.skill
            abilities_desc = ", ".join(
                f"{a.name}: {a.description}" for a in skill.abilities
            )
            parts.append(
                f"- [{skill.name}(v{skill.version})] {skill.description}\n"
                f"  能力: {abilities_desc}"
            )

        return "\n".join(parts)

    # ─── 批量注册到 ToolRegistry ─────────────

    def register_all_tools(self, tool_registry) -> int:
        """
        将所有技能下所有 Tool 注册到指定 ToolRegistry。
        
        Returns:
            注册的 Tool 数量
        """
        count = 0
        for skill in self._skills.values():
            before = len(tool_registry._tools) if hasattr(tool_registry, '_tools') else 0
            skill.register_to(tool_registry)
            after = len(tool_registry._tools) if hasattr(tool_registry, '_tools') else 0
            count += (after - before)
        logger.info(f"[SkillManager] 批量注册工具: {count} 个")
        return count

    def unregister_all_tools(self, tool_registry) -> None:
        """从 ToolRegistry 注销所有技能下的 Tool"""
        for skill in self._skills.values():
            skill.unregister_from(tool_registry)

    # ─── 序列化 ──────────────────────────────

    def to_dict(self) -> list[dict[str, Any]]:
        """序列化为字典列表（用于 API 展示）"""
        return [skill.to_dict() for skill in self._skills.values()]

    # ─── 属性 ────────────────────────────────

    @property
    def count(self) -> int:
        return len(self._skills)

    @property
    def discovery_result(self) -> SkillDiscoveryResult | None:
        return self._discovery_result

    def __len__(self) -> int:
        return len(self._skills)

    def __contains__(self, name: str) -> bool:
        return name in self._skills

    def __repr__(self) -> str:
        return f"<SkillManager count={len(self._skills)}>"
