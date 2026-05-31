# PrismProxy 桌面端跨平台方案

## 架构总览

```
┌─────────────────────────────────────────────────────────┐
│                    桌面壳 (Tauri)                         │
│  ┌───────────────────────────────────────────────────┐  │
│  │              前端 UI (React + TS)                  │  │
│  │  ┌─────────┐ ┌─────────┐ ┌──────┐ ┌──────────┐  │  │
│  │  │ 流量页面 │ │ 规则页面 │ │ AI   │ │ 设置页面  │  │  │
│  │  └────┬────┘ └────┬────┘ └──┬───┘ └────┬─────┘  │  │
│  │       └───────────┴─────────┴──────────┘         │  │
│  │                      │ gRPC-Web                   │  │
│  └──────────────────────┼────────────────────────────┘  │
│                         │                                │
│  ┌──────────────────────┼────────────────────────────┐  │
│  │              Go 后端 (嵌入进程)                     │  │
│  │  ┌──────────┐ ┌──────────┐ ┌───────────────────┐ │  │
│  │  │ gRPC     │ │ 代理引擎  │ │ AI / 规则 / 断点   │ │  │
│  │  │ Server   │ │ MITM     │ │ 重写 / 集合 / ...  │ │  │
│  │  └──────────┘ └──────────┘ └───────────────────┘ │  │
│  │  ┌──────────┐ ┌──────────┐                       │  │
│  │  │ SQLite   │ │ WebSocket│  (实时推送保留 WS)     │  │
│  │  └──────────┘ └──────────┘                       │  │
│  └───────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

## 技术选型

| 层次 | 技术 | 理由 |
|------|------|------|
| 桌面壳 | Tauri 2.x | 轻量(~5MB)、跨平台(Win/Mac/Linux)、Rust 安全 |
| 前端 | React 19 + TypeScript + TailwindCSS | 已有代码，无需迁移 |
| 通信 | gRPC-Web (前端) + gRPC (Go 原生) | 类型安全、高性能、双向流 |
| 后端 | Go (嵌入 Tauri sidecar) | 已有代码，直接复用 |
| 数据库 | SQLite | 已有，嵌入式无需额外服务 |
| 实时推送 | WebSocket (保留) | gRPC 流式可替代但 WS 更简单 |

## 为什么选 gRPC

| 对比项 | REST API | gRPC |
|--------|----------|------|
| 类型安全 | 弱 (JSON) | 强 (Protobuf) |
| 性能 | JSON 序列化 | 二进制序列化，快 5-10x |
| 流式 | 需要 WebSocket | 原生支持双向流 |
| 代码生成 | 手写 | 自动生成客户端/服务端 |
| 浏览器兼容 | 原生 | 需要 gRPC-Web 代理 |
| 文档 | 需要 Swagger | .proto 即文档 |

## 为什么选 Tauri 而不是 Electron

| 对比项 | Electron | Tauri |
|--------|----------|-------|
| 包体积 | ~150MB | ~5MB |
| 内存占用 | ~200MB | ~30MB |
| 后端语言 | Node.js | Rust (可嵌入 Go sidecar) |
| 安全性 | 一般 | Rust 内存安全 |
| 跨平台 | Win/Mac/Linux | Win/Mac/Linux/Android/iOS |
| WebView | Chromium (自带) | 系统 WebView |

## 目录结构

```
prismProxy/
├── proto/                      # Protobuf 定义
│   ├── traffic.proto
│   ├── rules.proto
│   ├── breakpoints.proto
│   ├── rewrites.proto
│   ├── collections.proto
│   ├── environments.proto
│   ├── ai.proto
│   ├── codegen.proto
│   └── system.proto
├── cmd/
│   ├── proxy/main.go           # CLI 模式 (纯后端)
│   └── server/main.go          # gRPC 服务器
├── internal/
│   ├── grpc/                   # gRPC 服务实现
│   │   ├── server.go
│   │   ├── traffic_service.go
│   │   ├── rules_service.go
│   │   ├── breakpoints_service.go
│   │   ├── rewrites_service.go
│   │   ├── collections_service.go
│   │   ├── environments_service.go
│   │   ├── ai_service.go
│   │   ├── codegen_service.go
│   │   └── system_service.go
│   ├── proxy/                  # 代理引擎 (已有)
│   ├── traffic/                # 流量管理 (已有)
│   ├── rules/                  # 规则引擎 (已有)
│   ├── debugger/               # 断点调试 (已有)
│   ├── rewrite/                # 请求重写 (已有)
│   ├── collection/             # API 集合 (已有)
│   ├── environment/            # 环境变量 (已有)
│   ├── codegen/                # 代码生成 (已有)
│   ├── websocket/              # WebSocket (已有)
│   ├── ai/                     # AI 服务 (已有)
│   └── storage/                # 存储层 (已有)
├── desktop/                    # Tauri 桌面端
│   ├── src-tauri/
│   │   ├── Cargo.toml
│   │   ├── tauri.conf.json
│   │   ├── src/
│   │   │   └── main.rs         # Tauri 入口，启动 Go sidecar
│   │   └── icons/
│   ├── src/                    # 前端代码 (复用 web/)
│   │   ├── main.tsx
│   │   ├── App.tsx
│   │   ├── proto/              # gRPC-Web 生成的客户端
│   │   │   ├── traffic_pb.ts
│   │   │   ├── traffic_pb_service.ts
│   │   │   └── ...
│   │   ├── services/
│   │   │   ├── grpc.ts         # gRPC-Web 连接管理
│   │   │   ├── traffic.ts
│   │   │   └── ...
│   │   ├── stores/
│   │   ├── hooks/
│   │   ├── components/
│   │   └── pages/
│   ├── package.json
│   ├── vite.config.ts
│   └── tsconfig.json
├── web/                        # Web 版本 (保留，用于浏览器访问)
└── go.mod
```

## Protobuf 定义

### traffic.proto
```protobuf
syntax = "proto3";
package prismproxy;
option go_package = "prismproxy/proto";

