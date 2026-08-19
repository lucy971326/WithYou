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
  } catch (err) {
    statusEl.textContent = err instanceof Error ? err.message : "打开失败";
  } finally {
    openBtn.disabled = false;
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
