"""
JD-Agent 技能：日期时间工具
==============================
提供当前时间获取、日期计算、时差计算等能力。
所有功能纯本地执行，无需外部 API。
"""

from datetime import datetime, date, timedelta

from agent.skills.base import Skill, SkillAbility
from agent.tools.base import Tool


class GetCurrentTimeTool(Tool):
    """获取当前日期和时间信息"""

    @property
    def name(self) -> str:
        return "get_current_time"

    @property
    def description(self) -> str:
        return "获取当前的日期、时间、星期和农历信息"

    @property
    def parameters(self) -> dict:
        return {}

    async def execute(self, **kwargs) -> str:
        now = datetime.now()
        weekdays = ["星期一", "星期二", "星期三", "星期四", "星期五", "星期六", "星期日"]
        weekday = weekdays[now.weekday()]

        return (
            f"📅 当前时间\n"
            f"━━━━━━━━━━━━━━━━━━\n"
            f"📆 日期: {now.strftime('%Y年%m月%d日')}\n"
            f"🕐 时间: {now.strftime('%H:%M:%S')}\n"
            f"📌 星期: {weekday}\n"
            f"━━━━━━━━━━━━━━━━━━"
        )


class CalculateDateTool(Tool):
    """日期计算器：加减天数 / 计算日期差"""

    @property
    def name(self) -> str:
        return "calculate_date"

    @property
    def description(self) -> str:
        return "进行日期计算：给一个日期加上或减去若干天，或计算两个日期相差的天数"

    @property
    def parameters(self) -> dict:
        return {
            "operation": {
                "type": "string",
                "description": "操作类型：add（加天数）、subtract（减天数）、diff（计算相差天数）",
                "required": True,
            },
            "start_date": {
                "type": "string",
                "description": "起始日期，格式 YYYY-MM-DD，默认为今天",
                "required": False,
            },
            "days": {
                "type": "integer",
                "description": "天数（add/subtract 时必填）",
                "required": False,
            },
            "end_date": {
                "type": "string",
                "description": "结束日期，格式 YYYY-MM-DD（diff 时必填）",
                "required": False,
            },
        }

    async def execute(self, operation: str, start_date: str = None, days: int = None, end_date: str = None, **kwargs) -> str:
        try:
            start = date.fromisoformat(start_date) if start_date else date.today()
        except ValueError:
            return f"错误：日期格式无效 '{start_date}'，请使用 YYYY-MM-DD 格式"

        if operation == "add" or operation == "subtract":
            if days is None:
                return "错误：add/subtract 操作需要提供 days 参数"
            if operation == "subtract":
                days = -days
            result = start + timedelta(days=days)
            return (
                f"📅 日期计算结果\n"
                f"━━━━━━━━━━━━━━━━━━\n"
                f"{start.isoformat()} {'+' if days > 0 else ''}{days}天\n"
                f"  → {result.isoformat()} ({result.strftime('%A')})\n"
                f"━━━━━━━━━━━━━━━━━━"
            )

        elif operation == "diff":
            if not end_date:
                return "错误：diff 操作需要提供 end_date 参数"
            try:
                end = date.fromisoformat(end_date)
            except ValueError:
                return f"错误：日期格式无效 '{end_date}'，请使用 YYYY-MM-DD 格式"
            diff = (end - start).days
            direction = "相差" if diff >= 0 else "相差"
            return (
                f"📅 日期差计算结果\n"
                f"━━━━━━━━━━━━━━━━━━\n"
                f"  起始: {start.isoformat()}\n"
                f"  结束: {end.isoformat()}\n"
                f"  {direction} {abs(diff)} 天\n"
                f"━━━━━━━━━━━━━━━━━━"
            )

        else:
            return f"错误：不支持的操作 '{operation}'，支持 add/subtract/diff"


class TimeSinceTool(Tool):
    """计算距离某个日期过去了多久"""

    @property
    def name(self) -> str:
        return "time_since"

    @property
    def description(self) -> str:
        return "计算从某个日期到现在过去了多长时间（如：距离春节还有多少天）"

    @property
    def parameters(self) -> dict:
        return {
            "date_str": {
                "type": "string",
                "description": "目标日期，格式 YYYY-MM-DD",
                "required": True,
            },
            "label": {
                "type": "string",
                "description": "日期描述标签，如'春节'、'生日'等",
                "required": False,
            },
        }

    async def execute(self, date_str: str, label: str = "", **kwargs) -> str:
        try:
            target = date.fromisoformat(date_str)
        except ValueError:
            return f"错误：日期格式无效 '{date_str}'，请使用 YYYY-MM-DD 格式"

        today = date.today()
        diff = (target - today).days
        name = label or date_str

        if diff > 0:
            return f"⏳ 距离 {name} 还有 {diff} 天"
        elif diff == 0:
            return f"🎉 {name} 就是今天！"
        else:
            return f"⌛ {name} 已经过去 {abs(diff)} 天了"


class DatetimeSkill(Skill):
    """
    日期时间技能：日期时间查询与计算能力包。

    包含三个 Tool：
      - get_current_time: 获取当前日期时间
      - calculate_date:   日期计算
      - time_since:       距离日期计算
    """

    @property
    def name(self) -> str:
        return "datetime"

    @property
    def description(self) -> str:
        return "日期时间技能：获取当前时间、日期计算、时差计算等"

    @property
    def version(self) -> str:
        return "1.0.0"

    @property
    def categories(self) -> list[str]:
        return ["工具", "生活服务"]

    @property
    def abilities(self) -> list[SkillAbility]:
        return [
            SkillAbility(
                name="get_current_time",
                description="获取当前日期、时间和星期信息",
                keywords=["时间", "日期", "今天", "现在", "星期", "几号", "几月"],
                examples=[
                    "现在几点了？",
                    "今天星期几？",
                    "现在是什么时间？",
                ],
                tool=GetCurrentTimeTool(),
            ),
            SkillAbility(
                name="calculate_date",
                description="进行日期加减或计算两个日期相差的天数",
                keywords=["计算日期", "哪天", "多少天后", "多少天前", "日期差", "相差多少天"],
                examples=[
                    "100天后是几月几号？",
                    "从今天到月底还有多少天？",
                    "12月25日距离今天有多少天",
                ],
                tool=CalculateDateTool(),
            ),
            SkillAbility(
                name="time_since",
                description="计算距离某个日期过去了多久",
                keywords=["距离", "倒计时", "还有多少天", "还有几天", "还剩"],
                examples=[
                    "距离春节还有多少天？",
                    "还有几天到元旦？",
                    "距离2025年高考还有多少天",
                ],
                tool=TimeSinceTool(),
            ),
        ]


skill_instance = DatetimeSkill()
