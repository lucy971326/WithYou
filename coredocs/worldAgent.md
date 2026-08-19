# World Agent Runtime

## 公共世界感知与实时智能体架构草案 v0.1

> 目标：构建一个能够持续感知用户所处数字/现实环境、维护世界状态、进行后台认知，并通过 Realtime Omni 模型自然交流与行动的通用 Agent Runtime。

---

# 1. 核心思想

当前多数 AI Agent 的交互范式仍然是：

```text
用户提出问题
    ↓
Agent 获取上下文
    ↓
思考
    ↓
回答 / 调用工具
```

它本质上是一个 **Request → Response** 系统。

World Agent 希望变成：

```text
世界持续发生变化
        ↓
Agent 持续感知
        ↓
维护对世界的内部状态
        ↓
后台持续理解 / 推理 / 记忆
        ↓
用户随时可以自然交流
        ↓
Agent 可以回应、主动提醒或采取行动
```

核心变化：

> AI 不再等待用户描述世界。
>
> **AI 自己持续知道世界正在发生什么。**

---

# 2. 最重要的设计原则

整个系统不应该把 Realtime Omni Model 当成完整“大脑”。

Realtime 模型主要负责：

* 实时听
* 实时看
* 实时说
* 情绪表达
* 快速理解
* 低延迟交流
* 打断处理
* 简单 Tool Calling
* 当前时刻的自然反应

而复杂工作应该交给后台认知系统：

* 深度推理
* 网络搜索
* 资料理解
* 长期记忆
* 世界状态整理
* 因果分析
* 复杂 Tool Use
* Context 构建
* 情景预测

因此整体采用：

# Fast Brain + Slow Brain

即：

```text
Realtime Model
    =
Fast Brain
快速感知 / 交流 / 反应

Background LLM
    =
Slow Brain
理解 / 推理 / 搜索 / 规划
```

---

# 3. 整体架构

```text
                         REAL WORLD / DIGITAL WORLD
                                    │
              ┌─────────────────────┼─────────────────────┐
              │                     │                     │
              ▼                     ▼                     ▼
           Screen                 Audio               System
         Screenshot           Mic / Speaker        OS / App Events
         Video Frames          System Audio        Browser / IDE
              │                     │                     │
              └──────────────┬──────┴─────────────────────┘
                             ▼
                  ┌──────────────────────┐
                  │   Perception Layer   │
                  │                      │
                  │  看见了什么？        │
                  │  听见了什么？        │
                  │  系统发生了什么？    │
                  └──────────┬───────────┘
                             │
                        Observations
                             │
                             ▼
              ┌──────────────────────────────┐
              │      WORLD STATE LAYER       │
              │                              │
              │ Entity      实体             │
              │ State       状态             │
              │ Event       事件             │
              │ Relation    关系             │
              │ Activity    当前活动         │
              │ Attention   用户注意力       │
              │ Goal        用户目标         │
              │ Timeline    时间线           │
              │ Confidence  置信度           │
              │ Source      信息来源         │
              └──────────────┬───────────────┘
                             │
               ┌─────────────┴─────────────┐
               │                           │
               ▼                           ▼
      ┌─────────────────┐         ┌─────────────────┐
      │ Cognitive Core  │         │ Event / Memory  │
      │ Background LLM  │         │     System      │
      │                 │         │                 │
      │ Reasoning       │         │ Episodic Memory │
      │ Search          │         │ Semantic Memory │
      │ Planning        │         │ User Memory     │
      │ Tool Use        │         │ World History   │
      └────────┬────────┘         └────────┬────────┘
               │                           │
               └────────────┬──────────────┘
                            ▼
                 ┌──────────────────────┐
                 │   Context Compiler   │
                 │                      │
                 │ 这一刻 Realtime      │
                 │ 应该知道什么？       │
                 └──────────┬───────────┘
                            │
                     Situation Packet
                            │
                            ▼
                ┌────────────────────────┐
                │    Realtime Omni       │
                │                        │
                │ 👀 Vision              │
                │ 👂 Audio               │
                │ 🗣 Speech              │
                │ ⚡ Fast Interaction    │
                │ 🦾 Function Calling    │
                └───────────┬────────────┘
                            │
                            ▼
                       USER / WORLD
```

