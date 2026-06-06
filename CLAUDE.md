# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

PrismProxy 是类似 Reqable 的跨平台 HTTP 调试代理工具，集成 AI 能力。采用 **前端 → Rust → Go** 三层架构。

```
┌─────────────────┐     Tauri IPC      ┌─────────────────┐     gRPC      ┌─────────────────┐
│  React Frontend │ ──────────────────→ │  Rust (Tauri)   │ ────────────→ │   Go Backend    │
│                 │ ←────────────────── │                 │ ←──────────── │                 │
└─────────────────┘     invoke()        └─────────────────┘    tonic      └─────────────────┘
```

## 技术栈

### 后端
- Go 1.21+
- SQLite (modernc.org/sqlite, 纯 Go 无 CGO)
- gRPC + Protocol Buffers

### 前端
- React 19
- TypeScript 5
- TailwindCSS 4
- shadcn/ui (组件库)
- Zustand (状态管理)
- Lucide React (图标)
- Vite 6 (构建工具)

### 桌面端 (Rust)
- Tauri 2.x - 原生窗口、系统集成
- tonic - gRPC 客户端
- Sidecar 管理 - Go 进程生命周期

### 通信协议
- Tauri IPC (前端 → Rust) - `invoke('command', args)`
- gRPC (Rust → Go) - tonic 客户端
- Tauri Event (Rust → 前端) - 流式数据推送

## 常用命令

详见 [docs/development/README.md](docs/development/README.md#常用命令)

## 架构要点

### 通信架构

**前端 → Rust → Go** 三层通信：

1. **前端 → Rust**：通过 Tauri IPC (`invoke`)
   - 前端调用 `invoke('command_name', args)` 发起请求
   - Rust 层接收 IPC 调用，转换为 gRPC 请求

2. **Rust → Go**：通过 gRPC (tonic)
   - Rust 使用 tonic 客户端连接 Go 后端
   - 管理连接池、重连、超时

3. **流式数据**：gRPC Stream → Tauri Event
   - Go 后端推送 gRPC 流
   - Rust 转换为 Tauri 事件 (`app.emit`)
   - 前端通过 `listen` 监听事件

### Rust 层职责

Rust 层是 Tauri IPC 到 gRPC 的桥梁，只做转发，业务逻辑留在 Go 后端：

- **gRPC 客户端** (`src/grpc_client.rs`)：管理与 Go 后端的 gRPC 连接
- **IPC Handlers** (`src/commands/*.rs`)：接收前端 invoke 调用，转换为 gRPC 请求
- **Sidecar 管理** (`src/sidecar.rs`)：启动/停止 Go 进程，健康检查，自动重启
- **配置管理** (`src/config.rs`)：窗口状态、主题、本地设置

### Go 后端模块化设计

`internal/` 下每个子目录是一个独立模块，遵循统一模式：
- 模块核心逻辑 (如 `rules/engine.go`)
- gRPC 服务实现 (`internal/grpc/*_service.go`) 桥接模块与 gRPC 接口
- 部分模块有独立的 Store 层操作 SQLite

`internal/grpc/server.go` 中的 `Server` 结构体聚合所有模块，`registerServices()` 注册所有 gRPC 服务。

### Sidecar 模式

Go 后端编译为平台特定的二进制文件，作为 Tauri 的 **sidecar** 运行：
- 构建脚本: `scripts/build_sidecar.sh`
- 输出目录: `desktop/src-tauri/bin/`
- 命名规范: `prismproxy-server-{target-triple}` (如 `aarch64-apple-darwin`)
- Tauri 配置: `externalBin` 字段引用 sidecar
- 健康检查: 定期 HTTP 检测，崩溃自动重启

### 数据库

使用 SQLite，数据库文件默认 `./prismproxy.db`，通过环境变量 `DB_PATH` 可配置。
迁移系统在 `internal/storage/migration.go`，服务器启动时自动执行。

### Proto 定义

Proto 文件在 `proto/` 目录：
- Go 代码生成: `proto/gen/go/`
- Rust 代码生成: `desktop/src-tauri/src/gen/` (build.rs 编译)
- 修改 proto 后需运行 `bash scripts/gen_proto.sh go` 重新生成

## 开发规范

- 使用中文注释，简洁明了，专业名词用英文
- 不要过度设计，保持代码简洁易读
- 提交前格式化代码 (`go fmt ./...`)
- **所有回复和思考过程使用中文**
- 每个后端接口要有单元测试

## 前端设计风格

### 主题
- 深色主题，GitHub 风格配色
- 参考产品：Reqable、Charles、VS Code
- 目标用户：开发者

### 颜色系统
使用 CSS 变量，定义在 `src/index.css`：

| 变量 | 值 | 用途 |
|------|-----|------|
| `--bg-primary` | `#0d1117` | 页面背景 |
| `--bg-secondary` | `#161b22` | 卡片、侧边栏 |
| `--bg-panel` | `#1c2128` | 输入框、下拉菜单 |
| `--border` | `#30363d` | 边框 |
| `--text-primary` | `#e6edf3` | 主文本 |
| `--text-secondary` | `#8b949e` | 次要文本 |
| `--text-tertiary` | `#565f89` | 三级文本 |
| `--blue` | `#58a6ff` | 链接、主按钮 |
| `--green` | `#3fb950` | 成功、GET |
| `--yellow` | `#d29922` | 警告、PUT |
| `--red` | `#f85149` | 错误、DELETE |
| `--purple` | `#bc8cff` | 特殊标记 |

### 字体
```css
font-family: 'Geist Variable', -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Noto Sans SC', sans-serif;
```

### 设计原则
1. **信息密度优先** - 开发者工具需要一屏展示更多信息
2. **操作效率** - 常用操作 1-2 次点击完成
3. **专业可信** - 深色主题、精确的数据展示
4. **渐进披露** - 复杂功能通过层级逐步展示

## 前端目录结构

详见 [docs/development/architecture.md](docs/development/architecture.md#目录结构) 和 [frontend/code-style.md](docs/development/frontend/code-style.md#目录结构)

## 端口说明

| 服务 | 端口 | 说明 |
|------|------|------|
| HTTP 代理 | 8888 | 抓包代理端口 |
| gRPC | 9090 | 原生 gRPC 服务 |
| HTTP/gRPC-Web | 8080 | 浏览器访问的 HTTP 端点 |
| 前端开发 | 3000 | Vite 开发服务器 |

## 开发规范

详细的开发规范请参考 [docs/development/](docs/development/)：

| 文档 | 内容 |
|------|------|
| [README.md](docs/development/README.md) | 规范总览和快速参考 |
| [frontend/code-style.md](docs/development/frontend/code-style.md) | 前端代码风格规范 |
| [frontend/testing.md](docs/development/frontend/testing.md) | 前端测试规范 |
| [backend/code-style.md](docs/development/backend/code-style.md) | 后端代码风格规范 |
| [backend/testing.md](docs/development/backend/testing.md) | 后端测试规范 |
| [git-workflow.md](docs/development/git-workflow.md) | Git 工作流和提交规范 |
| [architecture.md](docs/development/architecture.md) | 项目架构说明 |
