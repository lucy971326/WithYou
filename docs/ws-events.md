# WithYou · Realtime WS 事件规格

> 前后端同一份合同。没写进本表的 `type`，客户端→云端 **直接拒**。云端→客户端未知 type **记日志并透传**，不崩。
> 依据：`docs/qwenRealtimeAPI/omniAPI/clientEvent.md`、`serverEvent.md`。Python SDK / 声音复刻不进正式路径。

两层别混：

| 层 | 通道 | 干什么 |
|----|------|--------|
| Realtime 面 | 浏览器 ↔ Go 中继 ↔ Qwen WS | 下面这些事件 |
| App 控制面 | HTTP | 选文件、富化进度、`/media`、剧情 JSON。不走这条 WS |

---

## 1. 建连与会话

```
WS 握手（Bearer key，国际站）
  → session.created          云端默认配置
  → session.update           我们覆盖
  → session.updated          之后才准发音频 / 图
```

`session.update` 建议固定这么发（V0）：

```json
{
  "event_id": "evt_...",
  "type": "session.update",
  "session": {
    "modalities": ["text", "audio"],
    "voice": "Tina",
    "audio": {
      "input":  { "format": { "type": "pcm", "sample_rate": 24000 } },
      "output": { "format": { "type": "pcm", "sample_rate": 24000 } }
    },
    "instructions": "<人设 + 当前已知剧情，见 §4>",
    "turn_detection": {
      "type": "semantic_vad",
      "threshold": 0.5,
      "silence_duration_ms": 800
    }
  }
}
```

约束：

- 输入采样率官方允许 `8000 | 16000 | 24000 | 48000`，默认 16k。我们 **显式 24k**，和输出、前端 LinearResampler 同一套。
- 音频格式用嵌套 `audio.*.format`，不用过时的 `input_audio_format`。
- **不要设 `idle_timeout_ms`**：看番时用户长时间不说话是正常的，设了模型会自己找话聊。
- `semantic_vad` 仅 3.5 Omni Realtime 有，能滤嗯嗯/背景音，比 `server_vad` 合适。
- 发过音频之后不能再改输入格式。

---

## 2. 客户端 → 云端（allowlist）

官方一共就这些。V0 实际会发的打 ✅，能发但不主动发的打 △，禁止的打 ⛔。

| type | V0 | 谁发 | 字段 | 何时 |
|------|----|------|------|------|
| `session.update` | ✅ | 前端或 Go | `session` | 收到 `session.created` 之后立刻。人设写这里，段切不重发 |
| `input_audio_buffer.append` | ✅ | 前端 | `audio` = base64 PCM | `session.updated` 之后，约每 100ms 一块 |
| `input_image_buffer.append` | ✅ | 前端 | `image` = JPEG base64 | `speech_started` 时，取开口前 500ms 那帧 |
| `response.cancel` | ✅ | 前端 | 无 | barge-in：用户开口且 AI 还在说。若云端 `interrupt_response=true` 已自动掐，重复 cancel 可能回 `error`，忽略即可 |
| `input_audio_buffer.commit` | △ | — | 无 | **VAD 开着时不要发**，云端自己提交 |
| `input_audio_buffer.clear` | △ | — | 无 | V0 不用 |
| `response.create` | △ | — | 可带 `response.instructions` | **VAD 开着时不要发**，会自动生成。SDK 允许给这一轮加一次性指令，V0 不用 |
| `conversation.item.create` | ✅ | 前端或 Go | 见下 | 剧情增量：只许 `input_text`。SDK 方法名 `create_item` |
| `session.finish` | △ | Go | 无 | SDK 有，官方事件表没有。干净收尾用，V0 可先 `close` |

`conversation.item.create` 文本形态（SDK 只声明 `item: dict`，对齐 OpenAI 兼容口径）：

