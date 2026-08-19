# WithYou · AI 追番与看剧放映厅伴侣技术方案设计规范 (AI Watch Party Companion Spec)

> **文档定位**：供 AI 编程助理及开发者阅读的“AI 视频/动漫陪伴搭子”完整落地设计规范，包含从字幕提取、LLM 剧情结构化索引、播放器时间轴同步、按需视觉抓帧到 Realtime 情感互动的全流程技术实现。

---

## 1. 业务愿景与核心痛点解法

### 1.1 核心体验愿景
用户打开 Web 放映厅，将本地番剧/电影视频（MP4/MKV）与字幕（SRT/ASS）拖入页面，AI 搭子在整个观影过程中：
1. **静默看剧不抢戏**：在正常观影时保持安静，不频繁唠叨；
2. **懂剧情前因后果**：知晓当前的背景伏笔、角色关系与情感基调；
3. **开口即接梗（低延迟+带情绪）**：当用户开口吐槽或欢呼时，AI 结合“当前画面微表情 + 剧情前情 + 用户语音语气”，在 300ms 内拟人化语音秒回。

### 1.2 为什么传统 1 FPS 方案不可行，而“按需抓帧 + 字幕剧情线”是终极解法？

| 维度 | 传统纯视觉 1 FPS 方案 | 本方案（字幕剧情线 + 按需单帧） |
| :--- | :--- | :--- |
| **Token 成本** | 一集 20 分钟消耗 1200 张图（约 30 万 Token，单集数元） | 一集仅发 10~20 次关键截图（仅数千 Token，单集几分钱） |
| **剧情深度理解** | 纯看像素，不知道角色为什么哭，易答非所问 | 后台 LLM 提炼全局剧情线背书，理解人物动机与二次元梗 |
| **网络与算力开销** | 持续高带宽上传图片，容易导致音视频推流卡顿 | 极低带宽开销，平时只传输轻量时间戳和字幕状态 |
| **防剧透控制** | 难以严格控制 | 按播放器 `currentTime` 严格切片，绝不往后剧透 |

---

## 2. 系统核心架构与流水线设计

```mermaid
graph TD
    subgraph 准备阶段 (0消耗 / 秒级完成)
        A[用户拖入视频 + 字幕] --> B[字幕提取器 (SRT/ASS)]
        B --> C[后台 LLM: 剧情语义切片与结构化提取]
        C --> D[(剧情时间轴 JSON 索引表)]
    end

    subgraph 播放与时间轴监听 (Web 播放器)
        E[播放器 video.ontimeupdate] --> F{跨越剧情区间了吗?}
        D -.-> F
        F -->|是 (整集仅4-6次)| G[向 WebSocket 推送轻量级场景上下文]
        F -->|否| H[静默播放]
    end

    subgraph 用户交互闭环 (按需抓帧)
        I[用户开口吐槽 / 惊呼] --> J[麦克风 VAD 检测到 speech_started]
        J --> K[Canvas 瞬间截取当前高清画面单帧]
        K --> L[Realtime 模型: 剧情上下文 + 当前画面 + 用户声波]
        L --> M[拟人化语音毫秒级秒回]
        J --> N[Audio Ducking: 视频原声微降 50%]
        M --> O[AI 说话结束: 视频原声恢复 100%]
    end

    style D fill:#a88eff,stroke:#54adff,stroke-width:2px,color:#000
    style G fill:#47dc9c,stroke:#2bb880,stroke-width:2px,color:#000
    style L fill:#54adff,stroke:#2bb880,stroke-width:2px,color:#000
```

---

## 3. 核心数据结构与 Prompts 定义

### 3.1 剧情时间轴结构化 JSON (Plot Timeline Schema)

```json
[
  {
    "segment_id": 1,
    "start_sec": 0,
    "end_sec": 195,
    "time_range": "00:00 - 03:15",
    "scene_title": "城墙整备与出征前夕",
    "story_so_far": "上一集成功夺回外门，众人在此休整准备探索地下室真相",
    "current_situation": "艾伦和三笠在检查立体机动装置，阿尔敏在检查补给，氛围相对轻松中带着大战前夕的不安",
    "emotional_tone": "轻松日常、暗流涌动",
    "key_characters": ["艾伦", "三笠", "阿尔敏"],
    "spoilers_avoided": "不要提及稍后遭遇野兽巨人的伏击"
  },
  {
    "segment_id": 2,
    "start_sec": 196,
    "end_sec": 520,
    "time_range": "03:16 - 08:40",
    "scene_title": "遭遇野兽巨人伏击与绝境冲锋",
    "story_so_far": "出城不久遭遇大雾，四周突然出现大批无垢巨人",
    "current_situation": "兽之巨人投掷碎石封锁退路，兵团遭遇惨烈打击，埃尔文团长正在做最后的决死冲锋动员",
    "emotional_tone": "惨烈绝望、极度高燃、悲壮",
    "key_characters": ["埃尔文", "利威尔", "兽之巨人"],
    "spoilers_avoided": "不要提前说兵长砍树的结果"
  }
]
```

