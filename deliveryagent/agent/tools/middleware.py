"""
Agent Tools: 工具调用中间件（迁移自原始项目）
"""
from typing import Callable
from utils.logger import logger


def monitor_tool_call(tool_name: str, args: dict, handler: Callable) -> str:
    """
    监控工具调用：记录日志、检测异常
    返回工具执行结果字符串
    """
    logger.info(f"[ToolMonitor] 执行工具: {tool_name}, 参数: {args}")

    try:
        result = handler()
        logger.info(f"[ToolMonitor] 工具 {tool_name} 执行成功")
        return result
    except Exception as e:
        logger.error(f"[ToolMonitor] 工具 {tool_name} 执行失败: {e}")
        raise
