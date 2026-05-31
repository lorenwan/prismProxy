# PrismProxy 桌面端详细实施计划

> 每个阶段的任务拆分、验收标准、依赖关系、预估工时

---

## Phase 1: Protobuf 定义 + gRPC 基础骨架

**目标**: 完成所有 proto 定义，生成代码，搭建 gRPC 服务器框架

**预估工时**: 5-7 天

### 1.1 项目结构初始化 (Day 1)

```
新建目录:
├── proto/                      # Protobuf 定义
├── cmd/server/                 # gRPC 服务器入口
├── internal/grpc/              # gRPC 服务实现
├── scripts/                    # 构建脚本
│   ├── gen_proto.sh            # 代码生成脚本
│   └── build_sidecar.sh        # Sidecar 构建脚本
└── desktop/                    # Tauri 项目
    ├── src-tauri/
    └── src/
```

**任务清单**:
- [ ] 创建 `proto/` 目录
- [ ] 创建 `cmd/server/` 目录
- [ ] 创建 `internal/grpc/` 目录
- [ ] 创建 `scripts/` 目录
- [ ] 更新 `go.mod` 添加 gRPC 依赖
- [ ] 安装 protoc 编译器
- [ ] 安装 Go/TS protoc 插件

**依赖安装**:
```bash
# protoc 编译器
apt install -y protobuf-compiler

# Go 插件
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# TypeScript 插件
npm install -g @protobuf-ts/plugin
```

**验收标准**:
- [ ] `protoc --version` 输出版本号
- [ ] `protoc-gen-go --version` 输出版本号
- [ ] `protoc-gen-go-grpc --version` 输出版本号

### 1.2 Proto 文件定义 (Day 2-3)

**任务清单**:

**traffic.proto** (流量服务):
- [ ] 定义 `TrafficEntry` 消息
- [ ] 定义 `RequestData` / `ResponseData` 消息
- [ ] 定义 `ListTrafficRequest` / `ListTrafficResponse`
- [ ] 定义 `TrafficStats` 消息
- [ ] 定义 `TrafficEvent` 消息 (实时事件)
- [ ] 定义 `TrafficService` 服务 (7 个 RPC)
  - `GetTraffic` - 获取单条
  - `ListTraffic` - 列表查询
  - `DeleteTraffic` - 删除
  - `ClearTraffic` - 清空
  - `GetTrafficStats` - 统计
  - `UpdateBookmark` - 更新书签
  - `WatchTraffic` - 实时流

**rules.proto** (规则服务):
- [ ] 定义 `Rule` / `RuleMatch` / `RuleAction` 消息
- [ ] 定义 `ModifySpec` / `BlockSpec` 消息
- [ ] 定义 `RulesService` 服务 (9 个 RPC)
  - `ListRules` / `GetRule` / `CreateRule` / `UpdateRule` / `DeleteRule`
  - `ToggleRule` / `ReorderRules`
  - `ImportRules` / `ExportRules`

**breakpoints.proto** (断点服务):
- [ ] 定义 `Breakpoint` / `BreakpointSession` 消息
- [ ] 定义 `BreakpointEvent` 消息
- [ ] 定义 `BreakpointsService` 服务 (9 个 RPC)
  - `ListBreakpoints` / `CreateBreakpoint` / `UpdateBreakpoint` / `DeleteBreakpoint`
  - `ToggleBreakpoint`
  - `ListSessions` / `ReleaseSession` / `ModifySession` / `DropSession`
  - `WatchBreakpoints` - 断点命中流

**rewrites.proto** (重写服务):
- [ ] 定义 `RewriteRule` / `RewriteAction` 消息
- [ ] 定义 `RewritesService` 服务 (6 个 RPC)
  - `ListRewrites` / `CreateRewrite` / `UpdateRewrite` / `DeleteRewrite`
  - `ToggleRewrite` / `ReorderRewrites`

**collections.proto** (集合服务):
- [ ] 定义 `Collection` / `CollectionItem` / `Folder` 消息
- [ ] 定义 `APIRequest` / `RequestBody` / `AuthConfig` / `Test` 消息
- [ ] 定义 `CollectionsService` 服务 (10 个 RPC)
  - `ListCollections` / `GetCollection` / `CreateCollection` / `UpdateCollection` / `DeleteCollection`
  - `ListRequests` / `CreateRequest` / `UpdateRequest` / `DeleteRequest`
  - `RunRequest`