service TrafficService {
  // 一元 RPC
  rpc GetTraffic(GetTrafficRequest) returns (TrafficEntry);
  rpc ListTraffic(ListTrafficRequest) returns (ListTrafficResponse);
  rpc DeleteTraffic(DeleteTrafficRequest) returns (Empty);
  rpc ClearTraffic(Empty) returns (Empty);
  rpc GetTrafficStats(Empty) returns (TrafficStats);
  rpc UpdateBookmark(UpdateBookmarkRequest) returns (Empty);
  rpc UpdateNotes(UpdateNotesRequest) returns (Empty);

  // 服务端流 - 实时推送新流量
  rpc WatchTraffic(WatchTrafficRequest) returns (stream TrafficEvent);
}

message TrafficEntry {
  int64 id = 1;
  string method = 2;
  string url = 3;
  string host = 4;
  string path = 5;
  int32 status_code = 6;
  string content_type = 7;
  int64 duration_ms = 8;
  int64 request_size = 9;
  int64 response_size = 10;
  string server_ip = 11;
  string timestamp = 12;
  bool bookmarked = 13;
  string notes = 14;
  string color = 15;
  repeated string tags = 16;
  RequestData request = 17;
  ResponseData response = 18;
}

message RequestData {
  map<string, string> headers = 1;
  bytes body = 2;
  string content_type = 3;
  int64 body_size = 4;
  bytes raw = 5;
}

message ResponseData {
  int32 status_code = 1;
  string status_text = 2;
  map<string, string> headers = 3;
  bytes body = 4;
  string content_type = 5;
  int64 body_size = 6;
  bytes raw = 7;
}

message ListTrafficRequest {
  int32 limit = 1;
  int32 offset = 2;
  repeated string methods = 3;
  repeated string hosts = 4;
  string path_pattern = 5;
  repeated int32 status_codes = 6;
  int64 min_duration = 7;
  int64 max_duration = 8;
  string search = 9;
  optional bool bookmarked = 10;
}

message ListTrafficResponse {
  repeated TrafficEntry entries = 1;
  int64 total = 2;
}

message TrafficEvent {
  enum Type {
    CREATED = 0;
    UPDATED = 1;
    DELETED = 2;
    CLEARED = 3;
  }
  Type type = 1;
  TrafficEntry entry = 2;
  int64 id = 3;
}

message TrafficStats {
  int64 total_requests = 1;
  int64 success_count = 2;
  int64 error_count = 3;
  double avg_duration_ms = 4;
  int64 max_duration_ms = 5;
  int64 min_duration_ms = 6;
}

message GetTrafficRequest { int64 id = 1; }
message DeleteTrafficRequest { int64 id = 1; }
message UpdateBookmarkRequest { int64 id = 1; bool bookmarked = 2; }
message UpdateNotesRequest { int64 id = 1; string notes = 2; }
message WatchTrafficRequest { repeated string event_types = 1; }
message Empty {}
```

### rules.proto
```protobuf
syntax = "proto3";
package prismproxy;
option go_package = "prismproxy/proto";

