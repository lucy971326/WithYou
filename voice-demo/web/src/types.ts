// 与 Go 端 internal/realtime/types.go 保持一致的事件契约。

export type EventType =
  // 客户端事件
  | "session.update"
  | "input_audio_buffer.append"
  | "input_audio_buffer.commit"
  | "input_audio_buffer.clear"
  | "conversation.item.create"
  | "response.create"
  // 服务端事件
  | "session.created"
  | "session.updated"
  | "error"
  | "input_audio_buffer.speech_started"
  | "input_audio_buffer.speech_stopped"
  | "conversation.item.input_audio_transcription.completed"
  | "response.created"
  | "response.audio.delta"
  | "response.audio.done"
  | "response.audio_transcript.delta"
  | "response.audio_transcript.done"
  | "response.text.delta"
  | "response.text.done"
  | "response.done";

export interface TurnDetection {
  type: "server_vad";
  prefix_padding_ms?: number;
}

export interface SessionConfig {
  modalities?: string[];
  instructions?: string;
  voice?: string;
  input_audio_format?: string;
  output_audio_format?: string;
  turn_detection?: TurnDetection | null;
}

export interface SessionInfo {
  id?: string;
  model?: string;
  voice?: string;
  input_audio_format?: string;
  output_audio_format?: string;
  modalities?: string[];
}

// 客户端要发送的事件。
export interface SessionUpdateEvent {
  event_id?: string;
  type: "session.update";
  session: SessionConfig;
}

export interface AppendAudioEvent {
  event_id?: string;
  type: "input_audio_buffer.append";
  audio: string;
}

export interface CommitAudioEvent {
  event_id?: string;
  type: "input_audio_buffer.commit";
}

export interface ClearAudioEvent {
  event_id?: string;
  type: "input_audio_buffer.clear";
}

export interface ItemCreateEvent {
  event_id?: string;
  type: "conversation.item.create";
  item: {
    type: "message";
    role: "user";
    content: Array<{ type: "input_text"; text: string }>;
  };
}

export interface ResponseCreateEvent {
  event_id?: string;
  type: "response.create";
}

export type ClientEvent =
  | SessionUpdateEvent
  | AppendAudioEvent
  | CommitAudioEvent
  | ClearAudioEvent
  | ItemCreateEvent
  | ResponseCreateEvent;

// 服务端发送的事件。
export interface SessionCreatedEvent {
  type: "session.created";
  session: SessionInfo;
}

export interface SessionUpdatedEvent {
  type: "session.updated";
  session: SessionInfo;
}

export interface ErrorEvent {
  type: "error";
  error: { type?: string; code?: string; message?: string };
}

export interface SpeechStartedEvent {
  type: "input_audio_buffer.speech_started";
}

export interface SpeechStoppedEvent {
  type: "input_audio_buffer.speech_stopped";
}

export interface InputTranscriptionCompletedEvent {
  type: "conversation.item.input_audio_transcription.completed";
  transcript: string;
}

export interface AudioDeltaEvent {
  type: "response.audio.delta";
  item_id?: string;
  delta: string;
}

export interface AudioDoneEvent {
  type: "response.audio.done";
}

export interface TranscriptDeltaEvent {
  type: "response.audio_transcript.delta";
  delta: string;
}

export interface TranscriptDoneEvent {
  type: "response.audio_transcript.done";
  transcript: string;
}

export interface TextDeltaEvent {
  type: "response.text.delta";
  delta: string;
}

export interface TextDoneEvent {
  type: "response.text.done";
  text: string;
}

export interface ResponseDoneEvent {
  type: "response.done";
  response: { status: string };
}

export type ServerEvent =
  | SessionCreatedEvent
  | SessionUpdatedEvent
  | ErrorEvent
  | SpeechStartedEvent
  | SpeechStoppedEvent
  | InputTranscriptionCompletedEvent
  | AudioDeltaEvent
  | AudioDoneEvent
  | TranscriptDeltaEvent
  | TranscriptDoneEvent
  | TextDeltaEvent
  | TextDoneEvent
  | ResponseDoneEvent;

export function parseServerEvent(raw: string): ServerEvent {
  return JSON.parse(raw) as ServerEvent;
}
