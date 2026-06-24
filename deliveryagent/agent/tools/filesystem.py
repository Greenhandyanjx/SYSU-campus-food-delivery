"""
Agent Tools: 文件系统工具
参考 nanobot 的 agent/tools/filesystem.py

提供文件的读取/写入/编辑/列表功能。
"""

from pathlib import Path
from typing import Any

from agent.tools.base import Tool


class ReadFileTool(Tool):
    """读取文件内容"""

    def __init__(self, workspace: Path):
        self._workspace = workspace

    @property
    def name(self) -> str:
        return "read_file"

    @property
    def description(self) -> str:
        return "读取文件内容。适用于读取代码文件、配置文件、文本文件等。"

    @property
    def parameters(self) -> dict[str, Any]:
        return {
            "path": {
                "type": "string",
                "description": "文件路径（相对于工作区或绝对路径）",
                "required": True,
            },
        }

    async def execute(self, path: str, **kwargs: Any) -> str:
        target = Path(path)
        if not target.is_absolute():
            target = self._workspace / target

        if not target.exists():
            return f"错误: 文件不存在: {target}"

        if not target.is_file():
            return f"错误: 不是文件: {target}"

        try:
            content = target.read_text(encoding="utf-8")
            # 对大文件截断显示
            max_chars = 50000
            if len(content) > max_chars:
                content = content[:max_chars] + f"\n\n... (文件过长，已截断前 {max_chars} 字符)"
            return content
        except Exception as e:
            return f"错误读取文件: {e}"


class WriteFileTool(Tool):
    """写入文件内容"""

    def __init__(self, workspace: Path):
        self._workspace = workspace

    @property
    def name(self) -> str:
        return "write_file"

    @property
    def description(self) -> str:
        return "写入或创建文件。注意：这会覆盖已有文件！"

    @property
    def parameters(self) -> dict[str, Any]:
        return {
            "path": {
                "type": "string",
                "description": "文件路径（相对于工作区或绝对路径）",
                "required": True,
            },
            "content": {
                "type": "string",
                "description": "文件内容",
                "required": True,
            },
        }

    async def execute(self, path: str, content: str, **kwargs: Any) -> str:
        target = Path(path)
        if not target.is_absolute():
            target = self._workspace / target

        try:
            target.parent.mkdir(parents=True, exist_ok=True)
            target.write_text(content, encoding="utf-8")
            return f"成功写入文件: {target} ({len(content)} 字符)"
        except Exception as e:
            return f"错误写入文件: {e}"
