"""
Agent Core: 错误回退策略
当工具调用连续失败时的处理
"""
from typing import Optional
from utils.logger import logger
from agent.core.tool_registry import get_all_tools


class FallbackStrategy:
    """错误回退策略"""

    def __init__(self, max_consecutive_failures: int = 3):
        self.max_consecutive_failures = max_consecutive_failures
        self.consecutive_failures = 0
        self.failed_tools: set[str] = set()

    def on_tool_success(self):
        """工具调用成功时重置计数器"""
        self.consecutive_failures = 0

    def on_tool_failure(self, tool_name: str, error: str) -> Optional[str]:
        """
        工具调用失败时的处理策略
        返回给模型的提示信息
        """
        self.consecutive_failures += 1
        self.failed_tools.add(tool_name)

        logger.warning(
            f"[Fallback] 工具 {tool_name} 失败 "
            f"(连续失败: {self.consecutive_failures}/{self.max_consecutive_failures})"
        )

        if self.consecutive_failures >= self.max_consecutive_failures:
            return ("我遇到了技术问题，无法完成您请求的操作。"
                    "建议您稍后再试，或者联系技术支持。")

        # 建议使用其他工具
        alternatives = [t.name for t in get_all_tools() if t.name != tool_name]
        if alternatives:
            return f"工具 '{tool_name}' 执行失败。您可以尝试使用以下工具: {', '.join(alternatives)}"

        return None

    def reset(self):
        """重置状态"""
        self.consecutive_failures = 0
        self.failed_tools.clear()