**environments.proto** (环境服务):
- [ ] 定义 `Environment` / `Variable` 消息
- [ ] 定义 `EnvironmentsService` 服务 (5 个 RPC)
  - `ListEnvironments` / `CreateEnvironment` / `UpdateEnvironment` / `DeleteEnvironment`
  - `ActivateEnvironment`

**ai.proto** (AI 服务):
- [ ] 定义 `ChatMessage` / `ChatRequest` / `ChatChunk` 消息
- [ ] 定义 `AnalysisResult` / `SecurityReport` / `TestCase` 消息
- [ ] 定义 `AIService` 服务 (5 个 RPC)
  - `Chat` - 流式聊天
  - `AnalyzeTraffic` - 流量分析
  - `SecurityCheck` - 安全检测
  - `GenerateTests` - 测试生成
  - `GetProviders` - 获取可用 Provider

**codegen.proto** (代码生成):
- [ ] 定义 `CodeGenRequest` / `CodeGenResult` 消息
- [ ] 定义 `CodeGenService` 服务 (2 个 RPC)
  - `GenerateFromTraffic` - 从流量生成
  - `GenerateFromRequest` - 从请求定义生成

**system.proto** (系统服务):
- [ ] 定义 `SystemStatus` / `Settings` / `ProxySettings` / `AISettings` 消息
- [ ] 定义 `SystemService` 服务 (4 个 RPC)
  - `GetStatus` / `GetSettings` / `UpdateSettings` / `DownloadCert`

**公共类型** (common.proto):
- [ ] 定义 `Empty` 消息
- [ ] 定义 `Pagination` 消息
- [ ] 定义 `TimeRange` 消息

**验收标准**:
- [ ] 所有 .proto 文件通过 `protoc --lint_out=. validation` 检查
- [ ] proto 文件间无循环依赖
- [ ] 消息命名与现有 Go 结构体一致

### 1.3 代码生成 (Day 3)

**gen_proto.sh**:
```bash
#!/bin/bash
set -e

PROTO_DIR="./proto"
GO_OUT="./proto/gen/go"
TS_OUT="./desktop/src/proto"

mkdir -p $GO_OUT $TS_OUT

# 生成 Go 代码
protoc \
  --proto_path=$PROTO_DIR \
  --go_out=$GO_OUT --go_opt=paths=source_relative \
  --go-grpc_out=$GO_OUT --go-grpc_opt=paths=source_relative \
  $PROTO_DIR/*.proto

# 生成 TypeScript 代码
protoc \
  --proto_path=$PROTO_DIR \
  --ts_out=$TS_OUT \
  $PROTO_DIR/*.proto

echo "Protobuf 代码生成完成"
```

**任务清单**:
- [ ] 编写 `scripts/gen_proto.sh`
- [ ] 运行脚本生成 Go 代码到 `proto/gen/go/`
- [ ] 运行脚本生成 TS 代码到 `desktop/src/proto/`
- [ ] 验证生成的 Go 代码可编译
- [ ] 验证生成的 TS 代码类型正确

**验收标准**:
- [ ] `go build ./proto/gen/go/...` 通过
- [ ] 生成的 TS 文件包含正确的类型定义

### 1.4 gRPC 服务器骨架 (Day 4-5)

**cmd/server/main.go**:
```go
func main() {
    // 1. 初始化存储
    // 2. 初始化各模块 (traffic, rules, debugger, rewrite, ...)
    // 3. 创建 gRPC 服务器
    // 4. 注册所有服务
    // 5. 启动监听
    // 6. 优雅关闭
}
```

**internal/grpc/server.go**:
```go
type Server struct {
    grpcServer *grpc.Server
    traffic    *traffic.Manager
    rules      *rules.Engine
    debugger   *debugger.Debugger
    rewrite    *rewrite.Engine
    collection *collection.Manager
    environment *environment.Manager
    ai         *ai.Service
    codegen    *codegen.Generator
    storage    *storage.Storage
}

func New(cfg Config) *Server { ... }
func (s *Server) Start(addr string) error { ... }
func (s *Server) Stop() { ... }
```

