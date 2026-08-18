package realtime

import (
	"encoding/json"
	"fmt"
)

// EventType 标识 Realtime 事件类型，JSON 中对应 "type" 字段。
type EventType string

// 客户端 -> 服务端事件。
const (
	EventSessionUpdate          EventType = "session.update"
	EventInputAudioBufferAppend EventType = "input_audio_buffer.append"
	EventInputAudioBufferCommit EventType = "input_audio_buffer.commit"
	EventInputAudioBufferClear  EventType = "input_audio_buffer.clear"
	EventConversationItemCreate EventType = "conversation.item.create"
	EventResponseCreate         EventType = "response.create"
)

// 服务端 -> 客户端事件。
const (
	EventSessionCreated                   EventType = "session.created"
	EventSessionUpdated                   EventType = "session.updated"
	EventError                            EventType = "error"
	EventInputAudioBufferSpeechStarted    EventType = "input_audio_buffer.speech_started"
	EventInputAudioBufferSpeechStopped    EventType = "input_audio_buffer.speech_stopped"
	EventInputAudioBufferCommitted        EventType = "input_audio_buffer.committed"
	EventConversationItemCreated          EventType = "conversation.item.created"
	EventInputAudioTranscriptionCompleted EventType = "conversation.item.input_audio_transcription.completed"
	EventResponseCreated                  EventType = "response.created"
	EventResponseAudioDelta               EventType = "response.audio.delta"
	EventResponseAudioDone                EventType = "response.audio.done"
	EventResponseAudioTranscriptDelta     EventType = "response.audio_transcript.delta"
	EventResponseAudioTranscriptDone      EventType = "response.audio_transcript.done"
	EventResponseTextDelta                EventType = "response.text.delta"
	EventResponseTextDone                 EventType = "response.text.done"
	EventResponseDone                     EventType = "response.done"
)

// TurnDetection 服务端 VAD 的配置。
type TurnDetection struct {
	Type                     string `json:"type"`
	PrefixPaddingMs          int    `json:"prefix_padding_ms,omitempty"`
	SilenceDurationMs        int    `json:"silence_duration_ms,omitempty"`
	EnergyAwakenessThreshold int    `json:"energy_awakeness_threshold,omitempty"`
}

// SessionConfig 会话配置，用于 session.update。
type SessionConfig struct {
	Modalities        []string       `json:"modalities,omitempty"`
	Instructions      string         `json:"instructions,omitempty"`
	Voice             string         `json:"voice,omitempty"`
	InputAudioFormat  string         `json:"input_audio_format,omitempty"`
	OutputAudioFormat string         `json:"output_audio_format,omitempty"`
	TurnDetection     *TurnDetection `json:"turn_detection,omitempty"`
}

// SessionUpdateEvent 创建 / 更新会话。
type SessionUpdateEvent struct {
	EventID string        `json:"event_id,omitempty"`
	Type    EventType     `json:"type"`
	Session SessionConfig `json:"session"`
}

// AppendAudioEvent 追加一块 base64 PCM16 音频。
type AppendAudioEvent struct {
	EventID string    `json:"event_id,omitempty"`
	Type    EventType `json:"type"`
	Audio   string    `json:"audio"`
}

// CommitAudioEvent 手动提交输入音频（关 VAD 时使用）。
type CommitAudioEvent struct {
	EventID string    `json:"event_id,omitempty"`
	Type    EventType `json:"type"`
}

// ClearAudioEvent 清空输入音频缓冲（关 VAD 时新一轮输入前）。
type ClearAudioEvent struct {
	EventID string    `json:"event_id,omitempty"`
	Type    EventType `json:"type"`
}

// ConversationContent 消息内容块。
type ConversationContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// ConversationItemCreateEvent 注入一条文本消息。
type ConversationItemCreateEvent struct {
	EventID string           `json:"event_id,omitempty"`
	Type    EventType        `json:"type"`
	Item    ConversationItem `json:"item"`
}

// ConversationItem 会话中的一条消息。
type ConversationItem struct {
	Type    string                `json:"type"`
	Role    string                `json:"role"`
	Content []ConversationContent `json:"content,omitempty"`
}

// ResponseCreateEvent 触发模型推理。
type ResponseCreateEvent struct {
	EventID string    `json:"event_id,omitempty"`
	Type    EventType `json:"type"`
}

// ServerEvent 是所有服务端事件的统一接口，供 handler 做类型断言。
type ServerEvent interface {
	EventType() EventType
}

// SessionInfo 服务端返回的会话信息。
type SessionInfo struct {
	ID                string   `json:"id,omitempty"`
	Model             string   `json:"model,omitempty"`
	Modalities        []string `json:"modalities,omitempty"`
	Instructions      string   `json:"instructions,omitempty"`
	Voice             string   `json:"voice,omitempty"`
	InputAudioFormat  string   `json:"input_audio_format,omitempty"`
	OutputAudioFormat string   `json:"output_audio_format,omitempty"`
}

// SessionCreatedEvent 连接建立后的第一个事件。
type SessionCreatedEvent struct {
	Type    EventType   `json:"type"`
	Session SessionInfo `json:"session"`
}

func (SessionCreatedEvent) EventType() EventType { return EventSessionCreated }

