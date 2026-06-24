@echo off
REM JD-Agent 一键启动：仅 Streamlit 前端（需先启动 API）

echo 启动 JD-Agent Streamlit 前端...
echo 确保 FastAPI 后端已在 http://127.0.0.1:8000 运行
echo.

streamlit run main.py --server.port 8501
