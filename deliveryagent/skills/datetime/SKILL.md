---
name: datetime
description: "日期时间工具：获取当前时间、日期计算、时差计算。Use when: user asks about current date/time, needs date arithmetic, or wants to calculate time between dates."
version: "1.0.0"
categories: ["工具", "生活服务"]
metadata:
  openclaw:
    emoji: "📅"
    requires:
      python: []
---

# Datetime Skill

提供日期时间相关的实用工具。

## When to Use

✅ **USE this skill when:**

- "现在几点了？"
- "今天星期几？"
- "100天后是几月几号？"
- "两个日期之间相差多少天？"
- "2025年春节还有多少天？"

## When NOT to Use

❌ **DON'T use this skill when:**

- 需要定时/倒计时功能 → 用系统定时器
- 时区转换涉及夏令时 → 仅支持标准时区偏移

## Abilities

### get_current_time
获取当前日期、时间、星期信息。

**参数：** 无

### calculate_date
进行日期计算（加减天数、计算两个日期之间的差值）。

**参数：**
- `operation` (string, 必填): 操作类型，可选 "add"（加天数）、"subtract"（减天数）、"diff"（计算相差天数）
- `start_date` (string, 可选): 起始日期，格式 YYYY-MM-DD，默认今天
- `days` (integer, 可选): 天数（add/subtract 时必填）
- `end_date` (string, 可选): 结束日期，格式 YYYY-MM-DD（diff 时必填）

### time_since
计算距离某个日期过去了多久。

**参数：**
- `date_str` (string, 必填): 目标日期，格式 YYYY-MM-DD
- `label` (string, 可选): 日期描述，如"春节"、"生日"
