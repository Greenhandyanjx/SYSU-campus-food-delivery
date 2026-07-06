"""
User Profile Store: 用户画像存储
===================================
存储用户的饮食偏好、忌口、常去店铺等个性化信息。

持久化方式：以 JSON 文件存储到 workspace/profiles/{username}.json
（无 PostgreSQL 时的降级方案）

设计原则：
- 增量合并：每次只提取增量信息，与已有画像合并
- 异步非阻塞：偏好提取在后台运行，不阻塞用户响应
- 自动生效：提取到的偏好自动注入到后续对话的 System Prompt 中
"""

import json
from datetime import datetime
from pathlib import Path
from typing import Any

from loguru import logger

# 默认画像结构
DEFAULT_PROFILE: dict[str, Any] = {
    "cuisines": [],          # 偏好菜系，如 ["粤菜", "日料"]
    "dislikes": [],          # 忌口，如 ["辣", "香菜"]
    "favorite_stores": [],   # 常去店铺ID，如 [12, 35]
    "price_range": "",       # 价格偏好："经济" / "中等" / "高端"
    "order_times": [],       # 常下单时间段，如 ["11:30", "18:00"]
    "notes": "",             # 其他备注信息
    "updated_at": "",        # 最后更新时间 ISO 格式
}


class UserProfileStore:
    """用户画像存储。

    职责：
    - 按用户名读取/写入画像 JSON 文件
    - 增量合并新提取的偏好
    - 内存缓存避免重复读盘

    使用方式：
        store = UserProfileStore(workspace)
        profile = store.get("aaa")
        store.merge("aaa", {"cuisines": ["粤菜"]})
    """

    def __init__(self, workspace: Path) -> None:
        self._profiles_dir = workspace / "profiles"
        self._profiles_dir.mkdir(parents=True, exist_ok=True)
        self._cache: dict[str, dict] = {}

    def _profile_path(self, username: str) -> Path:
        return self._profiles_dir / f"{username}.json"

    def get(self, username: str) -> dict:
        """获取用户画像（含内存缓存）。"""
        if username in self._cache:
            return self._cache[username]

        path = self._profile_path(username)
        if path.exists():
            try:
                with open(path, encoding="utf-8") as f:
                    data = json.load(f)
                    self._cache[username] = data
                    return data
            except Exception as e:
                logger.warning(f"[Profile] 读取 {username} 画像失败: {e}")

        profile = dict(DEFAULT_PROFILE)
        self._cache[username] = profile
        return profile

    def save(self, username: str, profile: dict) -> None:
        """保存用户画像到文件 + 更新缓存。"""
        profile["updated_at"] = datetime.now().isoformat()
        self._cache[username] = profile
        path = self._profile_path(username)
        try:
            with open(path, "w", encoding="utf-8") as f:
                json.dump(profile, f, ensure_ascii=False, indent=2)
        except Exception as e:
            logger.warning(f"[Profile] 保存 {username} 画像失败: {e}")

    def merge(self, username: str, updates: dict) -> dict:
        """增量合并用户画像。列表字段去重追加，标量字段直接覆盖。"""
        current = self.get(username)
        changed = False

        for key, value in updates.items():
            if not value and value != "":
                continue
            if key not in DEFAULT_PROFILE:
                continue

            if isinstance(value, list):
                existing = set(current.get(key, []) or [])
                new_items = [v for v in value if v and v not in existing]
                if new_items:
                    current[key] = list(existing) + new_items
                    changed = True
            elif isinstance(value, str) and value.strip():
                if current.get(key) != value.strip():
                    current[key] = value.strip()
                    changed = True

        if changed:
            self.save(username, current)

        return current

    def has_profile_data(self, username: str) -> bool:
        """判断用户是否有有效画像数据。"""
        profile = self.get(username)
        for key, value in profile.items():
            if key == "updated_at":
                continue
            if isinstance(value, list) and value:
                return True
            if isinstance(value, str) and value.strip():
                return True
        return False
