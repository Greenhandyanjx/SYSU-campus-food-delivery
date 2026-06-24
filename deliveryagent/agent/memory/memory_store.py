"""
Agent Core: 记忆系统（三层记忆）
参考 nanobot 的 agent/memory.py 设计

记忆系统分为四层：
1. MEMORY.md（长期事实）— 跨会话持久化的用户偏好、关键事实
2. HISTORY.md（历史日志）— grep 可搜索的行为日志
3. history.jsonl（中期压缩记忆）— Consolidator 将旧对话压缩为摘要条目
4. Session Messages（短期记忆）— 当前会话上下文（由 session/manager.py 管理）

关键特点：
- MemoryStore 负责文件级别的读写
- HistoryJsonlStore 负责 history.jsonl 的追加/读取（对标 nanobot 的 history.jsonl）
- MemoryConsolidator 负责在合适的时机（按时间/消息数）自动 consolidation
- Consolidation 通过 LLM tool call 让模型自己决定写什么
- Dream 定期从 history.jsonl 提取原子事实更新 MEMORY.md
"""

import json
import os
from datetime import datetime, timezone
from pathlib import Path
from typing import TYPE_CHECKING, Any

from loguru import logger

from agent.core.types import LLMResponse, ToolCallRequest

if TYPE_CHECKING:
    from agent.core.llm_provider import LLMProvider
    from agent.session.manager import Session, SessionManager


class MemoryStore:
    """
    Memory 文件存储层。
    
    管理两个文件：
    - memory/MEMORY.md: 长期事实（跨会话持久）
    - memory/HISTORY.md: 历史日志（可 grep 搜索）
    
    参考 nanobot 的 MemoryStore，但简化了 history.jsonl 的迁移逻辑。
    """

    def __init__(self, workspace: Path):
        self.memory_dir = _ensure_dir(workspace / "memory")
        self.memory_file = self.memory_dir / "MEMORY.md"
        self.history_file = self.memory_dir / "HISTORY.md"
        self.history_jsonl = HistoryJsonlStore(workspace)  # 新增：history.jsonl 管线

    # ─── MEMORY.md 读写 ────────────────────────

    def read_long_term(self) -> str:
        """读取长期记忆"""
        if self.memory_file.exists():
            return self.memory_file.read_text(encoding="utf-8")
        return ""

    def write_long_term(self, content: str) -> None:
        """写入长期记忆"""
        self.memory_file.write_text(content, encoding="utf-8")

    def get_memory_context(self) -> str:
        """获取用于 system prompt 的记忆上下文"""
        content = self.read_long_term()
        return f"## Long-term Memory\n{content}" if content else ""

    # ─── HISTORY.md 操作 ───────────────────────

    def append_history(self, entry: str) -> None:
        """追加一条历史日志"""
        with open(self.history_file, "a", encoding="utf-8") as f:
            f.write(entry.rstrip() + "\n\n")


class HistoryJsonlStore:
    """history.jsonl 存储层（对标 nanobot 的 history.jsonl 管线）

    格式：每行一个 JSON，包含：
    - type: "consolidation" | "dream_note" | "user_event"
    - summary: 摘要文本
    - message_count: 压缩的消息数
    - cursor: 消息位置指针
    - timestamp: ISO 时间戳
    - session_key: 会话标识
    """

    def __init__(self, workspace: Path, max_entries: int = 500):
        self.history_file = _ensure_dir(workspace / "memory") / "history.jsonl"
        self.max_entries = max_entries

    def append(self, entry: dict) -> None:
        """追加一条 JSON 到 history.jsonl"""
        with open(self.history_file, "a", encoding="utf-8") as f:
            f.write(json.dumps(entry, ensure_ascii=False) + "\n")

    def read_all(self) -> list[dict]:
        """读取全部条目"""
        entries: list[dict] = []
        if not self.history_file.exists():
            return entries
        with open(self.history_file, "r", encoding="utf-8") as f:
            for line in f:
                line = line.strip()
                if line:
                    try:
                        entries.append(json.loads(line))
                    except json.JSONDecodeError:
                        continue
        return entries

    def read_since_cursor(self, cursor: int) -> list[dict]:
        """返回 cursor 之后的所有条目"""
        return [e for e in self.read_all() if e.get("cursor", 0) > cursor]

    def compact(self) -> None:
        """超限时丢弃最旧条目"""
        if self.max_entries <= 0:
            return
        entries = self.read_all()
        if len(entries) <= self.max_entries:
            return
        kept = entries[-self.max_entries:]
        self._overwrite(kept)

    def _overwrite(self, entries: list[dict]) -> None:
        """原子覆盖写入"""
        tmp = self.history_file.with_suffix(".jsonl.tmp")
        try:
            with open(tmp, "w", encoding="utf-8") as f:
                for entry in entries:
                    f.write(json.dumps(entry, ensure_ascii=False) + "\n")
                f.flush()
                os.fsync(f.fileno())
            os.replace(tmp, self.history_file)
        except BaseException:
            tmp.unlink(missing_ok=True)
            raise

    def get_last_dream_cursor(self) -> int:
        """读取 Dream 已处理到的 cursor 位置"""
        cursor_file = self.history_file.parent / ".dream_cursor"
        try:
            return int(cursor_file.read_text(encoding="utf-8").strip())
        except (FileNotFoundError, ValueError):
            return 0

    def set_last_dream_cursor(self, cursor: int) -> None:
        """记录 Dream 最后一次处理到的 cursor 位置"""
        cursor_file = self.history_file.parent / ".dream_cursor"
        cursor_file.write_text(str(cursor), encoding="utf-8")