---

# 4. 核心：公共世界状态层

World State Layer 是整个架构真正的核心。

它不是普通 Memory。

也不是 RAG。

也不是一堆截图。

它是：

> **Agent 对“当前世界是什么样”的内部表示。**

它维护的是一个持续变化的 World State。

---

# 5. 世界状态基础数据模型

所有领域最终尽可能统一成几个基础概念：

```text
Entity
Event
State
Relation
Activity
Goal
Attention
Action
Time
Confidence
Source
```

---

## Entity

世界中持续存在的对象。

例如：

```yaml
entity:
  id: minecraft_creeper_17
  type: game_character
  name: Creeper
  properties:
    hostile: true
```

或者：

```yaml
entity:
  id: vscode_terminal_01
  type: ui_component
  application: VSCode
```

---

## Event

世界中发生的事情。

```yaml
event:
  type: explosion
  timestamp: 10:31:07

  subjects:
    - minecraft_creeper_17

  objects:
    - user_player

  result:
    user_health: decreased
```

---

## State

某个实体当前是什么状态。

```yaml
state:
  entity: user_player
  health: low
  location: village
```

---

## Relation

实体之间的关系。

```yaml
relation:
  subject: minecraft_creeper_17
  relation: near
  object: user_player
```

或者：

```yaml
relation:
  subject: anime_character_A
  relation: knows_secret
  object: secret_X
```

---

## Activity

用户当前正在做什么。

```yaml
activity:
  type: debugging
  application: VSCode
  project: backend_server
```

或者：

```yaml
activity:
  type: watching_video
  title: episode_07
  position: 00:17:32
```

---

## Attention

用户当前大概率关注什么。

例如：

```yaml
attention:
  target: terminal_error
  confidence: 0.91
```

---

## Goal

用户可能正在尝试完成什么。

```yaml
goal:
  description: start FastAPI server
  confidence: 0.86
```

注意：

Goal 可以是推测。

所以必须带 Confidence。

---

# 6. Observation 与 World State 分离

非常重要：

**感知结果不能直接等于世界事实。**

例如截图模型说：

```text
我看到一个绿色生物。
```

这只是：

```yaml
observation:
  visual_object: green_creature
  confidence: 0.81
```

后台结合：

* 游戏名称 Minecraft
* 绿色生物
* 爆炸声
* 玩家掉血

才能更新 World State：

```yaml
entity:
  type: creeper
  confidence: 0.97
```

所以：

```text
Observation
     ↓
Fusion / Reasoning
     ↓
World State
```

这样可以避免 VLM 一次识别错误污染整个世界状态。

---

# 7. 信息来源优先级

不要什么都交给 LLM 看截图猜。

信息应该遵循：

```text
确定性结构化数据
      >
应用提供的语义数据
      >
模型视觉 / 音频推断
```

例如：

当前窗口：

```text
OS API
```

比截图判断可靠。

当前视频播放时间：

```text
Player API
```

比模型根据画面猜可靠。

网页内容：

```text
DOM
```

比截图 OCR 可靠。

VSCode 当前文件：

```text
IDE Extension / API
```

比截图识别可靠。

---

# 8. Perception Adapter

所有领域能力都应该以 Adapter 形式接入。

核心系统不应该写：

```text
AnimeAgent
GameAgent
CodingAgent
BrowserAgent
```

而应该写成：

```text
World State Core
       ↑
       │
Adapters
```

例如：

```text
Screen Adapter
Audio Adapter
OS Adapter
Browser Adapter
Player Adapter
Subtitle Adapter
IDE Adapter
Game Adapter
3D Adapter
Camera Adapter
```

它们统一输出 Observation。

---

# 9. Anime 不是独立系统

动漫场景只是一种拥有丰富数据源的 World Adapter。

例如：

```text
播放器
  ↓
current_timestamp

字幕
  ↓
dialogue / speaker / timeline

网络剧情资料
  ↓
episode summary

截图
  ↓
current visual scene
```

后台整理：

```text
Episode Knowledge
        ↓
World State
```

