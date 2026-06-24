"""
SkillLoader: SKILL.md 解析器 + 文件系统自动发现
==================================================

对标 OpenClaw 的技能发现机制：
  OpenClaw 扫描 skills/ 目录 → 读取 SKILL.md 的 yaml 头部 →
  解析 metadata / requires / description → 加载对应的 __init__.py

解构流程：
  1. scan(): 扫描 skills/ 目录，发现所有子目录
  2. discover(): 对每个子目录，读取 SKILL.md → 加载 __init__.py
  3. 返回 SkillDiscoveryResult（发现的技能元数据）

本系统支持两种技能形式：
  a) 带 SKILL.md 的"文档化技能"（推荐，会自动发现 Tool 继承的类）
  b) 纯代码技能（只有 __init__.py，但建议至少提供 SKILL.md）
"""

from __future__ import annotations

import importlib.util
import os
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any

import yaml
from loguru import logger

from agent.skills.base import Skill


# ─────────────────────────────────────────────────────────
# 发现结果
# ─────────────────────────────────────────────────────────


@dataclass
class SkillDiscoveryResult:
    """
    技能发现结果。
    
    discovered:  成功加载的技能列表
    failed:      加载失败的技能列表（附错误信息）
    total_found: 总共发现的技能目录数
    """
    discovered: list[Skill] = field(default_factory=list)
    failed: list[dict[str, str]] = field(default_factory=list)

    @property
    def total_found(self) -> int:
        return len(self.discovered) + len(self.failed)

    @property
    def success_count(self) -> int:
        return len(self.discovered)

    @property
    def failed_count(self) -> int:
        return len(self.failed)


@dataclass
class SkillManifest:
    """
    SKILL.md 解析结果（yaml 头部内容）。
    
    对应 OpenClaw SKILL.md 的 yaml front-matter：
      ---
      name: weather
      description: "..."
      metadata:
        openclaw:
          emoji: ☔
          requires: { bins: ["curl"] }
      ---
    """
    name: str = ""
    description: str = ""
    version: str = "0.1.0"
    metadata: dict[str, Any] = field(default_factory=dict)
    categories: list[str] = field(default_factory=list)

    @property
    def requires(self) -> dict[str, Any]:
        """从 metadata 中提取 requires（OpenClaw 模式）"""
        if "openclaw" in self.metadata and isinstance(self.metadata["openclaw"], dict):
            return self.metadata["openclaw"].get("requires", {})
        return {}

    @property
    def emoji(self) -> str:
        """从 metadata 中提取 emoji（OpenClaw 模式）"""
        if "openclaw" in self.metadata and isinstance(self.metadata["openclaw"], dict):
            return self.metadata["openclaw"].get("emoji", "📦")
        return "📦"

    @classmethod
    def from_yaml_header(cls, text: str) -> SkillManifest | None:
        """
        从 SKILL.md 的 yaml front-matter 中解析元数据。
        
        SKILL.md 格式：
          ---
          name: weather
          description: "..."
          ---
          正文内容...
        """
        text = text.lstrip()
        if not text.startswith("---"):
            return None

        # 找到结束的 ---
        end_idx = text.find("---", 3)
        if end_idx == -1:
            return None

        yaml_str = text[3:end_idx].strip()
        try:
            data = yaml.safe_load(yaml_str)
            if not isinstance(data, dict):
                return None
            return cls(
                name=data.get("name", ""),
                description=data.get("description", ""),
                version=data.get("version", "0.1.0"),
                metadata=data.get("metadata", {}),
                categories=data.get("categories", []),
            )
        except yaml.YAMLError as e:
            logger.warning(f"[SkillLoader] YAML 解析失败: {e}")
            return None


# ─────────────────────────────────────────────────────────
# SkillLoader
# ─────────────────────────────────────────────────────────