service RulesService {
  rpc ListRules(Empty) returns (ListRulesResponse);
  rpc GetRule(GetRuleRequest) returns (Rule);
  rpc CreateRule(Rule) returns (Rule);
  rpc UpdateRule(Rule) returns (Rule);
  rpc DeleteRule(DeleteRuleRequest) returns (Empty);
  rpc ToggleRule(ToggleRuleRequest) returns (ToggleRuleResponse);
  rpc ReorderRules(ReorderRulesRequest) returns (Empty);
  rpc ImportRules(ImportRulesRequest) returns (ImportRulesResponse);
  rpc ExportRules(Empty) returns (ExportRulesResponse);
}

message Rule {
  string id = 1;
  string name = 2;
  bool enabled = 3;
  int32 priority = 4;
  RuleMatch match = 5;
  RuleAction action = 6;
  string created_at = 7;
  string updated_at = 8;
  int32 hit_count = 9;
}

message RuleMatch {
  string url_pattern = 1;
  string host_pattern = 2;
  repeated string methods = 3;
  repeated string content_types = 4;
  string header_name = 5;
  string header_value = 6;
  string header_match_type = 7;
}

message RuleAction {
  string type = 1;
  string local_path = 2;
  string remote_url = 3;
  ModifySpec modify = 4;
  BlockSpec block = 5;
  int32 delay_ms = 6;
  string mock_body = 7;
}

message ModifySpec {
  map<string, string> add_headers = 1;
  repeated string remove_headers = 2;
  map<string, string> set_headers = 3;
  string body_replace = 4;
  string url_replace = 5;
}

message BlockSpec {
  int32 status_code = 1;
  map<string, string> headers = 2;
  string body = 3;
}

message ListRulesResponse { repeated Rule rules = 1; }
message GetRuleRequest { string id = 1; }
message DeleteRuleRequest { string id = 1; }
message ToggleRuleRequest { string id = 1; }
message ToggleRuleResponse { string id = 1; bool enabled = 2; }
message ReorderRulesRequest { repeated string ids = 1; }
message ImportRulesRequest { bytes data = 1; string format = 2; }
message ImportRulesResponse { int32 imported = 1; int32 skipped = 2; }
message ExportRulesResponse { bytes data = 1; string format = 2; }
```

### breakpoints.proto
```protobuf
syntax = "proto3";
package prismproxy;
option go_package = "prismproxy/proto";

service BreakpointsService {
  rpc ListBreakpoints(Empty) returns (ListBreakpointsResponse);
  rpc CreateBreakpoint(Breakpoint) returns (Breakpoint);
  rpc UpdateBreakpoint(Breakpoint) returns (Breakpoint);
  rpc DeleteBreakpoint(DeleteBreakpointRequest) returns (Empty);
  rpc ToggleBreakpoint(ToggleBreakpointRequest) returns (ToggleBreakpointResponse);
  rpc ListSessions(Empty) returns (ListSessionsResponse);
  rpc ReleaseSession(ReleaseSessionRequest) returns (Empty);
  rpc ModifySession(ModifySessionRequest) returns (Empty);
  rpc DropSession(DropSessionRequest) returns (Empty);

  // 服务端流 - 断点命中通知
  rpc WatchBreakpoints(Empty) returns (stream BreakpointEvent);
}

message Breakpoint {
  string id = 1;
  bool enabled = 2;
  string phase = 3;
  RuleMatch match = 4;
  string action_type = 5;
  ModifySpec auto_modify = 6;
  int32 hit_count = 7;
}

message BreakpointSession {
  string id = 1;
  string breakpoint_id = 2;
  int64 transaction_id = 3;
  string phase = 4;
  string status = 5;
  TrafficEntry original = 6;
  TrafficEntry modified = 7;
  string created_at = 8;
}

message BreakpointEvent {
  string type = 1;
  BreakpointSession session = 2;
}

message ListBreakpointsResponse { repeated Breakpoint breakpoints = 1; }
message DeleteBreakpointRequest { string id = 1; }
message ToggleBreakpointRequest { string id = 1; }
message ToggleBreakpointResponse { string id = 1; bool enabled = 2; }
message ListSessionsResponse { repeated BreakpointSession sessions = 1; }
message ReleaseSessionRequest { string id = 1; }
message ModifySessionRequest { string id = 1; TrafficEntry modified = 2; }
message DropSessionRequest { string id = 1; }
```

### ai.proto
```protobuf
syntax = "proto3";
package prismproxy;
option go_package = "prismproxy/proto";

service AIService {
  rpc Chat(ChatRequest) returns (stream ChatChunk);
  rpc AnalyzeTraffic(AnalyzeRequest) returns (AnalysisResult);
  rpc SecurityCheck(SecurityCheckRequest) returns (SecurityReport);
  rpc GenerateTests(TestGenRequest) returns (TestGenResult);
  rpc GetProviders(Empty) returns (ProvidersResponse);
}

