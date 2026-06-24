#!/bin/bash
# ─────────────────────────────────────────────────────────
# JD-Agent 一键启动脚本 (Linux / WSL)
# ─────────────────────────────────────────────────────────
# 使用方式:
#   chmod +x start.sh && ./start.sh
# ─────────────────────────────────────────────────────────

set -e

echo "============================================"
echo "  JD-Agent 启动脚本"
echo "  前后端分离模式: Streamlit + FastAPI"
echo "============================================"

# ── 激活虚拟环境（如有） ──
if [ -d "venv" ]; then
    echo "[✓] 使用虚拟环境"
    source venv/bin/activate
fi

# ── 配置 ──
API_PORT=${API_PORT:-8000}
STREAMLIT_PORT=${STREAMLIT_PORT:-8501}

echo ""
echo "[1/2] 启动 FastAPI 后端 (端口 $API_PORT)..."
echo "      API Docs: http://localhost:$API_PORT/docs"
echo ""

# 后台启动 API
uvicorn api.app:app --host 127.0.0.1 --port $API_PORT --reload &
API_PID=$!

# 等待 API 就绪
echo "      等待 API 就绪..."
for i in $(seq 1 15); do
    if curl -s http://127.0.0.1:$API_PORT/api/v1/health > /dev/null 2>&1; then
        echo "      [✓] API 就绪"
        break
    fi
    sleep 1
done

echo ""
echo "[2/2] 启动 Streamlit 前端 (端口 $STREAMLIT_PORT)..."
echo "      打开浏览器: http://localhost:$STREAMLIT_PORT"
echo "      按 Ctrl+C 停止所有服务"
echo ""

# 前台启动 Streamlit
streamlit run main.py --server.port $STREAMLIT_PORT

# 清理
echo ""
echo "正在停止 API..."
kill $API_PID 2>/dev/null || true
echo "[✓] 已停止所有服务"
