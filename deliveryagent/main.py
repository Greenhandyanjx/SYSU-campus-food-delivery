"""
Streamlit 前端页面：JD-Agent 聊天界面
==============================================
修改说明：
  v1 → v2（本版本）将 Streamlit 从**直接实例化 Orchestrator**
  改为**调用后端 FastAPI**，实现前后端分离。

架构：
  浏览器 ──HTTP──→ Streamlit (8501) ──HTTP──→ FastAPI (8000) ──→ AgentOrchestrator

为什么这样设计：
  1. 面试亮点：展示前后端分离架构
  2. 可扩展：Streamlit 只是前端之一，WebSocket / 移动端可复用同一 API
  3. 解耦：前端升级不影响 Agent 逻辑，同一份 API 服务可被多个前端调用

使用方式：
  # 终端 1：启动 API
  python -m api.app

  # 终端 2：启动 Streamlit 前端
  streamlit run main.py --server.port 8501
"""
import os
import time
from typing import Optional

import httpx
import streamlit as st
from loguru import logger

# ─────────────────────────────────────────────────────────
# 配置
# ─────────────────────────────────────────────────────────

# API 地址：默认 localhost:8000，可通过环境变量覆盖
API_BASE_URL = os.environ.get("JD_AGENT_API_URL", "http://127.0.0.1:8000")

st.set_page_config(
    page_title="JD-Agent 智能助手",
    page_icon="🤖",
    layout="wide",
)

st.title("🤖 JD-Agent 智能助手")
st.caption(
    "前后端分离架构 · Streamlit → FastAPI → AgentOrchestrator | "
    "支持流式输出 · 工具调用 · 四层记忆 · 会话持久化"
)
st.divider()


# ─────────────────────────────────────────────────────────
# 工具函数
# ─────────────────────────────────────────────────────────


@st.cache_resource
def get_api_status() -> dict:
    """
    检查后端 API 是否在线，并缓存结果。

    返回示例: {"status": "ok", "tools_count": 8, "sessions_count": 3}
    失败时: {"status": "error", "detail": "连接失败"}
    """
    try:
        resp = httpx.get(f"{API_BASE_URL}/api/v1/health", timeout=5.0)
        if resp.status_code == 200:
            return resp.json()
        return {"status": "error", "detail": f"HTTP {resp.status_code}"}
    except httpx.RequestError as e:
        logger.warning(f"[Streamlit] 后端 API 连接失败: {e}")
        return {"status": "error", "detail": str(e)}


def send_message(message: str, session_id: Optional[str] = None) -> str:
    """
    向 API 发送非流式消息，返回完整回复。

    参数:
      message:    用户输入
      session_id: 会话 ID（不传则自动生成）
    """
    payload = {"message": message, "stream": False}
    if session_id:
        payload["session_id"] = session_id
    try:
        resp = httpx.post(
            f"{API_BASE_URL}/api/v1/chat",
            json=payload,
            timeout=120.0,
        )
        resp.raise_for_status()
        data = resp.json()
        # 存下返回的 session_id 供后续使用
        if "session_id" in data:
            st.session_state.current_session_id = data["session_id"]
        return data.get("response", "")
    except httpx.RequestError as e:
        logger.error(f"[Streamlit] API 调用失败: {e}")
        return f"🚨 后端连接失败: {e}"
    except Exception as e:
        logger.error(f"[Streamlit] 请求出错: {e}")
        return f"🚨 请求出错: {e}"


