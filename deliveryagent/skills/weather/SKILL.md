---
name: weather
description: "获取天气信息和天气预报。Use when: user asks about weather, temperature, or forecasts for any location. NOT for: historical weather data, severe weather alerts, or detailed meteorological analysis."
version: "1.0.0"
categories: ["信息获取", "生活服务"]
metadata:
  openclaw:
    emoji: "☀️"
    requires:
      python: []
---

# Weather Skill

获取实时天气、天气预报和空气质量信息。

## When to Use

✅ **USE this skill when:**

- "今天天气怎么样？"
- "明天会下雨吗？"
- "北京的空气质量如何？"
- "这周末的天气预报"

## When NOT to Use

❌ **DON'T use this skill when:**

- 历史天气数据分析 → 使用气象数据 API
- 气候趋势分析 → 使用专业数据源
- 恶劣天气警报 → 查阅官方气象预警

## Abilities

### get_weather
获取指定城市的实时天气信息，包括天气状况、气温、湿度和风力。

**参数：**
- `city` (string, 必填): 城市名称

**示例输出：**
```
🌆 上海 今日天气
━━━━━━━━━━━━━━━━━━
☀️ 天气: 多云
🌡️  气温: 28°C
💧 湿度: 65%
🌬️  风力: 3级
━━━━━━━━━━━━━━━━━━
```

### get_forecast
获取未来3-7天的天气预报。

**参数：**
- `city` (string, 必填): 城市名称
- `days` (integer, 可选): 预报天数（1-7），默认3

### get_air_quality
获取空气质量指数（AQI）和污染等级。

**参数：**
- `city` (string, 必填): 城市名称
