"""
Agent Core: 消息总线
参考 nanobot 的 bus/queue.py + bus/events.py 设计

消息总线是解耦 Agent 核心与通道的关键组件：
- 通道（CLI / Streamlit / API）→ publish_inbound() → 队列
- Agent Loop → consume_inbound() → 处理 → publish_outbound()
- 通道 → consume_outbound() → 发送给用户

这样 Agent 核心不需要知道消息是来自 Streamlit 还是 API 还是 CLI，
做到了通道无关。
"""

import asyncio

from agent.core.types import InboundMessage, OutboundMessage


class MessageBus:
    """
    异步消息队列。
    
    用两个 asyncio.Queue 解耦通道和 Agent：
    - inbound: 通道 → Agent
    - outbound: Agent → 通道
    """

    def __init__(self):
        self.inbound: asyncio.Queue[InboundMessage] = asyncio.Queue()
        self.outbound: asyncio.Queue[OutboundMessage] = asyncio.Queue()

    # ─── 入站：通道→Agent ─────────────────────

    async def publish_inbound(self, msg: InboundMessage) -> None:
        """通道将用户消息发布到 Agent 队列"""
        await self.inbound.put(msg)

    async def consume_inbound(self) -> InboundMessage:
        """Agent 消费下一条入站消息（阻塞直到有消息）"""
        return await self.inbound.get()

    # ─── 出站：Agent→通道 ─────────────────────

    async def publish_outbound(self, msg: OutboundMessage) -> None:
        """Agent 将响应发布到通道队列"""
        await self.outbound.put(msg)

    async def consume_outbound(self) -> OutboundMessage:
        """通道消费下一条出站消息（阻塞直到有消息）"""
        return await self.outbound.get()

    @property
    def inbound_size(self) -> int:
        return self.inbound.qsize()

    @property
    def outbound_size(self) -> int:
        return self.outbound.qsize()
