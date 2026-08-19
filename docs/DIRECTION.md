# WithYou · 技术方向与验证记录

> **文档定位**：认知文档（非完整架构书）。说明我们采用的模型、计划方向、大体实现方式，以及已实测验证的成果、潜在问题与当前解方。关键决策点在此都有据可查。

---

## 1. 当前采用的模型

| 项 | 值 |
|----|----|
| 模型 | `qwen3.5-omni-plus-realtime`（Qwen 最新一代全模态，输入含视频/图像/音频/文本） |
| 连接端点 | **国际站**：`wss://dashscope-intl.aliyuncs.com/api-ws/v1/realtime`（不是国内站） |
| 认证 | `Authorization: Bearer <API Key>`（我们的 key 来自 qwencloud.com，走国际站） |
| 协议 | Realtime WebSocket。音频/文本事件对齐 OpenAI-Compatible；**图像走 Qwen 专属扩展 `input_image_buffer.append`** |

> 关键点：我们的真实 key 在国内站 (`dashscope.aliyuncs.com`) 返回 401，国际站 (`dashscope-intl`) 秒通。后续所有项目配置都要用国际站。

---

## 2. 技术栈（已拍板）

> **Relay = 中继**。不是框架名。浏览器不能拿 API Key 直连模型，所以本机 Go 夹在中间：一边接浏览器的音频/抓帧/剧情文本，一边用裸 WebSocket 跟 Qwen Realtime 对话。Key 不出本机，Qwen 专属塞图协议也收敛在这一层。

形态：本机 Web 应用（Go 听 `localhost`），不是公网上传站。

| 层 | 选型 | 不选 / 为什么 |
|----|------|----------------|
| 后端语言 | Go 1.22+ | — |
| HTTP | 标准库 `net/http` | 不上 Gin/Echo/Chi。就几个接口 + 一块中继 |
| WebSocket | `github.com/coder/websocket`（协议库） | 只当传输，不当厂商 SDK |
| Qwen Realtime | **自己写 WS 事件帧**，对 `wss://dashscope-intl.aliyuncs.com/api-ws/v1/realtime` | 不用 dashscope Realtime SDK |
| DeepSeek 富化 | **OpenAI 官方 Go SDK**（`openai-go`），baseURL 指到 `https://api.deepseek.com` | Realtime 不用这套 SDK |
| 抽字幕 | 本机 `ffmpeg`，Go `os/exec` 调 | 不绑 goav 之类 |
| 选文件 | Go 弹 Windows 原生选框，拿真实路径 | 浏览器拖文件拿不到真路径，也不上传整片 |
| 播放 | 浏览器 `<video src="/media">`，Go 按 HTTP Range 读本机文件 | 视频不进服务器磁盘副本，不是网盘 |
| 字幕解析 | 自己写 SRT/ASS 小解析器 | `forward/` 里 Python 版当对照，不进正式路径 |
| 缓存 | 本地目录，按文件内容 hash 存剧情 JSON | V0 不上 Postgres / Redis |
| 前端 | Vite + TypeScript，无框架 | 不上 React/Vue。一页看片，核心是 video / Web Audio / WS |
| 配置 | 环境变量（`QWEN_API_KEY` / `DEEPSEEK_API_KEY` / 模型名） | 不上 Viper |
| 实验代码 | `forward/` 保留作对照，正式实现全 Go | pyproject 里的 dashscope/openai 不进生产 |

**目录**：
```
backend/cmd/server/              入口
backend/internal/config/         读环境变量
backend/internal/library/        选文件 + /media Range
backend/internal/plot/           ffmpeg 抽轨 + 解析 + DeepSeek 富化 + 缓存
backend/internal/realtime/       浏览器 ↔ Qwen 的 WS 中继
frontend/                        Vite + TS
```

---

## 3. 计划方向（大体怎么做）

**一句话**：做一个"AI 陪伴你看番/看剧"的本机 Web 应用 —— 打开带软字幕轨的视频，Realtime 模型结合画面与剧情，在你吐槽时低延迟秒回，且绝不剧透。