def send_message_stream(message: str, response_container, session_id: Optional[str] = None) -> str:
    """
    向 API 发送流式消息（SSE），逐步渲染回复。

    流程:
      1. 发送 POST /chat 请求（stream: true）
      2. 逐行读取 SSE 流，每收到一个 chunk 就更新界面
      3. 遇到 [DONE] 标记结束

    返回完整回复文本。
    """
    payload = {"message": message, "stream": True}
    if session_id:
        payload["session_id"] = session_id
    try:
        with httpx.stream(
            "POST",
            f"{API_BASE_URL}/api/v1/chat",
            json=payload,
            timeout=120.0,
        ) as resp:
            resp.raise_for_status()
            full_text = ""
            for line in resp.iter_lines():
                if not line.startswith("data: "):
                    continue
                data_str = line[6:]  # 去掉 "data: " 前缀
                if data_str == "[DONE]":
                    break
                # 尝试 JSON 解码
                import json
                try:
                    chunk_data = json.loads(data_str)
                    text = chunk_data.get("text", "")
                except json.JSONDecodeError:
                    text = data_str
                full_text += text
                response_container.markdown(full_text + "▌")
            return full_text
    except httpx.RequestError as e:
        logger.error(f"[Streamlit] 流式请求失败: {e}")
        return f"🚨 后端连接失败: {e}"
    except Exception as e:
        logger.error(f"[Streamlit] 流式请求出错: {e}")
        return f"🚨 请求出错: {e}"


def fetch_history(session_key: Optional[str] = None) -> list[dict]:
    """
    从 API 拉取当前会话的历史消息。

    用于页面初始化时恢复对话上下文。
    """
    try:
        params = {}
        if session_key:
            params["session_key"] = session_key
        resp = httpx.get(f"{API_BASE_URL}/api/v1/history", params=params, timeout=5.0)
        if resp.status_code == 200:
            data = resp.json()
            return data.get("history", [])
    except Exception as e:
        logger.warning(f"[Streamlit] 拉取历史失败: {e}")
    return []


# ─────────────────────────────────────────────────────────
# Session State 初始化
# ─────────────────────────────────────────────────────────

if "current_session_id" not in st.session_state:
    st.session_state.current_session_id = "cli:direct"  # 固定会话 key，读写与 CLI 一致

if "messages" not in st.session_state:
    st.session_state.messages = []              # 前端展示用消息列表

# ── 启动时尝试拉取历史（从 PG 读取，JSONL 保底） ──
if "history_loaded" not in st.session_state:
    history = fetch_history(session_key=st.session_state.current_session_id)
    if history:
        for msg in history:
            role = msg.get("role", "")
            # 跳过工具调用专用的 assistant 消息（content 空 + 有 tool_calls）
            if role == "assistant" and not msg.get("content") and msg.get("tool_calls"):
                continue
            if role in ("user", "assistant"):
                # content 可能为 None/Pydantic null，统一转为空字符串
                content = msg.get("content") or ""
                st.session_state.messages.append({
                    "role": role,
                    "content": content,
                })
    st.session_state.history_loaded = True


# ─────────────────────────────────────────────────────────
# 侧边栏
# ─────────────────────────────────────────────────────────