// SessionUpdatedEvent 会话配置更新后的回执。
type SessionUpdatedEvent struct {
	Type    EventType   `json:"type"`
	Session SessionInfo `json:"session"`
}

func (SessionUpdatedEvent) EventType() EventType { return EventSessionUpdated }

// ErrorInfo 错误详情。
type ErrorInfo struct {
	Type    string `json:"type,omitempty"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	EventID string `json:"event_id,omitempty"`
}

// ErrorEvent 服务端错误。
type ErrorEvent struct {
	Type  EventType `json:"type"`
	Error ErrorInfo `json:"error"`
}

func (ErrorEvent) EventType() EventType { return EventError }

// AudioDeltaEvent 模型生成的音频增量块（base64）。
type AudioDeltaEvent struct {
	Type   EventType `json:"type"`
	ItemID string    `json:"item_id,omitempty"`
	Delta  string    `json:"delta"`
}

func (AudioDeltaEvent) EventType() EventType { return EventResponseAudioDelta }

// AudioDoneEvent 音频块流结束。
type AudioDoneEvent struct {
	Type       EventType `json:"type"`
	ItemID     string    `json:"item_id,omitempty"`
	ResponseID string    `json:"response_id,omitempty"`
}

func (AudioDoneEvent) EventType() EventType { return EventResponseAudioDone }

// TranscriptionDeltaEvent 转写增量（AI 音频转录）。
type TranscriptionDeltaEvent struct {
	Type   EventType `json:"type"`
	ItemID string    `json:"item_id,omitempty"`
	Delta  string    `json:"delta"`
}

func (TranscriptionDeltaEvent) EventType() EventType { return EventResponseAudioTranscriptDelta }

// TranscriptionDoneEvent 转写完成（AI 音频转录）。
type TranscriptionDoneEvent struct {
	Type       EventType `json:"type"`
	ItemID     string    `json:"item_id,omitempty"`
	Transcript string    `json:"transcript"`
}

func (TranscriptionDoneEvent) EventType() EventType { return EventResponseAudioTranscriptDone }

// TextDeltaEvent 模型文本增量。
type TextDeltaEvent struct {
	Type   EventType `json:"type"`
	ItemID string    `json:"item_id,omitempty"`
	Delta  string    `json:"delta"`
}

func (TextDeltaEvent) EventType() EventType { return EventResponseTextDelta }

// TextDoneEvent 模型文本完成。
type TextDoneEvent struct {
	Type   EventType `json:"type"`
	ItemID string    `json:"item_id,omitempty"`
	Text   string    `json:"text"`
}

func (TextDoneEvent) EventType() EventType { return EventResponseTextDone }

// ResponseDoneEvent 本轮推理结束，包含最终状态。
type ResponseDoneEvent struct {
	Type     EventType      `json:"type"`
	Response ResponseObject `json:"response"`
}

func (ResponseDoneEvent) EventType() EventType { return EventResponseDone }

// ResponseObject 响应对象概要。
type ResponseObject struct {
	ID     string       `json:"id,omitempty"`
	Status string       `json:"status,omitempty"`
	Output []OutputItem `json:"output,omitempty"`
}

// OutputItem 响应中的输出项。
type OutputItem struct {
	ID      string        `json:"id,omitempty"`
	Type    string        `json:"type,omitempty"`
	Status  string        `json:"status,omitempty"`
	Role    string        `json:"role,omitempty"`
	Content []ContentPart `json:"content,omitempty"`
}

// ContentPart 输出项的内容块。
type ContentPart struct {
	Type       string `json:"type,omitempty"`
	Text       string `json:"text,omitempty"`
	Transcript string `json:"transcript,omitempty"`
}

// ParseServerEvent 读取 JSON 中的 type 并分发到对应类型化结构体。
func ParseServerEvent(data []byte) (ServerEvent, error) {
	var header struct {
		Type EventType `json:"type"`
	}
	if err := json.Unmarshal(data, &header); err != nil {
		return nil, fmt.Errorf("解析事件类型失败: %w", err)
	}

	switch header.Type {
	case EventSessionCreated:
		return unmarshalServer[SessionCreatedEvent](data)
	case EventSessionUpdated:
		return unmarshalServer[SessionUpdatedEvent](data)
	case EventError:
		return unmarshalServer[ErrorEvent](data)
	case EventResponseAudioDelta:
		return unmarshalServer[AudioDeltaEvent](data)
	case EventResponseAudioDone:
		return unmarshalServer[AudioDoneEvent](data)
	case EventResponseAudioTranscriptDelta:
		return unmarshalServer[TranscriptionDeltaEvent](data)
	case EventResponseAudioTranscriptDone:
		return unmarshalServer[TranscriptionDoneEvent](data)
	case EventResponseTextDelta:
		return unmarshalServer[TextDeltaEvent](data)
	case EventResponseTextDone:
		return unmarshalServer[TextDoneEvent](data)
	case EventResponseDone:
		return unmarshalServer[ResponseDoneEvent](data)
	default:
		return nil, fmt.Errorf("未支持的服务端事件类型: %s", header.Type)
	}
}

// unmarshalServer 把原始 JSON 解码成某个服务端事件类型。
func unmarshalServer[T ServerEvent](data []byte) (ServerEvent, error) {
	var event T
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, fmt.Errorf("解析 %s 内容失败: %w", event.EventType(), err)
	}
	return event, nil
}