Realtime 并不知道这是特殊 Anime Pipeline。

它最终看到的仍然只是：

```text
当前发生什么
之前发生什么
涉及哪些角色
用户已经知道什么
```

---

# 10. Anime Episode Graph

动漫领域可以额外构建：

```text
Episode
│
├── Characters
│
├── Events
│
├── Relations
│
├── Knowledge State
│
├── Locations
│
└── Timeline
```

例如：

```yaml
event:
  id: event_82
  timestamp: 00:08:21

  actors:
    - character_A

  event:
    overheard_secret: secret_X

  importance: high
```

之后：

```yaml
relation:
  subject: character_A
  relation: knows
  object: secret_X
```

---

# 11. 防剧透机制

必须维护：

```text
known_until
```

例如：

```yaml
user_progress:
  episode: 7
  known_until: 00:17:32
```

所有 Episode Retrieval 必须：

```text
event.timestamp <= known_until
```

未来剧情即使已经被后台 LLM 搜索和理解，也绝不能提供给 Realtime。

---

# 12. 游戏场景

游戏没有天然字幕剧情网。

因此更多依赖：

```text
Screen
Audio
Input
Game Events
Logs
Telemetry
```

例如：

```text
10:31:05
绿色生物进入画面

10:31:07
爆炸声

10:31:07
生命值下降
```

后台认知：

```text
Possible Event:
用户遭到 Creeper 爆炸攻击

confidence = 0.96
```

Realtime 不需要看过去 20 秒完整视频。

只需要：

```text
Current Activity:
Minecraft

Recent Important Event:
用户刚被 Creeper 偷袭爆炸

Current State:
生命值较低

User Reaction:
明显受到惊吓
```

然后：

```text
用户：
“卧槽！”

AI：
“哈哈哈哈哈你被苦力怕偷屁股了 😂”
```

---

# 13. 普通桌面场景

World Agent 不需要知道用户一定在执行什么“任务”。

例如：

```text
active_application = VSCode

recent_event:
FastAPI process exited

terminal:
Address already in use
Port 8000
```

World State：

```yaml
activity:
  type: debugging

problem:
  type: port_conflict
  port: 8000

attention:
  target: terminal
```

用户：

```text
“妈的怎么又不行。”
```

Realtime 可以理解：

```text
“8000 端口又被占了 😂”
```

用户不需要重新描述问题。

这就是 World Awareness 的价值。

---

# 14. Realtime 不应该持续吃所有视觉

持续：

```text
1 FPS
2 FPS
```

把整个屏幕不断塞进 Realtime Context，是非常低效的。

更好的方案是：

## Reactive Vision

用户一开口：

```text
立即截图
```

提供：

```text
Current Screenshot
+
Current World State
+
Relevant Events
```

这样视觉 Token 消耗极低。

---

# 15. Ambient Vision

另外维护一个非常低频的后台视觉系统：

```text
5s
10s
30s
```

或者使用：

```text
Scene Change Detection
Motion Detection
UI Change Detection
Audio Spike
Window Change
```

只有世界发生明显变化时才分析。

目的不是直接陪用户聊天。

而是：

> **保持 World State 与现实同步。**

---

# 16. Event-Driven Perception

未来应该尽可能从：

```text
Fixed Sampling
```

变成：

```text
Event Driven Sampling
```

例如：

```text
窗口变化
→ Observe

视频场景切换
→ Observe

出现明显运动
→ Observe

系统产生错误
→ Observe

游戏出现剧烈音效
→ Observe

用户突然说话
→ Observe
```

这比持续截图更加高效。

---

# 17. Event Memory

World State 只保存：

> 当前世界。

历史应该进入 Event Memory。

例如：

```text
10:30 用户进入洞穴
10:31 Creeper 爆炸
10:32 用户回村
10:34 打开箱子
```

所以：

```text
World State = Working Memory

Event Memory = Episodic Memory
```

---

# 18. Memory 分层

建议至少分：

```text
Working Memory
Episodic Memory
Semantic Memory
User Memory
```

## Working Memory

当前正在发生什么。

生命周期：

```text
秒 / 分钟
```

---

## Episodic Memory

过去发生过什么。

