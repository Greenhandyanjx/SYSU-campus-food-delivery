"""
缓存后端抽象接口（CacheBackend）

为什么要有抽象接口？
  四层缓存未来可能换不同的后端（如 LLM cache 用本地内存更快，
  Session cache 用 Redis），抽象接口让上层代码不依赖具体实现。

契约：
  - get(key)  → 值 or None（不存在/已过期）
  - set(key, value, ttl)  → 设置值 + 过期时间
  - delete(key)  → 删除
  - exists(key)  → bool
  - clear_prefix(prefix)  → 按前缀批量删除
"""

from abc import ABC, abstractmethod
from typing import Any, Optional


class CacheBackend(ABC):
    """缓存后端抽象接口——所有缓存实现都必须遵守此契约"""

    @abstractmethod
    async def get(self, key: str) -> Optional[Any]:
        """获取缓存值。key 不存在或已过期返回 None。"""
        ...

    @abstractmethod
    async def set(self, key: str, value: Any, ttl: int) -> None:
        """
        设置缓存值。

        Args:
            key: 缓存键
            value: 缓存值（会被 JSON 序列化）
            ttl: 过期时间（秒）。ttl=0 表示永不过期（慎用）
        """
        ...

    @abstractmethod
    async def delete(self, key: str) -> None:
        """删除单个缓存键。key 不存在时静默成功（不抛异常）。"""
        ...

    @abstractmethod
    async def exists(self, key: str) -> bool:
        """判断 key 是否存在且未过期。"""
        ...

    @abstractmethod
    async def clear_prefix(self, prefix: str) -> int:
        """
        删除所有以 prefix 开头的 key（批量失效）。

        典型用途：
        - LLM 缓存失效：session 收到新消息后，clear_prefix("llm:{session_key}")
        - 本质是 Redis SCAN + DEL 操作

        返回实际删除的数量。
        """
        ...
