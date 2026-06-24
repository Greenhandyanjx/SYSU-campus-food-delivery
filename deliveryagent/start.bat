@echo off
chcp 65001 >nul
title JD-Agent 启动器
REM ─────────────────────────────────────────────────────────
REM JD-Agent 一键启动脚本 (Windows)
REM ─────────────────────────────────────────────────────────
REM 使用方式: 双击 start.bat 或命令行运行
REM 说明:
REM   本脚本会同时启动后端 API (FastAPI) 和前端 (Streamlit)
REM   - API 后端: http://localhost:8000
REM   - 前端界面: http://localhost:8501
REM   - API 文档: http://localhost:8000/docs
REM ─────────────────────────────────────────────────────────

cd /d "%~dp0"

echo ============================================
echo   JD-Agent 智能助手 - 一键启动
echo   前后端分离: Streamlit ^(8501^) + FastAPI ^(8000^)
echo ============================================

REM ── 检查 Python ──
where python >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo [错误] 未找到 Python，请安装 Python 3.10+
    pause
    exit /b 1
)

REM ── 检查关键依赖 ──
python -c "import fastapi, uvicorn, streamlit, httpx" >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo [提示] 正在安装依赖...
    pip install -r requirements.txt >nul 2>&1
    if %ERRORLEVEL% neq 0 (
        echo [错误] 依赖安装失败，请手动运行: pip install -r requirements.txt
        pause
        exit /b 1
    )
)

REM ── 检查端口占用（防止启动失败） ──
python -c "import socket; s=socket.socket(); s.settimeout(1); exit(0 if s.connect_ex(('127.0.0.1',8000))!=0 else 1); s.close()" >nul 2>&1
if %ERRORLEVEL% equ 1 (
    echo [警告] 端口 8000 已被占用，可能已有 API 实例在运行
    choice /c YN /m "是否继续？"
    if errorlevel 2 exit /b 0
)

echo.
echo [1/2] 启动 API 后端 (http://localhost:8000)...
echo       API 文档: http://localhost:8000/docs

REM 后台启动 API
start "JD-Agent API" cmd /c "python -m api.app"

REM 等待 API 就绪（健康检查循环，最多等30秒）
echo 等待 API 就绪中...
set WAIT_COUNT=0
:wait_loop
set /a WAIT_COUNT+=1
if %WAIT_COUNT% gtr 15 (
    echo [错误] API 启动超时，请检查控制台输出
    pause
    exit /b 1
)
timeout /t 2 /nobreak >nul
python -c "import urllib.request; urllib.request.urlopen('http://127.0.0.1:8000/api/v1/health', timeout=2)" >nul 2>&1
if %ERRORLEVEL% neq 0 goto wait_loop

echo [✓] API 后端已就绪

echo.
echo [2/2] 启动 Streamlit 前端 (http://localhost:8501)...
echo       请在打开的浏览器窗口中开始对话
echo.
echo 提示: 关闭此窗口不会关闭 API 后端（在 API 窗口按 Ctrl+C 停止）
echo.
start "JD-Agent Streamlit" cmd /c "streamlit run main.py --server.port 8501"

echo [✓] 前端已启动
echo.
echo ============================================
echo   全部服务已启动！
echo   前端界面: http://localhost:8501
echo   API 文档: http://localhost:8000/docs
echo ============================================
echo.
echo 按任意键查看启动说明...
pause >nul

echo.
echo ── 快速操作指南 ──
echo   ✅ 在浏览器打开 http://localhost:8501 开始聊天
echo   📖 API 文档: http://localhost:8000/docs
echo   🛑 关闭 API: 在 API 窗口按 Ctrl+C
echo.
echo 如需单独启动:
echo   终端1: python -m api.app
echo   终端2: streamlit run main.py --server.port 8501
echo.
pause