```json
{
  "event_id": "evt_...",
  "type": "conversation.item.create",
  "item": {
    "type": "message",
    "role": "user",
    "content": [
      { "type": "input_text", "text": "<当前段剧情 JSON 压缩文本>" }
    ]
  }
}
```

- `input_text` ✅ 剧情 / 倒退提示 / 快进补全
- `input_image` ⛔ 实测掐连接。图只走 `input_image_buffer.append`

图的硬限制（官方）：

- JPG/JPEG，建议 480p/720p，最大 1080p
- 编码前 ≤190KB，base64 ≤256KB
- **每秒最多 1 张**
- **本轮必须先有过至少一次 `input_audio_buffer.append`**
- 跟音频一起在 `commit` 时提交（VAD 模式下 commit 由云端做）
- 必须赶在 `speech_stopped` 前发出

⛔ 禁止：

- `conversation.item.create` 里塞 `input_image`（实测掐连接）
- `conversation.item.truncate` / `delete`（SDK 都没包）

---

## 3. 云端 → 客户端

Go 中继原样转发。前端按 type 分流。

### 3.1 必须处理

| type | 前端做什么 |
|------|------------|
| `error` | 亮出来；`invalid_request_error` 多半是我们发错事件 |
| `session.created` | 紧接着发 `session.update` |
| `session.updated` | 打开「可以推麦 / 可以塞图」门闩 |
| `input_audio_buffer.speech_started` | ① 从环形缓冲取 500ms 前 JPEG → `input_image_buffer.append` ② 若 AI 在说话：停播 + 可选 `response.cancel` |
| `input_audio_buffer.speech_stopped` | 本轮图必须已经发出 |
| `response.created` | 标记 AI 开始说，视频原声 duck 50% |
| `response.audio.delta` | base64 PCM 排队进 Web Audio（24kHz） |
| `response.audio.done` | 本段音频结束 |
| `response.audio_transcript.delta` | 字幕条（AI 在说啥） |
| `response.audio_transcript.done` | 字幕定稿 |
| `response.done` | 解除 duck；`status` 可能是 completed / failed / incomplete |
| `response.text.delta` / `response.text.done` | 仅 `modalities=["text"]` 时会出现。V0 双模态时主要走 transcript |

### 3.2 透传即可（V0 不必做业务）

`input_audio_buffer.committed` / `cleared`  
`conversation.item.created`  
`conversation.item.input_audio_transcription.delta` / `.completed` / `.failed`  
`response.output_item.added` / `.done`  
`response.content_part.added` / `.done`

用户 ASR 的 delta 用 `text + stash` 拼预览；最终以 `.completed.transcript` 为准。V0 可以不展示用户转写。

---

## 4. 剧情文本怎么进模型

官方 `clientEvent.md` 漏了，DashScope SDK 源码写明了：`create_item` → `conversation.item.create`，`item` 是任意 dict。

V0：

| 内容 | 事件 |
|------|------|
| 人设 + 本集完整剧情 | 开场一次 `session.update.instructions` |
| 总览 / 跨段 / 快进补全 / 倒退提示 | `conversation.item.create` + `input_text` |
| 当前画面 | `input_image_buffer.append` |
| 用户说话 | `input_audio_buffer.append`（VAD 自动收） |

段切只追加文本 item，不重写整段 instructions。

`conversation.item.create` 若被拒：**直接报错，停在这**。不准偷偷改成 `session.update.instructions`。方案要改就显式重开，不留隐式兜底。

---

## 5. V0 主路径（开着 VAD）

```
开麦循环     input_audio_buffer.append          前端 → 云
开口         speech_started                     云 → 前端
             input_image_buffer.append          前端 → 云   （500ms 前那帧）
闭嘴         speech_stopped                     云 → 前端
             （云端自己 commit，自己 response.create）
AI 说话      response.audio.delta × N           云 → 前端
             response.audio_transcript.delta
结束         response.done
```

快进 / 倒退 / 跨段：App 层算出该灌哪段文本，再发 `conversation.item.create`。不另发明 WS type。
