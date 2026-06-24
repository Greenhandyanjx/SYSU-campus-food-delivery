"""
JD-Agent Skill: 天气查询（生产级）
==============================
使用 wttr.in 免费天气 API 获取实时天气数据。
无需 API Key，零配置即可使用。

支持的天气能力：
  - get_weather:      实时天气（天气状况、温度、湿度、风力、体感温度、能见度、紫外线、日出日落）
  - get_forecast:     未来天气预报（逐日预报，含温度范围、天气状况）
  - get_air_quality:  空气质量（从能见度和气象条件估算）
"""

from agent.skills.base import Skill, SkillAbility
from agent.tools.base import Tool

import httpx

# ── 城市名称映射（中文 → wttr.in 支持的英文名） ──
CITY_MAP = {
    "北京": "Beijing", "上海": "Shanghai", "广州": "Guangzhou",
    "深圳": "Shenzhen", "杭州": "Hangzhou", "成都": "Chengdu",
    "武汉": "Wuhan", "南京": "Nanjing", "重庆": "Chongqing",
    "西安": "Xi%27an", "珠海": "Zhuhai", "苏州": "Suzhou",
    "天津": "Tianjin", "长沙": "Changsha", "青岛": "Qingdao",
    "大连": "Dalian", "昆明": "Kunming", "厦门": "Xiamen",
    "哈尔滨": "Harbin", "郑州": "Zhengzhou", "济南": "Jinan",
    "福州": "Fuzhou", "南宁": "Nanning", "贵阳": "Guiyang",
    "海口": "Haikou", "拉萨": "Lhasa", "银川": "Yinchuan",
    "西宁": "Xining", "呼和浩特": "Hohhot", "乌鲁木齐": "Urumqi",
    "台北": "Taipei", "香港": "Hong+Kong", "澳门": "Macau",
}

# ── 天气状况 emoji 映射 ──
WEATHER_EMOJI = {
    "sunny": "☀️", "clear": "🌙", "partly cloudy": "⛅",
    "cloudy": "☁️", "overcast": "☁️", "mist": "🌫️",
    "fog": "🌫️", "patchy rain possible": "🌦️",
    "light rain": "🌦️", "moderate rain": "🌧️",
    "heavy rain": "🌧️", "rain": "🌧️", "light snow": "🌨️",
    "snow": "❄️", "heavy snow": "❄️", "sleet": "🌨️",
    "thunder": "⛈️", "thunderstorm": "⛈️",
    "patchy light drizzle": "🌦️",
}

WIND_DIR_EMOJI = {
    "N": "⬆️", "NNE": "↗️", "NE": "↗️", "ENE": "↗️",
    "E": "➡️", "ESE": "↘️", "SE": "↘️", "SSE": "↘️",
    "S": "⬇️", "SSW": "↙️", "SW": "↙️", "WSW": "↙️",
    "W": "⬅️", "WNW": "↖️", "NW": "↖️", "NNW": "↖️",
}


def _city_to_english(city: str) -> str:
    """将城市名转为 wttr.in 可用的英文标识"""
    return CITY_MAP.get(city.strip(), city.strip())


def _get_emoji(desc: str) -> str:
    """根据天气描述返回 emoji"""
    if not desc:
        return "❓"
    d = desc.lower().strip()
    for key, emoji in WEATHER_EMOJI.items():
        if key in d:
            return emoji
    return "🌡️"


def _wind_emoji(dir_deg: int | None) -> str:
    """根据风向角度返回箭头 emoji"""
    if dir_deg is None:
        return ""
    dirs = ["N", "NNE", "NE", "ENE", "E", "ESE", "SE", "SSE",
            "S", "SSW", "SW", "WSW", "W", "WNW", "NW", "NNW"]
    idx = round(dir_deg / 22.5) % 16
    return WIND_DIR_EMOJI.get(dirs[idx], "")


