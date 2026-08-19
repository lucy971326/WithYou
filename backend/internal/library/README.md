# library

打开一部本机视频，让浏览器播。**不上传、不解码**：选框拿路径，`/media` 按 Range 读盘。

```mermaid
flowchart LR
  U[点打开] --> P[Picker 系统选框]
  P --> F[原片路径]
  F --> Y[Playable 探测编码]
  Y --> S[(当前文件 state)]
  S --> M[Media Range]
  M --> V[浏览器 video]
```

| 能力 | 干什么 |
|------|--------|
| `Picker` | Windows 原生选框，只要路径 |
| `Playable` | ffprobe 标一下 Chrome 能不能硬解，**不转码** |
| `Media` | `GET /media` 读原片；路径不回给前端 |
| `HTTP` | `/api/open`、`/api/current`、`/media` |

`plot` 只问 `Media.Current()` 要原片路径去抽字幕。
