import { defineConfig } from "vite";

export default defineConfig({
  server: {
    host: "127.0.0.1",
    port: 5173,
    proxy: {
      "/api": { target: "http://127.0.0.1:8080", ws: true },
      "/media": "http://127.0.0.1:8080",
    },
  },
});
