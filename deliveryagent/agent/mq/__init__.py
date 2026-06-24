"""
MQ (Message Queue) 包 — RabbitMQ 异步任务队列

将非关键路径的耗时任务（Dream、Consolidation、Session Cleanup）
投递到 RabbitMQ，由后台 Worker 异步处理，不阻塞用户请求响应。

架构：
┌──────────────────────────────────────────────────────┐
│ AgentLoop._process_message()                         │
│   用户响应已发送 ↓                                    │
│                                                      │
│   ┌─► MQProducer.publish("dream", {})               │
│   │         ↓                                        │
│   │    RabbitMQ Exchange: agent.tasks                │
│   │         ↓                                        │
│   │    Queue: agent.dream                            │
│   │         ↓                                        │
│   │    MQConsumer → handle_dream()                   │
│   │                                                 │
│   └─► MQProducer.publish("consolidation", {...})     │
│             ↓                                        │
│        RabbitMQ Exchange: agent.tasks                │
│             ↓                                        │
│        Queue: agent.consolidation                    │
│             ↓                                        │
│        MQConsumer → handle_consolidation()           │
│                                                      │
└──────────────────────────────────────────────────────┘

使用方式：
    # 加载配置
    from agent.mq import get_mq_config
    config = get_mq_config()

    # 生产消息（在 AgentLoop 中）
    from agent.mq.producer import MQProducer
    producer = MQProducer(config)
    await producer.publish("dream", {})

    # 消费消息（后台 Worker，在 Orchestrator 中启动）
    from agent.mq.consumer import MQConsumer
    consumer = MQConsumer(config, agent_loop)
    asyncio.create_task(consumer.start())
"""
from __future__ import annotations

import os
from pathlib import Path
from typing import Any

from loguru import logger


def get_mq_config(config_path: str | Path | None = None) -> dict | None:
    """加载 RabbitMQ 配置。

    优先级（从高到低）：
    1. 环境变量 MQ_URL（Docker/K8s 环境注入）
    2. YAML 配置文件 config/mq.yml
    3. 返回 None（禁用 MQ——优雅降级）

    设计原则（与 Redis 缓存一致）：
    - 配置文件不存在 → 返回 None（MQ 功能静默禁用）
    - 配置文件存在但损坏 → 日志警告 + 返回 None
    - 使用方收到 None 后跳过所有 MQ 操作

    Returns:
        dict | None: MQ 配置字典，None = 不可用（不阻塞业务）
    """
    if config_path is None:
        # 默认搜索路径：项目根目录下的 config/mq.yml
        for candidate in [
            Path.cwd() / "config" / "mq.yml",
            Path(__file__).parent.parent.parent / "config" / "mq.yml",
        ]:
            if candidate.exists():
                config_path = candidate
                break

    if config_path is None or not Path(config_path).exists():
        logger.info("[MQ] 未找到配置文件，MQ 功能未启用（如需启用请创建 config/mq.yml）")
        return None

    try:
        import yaml
        raw = yaml.safe_load(Path(config_path).read_text(encoding="utf-8"))
        if not raw or "rabbitmq" not in raw:
            logger.warning(f"[MQ] 配置文件 {config_path} 格式错误，MQ 未启用")
            return None

        config = raw["rabbitmq"]

        # 环境变量优先级最高
        env_url = os.environ.get("MQ_URL")
        if env_url:
            config["url"] = env_url

        # 同时保留顶级 key（方便各组件直接使用）
        config["queues"] = config.get("queues", {})
        config["exchange"] = config.get("exchange", "agent.tasks")
        config["exchange_type"] = config.get("exchange_type", "direct")

        logger.info(f"[MQ] 配置已加载: {config.get('url', 'unknown')}")
        return config

    except Exception as e:
        logger.warning(f"[MQ] 配置加载失败: {e}（MQ 未启用）")
        return None
