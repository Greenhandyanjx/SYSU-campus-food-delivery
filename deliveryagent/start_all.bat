@echo off
chcp 65001 >nul
title JD-Agent 外卖全栈启动器
REM ─────────────────────────────────────────────────────────
REM JD-Agent 外卖全栈一键启动脚本 (Windows)
REM 启动六项服务:
REM   0. Docker 容器    (Redis:6379, PostgreSQL:5432, RabbitMQ:5672)
REM   1. Go 后端        (配送数据服务) → localhost:3000
REM   2. FastAPI        (Agent 聊天)   → localhost:8000
REM   3. Vue 前端       (外卖平台界面)  → localhost:5173
REM
REM 用户端两个 AI 入口已集成在前端中:
REM   - 右下角 FAB 按钮 → AgentPanel.vue（浮动面板）
REM   - 底部栏 "AI 助手" → /user/agent（全屏页面）
REM ─────────────────────────────────────────────────────────

cd /d "%~dp0"

setlocal enabledelayedexpansion

set "GREEN=[32m"
set "YELLOW=[33m"
set "RED=[31m"
set "CYAN=[36m"
set "RESET=[0m"

echo ============================================
echo   JD-Agent  外卖智能助手 - 全栈一键启动
echo   六项服务: Docker ^(Redis/PG/RMQ^) + Go + FastAPI + Vue
echo ============================================
echo.

REM ──────────────────────────────────────
REM 检查 Python + Node
REM ──────────────────────────────────────
where python >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo %RED%[错误]%RESET% 未找到 Python
    pause
    exit /b 1
)
where node >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo %RED%[错误]%RESET% 未找到 Node.js
    pause
    exit /b 1
)

REM ──────────────────────────────────────
REM 定义路径（本脚本位于 deliveryagent/ 目录下）
REM ──────────────────────────────────────
set "PROJECT_ROOT=D:\桌面\软件工程中级实训\SYSU-campus-food-delivery"
set "GO_BACKEND_DIR=%PROJECT_ROOT%\backend"
set "GO_BACKEND_EXE=%GO_BACKEND_DIR%\server.exe"
set "FRONTEND_DIR=%PROJECT_ROOT%\frontend"
set "AGENT_DIR=%CD%"

REM ──────────────────────────────────────
REM Step 0: Docker 容器 (Redis / PostgreSQL / RabbitMQ)
REM ──────────────────────────────────────
echo [0/4] Docker 基础设施...

where docker >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo   %YELLOW%[!]%RESET% 未找到 Docker，跳过容器启动
    echo   请确保 Redis(6379) PostgreSQL(5432) RabbitMQ(5672) 已手动运行
) else (
    set "DOCKER_CONTAINERS=jd-agent-redis jd-agent-postgres jd-agent-rabbitmq"
    for %%c in (%DOCKER_CONTAINERS%) do (
        docker ps --format "{{.Names}}" | findstr /C:"%%c" >nul 2>&1
        if !ERRORLEVEL! equ 0 (
            echo   %GREEN%[✓]%RESET% %%c 已在运行
        ) else (
            docker ps -a --format "{{.Names}}" | findstr /C:"%%c" >nul 2>&1
            if !ERRORLEVEL! equ 0 (
                echo   %YELLOW%[~]%RESET% 启动 %%c...
                docker start %%c >nul 2>&1
                if !ERRORLEVEL! equ 0 (
                    echo   %GREEN%[✓]%RESET% %%c 已启动
                ) else (
                    echo   %RED%[✗]%RESET% %%c 启动失败
                )
            ) else (
                echo   %YELLOW%[~]%RESET% %%c 不存在，请先创建容器
            )
        )
    )
)

REM ──────────────────────────────────────
REM Step 1: Go 后端 (配送数据) — port 3000
REM ──────────────────────────────────────
echo [1/4] Go 后端配送服务 (port 3000)...