**internal/grpc/traffic_service.go**:
```go
type TrafficService struct {
    pb.UnimplementedTrafficServiceServer
    manager *traffic.Manager
    hub     *websocket.Hub
}

func (s *TrafficService) GetTraffic(ctx context.Context, req *pb.GetTrafficRequest) (*pb.TrafficEntry, error) { ... }
func (s *TrafficService) ListTraffic(ctx context.Context, req *pb.ListTrafficRequest) (*pb.ListTrafficResponse, error) { ... }
func (s *TrafficService) WatchTraffic(req *pb.WatchTrafficRequest, stream pb.TrafficService_WatchTrafficServer) error { ... }
```

**任务清单**:
- [ ] 实现 `cmd/server/main.go` 入口
- [ ] 实现 `internal/grpc/server.go` 服务器框架
- [ ] 实现 `internal/grpc/traffic_service.go` (第一个完整服务)
- [ ] 实现 gRPC 拦截器 (日志、panic 恢复)
- [ ] 实现 gRPC-Web 代理 (grpcweb.WrapServer)
- [ ] 实现优雅关闭 (SIGTERM/SIGINT 处理)
- [ ] 添加健康检查服务 (grpc.health.v1)

**gRPC-Web 代理**:
```go
import "github.com/improbable-eng/grpc-web/go/grpcweb"

wrappedServer := grpcweb.WrapServer(grpcServer,
    grpcweb.WithCorsForRegisteredEndpointsOnly(false),
    grpcweb.WithOriginFunc(func(origin string) bool { return true }),
)

httpServer := &http.Server{
    Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if wrappedServer.IsGrpcWebRequest(r) {
            wrappedServer.ServeHTTP(w, r)
            return
        }
        // 其他 HTTP 处理
    }),
}
```

**验收标准**:
- [ ] `go build ./cmd/server/` 编译通过
- [ ] 启动服务器后可接受 gRPC 连接
- [ ] gRPC-Web 请求可被浏览器发送
- [ ] `grpcurl -plaintext localhost:9090 list` 显示服务列表

### 1.5 集成测试 (Day 6-7)

**任务清单**:
- [ ] 编写 TrafficService 单元测试
- [ ] 编写 gRPC 集成测试 (启动服务器 + 客户端调用)
- [ ] 验证流式 RPC (WatchTraffic) 工作正常
- [ ] 验证 gRPC-Web 代理工作正常
- [ ] 性能基准测试 (对比 REST)

**验收标准**:
- [ ] `go test ./internal/grpc/...` 全部通过
- [ ] gRPC 请求延迟 < 1ms (本地)
- [ ] 流式推送延迟 < 50ms

---

## Phase 2: 完善所有 gRPC 服务

**目标**: 实现所有 gRPC 服务，替换 REST API

**预估工时**: 7-10 天

### 2.1 RulesService (Day 1-2)

**internal/grpc/rules_service.go**:
```go
type RulesService struct {
    pb.UnimplementedRulesServiceServer
    engine *rules.Engine
}
```

**任务清单**:
- [ ] 实现 `ListRules` - 调用 engine.ListRules()
- [ ] 实现 `GetRule` - 调用 engine.GetRule(id)
- [ ] 实现 `CreateRule` - 调用 engine.CreateRule(rule)
- [ ] 实现 `UpdateRule` - 调用 engine.UpdateRule(rule)
- [ ] 实现 `DeleteRule` - 调用 engine.DeleteRule(id)
- [ ] 实现 `ToggleRule` - 调用 engine.ToggleRule(id)
- [ ] 实现 `ReorderRules` - 调用 engine.ReorderRules(ids)
- [ ] 实现 `ImportRules` - JSON/Postman 格式解析
- [ ] 实现 `ExportRules` - JSON 格式导出
- [ ] 编写单元测试

**proto ↔ Go 转换函数**:
```go
func ruleToProto(r *rules.Rule) *pb.Rule { ... }
func protoToRule(p *pb.Rule) *rules.Rule { ... }
func matchToProto(m *rules.RuleMatch) *pb.RuleMatch { ... }
func protoToMatch(p *pb.RuleMatch) *rules.RuleMatch { ... }
```

**验收标准**:
- [ ] 所有 CRUD 操作正常
- [ ] 导入/导出格式正确
- [ ] 单元测试覆盖率 > 80%

