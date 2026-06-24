"""
Agent Core: Schema 校验器
校验工具调用的参数是否符合 Schema 定义
"""
import json
from typing import Any, Optional
from utils.logger import logger
from agent.core.tool_registry import Tool


class ValidationError(Exception):
    """参数校验失败"""
    pass


def validate_tool_call(tool: Tool, args: dict) -> None:
    """
    校验工具调用的参数
    抛出 ValidationError 如果校验失败
    """
    # 1. 检查必填参数
    for required_param in tool.schema.required:
        if required_param not in args:
            raise ValidationError(
                f"工具 '{tool.name}' 缺少必填参数: {required_param}"
            )

    # 2. 检查是否有未知参数
    for param_name in args:
        if param_name not in tool.schema.parameters:
            logger.warning(f"[SchemaValidator] 工具 '{tool.name}' 收到未知参数: {param_name}")

    # 3. 类型校验
    for param_name, param_value in args.items():
        if param_name not in tool.schema.parameters:
            continue

        param_schema = tool.schema.parameters[param_name]
        expected_type = param_schema.get("type", "string")

        type_ok = _check_type(param_value, expected_type)
        if not type_ok:
            raise ValidationError(
                f"参数 '{param_name}' 类型错误: 期望 {expected_type}, "
                f"实际 {type(param_value).__name__} (值: {param_value})"
            )


def _check_type(value: Any, expected_type: str) -> bool:
    """检查值是否符合期望类型"""
    type_map = {
        "string": lambda v: isinstance(v, str),
        "integer": lambda v: isinstance(v, int) and not isinstance(v, bool),
        "number": lambda v: isinstance(v, (int, float)) and not isinstance(v, bool),
        "boolean": lambda v: isinstance(v, bool),
        "array": lambda v: isinstance(v, (list, tuple)),
        "object": lambda v: isinstance(v, dict),
    }
    checker = type_map.get(expected_type)
    if checker is None:
        return True  # 未知类型跳过
    return checker(value)


def safe_tool_execute(tool: Tool, args: dict, max_retries: int = 3) -> tuple[bool, Any, str]:
    """
    安全执行工具：校验 + 重试 + 错误处理
    返回: (是否成功, 结果, 错误信息)
    """
    # 参数校验
    try:
        validate_tool_call(tool, args)
    except ValidationError as e:
        return False, None, str(e)

    # 执行（带重试）
    last_error = ""
    for attempt in range(max_retries):
        try:
            result = tool.execute(**args)
            return True, result, ""
        except Exception as e:
            last_error = str(e)
            logger.warning(f"[SafeExecute] 工具 '{tool.name}' 第 {attempt+1}/{max_retries} 次重试失败: {last_error}")
            if attempt < max_retries - 1:
                import time
                wait_time = 2 ** attempt  # 指数退避
                time.sleep(wait_time)

    return False, None, f"工具 '{tool.name}' 执行失败（已重试 {max_retries} 次）: {last_error}"
