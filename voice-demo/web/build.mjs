import { build } from "esbuild";

// 页面主逻辑：打包成一个 IIFE，浏览器 script 直接加载。
await build({
  entryPoints: ["src/app.ts"],
  bundle: true,
  outfile: "assets/app.js",
  format: "iife",
  target: "es2022",
});

// AudioWorklet：必须在独立文件里注册，且不能依赖外部模块。
await build({
  entryPoints: ["src/worklet.ts"],
  bundle: true,
  outfile: "assets/worklet.js",
  format: "iife",
  target: "es2022",
});

console.log("前端构建完成: assets/app.js, assets/worklet.js");
