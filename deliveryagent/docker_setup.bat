@echo off
title JD-Agent Docker 镜像预拉取
cd /d "%~dp0"
setlocal enabledelayedexpansion

echo ============================================
echo   镜像预拉取脚本（DaoCloud 加速）
echo ============================================
echo.

REM 检查 Docker
where docker >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo [错误] 未找到 docker 命令
    pause
    exit /b 1
)

REM DaoCloud 镜像前缀
set "MIRROR=docker.m.daocloud.io"

REM 需要预拉取的镜像列表
set IMAGES[0]=python:3.10-slim
set IMAGES[1]=redis:7-alpine
set IMAGES[2]=postgres:16-alpine
set IMAGES[3]=rabbitmq:3-management-alpine
set COUNT=4

set SUCCESS=0
set FAIL=0

for /L %%i in (0,1,3) do (
    for /f "tokens=*" %%img in ("!IMAGES[%%i]!") do (
        echo [%%i/4] %%img ...

        REM 检查是否已存在
        docker image inspect %%img >nul 2>&1
        if !ERRORLEVEL! equ 0 (
            echo   -> 已存在，跳过
            set /a SUCCESS+=1
        ) else (
            echo   -> 从 %MIRROR%/library/%%img 拉取...
            docker pull %MIRROR%/library/%%img
            if !ERRORLEVEL! equ 0 (
                docker tag %MIRROR%/library/%%img %%img
                echo   -> OK
                set /a SUCCESS+=1
            ) else (
                echo   -> 失败
                set /a FAIL+=1
            )
        )
    )
)

echo.
echo ============================================
echo  完成：成功 %SUCCESS% / 共 %COUNT%
echo ============================================
echo.
echo 提示：如果 jd-agent 需 build，运行：
echo   docker build -t jd-agent .
echo.
pause
