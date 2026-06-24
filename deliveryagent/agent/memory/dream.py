"""
Dream — 长期记忆整理器
对标 nanobot 的 Dream：从 history.jsonl 中提取原子事实，增量更新 MEMORY.md

工作流（参考 nanobot Dream 的两阶段设计）：
Phase 1: 从 history.jsonl 读取自上次处理后新增的 consolidation 条目
Phase 2: 调用 LLM 提取原子事实 → 去重 → 增量编辑 MEMORY.md

与 nanobot Dream 的差异：
- 没有 tool-calling loop，直接用 LLM API 回调
- 一次处理一批，不循环
"""
from __future__ import annotations

import json
import re
from datetime import datetime
from pathlib import Path
from typing import Any, Callable

from loguru import logger

from agent.memory.memory_store import MemoryStore, HistoryJsonlStore


# 用于分析历史摘要的提示词
_DREAM_ANALYSIS_PROMPT = """你是一个记忆提取器。请从以下对话摘要中提取重要的事实和信息。

请按分类输出（中文）：
- **用户事实**: 个人信息、偏好、观点、习惯
- **决定**: 做出的选择、达成的结论
- **解决方案**: 成功的解决思路
- **事件**: 计划、截止日期、重要发生
- **偏好**: 沟通风格、工具偏好

注意：
- 每行输出一条原子事实，用短横线开头
- 跳过技术细节（可从代码推导的）
- 如果没什么值得记录的，输出: (nothing)
- 只输出事实列表，不要解释

对话摘要：
{summary}
"""

_DREAM_DEDUP_PROMPT = """你是记忆去重专家。下面是现有的 MEMORY.md 和新提取的事实。

请对比分析：
1. 新事实中哪些与已有内容重复（完全或高度重叠）→ 放入 to_skip
2. 新事实中哪些是已有内容的对立/修正 → 放入 to_replace (用新事实替换旧内容片段)
3. 新事实中哪些是全新的 → 放入 to_add

现有 MEMORY.md：
{existing_memory}

待添加的新事实：
{new_facts}

请返回 JSON 格式（不要其他内容）：
{{
    "to_skip": ["重复事实A", "重复事实B"],
    "to_replace": {{"旧内容片段": "新内容片段"}},
    "to_add": ["新增事实A", "新增事实B"]
}}
"""


