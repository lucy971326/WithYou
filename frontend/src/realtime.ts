const INPUT_RATE = 24000;
const FRAME_MS = 250;
const LOOKBACK_MS = 500;
const JPEG_MAX_RAW = 192 * 1024;
const JPEG_MAX_B64 = 256 * 1024;
const FRAME_MAX_W = 1280;
const FRAME_MAX_H = 720;

export type SubSegment = {
  start_sec: number;
  end_sec: number;
  beat: string;
  summary: string;
  key_dialogue: string;
  visual_scene: string;
  emotion: string;
  story_so_far: string;
};

export type PlotDoc = {
  title: string;
  overview: {
    grand_summary: string;
    key_characters: string[];
    key_plot_points: string[];
  };
  major_segments: Array<{
    start_sec: number;
    end_sec: number;
    title: string;
    summary: string;
    sub_segments: SubSegment[];
  }>;
};

export type TimelineEvent =
  | { type: "user_delta"; text: string; stash: string }
  | { type: "user_done"; transcript: string }
  | { type: "ai_start" }
  | { type: "ai_delta"; delta: string }
  | { type: "ai_done"; transcript?: string; interrupted?: boolean }
  | { type: "system"; label: string; detail?: string };

type Frame = { at: number; b64: string; raw: number };

export class RealtimeWatch {
  private readonly video: HTMLVideoElement;
  private readonly onStatus: (s: string) => void;
  private readonly onBeat?: (seg: SubSegment) => void;
  private readonly onTimeline?: (ev: TimelineEvent) => void;
  private ws: WebSocket | null = null;
  private plot: PlotDoc | null = null;
  private voice = "Tina";
  private aiDoneEmitted = true;
  private gate = false;
  private audioSent = false;
  private aiTalking = false;
  private lastSeg = -1;
  private frames: Frame[] = [];
  private frameTimer = 0;
  private frameCanvas: HTMLCanvasElement | null = null;
  private captureCtx: AudioContext | null = null;
  private playCtx: AudioContext | null = null;
  private proc: ScriptProcessorNode | null = null;
  private mic: MediaStream | null = null;
  private pcmTail: Float32Array = new Float32Array(0);
  private playAt = 0;
  private sources: AudioBufferSourceNode[] = [];
  private savedVolume = 1;

  constructor(
    video: HTMLVideoElement,
    onStatus: (s: string) => void,
    onBeat?: (seg: SubSegment) => void,
    onTimeline?: (ev: TimelineEvent) => void,
  ) {
    this.video = video;
    this.onStatus = onStatus;
    this.onBeat = onBeat;
    this.onTimeline = onTimeline;
  }

  setVoice(voice: string): void {
    this.voice = voice;
    if (this.gate) {
      // 音频配置发过之后不能再改，换音色只发 voice，不能全量重发 session。
      const ok = this.send({ type: "session.update", session: { voice } });
      if (ok) {
        this.onTimeline?.({ type: "system", label: "音色已切换", detail: voice });
      }
      log("→ session.update voice", voice);
    }
  }

  async start(plot: PlotDoc): Promise<void> {
    this.stop();
    this.plot = plot;
    this.lastSeg = -1;
    log("ws connecting");
    const proto = location.protocol === "https:" ? "wss" : "ws";
    const ws = new WebSocket(`${proto}://${location.host}/api/realtime`);
    this.ws = ws;
    ws.addEventListener("open", () => {
      log("ws open");
      this.onStatus("Realtime 已连上，等 session");
    });
    ws.addEventListener("close", (ev) => {
      log("ws close", ev.code, ev.reason);
      this.gate = false;
      this.onStatus(`Realtime 断开 ${ev.code}`);
    });
    ws.addEventListener("error", () => {
      log("ws error");
    });
    ws.addEventListener("message", (ev) => {
      if (typeof ev.data !== "string") {
        log("drop non-text from server");
        return;
      }
      void this.onServer(ev.data);
    });
  }

  stop(): void {
    this.gate = false;
    this.audioSent = false;
    this.aiTalking = false;
    if (this.frameTimer) {
      window.clearInterval(this.frameTimer);
      this.frameTimer = 0;
    }
    this.proc?.disconnect();
    this.proc = null;
    this.mic?.getTracks().forEach((t) => t.stop());
    this.mic = null;
    void this.captureCtx?.close();
    this.captureCtx = null;
    this.stopPlayback();
    void this.playCtx?.close();
    this.playCtx = null;
    this.ws?.close();
    this.ws = null;
    this.video.volume = this.savedVolume;
    log("stopped");
  }

