# voice-demo

StepFun（阶跃星辰）Realtime API 的实时双向语音验证版。后端 Go + `gorilla/websocket`，前端单一 Vanilla TS 页面。

## 快速开始

```bash
# 1. 准备 Key（复制样例并按需修改）
copy .env.example .env

# 2. 构建前端（把 TS 编译成 web/assets/*.js）
node web/build.mjs

# 3. 拉取 Go 依赖并启动
go mod tidy
go run ./cmd/server

# 浏览器打开
http://localhost:8080
```

打开页面后点"开始对话"，授权麦克风即可语音交互；也支持文本输入调试。

## 常用命令

```bash
go run ./cmd/server        # 启动
go build ./...             # 编译检查
go vet ./...               # 静态检查
go test ./...              # 跑测试
node web/build.mjs         # 重新构建前端
npx tsc --noEmit -p web    # 前端类型检查
```

## 配置

见 `.env`（已 gitignore，不提交）。字段说明见 `计划书.md` 第 8 节。

## 目录

见 `计划书.md` 第 5 节。
