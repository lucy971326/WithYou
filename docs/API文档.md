> ## Documentation Index
> Fetch the complete documentation index at: https://platform.stepfun.com/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# 双向实时语音

开启实时语音通话能力，支持语音和文本输入，并支持输出音频。

## 快速体验

我们提供了模型体验中心，您可以点击下方链接快速体验阶跃星辰语音模型的能力。

* [语音克隆](https://www.stepfun.com/studio/audio?tab=clone) - 上传 5-10 秒音频即可复刻您的专属音色，支持多种情感方案，适用于个性化语音生成场景。
* [实时对话](https://www.stepfun.com/studio/audio?tab=voice-chat) - 提供极低延迟的双向语音流式交互，支持服务端 VAD 和 联网搜索，适用于智能助手和客服场景。
* [语音推理](https://www.stepfun.com/studio/audio?tab=conversation) - 具备深度声音理解能力，能听懂语气背后的深意，通过边想边说的推理模式，大幅提高对话质量。

同时也提供了开源的 Realtime API Demo 供您参考：[Step-Realtime-Console](https://github.com/stepfun-ai/Step-Realtime-Console)。

## 请求方式

WebSocket

## 请求地址

```text theme={null}
wss://api.stepfun.com/v1/realtime
```

## 请求头

* `Authorization` `string` ***required*** <br /> 鉴权使用的 KEY，其值为 `Bearer STEP_API_KEY`

## 请求参数

* `model` `string` ***required***<br />需要使用的模型名称，当前仅支持 `stepaudio-2.5-realtime` `step-1o-audio` `step-audio-2` `step-audio-2-mini` `step-audio-r1.5`

## 调用说明

Realtime API 需要在服务链接成功后，发送对应的 Client Event ，获取对应的 Server Event ，来完成交互。

### 公共参数

以下为 Client Event 和 Server Event 的公共参数

| **字段名**   | **类型** | **描述**            |
| --------- | ------ | ----------------- |
| event\_id | string | 事件 ID             |
| type      | string | Event 类型，可选项见下方说明 |

## Client Event 列表

### 创建/更新 session

type: `session.update`

发送此事件以创建或更新会话的默认配置。客户端可以随时发送此事件来更新 session 配置，任何字段都可能随时更新，除了 "voice" 之外。服务器将使用 `session.updated` 事件进行响应。

* `modalities` `array<string>`<br />模型可以使用的模态集。 固定为 `["text", "audio"]`

* `instructions` `string`<br />默认系统指令（即 system message）附加到 model 调用之前。此字段允许客户端指导模型获得所需的响应。该模型可以被指导回答内容和格式（例如，"要非常简洁"、"行为友好"、"这里有好回应的示例"）和音频行为（例如，"快声说"、"在你的声音中注入情感"、"经常大笑"）。不能保证模型会遵循这些说明，但它们会为模型提供有关所需行为的指导。

* `voice` `string`<br />生成时使用的音色信息，支持[官方音色](/docs/zh/guide/tts#支持音色)，和自定义音色。自定义音色传入对应的音色 ID 即可，可以通过[查询音色列表](/docs/zh/api-reference/audio/list_voice)查看当前可用的音色 ID。

  如果需要使用复刻音色，则需要使用[音色复刻](/docs/zh/api-reference/audio/create-voice)获取的 voiceid。

  * `stepaudio-2.5-realtime` 模型的对应音色复刻生成效果和 StepAudio 2.5 TTS 相当，可以使用 StepAudio 2.5 TTS 试听复刻效果。实际复刻时，使用 StepAudio 2.5 TTS 作为复刻的模型。

  如果使用 `step-audio-2-mini`，支持 `qingchunshaonv`（青春少女）和 `wenrounansheng`（温柔男声），且需要在 instructions 结尾处对应添加"请使用默认男声与用户交流"或"请使用默认女声与用户交流"。

  如果使用 `step-audio-2`，支持 `qingchunshaonv`（青春少女），`wenrounansheng`（温柔男声），`elegantgentle-female`（高雅女声），`livelybreezy-female`（活力女声）。之外，模型支持音色克隆，你可以上传音色片段获取自定义 voice，参考[音色复刻](/docs/zh/api-reference/audio/create-voice)。

* `turn_detection` `object` optional<br />ServerVAD 参数，默认关闭
  * `type` `string` ***required***<br />当前仅支持 `server_vad`， 配置后开启服务端 VAD。
  * `prefix_padding_ms` `integer` ***optional***<br />只在 type 为 `server_vad` 时有效，当检测到用户语音开始时，回溯添加的前置音频时长，单位为毫秒，默认值为 500 毫秒。
  * `silence_duration_ms` `integer` ***optional***<br />只在 type 为 `server_vad` 时有效，静音检测时长，单位为毫秒，默认值为 100 毫秒，当检测到用户语音结束后，静音达到该时长则认为用户语音输入结束。
  * `energy_awakeness_threshold` `integer` ***optional***<br />只在 type 为 `server_vad` 时有效，能量唤醒阈值，范围为0-5000，默认值为 2500，当音频能量超过该阈值时认为用户开始说话。

* `input_audio_format` `string`<br />输入音频的格式。 当前仅支持 `pcm16`

* `output_audio_format` `string`<br />输出音频的格式。 当前仅支持 `pcm16`

* `tools` `object array` ***optional***<br />Toolcall 支持的函数列表
  * `type` `string`<br />工具类型，总是为 `function`、`retrieval`
  * `function` `object`<br />函数内容的描述

    * `name` `string`<br />函数名称，要求纯英文、数字和\_-符号，建议不要超过64个字符长度。
    * `description` `string`<br />函数描述，支持中英文，它用于告诉模型函数实现何种功能和目的，方便模型来判断和选择。
    * `parameters` `object`<br />函数的参数
      * `type` `object`<br />参数描述，一般为 object
      * `properties` `object`<br />函数参数内容，以 key 为参数的名称，然后通过 `type` 和 `description` 描述参数的类型和介绍。
        * `type` `string|number|integer|object|array|boolean`<br />参数类型，可以参考[json-schema](https://json-schema.org/understanding-json-schema/reference/type)介绍
        * `description` `string`<br />函数参数描述，支持中英文，它用于告诉模型函数参数的含义。

    当 type 为 `retrieval` 时

    * `description` `string`<br />函数描述，支持中英文，它用于告诉模型函数实现何种功能和目的，方便模型来判断和选择。
    * `options` `object`<br />函数的参数
      * `vector_store_id` `string`<br />知识库的 ID
      * `prompt_template` `string`<br />知识库召回内容插入文档的模板。默认为 `从文档{{knowledge}}中找到问题{{query}}的答案。根据文档内容中的语句找到答案, 如果文档中没用答案则告诉用户找不到`，其中 `{{knowledge}}` 为召回的内容、 `{{query}}` 为用户的查询内容。可以根据实际情况修改。

**Sample**

```json theme={null}
{
    "event_id": "event_abc",
    "type": "session.update",
    "session": {
        "modalities": ["text", "audio"],
        "instructions": "你是由阶跃星辰提供的AI聊天助手，你擅长中文，英文，以及多种其他语言的对话。",
        "voice": "linjiajiejie",
        "input_audio_format": "pcm16",
        "output_audio_format": "pcm16",
        "tools": [
            {
                "type": "retrieval",
                "function": {
                    "description": "本知识库可以回答用户十万个为什么相关问题",
                    "options": {
                        "vector_store_id": "164643690285936640",
                        "prompt_template": "从文档{{knowledge}}中找到问题{{query}}的答案。根据文档内容中的语句找到答案, 如果文档中没用答案则告诉用户找不到"
                    }
                }
            },
            {
                "type": "retrieval",
                "function": {
                    "description": "本知识库可以回答用户redis安装等相关问题",
                    "options": {
                        "vector_store_id": "164643837904470016",
                        "prompt_template": "从文档{{knowledge}}中找到问题{{query}}的答案。根据文档内容中的语句找到答案, 如果文档中没用答案则告诉用户找不到"
                    }
                }
            }
        ],
        "turn_detection": {
            "type": "server_vad",
            "prefix_padding_ms": 500
        }
    }
}
```

### 音频内容追加

type: `input_audio_buffer.append`

发送此事件以将音频字节追加到输入音频缓冲区，服务器不会向此事件发送确认响应，在 ServerVAD 模式下会触发大模型推理。

* `audio` `string`<br />编码的音频 Base64编码字节。必须采用会话配置中 input\_audio\_format 字段指定的格式。

**Sample**

```json theme={null}
{
    "event_id": "event_abc",
    "type": "input_audio_buffer.append",
    "audio": "Base64EncodedAudioData"
}
```

### 音频内容提交

type: `input_audio_buffer.commit`

发送此事件以提交用户输入音频缓冲区进行推理，将在对话中创建新的用户消息项，服务器将使用 `input_audio_buffer.committed` 事件进行响应。如果输入音频缓冲区为空，则此事件将产生错误。

**Sample**

```json theme={null}
{
    "event_id": "event_abc",
    "type": "input_audio_buffer.commit"
}
```

### 音频内容清理

type:  `input_audio_buffer.clear`

发送此事件以提交用户输入音频缓冲区进行清理，服务器将使用 input\_audio\_buffer.cleared 事件进行响应。

**Sample**

```json theme={null}
{
    "event_id": "event_abc",
    "type": "input_audio_buffer.clear"
}
```

### 会话消息添加

type:  `conversation.item.create`

将新会话添加到会话上下文中，包括消息、函数调用和函数调用响应。此事件既可用于填充对话的 "历史记录" ，也可用于在途中添加新消息项，但当前限制是它无法填充 Assistant 音频消息。如果成功，服务器将使用 `conversation.item.created` 事件进行响应，否则将发送错误事件。

* `previous_item_id` `string`<br />前一项消息的 ID

* `content` `string`<br />消息的内容，适用于消息项，见消息参数

**Sample**

```json theme={null}
{
    "event_id": "event_abc",
    "type": "conversation.item.create",
    "item": {
        "id": "msg_001",
        "type": "message",
        "role": "user",
        "content": [
            {
                "type": "input_text",
                "text": "你好"
            }
        ]
    }
}
```

### 会话消息删除

type:  `conversation.item.delete`

当您想从对话历史记录中删除任何 item 时，发送此事件。服务器将使用 conversation.item.deleted 事件进行响应，当该 item 在对话历史记录中不存在，服务器将响应错误。

* `item_id` `string`<br />想删除的消息 ID

**Sample**

```json theme={null}
{
    "event_id": "event_abc",
    "type": "conversation.item.delete",
    "item_id": "msg_003"
}
```

### 推理提交

type:  `response.create`

此事件指示服务器创建 Response，这意味着触发模型推理，服务端会返回

**Sample**

```json theme={null}
{
    "event_id": "event_abc",
    "type": "response.create"
}
```

### 推理取消

type:  `response.cancel`

将此事件发送以取消正在进行的响应。服务器将返回一个 response.cancelled 事件或在没有可取消的响应时返回一个错误。

```json theme={null}
{
    "event_id": "event_abc",
    "type": "response.cancel"
}
```

## Server Event 列表

### 错误事件

type:  `error`

服务器执行过程中发生错误时返回，这可能是客户端问题或服务器问题，会话会继续保留。

* `type` `string`<br />错误类型（例如: "invalid\_request\_error"、"server\_error"）。

* `code` `string`<br />错误代码（如果有）。

* `message` `string`<br />可解释的错误消息。

* `event_id` `string`<br />导致错误的 client 事件的 event\_id（如果适用）。

**Sample**

```json theme={null}
{
    "event_id": "event_bcd",
    "type": "error",
    "error": {
        "type": "invalid_request_error",
        "code": "invalid_param",
        "message": "音频内容不完整",
        "event_id": "event_567"
    }
}
```

### 创建 session 响应

type:  `session.created`

创建 Session 时返回。当新连接建立为第一个服务器事件时自动发出。此事件将包含默认的 Session 配置。

* `modalities` `array<string>`<br />模型可以使用的模态集。 固定为 `["text", "audio"]`

* `instructions` `string`<br />默认系统指令（即 system message）附加到 model 调用之前。此字段允许客户端指导模型获得所需的响应。该模型可以被指导回答内容和格式（例如，"要非常简洁"、"行为友好"、"这里有好回应的示例"）和音频行为（例如，"快声说"、"在你的声音中注入情感"、"经常大笑"）。不能保证模型会遵循这些说明，但它们会为模型提供有关所需行为的指导。

* `voice` `string`<br />生成时使用的音色信息，支持官方音色，未来会支持自定义音色。

* `input_audio_format` `string`<br />输入音频的格式。 当前仅支持 `pcm16`

* `output_audio_format` `string`<br />输入音频的格式。 当前仅支持 `pcm16`

**sample**

```json theme={null}
{
    "event_id": "event_def",
    "type": "session.created",
    "session": {
        "id": "sess_001",
        "object": "realtime.session",
        "model": "step-1o-audio",
        "modalities": ["text", "audio"],
        "instructions": "你是由阶跃星辰提供的AI聊天助手，你擅长中文，英文，以及多种其他语言的对话。",
        "voice": "linjiajiejie",
        "input_audio_format": "pcm16",
        "output_audio_format": "pcm16",
        "max_response_output_tokens": "4096"
    }
}
```

### 更新 session 响应

type:  `session.updated`

更新 Session 时返回。当新连接建立为第一个服务器事件时自动发出。此事件将包含默认的 Session 配置。

**Sample**

```json theme={null}
{
    "event_id": "event_def",
    "type": "session.created",
    "session": {
        "modalities": ["text", "audio"],
        "instructions": "你是由阶跃星辰提供的AI聊天助手，你擅长中文，英文，以及多种其他语言的对话。",
        "voice": "wenrounansheng",
        "input_audio_format": "pcm16",
        "output_audio_format": "pcm16",
        "max_response_output_tokens": 4096
    }
}
```

### 音频输入激活开始 （VAD）

type:  `input_audio_buffer.speech_started`

输入音频的人声有效输入开始事件通知，一般用于打断场景。

* `audio_start_ms` `string`<br />音频的开始时间

* `item_id` `string`<br />项目的 ID。

**Sample**

```json theme={null}
{
    "event_id": "event_bcd",
    "type": "input_audio_buffer.speech_started",
    "audio_start_ms": 1000,
    "item_id": "msg_003"
}
```

### 音频输入激活结束 （VAD）

type:  `input_audio_buffer.speech_stopped`

输入音频的人声有效输入结束事件通知。

* `response_id` `string`<br />回复 id。

* `item_id` `string`<br />项目的 ID。

**Sample**

```json theme={null}
{
    "event_id": "event_1718",
    "type": "input_audio_buffer.speech_stopped",
    "audio_end_ms": 2000,
    "item_id": "msg_003"
}
```

### 音频内容流式返回

type:  `response.audio.delta`

更新模型生成的音频时返回。

* `response_id` `string`<br />回复 id。

* `item_id` `string`<br />项目的 ID。

* `output_index` `int`<br />响应中输出项的索引。

* `delta` `string`<br />Base64 编码的音频数据增量，音频格式同 session 创建的 output\_audio\_format

**Sample**

```json theme={null}
{
    "event_id": "event_bcd",
    "type": "response.audio.delta",
    "item_id": "msg_008",
    "delta": "Base64EncodedAudioDelta"
}
```

### 音频内容流式结束

type:  `response.audio.done`

在模型生成的音频完成时返回。当 Response 中断、不完整或取消时也会发出.

* `response_id` `string`<br />回复 id。

* `item_id` `string`<br />项目的 ID。

```json theme={null}
{
    "event_id": "event_bcd",
    "type": "response.audio.done",
    "response_id": "traceid",
    "item_id": "msg_008"
}
```

### 音频内容文字流返回

type:  `response.audio_transcript.delta`

当客户端提交输入音频缓冲区时返回。

* `response_id` `string`<br />回复 id。

* `item_id` `string`<br />项目的 ID。

* `output_index` `int`<br />响应中输出项的索引。

* `delta` `string`<br />转录增量。

```json theme={null}
{
    "event_id": "event_bcd",
    "type": "response.audio_transcript.delta",
    "item_id": "msg_002",
    "output_index": 0,
    "delta": "Hello, how can I a"
}
```

### 音频内容文字流结束

type:  `response.audio_transcript.done`

当模型生成的音频输出转录完成流式处理时返回。当 Response 中断、不完整或取消时也会发出。

* `response_id` `string`<br />回复 id。

* `item_id` `string`<br />项目的 ID。

* `output_index` `int`<br />响应中输出项的索引。

* `transcript` `string`<br />音频的完整文字内容。

```json theme={null}
{
    "event_id": "event_4748",
    "type": "response.audio_transcript.done",
    "response_id": "resp_001",
    "item_id": "msg_008",
    "content_index": 0,
    "transcript": "Hello, how can I assist you today?"
}
```

### 推理思考流返回

type:  `response.thinking.delta`

当模型生成推理思考过程时流式返回。

* `response_id` `string`<br />回复 id。

* `item_id` `string`<br />项目的 ID。

* `output_index` `int`<br />响应中输出项的索引。

* `content_index` `int`<br />项目内容数组中内容部分的索引。

* `delta` `string`<br />思考过程的增量内容。

```json theme={null}
{
    "event_id": "event_thinking_delta",
    "type": "response.thinking.delta",
    "response_id": "resp_001",
    "item_id": "msg_008",
    "output_index": 0,
    "content_index": 0,
    "delta": "通过分析..."
}
```

### 推理思考流结束

type:  `response.thinking.done`

当模型推理思考过程结束时返回。

* `response_id` `string`<br />回复 id。

* `item_id` `string`<br />项目的 ID。

* `output_index` `int`<br />响应中输出项的索引。

* `content_index` `int`<br />项目内容数组中内容部分的索引。

* `thinking` `string`<br />思考过程的完整文字内容。

```json theme={null}
{
    "event_id": "event_thinking_done",
    "type": "response.thinking.done",
    "response_id": "resp_001",
    "item_id": "msg_008",
    "output_index": 0,
    "content_index": 0,
    "thinking": "完整思考过程..."
}
```

### 回复文字内容流式返回

type: `response.text.delta`

在模型生成的文本输出时返回。

* `response_id` `string`<br />回复 id。

* `item_id` `string`<br />响应项目的 ID。

* `output_index` `int`<br />响应中输出项的索引。

* `delta` `string`<br />生成的文本内容增量。

* `content_index` `int`<br />包含文本内容部分的索引。

```json theme={null}
{
    "event_id": "event_4950",
    "type": "response.text.delta",
    "response_id": "resp_001",
    "item_id": "msg_009",
    "output_index": 0,
    "content_index": 0,
    "delta": "Sure, "
}
```

### 回复文字内容流式结束

type: `response.text.done`

在模型生成的文本输出完成时返回。当 Response 中断、不完整或取消时也会发出。

* `response_id` `string`<br />回复 id。

* `item_id` `string`<br />响应项目的 ID。

* `output_index` `int`<br />响应中输出项的索引。

* `text` `string`<br />生成的完整文本内容。

* `content_index` `int`<br />包含文本内容部分的索引。

```json theme={null}
{
    "event_id": "event_5152",
    "type": "response.text.done",
    "response_id": "resp_001",
    "item_id": "msg_009",
    "output_index": 0,
    "content_index": 0,
    "text": "Sure, I can help with that."
}
```

### 会话消息创建响应

type: `conversation.item.created`

创建对话项时返回，服务器正在生成一个 Response，如果成功，将生成一个或两个 Item，其类型为 message；

* `id` `string`<br />消息的唯一 ID，非必需，如果客户端未提供，服务器将自动生成一个。

* `type` `string`<br />项目的类型，通常为 message

* `role` `string`<br />消息发送者的角色（用户、助手、系统），仅适用于消息项。

* `status` `string`<br />项目的状态 （completed， incomplete）。这些对会话没有影响。

* `content` `string`<br />消息的内容，适用于消息项。

```json theme={null}
{
	"event_id": "event_bcd",
	"type": "conversation.item.created",
	"previous_item_id": "msg_001",
	"item": {
		"id": "msg_002",
		"object": "realtime.item",
		"type": "message",
		"status": "completed",
		"role": "user",
		"content": [{
			"type": "input_text",
			"transcript": "你好"
		}]
	}
}
```

### 会话消息删除响应

type: `conversation.item.deleted`

当客户端使用 `conversation.item.delete` 事件删除对话中的项目时返回。此事件用于将服务器对对话历史与客户端的进行同步。

* `item_id` `string`<br />会话消息 ID

```json theme={null}
{
    "event_id": "event_bcd",
    "type": "conversation.item.deleted",
    "item_id": "msg_001"
}
```

### 用户提交音频的转录结果完成

type: `conversation.item.input_audio_transcription.completed`

此事件是写入用户音频缓冲区的用户音频的音频转录（ASR）的输出。当客户端 commit 缓存区的音频，或在 server\_vad 模式下缓冲区音频被提交时，转录开始。转录随响应创建异步运行，因此此事件可能发生在响应事件之前或之后。

* `item_id` `string`<br />会话消息 ID
* `content_index` `int`<br />包含音频内容部分的索引。
* `transcript` `string`<br />音频的完整文字内容。

```json theme={null}
{
    "event_id": "event_2122",
    "type": "conversation.item.input_audio_transcription.completed",
    "item_id": "msg_003",
    "content_index": 0,
    "transcript": "你好"
}
```

### 音频内容提交响应

type: `input_audio_buffer.committed`

当客户端提交输入音频缓冲区时返回。

* `previous_item_id` `string`<br />会话前一个消息 ID

* `item_id` `string`<br />会话消息 ID

```json theme={null}
{
    "event_id": "event_bcd",
    "type": "input_audio_buffer.committed",
    "previous_item_id": "msg_001",
    "item_id": "msg_002"
}
```

### 音频内容删除响应

type: `input_audio_buffer.cleared`

当客户端使用 input\_audio\_buffer.clear 事件清除输入音频缓冲区时返回。

```json theme={null}

{
    "event_id": "event_1314",
    "type": "input_audio_buffer.cleared"
}
```

### 推理有输出项目产生

type: `response.output_item.added`

在响应生成期间创建新项目时返回。

* `output_index` `int`<br />响应中输出项的索引。

* `item` `object`<br />输出项的对象。

  * `id` `string`<br />项目的 ID。

  * object
    <br />   始终为 `realtime.item`

  * type
    <br />  item 类型,目前仅支持 `message`

  * status
    <br />  项目状态,可选值为 `completed`, `incomplete`,`in_progress`

  * role
    <br />   该项目的 role， 仅适用于 message 项目，可选值为 `user`, `assistant`, `system`

  * content
    <br />   消息的内容，适用于 message 项目。
    <br />   role=system 的消息项仅支持 input\_text 内容
    <br />   role=user 的消息项支持 input\_text 和 input\_audio 内容
    <br />   role=assistant 角色支持文本内容的消息项。

```json theme={null}
{
    "event_id": "event_3334",
    "type": "response.output_item.added",
    "response_id": "resp_001",
    "output_index": 0,
    "item": {
        "id": "msg_007",
        "object": "realtime.item",
        "type": "message",
        "status": "in_progress",
        "role": "assistant",
        "content": []
    }
}
```

### 推理有输出项目完成

type: `response.output_item.done`

当一个项目完成时返回。当一个响应处于以下状态时 `interrupted`, `incomplete`,  `cancelled` 也会发出。

* `output_index` `int`<br />响应中输出项的索引。

* `item` `object`<br />输出项的对象。

  * `id` `string`<br />项目的 ID。

  * object
    <br />   始终为 `realtime.item`

  * type
    <br />  item 类型,目前仅支持 `message`

  * status
    <br />  项目状态,可选值为 `completed`, `incomplete`,`in_progress`

  * role
    <br />   该项目的 role， 仅适用于 message 项目，可选值为 `user`, `assistant`, `system`

  * content
    <br />   消息的内容，适用于 message 项目。
    <br />   role=system 的消息项仅支持 input\_text 内容
    <br />   role=user 的消息项支持 input\_text 和 input\_audio 内容
    <br />   role=assistant 角色支持文本内容的消息项。

```json theme={null}
{
    "event_id": "event_3536",
    "type": "response.output_item.done",
    "response_id": "resp_001",
    "output_index": 0,
    "item": {
        "id": "msg_007",
        "object": "realtime.item",
        "type": "message",
        "status": "completed",
        "role": "assistant",
        "content": [
            {
                "type": "text",
                "text": "Sure, I can help with that."
            }
        ]
    }
}
```

### 推理的输出项目中有新的内容生成

type: `response.content_part.added`

在响应生成期间，当向 assistant 消息项添加新的内容部分时返回。

* `response_id` `string`<br />响应 id
* `item_id` `string`<br />对应的项目 id
* `content_index` `int`<br />项目内容数组中内容部分的索引。
* `output_index` `int`<br />响应中输出项的索引。
* `part` `object`
  * `type` `string`<br />类型，支持 `text`, `audio` 或 `thinking`
  * `audio` `string`<br />audio: base64 encode 后的音频数据（当 type=audio 时有这个字段）
  * `text` `string`<br />生成的文本内容（当 type=text 时有这个字段）
  * `thinking` `string`<br />生成的推理思考内容（当 type=thinking 时有这个字段）
  * `transcript` `string`<br />音频的转录内容 （当 type=audio 时有这个字段）

```json theme={null}
{
    "event_id": "event_3738",
    "type": "response.content_part.added",
    "response_id": "resp_001",
    "item_id": "msg_007",
    "output_index": 0,
    "content_index": 0,
    "part": {
        "type": "text",
        "text": ""
    }
}
```

### 推理的输出项目中有新的内容完成

type: `response.content_part.done`

当一个 `content_part` 完成时返回。当它对应的响应处于以下状态时 `interrupted`, `incomplete`, or `cancelled` 也会发出。

* `response_id` `string`<br />响应 id
* `item_id` `string`<br />对应的项目 id
* `content_index` `int`<br />项目内容数组中内容部分的索引。
* `output_index` `int`<br />响应中输出项的索引。
* `part` `object`
  * `type` `string`<br />类型，支持 `text`, `audio` 或 `thinking`
  * `audio` `string`<br />audio: base64 encode 后的音频数据（当 type=audio 时有这个字段）
  * `text` `string`<br />生成的文本内容（当 type=text 时有这个字段）
  * `thinking` `string`<br />生成的推理思考内容（当 type=thinking 时有这个字段）
  * `transcript` `string`<br />音频的转录内容 （当 type=audio 时有这个字段）

```json theme={null}
{
    "event_id": "event_3940",
    "type": "response.content_part.done",
    "response_id": "resp_001",
    "item_id": "msg_007",
    "output_index": 0,
    "content_index": 0,
    "part": {
        "type": "text",
        "text": "Sure, I can help with that."
    }
}
```

### 推理创建响应

type: `response.created`

创建新 Response 时返回。响应创建的第一个事件，其中响应的初始状态为 in\_progress。

* `id` `string`<br />响应的唯一 ID。

* `object` `string`<br />对象类型必须为:realtime.response

* `status` `string`<br />响应的最终状态 （completed， cancelled， failed， incomplete）。

* `output` `list`<br />响应生成的输出项列表。

```json theme={null}
{
    "event_id": "event_3132",
    "type": "response.done",
    "response": {
        "id": "resp_001",
        "object": "realtime.response",
        "status": "completed",
        "status_details": null,
        "output": [
            {
                "id": "msg_006",
                "object": "realtime.item",
                "type": "message",
                "status": "completed",
                "role": "assistant",
                "content": [
                    {
                        "type": "text",
                        "text": "Sure, how can I assist you today?"
                    }
                ]
            }
        ]
    }
}
```

### 推理结束响应

type:  `response.done`

当 Response 完成流式处理时返回。始终发出，无论最终状态如何。response.done 事件中包含的 Response 对象将包括 Response 中的所有输出 Item，但会省略原始音频数据。

* `id` `string`<br />响应的唯一 ID。

* `object` `string`<br />对象类型必须为:realtime.response

* `status` `string`<br />响应的最终状态 （completed， cancelled， failed， incomplete）。

* `output` `list`<br />响应生成的输出项列表。

```json theme={null}
{
    "event_id": "event_bcd",
    "type": "response.done",
    "response": {
        "id": "resp_001",
        "object": "realtime.response",
        "status": "completed",
        "status_details": null,
        "output": [
            {
                "id": "msg_006",
                "object": "realtime.item",
                "type": "message",
                "status": "completed",
                "role": "assistant",
                "content": [
                    {
                        "type": "text",
                        "text": "你好"
                    }
                ]
            }
        ]
    }
}
```
