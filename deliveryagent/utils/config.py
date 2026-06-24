"""
配置加载：读取 YAML 配置文件
（迁移自原始项目的 config_handler.py，适配新路径）
"""
import os
import yaml
from utils.path_tool import get_abs_path


def load_rag_config(config_path: str = None, encoding: str = "utf-8"):
    path = config_path or get_abs_path("config/rag.yml")
    with open(path, "r", encoding=encoding) as f:
        return yaml.load(f, Loader=yaml.FullLoader)


def load_chroma_config(config_path: str = None, encoding: str = "utf-8"):
    # 优先读取 vector_store.yml（新名称），不存在则回退 chroma.yml
    path = config_path or get_abs_path("config/vector_store.yml")
    if not os.path.exists(path):
        path = config_path or get_abs_path("config/chroma.yml")
    with open(path, "r", encoding=encoding) as f:
        return yaml.load(f, Loader=yaml.FullLoader)


def load_prompts_config(config_path: str = None, encoding: str = "utf-8"):
    path = config_path or get_abs_path("config/prompts.yml")
    with open(path, "r", encoding=encoding) as f:
        return yaml.load(f, Loader=yaml.FullLoader)


def load_agent_config(config_path: str = None, encoding: str = "utf-8"):
    path = config_path or get_abs_path("config/agent.yml")
    with open(path, "r", encoding=encoding) as f:
        return yaml.load(f, Loader=yaml.FullLoader)


def load_database_config(config_path: str = None, encoding: str = "utf-8"):
    """
    加载 PostgreSQL 数据库配置。

    返回 dict，含 dsn / pool_min_size / pool_max_size / connect_timeout。
    如果配置文件不存在或 dsn 为空，返回空字典（表示不启用 PG）。
    """
    path = config_path or get_abs_path("config/database.yml")
    if not os.path.exists(path):
        return {}
    with open(path, "r", encoding=encoding) as f:
        cfg = yaml.load(f, Loader=yaml.FullLoader) or {}
    db_cfg = cfg.get("database", {})
    # 去掉 dsn 为 None 或空字符串的情况
    dsn = db_cfg.get("dsn") or os.environ.get("DATABASE_URL") or ""
    if not dsn:
        return {}
    return {
        "dsn": dsn,
        "pool_min_size": db_cfg.get("pool_min_size", 1),
        "pool_max_size": db_cfg.get("pool_max_size", 10),
        "connect_timeout": db_cfg.get("connect_timeout", 10.0),
    }


rag_conf = load_rag_config()
chroma_conf = load_chroma_config()
prompts_conf = load_prompts_config()
agent_conf = load_agent_config()
