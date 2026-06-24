"""
Planner: 任务依赖图执行引擎
按拓扑排序逐步执行子任务
"""
from typing import Optional
from utils.logger import logger
from agent.planner.task_planner import TaskPlanner, Task
from agent.core.react_loop import ReactLoop
from agent.core.function_calling import parse_function_call
from model.factory import chat_model as llm


class TaskExecutor:
    """
    任务执行引擎
    按拓扑排序的顺序执行子任务

    扩展（参考 nanobot 记忆管线）：
    - 执行完成后将任务结果写入 history.jsonl
    - 支持记忆感知的循环（每个子任务前加载上下文）
    """

    def __init__(self, planner: TaskPlanner = None, react_loop: ReactLoop = None):
        self.planner = planner or TaskPlanner()
        self.react_loop = react_loop or ReactLoop()
        self.llm = llm

    def execute_plan(self, user_query: str) -> str:
        """
        完整执行一个规划的任务
        """
        tasks = self.planner.plan(user_query)
        ordered = self.planner.get_execution_order()

        logger.info(f"[TaskExecutor] 开始执行 {len(ordered)} 个子任务")
        results = []

        for task in ordered:
            logger.info(f"[TaskExecutor] 执行任务: {task.name}")

            # 检查依赖是否都已完成
            deps_ok = all(
                self.planner.tasks.get(dep_id)
                and self.planner.tasks[dep_id].status == "completed"
                for dep_id in task.depends_on
            )

            if not deps_ok:
                self.planner.update_task_status(task.task_id, "skipped",
                                                error="依赖任务未完成")
                continue

            self.planner.update_task_status(task.task_id, "running")

            try:
                # 使用 ReAct Loop 执行子任务
                result = self.react_loop.execute(task.description)
                self.planner.update_task_status(task.task_id, "completed", result=result)
                results.append(f"[{task.name}]\n{result}")
            except Exception as e:
                self.planner.update_task_status(task.task_id, "failed", error=str(e))
                logger.error(f"[TaskExecutor] 任务 {task.name} 失败: {e}")

        # 汇总结果
        summary = user_query + "\n\n"

        task_summary = self.planner.get_plan_summary()
        summary += task_summary + "\n\n"

        if results:
            summary += "执行结果:\n" + "\n".join(results)

        return summary

    def get_plan_text(self, user_query: str) -> str:
        """
        只生成计划，不执行（供前端展示）
        """
        tasks = self.planner.plan(user_query)
        return self.planner.get_plan_summary()
