"""
Agent Core: 工具注册中心
自实现工具注册、Schema生成、调度分发，不依赖LangChain框架
"""
from typing import Callable, Dict, List, Optional, Any, get_type_hints
import inspect
import json
from datetime import datetime
from utils.logger import logger


class ToolSchema:
    """工具 Schema：描述工具的输入输出格式"""

    def __init__(self, name: str, description: str, parameters: dict, required: list[str]):
        self.name = name
        self.description = description
        self.parameters = parameters  # JSON Schema 格式
        self.required = required

    def to_openai_format(self) -> dict:
        """转为 OpenAI Function Calling 格式"""
        return {
            "type": "function",
            "function": {
                "name": self.name,
                "description": self.description,
                "parameters": {
                    "type": "object",
                    "properties": self.parameters,
                    "required": self.required,
                }
            }
        }

    def to_dict(self) -> dict:
        return {
            "name": self.name,
            "description": self.description,
            "parameters": self.parameters,
            "required": self.required,
        }


class Tool:
    """工具封装：每个工具是一个可调用对象 + Schema"""

    def __init__(self, name: str, description: str, fn: Callable,
                 parameters: Optional[dict] = None,
                 required: Optional[list[str]] = None):
        self.name = name
        self.description = description
        self.fn = fn
        self.schema = ToolSchema(name, description, parameters or {}, required or [])

    async def execute(self, **kwargs) -> Any:
        """执行工具，支持异步"""
        logger.info(f"[Tool] 执行工具: {self.name}, 参数: {kwargs}")
        try:
            if inspect.iscoroutinefunction(self.fn):
                result = await self.fn(**kwargs)
            else:
                result = self.fn(**kwargs)
            logger.info(f"[Tool] 工具 {self.name} 执行成功")
            return result
        except Exception as e:
            logger.error(f"[Tool] 工具 {self.name} 执行失败: {str(e)}")
            raise

    def validate_args(self, args: dict) -> tuple[bool, str]:
        """校验参数是否符合 Schema"""
        for required_param in self.schema.required:
            if required_param not in args:
                return False, f"缺少必要参数: {required_param}"

        for param_name, param_value in args.items():
            if param_name not in self.schema.parameters:
                return False, f"未知参数: {param_name}"

            param_schema = self.schema.parameters[param_name]
            param_type = param_schema.get("type", "string")

            # 基本类型校验
            if param_type == "string" and not isinstance(param_value, str):
                return False, f"参数 {param_name} 应为 string 类型"
            elif param_type == "integer" and not isinstance(param_value, int):
                return False, f"参数 {param_name} 应为 integer 类型"
            elif param_type == "number" and not isinstance(param_value, (int, float)):
                return False, f"参数 {param_name} 应为 number 类型"

        return True, ""


# 内置装饰器，用于标注工具函数
_tool_registry: Dict[str, Tool] = {}


def tool(name: str = None, description: str = None):
    """工具装饰器：将函数注册为 Agent 工具"""
    def decorator(func):
        nonlocal name, description
        tool_name = name or func.__name__
        tool_desc = description or func.__doc__ or ""

        # 自动推断参数 Schema
        sig = inspect.signature(func)
        parameters = {}
        required = []

        for param_name, param in sig.parameters.items():
            param_type = str if param.annotation is inspect.Parameter.empty else param.annotation
            json_type = _python_type_to_json_type(param_type)
            param_desc = ""
            parameters[param_name] = {
                "type": json_type,
                "description": param_desc,
            }
            if param.default is inspect.Parameter.empty:
                required.append(param_name)

        tool_obj = Tool(tool_name, tool_desc, func, parameters, required)
        _tool_registry[tool_name] = tool_obj
        logger.info(f"[ToolRegistry] 注册工具: {tool_name}")
        return func

    return decorator


def _python_type_to_json_type(py_type) -> str:
    """Python类型 → JSON Schema 类型映射"""
    type_map = {
        str: "string",
        int: "integer",
        float: "number",
        bool: "boolean",
        list: "array",
        dict: "object",
    }
    return type_map.get(py_type, "string")


def get_tool(name: str) -> Optional[Tool]:
    """根据名称获取工具"""
    return _tool_registry.get(name)


def get_all_tools() -> list[Tool]:
    """获取所有注册的工具"""
    return list(_tool_registry.values())


def get_tool_schemas() -> list[dict]:
    """获取所有工具的 Schema（OpenAI Format）"""
    return [tool.schema.to_openai_format() for tool in _tool_registry.values()]


def get_tool_descriptions() -> str:
    """获取工具描述的纯文本格式（用于非FunctionCalling场景）"""
    lines = []
    for tool_obj in _tool_registry.values():
        params_desc = []
        for p_name, p_schema in tool_obj.schema.parameters.items():
            is_required = "（必填）" if p_name in tool_obj.schema.required else "（可选）"
            params_desc.append(f"    - {p_name} ({p_schema['type']}){is_required}")
        params_str = "\n".join(params_desc)
        lines.append(f"- {tool_obj.name}: {tool_obj.description}\n{params_str}")
    return "\n".join(lines)


def clear_registry():
    """清空工具注册（测试用）"""
    _tool_registry.clear()
