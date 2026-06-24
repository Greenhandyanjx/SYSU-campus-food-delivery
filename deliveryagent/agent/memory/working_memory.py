"""
Memory: 工作记忆
保存当前任务的执行状态、中间结果、子任务进度
"""
from typing import Any, Optional
from datetime import datetime
from utils.logger import logger


class WorkingMemory:
    """
    工作记忆：跟踪当前任务的执行状态
    类似于 Agent 的"草稿纸"，Task Planner 写入，ReAct Loop 读取
    """

    def __init__(self):
        self._data: dict[str, Any] = {}
        self._task_stack: list[dict] = []  # 任务调用栈
        self.current_task: Optional[str] = None

    def set(self, key: str, value: Any) -> None:
        """设置工作记忆中的值"""
        self._data[key] = value
        logger.debug(f"[WorkingMemory] 设置: {key}")

    def get(self, key: str, default: Any = None) -> Any:
        """获取工作记忆中的值"""
        return self._data.get(key, default)

    def delete(self, key: str) -> None:
        """删除工作记忆中的某个值"""
        self._data.pop(key, None)
        logger.debug(f"[WorkingMemory] 删除: {key}")

    def push_task(self, task: dict) -> None:
        """压入子任务到栈"""
        task["_pushed_at"] = datetime.now().isoformat()
        self._task_stack.append(task)
        self.current_task = task.get("name", "unknown")
        logger.info(f"[WorkingMemory] 推入任务: {task.get('name')}")

    def pop_task(self) -> Optional[dict]:
        """弹出栈顶子任务"""
        if not self._task_stack:
            return None
        task = self._task_stack.pop()
        self.current_task = self._task_stack[-1].get("name") if self._task_stack else None
        logger.info(f"[WorkingMemory] 完成任务: {task.get('name')}")
        return task

    def get_task_stack(self) -> list[dict]:
        """获取任务栈"""
        return self._task_stack

    def get_context(self) -> dict:
        """
        获取完整的工作记忆上下文
        供 LLM 构建 prompt 时使用
        """
        return {
            "current_task": self.current_task,
            "task_stack": len(self._task_stack),
            "data_keys": list(self._data.keys()),
        }

    def clear(self) -> None:
        """清空工作记忆"""
        self._data.clear()
        self._task_stack.clear()
        self.current_task = None
        logger.info("[WorkingMemory] 工作记忆已清空")