message ChatMessage {
  string role = 1;
  string content = 2;
}

message ChatRequest {
  repeated ChatMessage messages = 1;
  string provider = 2;
  string model = 3;
  bool stream = 4;
}

message ChatChunk {
  string content = 1;
  bool done = 2;
  string provider = 3;
}

message AnalyzeRequest {
  repeated int64 traffic_ids = 1;
}

message AnalysisResult {
  string summary = 1;
  repeated Issue issues = 2;
  repeated Suggestion suggestions = 3;
}

message Issue {
  string severity = 1;
  string type = 2;
  string title = 3;
  string detail = 4;
}

message Suggestion {
  string category = 1;
  string title = 2;
  string detail = 3;
}

message SecurityCheckRequest {
  int64 traffic_id = 1;
}

message SecurityReport {
  string risk_level = 1;
  repeated SecurityFinding findings = 2;
  string summary = 3;
}

message SecurityFinding {
  string severity = 1;
  string category = 2;
  string title = 3;
  string description = 4;
  string remediation = 5;
}

message TestGenRequest {
  repeated int64 traffic_ids = 1;
  string framework = 2;
}

message TestGenResult {
  repeated TestCase cases = 1;
}

message TestCase {
  string name = 1;
  string description = 2;
  string method = 3;
  string url = 4;
  string code = 5;
}

message ProvidersResponse {
  repeated string providers = 1;
  string default_provider = 2;
}
```

### system.proto
```protobuf
syntax = "proto3";
package prismproxy;
option go_package = "prismproxy/proto";

service SystemService {
  rpc GetStatus(Empty) returns (SystemStatus);
  rpc GetSettings(Empty) returns (Settings);
  rpc UpdateSettings(Settings) returns (Settings);
  rpc DownloadCert(Empty) returns (CertData);
}

message SystemStatus {
  string version = 1;
  string proxy_addr = 2;
  string api_addr = 3;
  bool proxy_running = 4;
  int64 uptime_seconds = 5;
  int64 total_traffic = 6;
  int32 ws_clients = 7;
}

message Settings {
  ProxySettings proxy = 1;
  AISettings ai = 2;
}

message ProxySettings {
  int32 port = 1;
  bool mitm_enabled = 2;
  bool capture_websocket = 3;
  int64 max_body_size = 4;
  repeated string exclude_domains = 5;
}

message AISettings {
  string default_provider = 1;
  string openai_api_key = 2;
  string openai_base_url = 3;
  string openai_model = 4;
  string claude_api_key = 5;
  string claude_base_url = 6;
  string claude_model = 7;
  string ollama_base_url = 8;
  string ollama_model = 9;
}

message CertData {
  bytes cert = 1;
  bytes key = 2;
}
```

## Tauri 配置

### tauri.conf.json
```json
{
  "productName": "PrismProxy",
  "version": "1.0.0",
  "identifier": "com.prismproxy.app",
  "build": {
    "frontendDist": "../dist",
    "devUrl": "http://localhost:3000",
    "beforeDevCommand": "npm run dev",
    "beforeBuildCommand": "npm run build"
  },
  "app": {
    "windows": [
      {
        "title": "PrismProxy",
        "width": 1400,
        "height": 900,
        "minWidth": 1000,
        "minHeight": 600,
        "resizable": true,
        "decorations": true,
        "transparent": false
      }
    ],
    "security": {
      "csp": null
    }
  },
  "bundle": {
    "active": true,
    "targets": "all",
    "icon": [
      "icons/32x32.png",
      "icons/128x128.png",
      "icons/icon.icns",
      "icons/icon.ico"
    ],
    "externalBin": ["bin/prismproxy-server"]
  }
}
```

### src-tauri/src/main.rs
```rust
use tauri::{Manager, WindowEvent};
use tauri_plugin_shell::ShellExt;