python -c "import socket; exit(0 if socket.connect_ex(('127.0.0.1',3000))!=0 else 1)" >nul 2>&1
if %ERRORLEVEL% equ 1 (
    echo   %GREEN%[✓]%RESET% 端口 3000 已在运行
) else (
    if exist "%GO_BACKEND_EXE%" (
        echo   %YELLOW%[~]%RESET% 启动 Go 后端...
        start "Go-Backend-3000" "%GO_BACKEND_EXE%"
        timeout /t 3 /nobreak >nul
        python -c "import socket; exit(0 if socket.connect_ex(('127.0.0.1',3000))!=0 else 1)" >nul 2>&1
        if !ERRORLEVEL! equ 0 (
            echo   %RED%[✗]%RESET% Go 后端启动失败
            pause
        ) else (
            echo   %GREEN%[✓]%RESET% Go 后端已启动
        )
    ) else (
        echo   %YELLOW%[!]%RESET% 未找到 Go 后端: %GO_BACKEND_EXE%
        pause
    )
)

REM ──────────────────────────────────────
REM Step 2: FastAPI Agent 后端 — port 8000
REM ──────────────────────────────────────
echo [2/4] FastAPI Agent 后端 (port 8000)...

python -c "import socket; exit(0 if socket.connect_ex(('127.0.0.1',8000))!=0 else 1)" >nul 2>&1
if %ERRORLEVEL% equ 1 (
    echo   %GREEN%[✓]%RESET% 端口 8000 已在运行
) else (
    echo   %YELLOW%[~]%RESET% 启动 FastAPI...
    start "FastAPI-Agent-8000" cmd /c "uvicorn api.app:app --host 127.0.0.1 --port 8000"

    set /a WAIT=0
    :wait_api
    set /a WAIT+=1
    if !WAIT! gtr 10 (
        echo   %RED%[✗]%RESET% API 启动超时
        pause
        goto :skip_api_wait
    )
    timeout /t 2 /nobreak >nul
    python -c "import urllib.request; urllib.request.urlopen('http://127.0.0.1:8000/api/v1/health',timeout=2)" >nul 2>&1
    if !ERRORLEVEL! neq 0 goto :wait_api
    :skip_api_wait
    echo   %GREEN%[✓]%RESET% FastAPI 已就绪 ^(http://localhost:8000^)
)

REM ──────────────────────────────────────
REM Step 3: Vue 前端开发服务器 — port 5173
REM ──────────────────────────────────────
echo [3/4] Vue 前端 (port 5173)...

python -c "import socket; exit(0 if socket.connect_ex(('127.0.0.1',5173))!=0 else 1)" >nul 2>&1
if %ERRORLEVEL% equ 1 (
    echo   %GREEN%[✓]%RESET% 端口 5173 已在运行
) else (
    if exist "%FRONTEND_DIR%\package.json" (
        echo   %YELLOW%[~]%RESET% 启动 Vite 开发服务器...
        start "Vue-Frontend-5173" /D "%FRONTEND_DIR%" npm run dev
        timeout /t 5 /nobreak >nul
        python -c "import socket; exit(0 if socket.connect_ex(('127.0.0.1',5173))!=0 else 1)" >nul 2>&1
        if !ERRORLEVEL! equ 0 (
            echo   %YELLOW%[~]%RESET% Vite 启动中，等待就绪...
        ) else (
            echo   %GREEN%[✓]%RESET% 前端已启动
        )
    ) else (
        echo   %RED%[!]%RESET% 未找到前端项目: %FRONTEND_DIR%
        pause
    )
)

echo.
echo ============================================
echo   全栈服务启动完成！
echo ============================================
echo.
echo   %CYAN%服务状态:%RESET%
echo   ┌──────────────────────┬──────────┬──────────┐
echo   │ 组件                 │ 端口     │ 状态     │
echo   ├──────────────────────┼──────────┼──────────┤

docker ps --format "{{.Names}}" | findstr /C:"jd-agent-redis" >nul 2>&1
if %ERRORLEVEL% equ 0 (echo   │ Redis 缓存            │ 6379     │ %%GREEN%%[运行]%%RESET%% │) else (echo   │ Redis 缓存            │ 6379     │ %%RED%%[离线]%%RESET%% │)

docker ps --format "{{.Names}}" | findstr /C:"jd-agent-postgres" >nul 2>&1
if %ERRORLEVEL% equ 0 (echo   │ PostgreSQL 数据库     │ 5432     │ %%GREEN%%[运行]%%RESET%% │) else (echo   │ PostgreSQL 数据库     │ 5432     │ %%RED%%[离线]%%RESET%% │)

docker ps --format "{{.Names}}" | findstr /C:"jd-agent-rabbitmq" >nul 2>&1
if %ERRORLEVEL% equ 0 (echo   │ RabbitMQ 消息队列     │ 5672     │ %%GREEN%%[运行]%%RESET%% │) else (echo   │ RabbitMQ 消息队列     │ 5672     │ %%RED%%[离线]%%RESET%% │)

python -c "import socket; exit(0 if socket.connect_ex(('127.0.0.1',3000))!=0 else 1)" >nul 2>&1
if %ERRORLEVEL% equ 0 (echo   │ Go 配送后端           │ 3000     │ %%RED%%[离线]%%RESET%% │) else (echo   │ Go 配送后端           │ 3000     │ %%GREEN%%[运行]%%RESET%% │)

python -c "import socket; exit(0 if socket.connect_ex(('127.0.0.1',8000))!=0 else 1)" >nul 2>&1
if %ERRORLEVEL% equ 0 (echo   │ FastAPI Agent         │ 8000     │ %%RED%%[离线]%%RESET%% │) else (echo   │ FastAPI Agent         │ 8000     │ %%GREEN%%[运行]%%RESET%% │)

python -c "import socket; exit(0 if socket.connect_ex(('127.0.0.1',5173))!=0 else 1)" >nul 2>&1
if %ERRORLEVEL% equ 0 (echo   │ Vue 前端              │ 5173     │ %%RED%%[离线]%%RESET%% │) else (echo   │ Vue 前端              │ 5173     │ %%GREEN%%[运行]%%RESET%% │)

echo   └──────────────────────┴──────────┴──────────┘
echo.
echo   %CYAN%访问入口:%RESET%
echo   🏠 外卖平台  → http://localhost:5173
echo   📖 API 文档  → http://localhost:8000/docs
echo.
echo   %CYAN%AI 助手入口（已集成在前端中）:%RESET%
echo   🔵 右下角 FAB 按钮 — 浮动聊天面板
echo   🤖 底部栏 "AI 助手" 标签 — 全屏聊天页
echo.
echo   %YELLOW%提示:%RESET% 关闭此窗口不会停止服务。
echo   如需一键停止，请按 K 后回车。
echo.

:menu
echo   ── 请选择 ──
echo   [O] 打开浏览器
echo   [R] 刷新状态
echo   [K] 停止所有服务
echo   [Q] 退出
set /p choice="输入选项后回车 (O/R/K/Q): "

if /i "!choice!"=="O" (
    start http://localhost:5173
    goto :menu
)
if /i "!choice!"=="R" (
    cls
    goto :refresh
)
if /i "!choice!"=="K" (
    goto :kill_all
)
if /i "!choice!"=="Q" (
    echo 退出启动器。
    goto :eof
)
goto :menu

:refresh
echo.
echo 重新检测...

:kill_all
echo.
echo %YELLOW%正在停止所有服务...%RESET%
for %%p in (5173 8000 3000) do (
    for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%%p " ^| findstr LISTEN') do (
        if not "%%a"=="" (
            taskkill /F /PID %%a >nul 2>&1 && echo   [✓] 端口 %%p (PID %%a) 已停止
        )
    )
)
echo   [i] Docker 容器服务未停止（docker stop 可手动停止）
echo.
echo %GREEN%所有进程服务已停止。%RESET%
pause
exit /b 0
