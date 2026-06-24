"""
agent.cache — 缓存系统（四层架构）

本包实现 Day 2 的 Redis 四层缓存方案：
  ① LLM Response Cache — 省 API 费用（相同上下文命中后跳过 LLM 调用）
  ② RAG Cache         — 省向量检索时间（相同 query 避免重复多路召回）
  ③ Rate Limiter      — 防刷（滑动窗口限流）
  ④ Session Cache     — 减少 PG 读（L1=内存 → L2=Redis → L3=PG → L4=JSONL）
"""

from agent.cache.backend import CacheBackend
from agent.cache.redis_cache import RedisCache, RedisRateLimiter, get_redis_cache, get_redis_config

__all__ = [
    "CacheBackend",
    "RedisCache",
    "RedisRateLimiter",
    "get_redis_cache",
    "get_redis_config",
]