### 2.2 BreakpointsService (Day 2-3)

**internal/grpc/breakpoints_service.go**:
```go
type BreakpointsService struct {
    pb.UnimplementedBreakpointsServiceServer
    debugger *debugger.Debugger
}
```

**任务清单**:
- [ ] 实现断点 CRUD (5 个 RPC)
- [ ] 实现会话管理 (4 个 RPC)
- [ ] 实现 `WatchBreakpoints` 流式推送
  - 当断点命中时，通过 channel 推送事件
  - 客户端断开时清理订阅
- [ ] 编写单元测试

**流式实现**:
```go
func (s *BreakpointsService) WatchBreakpoints(
    _ *emptypb.Empty,
    stream pb.BreakpointsService_WatchBreakpointsServer,
) error {
    ch := s.debugger.Subscribe()
    defer s.debugger.Unsubscribe(ch)

    for {
        select {
        case event := <-ch:
            if err := stream.Send(event); err != nil {
                return err
            }
        case <-stream.Context().Done():
            return nil
        }
    }
}
```

**验收标准**:
- [ ] 断点命中时前端能收到实时通知
- [ ] 会话释放/修改/丢弃操作正常
- [ ] 多客户端同时监听不冲突

### 2.3 RewritesService (Day 3-4)

**任务清单**:
- [ ] 实现重写规则 CRUD (4 个 RPC)
- [ ] 实现 Toggle/Reorder (2 个 RPC)
- [ ] 编写单元测试

**验收标准**:
- [ ] 重写规则 CRUD 正常
- [ ] 规则优先级排序正确

### 2.4 CollectionsService + EnvironmentsService (Day 4-6)

**任务清单**:
- [ ] 实现集合 CRUD (5 个 RPC)
- [ ] 实现请求 CRUD (4 个 RPC)
- [ ] 实现 `RunRequest` - 执行请求并返回结果
  - 解析环境变量 `{{variable}}`
  - 执行认证 (Basic/Bearer/API-Key)
  - 记录到流量
- [ ] 实现环境 CRUD (4 个 RPC)
- [ ] 实现 `ActivateEnvironment` (1 个 RPC)
- [ ] 编写单元测试

**环境变量替换**:
```go
func (s *CollectionsService) replaceVariables(
    input string,
    env *environment.Environment,
) string {
    for _, v := range env.Variables {
        input = strings.ReplaceAll(input, "{{"+v.Key+"}}", v.Value)
    }
    return input
}
```

**验收标准**:
- [ ] 集合/请求 CRUD 正常
- [ ] 环境变量替换正确
- [ ] `RunRequest` 能发送真实请求并记录

### 2.5 AIService (Day 6-8)

**任务清单**:
- [ ] 实现 `Chat` 流式 RPC
  - 接收 ChatRequest
  - 调用 AI Provider 的 StreamChat
  - 逐块推送 ChatChunk
- [ ] 实现 `AnalyzeTraffic`
  - 从数据库加载流量
  - 调用 service.AnalyzeTraffic
  - 返回 AnalysisResult
- [ ] 实现 `SecurityCheck`
  - 加载单条流量
  - 调用 service.SecurityCheck
  - 返回 SecurityReport
- [ ] 实现 `GenerateTests`
  - 加载流量
  - 调用 service.GenerateTests
  - 返回 TestCase 列表
- [ ] 实现 `GetProviders`
- [ ] 编写单元测试

**流式 Chat 实现**:
```go
func (s *AIService) Chat(
    req *pb.ChatRequest,
    stream pb.AIService_ChatServer,
) error {
    ctx := stream.Context()
    ch, err := s.ai.StreamChat(ctx, protoToChatRequest(req))
    if err != nil {
        return err
    }

    for chunk := range ch {
        if err := stream.Send(&pb.ChatChunk{
            Content:  chunk.Content,
            Done:     chunk.Done,
            Provider: chunk.Provider,
        }); err != nil {
            return err
        }
    }
    return nil
}
```

**验收标准**:
- [ ] 流式聊天正常工作
- [ ] 流量分析返回正确结果
- [ ] 安全检测返回正确报告
- [ ] 测试生成返回正确用例

### 2.6 SystemService + CodeGenService (Day 8-9)

