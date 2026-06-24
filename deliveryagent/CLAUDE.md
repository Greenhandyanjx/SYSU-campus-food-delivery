# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 启动方式

```bash
# 终端 1：FastAPI 后端（必须）
uvicorn api.app:app --host 127.0.0.1 --port 8000 --reload

# 终端 2：Streamlit 前端
streamlit run main.py --server.port 8501
```

或一键启动：`start.bat`（Windows）/ `start.sh`（WSL）

## 架构概览

```
Streamlit(8501) --HTTP--> FastAPI(8000) --> AgentOrchestrator --> AgentLoop(ReAct)
```

- **前/后端分离** — Streamlit 只做 UI，不直接实例化 Agent
- **AgentOrchestrator** — 全局单例，编排 MessageBus + AgentLoop + Provider + Tools + Skills
- **AgentLoop** — ReAct 核心 (Think→Act→Observe)，per-session 锁保证同会话串行
- **Memory** — 三层：Session（短期）→ history.jsonl / Consolidator（中期）→ MEMORY.md / Dream（长期）
- **Skills** — 自动发现 `skills/` 目录下的插件，通过意图匹配触发
- **工具** — 通过 ToolRegistry 注册，支持并行执行

## 关键目录

| 目录 | 用途 |
|------|------|
| `agent/` | 核心 Agent 框架（loop, bus, tools, memory, session, providers, skills, planner） |
| `api/` | FastAPI 路由 + 入口 |
| `rag/` | RAG 检索增强（查询改写、多路召回、重排序） |
| `skills/` | 技能插件（weather, datetime, rag, template） |
| `config/` | YAML 配置（agent, vector_store, rag, prompts） |
| `sessions/` | 会话持久化 JSONL |
| `memory/` | MEMORY.md / HISTORY.md / history.jsonl |

## 关键约定

- 你是一个企业的ai应用开发高级工程师，在辅助我编写agent代码并给出关键帮助理解的注释、帮我提出多种工程实现计划让我做出选择，以帮我培养企业级的工程思想和增加我面试agent开发的全栈岗位通过的成功率
- LLM Provider 通过 OpenAIProvider 实现（兼容 DeepSeek / 通义千问），抽象接口在 `agent/core/llm_provider.py`
- Session 持久化为 JSONL，每条消息含 role/content/timestamp，tool_calls 单独存字段
- assistant 消息的 content 字段必须为字符串（不允许 None/null），避免 LLM 渲染异常
- 流式聊天走 SSE，非流式走 JSON，路由在 `api/routes.py`
- 环境变量配置参考 `.env.example`

## 环境变量

- `DEEPSEEK_API_KEY` — LLM API Key（也可用 `OPENAI_API_KEY` 或 `DASHSCOPE_API_KEY`）
- `DEFAULT_MODEL` — 默认模型（deepseek-chat）
- `JDAGENT_MAX_CONCURRENT_REQUESTS` — 并发数（默认 3）
