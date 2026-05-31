# PrismProxy - HTTP 调试工具

## 项目概述
类似 Reqable 的跨平台 HTTP 调试代理工具，集成 AI 能力。

## 技术栈
- **后端**: Go + SQLite + gRPC
- **前端**: React + TypeScript + TailwindCSS
- **桌面端**: Tauri 2.x
- **通信**: gRPC (Protocol Buffers)

## 项目结构
```
prismProxy/
├── cmd/
│   ├── proxy/              # HTTP 代理入口
│   └── server/             # gRPC 服务器入口
├── internal/               # Go 后端模块 (18 个)
│   ├── ai/                 # AI 服务 (OpenAI/Claude/Ollama)
│   ├── api/                # REST API
│   ├── cert/               # 证书管理
│   ├── codegen/            # 代码生成
│   ├── collection/         # API 集合管理
│   ├── debugger/           # 断点调试
│   ├── diff/               # Diff 对比
│   ├── environment/        # 环境变量
│   ├── grpc/               # gRPC 服务实现 (14 个)
│   ├── perf/               # 性能分析
│   ├── proxy/              # 代理引擎 (HTTP/HTTPS MITM)
│   ├── rewrite/            # 请求重写
│   ├── rules/              # 规则引擎
│   ├── script/             # 脚本引擎 (expr-lang)
│   ├── search/             # 搜索增强
│   ├── storage/            # 存储层 (SQLite)
│   ├── traffic/            # 流量管理
│   └── websocket/          # WebSocket
├── proto/                  # Protobuf 定义 (15 个)
│   └── gen/go/             # 生成的 Go 代码
├── desktop/                # Tauri 桌面应用
│   ├── src/
│   │   ├── components/     # React 组件
│   │   │   ├── layout/     # 布局组件
│   │   │   ├── traffic/    # 流量组件
│   │   │   └── ui/         # UI 组件库
│   │   ├── pages/          # 页面 (11 个)
│   │   ├── services/       # API 服务层
│   │   ├── stores/         # 状态管理
│   │   └── types/          # TypeScript 类型
│   └── src-tauri/          # Rust 配置
├── scripts/                # 构建脚本
└── docs/                   # 文档
```

## 常用命令

```bash
# 构建后端
go build -o prismproxy-server ./cmd/server

# 运行 gRPC 服务器
go run ./cmd/server --port 9090

# 前端开发
cd desktop && npm run dev

# 构建桌面应用
cd desktop && npx tauri build

# 生成 Proto 代码
bash scripts/gen_proto.sh go

# 代码检查
go vet ./...
go fmt ./...
```

## 开发规范

- 使用中文注释，简洁明了，专业名词用英文
- 功能分支开发，PR review
- 提交前格式化代码
- 不要过度设计，保持代码简洁易读
- 提交前需要本地 review

## 当前状态

所有核心功能已实现，详见 docs/design/DESIGN.md。
