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
| plot | Qwen API key、模型名 |
| realtime | Qwen key、Realtime URL / 模型 |

Key 只活在本机进程里。缺 Realtime key 时服务仍能起来，升级 WS 再 503。
