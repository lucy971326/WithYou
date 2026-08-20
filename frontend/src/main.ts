import "./style.css";
import { RealtimeWatch, type PlotDoc, type SubSegment, type TimelineEvent } from "./realtime";

const openBtn = document.querySelector<HTMLButtonElement>("#open-btn")!;
const emptyOpenBtn = document.querySelector<HTMLButtonElement>("#empty-open-btn")!;
const statusText = document.querySelector<HTMLSpanElement>("#status-text")!;
const assistantState = document.querySelector<HTMLSpanElement>("#assistant-state")!;
const assistantStatus = document.querySelector<HTMLParagraphElement>("#assistant-status")!;
const player = document.querySelector<HTMLVideoElement>("#player")!;
const stage = document.querySelector<HTMLDivElement>("#stage")!;
const subEl = document.querySelector<HTMLDivElement>("#subtitles")!;
const empty = document.querySelector<HTMLDivElement>("#empty")!;
const mediaMeta = document.querySelector<HTMLDivElement>("#media-meta")!;
const voiceSelect = document.querySelector<HTMLSelectElement>("#voice-select")!;
const voiceName = document.querySelector<HTMLSpanElement>("#voice-name")!;
const voiceDesc = document.querySelector<HTMLParagraphElement>("#voice-desc")!;
const plotTitleEl = document.querySelector<HTMLDivElement>("#plot-title")!;
const currentBeatEl = document.querySelector<HTMLDivElement>("#current-beat")!;
const beatDetailEl = document.querySelector<HTMLDivElement>("#beat-detail")!;
const transcriptList = document.querySelector<HTMLDivElement>("#transcript-list")!;
const clearTranscriptBtn = document.querySelector<HTMLButtonElement>("#clear-transcript")!;

type Cue = { start_sec: number; end_sec: number; text: string };
type VoiceOption = {
  voice: string;
  name?: string;
  description?: string;
  custom?: boolean;
  target_model?: string;
};
type VoicesResponse = {
  default_voice: string;
  preset: VoiceOption[];
  custom: VoiceOption[];
  custom_error?: string;
};
type OpenResponse = {
  name: string;
  size: number;
  codec: string;
  pixelFormat: string;
  browserSafe: boolean;
};
type ErrorResponse = { error: string };
type SubtitleDoc = { format: string; count: number; cues: Cue[] };
type EnrichResponse = {
  title: string;
  cached: boolean;
  major_count: number;
  sub_count: number;
  grand_summary: string;
};

type TranscriptEntry = {
  id: string;
  kind: "user" | "ai" | "event";
  text: string;
  detail?: string;
  status?: "done" | "interrupted";
  at: number;
};
type TranscriptResponse = {
  name: string;
  entries: TranscriptEntry[];
};

let cues: Cue[] = [];
let watch: RealtimeWatch | null = null;
let currentVoice = "Tina";
let voices: VoiceOption[] = [];
let transcript: TranscriptEntry[] = [];
let userDraftId: string | null = null;
let aiDraftId: string | null = null;
let persistQueue: Promise<void> = Promise.resolve();

openBtn.addEventListener("click", () => {
  void openVideo();
});

emptyOpenBtn.addEventListener("click", () => {
  void openVideo();
});

player.addEventListener("loadeddata", () => {
  if (player.videoWidth === 0 && player.duration > 0) {
    mediaMeta.textContent = `${mediaMeta.textContent} · 只有声音：浏览器解不了这条视频轨`;
  }
});

player.addEventListener("timeupdate", () => {
  paintCue(player.currentTime);
  watch?.onTime(player.currentTime);
});

player.addEventListener("seeked", () => {
  paintCue(player.currentTime);
  watch?.onTime(player.currentTime);
});

voiceSelect.addEventListener("change", () => {
  currentVoice = voiceSelect.value;
  watch?.setVoice(currentVoice);
  updateVoiceMeta();
});

clearTranscriptBtn.addEventListener("click", () => {
  void clearTranscript();
});

void loadVoices();
renderTranscript();

