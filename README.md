# PrismProxy

> 类 Reqable 的跨平台 HTTP 调试代理工具，集成 AI 能力

## 功能特性

**核心功能：**
- HTTP/HTTPS 代理抓包 (MITM)
- 流量实时查看与过滤
- 规则引擎 (匹配/动作)
- 断点调试 (暂停/修改/释放)
- 请求重写 (7 种类型)
- API 集合管理 (类 Postman)
- 环境变量管理
- 脚本引擎 (表达式)
- 代码生成 (cURL/Python/JS/Go/Java/PHP)
- Diff 对比 (Headers/Body/JSON)
- 性能分析 (P50/P90/P99)
- 证书管理 (CA 生成/域名证书)
- 全文搜索与过滤器
- AI 分析 (OpenAI/Claude/Ollama)

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
├── internal/               # Go 后端模块
│   ├── ai/                 # AI 服务
│   ├── api/                # REST API
│   ├── cert/               # 证书管理
│   ├── codegen/            # 代码生成
│   ├── collection/         # API 集合
│   ├── debugger/           # 断点调试
│   ├── diff/               # Diff 对比
│   ├── environment/        # 环境变量
│   ├── grpc/               # gRPC 服务实现
│   ├── perf/               # 性能分析
│   ├── proxy/              # 代理引擎
│   ├── rewrite/            # 请求重写
│   ├── rules/              # 规则引擎
│   ├── script/             # 脚本引擎
│   ├── search/             # 搜索增强
│   ├── storage/            # 存储层
│   ├── traffic/            # 流量管理
│   └── websocket/          # WebSocket
├── proto/                  # Protobuf 定义
│   └── gen/go/             # 生成的 Go 代码
├── desktop/                # Tauri 桌面应用
│   ├── src/                # React 前端
│   └── src-tauri/          # Rust 配置
├── scripts/                # 构建脚本
└── docs/                   # 文档
```

## 快速开始

### 安装

```bash
# Debian/Ubuntu
sudo dpkg -i PrismProxy_1.0.0_arm64.deb

# 或直接运行
./prismproxy-desktop
```

### 从源码构建

```bash
# 后端
go build -o prismproxy-server ./cmd/server

# 前端
cd desktop && npm install && npm run build

# 桌面应用
cd desktop && npx tauri build
```

### 开发模式

```bash
# 启动 gRPC 服务器
go run ./cmd/server --port 9090

# 启动前端开发服务器
cd desktop && npm run dev
```

## 开发规范

- 使用中文注释，简洁明了，专业名词用英文
- 功能分支开发，PR review
- 提交前格式化代码
- 不要过度设计，保持代码简洁易读

## 许可证

MIT License