  onTime(t: number): void {
    if (!this.gate || !this.plot) {
      return;
    }
    const segs = flatten(this.plot);
    const idx = segs.findIndex((s) => t >= s.start_sec && t < s.end_sec);
    if (idx < 0 || idx === this.lastSeg) {
      return;
    }
    if (this.lastSeg >= 0 && idx > this.lastSeg + 1) {
      const skipped = segs.slice(this.lastSeg + 1, idx);
      this.sendItem(
        `【快进】跳过：${skipped.map((s) => s.beat).join(" / ")}。现在进入：\n${formatSeg(segs[idx]!)}`,
        "剧情更新",
        `快进，进入「${segs[idx]!.beat}」`,
      );
    } else if (this.lastSeg >= 0 && idx < this.lastSeg) {
      this.sendItem(`【倒退】回到 ${formatSeg(segs[idx]!)}`, "剧情更新", `倒退，回到「${segs[idx]!.beat}」`);
    } else {
      this.sendItem(`【当前剧情】\n${formatSeg(segs[idx]!)}`, "剧情更新", `当前节拍「${segs[idx]!.beat}」`);
    }
    this.lastSeg = idx;
    this.onBeat?.(segs[idx]!);
    log("segment", idx, segs[idx]?.beat);
  }

  private async onServer(raw: string): Promise<void> {
    let msg: { type?: string; delta?: string; text?: string; stash?: string; transcript?: string };
    try {
      msg = JSON.parse(raw) as { type?: string; delta?: string; text?: string; stash?: string; transcript?: string };
    } catch {
      log("bad json from server");
      return;
    }
    const typ = msg.type ?? "";
    if (typ !== "response.audio.delta") {
      log("←", typ, raw.length);
    }
    switch (typ) {
      case "session.created":
        this.sendSessionUpdate(true);
        break;
      case "session.updated":
        this.gate = true;
        this.onStatus("Realtime 可说话");
        log("gate open, start mic+frames");
        await this.startMic();
        this.startFrames();
        this.onTime(this.video.currentTime);
        break;
      case "input_audio_buffer.speech_started":
        if (this.aiTalking) {
          this.stopPlayback();
          this.send({ type: "response.cancel" });
          log("barge-in cancel");
        }
        this.sendLookbackFrame();
        break;
      case "response.created":
        this.aiDoneEmitted = false;
        this.aiTalking = true;
        this.savedVolume = this.video.volume;
        this.video.volume = Math.min(this.savedVolume, 0.5);
        this.onTimeline?.({ type: "ai_start" });
        break;
      case "conversation.item.input_audio_transcription.delta":
        this.onTimeline?.({ type: "user_delta", text: msg.text ?? "", stash: msg.stash ?? "" });
        break;
      case "conversation.item.input_audio_transcription.completed":
        this.onTimeline?.({ type: "user_done", transcript: msg.transcript ?? "" });
        break;
      case "response.audio_transcript.delta":
        this.onTimeline?.({ type: "ai_delta", delta: msg.delta ?? "" });
        break;
      case "response.audio_transcript.done":
        this.aiDoneEmitted = true;
        this.onTimeline?.({ type: "ai_done", transcript: msg.transcript ?? "" });
        break;
      case "response.audio.delta":
        if (msg.delta) {
          this.enqueuePlay(msg.delta);
        }
        break;
      case "response.done":
        this.aiTalking = false;
        this.video.volume = this.savedVolume;
        if (!this.aiDoneEmitted) {
          this.onTimeline?.({ type: "ai_done", interrupted: true });
        }
        this.aiDoneEmitted = true;
        log("response done");
        break;
      case "error":
        log("server error", raw.slice(0, 400));
        this.onStatus("Realtime error");
        break;
      default:
        break;
    }
  }

  private sendSessionUpdate(notify = false): void {
    const plot = this.plot;
    if (!plot) {
      return;
    }
    const instructions = [
      "你是和用户一起看这部片子的同伴。用中文短句说话。",
      "以下面的剧情 JSON 为事实，不要编造没发生过的互动。",
      "优先回应当前播放段和画面。",
      JSON.stringify(plot),
    ].join("\n");
    const ok = this.send({
      type: "session.update",
      session: {
        modalities: ["text", "audio"],
        voice: this.voice,
        audio: {
          input: { format: { type: "pcm", sample_rate: 24000 } },
          output: { format: { type: "pcm", sample_rate: 24000 } },
        },
        input_audio_transcription: {
          model: "qwen3-asr-flash-realtime",
        },
        instructions,
        turn_detection: {
          type: "semantic_vad",
          threshold: 0.5,
          silence_duration_ms: 800,
        },
      },
    });
    if (ok && notify) {
      this.onTimeline?.({ type: "system", label: "剧情已灌注", detail: "开场整集剧情" });
    }
    log("→ session.update plot=", plot.title, "json_bytes=", instructions.length);
  }