**任务清单**:
- [ ] 实现 `GetStatus` - 系统状态
- [ ] 实现 `GetSettings` / `UpdateSettings` - 配置管理
- [ ] 实现 `DownloadCert` - CA 证书下载
- [ ] 实现代码生成 (2 个 RPC)
  - cURL
  - Python (requests)
  - JavaScript (fetch)
  - Go (net/http)
- [ ] 编写单元测试

**验收标准**:
- [ ] 系统状态正确
- [ ] 配置读写正常
- [ ] 代码生成格式正确

### 2.7 集成测试 + 性能测试 (Day 9-10)

**任务清单**:
- [ ] 端到端集成测试
  - 启动 gRPC 服务器
  - 创建规则 → 触发规则 → 验证结果
  - 创建断点 → 发送请求 → 验证暂停
  - 发送聊天消息 → 验证流式响应
- [ ] 性能基准测试
  - gRPC vs REST QPS 对比
  - 流式推送延迟测试
  - 内存占用测试
- [ ] 编写测试报告

**验收标准**:
- [ ] 所有集成测试通过
- [ ] gRPC QPS > REST QPS 的 2x
- [ ] 流式推送延迟 < 50ms

---

## Phase 3: Tauri 桌面端

**目标**: 完成 Tauri 桌面壳，前端迁移到 gRPC-Web

**预估工时**: 7-10 天

### 3.1 Tauri 项目初始化 (Day 1)

**任务清单**:
- [ ] 安装 Rust 工具链
  ```bash
  curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
  ```
- [ ] 安装 Tauri CLI
  ```bash
  cargo install tauri-cli
  ```
- [ ] 初始化 Tauri 项目
  ```bash
  cargo tauri init
  ```
- [ ] 配置 `tauri.conf.json`
  - 窗口大小 1400x900
  - 最小大小 1000x600
  - 标题 "PrismProxy"
  - 图标配置
- [ ] 配置 sidecar 路径
  ```json
  {
    "bundle": {
      "externalBin": ["bin/prismproxy-server"]
    }
  }
  ```
- [ ] 验证 `cargo tauri dev` 可启动

**验收标准**:
- [ ] `cargo tauri dev` 启动空白窗口
- [ ] 窗口大小、标题正确
- [ ] 开发服务器可访问

### 3.2 Go Sidecar 集成 (Day 2-3)

**任务清单**:
- [ ] 编写 `src-tauri/src/main.rs`
  ```rust
  // 启动 Go sidecar
  let sidecar = app.shell().sidecar("prismproxy-server").unwrap();
  let (rx, child) = sidecar
      .args(&["--port", "9090", "--proxy-port", "8080"])
      .spawn()
      .expect("Failed to spawn sidecar");
  ```
- [ ] 实现 sidecar 生命周期管理
  - 启动时自动启动
  - 窗口关闭时自动停止
  - 崩溃时自动重启
- [ ] 实现 sidecar 日志转发
  ```rust
  // 读取 sidecar stdout/stderr
  tok::spawn(async move {
      while let Some(event) = rx.recv().await {
          match event {
              CommandEvent::Stdout(line) => {
                  window.emit("sidecar-log", line).unwrap();
              }
              CommandEvent::Stderr(line) => {
                  window.emit("sidecar-error", line).unwrap();
              }
              _ => {}
          }
      }
  });
  ```
- [ ] 实现 sidecar 健康检查
  - 启动后轮询 gRPC 端口
  - 超时则报错
- [ ] 验证 sidecar 自动启动/停止

**验收标准**:
- [ ] 应用启动时 Go 服务器自动启动
- [ ] 应用关闭时 Go 服务器自动停止
- [ ] Go 服务器崩溃时自动重启
- [ ] 日志可查看

### 3.3 前端迁移 (Day 3-5)

**任务清单**:
- [ ] 安装前端依赖
  ```bash
  cd desktop
  npm install @protobuf-ts/plugin @protobuf-ts/grpcweb-transport
  npm install @tauri-apps/api @tauri-apps/plugin-shell
  ```
- [ ] 生成 TypeScript 客户端代码
  ```bash
  protoc --ts_out=src/proto/ --proto_path=../proto/ ../proto/*.proto
  ```