例如：

```text
昨天用户在 Minecraft 被 Creeper 炸死三次。
```

---

## Semantic Memory

从历史中总结出的知识。

例如：

```text
用户玩 Minecraft 时特别讨厌 Creeper。
```

---

## User Memory

长期个人偏好。

例如：

```text
用户喜欢 AI 轻松吐槽，而不是客服式回答。
```

---

# 19. Cognitive Core

后台强 LLM 不应该每秒运行。

它应该由事件触发。

例如：

```text
Observation 无法解释
→ Think

发生复杂事件
→ Think

用户提出复杂问题
→ Think

World State 存在冲突
→ Think

需要搜索
→ Think

需要长期规划
→ Think
```

任务包括：

```text
Reasoning
Search
Entity Resolution
Cause Analysis
Memory Retrieval
Knowledge Updating
Planning
Complex Tool Use
```

---

# 20. Context Compiler

这是整个系统第二关键的模块。

World State 可能非常大。

Memory 更大。

Realtime 不应该直接读取全部信息。

它应该只收到：

# Situation Packet

例如：

```yaml
situation:

  current_activity:
    user is debugging FastAPI

  current_focus:
    terminal error

  current_event:
    server start failed

  recent_events:
    - attempted server start
    - port 8000 already occupied

  relevant_memory:
    - similar problem occurred earlier today

  entities:
    - FastAPI server
    - process using port 8000

  user_state:
    mildly frustrated

  uncertainty:
    unknown which process owns port

  available_actions:
    - inspect_process
    - terminate_process
```

Realtime 只需要这几百到几千 Token。

---

# 21. Context Compiler 的核心问题

每次用户说话时：

```text
User Utterance
      +
Current World State
      +
Recent Events
      +
Relevant Memory
      +
Current Screenshot
```

Compiler 要回答：

> **这一秒，Realtime 最需要知道什么？**

它本质上是：

```text
World → Context
```

的编译器。

---

# 22. 为什么叫 Compiler

因为它不是普通 Search。

它会：

```text
筛选
压缩
排序
解释
消歧
限定知识边界
生成当前情景
```

最后输出适合 Realtime 使用的 Context。

因此：

```text
Raw World State
        ↓
Context Compiler
        ↓
Situation Packet
```

和：

```text
Source Code
    ↓
Compiler
    ↓
Executable
```

非常像。

---

# 23. Realtime Omni 的角色

Realtime 模型应该被视为：

# Interaction Model

而不是：

# General Intelligence Core

职责：

```text
自然说话
理解当前截图
理解用户声音
感知情绪
快速响应
打断
插话
表达
简单工具调用
```

不负责：

```text
维护所有历史
理解所有领域
保存整个世界
长时间深度推理
整理长期知识
```

---

# 24. Tool System

Tool 可以作用于 World。

例如：

```text
Browser Tool
File Tool
OS Tool
Search Tool
Calendar Tool
IDE Tool
Smart Home Tool
Game Tool
Robot Tool
```

执行以后：

```text
Tool Result
      ↓
Observation
      ↓
World State Update
```

形成真正的：

```text
Observe
   ↓
Understand
   ↓
Act
   ↓
Observe Result
```

闭环。

---

# 25. 主动性

真正的 World Agent 不能只有用户说话才运行。

它还要判断：

```text
世界发生了一件重要事情
          ↓
Should I Speak?
```

例如：

动漫：

```text
巨大剧情反转
→ 可以轻微反应
```

游戏：

```text
玩家即将死亡
→ 可以提醒
```

桌面：

```text
后台任务执行完成
→ 可以通知
```

但真正困难的是：

> **什么时候闭嘴。**

所以需要单独存在：

# Proactive Behavior Policy

---

# 26. Attention Model

未来可以维护：

```text
Agent Attention
User Attention
```

例如：

用户当前注意：

```text
Terminal Error
```

而不是：

```text
桌面右下角天气插件
```

即使感知系统同时看见二者，Context Compiler 也应该提高 Terminal 权重。

---

# 27. Confidence

World State 中不能把所有推断都当事实。

每一个推断最好带：

```text
confidence
source
timestamp
```