def _ensure_dir(path: Path) -> Path:
    path.mkdir(parents=True, exist_ok=True)
    return path


class MemoryConsolidator:
    """
    记忆 Consolidation 引擎。
    
    在合适的时机调用 LLM 来总结对话并更新 MEMORY.md / HISTORY.md。
    
    触发时机：
    - 每 N 条消息后（按消息数）
    - 每 T 分钟（按时间间隔）
    
    参考 nanobot 的 MemoryConsolidator，但更简洁。
    """

    def __init__(
        self,
        workspace: Path,
        provider: "LLMProvider",
        model: str,
        memory_store: MemoryStore,
        context_builder: Any,  # ContextBuilder, 用于 build_messages
        sessions: "SessionManager",
        consolidate_every_n: int = 20,  # 每 N 条消息后触发
        consolidate_interval_min: int = 15,  # 最短间隔 15 分钟
    ):
        self.workspace = workspace
        self.provider = provider
        self.model = model
        self.memory = memory_store
        self.context_builder = context_builder
        self.sessions = sessions
        self._consolidate_every_n = consolidate_every_n
        self._consolidate_interval_min = consolidate_interval_min
        self._last_consolidation_time: float = 0.0  # epoch 时间戳

    async def maybe_consolidate(self, session: "Session") -> None:
        """
        检查是否需要进行记忆 consolidation。
        
        条件：
        1. 自上次 consolidation 以来有新消息
        2. 达到消息数阈值或时间间隔
        """
        new_count = len(session.messages) - session.last_consolidated
        if new_count == 0:
            return

        # 双阈值：消息数达到阈值，或时间达标且有一定量的新消息
        now = datetime.now().timestamp()
        time_since_last = now - self._last_consolidation_time

        hit_count = new_count >= self._consolidate_every_n
        hit_time = (
            time_since_last >= self._consolidate_interval_min * 60
            and new_count >= min(5, self._consolidate_every_n // 2)
        )

        if not hit_count and not hit_time:
            return

        logger.info(f"触发记忆 consolidation "
                     f"(消息数阈值: {new_count}/{self._consolidate_every_n}, "
                     f"时间间隔: {int(time_since_last)}s/{self._consolidate_interval_min * 60}s)")

        await self._consolidate(session)
        self._last_consolidation_time = now

    async def _consolidate(self, session: "Session") -> None:
        """
        执行记忆 consolidation。
        
        流程：
        1. 获取未 consolidated 的对话历史
        2. 用 LLM 分析对话生成 history_entry 摘要
        3. 写入 history.jsonl 和 HISTORY.md
        4. 更新 session 的 last_consolidated 指针
        """
        # 获取未 consolidated 的历史
        unconsolidated = session.messages[session.last_consolidated:]
        if not unconsolidated:
            return

        # 构建 consolidation prompt
        now = datetime.now()
        date_str = now.strftime("%Y-%m-%d %H:%M")
        messages = [
            {
                "role": "system",
                "content": (
                    f"当前日期时间：{date_str}\n\n"
                    "你是一个对话日志系统。请分析以下对话历史，用一句话概括关键内容。"
                    "返回 JSON 格式：\n\n"
                    '{\n'
                    '  "history_entry": "当前日期 YYYY-MM-DD HH:MM:SS 摘要"\n'
                    '}\n\n'
                    f"history_entry 必须以「{now.strftime('%Y-%m-%d')} 」开头，"
                    "不要使用其他日期！"
                ),
            },
            {
                "role": "user",
                "content": (
                    f"## 当前日期\n{now.strftime('%Y年%m月%d日 %H:%M')}\n\n"
                    f"## 这段对话历史（共 {len(unconsolidated)} 条消息）\n"
                    + json.dumps([
                        {"role": m["role"], "content": (m.get("content") or "")[:500]}
                        for m in unconsolidated[-10:]  # 只取最近 10 条，避免 token 爆炸
                    ], ensure_ascii=False, indent=2)
                    + "\n\n请返回 JSON 格式的 history_entry。"
                    f"\n\n重要：history_entry 必须以「{now.strftime('%Y-%m-%d')} 」开头，不要使用其他日期！"
                ),
            },
        ]

        try:
            response = await self.provider.chat(
                messages=messages,
                model=self.model,
                max_tokens=2048,
                temperature=0.3,
            )

            if response.content:
                history_entry = self._parse_consolidation_result(response.content)
                if history_entry:
                    # 写入 HISTORY.md
                    self.memory.append_history(history_entry)
                    # 写入 history.jsonl（Dream 的素材）
                    self.memory.history_jsonl.append({
                        "type": "consolidation",
                        "session_key": session.key,
                        "summary": history_entry,
                        "message_count": len(unconsolidated),
                        "cursor": len(session.messages),
                        "timestamp": datetime.now(tz=timezone.utc).isoformat(),
                    })
                    self.memory.history_jsonl.compact()
                    # 更新 session 指针
                    session.last_consolidated = len(session.messages)
                    logger.info("记忆 consolidation 完成")
        except Exception as e:
            logger.error(f"记忆 consolidation 失败: {e}")

    @staticmethod
    def _parse_consolidation_result(content: str) -> str | None:
        """解析 LLM 返回的 consolidation 结果，提取 history_entry"""
        from agent.utils.json_parser import extract_json

        data = extract_json(content)
        if data and isinstance(data, dict):
            entry = data.get("history_entry", "")
            return entry.strip() if entry else None
        return None
