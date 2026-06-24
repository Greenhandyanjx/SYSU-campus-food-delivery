"""
Agent Core: ReAct 循环（自实现）
不依赖 LangChain Agent 框架，自实现 Think → Act → Observe 闭环
"""
import json
from typing import Optional, Generator, AsyncGenerator
from datetime import datetime
from utils.logger import logger
from model.factory import chat_model
from agent.core.tool_registry import get_tool, get_all_tools, get_tool_schemas
from agent.core.function_calling import parse_function_call
from agent.core.schema_validator import safe_tool_execute
from agent.memory.short_term import ShortTermMemory
from agent.memory.working_memory import WorkingMemory
from utils.prompt_loader import load_system_prompts


class ReactLoop:
    """
    ReAct 推理循环
    流程：Think（推理）→ Act（工具调用）→ Observe（观察结果）→ 循环直到信息足够
    """

    def __init__(self, max_iterations: int = 10, system_prompt: Optional[str] = None):
        self.max_iterations = max_iterations
        self.system_prompt = system_prompt or load_system_prompts()
        self.short_term_memory = ShortTermMemory(max_messages=50)
        self.working_memory = WorkingMemory()
        self.model = chat_model

    def _build_messages(self, user_query: str) -> list[dict]:
        """构建发送给模型的消息列表"""
        messages = [{"role": "system", "content": self.system_prompt}]

        # 添加工具描述到 system prompt
        tool_schemas = get_tool_schemas()
        if tool_schemas:
            messages.append({
                "role": "system",
                "content": f"你拥有以下工具可用:\n{json.dumps(tool_schemas, ensure_ascii=False, indent=2)}"
            })

        # 添加短期记忆（历史会话）
        history = self.short_term_memory.get_recent()
        messages.extend(history)

        # 添加用户当前问题
        messages.append({"role": "user", "content": user_query})

        return messages

    def _call_model(self, messages: list[dict]) -> str:
        """调用 LLM 获取回复"""
        try:
            response = self.model.invoke(messages)
            content = response.content if hasattr(response, 'content') else str(response)
            return content
        except Exception as e:
            logger.error(f"[ReAct] 模型调用失败: {str(e)}")
            return f"模型调用出错: {str(e)}"

    def _is_final_answer(self, response: str) -> bool:
        """判断模型是否给出了最终回复（没有调用工具）"""
        # 如果回复包含明确的工具调用标记，说明还需要继续
        fc = parse_function_call(response)
        if fc:
            return False
        return True

    def execute(self, user_query: str) -> str:
        """
        同步执行 ReAct 循环
        返回最终回复
        """
        messages = self._build_messages(user_query)
        logger.info(f"[ReAct] 开始处理用户查询: {user_query[:50]}...")

        for iteration in range(self.max_iterations):
            logger.info(f"[ReAct] 第 {iteration + 1}/{self.max_iterations} 轮推理")

            # === Think 阶段 ===
            response = self._call_model(messages)

            # 检查是否是 Function Call
            fc = parse_function_call(response)

            if fc is None:
                # === Final Answer 阶段 ===
                logger.info(f"[ReAct] 模型给出最终回复，循环结束")
                self.short_term_memory.add("user", user_query)
                self.short_term_memory.add("assistant", response)
                return response

            # === Act 阶段 ===
            tool = get_tool(fc.name)
            if tool is None:
                error_msg = f"工具 '{fc.name}' 不存在"
                logger.warning(f"[ReAct] {error_msg}, 可用工具: {[t.name for t in get_all_tools()]}")
                messages.append({"role": "assistant", "content": response})
                messages.append({"role": "tool", "content": error_msg, "tool_call_id": fc.name})
                continue

            logger.info(f"[ReAct] 调用工具: {fc.name}, 参数: {fc.arguments}")

            # === Observe 阶段 ===
            success, result, error = safe_tool_execute(tool, fc.arguments)

            observation = ""
            if success:
                observation = f"工具 '{fc.name}' 返回结果: {result}"
                # 保存到工作记忆
                self.working_memory.set(f"tool_result_{fc.name}", result)
            else:
                observation = f"工具 '{fc.name}' 执行失败: {error}"

            logger.info(f"[ReAct] 工具返回: {observation[:100]}...")

            # 将工具调用和结果追加到消息历史
            messages.append({"role": "assistant", "content": response})
            messages.append({"role": "tool", "content": observation, "tool_call_id": fc.name})

        # 达到最大迭代次数
        fallback_msg = ("我已经进行了多次尝试，但仍无法完全回答您的问题。"
                        "请尝试更具体地描述您的需求，或者换个角度提问。")
        self.short_term_memory.add("user", user_query)
        self.short_term_memory.add("assistant", fallback_msg)
        return fallback_msg

    def execute_stream(self, user_query: str) -> Generator[str, None, str]:
        """
        流式执行 ReAct 循环
        每次 yield 一个文本块，最终返回完整回复
        """
        messages = self._build_messages(user_query)
        logger.info(f"[ReAct] 开始流式处理: {user_query[:50]}...")

        for iteration in range(self.max_iterations):
            yield f"\n[推理步骤 {iteration + 1}]\n"

            response = self._call_model(messages)
            fc = parse_function_call(response)

            if fc is None:
                # 最终回复
                yield response
                self.short_term_memory.add("user", user_query)
                self.short_term_memory.add("assistant", response)
                return response

            # 执行工具
            yield f"[调用工具: {fc.name}]\n"

            tool = get_tool(fc.name)
            if tool is None:
                error_msg = f"工具 '{fc.name}' 不存在"
                yield error_msg
                messages.append({"role": "assistant", "content": response})
                messages.append({"role": "tool", "content": error_msg, "tool_call_id": fc.name})
                continue

            success, result, error = safe_tool_execute(tool, fc.arguments)
            observation = f"工具 '{fc.name}' 返回结果: {result}" if success else f"工具 '{fc.name}' 执行失败: {error}"

            if success:
                self.working_memory.set(f"tool_result_{fc.name}", result)

            yield f"[工具结果: {observation[:80]}...]\n"

            messages.append({"role": "assistant", "content": response})
            messages.append({"role": "tool", "content": observation, "tool_call_id": fc.name})

        fallback = "已达到最大推理轮数，请尝试换个角度提问。"
        yield fallback
        self.short_term_memory.add("user", user_query)
        self.short_term_memory.add("assistant", fallback)
        return fallback
