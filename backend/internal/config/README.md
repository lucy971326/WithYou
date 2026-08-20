# config

进程启动时读一次环境 / `.env`，得到一份没有行为的 `Config`。不持有连接，不提供 HTTP。

```mermaid
flowchart LR
  E[.env / 环境变量] --> L[Load]
  L --> C[Config]
  C --> S[cmd/server]
```

| 给谁 | 什么 |
|------|------|
| library / HTTP | `Addr` |
| plot | Qwen API key、模型名、当前站点的 Responses API 地址 |
| realtime | Qwen key、当前站点的 Realtime URL / 模型、默认音色、音色列表接口 |

通过 `QWEN_SITE=intl` 或 `QWEN_SITE=cn` 选择国际站 / 国内站；两站的 API Key 不通用。
Key 只活在本机进程里。缺 Realtime key 时服务仍能起来，升级 WS 再 503。