- [ ] 实现 `services/grpc.ts` 连接管理
  ```typescript
  const transport = new GrpcWebFetchTransport({
    baseUrl: "http://localhost:9090",
  });
  ```
- [ ] 迁移 TrafficPage
  - `getTrafficList` → `trafficClient.listTraffic()`
  - WebSocket → `trafficClient.watchTraffic()` 流
- [ ] 迁移 RulesPage
  - REST 调用 → `rulesClient.xxx()`
- [ ] 迁移 BreakpointsPage
  - REST 调用 → `breakpointsClient.xxx()`
  - 断点命中通知 → `breakpointsClient.watchBreakpoints()` 流
- [ ] 迁移 AiPage
  - REST 流式 → `aiClient.chat()` 流
- [ ] 迁移 SettingsPage
  - REST 调用 → `systemClient.xxx()`
- [ ] 删除旧的 REST 服务层
- [ ] 删除 WebSocket hook (用 gRPC 流替代)

**验收标准**:
- [ ] 所有页面功能正常
- [ ] 实时流量推送正常
- [ ] 断点命中通知正常
- [ ] AI 聊天流式输出正常
- [ ] 无 TypeScript 编译错误

### 3.4 深色主题 + UI 完善 (Day 5-6)

**任务清单**:
- [ ] 完善深色主题样式
- [ ] 添加流量列表状态码颜色
- [ ] 添加请求/响应详情面板
- [ ] 添加 JSON 格式化显示
- [ ] 添加搜索/过滤 UI
- [ ] 添加侧边栏导航高亮
- [ ] 添加状态栏信息

**验收标准**:
- [ ] UI 与 Reqable 视觉风格接近
- [ ] 所有交互流畅
- [ ] 无布局错乱

### 3.5 构建脚本 (Day 6-7)

**scripts/build_sidecar.sh**:
```bash
#!/bin/bash
set -e

PLATFORM=$1  # windows, macos, linux
ARCH=$2      # x86_64, aarch64

case $PLATFORM in
  windows)
    GOOS=windows GOARCH=amd64 go build -o desktop/src-tauri/bin/prismproxy-server.exe ./cmd/server/
    ;;
  macos)
    if [ "$ARCH" = "aarch64" ]; then
      GOOS=darwin GOARCH=arm64 go build -o desktop/src-tauri/bin/prismproxy-server-aarch64-apple-darwin ./cmd/server/
    else
      GOOS=darwin GOARCH=amd64 go build -o desktop/src-tauri/bin/prismproxy-server-x86_64-apple-darwin ./cmd/server/
    fi
    ;;
  linux)
    GOOS=linux GOARCH=amd64 go build -o desktop/src-tauri/bin/prismproxy-server-x86_64-unknown-linux-gnu ./cmd/server/
    ;;
esac
```

**任务清单**:
- [ ] 编写 `scripts/build_sidecar.sh`
- [ ] 配置 Tauri 构建目标 (Windows/macOS/Linux)
- [ ] 配置图标 (多分辨率)
  - 32x32, 128x128, 256x256, 512x512
  - .icns (macOS)
  - .ico (Windows)
- [ ] 测试各平台构建

**验收标准**:
- [ ] `cargo tauri build` 在当前平台成功
- [ ] 生成的安装包可正常安装运行

---

## Phase 4: 打包发布

**目标**: 多平台打包、自动更新、文档

**预估工时**: 5-7 天

### 4.1 Windows 打包 (Day 1)

**任务清单**:
- [ ] 交叉编译 Go sidecar (Windows amd64)
  ```bash
  GOOS=windows GOARCH=amd64 CGO_ENABLED=1 \
    CC=x86_64-w64-mingw32-gcc \
    go build -o prismproxy-server.exe ./cmd/server/
  ```
- [ ] 配置 Windows 安装程序
  - NSIS 或 WiX
  - 自动添加 PATH
  - 创建桌面快捷方式
  - 文件关联 .har
- [ ] 测试 Windows 安装/卸载

**验收标准**:
- [ ] .msi 安装包可在 Windows 10/11 安装
- [ ] 应用可正常启动运行
- [ ] 卸载后无残留

### 4.2 macOS 打包 (Day 2)

