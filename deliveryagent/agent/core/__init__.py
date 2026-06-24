"""Agent Core 模块：ReAct 循环核心、LLM Provider 抽象、消息类型定义

参考 nanobot 架构重新设计：

模块           | 功能                    | 对应 nanobot
--------------|------------------------|-------------------
types.py      | 核心数据类（DTO）       | bus/events.py
llm_provider.py | LLM Provider 抽象接口 | providers/base.py
tools/base.py | 工具基类                | agent/tools/base.py
tools/registry.py | 工具注册中心          | agent/tools/registry.py

保留兼容：
- function_calling.py（原有解析器，作为备用格式支持）
- tool_registry.py（原有的装饰器式注册，保持向后兼容）
- schema_validator.py（参数校验逻辑保留）
- fallback.py（错误回退策略保留）
"""
