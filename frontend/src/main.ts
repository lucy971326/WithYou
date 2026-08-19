import "./style.css";

const openBtn = document.querySelector<HTMLButtonElement>("#open-btn")!;
const statusEl = document.querySelector<HTMLSpanElement>("#status")!;
const player = document.querySelector<HTMLVideoElement>("#player")!;
const empty = document.querySelector<HTMLParagraphElement>("#empty")!;

type OpenResponse = {
  name: string;
  size: number;
  codec: string;
  pixelFormat: string;
  browserSafe: boolean;
};
type ErrorResponse = { error: string };
type SubtitleDoc = { format: string; count: number };
type EnrichResponse = {
  title: string;
  cached: boolean;
  major_count: number;
  sub_count: number;
  grand_summary: string;
};

openBtn.addEventListener("click", () => {
  void openVideo();
});

player.addEventListener("loadeddata", () => {
  if (player.videoWidth === 0 && player.duration > 0) {
    statusEl.textContent = `${statusEl.textContent} · 只有声音：浏览器解不了这条视频轨`;
  }
});

async function openVideo(): Promise<void> {
  openBtn.disabled = true;
  statusEl.textContent = "在系统窗口里选文件…";
  try {
    const resp = await fetch("/api/open", { method: "POST" });
    const body = (await resp.json()) as OpenResponse & ErrorResponse;
    if (resp.status === 409) {
      statusEl.textContent = "已取消";
      return;
    }
    if (!resp.ok) {
      statusEl.textContent = body.error || "打开失败";
      return;
    }
    statusEl.textContent = formatStatus(body);
    empty.classList.add("is-hidden");
    player.classList.add("is-on");
    player.src = `/media?t=${Date.now()}`;
    await player.play().catch(() => {
      /* 浏览器可能拦自动播放，用户点控件即可 */
    });
    await loadSubtitles();
  } catch (err) {
    statusEl.textContent = err instanceof Error ? err.message : "打开失败";
  } finally {
    openBtn.disabled = false;
  }
}

async function loadSubtitles(): Promise<void> {
  const prev = statusEl.textContent ?? "";
  statusEl.textContent = `${prev} · 抽字幕…`;
  try {
    const resp = await fetch("/api/plot/subtitles", { method: "POST" });
    const body = (await resp.json()) as SubtitleDoc & ErrorResponse;
    if (!resp.ok) {
      statusEl.textContent = `${prev} · ${body.error || "抽字幕失败"}`;
      return;
    }
    statusEl.textContent = `${prev} · 字幕 ${body.count} 条`;
    await loadPlot();
  } catch (err) {
    statusEl.textContent = `${prev} · ${err instanceof Error ? err.message : "抽字幕失败"}`;
  }
}

async function loadPlot(): Promise<void> {
  const prev = statusEl.textContent ?? "";
  statusEl.textContent = `${prev} · 富化剧情…`;
  try {
    const resp = await fetch("/api/plot/enrich", { method: "POST" });
    const body = (await resp.json()) as EnrichResponse & ErrorResponse;
    if (!resp.ok) {
      statusEl.textContent = `${prev} · ${body.error || "富化失败"}`;
      return;
    }
    const via = body.cached ? "缓存" : "DeepSeek";
    statusEl.textContent = `${prev} · 剧情 ${body.major_count} 段 / ${body.sub_count} 节（${via}）`;
  } catch (err) {
    statusEl.textContent = `${prev} · ${err instanceof Error ? err.message : "富化失败"}`;
  }
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
