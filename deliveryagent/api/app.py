"""
FastAPI 主应用入口: JD-Agent API 服务
======================================
JD-Agent 的 RESTful API 网关，底层由 AgentOrchestrator 驱动。

启动方式:
  方式一（开发）:  python -m api.app
  方式二（开发）:  uvicorn api.app:app --reload --host 0.0.0.0 --port 8000
  方式三（生产）:  uvicorn api.app:app --host 0.0.0.0 --port 8000 --workers 4

API 文档:
  - Swagger UI:  http://localhost:8000/docs
  - ReDoc:       http://localhost:8000/redoc

端点:
  POST /api/v1/chat       聊天（支持 SSE 流式）
  GET  /api/v1/health     健康检查
  GET  /api/v1/history    会话历史
"""
import os
from contextlib import asynccontextmanager

import uvicorn
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from loguru import logger

from api.routes import router


# ── Lifespan（替代弃用的 on_event） ────────────────────

@asynccontextmanager
async def lifespan(app: FastAPI):
    """服务生命周期管理"""
    logger.info("🚀 JD-Agent API 服务启动中...")
    yield
    # shutdown 逻辑
    from api.routes import _orchestrator
    if _orchestrator is not None:
        try:
            await _orchestrator.shutdown()
            logger.info("[API] AgentOrchestrator 已关闭")
        except Exception as e:
            logger.warning(f"[API] 关闭 AgentOrchestrator 时出错: {e}")


# ── 创建 FastAPI 应用 ────────────────────────────────────

app = FastAPI(
    title="JD-Agent API",
    description=(
        "自实现 AI Agent 框架 REST API。\n\n"
        "功能:\n"
        "- 多轮对话（支持流式 SSE）\n"
        "- 工具调用（RAG 检索、天气、订单查询等）\n"
        "- 四层记忆系统（MEMORY/HISTORY/history.jsonl/Session）\n"
        "- 会话持久化（JSONL）\n\n"
        "基于 ReAct 架构 · 参考 OpenClaw → nanobot 设计"
    ),
    version="1.0.0",
    lifespan=lifespan,
    docs_url="/docs",
    redoc_url="/redoc",
)

# ── CORS（允许 Streamlit 前端跨域调用）──────────────────

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],             # 生产环境应限制为前端地址
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# ── 注册路由 ─────────────────────────────────────────

app.include_router(router, prefix="/api/v1")


# ── 直接入口 ──────────────────────────────────────────

def run(host: str = "127.0.0.1", port: int = 8000):
    """启动 API 服务（适用于脚本调用）"""
    # 允许通过环境变量覆盖
    _host = os.environ.get("API_HOST", host)
    _port = int(os.environ.get("API_PORT", str(port)))
    logger.info(f"🚀 JD-Agent API 启动在 http://{_host}:{_port}")
    logger.info(f"📄 API Docs: http://{_host}:{_port}/docs")
    uvicorn.run(
        "api.app:app",
        host=_host,
        port=_port,
        reload=os.environ.get("API_RELOAD", "false").lower() == "true",
        log_level=os.environ.get("LOG_LEVEL", "info").lower(),
    )


if __name__ == "__main__":
    run()
