"""
Agent Core: 工具基类
参考 nanobot 的 agent/tools/base.py 设计

Tool 是 Agent 可以调用的能力接口。
每个 Tool 需要定义：
- name: 工具名称
- description: 功能描述（LLM 用来决定何时调用）
- parameters: JSON Schema 参数定义
- execute(): 执行逻辑（异步）
"""

from abc import ABC, abstractmethod
from typing import Any


class Tool(ABC):
    """
    工具抽象基类。
    
    所有 Agent 工具（读取文件、执行命令、搜索 web 等）
    都继承此类并实现抽象方法。
    """

    # 类型映射：JSON Schema type → Python type
    _TYPE_MAP = {
        "string": str,
        "integer": int,
        "number": (int, float),
        "boolean": bool,
        "array": list,
        "object": dict,
    }

    @property
    @abstractmethod
    def name(self) -> str:
        """工具名称（用于 Function Calling 中的函数名）"""
        pass

    @property
    @abstractmethod
    def description(self) -> str:
        """工具描述（LLM 理解工具用途的关键）"""
        pass

    @property
    @abstractmethod
    def parameters(self) -> dict[str, Any]:
        """JSON Schema 格式的参数定义"""
        pass

    @abstractmethod
    async def execute(self, **kwargs: Any) -> Any:
        """
        执行工具逻辑。
        
        Args:
            **kwargs: 由 LLM 解析的工具参数
        
        Returns:
            字符串或内容块列表
        """
        pass

    def to_schema(self) -> dict[str, Any]:
        """
        Convert to OpenAI Function Calling format Schema.
        Strips inner "required" flags from property schemas.
        """
        raw_params = self.parameters or {}
        properties = {}
        required = []
        for name, schema in raw_params.items():
            if isinstance(schema, dict):
                is_required = schema.get("required", False)
                if is_required:
                    required.append(name)
                # Remove inner "required" from property — only valid at top level
                prop = {k: v for k, v in schema.items() if k != "required"}
            else:
                prop = schema
            properties[name] = prop
        parameters = {
            "type": "object",
            "properties": properties,
        }
        if required:
            parameters["required"] = required
        return {
            "type": "function",
            "function": {
                "name": self.name,
                "description": self.description,
                "parameters": parameters,
            },
        }

    def cast_params(self, params: dict[str, Any]) -> dict[str, Any]:
        """
        根据 Schema 类型强制转换参数（提高 LLM 输出兼容性）。
        
        LLM 有时会传字符串类型的数字，这里尝试自动转换。
        """
        schema = self.parameters or {}
        result = dict(params)
        for key, value in params.items():
            if key in schema and schema[key].get("type") == "number":
                if isinstance(value, str):
                    try:
                        result[key] = float(value)
                    except (ValueError, TypeError):
                        pass
            elif key in schema and schema[key].get("type") == "integer":
                if isinstance(value, str):
                    try:
                        result[key] = int(value)
                    except (ValueError, TypeError):
                        pass
        return result

    def validate_params(self, params: dict[str, Any]) -> list[str]:
        """校验参数是否符合 Schema 定义。返回错误列表"""
        errors = []
        schema = self.parameters or {}

        for name, field_schema in schema.items():
            if field_schema.get("required", False) and name not in params:
                errors.append(f"缺少必填参数: {name}")

        for name, value in params.items():
            if name not in schema:
                continue
            expected_type = schema[name].get("type", "string")
            py_type = self._TYPE_MAP.get(expected_type)
            if py_type and not isinstance(value, py_type):
                errors.append(
                    f"参数 '{name}' 类型错误: 期望 {expected_type}, "
                    f"实际 {type(value).__name__}"
                )

        return errors