例如：

```yaml
fact:
  description: user is confused by terminal error

  confidence: 0.72

  source:
    - terminal observation
    - user utterance
```

这样后台系统以后可以纠正。

---

# 28. Source of Truth

最好保留一个 Event / Observation Log：

```text
Observation 001
Observation 002
Tool Event 003
User Event 004
Observation 005
```

World State 是从这些 Event 推导出来的。

因此：

```text
Event Log
   =
事实历史

World State
   =
当前状态
```

当状态出现错误，可以重新计算或纠正。

---

# 29. 通用领域架构

最终不是：

```text
Anime AI
Game AI
Desktop AI
Coding AI
```

而是：

```text
                 WORLD AGENT
                      │
              Universal Runtime
                      │
        ┌─────────────┼─────────────┐
        │             │             │
      Anime          Game         Desktop
     Adapter        Adapter       Adapter
```

Adapter 提高特定领域的感知质量。

但 Agent 本身不依赖领域。

---

# 30. Adapter 的本质

Adapter 不是 Agent。

它只是：

> **让 World Agent 在某个领域拥有更好的感官。**

Anime Adapter：

```text
字幕
时间轴
剧情资料
```

Browser Adapter：

```text
DOM
URL
网页语义
```

IDE Adapter：

```text
代码
Git
Terminal
Diagnostics
```

Game Adapter：

```text
Telemetry
HUD
Logs
Game State
```

未来 3D Adapter：

```text
Point Cloud
Depth
Pose
Spatial Graph
```

---

# 31. 从数字世界走向现实世界

目前 World State 可以是：

```text
Window
Application
Web Page
Video
Game Character
Error Message
```

未来可以自然扩展成：

```text
Person
Chair
Desk
Tool
Car
Room
Position
Velocity
Pose
Reachability
```

例如：

```yaml
entity:
  id: screwdriver_01

  type: tool

  position:
    x: 1.4
    y: 0.8
    z: 0.9

  relation:
    on: workbench

  reachable_by:
    - user
```

架构不需要发生本质变化。

---

# 32. 世界感知能力

真正的未来不是：

```text
Input:
Text
Image
Audio
Video
```

而是：

```text
Input:
World State
```

模型理解的不再是：

> 一张图里有什么。

而是：

> **世界中有哪些实体，它们在哪里，是什么状态，刚刚发生了什么，它们可能接下来发生什么。**

---

# 33. World Agent 与传统 Agent 的区别

传统 Agent：

```text
Task Oriented
```

核心问题：

> 我要怎么完成任务？

World Agent：

```text
World Oriented
```

首先回答：

> **现在世界是什么样？**

然后才决定：

> 我需要做什么？

---

# 34. World Agent 的核心循环

```text
PERCEIVE
感知

↓

UPDATE
更新世界状态

↓

UNDERSTAND
理解当前情景

↓

REMEMBER
形成记忆

↓

THINK
必要时后台推理

↓

INTERACT
自然交流

↓

ACT
必要时行动

↓

OBSERVE RESULT
再次观察
```

不断循环。

---

# 35. 一个非常重要的架构原则

## 世界一直存在，但 LLM 不应该一直思考。

因此：

```text
Continuous World State
+
Event-Driven Intelligence
```

而不是：

```text
Continuous LLM Inference
```

这能显著降低：

```text
延迟
Token
成本
计算量
上下文污染
```

---

# 36. V0 最小验证版本

第一版完全没必要真的实现完整 World Agent。

只验证核心：

```text
Screen
Audio
OS
   ↓
Observation
   ↓
Simple World State
   ↓
Context Compiler
   ↓
Qwen Realtime
```

World State 第一版甚至只需要：

```text
current_app
current_activity
current_focus
current_event
recent_events
current_entities
```

---

# 37. V0 场景

建议首先验证三个非常不同的场景。

## 场景 A：看番

目标：

```text
“他为什么这么说？”
```

AI 能理解：

```text
当前画面
+
当前剧情
+
历史剧情
```

---

## 场景 B：游戏

例如 Minecraft。

用户：

```text
“卧槽！”
```

AI 根据世界状态知道：