**任务清单**:
- [ ] 交叉编译 Go sidecar (macOS arm64 + amd64)
- [ ] 配置 macOS 签名 (可选)
  - Developer ID 证书
  - 公证 (notarization)
- [ ] 生成 .dmg 安装包
- [ ] 测试 macOS 安装/运行

**验收标准**:
- [ ] .dmg 可在 macOS 安装
- [ ] arm64 和 x86_64 均可运行
- [ ] Gatekeeper 不阻止运行

### 4.3 Linux 打包 (Day 3)

**任务清单**:
- [ ] 编译 Go sidecar (Linux amd64)
- [ ] 生成 .deb 包
  ```bash
  cargo tauri build --target deb
  ```
- [ ] 生成 .AppImage (可选)
- [ ] 测试 Ubuntu/Debian 安装

**验收标准**:
- [ ] .deb 可在 Ubuntu 22.04+ 安装
- [ ] 应用可正常启动运行

### 4.4 自动更新 (Day 4-5)

**任务清单**:
- [ ] 配置 Tauri Updater
  ```json
  {
    "updater": {
      "active": true,
      "endpoints": [
        "https://releases.prismproxy.com/{{target}}/{{arch}}/{{current_version}}"
      ],
      "pubkey": "dW50cnVzdGVkIGNvbW1lbnQ6..."
    }
  }
  ```
- [ ] 搭建更新服务器 (GitHub Releases 或自建)
- [ ] 实现更新检查逻辑
- [ ] 实现下载和安装
- [ ] 实现更新通知 UI

**验收标准**:
- [ ] 新版本发布后客户端自动检测
- [ ] 用户确认后自动下载安装
- [ ] 更新后数据不丢失

### 4.5 文档 + README (Day 5-6)

**任务清单**:
- [ ] 编写 README.md
  - 项目介绍
  - 功能特性
  - 安装方法
  - 使用说明
  - 开发指南
- [ ] 编写 CONTRIBUTING.md
  - 开发环境搭建
  - 代码规范
  - PR 流程
- [ ] 编写 CHANGELOG.md
- [ ] 截图/GIF 演示

**验收标准**:
- [ ] README 清晰完整
- [ ] 新开发者可按文档搭建环境

### 4.6 CI/CD (Day 6-7)

**GitHub Actions**:
```yaml
name: Build & Release
on:
  push:
    tags: ['v*']

jobs:
  build:
    strategy:
      matrix:
        include:
          - os: ubuntu-latest
            target: x86_64-unknown-linux-gnu
          - os: macos-latest
            target: aarch64-apple-darwin
          - os: windows-latest
            target: x86_64-pc-windows-msvc
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v4
      - uses: dtolnay/rust-toolchain@stable
      - run: ./scripts/build_sidecar.sh
      - run: cargo tauri build
      - uses: softprops/action-gh-release@v1
        with:
          files: |
            desktop/src-tauri/target/release/bundle/*
```

**任务清单**:
- [ ] 配置 GitHub Actions
- [ ] 配置多平台构建矩阵
- [ ] 配置自动发布 (tag 触发)
- [ ] 配置构建缓存
- [ ] 测试 CI 流程

**验收标准**:
- [ ] 推送 tag 后自动构建三平台
- [ ] 构建产物自动上传到 GitHub Releases

---

## 里程碑时间线

```
Week 1-2:  Phase 1 - Proto + gRPC 骨架
Week 3-4:  Phase 2 - 完善 gRPC 服务
Week 5-6:  Phase 3 - Tauri 桌面端
Week 7:    Phase 4 - 打包发布

总计: 7 周 (约 1.5 个月)
```

## 风险与应对

| 风险 | 影响 | 应对 |
|------|------|------|
| 交叉编译 CGO (SQLite) 困难 | 阻塞多平台构建 | 使用 mattn/go-sqlite3，配置 mingw 交叉编译器 |
| Tauri + Go sidecar 通信不稳定 | 影响用户体验 | 添加健康检查、自动重启、错误重试 |
| gRPC-Web 浏览器兼容性 | 部分浏览器不可用 | 使用 improbable-eng/grpc-web 代理 |
| Protobuf 消息与 Go 结构体不一致 | 转换函数多 | 统一命名规范，自动生成转换代码 |
| 前端迁移工作量大 | 延期 | 逐页面迁移，保留 REST 作为降级方案 |
