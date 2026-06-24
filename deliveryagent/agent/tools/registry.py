"""
Agent Core: 工具注册中心
参考 nanobot 的 agent/tools/registry.py 设计

管理所有 Agent 工具的注册、查询和执行调度。
"""

from typing import Any

from loguru import logger
from agent.tools.base import Tool


class ToolRegistry:
    """
    工具注册中心。
    
    功能：
    - register() / unregister(): 动态注册/注销工具
    - get_definitions(): 获取所有工具的 OpenAI 格式定义
    - execute(): 执行指定工具（含参数校验和类型转换）
    """

    def __init__(self):
        self._tools: dict[str, Tool] = {}

    def register(self, tool: Tool) -> None:
        """注册一个工具"""
        self._tools[tool.name] = tool
        logger.info(f"[ToolRegistry] 注册工具: {tool.name}")

    def unregister(self, name: str) -> None:
        """注销一个工具"""
        self._tools.pop(name, None)
        logger.info(f"[ToolRegistry] 注销工具: {name}")

    def get(self, name: str) -> Tool | None:
        """根据名称获取工具"""
        return self._tools.get(name)

    def has(self, name: str) -> bool:
        """检查工具是否存在"""
        return name in self._tools

    def get_definitions(self) -> list[dict[str, Any]]:
        """获取所有工具的 OpenAI Function Calling 格式定义"""
        return [tool.to_schema() for tool in self._tools.values()]

    async def execute(self, name: str, params: dict[str, Any]) -> Any:
        """
        执行指定工具。
        
        流程：
        1. 查找工具
        2. 参数类型转换
        3. 参数校验
        4. 执行
        5. 返回结果

        如果失败，返回带有 _HINT 的错误信息（提示 LLM 换种方式重试）
        """
        _HINT = "\n\n[分析上面的错误后尝试不同的方式。]"

        tool = self._tools.get(name)
        if not tool:
            return f"错误: 工具 '{name}' 不存在。可用工具: {', '.join(self._tools.keys())}"

        try:
            # 尝试类型转换
            params = tool.cast_params(params)

            # 参数校验
            errors = tool.validate_params(params)
            if errors:
                return f"错误: 工具 '{name}' 参数无效: {'; '.join(errors)}" + _HINT

            # 执行
            result = await tool.execute(**params)
            if isinstance(result, str) and result.startswith("Error"):
                return result + _HINT
            return result

        except Exception as e:
            return f"错误执行 {name}: {str(e)}" + _HINT

    @property
    def tool_names(self) -> list[str]:
        return list(self._tools.keys())

    def __len__(self) -> int:
        return len(self._tools)

    def __contains__(self, name: str) -> bool:
        return name in self._tools