class SkillLoader:
    """
    技能加载器：文件系统扫描 + SKILL.md 解析 + 代码加载。
    
    用法：
      loader = SkillLoader(Path("skills"))
      result = await loader.discover_all()
      for skill in result.discovered:
          skill.register_to(tool_registry)
    """

    def __init__(self, skills_dir: str | Path):
        """
        Args:
            skills_dir: 技能目录根路径（扫描该目录的所有子目录）
        """
        self._skills_dir = Path(skills_dir).expanduser().resolve()

    # ─── 扫描发现 ────────────────────────────

    async def discover_all(self) -> SkillDiscoveryResult:
        """
        扫描所有技能目录，发现并加载技能。
        
        流程：
          1. 列出 skills/ 下所有子目录
          2. 对每个子目录，读取 SKILL.md
          3. 尝试加载 __init__.py 中的 skill_instance
          4. 返回所有成功/失败的结果
        """
        result = SkillDiscoveryResult()

        if not self._skills_dir.exists():
            logger.info(f"[SkillLoader] 技能目录不存在: {self._skills_dir}，跳过扫描")
            return result

        # 列出所有子目录
        for entry in sorted(self._skills_dir.iterdir()):
            if not entry.is_dir():
                continue
            if entry.name.startswith("_"):
                continue  # 跳过 __pycache__ 等特殊目录

            try:
                skill = await self._load_skill(entry)
                if skill is not None:
                    result.discovered.append(skill)
                    logger.info(f"[SkillLoader] ✅ 已发现技能: {skill.name} ({entry.name})")
                else:
                    result.failed.append({
                        "dir": entry.name,
                        "error": "加载失败：__init__.py 未导出 skill_instance",
                    })
            except Exception as e:
                result.failed.append({
                    "dir": entry.name,
                    "error": str(e),
                })
                logger.warning(f"[SkillLoader] ❌ 技能加载失败: {entry.name}: {e}")

        logger.info(
            f"[SkillLoader] 技能扫描完成: "
            f"{result.success_count} 成功 / {result.failed_count} 失败"
        )
        return result

    async def _load_skill(self, skill_dir: Path) -> Skill | None:
        """
        加载单个技能目录。
        
        流程：
          1. 读取 SKILL.md → 解析 yaml 头部
          2. 尝试加载 __init__.py → 获取 skill_instance
          3. 如果只有 __init__.py 没有 SKILL.md → 仍然加载
          4. 如果只有 SKILL.md 没有 __init__.py → 跳过（需要代码）
        """
        manifest: SkillManifest | None = None
        skill_instance: Skill | None = None

        skill_md = skill_dir / "SKILL.md"
        init_py = skill_dir / "__init__.py"

        # ── 1. 读取 SKILL.md ──
        if skill_md.exists():
            content = skill_md.read_text(encoding="utf-8")
            manifest = SkillManifest.from_yaml_header(content)

        # ── 2. 加载 __init__.py ──
        if not init_py.exists():
            # 只有 SKILL.md 没有 __init__.py → 纯文档化技能
            # 对应 OpenClaw 的天气技能：需要有 __init__.py 才能加载
            logger.debug(f"[SkillLoader] 跳过 {skill_dir.name}：缺少 __init__.py")
            return None

        skill_instance = await self._load_init_py(init_py, manifest, skill_dir)
        return skill_instance

    async def _load_init_py(
        self,
        init_py: Path,
        manifest: SkillManifest | None,
        skill_dir: Path,
    ) -> Skill | None:
        """
        动态加载 __init__.py 并提取 skill_instance。
        
        使用 importlib 动态导入，避免污染全局命名空间。

        约定：
          技能 __init__.py 必须导出名为 skill_instance 的 Skill 实例。
          （对标 OpenClaw 每个 skill 目录提供一个 Skill 实例）
        """
        module_name = f"skills_{skill_dir.name}_{id(skill_dir)}"

        spec = importlib.util.spec_from_file_location(module_name, init_py)
        if spec is None or spec.loader is None:
            return None

        module = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(module)

        # 查找 Skill 实例
        # 1. 优先找 skill_instance 变量
        # 2. 其次找 Skill 子类的实例
        skill = getattr(module, "skill_instance", None)
        if skill is None:
            # 查找 Skill 子类的实例
            for attr_name in dir(module):
                attr = getattr(module, attr_name)
                if isinstance(attr, Skill) and type(attr) is not Skill:
                    skill = attr
                    break

        if skill is None:
            return None

        # 设置技能目录
        if skill.skill_dir is None:
            skill.skill_dir = skill_dir

        return skill

    # ─── 辅助方法 ────────────────────────────

    def get_skill_dirs(self) -> list[Path]:
        """获取所有候选技能目录名"""
        if not self._skills_dir.exists():
            return []
        return [
            entry for entry in sorted(self._skills_dir.iterdir())
            if entry.is_dir() and not entry.name.startswith("_")
        ]

    @property
    def skills_dir(self) -> Path:
        return self._skills_dir