  private sendItem(text: string, label: string, detail: string): void {
    const ok = this.send({
      type: "conversation.item.create",
      item: {
        type: "message",
        role: "user",
        content: [{ type: "input_text", text }],
      },
    });
    if (ok) {
      this.onTimeline?.({ type: "system", label, detail });
    }
    log("→ item.create", text.slice(0, 80));
  }

  private send(body: Record<string, unknown>): boolean {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      log("send skipped, ws not open", body.type);
      return false;
    }
    const payload = { event_id: newId(), ...body };
    this.ws.send(JSON.stringify(payload));
    return true;
  }

  private async startMic(): Promise<void> {
    this.mic = await navigator.mediaDevices.getUserMedia({
      audio: { echoCancellation: true, noiseSuppression: true },
    });
    const ctx = new AudioContext();
    this.captureCtx = ctx;
    if (ctx.state === "suspended") {
      await ctx.resume();
    }
    const src = ctx.createMediaStreamSource(this.mic);
    const proc = ctx.createScriptProcessor(4096, 1, 1);
    this.proc = proc;
    proc.onaudioprocess = (ev) => {
      if (!this.gate) {
        return;
      }
      const input = ev.inputBuffer.getChannelData(0);
      const resampled = resample(input, ctx.sampleRate, INPUT_RATE);
      this.pcmTail = concat(this.pcmTail, resampled);
      const chunk = Math.floor(INPUT_RATE / 10);
      while (this.pcmTail.length >= chunk) {
        const piece = this.pcmTail.subarray(0, chunk);
        const rest = this.pcmTail.subarray(chunk);
        const next = new Float32Array(rest.length);
        next.set(rest);
        this.pcmTail = next;
        this.send({ type: "input_audio_buffer.append", audio: floatToPcm16Base64(piece) });
        this.audioSent = true;
      }
    };
    const mute = ctx.createGain();
    mute.gain.value = 0;
    src.connect(proc);
    proc.connect(mute);
    mute.connect(ctx.destination);
    log("mic started context_rate=", ctx.sampleRate);
  }

  private startFrames(): void {
    this.frames = [];
    this.frameTimer = window.setInterval(() => {
      void this.grabFrame();
    }, FRAME_MS);
  }

  private async grabFrame(): Promise<void> {
    const packed = await encodeJpegFrame(this.video, this.ensureCanvas());
    if (!packed) {
      return;
    }
    this.frames.push({ at: performance.now(), b64: packed.b64, raw: packed.raw });
    if (this.frames.length > 6) {
      this.frames.shift();
    }
  }

  private ensureCanvas(): HTMLCanvasElement {
    if (!this.frameCanvas) {
      this.frameCanvas = document.createElement("canvas");
    }
    return this.frameCanvas;
  }

  private sendLookbackFrame(): void {
    if (!this.audioSent) {
      log("skip image: no audio sent yet");
      return;
    }
    const target = performance.now() - LOOKBACK_MS;
    let best: Frame | null = null;
    for (const f of this.frames) {
      if (f.at <= target) {
        best = f;
      }
    }
    if (!best) {
      best = this.frames[this.frames.length - 1] ?? null;
    }
    if (!best) {
      log("skip image: ring empty");
      return;
    }
    const ok = this.send({ type: "input_image_buffer.append", image: best.b64 });
    if (ok) {
      this.onTimeline?.({ type: "system", label: "截图已发送", detail: `JPEG ${best.raw}B` });
    }
    log(
      "→ image_buffer jpeg=",
      best.raw,
      "b64=",
      best.b64.length,
      "lookback_ms=",
      Math.round(performance.now() - best.at),
    );
  }

  private enqueuePlay(b64: string): void {
    const pcm = pcm16Base64ToFloat(b64);
    if (pcm.length === 0) {
      return;
    }
    const ctx = this.ensurePlayCtx();
    const buf = ctx.createBuffer(1, pcm.length, INPUT_RATE);
    buf.getChannelData(0).set(pcm);
    const src = ctx.createBufferSource();
    src.buffer = buf;
    src.connect(ctx.destination);
    const now = ctx.currentTime;
    if (this.playAt < now) {
      this.playAt = now;
    }
    src.start(this.playAt);
    this.playAt += buf.duration;
    this.sources.push(src);
    src.onended = () => {
      this.sources = this.sources.filter((s) => s !== src);
    };
  }

  private ensurePlayCtx(): AudioContext {
    if (!this.playCtx) {
      this.playCtx = new AudioContext({ sampleRate: INPUT_RATE });
      this.playAt = 0;
    }
    return this.playCtx;
  }

  private stopPlayback(): void {
    for (const s of this.sources) {
      try {
        s.stop();
      } catch {
        /* already stopped */
      }
    }
    this.sources = [];
    if (this.playCtx) {
      this.playAt = this.playCtx.currentTime;
    }
    this.video.volume = this.savedVolume;
    this.aiTalking = false;
  }
}