**整体架构（三段式）**：
```
Web 前端（浏览器，只负责播控 / 麦 / 抓帧）
   │  音频流 + 开口前 500ms 抓帧 + 剧情上下文
   ▼
Go 中继（本机：选文件 / 抽字幕 / 富化 / Realtime 转发）
   │  自己写 WS：音频对齐 OpenAI 协议，图片走 Qwen input_image_buffer.append
   ▼
云端 Realtime 模型（Qwen3.5-Omni-Plus-Realtime）
```

### 核心链路
1. **离线富化**（打开视频时跑一次，全在 Go）：
   - 前端点「打开」→ Go 弹系统选框拿到本机路径 → `ffmpeg` 抽软字幕轨 → 解析时间轴 → DeepSeek **一次结构化生成**（schema 失败再重试）→ **剧情 JSON**
   - 浏览器同时用 `/media` Range 播这个本机文件（不上传）
   - 富化目的是补"字幕缺失的世界知识"（画面、动机、情感、梗）。V0 不上完整 Agent，也不上用户记忆。
2. **运行时注入**（分级、省 token、防剧透）：
   - 开始播放 → 注入**整体剧情总览**（压缩版）
   - 每到一个剧情段 → 注入**该段文本剧情 + 该段典型画面一帧**
   - 用户开口 → 从环形缓冲取 **开口前 500ms** 的帧（不是当前帧）
3. **回答**：Realtime 结合 剧情上下文 + 画面 + 用户语音，流式秒回。

### 实现边界决策（认可后定稿）
> 前端、后端（Go）、Realtime 三方职责与注入通道的明确约定。

1. **抽轨 / 解析 / 富化全走 Go**：前端不上传文件。Go 弹系统选框拿本机路径 → `ffmpeg -map 0:s:0` 抽软字幕 → 解析为带时间戳的对白 JSON → OpenAI SDK 调 DeepSeek → 回剧情 JSON。Key 不能进浏览器。播放走 `/media` Range。
2. **剧情 JSON 建议后端缓存**：一集只富化一次，之后同集直接复用（文件/内存缓存），避免重复消耗 token。前端拿到后缓存使用，刷新不重复富化。
3. **富化 LLM ≠ Realtime 模型**：富化用 DeepSeek（V4-Flash 或同级文本模型），不必用昂贵的 realtime —— 后台慢脑与前台快脑是两个独立调用、两套 key。
4. **两条注入通道别混**：
   - 剧情增量（文本）→ `conversation.item.create` 塞 `input_text`
   - 当前帧截图 → `input_image_buffer.append`（**不是** conversation.item.create，实测会被拒）
5. **两个独立触发节奏**：
   - 剧情增量注入：**跨剧情段**时触发（低频，每几分钟一次）
   - 实时抓帧：**用户开口（VAD `speech_started`）** 时触发（高频，每次吐槽）
6. **截图时机**：官方要求图片在 `speech_stopped` 前发送，所以抓帧须紧跟 `speech_started`；环形缓冲取 **开口前 500ms** 那一帧（情绪触发点，不是事后静态画面）。
7. **仅接受"带软字幕轨的视频"**（已拍板）：V0 输入范围锁定为内嵌软字幕轨（mkv/mp4）的视频，Go 侧 `ffmpeg -map 0:s:0` 自动提取字幕 → 富化。**原因**：软字幕轨与视频打包、时间轴天然 100% 对齐，从源头消灭"字幕对不齐"问题；用户无需额外下载字幕文件。V0 **不支持**：硬字幕（需 OCR）、无字幕裸片（需 ASR）、外挂字幕文件（无法保证与视频对齐）。后续可作为适配器扩展。
8. **剧情段定位容忍误差**：软字幕轨虽对齐，但不同压制帧率/剪辑仍可能引入偏移。V0 提供用户手动字幕偏移微调（±几秒）；V1 再用"对白内容匹配剧情段 key_dialogue"做自适应定位。

