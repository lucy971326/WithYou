# WithYou · 核心技术文档与架构设计总览

欢迎查阅 **WithYou（实时全模态 AI 伴侣系统）** 的技术设计与实现规范文档。本目录专为开发者及后续协作的 AI Agent 设计，具备高精度、高结构化与可直接代码落地的特性。

---

## 📑 核心文档索引

| 文档名称 | 核心内容与技术定位 |
| :--- | :--- |
| **[01_REALTIME_CORE_ARCHITECTURE.md](./01_REALTIME_CORE_ARCHITECTURE.md)** | **Realtime 底层协议与音频管线规范**<br/>• Realtime WebSocket 双向流式协议标准全解析<br/>• 44.1k/48k ➔ 24kHz 高保真线性重采样器原理<br/>• Web Audio 生命周期、低延迟排队与毫秒级打断 (Barge-in)<br/>• 快思考 Realtime + 慢思考 LLM 双脑协同架构 |
| **[02_AI_MOVIE_WATCH_COMPANION_SPEC.md](./02_AI_MOVIE_WATCH_COMPANION_SPEC.md)** | **AI 追番与看剧放映厅伴侣方案规范**<br/>• 视频字幕 ➔ 后台 LLM 剧情结构化 JSON 索引流水线<br/>• Web 播放器时间轴监听与场景区间静默推送<br/>• 开口瞬间“按需抓帧”与音量闪避 (Audio Ducking)<br/>• 提示词工程、防剧透机制与多端演进路线 |

---

## 🧭 架构核心关键决策速查 (Key Decisions Summary)

1. **协议层**：
   - 统一对齐 **OpenAI-Compatible Realtime WebSocket 协议**，原生通吃阿里通义千问（Qwen3.5-Omni）、阶跃星辰（StepAudio 2.5）和 OpenAI GPT-4o Realtime，拒绝平台锁定。
2. **音频层**：
   - 输入：前端内置 `LinearResampler` 强制下采样至 `24000Hz 16-bit Mono PCM`（Little-Endian），每 100ms 打包一块；
   - 输出：`AudioPlayer` 采用 Web Audio 时钟排队调度，并在收到 `speech_started` 时瞬间 `clear()` 清空缓冲，实现 0 延迟真人级打断。
3. **看番/视频陪伴层**：
   - 彻底摒弃高成本无脑 1 FPS 全程推图；
   - 采用 **“后台字幕生成剧情分段 JSON + 播放器时间轴静默同步 + 用户开口时按需抓取单张高清帧”** 的颠覆性架构，将 Token 成本暴砍 98%，同时把剧情理解与梗文化体验拉到极致。

---

## 🛠️ 推荐开发与演进技术栈
* **前端**：TypeScript + HTML5 Video + Web Audio API + Canvas
* **后端**：Go / Node.js / Python（标准 WebSocket 中转代理 + 本地配置下发）
* **未来桌面端打包**：Wails (Go + Webview) 或 Tauri (Rust + Webview) 打造透明桌宠悬浮伴侣。
