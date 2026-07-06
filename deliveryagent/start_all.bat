@echo off
title JD-Agent Launcher
cd /d "%~dp0"
setlocal enabledelayedexpansion

echo ============================================
echo   JD-Agent Delivery Assistant
echo ============================================
echo.

where python >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo [ERROR] Python not found
    pause
    exit /b 1
)

python start_all.py --start
if %ERRORLEVEL% neq 0 (
    echo [ERROR] Failed to start services
    pause
    exit /b 1
)

:menu
cls
python start_all.py --status
echo.
echo   == Menu ==
echo   [O] Open browser
echo   [R] Refresh status
echo   [D] Toggle Docker infra
echo   [K] Stop all services
echo   [Q] Quit
set /p choice=" Select and press Enter: "

if /i "!choice!"=="O" (
    start http://localhost:5173
    goto :menu
)
if /i "!choice!"=="R" (
    goto :menu
)
if /i "!choice!"=="D" (
    python start_all.py --docker
    echo.
    pause
    goto :menu
)
if /i "!choice!"=="K" (
    python start_all.py --stop
    goto :eof
)
if /i "!choice!"=="Q" (
    python start_all.py --stop
    goto :eof
)
goto :menu
