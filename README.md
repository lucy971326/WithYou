# WithYou

> 看番陪伴。
> 能看你的屏幕 能真的理解剧情 能陪你吐槽
> 音色多种可选 推荐“知芝” 很好听

## 先创建.env 

复制配置项.env.example 到 .env
最好 设置 QWEN_SITE=cn 国内站点 然后从 https://www.qianwenai.com/ 平台获取APIkey 填入，（无需购买有免费额度够玩一阵子了）

cn国内站点用完再去国际站 把QWEN_SITE=intl 替换上去， 去https://www.qwencloud.com/ 平台拿APIkey


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


## ！！！注意

目前不支持 普通视频，仅支持内挂字幕视频(可以百度了解下)，因为LLM需要去通过 字幕来组建剧情线从而到阶段后注入给Realtime模型