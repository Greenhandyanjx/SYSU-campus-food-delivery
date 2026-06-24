"""
API Schemas（已迁移至 api/routes.py）

Pydantic 模型已内联至 api/routes.py 以保持路由定义自包含。
本文件保留为空桩，避免已有引用报错。
"""
from api.routes import ChatRequest, ChatResponse, HealthResponse, HistoryResponse, HistoryMessage
