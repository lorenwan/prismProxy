# PrismProxy - HTTP 调试工具

## 项目概述
类似 Reqable 的现代化 HTTP 调试和抓包工具，集成 AI 能力。

## 技术栈
- 后端: Go + Gin + SQLite + WebSocket
- 前端: React + TypeScript + TailwindCSS
- 桌面端: Tauri
- AI: OpenAI/Claude/Ollama 适配层

## 开发规范
- 使用中文注释，简洁明了，专业名词用英文
- 所有 feature 变更使用分支开发
- 代码格式化后提交
- 不要过度设计，保持代码简洁易读
- 提交前需要本地 review

## 项目结构
```
prismProxy/
├── cmd/
│   └── proxy/
│       └── main.go          # 程序入口
├── internal/
│   ├── proxy/               # 代理引擎
│   ├── traffic/             # 流量管理
│   ├── rules/               # 规则引擎
│   ├── debugger/            # 断点调试
│   ├── ai/                  # AI 服务
│   ├── collection/          # API 集合
│   ├── storage/             # 存储层
│   ├── api/                 # REST API
│   └── websocket/           # WebSocket
├── web/                     # 前端代码
├── docs/                    # 文档
├── DESIGN.md                # 详细设计文档
└── README.md
```

## 常用命令
```bash
# 构建
go build -o prismproxy ./cmd/proxy

# 运行
./prismproxy

# 测试
go test ./...

# 格式化
go fmt ./...
gofmt -w .
```

## 当前任务
按照 DESIGN.md 中的实施计划，从 Phase 1 开始实现：
1. 项目结构重组
2. SQLite 存储层
3. 数据库迁移系统
4. 代理引擎重构
5. 基础 REST API

## 详细设计
请参考 DESIGN.md 文件，包含所有模块的详细设计。
