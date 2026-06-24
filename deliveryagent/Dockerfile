# ─────────────────────────────────────────────────────────
# Dockerfile — JD-Agent 多阶段构建
# ─────────────────────────────────────────────────────────
# 用途:
#   将 JD-Agent (FastAPI API + Streamlit UI) 打包为 Docker 镜像
#
# 构建:
#   docker build -t jd-agent:latest .
#
# 运行 API:
#   docker run -d -p 8000:8000 -e DEEPSEEK_API_KEY=xxx jd-agent:latest
#
# 运行 Streamlit（需要依赖 API 容器）:
#   docker run -d -p 8501:8501 \
#     -e JD_AGENT_API_URL=http://host.docker.internal:8000 \
#     jd-agent:latest \
#     streamlit run main.py --server.port 8501 --server.address 0.0.0.0
# ─────────────────────────────────────────────────────────

# ── Stage 1: 构建阶段 ──
FROM python:3.10-slim AS builder

WORKDIR /app

# 安装编译依赖
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    && rm -rf /var/lib/apt/lists/*

# 只复制依赖文件提前构建缓存层
COPY requirements.txt .
RUN pip install --no-cache-dir --user -r requirements.txt

# ── Stage 2: 运行阶段 ──
FROM python:3.10-slim

WORKDIR /app

# 从 builder 阶段复制已安装的 Python 包
COPY --from=builder /root/.local /root/.local
ENV PATH=/root/.local/bin:$PATH

# 复制项目代码
COPY . .

# 环境变量
ENV PYTHONPATH=/app
ENV PYTHONUNBUFFERED=1

# 暴露端口
EXPOSE 8000   # FastAPI
EXPOSE 8501   # Streamlit

# 默认入口：FastAPI API 服务
# 可通过 CMD 覆盖来启动 Streamlit
CMD ["uvicorn", "api.app:app", "--host", "0.0.0.0", "--port", "8000"]