```text
刚刚发生 Creeper 爆炸
```

而不是机械询问：

```text
“发生什么了？”
```

---

## 场景 C：桌面

例如开发。

用户：

```text
“怎么又不行？”
```

AI 能根据：

```text
Current App
Terminal
Recent Events
```

理解用户正在指什么。

---

# 38. V0 成功标准

不是：

```text
AI 会很多功能
```

而是：

> **用户产生明显的“它知道我正在干什么”的感觉。**

这是整个项目最核心的体验指标。

---

# 39. V1

在 V0 之后加入：

```text
Event Memory
Semantic Memory
Background LLM
Search
Basic Tools
Attention
Confidence
```

---

# 40. V2

加入：

```text
Proactive Behavior
Long-term User Model
Multi-Application World State
Complex Tool Planning
Continuous Context Updating
```

---

# 41. V3

进入：

```text
Camera
Physical Environment
Spatial Perception
3D World State
AR Output
Embodied Agents
```

---

# 42. 最终方向

完整 World Agent 应当逐渐拥有：

```text
Perception
世界感知

World State
世界状态

Memory
世界记忆

Reasoning
理解与推理

Attention
注意力

Prediction
世界预测

Action
改变世界

Interaction
与人自然交流
```

也就是：

```text
          WORLD
            ↕
    ┌─────────────────┐
    │   WORLD AGENT   │
    │                 │
    │ Perception      │
    │ World Model     │
    │ Memory          │
    │ Reasoning       │
    │ Attention       │
    │ Action          │
    │ Interaction     │
    └─────────────────┘
            ↕
           USER
```

---

# 43. 这条路线与 Jarvis 的关系

Jarvis 真正重要的并不是：

```text
语音助手
```

而是：

> **它和 Tony 共享同一个世界。**

Tony 不需要不断向 Jarvis 描述：

```text
“我现在正在修哪个零件。”
```

Jarvis 本身知道。

Tony 不需要说：

```text
“我说的那个东西是左边第三个螺丝。”
```

Jarvis拥有空间和注意力上下文。

因此：

```text
Realtime Omni
=
眼睛 + 耳朵 + 嘴

World State
=
对当前世界的认识

Memory
=
经历

Reasoning LLM
=
慢思考

Tool / Computer Use / Robot
=
手脚

Context Compiler
=
注意力与工作记忆

Proactive Policy
=
主动行为
```

这些部分组合以后，才开始真正形成：

# Jarvis Runtime

---

# 44. 最终核心判断

过去 AI 的核心问题是：

> **“如何回答用户？”**

Agent 时代变成：

> **“如何完成任务？”**

而 World Agent 时代可能会变成：

> **“如何持续存在于一个世界中？”**

这要求 AI 不再只拥有 Conversation Context。

而是拥有：

# World Context

---

# 45. 一句话定义

> **World Agent 是一种持续感知环境、维护内部世界状态、形成长期记忆，并能够基于当前情景自然交流、推理和行动的智能体。**

其核心架构不是：

```text
LLM + Tools
```

而是：

```text
Perception
+
Universal World State
+
Event Memory
+
Cognitive Core
+
Context Compiler
+
Realtime Interaction
+
Tools / Actions
```

---

# 46. 当前最值得深入设计的三个模块

下一阶段优先级：

## ① World State Schema

解决：

> **世界应该怎么表示？**

---

## ② Observation → World State

解决：

> **看到、听到、系统发生的信息，如何转化成稳定世界状态？**

---

## ③ Context Compiler

解决：

> **世界那么大，这一秒 Realtime 到底应该知道什么？**

如果这三个问题解决得足够漂亮：

**Realtime 模型本身就会变成可替换组件。**

Qwen、Gemini、SeedRealtime 或未来任何 Omni Model 都只是 Runtime 的“交互器官”。

真正的核心资产将是：

# World Agent Runtime

---

## 最后一句

**不要给 AI 为每一个场景单独打造一个大脑。**

应该：

> **给 AI 一个世界。**

然后让动漫、游戏、桌面、浏览器、代码、现实空间，都只是这个世界的不同区域。

这可能才是从 Agent 走向真正持续智能体的一条路。
