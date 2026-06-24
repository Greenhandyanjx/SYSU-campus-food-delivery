# SKILL PATTERN — JD-Agent 技能规范

> 参考 **[OpenClaw Skill 架构](https://docs.openclaw.ai)** 设计
>
> Skill 是一组相关能力的封装，是 JD-Agent 的"能力包"。

---

## 一、什么是 Skill

**Skill（技能）** 是对标 OpenClaw 中 `skills/` 目录的概念。一个 Skill 包含：

```
skills/weather/              ← 技能目录
├── SKILL.md                 ← 技能声明文档（YAML 头部 + Markdown 说明）
└── __init__.py              ← 技能代码实现（导出 skill_instance）
```

Skill 和 Tool 的区别：

| 维度 | Tool | Skill |
|------|------|-------|
| 粒度 | **原子操作** — 读文件、查天气 | **能力包** — 一组相关 Tool |
| 声明方式 | Python 类继承 `Tool` | SKILL.md 文档 + `Skill` 子类 |
| 发现机制 | 代码显式注册 | **文件系统自动发现** |
| 意图匹配 | Function Calling | **关键词+示例匹配** |
| 附带资源 | 无 | SKILL.md + 数据文件 + 模板 |

---

## 二、Skill 目录结构

### 最小结构

一个可被自动发现的 Skill 目录至少包含：

```python
skills/your_skill/
├── SKILL.md         # 必须：技能声明
└── __init__.py      # 必须：导出 skill_instance
```

### 完整结构

```python
skills/your_skill/
├── SKILL.md             # 技能声明文档
├── __init__.py          # 代码实现（导出 skill_instance）
├── prompts/             # （可选）Prompt 模板
│   └── system.txt
├── data/                # （可选）数据文件
│   └── config.json
└── README.md            # （可选）开发者文档
```

---

## 三、SKILL.md 格式

### YAML 头部

必须包含 `---` 包裹的 YAML 头部（对标 OpenClaw）：

```yaml
---
name: weather
description: "获取天气信息和天气预报"
version: "1.0.0"
categories: ["信息获取", "实用工具"]
metadata:
  openclaw:                    # OpenClaw 兼容字段
    emoji: "☀️"               # 技能图标
    requires:                  # 依赖声明
      bins: ["curl"]
      env: ["WEATHER_API_KEY"]
---
```

### Markdown 正文

```
# Weather Skill

SDK 级的天气能力包。当用户询问天气时，调用此技能。

## When to Use

✅ **USE this skill when:**

- "今天天气怎么样？"
- "明天会下雨吗？"

## When NOT to Use

❌ **DON'T use this skill when:**

- 历史天气数据分析
- 长期气候趋势
```

---

## 四、`__init__.py` 格式

### 最小示例

```python
from agent.skills.base import Skill, SkillAbility, SkillAsset
from agent.tools.base import Tool


# ── 1. 定义 Tool ──

class MyTool(Tool):
    @property
    def name(self) -> str:
        return "my_tool"

    @property
    def description(self) -> str:
        return "我的工具"

    @property
    def parameters(self) -> dict:
        return {}

    async def execute(self, **kwargs) -> str:
        return "执行结果"


# ── 2. 定义 Skill ──

class MySkill(Skill):
    @property
    def name(self) -> str:
        return "my_skill"

    @property
    def description(self) -> str:
        return "我的技能包"

    @property
    def version(self) -> str:
        return "1.0.0"

    @property
    def abilities(self) -> list[SkillAbility]:
        return [
            SkillAbility(
                name="my_tool",
                description="执行我的操作",
                keywords=["关键字A", "关键字B"],
                examples=["示例查询"],
                tool=MyTool(),
            ),
        ]

    @property
    def categories(self) -> list[str]:
        return ["信息获取"]

    async def initialize(self) -> None:
        # 技能初始化逻辑（可选）
        pass


# ── 3. 导出实例（最关键的一行！）──

skill_instance = MySkill()
```

### 导出约定

`__init__.py` **必须**导出一个名为 `skill_instance` 的变量，类型为 `Skill` 子类实例。

```python
# ✅ 正确
skill_instance = MySkill()

# ❌ 错误：命名不对，Loader 找不到
my_skill = MySkill()
```

---

## 五、Skill 生命周期

### 启动时

```
AgentLoop.__init__()
  │
  ├─ SkillLoader.discover_all()
  │    ├─ 扫描 skills/ 下所有子目录
  │    ├─ 读取 SKILL.md → 解析 yaml 头部
  │    └─ 加载 __init__.py → 获取 skill_instance
  │
  ├─ SkillManager 持有所有 Skill
  │
  └─ SkillManager.register_all_tools(ToolRegistry)
       └─ 每个 Skill 的每个 Tool → ToolRegistry
```

### 对话中

```
用户输入 → AgentLoop
  │
  ├─ SkillManager.match(query) → [SkillMatchResult]
  │    └─ 智能匹配最相关的技能
  │
  ├─ 匹配结果注入 System Prompt
  │    └─ LLM 能看到当前有哪些技能可用
  │
  └─ LLM 通过 Function Calling 调用技能下的 Tool
       └─ 一切照常：ToolRegistry.execute()
```

---

## 六、编写指南

### 1. 一个 Tool 一个能力

每个 `SkillAbility` 对应一个具体的 Tool，Tool 的逻辑应当和能力描述一致。

### 2. 关键词很重要

`keywords` 是意图匹配的关键。写全同义词：
```python
# ✅ 好
keywords=["天气", "气温", "下雨", "下雪", "台风", "预报"]

# ❌ 不好：太少了
keywords=["天气"]
```

### 3. 示例查询帮助文档化

`examples` 既用于匹配，也方便给 LLM 展示：
```python
examples=[
    "今天上海天气怎么样？",
    "北京明天会下雨吗？",
    "这个周末适合出去玩吗？",
]
```

### 4. 初始化钩子

如果技能需要连接外部服务、加载模型、检查环境，在 `initialize()` 中完成：
```python
async def initialize(self) -> None:
    import os
    api_key = os.environ.get("MY_API_KEY")
    if not api_key:
        raise RuntimeError("MY_API_KEY 未设置")
    self._client = MyClient(api_key)
```

### 5. 依赖声明

在 SKILL.md 中声明运行时依赖，`SkillManager` 会在注册时检查：
```yaml
metadata:
  openclaw:
    requires:
      bins: ["ffmpeg", "curl"]
      env: ["OPENAI_API_KEY", "DATA_DIR"]
      python: ["numpy>=1.24", "pandas"]
```

---

## 七、内置技能

JD-Agent 自带以下技能（位于 `skills/` 目录）：

| 技能 | 能力 | 位置 |
|------|------|------|
| 🌤️ **Weather** | 天气查询 | `skills/weather/` |
| 📚 **RAG** | 知识库检索 | `skills/rag/` |
| 📋 **Template** | 技能模板 | `skills/template/` |

---

## 八、FAQ

### Q：Skill 和 Tool 到底什么关系？

A：Skill 是"能力包"，Tool 是"原子操作"。一个 Skill 可以包含多个 Tool。

```
Skill: WeatherSkill
  ├── Tool: get_weather()       → 查当前天气
  ├── Tool: get_forecast()      → 查天气预报
  └── Tool: get_air_quality()   → 查空气质量
```

### Q：什么时候需要写一个新的 Skill？

A：当你有**一组相关的 Tool** 需要一起管理、有**文档**要附带、希望被**自动发现**时。

### Q：能只写 Tool 不写 Skill 吗？

A：可以。Tool 仍然可以通过 `register_all_tools()` 手动注册。Skill 是在 Tool 之上的组织层。

### Q：意图匹配怎么工作的？

A：`match_query()` 方法遍历技能的 keywords 和 examples：
- 关键词命中 → +0.3 分
- 示例中关键短语匹配 → +0.2 分
- 总分上限 1.0
- 可按分数降序取 top-k 个技能
- 子类可重写 match_query() 实现更好的匹配
