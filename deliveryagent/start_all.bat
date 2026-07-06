@echo off
chcp 65001 >nul
title JD-Agent 外卖全栈启动器
REM ─────────────────────────────────────────────────────────
REM JD-Agent 外卖全栈一键启动脚本 (Windows)
REM ─────────────────────────────────────────────────────────

cd /d "%~dp0"
setlocal enabledelayedexpansion

REM ── 关闭 ANSI 彩色（Windows 批处理兼容性差），只用纯文本 ──
echo ============================================
echo   JD-Agent 外卖智能助手 - 全栈一键启动
echo ============================================
echo.

REM ──────────────────────────────────────
REM 检查 Python + Node
REM ──────────────────────────────────────
where python >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo [错误] 未找到 Python，请先安装 Python 3.10+
    pause
    exit /b 1
)
where node >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo [错误] 未找到 Node.js，请先安装 Node 18+
    pause
    exit /b 1
)

REM ──────────────────────────────────────
REM 定义路径
REM ──────────────────────────────────────
set "GO_BACKEND_DIR=D:\桌面\软件工程中级实训\SYSU-campus-food-delivery\backend"
set "GO_BACKEND_EXE=%GO_BACKEND_DIR%\server.exe"
set "FRONTEND_DIR=D:\桌面\软件工程中级实训\SYSU-campus-food-delivery\frontend"
set "AGENT_DIR=%CD%"

REM ── 验证路径存在 ──
if not exist "%GO_BACKEND_EXE%" (
    echo [错误] 未找到 Go 后端程序:
    echo   %GO_BACKEND_EXE%
    echo 请先编译 Go 后端或将 server.exe 放在正确位置。
    pause
    exit /b 1
)
if not exist "%FRONTEND_DIR%\package.json" (
    echo [错误] 未找到 Vue 前端项目:
    echo   %FRONTEND_DIR%
    pause
    exit /b 1
)

REM ──────────────────────────────────────
REM Step 1: Go 后端 (port 3000)
REM ──────────────────────────────────────
echo [1/3] Go 后端配送服务 (port 3000)...

python -c "import socket; exit(0 if socket.connect_ex(('127.0.0.1',3000))!=0 else 1)" >nul 2>&1
if %ERRORLEVEL% equ 1 (
    echo   [OK] 端口 3000 已在运行
) else (
    echo   [..] 启动 Go 后端...
    start "Go-Backend-3000" "%GO_BACKEND_EXE%"
    timeout /t 4 /nobreak >nul
    python -c "import socket; exit(0 if socket.connect_ex(('127.0.0.1',3000))!=0 else 1)" >nul 2>&1
    if !ERRORLEVEL! equ 0 (
        echo   [FAIL] Go 后端启动失败，请检查 %GO_BACKEND_EXE%
        pause
        exit /b 1
    ) else (
        echo   [OK] Go 后端已启动
    )
)

REM ──────────────────────────────────────
REM Step 2: FastAPI Agent (port 8000)
REM ──────────────────────────────────────
echo [2/3] FastAPI Agent 后端 (port 8000)...

python -c "import socket; exit(0 if socket.connect_ex(('127.0.0.1',8000))!=0 else 1)" >nul 2>&1
if %ERRORLEVEL% equ 1 (
    echo   [OK] 端口 8000 已在运行
) else (
    echo   [..] 启动 FastAPI...
    start "FastAPI-Agent-8000" cmd /c "uvicorn api.app:app --host 127.0.0.1 --port 8000 --reload"

    REM 等待 API 就绪（最多 20 秒）
    set WAIT_CNT=0
    :wait_api
    set /a WAIT_CNT+=1
    if !WAIT_CNT! gtr 20 (
        echo   [WARN] FastAPI 启动超时，请手动检查
        goto :skip_api_wait
    )
    timeout /t 1 /nobreak >nul
    python -c "import urllib.request; urllib.request.urlopen('http://127.0.0.1:8000/api/v1/health',timeout=2)" >nul 2>&1
    if !ERRORLEVEL! neq 0 goto :wait_api
    :skip_api_wait
    echo   [OK] FastAPI 已就绪 (http://localhost:8000)
)

REM ──────────────────────────────────────
REM Step 3: Vue 前端 (port 5173)
REM ──────────────────────────────────────
echo [3/3] Vue 前端 (port 5173)...

python -c "import socket; exit(0 if socket.connect_ex(('127.0.0.1',5173))!=0 else 1)" >nul 2>&1
if %ERRORLEVEL% equ 1 (
    echo   [OK] 端口 5173 已在运行
) else (
    echo   [..] 启动 Vite 开发服务器...
    start "Vue-Frontend-5173" /D "%FRONTEND_DIR%" cmd /c "npm run dev"
    timeout /t 6 /nobreak >nul
    python -c "import socket; exit(0 if socket.connect_ex(('127.0.0.1',5173))!=0 else 1)" >nul 2>&1
    if !ERRORLEVEL! equ 0 (
        echo   [..] Vite 启动中，请稍候...
    ) else (
        echo   [OK] 前端已启动
    )
)

echo.
echo ============================================
echo   全栈服务启动完成！
echo ============================================
echo.
echo   Site:   http://localhost:5173
echo   API:    http://localhost:8000/docs
echo.

REM ── 检测各服务状态（用 helper 脚本避免 cmd 多行字符串问题） ──
python "%~dp0_check_status.py" 2>nul || call :_simple_status
goto :menu

REM ── 备用状态检测（如果 helper 脚本不存在） ──
:_simple_status
echo 服务状态:
python -c "import socket; s=socket.socket(); s.settimeout(2); print('  3000 Go后端:   running' if s.connect_ex(('127.0.0.1',3000))==0 else '  3000 Go后端:   offline'); s.close()"
python -c "import socket; s=socket.socket(); s.settimeout(2); print('  8000 FastAPI:  running' if s.connect_ex(('127.0.0.1',8000))==0 else '  8000 FastAPI:  offline'); s.close()"
python -c "import socket; s=socket.socket(); s.settimeout(2); print('  5173 Vue前端:  running' if s.connect_ex(('127.0.0.1',5173))==0 else '  5173 Vue前端:  offline'); s.close()"
goto :eof

:menu
echo.
echo   ── 操作 ──
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
echo 重新检测...
python "%~dp0_check_status.py" 2>nul || call :_simple_status
goto :menu

:kill_all
echo 正在停止所有服务...
for %%p in (5173 8000 3000) do (
    for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%%p " ^| findstr LISTEN') do (
        if not "%%a"=="" (
            taskkill /F /PID %%a >nul 2>&1 && echo   [OK] Port %%p (PID %%a) stopped
        )
    )
)
echo 所有服务已停止。
pause
exit /b 0