function flatten(plot: PlotDoc): SubSegment[] {
  const out: SubSegment[] = [];
  for (const m of plot.major_segments ?? []) {
    for (const s of m.sub_segments ?? []) {
      out.push(s);
    }
  }
  return out;
}

function formatSeg(s: SubSegment): string {
  return [
    `${s.start_sec}-${s.end_sec}s ${s.beat}`,
    s.summary,
    `台词：${s.key_dialogue}`,
    `画面：${s.visual_scene}`,
    `情绪：${s.emotion}`,
    `前情：${s.story_so_far}`,
  ].join("\n");
}

function newId(): string {
  return "evt_" + crypto.randomUUID().replace(/-/g, "");
}

function resample(input: Float32Array, fromRate: number, toRate: number): Float32Array {
  if (fromRate === toRate) {
    return new Float32Array(input);
  }
  const ratio = fromRate / toRate;
  const outLen = Math.max(1, Math.floor(input.length / ratio));
  const out = new Float32Array(outLen);
  for (let i = 0; i < outLen; i++) {
    const x = i * ratio;
    const i0 = Math.floor(x);
    const f = x - i0;
    const a = input[i0] ?? 0;
    const b = input[i0 + 1] ?? a;
    out[i] = a + (b - a) * f;
  }
  return out;
}

function concat(a: Float32Array, b: Float32Array): Float32Array {
  const out = new Float32Array(a.length + b.length);
  out.set(a, 0);
  out.set(b, a.length);
  return out;
}

function floatToPcm16Base64(input: Float32Array): string {
  const bytes = new Uint8Array(input.length * 2);
  const view = new DataView(bytes.buffer);
  for (let i = 0; i < input.length; i++) {
    const s = Math.max(-1, Math.min(1, input[i]!));
    view.setInt16(i * 2, s < 0 ? s * 0x8000 : s * 0x7fff, true);
  }
  return u8ToB64(bytes);
}

function pcm16Base64ToFloat(b64: string): Float32Array {
  const bin = atob(b64);
  const out = new Float32Array(bin.length / 2);
  const view = new DataView(new ArrayBuffer(bin.length));
  for (let i = 0; i < bin.length; i++) {
    view.setUint8(i, bin.charCodeAt(i));
  }
  for (let i = 0; i < out.length; i++) {
    out[i] = view.getInt16(i * 2, true) / 0x8000;
  }
  return out;
}

function u8ToB64(bytes: Uint8Array): string {
  let bin = "";
  const step = 0x8000;
  for (let i = 0; i < bytes.length; i += step) {
    bin += String.fromCharCode(...bytes.subarray(i, i + step));
  }
  return btoa(bin);
}

function fitFrame(srcW: number, srcH: number, maxW: number, maxH: number): { w: number; h: number } {
  const r = Math.min(maxW / srcW, maxH / srcH, 1);
  return {
    w: Math.max(1, Math.round(srcW * r)),
    h: Math.max(1, Math.round(srcH * r)),
  };
}

async function encodeJpegFrame(
  video: HTMLVideoElement,
  canvas: HTMLCanvasElement,
): Promise<{ b64: string; raw: number } | null> {
  if (video.readyState < 2 || video.videoWidth === 0) {
    return null;
  }
  let maxW = FRAME_MAX_W;
  let maxH = FRAME_MAX_H;
  for (let scale = 0; scale < 4; scale++) {
    const { w, h } = fitFrame(video.videoWidth, video.videoHeight, maxW, maxH);
    canvas.width = w;
    canvas.height = h;
    const ctx = canvas.getContext("2d");
    if (!ctx) {
      return null;
    }
    ctx.drawImage(video, 0, 0, w, h);
    for (let q = 0.7; q >= 0.35; q -= 0.1) {
      const blob = await canvasToJpeg(canvas, q);
      if (!blob || blob.size > JPEG_MAX_RAW) {
        continue;
      }
      const b64 = await blobToB64(blob);
      if (b64.length > JPEG_MAX_B64) {
        continue;
      }
      return { b64, raw: blob.size };
    }
    maxW = Math.floor(maxW * 0.75);
    maxH = Math.floor(maxH * 0.75);
  }
  log("frame drop: cannot fit jpeg<=192KB and b64<=256KB");
  return null;
}

function canvasToJpeg(canvas: HTMLCanvasElement, quality: number): Promise<Blob | null> {
  return new Promise((resolve) => {
    canvas.toBlob((b) => resolve(b), "image/jpeg", quality);
  });
}

function blobToB64(blob: Blob): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => {
      const s = String(reader.result ?? "");
      const comma = s.indexOf(",");
      resolve(comma >= 0 ? s.slice(comma + 1) : s);
    };
    reader.onerror = () => reject(reader.error);
    reader.readAsDataURL(blob);
  });
}

function log(...args: unknown[]): void {
  console.log("[withyou]", ...args);
}
