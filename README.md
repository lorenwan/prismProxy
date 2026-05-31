# PrismProxy

> 类 Reqable 的跨平台 HTTP 调试代理工具，集成 AI 能力

![Platform](https://img.shields.io/badge/platform-macOS%20|%20Linux%20|%20Windows-blue)
![License](https://img.shields.io/badge/license-MIT-green)

## 功能特性

### 核心功能

- **HTTP/HTTPS 代理抓包** - MITM 中间人攻击，查看加密流量
- **流量实时监控** - 实时查看请求/响应，支持过滤和搜索
- **规则引擎** - 自定义匹配规则和动作
- **断点调试** - 暂停请求，修改后释放
- **请求重写** - 7 种重写类型（Header/Body/URL/状态码等）
- **API 集合管理** - 类 Postman 的请求管理
- **环境变量** - 多环境切换（开发/测试/生产）
- **脚本引擎** - 表达式计算和数据处理
- **代码生成** - 一键生成 cURL/Python/JS/Go/Java/PHP 代码
- **Diff 对比** - 对比 Headers/Body/JSON 差异
- **性能分析** - P50/P90/P99 延迟统计
- **证书管理** - CA 证书生成和域名证书管理
- **全文搜索** - 跨请求搜索和过滤器
- **AI 分析** - 集成 OpenAI/Claude/Ollama 分析请求

### 代理控制

- **HTTP 代理开关** - 用户手动启动/停止代理服务
- **系统代理开关** - 一键设置系统代理，所有流量自动经过 PrismProxy
- **跨平台支持** - macOS (networksetup)、Linux (gsettings)、Windows (注册表)

## 技术栈

- **后端**: Go 1.21+ / SQLite / gRPC
- **前端**: React 19 / TypeScript 5 / TailwindCSS 4
- **桌面端**: Tauri 2.x (Rust)
- **通信协议**: Protocol Buffers / gRPC-Web

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
│   ├── script/             # 脚本引擎
│   ├── search/             # 搜索增强
│   ├── storage/            # 存储层 (SQLite)
│   ├── traffic/            # 流量管理
│   └── websocket/          # WebSocket
├── proto/                  # Protobuf 定义 (15 个)
│   └── gen/go/             # 生成的 Go 代码
├── desktop/                # Tauri 桌面应用
│   ├── src/                # React 前端
│   │   ├── components/     # UI 组件
│   │   ├── pages/          # 页面 (11 个)
│   │   ├── services/       # API 服务层
│   │   └── types/          # TypeScript 类型
│   └── src-tauri/          # Rust 配置
├── scripts/                # 构建脚本
└── docs/                   # 设计文档
```

## 快速开始

### 安装

#### 下载预编译版本

从 [GitHub Releases](https://github.com/lorenwan/prismProxy/releases) 下载对应平台的安装包：

- **macOS**: `PrismProxy_1.0.0_aarch64.dmg`
- **Linux (Debian/Ubuntu)**: `PrismProxy_1.0.0_arm64.deb`
- **Linux (RHEL/Fedora)**: `PrismProxy_1.0.0-1.aarch64.rpm`
- **Windows**: `PrismProxy_1.0.0_x64-setup.exe`

#### Debian/Ubuntu 安装

```bash
sudo dpkg -i PrismProxy_1.0.0_arm64.deb
```

#### 直接运行

```bash
chmod +x prismproxy-desktop
./prismproxy-desktop
```

### 从源码构建

#### 前置依赖

- Go 1.21+
- Node.js 18+
- Rust 1.70+
- Protocol Buffers Compiler (protoc)

```bash
# 安装 protoc (macOS)
brew install protobuf

# 安装 protoc (Ubuntu)
sudo apt install protobuf-compiler

# 安装 Go 插件
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 安装 Tauri CLI
cargo install tauri-cli
```

#### 构建步骤

```bash
# 1. 克隆仓库
git clone https://github.com/lorenwan/prismProxy.git
cd prismProxy

# 2. 生成 Proto 代码
bash scripts/gen_proto.sh go

# 3. 构建后端
go build -o prismproxy-server ./cmd/server

# 4. 构建前端和桌面应用
cd desktop
npm install
npm run build
npx tauri build
```

### 开发模式

```bash
# 启动桌面应用开发模式（推荐）
cd desktop
npm run dev

# 这会同时启动：
# - 前端开发服务器 (http://localhost:3000)
# - Tauri 桌面窗口
# - Go sidecar (gRPC 服务器)

# 如果只需要前端开发（不启动 Tauri 窗口）
cd desktop
npm run frontend:dev

# 单独启动后端
go run ./cmd/server --port 9090
```

## 使用说明

### 代理配置

1. 启动 PrismProxy
2. 进入 **设置** 页面
3. 点击 **HTTP 代理服务** 开关启动代理
4. （可选）点击 **系统代理** 开关自动设置系统代理

### 手动配置代理

如果不想使用系统代理，可以手动配置浏览器：

1. 打开浏览器代理设置
2. 设置 HTTP 代理为 `127.0.0.1:8888`
3. 设置 HTTPS 代理为 `127.0.0.1:8888`

### HTTPS 抓包

1. 在设置页面点击 **下载 CA 证书**
2. 安装证书到系统信任列表
3. 启用 **MITM** 开关
4. 即可查看 HTTPS 请求内容

### AI 分析

1. 在设置页面配置 AI Provider（OpenAI/Claude/Ollama）
2. 填写 API Key 和模型信息
3. 在流量页面选择请求，点击 **AI 分析**

## 端口说明

| 服务 | 端口 | 说明 |
|------|------|------|
| HTTP 代理 | 8888 | 抓包代理端口 |
| gRPC | 9090 | 后端 API 服务 |
| 前端开发 | 3000 | Vite 开发服务器 |

## 开发规范

- 使用中文注释，简洁明了，专业名词用英文
- 功能分支开发，PR review
- 提交前格式化代码（`go fmt ./...`）
- 不要过度设计，保持代码简洁易读
- 提交前需要本地 review

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！

## 致谢

- [Reqable](https://reqable.com/) - 设计灵感
- [Tauri](https://tauri.app/) - 桌面应用框架
- [gRPC](https://grpc.io/) - 通信协议
