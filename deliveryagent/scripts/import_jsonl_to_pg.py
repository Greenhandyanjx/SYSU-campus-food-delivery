"""
JSONL → PostgreSQL 历史数据迁移脚本。

将 sessions/ 目录下所有 .jsonl 文件中的旧聊天记录导入到 PostgreSQL。

用法:
  cd deliveryagent/
  python scripts/import_jsonl_to_pg.py

幂等安全：PgBackend.save() 使用 UPSERT + 增量写入，可重复运行。
"""

import argparse
import json
import os
import sys
from datetime import datetime
from pathlib import Path

# ── 将项目根加入 path，使内部 import 可用 ──
_HERE = Path(__file__).resolve().parent
_PROJECT_ROOT = _HERE.parent
sys.path.insert(0, str(_PROJECT_ROOT))


def load_pg_config() -> dict | None:
    """从 config/database.yml 加载 PG 配置，失败返回 None"""
    config_path = _PROJECT_ROOT / "config" / "database.yml"
    if not config_path.exists():
        print(f"[错误] 配置文件不存在: {config_path}")
        return None

    import yaml

    with open(config_path, "r", encoding="utf-8") as f:
        cfg = yaml.load(f, Loader=yaml.FullLoader) or {}

    db_cfg = cfg.get("database", {})
    dsn = db_cfg.get("dsn") or os.environ.get("DATABASE_URL") or ""
    if not dsn:
        print("[错误] database.yml 中未配置 dsn，也没有 DATABASE_URL 环境变量")
        return None

    return {
        "dsn": dsn,
        "pool_min_size": db_cfg.get("pool_min_size", 1),
        "pool_max_size": db_cfg.get("pool_max_size", 10),
        "connect_timeout": db_cfg.get("connect_timeout", 10.0),
    }


def parse_jsonl(filepath: Path) -> dict | None:
    """解析单个 .jsonl 文件，返回 {key, messages, created_at, updated_at, metadata, last_consolidated}"""
    try:
        with open(filepath, "r", encoding="utf-8") as f:
            lines = f.readlines()
    except Exception as e:
        print(f"  [跳过] 读取失败: {e}")
        return None

    if not lines:
        print(f"  [跳过] 空文件")
        return None

    # 第一行：元数据
    try:
        meta = json.loads(lines[0].strip())
    except json.JSONDecodeError:
        print(f"  [跳过] 首行（元数据）JSON 解析失败")
        return None

    if meta.get("_type") != "metadata":
        print(f"  [跳过] 首行不是 metadata（_type={meta.get('_type')}）")
        return None

    key = meta.get("key") or filepath.stem.replace("_", ":", 1)

    # 解析时间
    def _parse_ts(ts_str: str | None) -> datetime:
        if ts_str:
            try:
                return datetime.fromisoformat(ts_str)
            except (ValueError, TypeError):
                pass
        return datetime.now()

    created_at = _parse_ts(meta.get("created_at"))
    updated_at = _parse_ts(meta.get("updated_at"))
    metadata = meta.get("metadata", {})
    last_consolidated = meta.get("last_consolidated", 0)

    # 后续行：消息
    messages = []
    for i, line in enumerate(lines[1:], start=2):
        line = line.strip()
        if not line:
            continue
        try:
            msg = json.loads(line)
            # 确保 content 不为 None
            if msg.get("content") is None:
                msg["content"] = ""
            messages.append(msg)
        except json.JSONDecodeError:
            print(f"    警告: 第 {i} 行 JSON 解析失败，已跳过")
            continue

    return {
        "key": key,
        "messages": messages,
        "created_at": created_at,
        "updated_at": updated_at,
        "metadata": metadata,
        "last_consolidated": last_consolidated,
    }


async def main():
    parser = argparse.ArgumentParser(description="JSONL → PG 历史数据迁移")
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="仅扫描统计，不实际写入 PG",
    )
    args = parser.parse_args()

    # 1. 加载 PG 配置
    print("=" * 60)
    print("JSONL → PostgreSQL 历史数据迁移")
    print("=" * 60)

    pg_config = load_pg_config()
    if not pg_config:
        sys.exit(1)

    # 脱敏打印
    dsn = pg_config["dsn"]
    masked = dsn.split("@")[0].rsplit(":", 1)[0] + ":****@" + dsn.split("@")[1] if "@" in dsn else dsn
    print(f"PG DSN: {masked}")

    # 2. 扫描 JSONL 文件
    sessions_dir = _PROJECT_ROOT / "sessions"
    if not sessions_dir.exists():
        print(f"[错误] sessions 目录不存在: {sessions_dir}")
        sys.exit(1)

    jsonl_files = sorted(sessions_dir.glob("*.jsonl"))
    print(f"扫描到 {len(jsonl_files)} 个 .jsonl 文件\n")

    if not jsonl_files:
        print("没有需要迁移的数据。")
        return

    # 3. 初始化 PG 后端
    if not args.dry_run:
        from agent.session.pg_store import PgBackend

        backend = PgBackend(pg_config)
        try:
            await backend.initialize()
            print("PostgreSQL 连接成功\n")
        except Exception as e:
            print(f"[错误] PostgreSQL 连接失败: {e}")
            sys.exit(1)

    # 4. 逐个导入
    total = len(jsonl_files)
    success = 0
    failed = 0
    total_messages = 0

    for fp in jsonl_files:
        print(f"[{success + failed + 1}/{total}] {fp.name} ...", end=" ")

        data = parse_jsonl(fp)
        if data is None:
            print("x 解析失败")
            failed += 1
            continue

        msg_count = len(data["messages"])
        print(f"{msg_count} 条消息", end="")

        if msg_count == 0:
            print(" → 跳过（空会话）")
            continue

        if args.dry_run:
            print(" → [DRY RUN] 跳过写入")
            total_messages += msg_count
            success += 1
            continue

        # 写入 PG
        try:
            ok = await backend.save(
                key=data["key"],
                messages=data["messages"],
                created_at=data["created_at"],
                updated_at=data["updated_at"],
                metadata=data["metadata"],
                last_consolidated=data["last_consolidated"],
            )
            if ok:
                print(" -> OK")
                total_messages += msg_count
                success += 1
            else:
                print(" -> FAILED（保存返回 False）")
                failed += 1
        except Exception as e:
            print(f" -> ERROR: {e}")
            failed += 1

    # 5. 关闭连接
    if not args.dry_run:
        await backend.close()

    # 6. 统计
    print("\n" + "=" * 60)
    print(f"迁移完成：总计 {total} 个文件")
    print(f"  成功: {success}")
    print(f"  失败: {failed}")
    print(f"  总消息数: {total_messages}")
    if args.dry_run:
        print("  [DRY RUN] 未实际写入任何数据")
    print("=" * 60)


if __name__ == "__main__":
    import asyncio

    asyncio.run(main())
