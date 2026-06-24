"""
Agent Tools: Web 工具
参考 nanobot 的 agent/tools/web.py

提供网页抓取和搜索功能。
"""

from typing import Any

import httpx

from agent.tools.base import Tool


class WebFetchTool(Tool):
    """抓取网页内容并转为 Markdown"""

    def __init__(self, proxy: str | None = None):
        self._proxy = proxy

    @property
    def name(self) -> str:
        return "web_fetch"

    @property
    def description(self) -> str:
        return "抓取指定 URL 的网页内容并转为可读的 Markdown。适用于访问在线文档、网页等。"

    @property
    def parameters(self) -> dict[str, Any]:
        return {
            "url": {
                "type": "string",
                "description": "要抓取的 URL",
                "required": True,
            },
        }

    async def execute(self, url: str, **kwargs: Any) -> str:
        try:
            client_kw = {}
            if self._proxy:
                client_kw["proxy"] = self._proxy

            async with httpx.AsyncClient(**client_kw, timeout=30.0) as client:
                response = await client.get(url, follow_redirects=True)
                response.raise_for_status()

                content_type = response.headers.get("content-type", "")
                text = response.text

                if "text/html" in content_type:
                    # 简单提取文本：去除 HTML 标签
                    import re
                    text = re.sub(r'<script[^>]*>.*?</script>', '', text, flags=re.DOTALL)
                    text = re.sub(r'<style[^>]*>.*?</style>', '', text, flags=re.DOTALL)
                    text = re.sub(r'<[^>]+>', ' ', text)
                    text = re.sub(r'\s+', ' ', text).strip()
                elif "application/json" in content_type:
                    # JSON 响应，格式化展示
                    import json as json_mod
                    try:
                        data = json_mod.loads(text)
                        text = json_mod.dumps(data, ensure_ascii=False, indent=2)
                    except json_mod.JSONDecodeError:
                        pass

                # 截断
                max_chars = 50000
                if len(text) > max_chars:
                    text = text[:max_chars] + f"\n\n... (内容过长，已截断前 {max_chars} 字符)"

                return (
                    f"URL: {url}\n"
                    f"Content-Type: {content_type}\n"
                    f"Status: {response.status_code}\n\n"
                    f"{text}"
                )

        except httpx.HTTPError as e:
            return f"错误抓取网页: {type(e).__name__}: {e}"
        except Exception as e:
            return f"错误: {e}"