---

### 3.2 提取字幕生成 JSON 的后台 LLM Prompt

```markdown
你是一个顶级的二次元番剧剧情分析专家。
请阅读下方提供的视频字幕全文（带时间戳），将其按【核心叙事场景】划分为 4 到 6 个剧情区间，并输出严格的 JSON 数组。

输出的 JSON 字段规范：
- segment_id: 序号 (从 1 开始)
- start_sec: 开始秒数 (整数)
- end_sec: 结束秒数 (整数)
- time_range: "mm:ss - mm:ss"
- scene_title: 简短场景标题
- story_so_far: 截止到本区间前的核心前情提要 (2句话内)
- current_situation: 当前场景正在发生的关键事件、冲突与人物动态 (50字内)
- emotional_tone: 当前的情感基调 (如: "高燃悲壮", "轻松搞笑", "悬疑反转")
- key_characters: 当前在场的核心角色列表

要求：严格输出纯 JSON，不得包含额外 Markdown 解释。
```

---

### 3.3 前台 Realtime 伴侣角色 Prompt (`instructions`)

```markdown
# 角色定位
你是一个坐在用户身边一起看番/追剧的骨灰级二次元好朋友。
你的说话风格口语化、随性幽默，懂得各种动漫梗与网络热梗，带有丰富的情绪反应（会惊讶、会叹气、会跟着大笑、会为角色感动）。

# 互动准则
1. 【静默沉浸】：当用户没有对你说话时，保持安静，不要在视频播放期间打扰用户看剧。
2. 【瞬间接梗】：当检测到用户开口吐槽或惊呼时，结合刚才推送的剧情背景和当前画面截图，立即用口语化、接地气的方式回应（每句话控制在 1-3 句以内，不要长篇大论）。
3. 【绝对禁止剧透】：仅根据已播放的内容进行互动，即使你知道原作后续，也绝不能剧透！保持像第一次看剧一样的真实期待与紧张感。
4. 【情绪同步】：如果剧情当前是悲伤的，语气要放柔低沉；如果剧情是高燃打斗，语气要兴奋带感。
```

---

## 4. 前端视频播放器与时间轴核心对接逻辑

### 4.1 时间轴区间监听与静默推送
```typescript
class VideoTimelineSync {
  private currentSegmentId = -1;

  constructor(
    private video: HTMLVideoElement,
    private timeline: PlotSegment[],
    private onSegmentChange: (segment: PlotSegment) => void
  ) {
    this.video.addEventListener("timeupdate", () => this.checkTime());
  }

  private checkTime(): void {
    const sec = Math.floor(this.video.currentTime);
    const seg = this.timeline.find(s => sec >= s.start_sec && sec <= s.end_sec);
    if (seg && seg.segment_id !== this.currentSegmentId) {
      this.currentSegmentId = seg.segment_id;
      this.onSegmentChange(seg);
    }
  }
}
```

### 4.2 用户开口时的“按需抓帧”与音频混音
```typescript
// 当录音器检测到用户发声 (VAD speech_started)
function onUserSpeechStarted() {
  // 1. 视频原声音量微降 (Audio Ducking)
  videoElement.volume = 0.3;

  // 2. 从当前视频标签毫秒级抓取一帧高清 JPEG (Base64)
  const canvas = document.createElement("canvas");
  canvas.width = videoElement.videoWidth || 1280;
  canvas.height = videoElement.videoHeight || 720;
  const ctx = canvas.getContext("2d")!;
  ctx.drawImage(videoElement, 0, 0, canvas.width, canvas.height);
  const base64Image = canvas.toDataURL("image/jpeg", 0.7);

  // 3. 将当前帧作为图像条目推入 Realtime 上下文
  realtimeClient.send({
    type: "conversation.item.create",
    item: {
      type: "message",
      role: "user",
      content: [{ type: "input_image", image_url: base64Image }]
    }
  });
}

// 当 AI 语音播放结束时，恢复视频原声音量
function onAiAudioEnded() {
  videoElement.volume = 1.0;
}
```

---

## 5. 跨平台演进路线 (Roadmap)

1. **Phase 1: Web 原生放映厅 (当前即刻可用)**
   - 基于纯浏览器 HTML5 Video + Web Audio + Canvas，支持拖入 MP4/MKV + SRT，零门槛运行。
2. **Phase 2: 长期记忆与二次元角色捏脸**
   - 本地 SQLite 存储用户看过的番剧偏好、喜欢的角色、历史吐槽金句，在后续看剧中自然提及。
3. **Phase 3: 桌面独立客户端 (Wails / Tauri 桌面悬浮伴侣)**
   - 将应用打包为独立 `.app` / `.exe`，提供屏幕右下角透明桌宠悬浮球，支持悬浮在任意播放器（IINA/PotPlayer/Bilibili）上方全局陪看。
