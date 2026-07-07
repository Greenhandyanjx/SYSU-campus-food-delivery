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

:: Initialize Docker check on startup
python start_all.py --menu

:menu
cls
echo ============================================
echo   JD-Agent Service Control Panel
echo ============================================
python start_all.py --status
echo.
echo   == Start ==                == Stop ==
echo   [1] Go Backend            [5] Stop Go Backend
echo   [2] FastAPI Agent         [6] Stop FastAPI
echo   [3] Vue Frontend          [7] Stop Vue Frontend
echo   [4] Docker Infrastructure [8] Stop Docker
echo.
echo   [9] Refresh    [O] Open Browser(5173)
echo   [0] Stop All + Exit       [Q] Quit(No Stop)
echo.
set /p user_choice=" Press key: "
if not defined user_choice goto :menu
if /i "!user_choice!"=="q" goto :quit
if /i "!user_choice!"=="o" goto :open
if "!user_choice!"=="0" goto :stopall
if "!user_choice!"=="9" goto :refresh
if "!user_choice!"=="8" goto :stopdocker
if "!user_choice!"=="7" goto :stopvue
if "!user_choice!"=="6" goto :stopfastapi
if "!user_choice!"=="5" goto :stopgo
if "!user_choice!"=="4" goto :startdocker
if "!user_choice!"=="3" goto :startvue
if "!user_choice!"=="2" goto :startfastapi
if "!user_choice!"=="1" goto :startgo
goto :menu

:startgo
echo.
echo Starting Go Backend...
python start_all.py --start-service go
echo.
pause
goto :menu

:startfastapi
echo.
echo Starting FastAPI Agent...
python start_all.py --start-service fastapi
echo.
pause
goto :menu

:startvue
echo.
echo Starting Vue Frontend...
python start_all.py --start-service vue
echo.
pause
goto :menu

:startdocker
echo.
python start_all.py --start-service docker
echo.
pause
goto :menu

:stopgo
echo.
python start_all.py --stop-service go
echo.
pause
goto :menu

:stopfastapi
echo.
python start_all.py --stop-service fastapi
echo.
pause
goto :menu

:stopvue
echo.
python start_all.py --stop-service vue
echo.
pause
goto :menu

:stopdocker
echo.
python start_all.py --stop-service docker
echo.
pause
goto :menu

:refresh
goto :menu

:open
start http://localhost:5173
goto :menu

:stopall
echo.
python start_all.py --stop
echo.
pause
goto :menu

:quit
echo Exiting.
goto :eof
