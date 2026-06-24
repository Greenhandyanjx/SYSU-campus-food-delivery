"""
Planner: 任务规划器
借鉴 Claude Code 的 Plan-Solve 模式，
将复杂任务分解为可执行的子任务 DAG
"""
import json
from typing import Any, Optional
from datetime import datetime
from loguru import logger
from model.factory import chat_model as llm
from agent.memory.working_memory import WorkingMemory
from agent.memory.memory_store import MemoryStore


class Task:
    """单个任务节点"""

    def __init__(self, task_id: str, name: str, description: str,
                 depends_on: list[str] = None, tool_name: str = None):
        self.task_id = task_id
        self.name = name
        self.description = description
        self.depends_on = depends_on or []  # 依赖的任务ID列表
        self.tool_name = tool_name  # 执行该任务所需的工具
        self.status = "pending"  # pending | running | completed | failed | skipped
        self.result: Optional[str] = None
        self.error: Optional[str] = None
        self.created_at = datetime.now().isoformat()
        self.completed_at: Optional[str] = None

    def to_dict(self) -> dict:
        return {
            "task_id": self.task_id,
            "name": self.name,
            "description": self.description,
            "depends_on": self.depends_on,
            "tool_name": self.tool_name,
            "status": self.status,
            "result": self.result,
            "error": self.error,
        }


class TaskPlanner:
    """
    任务规划器：
    1. 接收用户复杂请求
    2. 调用LLM进行任务分解（结合长期记忆）
    3. 构建任务依赖图
    4. 逐步执行

    扩展（参考 nanobot 记忆管线）：
    - 规划时自动加载长期记忆（MEMORY.md），让 LLM 的拆解更明智
    - 任务执行后自动记录到 history.jsonl（供 Dream 归档）
    - 复合任务执行完成后触发 consolidation
    """

    def __init__(self, working_memory: WorkingMemory = None,
                 memory_store: MemoryStore = None,
                 history_jsonl: Any = None):
        self.working_memory = working_memory or WorkingMemory()
        self.tasks: dict[str, Task] = {}
        self.llm = llm
        self.memory_store = memory_store  # 新增：长期记忆注入
        self.history_jsonl = history_jsonl  # 新增：history.jsonl 注入

    def _build_plan_prompt(self, user_query: str) -> str:
        """构建含长期记忆感知的任务分解提示词。"""
        memory_context = ""
        if self.memory_store:
            long_term = self.memory_store.read_long_term()
            if long_term.strip():
                memory_context = f"\n用户已知信息：\n{long_term[:1500]}\n"

        return f"""你是一个任务规划器。请将以下用户请求分解为可执行的子任务列表。

用户请求: {user_query}{memory_context}

请以 JSON 格式返回任务列表，格式如下:
{{
    "tasks": [
        {{
            "task_id": "task_1",
            "name": "简短任务名",
            "description": "具体执行描述",
            "depends_on": [],
            "tool_name": null
        }},
        ...
    ]
}}

要求:
1. 每个任务应该是独立的、可执行的步骤
2. 明确标识任务间的依赖关系
3. 任务名简短明确
4. 工具名必须来自可用工具列表"""

    def plan(self, user_query: str, context: Optional[dict] = None) -> list[Task]:
        """
        根据用户请求生成任务计划
        返回任务列表（按执行顺序）
        """
        logger.info(f"[TaskPlanner] 开始规划任务: {user_query[:50]}...")

        plan_prompt = self._build_plan_prompt(user_query)

        try:
            response = self.llm.invoke([
                {"role": "system", "content": plan_prompt},
                {"role": "user", "content": user_query}
            ])

            content = response.content if hasattr(response, 'content') else str(response)
            tasks = self._parse_plan_response(content)
            return tasks

        except Exception as e:
            logger.error(f"[TaskPlanner] 规划失败: {e}")
            # fallback: 返回一个兜底任务
            return [Task("task_1", "处理用户请求", user_query)]

    def _parse_plan_response(self, response: str) -> list[Task]:
        """解析LLM返回的任务计划"""
        from agent.utils.json_parser import extract_json

        data = extract_json(response)
        if data and isinstance(data, dict):
            tasks_data = data.get("tasks", [])
            if tasks_data:
                tasks = []
                for t_data in tasks_data:
                    task = Task(
                        task_id=t_data.get("task_id", f"task_{len(tasks)+1}"),
                        name=t_data.get("name", "未知任务"),
                        description=t_data.get("description", ""),
                        depends_on=t_data.get("depends_on", []),
                        tool_name=t_data.get("tool_name"),
                    )
                    tasks.append(task)
                    self.tasks[task.task_id] = task
                return tasks

        # fallback: 返回一个兜底任务
        return [Task("task_1", "处理用户请求", response)]

    def get_execution_order(self) -> list[Task]:
        """
        拓扑排序：返回按依赖关系排序的任务列表
        """
        # Kahn 算法
        in_degree = {}
        for task_id, task in self.tasks.items():
            in_degree[task_id] = len(task.depends_on)

        queue = [tid for tid, deg in in_degree.items() if deg == 0]
        ordered = []

        while queue:
            tid = queue.pop(0)
            ordered.append(self.tasks[tid])
            # 简单的依赖传播：检查所有任务的依赖
            for task_id, task in self.tasks.items():
                if tid in task.depends_on:
                    in_degree[task_id] = in_degree.get(task_id, 1) - 1
                    if in_degree[task_id] == 0:
                        queue.append(task_id)

        return ordered

    def update_task_status(self, task_id: str, status: str,
                           result: str = None, error: str = None) -> None:
        """更新任务状态"""
        if task_id in self.tasks:
            self.tasks[task_id].status = status
            if result:
                self.tasks[task_id].result = result
            if error:
                self.tasks[task_id].error = error
            if status in ("completed", "failed"):
                self.tasks[task_id].completed_at = datetime.now().isoformat()
                self.working_memory.set(f"task_{task_id}_result", result or error)

    def get_plan_summary(self) -> str:
        """获取计划摘要文本"""
        ordered = self.get_execution_order()
        lines = [f"任务计划（共 {len(ordered)} 个步骤）:"]
        for i, task in enumerate(ordered, 1):
            status_icon = {
                "pending": "⏳",
                "running": "▶️",
                "completed": "✅",
                "failed": "❌",
                "skipped": "⏭️",
            }.get(task.status, "⏳")
            deps = f" [依赖: {', '.join(task.depends_on)}]" if task.depends_on else ""
            lines.append(f"  {i}. {status_icon} {task.name}{deps}")
        return "\n".join(lines)
