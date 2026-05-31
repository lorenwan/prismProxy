import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    port: 3000,
    proxy: {
      // gRPC-Web 代理到 Go 后端
      "/prismproxy.TrafficService": {
        target: "http://localhost:9090",
        changeOrigin: true,
      },
      "/prismproxy.RulesService": {
        target: "http://localhost:9090",
        changeOrigin: true,
      },
      "/prismproxy.BreakpointsService": {
        target: "http://localhost:9090",
        changeOrigin: true,
      },
      "/prismproxy.RewritesService": {
        target: "http://localhost:9090",
        changeOrigin: true,
      },
      "/prismproxy.CollectionsService": {
        target: "http://localhost:9090",
        changeOrigin: true,
      },
      "/prismproxy.EnvironmentsService": {
        target: "http://localhost:9090",
        changeOrigin: true,
      },
      "/prismproxy.AIService": {
        target: "http://localhost:9090",
        changeOrigin: true,
      },
      "/prismproxy.SystemService": {
        target: "http://localhost:9090",
        changeOrigin: true,
      },
      "/prismproxy.CodeGenService": {
        target: "http://localhost:9090",
        changeOrigin: true,
      },
      // REST 兼容 (旧接口)
      "/api": {
        target: "http://localhost:8080",
        changeOrigin: true,
      },
      "/ws": {
        target: "ws://localhost:8080",
        ws: true,
      },
    },
  },
  build: {
    outDir: "dist",
  },
});
