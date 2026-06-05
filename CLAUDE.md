# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

PrismProxy 是类似 Reqable 的跨平台 HTTP 调试代理工具，集成 AI 能力。采用 Go 后端 + React 前端 + Tauri 桌面端的架构。

## 技术栈

- **后端**: Go + SQLite (modernc.org/sqlite, 纯 Go 无 CGO) + gRPC
- **前端**: React 19 + TypeScript 5 + TailwindCSS 4 + Zustand (状态管理)
- **桌面端**: Tauri 2.x (Rust)
- **通信**: gRPC-Web (前端 ↔ 后端)，Protocol Buffers 定义接口

## 常用命令

```bash
# 构建后端
go build -o prismproxy-server ./cmd/server

# 运行后端服务器 (gRPC :9090 + HTTP/gRPC-Web :8080)
go run ./cmd/server --port 9090

# 前端开发 (启动 Tauri 桌面窗口 + 前端 + Go sidecar)
cd desktop && npm run dev

# 仅前端开发 (不启动 Tauri)
cd desktop && npm run frontend:dev

# 构建桌面应用
cd desktop && npx tauri build

# 构建 Go sidecar (用于 Tauri 打包)
bash scripts/build_sidecar.sh

# 生成 Proto 代码 (Go + TypeScript)
bash scripts/gen_proto.sh go

# 代码检查
go vet ./...
go fmt ./...
```

## 架构要点

### 通信架构

前端通过 **gRPC-Web** 协议与后端通信，不是 REST API。`cmd/server/main.go` 同时启动：
- gRPC 服务器 (:9090) - 原生 gRPC
- HTTP 服务器 (:8080) - 包装 gRPC-Web，处理浏览器请求

前端服务层 (`desktop/src/services/`) 每个模块对应一个 gRPC 服务调用。

### 模块化设计

`internal/` 下每个子目录是一个独立模块，遵循统一模式：
- 模块核心逻辑 (如 `rules/engine.go`)
- gRPC 服务实现 (`internal/grpc/*_service.go`) 桥接模块与 gRPC 接口
- 部分模块有独立的 Store 层操作 SQLite

`internal/grpc/server.go` 中的 `Server` 结构体聚合所有模块，`registerServices()` 注册所有 gRPC 服务。

### Tauri Sidecar 模式

Go 后端编译为平台特定的二进制文件，作为 Tauri 的 **sidecar** 运行：
- 构建脚本: `scripts/build_sidecar.sh`
- 输出目录: `desktop/src-tauri/bin/`
- 命名规范: `prismproxy-server-{target-triple}` (如 `aarch64-apple-darwin`)
- Tauri 配置: `externalBin` 字段引用 sidecar

### 数据库

使用 SQLite，数据库文件默认 `./prismproxy.db`，通过环境变量 `DB_PATH` 可配置。
迁移系统在 `internal/storage/migration.go`，服务器启动时自动执行。

### Proto 定义

Proto 文件在 `proto/` 目录，生成的 Go 代码在 `proto/gen/go/`。
修改 proto 后需运行 `bash scripts/gen_proto.sh go` 重新生成。

## 开发规范

- 使用中文注释，简洁明了，专业名词用英文
- 不要过度设计，保持代码简洁易读
- 提交前格式化代码 (`go fmt ./...`)

## 端口说明

| 服务 | 端口 | 说明 |
|------|------|------|
| HTTP 代理 | 8888 | 抓包代理端口 |
| gRPC | 9090 | 原生 gRPC 服务 |
| HTTP/gRPC-Web | 8080 | 浏览器访问的 HTTP 端点 |
| 前端开发 | 3000 | Vite 开发服务器 |
