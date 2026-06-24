"""
Agent Core: Function Calling 解析器
自实现模型输出 → 工具调用参数的解析
"""
import json
import re
from typing import Optional
from utils.logger import logger


class FunctionCall:
    """解析后的函数调用"""

    def __init__(self, name: str, arguments: dict):
        self.name = name
        self.arguments = arguments

    def __repr__(self):
        return f"FunctionCall(name={self.name}, args={self.arguments})"


def parse_function_call(model_response: str) -> Optional[FunctionCall]:
    """
    解析模型输出中的 Function Call 结构。
    支持多种格式：
    1. 标准 JSON 格式：{"name": "xxx", "arguments": {...}}
    2. XML/标签格式：<function_call>{"name":"xxx","arguments":{}}</function_call>
    3. 自然语言中嵌入的 JSON
    4. 支持通义千问格式（在content中返回tool_calls）
    """
    # 方法1: 尝试直接解析为 JSON
    text = model_response.strip()

    # 尝试提取 JSON 块
    json_pattern = r'```(?:json)?\s*([\s\S]*?)```'
    json_matches = re.findall(json_pattern, text)

    for json_str in json_matches:
        result = _try_parse_json(json_str.strip(), is_tool_call=True)
        if result:
            return result

    # 方法2: 查找 {...} 中是否有 function_call 结构
    brace_pattern = r'\{[^{}]*\}'
    brace_matches = re.findall(brace_pattern, text)

    for possible_json in brace_matches:
        result = _try_parse_json(possible_json, is_tool_call=True)
        if result:
            return result

    # 方法3: 查找 XML 风格的 function_call 标签
    xml_pattern = r'<function_call>\s*(\{[\s\S]*?\})\s*</function_call>'
    xml_matches = re.findall(xml_pattern, text)

    for xml_json in xml_matches:
        result = _try_parse_json(xml_json.strip(), is_tool_call=True)
        if result:
            return result

    # 方法4: 查找 tool_call 格式
    tool_pattern = r'<tool_call>\s*(\{[\s\S]*?\})\s*</tool_call>'
    tool_matches = re.findall(tool_pattern, text)

    for tool_json in tool_matches:
        result = _try_parse_json(tool_json.strip(), is_tool_call=True)
        if result:
            return result

    # 方法5: 统一 fallback — 用 extract_json 的完整回退链提取
    from agent.utils.json_parser import extract_json
    data = extract_json(text)
    if isinstance(data, dict):
        result = _try_parse_json(json.dumps(data, ensure_ascii=False), is_tool_call=True)
        if result:
            return result

    return None


def _try_parse_json(text: str, is_tool_call: bool = False) -> Optional[FunctionCall]:
    """尝试解析 JSON，支持 tool_call 格式和标准 function_call 格式"""
    try:
        data = json.loads(text)

        if is_tool_call:
            # 标准 Function Call 格式
            if "name" in data and "arguments" in data:
                args = data["arguments"]
                if isinstance(args, str):
                    args = json.loads(args)
                return FunctionCall(data["name"], args)

            # { "function": { "name": "...", "arguments": {...} } }
            if "function" in data:
                func_data = data["function"]
                if isinstance(func_data.get("arguments"), str):
                    func_data["arguments"] = json.loads(func_data["arguments"])
                return FunctionCall(func_data["name"], func_data["arguments"])

            # { "tool": "xxx", "args": {...} }
            if "tool" in data:
                return FunctionCall(data["tool"], data.get("args", {}))

        return None
    except (json.JSONDecodeError, KeyError):
        return None


def parse_function_calls_from_messages(messages: list[dict]) -> list[FunctionCall]:
    """
    从消息列表中提取所有 Function Call（兼容通义千问的 tool_calls 格式）
    """
    calls = []

    for msg in messages:
        msg_content = msg.get("content", "")

        # 检查 tool_calls 字段（通义千问格式）
        tool_calls = msg.get("tool_calls", [])
        for tc in tool_calls:
            if tc.get("type") == "function":
                func = tc.get("function", {})
                name = func.get("name", "")
                try:
                    args = json.loads(func.get("arguments", "{}"))
                except json.JSONDecodeError:
                    args = func.get("arguments", {})
                calls.append(FunctionCall(name, args))

        # 内容中也可能包含 Function Call
        if isinstance(msg_content, str) and msg_content:
            parsed = parse_function_call(msg_content)
            if parsed:
                calls.append(parsed)

    return calls
