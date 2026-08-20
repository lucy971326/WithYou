# WithYou

本机看番陪伴。整体规划与现状见 `docs/交接文档.md`。

## 跑第 1 块（打开并能播）

两个终端：

```text
cd backend
go run ./cmd/server
```

```text
cd frontend
npm install
npm run dev
```

浏览器开 http://127.0.0.1:5173 ，点「打开视频」，在系统窗口里选 mkv/mp4。

Chrome 对 mkv 支持不稳，播不动就先换 mp4。
