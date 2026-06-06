# 项目架构

## 概述

PrismProxy 采用 **前端 → Rust → Go** 三层架构。

### 职责划分

| 层级 | 职责 | 技术 |
|------|------|------|
| 前端 | UI 展示、用户交互、状态管理 | React, TypeScript, TailwindCSS |
| Rust | IPC 桥梁、Sidecar 管理、系统集成 | Tauri, tonic |
| Go | 业务逻辑、数据存储、gRPC 服务 | gRPC, SQLite |

---

## 前端架构

### 目录结构

```
desktop/src/
├── components/          # 通用组件（跨功能复用）
│   ├── layout/          # 布局组件（Header, Sidebar, StatusBar）
│   └── ui/              # UI 基础组件（shadcn/ui）
├── features/            # 功能模块（按业务拆分）
│   └── [module]/
│       ├── components/  # 模块专属组件
│       ├── index.ts     # 统一导出
│       ├── [module]Store.ts
│       └── [module]Service.ts
├── hooks/               # 自定义 hooks
├── lib/                 # 工具函数
├── pages/               # 页面组件（路由入口）
├── services/            # API 服务（Tauri IPC 调用）
└── types/               # 类型定义
```

### 组件规范

- 使用 shadcn/ui 组件库
- 按功能模块组织组件
- 同类组件出现超 2 次就封装复用

### 状态管理

- 使用 Zustand
- 每个功能模块独立 Store
- Store 与 Service 分离

### 路由

- 使用 React Router v7
- 页面组件作为路由入口

---

## 后端架构

### 目录结构

```
internal/
├── traffic/         # 流量模块
├── rules/           # 规则模块
├── ai/              # AI 模块
├── grpc/            # gRPC 服务实现（14 个服务）
├── storage/         # 存储层（SQLite）
├── proxy/           # HTTP 代理引擎
├── cert/            # 证书管理
├── codegen/         # 代码生成
├── collection/      # API 集合管理
├── debugger/        # 断点调试
├── diff/            # Diff 对比
├── environment/     # 环境变量
├── perf/            # 性能分析
├── rewrite/         # 请求重写
├── script/          # 脚本引擎
├── search/          # 搜索增强
└── websocket/       # WebSocket
```

### 模块化设计

- 每个模块独立目录
- 模块核心逻辑 + gRPC 服务实现
- 部分模块有独立 Store 层

### gRPC 服务

- 14 个服务实现
- 统一错误处理（gRPC Status Code）
- 流式数据支持（Server Streaming）

---

## 通信协议

### Tauri IPC (前端 → Rust)

- 前端调用 `invoke('command', args)`
- Rust 层接收 IPC 调用，转换为 gRPC 请求

### gRPC (Rust → Go)

- Rust 使用 tonic 客户端连接 Go 后端
- 管理连接池、重连、超时

### Tauri Event (Rust → 前端)

- Go 后端推送 gRPC 流
- Rust 转换为 Tauri 事件
- 前端通过 `listen` 监听事件

---

## 设计规范

引用 [DESIGN.md](../design/DESIGN.md)

- 深色主题，GitHub 风格配色
- 使用 CSS 变量
- shadcn/ui 组件库

## 产品说明

引用 [PRODUCT.md](../../PRODUCT.md)

- 产品定位：HTTP 调试代理工具
- 目标用户：开发者
- 核心功能：抓包、规则、AI 分析