async def _fetch_weather_data(city: str) -> dict | None:
    """
    从 wttr.in 获取 JSON 格式天气数据。

    返回 wttr.in 的完整 JSON 响应，或 None（出错时）。
    """
    eng_city = _city_to_english(city)
    url = f"https://wttr.in/{eng_city}?format=j1"
    try:
        async with httpx.AsyncClient(timeout=10.0) as client:
            resp = await client.get(url, follow_redirects=True)
            resp.raise_for_status()
            return resp.json()
    except httpx.HTTPError:
        return None
    except Exception:
        return None


# ─────────────────────────────────────────────────────────
# Tool 1: 实时天气
# ─────────────────────────────────────────────────────────

class GetWeatherTool(Tool):
    """获取指定城市的实时天气信息（生产级，对接 wttr.in 真实数据）"""

    @property
    def name(self) -> str:
        return "get_weather"

    @property
    def description(self) -> str:
        return (
            "获取指定城市的实时天气信息，包含天气状况、气温、湿度、风力风向、"
            "体感温度、能见度、紫外线指数、日出日落时间等详细数据。"
        )

    @property
    def parameters(self) -> dict:
        return {
            "city": {
                "type": "string",
                "description": "城市名称，如：北京、上海、广州、珠海",
                "required": True,
            },
        }

    async def execute(self, city: str, **kwargs) -> str:
        data = await _fetch_weather_data(city)
        if data is None:
            return (
                f"⚠️ 无法获取 {city} 的天气数据。\n"
                f"可能原因：城市名称不支持或网络连接异常。"
            )

        try:
            cc = data["current_condition"][0]
            temp = cc["temp_C"]
            feels_like = cc.get("FeelsLikeC", temp)
            humidity = cc["humidity"]
            weather_desc = cc["weatherDesc"][0]["value"]
            wind_speed = cc["windspeedKmph"]
            wind_dir = cc["winddir16Point"]
            wind_deg = cc.get("winddirdegree")
            visibility = cc.get("visibility", "N/A")
            uv_index = cc.get("uvIndex", "N/A")
            pressure = cc.get("pressure", "N/A")

            # 日出日落
            astro = {}
            if data.get("weather"):
                astro = data["weather"][0].get("astronomy", [{}])[0]
            sunrise = astro.get("sunrise", "N/A")
            sunset = astro.get("sunset", "N/A")

            emoji = _get_emoji(weather_desc)
            wind_emoji = _wind_emoji(int(wind_deg)) if wind_deg else ""

            # 根据温度给出建议
            t = int(temp)
            if t >= 38:
                advice = "🔥 极端高温，注意防暑降温，避免户外活动"
            elif t >= 35:
                advice = "🔥 酷热天气，注意防暑降温"
            elif t >= 30:
                advice = "🌡️ 天气炎热，注意补水和防晒"
            elif t >= 25:
                advice = "🌤️ 温暖舒适，适合户外活动"
            elif t >= 20:
                advice = "🌿 气温宜人，体感舒适"
            elif t >= 15:
                advice = "🍂 微凉，建议带件外套"
            elif t >= 10:
                advice = "🧥 偏凉，注意添衣保暖"
            elif t >= 0:
                advice = "🧣 天气寒冷，注意保暖"
            else:
                advice = "❄️ 严寒天气，注意防寒保暖"

            return (
                f"📍 {city} 实时天气\n"
                f"━━━━━━━━━━━━━━━━━━━━\n"
                f"{emoji} 天气：{weather_desc}\n"
                f"🌡️  气温：{temp}°C（体感 {feels_like}°C）\n"
                f"💧 湿度：{humidity}%\n"
                f"🌬️  风力：{wind_speed} km/h {wind_emoji}（{wind_dir}）\n"
                f"👁️  能见度：{visibility} km\n"
                f"☀️  紫外线：{uv_index}\n"
                f"📊 气压：{pressure} hPa\n"
                f"🌅 日出：{sunrise}  🌇 日落：{sunset}\n"
                f"━━━━━━━━━━━━━━━━━━━━\n"
                f"💡 {advice}"
            )
        except (KeyError, IndexError, ValueError):
            return (
                f"⚠️ 获取 {city} 天气数据时解析失败。\n"
                f"原始数据格式异常，请稍后重试。"
            )