async function loadVoices(): Promise<void> {
  try {
    const resp = await fetch("/api/voices");
    if (!resp.ok) {
      return;
    }
    const body = (await resp.json()) as VoicesResponse;
    voiceSelect.innerHTML = "";
    voices = [];

    const presetGroup = document.createElement("optgroup");
    presetGroup.label = "官方音色";
    for (const v of body.preset ?? []) {
      presetGroup.append(voiceOption(v));
      voices.push(v);
    }
    if (presetGroup.childElementCount > 0) {
      voiceSelect.append(presetGroup);
    }

    const customGroup = document.createElement("optgroup");
    customGroup.label = "我的音色";
    for (const v of body.custom ?? []) {
      customGroup.append(voiceOption(v));
      voices.push(v);
    }
    if (customGroup.childElementCount > 0) {
      voiceSelect.append(customGroup);
    }

    currentVoice = body.default_voice || "Tina";
    voiceSelect.value = currentVoice;
    if (!voiceSelect.value) {
      const first = voiceSelect.querySelector("option") as HTMLOptionElement | null;
      currentVoice = first?.value ?? "Tina";
      voiceSelect.value = currentVoice;
    }
    updateVoiceMeta();
  } catch {
    // 拉不到音色时保持默认 Tina，不阻塞开片。
  }
}

function voiceOption(v: VoiceOption): HTMLOptionElement {
  const opt = document.createElement("option");
  opt.value = v.voice;
  opt.textContent = v.name ? `${v.name}（${v.voice}）` : v.voice;
  opt.dataset.desc = v.description ?? "";
  return opt;
}

function updateVoiceMeta(): void {
  const found = voices.find((v) => v.voice === currentVoice);
  voiceName.textContent = found?.name || currentVoice;
  if (!found) {
    voiceDesc.textContent = "默认音色 Tina";
    return;
  }
  const parts = [found.description || "自定义音色"];
  if (found.custom) {
    parts.push("复刻音色");
  }
  if (found.target_model) {
    parts.push(found.target_model);
  }
  voiceDesc.textContent = parts.join(" · ");
}

async function openVideo(): Promise<void> {
  openBtn.disabled = true;
  emptyOpenBtn.disabled = true;
  statusText.textContent = "在系统窗口里选文件…";
  try {
    const resp = await fetch("/api/open", { method: "POST" });
    const body = (await resp.json()) as OpenResponse & ErrorResponse;
    if (resp.status === 409) {
      statusText.textContent = "已取消";
      return;
    }
    if (!resp.ok) {
      statusText.textContent = body.error || "打开失败";
      return;
    }
    mediaMeta.textContent = formatStatus(body);
    statusText.textContent = "已打开，准备字幕";
    empty.classList.add("is-hidden");
    stage.classList.remove("is-hidden");
    cues = [];
    subEl.textContent = "";
    watch?.stop();
    watch = null;
    assistantState.textContent = "待机";
    assistantState.dataset.state = "idle";
    assistantStatus.textContent = "正在准备字幕与剧情…";
    plotTitleEl.textContent = "还没富化剧情";
    currentBeatEl.textContent = "随播放进度更新";
    beatDetailEl.textContent = "";
    userDraftId = null;
    aiDraftId = null;
    transcript = [];
    await loadTranscript();
    player.src = `/media?t=${Date.now()}`;
    await player.play().catch(() => {
      /* 浏览器可能拦自动播放，用户点控件即可 */
    });
    await loadSubtitles();
  } catch (err) {
    statusText.textContent = err instanceof Error ? err.message : "打开失败";
  } finally {
    openBtn.disabled = false;
    emptyOpenBtn.disabled = false;
  }
}

async function loadSubtitles(): Promise<void> {
  statusText.textContent = "抽字幕中…";
  try {
    const resp = await fetch("/api/plot/subtitles", { method: "POST" });
    const body = (await resp.json()) as SubtitleDoc & ErrorResponse;
    if (!resp.ok) {
      statusText.textContent = body.error || "抽字幕失败";
      return;
    }
    cues = body.cues ?? [];
    paintCue(player.currentTime);
    statusText.textContent = `字幕 ${body.count} 条，正在富化剧情`;
    await loadPlot();
  } catch (err) {
    statusText.textContent = err instanceof Error ? err.message : "抽字幕失败";
  }
}