with st.sidebar:
    st.header("🔧 系统状态")

    # ── 后端状态 ──
    status = get_api_status()
    if status.get("status") == "ok":
        tools_n = status.get('tools_count', 0)
        skills_n = status.get('skills_count', 0)
        skills_names = status.get('skills_names', [])
        skill_badge = f" | 技能 {skills_n} 个" if skills_n > 0 else ""
        st.success(f"✅ API 在线 | 工具 {tools_n} 个{skill_badge}")
        # 技能标签展示
        if skills_names:
            st.markdown(" ".join([f"🏷️ {s}" for s in skills_names]))
    else:
        st.error(f"❌ API 离线: {status.get('detail', '连接超时')}")
        st.info("请先在终端启动 API:  python -m api.app")
        st.stop()  # 后端离线时阻止聊天

    # ── 会话信息 ──
    sid = st.session_state.current_session_id
    st.metric("会话 ID", sid or "待初始化")
    st.metric("消息数", len(st.session_state.messages))

    # ── 按钮区 ──
    col1, col2 = st.columns(2)
    with col1:
        if st.button("🔄 清空对话", use_container_width=True):
            st.session_state.messages = []
            st.session_state.current_session_id = None
            st.rerun()
    with col2:
        if st.button("📋 复制会话 ID", use_container_width=True):
            if st.session_state.current_session_id:
                st.info(f"已复制: {st.session_state.current_session_id}")
            else:
                st.info("尚无会话 ID")

    # ── 技能系统说明 ──
    with st.expander("🎯 技能系统（Skill System）"):
        if status.get('skills_count', 0) > 0:
            st.markdown(f"""
            ### 已加载技能
            
            共有 **{skills_n}** 个技能包被自动发现并加载：
            """)
            for sname in skills_names:
                st.markdown(f"- 🎯 **{sname}**")
        else:
            st.info("暂无 Skill 技能包被加载。")
        st.markdown("""
        ---
        技能系统特性：
        - 📂 **自动发现** — 扫描 `skills/` 目录
        - 📄 **SKILL.md** — 每个技能带文档声明
        - 🧠 **意图匹配** — 根据用户输入智能匹配合适技能
        - 🔌 **可插拔** — 新增/删除技能目录即可热更新
        """)

    # ── 架构说明 ──
    with st.expander("📖 架构说明"):
        st.markdown("""
        ### JD-Agent 架构（前后端分离 + 技能系统）

        ```
        ┌─────────────────┐     HTTP/SSE     ┌──────────────────┐
        │  Streamlit UI    │ ───────────────→ │  FastAPI Server   │
        │  (port 8501)     │ ←────────────── │  (port 8000)      │
        └─────────────────┘                  └────────┬─────────┘
                                                       │
                                              ┌────────▼─────────┐
                                              │ AgentOrchestrator │
                                              │  ┌───────────┐   │
                                              │  │ Skill     │   │
                                              │  │ Manager   │──→│ ←skills/ 目录
                                              │  └───────────┘   │
                                              └──────────────────┘
        ```

        | 组件 | 端口 | 职责 |
        |------|------|------|
        | Streamlit UI | 8501 | 聊天界面 &rarr; 调后端 API |
        | FastAPI | 8000 | REST & SSE 端点 |
        | Agent | 内部 | ReAct 循环 + 工具 + 记忆 + **技能系统** |
        """)

    # ── 环境变量说明 ──
    with st.expander("🌐 环境变量"):
        st.markdown(f"""
        - `DEEPSEEK_API_KEY` — DeepSeek API Key
        - `OPENAI_API_KEY`   — OpenAI API Key
        - `DEFAULT_MODEL`    — 默认模型（如 deepseek-chat）
        - `JD_AGENT_API_URL` — API 后端地址
          当前值: `{API_BASE_URL}`
        - `API_PORT`         — API 端口（默认 8000）
        """)


# ─────────────────────────────────────────────────────────
# 消息列表渲染
# ─────────────────────────────────────────────────────────

for msg in st.session_state.messages:
    with st.chat_message(msg["role"]):
        st.markdown(msg["content"])


# ─────────────────────────────────────────────────────────
# 用户输入处理
# ─────────────────────────────────────────────────────────

prompt = st.chat_input("请输入您的问题...")
if prompt:
    # ── 显示用户消息 ──
    with st.chat_message("user"):
        st.markdown(prompt)
    st.session_state.messages.append({"role": "user", "content": prompt})

    # ── 获取回复 ──
    with st.chat_message("assistant"):
        response_placeholder = st.empty()

        with st.spinner("🤔 思考中..."):
            # 优先用流式模式（用户体验更好）
            try:
                full_response = send_message_stream(
                    message=prompt,
                    response_container=response_placeholder,
                    session_id=st.session_state.current_session_id,
                )
            except Exception:
                # 流式失败则回退到非流式
                logger.warning("[Streamlit] 流式模式回退到非流式")
                response_placeholder.markdown("🔄 切换到普通模式...")
                full_response = send_message(
                    message=prompt,
                    session_id=st.session_state.current_session_id,
                )

        # 最终渲染（去掉光标）
        response_placeholder.markdown(full_response)

    st.session_state.messages.append({"role": "assistant", "content": full_response})
    st.rerun()
