"""
Memory: 短期记忆
管理当前会话的消息历史，支持滑动窗口
"""
from typing import Optional
from datetime import datetime
from utils.logger import logger


class ShortTermMemory:
    """
    短期记忆：保存当前会话的消息历史
    使用滑动窗口控制最大消息数
    """

    def __init__(self, max_messages: int = 50):
        self.max_messages = max_messages
        self.messages: list[dict] = []

    def add(self, role: str, content: str, metadata: Optional[dict] = None) -> None:
        """添加一条消息到短期记忆"""
        msg = {
            "role": role,
            "content": content,
            "timestamp": datetime.now().isoformat(),
        }
        if metadata:
            msg["metadata"] = metadata

        self.messages.append(msg)
        self._trim()

    def add_tool_call(self, tool_name: str, args: dict, result: str,
                      success: bool = True) -> None:
        """添加工具调用记录"""
        self.messages.append({
            "role": "tool",
            "content": result,
            "tool_call_id": tool_name,
            "timestamp": datetime.now().isoformat(),
        })

    def get_recent(self, count: Optional[int] = None) -> list[dict]:
        """获取最近的 N 条消息"""
        if count is None:
            return self.messages
        return self.messages[-count:]

    def get_all(self) -> list[dict]:
        """获取所有消息"""
        return self.messages

    def clear(self) -> None:
        """清空短期记忆"""
        self.messages.clear()
        logger.info("[ShortTermMemory] 短期记忆已清空")

    def _trim(self) -> None:
        """裁剪消息数量到最大限制"""
        if len(self.messages) > self.max_messages:
            self.messages = self.messages[-self.max_messages:]
            logger.debug(f"[ShortTermMemory] 裁剪到 {self.max_messages} 条")