async function loadPlot(): Promise<void> {
  statusText.textContent = "富化剧情中…";
  try {
    const resp = await fetch("/api/plot/enrich", { method: "POST" });
    const body = (await resp.json()) as EnrichResponse & ErrorResponse;
    if (!resp.ok) {
      statusText.textContent = body.error || "富化失败";
      return;
    }
    plotTitleEl.textContent = body.title || "未命名";
    const via = body.cached ? "缓存" : "Qwen";
    statusText.textContent = `剧情 ${body.major_count} 段 / ${body.sub_count} 节（${via}），连接陪看`;
    await startRealtime();
  } catch (err) {
    statusText.textContent = err instanceof Error ? err.message : "富化失败";
  }
}

async function startRealtime(): Promise<void> {
  statusText.textContent = "连接 Realtime…";
  try {
    const resp = await fetch("/api/plot/enrich");
    const plot = (await resp.json()) as PlotDoc & ErrorResponse;
    if (!resp.ok) {
      statusText.textContent = plot.error || "无剧情";
      return;
    }
    watch = new RealtimeWatch(player, onRealtimeStatus, paintBeat, onTimeline);
    watch.setVoice(currentVoice);
    await watch.start(plot);
  } catch (err) {
    statusText.textContent = err instanceof Error ? err.message : "Realtime 失败";
  }
}

function onRealtimeStatus(s: string): void {
  statusText.textContent = s;
  assistantStatus.textContent = s;
  if (s.includes("可说话")) {
    assistantState.textContent = "陪看中";
    assistantState.dataset.state = "live";
  } else if (s.includes("已连上")) {
    assistantState.textContent = "连接中";
    assistantState.dataset.state = "pending";
  } else if (s.includes("断开")) {
    assistantState.textContent = "离线";
    assistantState.dataset.state = "off";
  } else {
    assistantState.textContent = "待机";
    assistantState.dataset.state = "idle";
  }
}

function paintBeat(seg: SubSegment): void {
  currentBeatEl.textContent = seg.beat;
  beatDetailEl.textContent = [
    `剧情：${seg.summary}`,
    `台词：${seg.key_dialogue}`,
    `画面：${seg.visual_scene}`,
    `情绪：${seg.emotion}`,
  ].join("\n");
}

function paintCue(t: number): void {
  for (const cue of cues) {
    if (t >= cue.start_sec && t < cue.end_sec) {
      subEl.textContent = cue.text;
      return;
    }
  }
  subEl.textContent = "";
}

async function loadTranscript(): Promise<void> {
  try {
    const resp = await fetch("/api/transcript");
    if (!resp.ok) {
      transcript = [];
      renderTranscript();
      return;
    }
    const body = (await resp.json()) as TranscriptResponse;
    transcript = body.entries ?? [];
    if (transcript.length > 0) {
      transcript.push({
        id: newTranscriptId(),
        kind: "event",
        text: "已恢复历史记录，模型上下文已重置",
        at: Date.now(),
      });
    }
    renderTranscript();
  } catch {
    transcript = [];
    renderTranscript();
  }
}

async function clearTranscript(): Promise<void> {
  userDraftId = null;
  aiDraftId = null;
  transcript = [];
  renderTranscript();
  try {
    await fetch("/api/transcript", { method: "DELETE" });
  } catch {
    // 后端清不掉时也先清掉页面。
  }
}