### 关键技术决策（改过的）
- 音频重采样到 **24kHz PCM16**，每 ~100ms 打包，LinearResampler 处理声卡 44.1k/48k。
- Web Audio 排队播放 + `speech_started` 瞬时打断（Barge-in）。
- Audio Ducking：AI 说话时视频原声降约 50%，说完恢复。
- 供应商可切换：Relay 只认 OpenAI-compatible 音频协议，供应商差异收敛在 Relay 一层；图像扩展（当前仅 Qwen 支持）做能力感知。

---

## 4. 已验证的成果（可行性结论）

> 以下均为真实跑通的实验（qwen3.5-omni-plus-realtime，麦克风 + 图像注入）。

### 3.1 图像注入链路 —— ✅ 通
- 说话时用 `input_image_buffer.append` 推 JPG → VAD 自动提交 → 模型"看到"画面并回答。
- 图片合规要求（实测确认）：**必须 JPG**（PNG 不行）、原始 ≤190KB / base64 ≤256KB、发图前至少一次 `input_audio_buffer.append`。
- 注意：`conversation.item.create` 塞 `input_image` **服务端直接拒绝**（掐断连接）。正确姿势是 `input_image_buffer.append`。⚠️ 这是文档常见误区，务必避开。

### 3.2 裸模型能力（零剧情背景，仅单帧画面 + 系统人设）
| 能力 | 结果 |
|------|------|
| 画面表层理解（人/表情/氛围/动作） | ✅ 强 |
| 字幕 OCR + 语境推断 | ✅ 意外地强（能读画面字幕并推理对话冲突） |
| 认角色（热门番，特征明显） | ✅ 能（辉夜一次命中；我推的孩子靠字幕+知识联想也对） |
| 认角色（多角色/冷门新番） | ⚠️ 会垮（咒术伏黑惠认成虎杖；角色关系张冠李戴） |
| 多轮记忆 + 人设一致性 | ✅ 优秀（记得前文、认错自然、沙雕人设稳定） |
| 首字延迟 | ✅ ~1.6s |

### 3.3 关键翻车点（改为设计约束）
- **裸模型会自信编造具体剧情**（实测：问"哪一集"，它编出"第一季第九集温泉旅行"的幻觉）。→ 设计上必须用**真实剧情 JSON 注入约束**，堵住瞎编。

---

## 5. 潜在问题与当前解方

| 潜在问题 | 现象/风险 | 当前解方 |
|---------|----------|---------|
| **剧情幻觉** | 裸模型自信编造集数/剧情/角色互动 | 离线富化剧情 JSON 作为"事实源"注入；Realtime 只在该范围说话；防剧透边界 `known_until` |
| **冷门/新番识别差** | 训练知识不足，认不出角色/世界观 | V0：字幕 + DeepSeek 一次结构化生成补背景；联网搜索/Agent 后置 |
| **图像协议坑** | `conversation.item.create` 塞图被拒 | 统一走 `input_image_buffer.append`；JPG/体积合规 |
| **Realtime 非全知** | 当作"大脑"会露馅 | 双脑：Realtime = 五官/秒回；复杂认知 = 后台离线 LLM（富化阶段） |
| **身份精度** | 多角色关系会串 | 剧情 JSON 提供精确实体/关系，减少裸猜 |
| **Key/成本** | 模型按量计费、需防泄露 | Key 只放后端（Relay），前端不可见；上下文分级注入省 token |
| **认证端点混淆** | 国内/国际站 key 不通用 | 统一用国际站 `dashscope-intl`（已实测） |

---

## 6. 下一步（V0 清单）
1. **Go 中继**：透明转发（浏览器↔本机 Go↔Qwen），先跑通最小骨架。裸 WS，无 SDK。
2. **Web 前端**：Vite + TS。LinearResampler、环形缓冲突发抓帧、Web Audio 排队/打断、Ducking。
3. **离线富化器**：系统选框拿路径 → ffmpeg 抽软字幕 → DeepSeek（OpenAI SDK）一次结构化生成（失败重试）→ 剧情 JSON。
4. **注入编排**：剧情总览 + 分段 + 开口前 500ms 抓帧三合一。

> 本阶段验证已完成，以下进入正式实现。模型、协议、可用性均已确认。
