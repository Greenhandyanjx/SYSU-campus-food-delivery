@echo off
REM Delivery Agent 一键启动：仅 FastAPI 后端（开发调试用）
REM 启动后 API 文档在 http://localhost:8000/docs

echo 启动 Delivery Agent API 服务...
uvicorn api.app:app --host 127.0.0.1 --port 8000 --reload