function onTimeline(ev: TimelineEvent): void {
  switch (ev.type) {
    case "user_delta": {
      if (!userDraftId) {
        userDraftId = addDraft("user");
      }
      const entry = findEntry(userDraftId);
      if (entry) {
        entry.text = ev.text + ev.stash;
      }
      renderTranscript();
      break;
    }
    case "user_done": {
      const entry = userDraftId ? findEntry(userDraftId) : null;
      if (entry) {
        entry.text = ev.transcript;
        entry.status = "done";
        void persistEntry(entry);
      } else if (ev.transcript) {
        const created = finalEntry("user", ev.transcript, "done");
        void persistEntry(created);
      }
      userDraftId = null;
      renderTranscript();
      break;
    }
    case "ai_start": {
      aiDraftId = addDraft("ai");
      renderTranscript();
      break;
    }
    case "ai_delta": {
      const entry = aiDraftId ? findEntry(aiDraftId) : null;
      if (entry) {
        entry.text += ev.delta;
        renderTranscript();
      }
      break;
    }
    case "ai_done": {
      const entry = aiDraftId ? findEntry(aiDraftId) : null;
      if (entry) {
        entry.text = ev.transcript ?? entry.text;
        entry.status = ev.interrupted ? "interrupted" : "done";
        void persistEntry(entry);
      } else if (ev.transcript) {
        const created = finalEntry("ai", ev.transcript, ev.interrupted ? "interrupted" : "done");
        void persistEntry(created);
      }
      aiDraftId = null;
      renderTranscript();
      break;
    }
    case "system": {
      const created = finalEntry("event", ev.label, "done", ev.detail);
      void persistEntry(created);
      renderTranscript();
      break;
    }
  }
}

function addDraft(kind: "user" | "ai"): string {
  const entry: TranscriptEntry = { id: newTranscriptId(), kind, text: "", at: Date.now() };
  transcript.push(entry);
  return entry.id;
}

function findEntry(id: string): TranscriptEntry | undefined {
  return transcript.find((e) => e.id === id);
}

function finalEntry(
  kind: "user" | "ai" | "event",
  text: string,
  status: "done" | "interrupted",
  detail?: string,
): TranscriptEntry {
  const entry: TranscriptEntry = { id: newTranscriptId(), kind, text, status, at: Date.now() };
  if (detail) {
    entry.detail = detail;
  }
  transcript.push(entry);
  return entry;
}

function newTranscriptId(): string {
  return "tr_" + crypto.randomUUID().replace(/-/g, "");
}

async function persistEntry(entry: TranscriptEntry): Promise<void> {
  persistQueue = persistQueue.then(async () => {
    try {
      await fetch("/api/transcript", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ entry }),
      });
    } catch {
      // 写盘失败不阻塞 UI，记录仍在内存里。
    }
  });
  await persistQueue;
}

function renderTranscript(): void {
  transcriptList.textContent = "";
  if (transcript.length === 0) {
    const empty = document.createElement("p");
    empty.className = "transcript-empty";
    empty.textContent = "还没有对话记录";
    transcriptList.append(empty);
    return;
  }
  for (const entry of transcript) {
    transcriptList.append(transcriptRow(entry));
  }
  transcriptList.scrollTop = transcriptList.scrollHeight;
}

function transcriptRow(entry: TranscriptEntry): HTMLElement {
  if (entry.kind === "event") {
    const row = document.createElement("div");
    row.className = "tl-event";
    const label = document.createElement("span");
    label.textContent = entry.text;
    const detail = document.createElement("span");
    detail.className = "tl-event-detail";
    detail.textContent = entry.detail ? ` · ${entry.detail}` : "";
    row.append(label, detail);
    return row;
  }

  const row = document.createElement("div");
  row.className = entry.kind === "user" ? "tl-row tl-user" : "tl-row tl-ai";
  const who = document.createElement("span");
  who.className = "tl-who";
  who.textContent = entry.kind === "user" ? "你" : "AI";
  const bubble = document.createElement("div");
  bubble.className = "tl-bubble";
  let text = entry.text || (entry.status ? "" : "…");
  if (entry.status === "interrupted") {
    text = `${text}（被打断）`;
  }
  bubble.textContent = text;
  row.append(who, bubble);
  return row;
}

function formatStatus(info: OpenResponse): string {
  const codec = [info.codec, info.pixelFormat].filter(Boolean).join(" ");
  if (!info.codec) {
    return info.name;
  }
  if (info.browserSafe) {
    return codec ? `${info.name} · ${codec}` : info.name;
  }
  return `${info.name} · ${codec}（Chrome 常解不了，会只有声音）`;
}