# ─────────────────────────────────────────────────────────
# Tool 2: 天气预报
# ─────────────────────────────────────────────────────────

class GetForecastTool(Tool):
    """获取未来几天的天气预报（生产级，对接 wttr.in 真实数据）"""

    @property
    def name(self) -> str:
        return "get_forecast"

    @property
    def description(self) -> str:
        return "获取指定城市未来3-7天的天气预报，包含每日天气状况、最高/最低温度。"

    @property
    def parameters(self) -> dict:
        return {
            "city": {
                "type": "string",
                "description": "城市名称",
                "required": True,
            },
            "days": {
                "type": "integer",
                "description": "预报天数（1-7），默认3",
                "required": False,
            },
        }

    async def execute(self, city: str, days: int = 3, **kwargs) -> str:
        from datetime import datetime

        data = await _fetch_weather_data(city)
        if data is None:
            return (
                f"⚠️ 无法获取 {city} 的天气预报数据。\n"
                f"可能原因：城市名称不支持或网络连接异常。"
            )

        try:
            weather_list = data.get("weather", [])
            if not weather_list:
                return f"⚠️ {city} 暂无预报数据。"

            forecast_days = min(max(days, 1), 7)

            lines = [f"📅 {city} 未来 {forecast_days} 天天气预报"]
            lines.append(f"━━━━━━━━━━━━━━━━━━━━━━")

            for day in weather_list[:forecast_days]:
                date_str = day["date"]
                try:
                    dt = datetime.strptime(date_str, "%Y-%m-%d")
                    weekday = ["周一", "周二", "周三", "周四", "周五", "周六", "周日"][dt.weekday()]
                    display_date = f"{dt.month}月{dt.day}日 {weekday}"
                except ValueError:
                    display_date = date_str

                max_temp = day["maxtempC"]
                min_temp = day["mintempC"]
                desc = day["hourly"][0]["weatherDesc"][0]["value"]
                emoji = _get_emoji(desc)

                lines.append(f"  {display_date}")
                lines.append(f"    {emoji} {desc}")
                lines.append(f"    🌡️  {min_temp}°C ~ {max_temp}°C")

            return "\n".join(lines)

        except (KeyError, IndexError, ValueError):
            return f"⚠️ 获取预报数据时解析失败，请稍后重试。"


# ─────────────────────────────────────────────────────────
# Tool 3: 空气质量
# ─────────────────────────────────────────────────────────

