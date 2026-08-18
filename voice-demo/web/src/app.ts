import { MicrophoneStream, PcmPlayer } from "./audio";
import {
  parseServerEvent,
  type ClientEvent,
  type SessionConfig,
  type ServerEvent,
} from "./types";

// 页面元素集合，避免到处 getElementById。
interface Elements {
  status: HTMLElement;
  log: HTMLElement;
  userTranscript: HTMLElement;
  aiTranscript: HTMLElement;
  connectBtn: HTMLButtonElement;
  vadToggle: HTMLInputElement;
  commitBtn: HTMLButtonElement;
  clearBtn: HTMLButtonElement;
  textInput: HTMLInputElement;
  sendBtn: HTMLButtonElement;
}

class VoiceSession {
  private socket: WebSocket | null = null;
  private mic: MicrophoneStream | null = null;
  private player = new PcmPlayer();
  private aiTranscript = "";

  constructor(private readonly els: Elements) {}

  async toggleConnect(): Promise<void> {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      await this.disconnect();
    } else {
      await this.connect();
    }
  }

  sendText(): void {
    const text = this.els.textInput.value.trim();
    if (!text || !this.socket || this.socket.readyState !== WebSocket.OPEN) {
      return;
    }
    this.send({
      type: "conversation.item.create",
      item: {
        type: "message",
        role: "user",
        content: [{ type: "input_text", text }],
      },
    });
    this.send({ type: "response.create" });
    this.els.userTranscript.textContent = text;
    this.aiTranscript = "";
    this.els.aiTranscript.textContent = "";
    this.els.textInput.value = "";
  }

  async commitAudio(): Promise<void> {
    if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
      return;
    }
    this.send({ type: "input_audio_buffer.commit" });
    this.send({ type: "response.create" });
    this.log("已提交音频，等待回复");
  }

  clearAudio(): void {
    if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
      return;
    }
    this.send({ type: "input_audio_buffer.clear" });
    this.player.clear();
  }

  private async connect(): Promise<void> {
    await this.player.start();

    const protocol = window.location.protocol === "https:" ? "wss://" : "ws://";
    const socket = new WebSocket(`${protocol}${window.location.host}/ws`);
    this.socket = socket;

    socket.onopen = () => this.onOpen();
    socket.onmessage = (event: MessageEvent<string>) => this.onMessage(event.data);
    socket.onclose = () => this.onClose();
    socket.onerror = () => this.log("WebSocket 连接出错");
  }

  private async disconnect(): Promise<void> {
    if (this.mic) {
      await this.mic.stop();
      this.mic = null;
    }
    await this.player.stop();
    this.socket?.close();
    this.socket = null;
    this.setStatus("已断开");
  }

  private async onOpen(): Promise<void> {
    this.setStatus("已连接，开始说话");
    this.send(this.buildSessionUpdate());

    this.mic = new MicrophoneStream((base64) => {
      this.send({ type: "input_audio_buffer.append", audio: base64 });
    });
    try {
      await this.mic.start();
    } catch (error) {
      this.log(`麦克风不可用: ${String(error)}`);
    }

    this.els.connectBtn.textContent = "结束对话";
    this.refreshManualControls();
    this.log("已连接并开启麦克风");
  }

  private onClose(): void {
    this.setStatus("连接已断开");
    this.els.connectBtn.textContent = "开始对话";
    this.refreshManualControls();
  }

  private onMessage(raw: string): void {
    let event: ServerEvent;
    try {
      event = parseServerEvent(raw);
    } catch {
      this.log("收到无法解析的事件");
      return;
    }

    switch (event.type) {
      case "session.created":
      case "session.updated":
        this.setStatus("会话就绪");
        break;
      case "input_audio_buffer.speech_started":
        // 用户开始说话，打断 AI 正在播放的内容。
        this.player.clear();
        this.aiTranscript = "";
        this.els.aiTranscript.textContent = "";
        break;
      case "conversation.item.input_audio_transcription.completed":
        this.els.userTranscript.textContent = event.transcript;
        break;
      case "response.audio.delta":
        this.player.append(event.delta);
        break;
      case "response.audio_transcript.delta":
        this.aiTranscript += event.delta;
        this.els.aiTranscript.textContent = this.aiTranscript;
        break;
      case "response.audio_transcript.done":
        this.aiTranscript = event.transcript;
        this.els.aiTranscript.textContent = event.transcript;
        break;
      case "response.text.delta":
        this.aiTranscript += event.delta;
        this.els.aiTranscript.textContent = this.aiTranscript;
        break;
      case "response.text.done":
        this.aiTranscript = event.text;
        this.els.aiTranscript.textContent = event.text;
        break;
      case "response.done":
        this.log(`本轮结束: ${event.response.status}`);
        break;
      case "error":
        this.log(`错误: ${event.error.message ?? event.error.code ?? "未知"}`);
        break;
    }
  }

  private buildSessionUpdate(): ClientEvent {
    const session: SessionConfig = {
      modalities: ["text", "audio"],
      instructions: "你是用户的中文实时语音助手，任何时候都使用中文回答，回答尽量简洁自然。",
      voice: "linjiajiejie",
      input_audio_format: "pcm16",
      output_audio_format: "pcm16",
    };

    if (this.els.vadToggle.checked) {
      session.turn_detection = { type: "server_vad", prefix_padding_ms: 500 };
    } else {
      session.turn_detection = null;
    }

    return { type: "session.update", session };
  }

  private refreshManualControls(): void {
    const connected = Boolean(this.socket && this.socket.readyState === WebSocket.OPEN);
    const vadOff = connected && !this.els.vadToggle.checked;
    this.els.commitBtn.disabled = !vadOff;
    this.els.clearBtn.disabled = !vadOff;
    this.els.sendBtn.disabled = !connected;
    this.els.textInput.disabled = !connected;
  }

  private send(event: ClientEvent): void {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      this.socket.send(JSON.stringify(event));
    }
  }

  private setStatus(text: string): void {
    this.els.status.textContent = text;
  }

  private log(text: string): void {
    const line = document.createElement("div");
    line.textContent = `[${new Date().toLocaleTimeString()}] ${text}`;
    this.els.log.prepend(line);
  }
}

function byId<T extends HTMLElement>(id: string): T {
  const element = document.getElementById(id);
  if (!element) {
    throw new Error(`缺少页面元素 #${id}`);
  }
  return element as T;
}

function main(): void {
  const els: Elements = {
    status: byId("status"),
    log: byId("log"),
    userTranscript: byId("user-transcript"),
    aiTranscript: byId("ai-transcript"),
    connectBtn: byId("connect-btn"),
    vadToggle: byId<HTMLInputElement>("vad-toggle"),
    commitBtn: byId("commit-btn"),
    clearBtn: byId("clear-btn"),
    textInput: byId<HTMLInputElement>("text-input"),
    sendBtn: byId("send-btn"),
  };

  const session = new VoiceSession(els);

  els.connectBtn.addEventListener("click", () => void session.toggleConnect());
  els.sendBtn.addEventListener("click", () => session.sendText());
  els.textInput.addEventListener("keydown", (event) => {
    if (event.key === "Enter") {
      session.sendText();
    }
  });
  els.commitBtn.addEventListener("click", () => void session.commitAudio());
  els.clearBtn.addEventListener("click", () => session.clearAudio());
  els.vadToggle.addEventListener("change", () => els.vadToggle.blur());
}

window.addEventListener("DOMContentLoaded", main);