class Dream:
    """长期记忆整理器。

    用法：
        dream = Dream(store, llm_summarize_fn=..., llm_dedup_fn=...)
        result = dream.run()  # 处理所有未归档的 consolidation 条目
        n_removed = dream.tidy_memory()  # 清理过时条目
    """

    _MAX_BATCH_SIZE = 10
    _MAX_MEMORY_CHARS = 32000

    def __init__(
        self,
        store: MemoryStore,
        history_jsonl: HistoryJsonlStore,
        summarizer_fn: Callable[[str], str] | None = None,
        dedup_fn: Callable[[str], str] | None = None,
    ):
        """
        Args:
            store: MemoryStore 实例（提供 MEMORY.md 读写）
            history_jsonl: HistoryJsonlStore 实例（提供 history.jsonl 读写）
            summarizer_fn: LLM 回调，接收分析提示，返回原子事实列表文本
            dedup_fn: LLM 回调，接收去重提示，返回 JSON 去重结果
        """
        self.store = store
        self.history_jsonl = history_jsonl
        self.summarizer_fn = summarizer_fn or self._builtin_summarizer
        self.dedup_fn = dedup_fn or self._builtin_dedup

    # ─── 公开接口 ───────────────────────────────────────

    def pending_count(self) -> int:
        """返回未处理的 consolidation 条目数"""
        last_cursor = self.history_jsonl.get_last_dream_cursor()
        entries = self.history_jsonl.read_since_cursor(last_cursor)
        return len([
            e for e in entries
            if e.get("type") == "consolidation" and e.get("summary")
        ])

    def run(self) -> dict[str, Any]:
        """执行一轮 Dream 处理。

        Returns:
            {"added": int, "replaced": int, "skipped": int, "cursor": int}
        """
        last_cursor = self.history_jsonl.get_last_dream_cursor()
        entries = self.history_jsonl.read_since_cursor(last_cursor)

        # 只处理 consolidation 类型的条目
        target = [e for e in entries if e.get("type") == "consolidation" and e.get("summary")]
        if not target:
            new_cursor = max((e.get("cursor", last_cursor) for e in entries), default=last_cursor)
            self.history_jsonl.set_last_dream_cursor(new_cursor)
            logger.info(f"[Dream] 无新 consolidation 条目 (cursor={new_cursor})")
            return {"added": 0, "replaced": 0, "skipped": 0, "cursor": new_cursor}

        logger.info(f"[Dream] 处理 {len(target)} 条未归档的 consolidation 记录")

        # 分批
        batches = [target[i:i + self._MAX_BATCH_SIZE] for i in range(0, len(target), self._MAX_BATCH_SIZE)]

        all_new_facts: list[str] = []
        for batch in batches:
            combined = "\n\n---\n\n".join(e.get("summary", "") for e in batch)
            facts = self._extract_facts(combined)
            if facts:
                all_new_facts.extend(facts)

        if not all_new_facts:
            new_cursor = max(e.get("cursor", last_cursor) for e in target)
            self.history_jsonl.set_last_dream_cursor(new_cursor)
            return {"added": 0, "replaced": 0, "skipped": 0, "cursor": new_cursor}

        logger.info(f"[Dream] 提取到 {len(all_new_facts)} 条新事实")

        # 去重并合并
        existing = self.store.read_long_term()
        result = self._dedup(existing, all_new_facts)

        # 应用变更
        updated = self._apply_changes(existing, result)

        if updated != existing:
            self.store.write_long_term(updated)
            logger.info(f"[Dream] MEMORY.md 已更新: +{len(result.get('to_add', []))} "
                        f"替换{len(result.get('to_replace', {}))} "
                        f"跳过{len(result.get('to_skip', []))}")

        # 更新 cursor
        new_cursor = max(e.get("cursor", last_cursor) for e in target)
        self.history_jsonl.set_last_dream_cursor(new_cursor)

        return {
            "added": len(result.get("to_add", [])),
            "replaced": len(result.get("to_replace", {})),
            "skipped": len(result.get("to_skip", [])),
            "cursor": new_cursor,
        }

    def tidy_memory(self) -> int:
        """清理 MEMORY.md 中的过时条目（移除含 '(stale)' 标记的行）

        Returns: 移除的行数
        """
        content = self.store.read_long_term()
        if not content:
            return 0
        lines = content.split("\n")
        kept = [l for l in lines if "(stale)" not in l]
        removed = len(lines) - len(kept)
        if removed > 0:
            self.store.write_long_term("\n".join(kept))
            logger.info(f"[Dream] 清理了 {removed} 条过时记忆")
        return removed

    def status(self) -> dict[str, Any]:
        """获取 Dream 的状态信息"""
        last_cursor = self.history_jsonl.get_last_dream_cursor()
        entries = self.history_jsonl.read_all()
        pending = len([
            e for e in entries
            if e.get("type") == "consolidation" and e.get("cursor", 0) > last_cursor
        ])
        memory_size = len(self.store.read_long_term())
        return {
            "last_cursor": last_cursor,
            "total_entries": len(entries),
            "pending_entries": pending,
            "memory_size_chars": memory_size,
        }

    # ─── 内部：提取与去重 ─────────────────────────────

    def _extract_facts(self, summary: str) -> list[str]:
        """从摘要中提取原子事实。"""
        if not self.summarizer_fn:
            return []
        prompt = _DREAM_ANALYSIS_PROMPT.format(summary=summary)
        try:
            response = self.summarizer_fn(prompt)
            text = response.strip()
            if text in ("(nothing)", ""):
                return []

            facts = []
            for line in text.split("\n"):
                line = line.strip().strip("- *")
                if line and not line.startswith("#") and not line.startswith("("):
                    facts.append(line)
            return facts[:20]
        except Exception as e:
            logger.warning(f"[Dream] 提取事实失败: {e}")
            return []

    def _builtin_dedup(self, prompt: str) -> str:
        """无 LLM 时的内置降级：直接返回所有事实作为新增。"""
        return '{"to_skip": [], "to_replace": {}, "to_add": []}'

    def _builtin_summarizer(self, prompt: str) -> str:
        """无 LLM 时的内置降级：直接取摘要的前几句。"""
        # prompt 的最后一段是 "对话摘要：\n{summary}"
        parts = prompt.split("对话摘要：")
        if len(parts) >= 2:
            summary = parts[-1].strip()
            lines = [l.strip("- *") for l in summary.split("\n") if l.strip() and not l.startswith("#")]
            return "\n".join(lines[:5])
        return ""

    def _dedup(self, existing: str, new_facts: list[str]) -> dict[str, Any]:
        """去重：将新事实与已有 MEMORY.md 对比。"""
        if not self.dedup_fn:
            return self._simple_dedup(existing, new_facts)

        prompt = _DREAM_DEDUP_PROMPT.format(
            existing_memory=existing[:self._MAX_MEMORY_CHARS] or "(空)",
            new_facts="\n".join(f"- {f}" for f in new_facts),
        )
        try:
            response = self.dedup_fn(prompt)
            result = json.loads(response)
            result.setdefault("to_skip", [])
            result.setdefault("to_replace", {})
            result.setdefault("to_add", [])
            return result
        except (json.JSONDecodeError, Exception) as e:
            logger.warning(f"[Dream] LLM 去重失败，使用简单去重: {e}")
            return self._simple_dedup(existing, new_facts)

    def _simple_dedup(self, existing: str, new_facts: list[str]) -> dict[str, Any]:
        """简单子串匹配去重（无 LLM 时的备选）"""
        to_add = []
        to_skip = []
        for fact in new_facts:
            if fact.lower() in existing.lower():
                to_skip.append(fact)
            else:
                to_add.append(fact)
        return {"to_skip": to_skip, "to_replace": {}, "to_add": to_add}

    def _apply_changes(self, existing: str, result: dict[str, Any]) -> str:
        """根据去重结果更新 MEMORY.md"""
        lines = existing.split("\n") if existing else ["# Memory", ""]

        # 替换
        for old_content, new_content in result.get("to_replace", {}).items():
            if old_content in existing:
                existing = existing.replace(old_content, new_content)
                lines = existing.split("\n")

        # 新增
        to_add = result.get("to_add", [])
        if to_add:
            insert_pos = len(lines)
            for i, line in enumerate(lines):
                if line.strip() and not line.startswith("#"):
                    insert_pos = i + 1
                    break

            # 确保前面有空行
            if insert_pos < len(lines) and lines[insert_pos - 1].strip():
                lines.insert(insert_pos, "")
                insert_pos += 1

            timestamp = datetime.now().strftime("%Y-%m-%d %H:%M")
            for fact in reversed(to_add):
                lines.insert(insert_pos, f"- [{timestamp}] {fact}")

        return "\n".join(lines)

    def read_memory_file(self) -> str:
        """读取 MEMORY.md 全文的快捷方式"""
        return self.store.read_long_term()