class GetAirQualityTool(Tool):
    """获取空气质量信息（基于 wttr.in 能见度和气象数据估算）"""

    @property
    def name(self) -> str:
        return "get_air_quality"

    @property
    def description(self) -> str:
        return "获取指定城市的空气质量指数（AQI）和污染等级评估。"

    @property
    def parameters(self) -> dict:
        return {
            "city": {
                "type": "string",
                "description": "城市名称",
                "required": True,
            },
        }

    async def execute(self, city: str, **kwargs) -> str:
        data = await _fetch_weather_data(city)
        if data is None:
            return f"⚠️ 无法获取 {city} 的空气质量数据。"

        try:
            cc = data["current_condition"][0]
            visibility_km = cc.get("visibility", "10")
            humidity = int(cc["humidity"])
            desc = cc["weatherDesc"][0]["value"].lower()
            # 基于能见度和气象条件估算 AQI
            try:
                vis = float(visibility_km)
            except (ValueError, TypeError):
                vis = 10.0

            # 估算逻辑：能见度越高 → AQI 越低
            if vis >= 20:
                aqi = int(20 + (vis - 20) * 1.5)
            elif vis >= 10:
                aqi = int(50 - (vis - 10) * 3)
            elif vis >= 5:
                aqi = int(100 - (vis - 5) * 10)
            elif vis >= 2:
                aqi = int(150 - (vis - 2) * 17)
            else:
                aqi = 200

            # 高湿度 + 静稳天气 → 可能增加污染
            if humidity > 80 and ("mist" in desc or "fog" in desc):
                aqi = min(aqi + 30, 300)

            # 有雨 → 空气质量通常较好
            if any(w in desc for w in ["rain", "drizzle", "thunder"]):
                aqi = max(aqi - 40, 10)

            # 大风 → 扩散条件好
            wind_speed = int(cc.get("windspeedKmph", "0"))
            if wind_speed > 20:
                aqi = max(aqi - 20, 10)

            aqi = max(10, min(aqi, 300))

            if aqi <= 50:
                level, color, advice = "优", "🟢", "空气质量令人满意，可正常外出"
            elif aqi <= 100:
                level, color, advice = "良", "🟡", "空气质量可接受，敏感人群注意防护"
            elif aqi <= 150:
                level, color, advice = "轻度污染", "🟠", "敏感人群减少户外活动"
            elif aqi <= 200:
                level, color, advice = "中度污染", "🔴", "建议减少外出，佩戴口罩"
            else:
                level, color, advice = "重度污染", "⚫", "尽量避免外出，关闭门窗"

            return (
                f"🏭 {city} 空气质量（估算）\n"
                f"━━━━━━━━━━━━━━━━\n"
                f"  AQI：{aqi} {color}\n"
                f"  等级：{level}\n"
                f"  能见度：{visibility_km} km\n"
                f"━━━━━━━━━━━━━━━━\n"
                f"💡 {advice}"
            )
        except (KeyError, IndexError, ValueError):
            return f"⚠️ 获取空气质量数据时解析失败，请稍后重试。"


# ─────────────────────────────────────────────────────────
# Skill 定义
# ─────────────────────────────────────────────────────────

class WeatherSkill(Skill):
    """
    天气技能：完整的气象信息能力包（生产级）。

    包含三个 Tool：
      - get_weather:      实时天气（wttr.in 真实数据）
      - get_forecast:     天气预报（逐日预报）
      - get_air_quality:  空气质量（基于气象数据估算）
    """

    @property
    def name(self) -> str:
        return "weather"

    @property
    def description(self) -> str:
        return "天气查询技能：获取实时天气、天气预报和空气质量信息"

    @property
    def version(self) -> str:
        return "2.0.0"

    @property
    def categories(self) -> list[str]:
        return ["信息获取", "生活服务"]

    @property
    def abilities(self) -> list[SkillAbility]:
        return [
            SkillAbility(
                name="get_weather",
                description="获取指定城市的实时天气，包含天气状况、气温、湿度、风力、体感温度、能见度、紫外线、日出日落等",
                keywords=["天气", "气温", "温度", "多少度", "冷", "热", "暖和"],
                examples=[
                    "今天天气怎么样？",
                    "上海多少度？",
                    "北京今天冷不冷？",
                ],
                tool=GetWeatherTool(),
            ),
            SkillAbility(
                name="get_forecast",
                description="获取未来几天的天气预报，包含每日天气、最高/最低温度",
                keywords=["预报", "未来", "明天", "后天", "这周", "周末", "下周"],
                examples=[
                    "明天北京会下雨吗？",
                    "这周末适合出去玩吗？",
                    "未来三天的天气预报",
                ],
                tool=GetForecastTool(),
            ),
            SkillAbility(
                name="get_air_quality",
                description="获取空气质量指数（AQI）和污染等级评估",
                keywords=["空气质量", "AQI", "雾霾", "污染", "pm2.5", "pm10"],
                examples=[
                    "北京的空气质量怎么样？",
                    "今天有雾霾吗？",
                ],
                tool=GetAirQualityTool(),
            ),
        ]


# ─── 导出实例（SkillLoader 自动发现需要）───
skill_instance = WeatherSkill()