fn main() {
    tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .setup(|app| {
            // 启动 Go sidecar
            let sidecar_command = app.shell().sidecar("prismproxy-server").unwrap();
            let (mut _rx, _child) = sidecar_command
                .args(&["--port", "9090", "--proxy-port", "8080"])
                .spawn()
                .expect("Failed to spawn sidecar");

            // 存储子进程句柄以便关闭时清理
            app.manage(_child);

            Ok(())
        })
        .on_window_event(|window, event| {
            if let WindowEvent::CloseRequested { .. } = event {
                // 关闭 Go sidecar
                let child = window.state::<tauri::api::process::CommandChild>();
                child.kill().expect("Failed to kill sidecar");
            }
        })
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
```

## 前端 gRPC-Web 客户端

### services/grpc.ts
```typescript
import { GrpcWebFetchTransport } from "@protobuf-ts/grpcweb-transport";
import { TrafficServiceClient } from "../proto/traffic_pb_service";
import { RulesServiceClient } from "../proto/rules_pb_service";
import { BreakpointsServiceClient } from "../proto/breakpoints_pb_service";
import { AIServiceClient } from "../proto/ai_pb_service";
import { SystemServiceClient } from "../proto/system_pb_service";

const transport = new GrpcWebFetchTransport({
  baseUrl: "http://localhost:9090",
});

export const trafficClient = new TrafficServiceClient(transport);
export const rulesClient = new RulesServiceClient(transport);
export const breakpointsClient = new BreakpointsServiceClient(transport);
export const aiClient = new AIServiceClient(transport);
export const systemClient = new SystemServiceClient(transport);
```

### hooks/useTrafficStream.ts
```typescript
import { useEffect, useRef } from "react";
import { trafficClient } from "../services/grpc";
import { useTrafficStore } from "../stores/trafficStore";

export function useTrafficStream() {
  const abortRef = useRef<AbortController | null>(null);
  const { addTraffic, removeTraffic, clearTraffic } = useTrafficStore();

  useEffect(() => {
    const abort = new AbortController();
    abortRef.current = abort;

    const stream = trafficClient.watchTraffic(
      { eventTypes: ["CREATED", "DELETED", "CLEARED"] },
      { abort: abort.signal }
    );

    stream.responses.onMessage((event) => {
      switch (event.type) {
        case "CREATED":
          addTraffic(event.entry!);
          break;
        case "DELETED":
          removeTraffic(event.id);
          break;
        case "CLEARED":
          clearTraffic();
          break;
      }
    });

    return () => abort.abort();
  }, []);
}
```

## 实施计划

### Phase 1: Protobuf + gRPC 基础 (1-2周)
- [ ] 定义所有 .proto 文件
- [ ] 生成 Go gRPC 服务端代码
- [ ] 生成 TypeScript gRPC-Web 客户端代码
- [ ] 实现 gRPC 服务器 (internal/grpc/)
- [ ] 实现 TrafficService (第一个完整服务)

### Phase 2: 完善 gRPC 服务 (1-2周)
- [ ] 实现 RulesService
- [ ] 实现 BreakpointsService
- [ ] 实现 RewritesService
- [ ] 实现 CollectionsService
- [ ] 实现 EnvironmentsService
- [ ] 实现 AIService (含流式 Chat)
- [ ] 实现 SystemService

### Phase 3: Tauri 桌面端 (1-2周)
- [ ] 初始化 Tauri 项目
- [ ] 配置 Go sidecar 打包
- [ ] 实现 sidecar 启动/停止管理
- [ ] 前端从 REST 迁移到 gRPC-Web
- [ ] 流式数据用 gRPC server-stream 替代 WebSocket

### Phase 4: 打包发布 (1周)
- [ ] Windows 打包 (MSI/EXE)
- [ ] macOS 打包 (DMG)
- [ ] Linux 打包 (DEB/AppImage)
- [ ] 自动更新 (Tauri Updater)
- [ ] CA 证书自动安装引导

## 构建命令

```bash
# 1. 生成 Protobuf 代码
./scripts/gen_proto.sh

# 2. 构建 Go sidecar
GOOS=linux GOARCH=arm64 go build -o desktop/src-tauri/bin/prismproxy-server ./cmd/server/
GOOS=darwin GOARCH=arm64 go build -o desktop/src-tauri/bin/prismproxy-server-macos ./cmd/server/
GOOS=windows GOARCH=amd64 go build -o desktop/src-tauri/bin/prismproxy-server.exe ./cmd/server/

# 3. 构建 Tauri 桌面端
cd desktop && npm run tauri build
```

## 依赖

### Go
```
google.golang.org/grpc
google.golang.org/protobuf
github.com/grpc-ecosystem/grpc-gateway/v2  (可选，REST 兼容)
```

### 前端
```
@protobuf-ts/plugin          # Protobuf TS 代码生成
@protobuf-ts/grpcweb-transport  # gRPC-Web 传输
@protobuf-ts/runtime         # Protobuf 运行时
@tauri-apps/api              # Tauri API
@tauri-apps/plugin-shell     # Sidecar 管理
```

### Tauri (Rust)
```
tauri = { version = "2", features = ["shell"] }
tauri-plugin-shell = "2"
```
