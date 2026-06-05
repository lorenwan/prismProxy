# PrismProxy 架构迁移设计文档

## Context

**问题**：当前架构中，前端通过gRPC-Web直接与Go后端通信，存在以下问题：
- gRPC-Web迁移不完整，前端仍使用REST API作为过渡
- 缺少TypeScript proto生成管道
- 浏览器环境限制了某些功能（如证书管理、系统代理设置）

**目标**：迁移到"前端 → Rust → Go"架构，Rust层作为Tauri IPC到gRPC的桥梁。

**预期收益**：
- 统一通信协议，消除REST/gRPC-Web混用
- 利用Tauri原生能力，支持系统级操作
- 简化前端代码，类型安全的IPC调用
- 更好的错误处理和进程管理

---

## 架构设计

### 整体架构图

```
┌─────────────────────────────────────────────────────────────┐
│                      React Frontend                         │
│  ┌─────────────────────────────────────────────────────────┐│
│  │  Services Layer (services/*.ts)                         ││
│  │  - traffic.ts, rules.ts, ai.ts, etc.                   ││
│  │  - 调用 invoke('service:method', args)                  ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
                              │
                              │ Tauri IPC (invoke)
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      Rust Layer (Tauri)                     │
│  ┌─────────────────────────────────────────────────────────┐│
│  │  IPC Handlers (src/commands/*.rs)                       ││
│  │  - traffic.rs, rules.rs, ai.rs, etc.                   ││
│  │  - 接收IPC调用，转换为gRPC请求                           ││
│  └─────────────────────────────────────────────────────────┘│
│  ┌─────────────────────────────────────────────────────────┐│
│  │  gRPC Client (src/grpc_client.rs)                       ││
│  │  - tonic客户端，连接Go后端                               ││
│  │  - 管理连接池、重连、超时                                ││
│  └─────────────────────────────────────────────────────────┘│
│  ┌─────────────────────────────────────────────────────────┐│
│  │  Sidecar Manager (src/sidecar.rs)                       ││
│  │  - 启动/停止Go进程                                      ││
│  │  - 健康检查、自动重启                                   ││
│  └─────────────────────────────────────────────────────────┘│
│  ┌─────────────────────────────────────────────────────────┐│
│  │  Config Manager (src/config.rs)                         ││
│  │  - 窗口状态、主题、本地设置                              ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
                              │
                              │ gRPC (tonic)
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      Go Backend                             │
│  ┌─────────────────────────────────────────────────────────┐│
│  │  gRPC Server (internal/grpc/)                           ││
│  │  - 14个服务实现                                         ││
│  │  - TrafficService, RulesService, AIService, etc.        ││
│  └─────────────────────────────────────────────────────────┘│
│  ┌─────────────────────────────────────────────────────────┐│
│  │  Business Logic (internal/)                             ││
│  │  - traffic/, rules/, ai/, etc.                          ││
│  └─────────────────────────────────────────────────────────┘│
│  ┌─────────────────────────────────────────────────────────┐│
│  │  Storage (internal/storage/)                            ││
│  │  - SQLite 数据库                                        ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

### 数据流

#### 1. 普通请求流（Unary RPC）

```
前端                    Rust                    Go
  │                       │                       │
  │ invoke('traffic:list', {page: 1})             │
  │──────────────────────>│                       │
  │                       │                       │
  │                       │ grpc.TrafficService.List(TrafficListRequest)
  │                       │──────────────────────>│
  │                       │                       │
  │                       │ TrafficListResponse   │
  │                       │<──────────────────────│
  │                       │                       │
  │ TrafficListResponse   │                       │
  │<──────────────────────│                       │
```

#### 2. 流式请求流（Server Streaming RPC）

```
前端                    Rust                    Go
  │                       │                       │
  │ invoke('traffic:subscribe')                   │
  │──────────────────────>│                       │
  │                       │                       │
  │                       │ grpc.TrafficService.Subscribe(Empty)
  │                       │──────────────────────>│
  │                       │                       │
  │                       │ stream TrafficEvent   │
  │                       │<──────────────────────│
  │                       │                       │
  │ emit('traffic:event', data)                   │
  │<──────────────────────│                       │
  │                       │                       │
  │ listen('traffic:event')                       │
  │ (前端持续监听)         │                       │
```

---

## 组件设计

### 1. Rust gRPC客户端 (`src/grpc_client.rs`)

**职责**：管理与Go后端的gRPC连接，提供类型安全的调用接口。

**关键设计**：
```rust
pub struct GrpcClient {
    traffic_client: TrafficServiceClient<Channel>,
    rules_client: RulesServiceClient<Channel>,
    ai_client: AIServiceClient<Channel>,
    // ... 其他14个服务的客户端
}

impl GrpcClient {
    pub async fn new(addr: &str) -> Result<Self, tonic::Status> {
        // 创建连接，配置超时、重连策略
    }
    
    // 每个服务对应一组方法
    pub async fn list_traffic(&self, req: TrafficListRequest) -> Result<TrafficListResponse, tonic::Status> {
        self.traffic_client.clone().list(req).await.map(|r| r.into_inner())
    }
}
```

**依赖**：
- `tonic` - gRPC客户端
- `tonic-build` - 编译时生成客户端代码
- `prost` - Protocol Buffers消息类型

### 2. IPC Handlers (`src/commands/*.rs`)

**职责**：接收前端invoke调用，转换为gRPC请求，返回响应。

**设计模式**：每个gRPC服务对应一个commands模块。

```rust
// src/commands/traffic.rs
use tauri::command;
use crate::grpc_client::GrpcClient;
use crate::state::AppState;

#[command]
pub async fn list_traffic(
    state: State<'_, AppState>,
    page: Option<i32>,
    page_size: Option<i32>,
) -> Result<TrafficListResponse, String> {
    let client = state.grpc_client.lock().await;
    let req = TrafficListRequest {
        pagination: Some(Pagination {
            page: page.unwrap_or(1),
            page_size: page_size.unwrap_or(20),
            ..Default::default()
        }),
        ..Default::default()
    };
    
    client.list_traffic(req).await.map_err(|e| e.to_string())
}
```

**注册方式**（`src/main.rs`）：
```rust
fn main() {
    tauri::Builder::default()
        .manage(AppState::default())
        .invoke_handler(tauri::generate_handler![
            commands::traffic::list_traffic,
            commands::traffic::get_traffic,
            commands::traffic::subscribe_traffic,
            commands::rules::list_rules,
            // ... 其他命令
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
```

### 3. 流式RPC处理

**策略**：gRPC流转换为Tauri事件。

```rust
// src/commands/traffic.rs
#[command]
pub async fn subscribe_traffic(
    app: AppHandle,
    state: State<'_, AppState>,
) -> Result<(), String> {
    let mut client = state.grpc_client.lock().await;
    let mut stream = client.subscribe_traffic().await
        .map_err(|e| e.to_string())?
        .into_inner();
    
    // 在后台任务中处理流
    tokio::spawn(async move {
        while let Some(event) = stream.message().await.unwrap_or(None) {
            // 将gRPC事件转换为Tauri事件
            app.emit("traffic:event", event).unwrap();
        }
    });
    
    Ok(())
}
```

**前端监听**：
```typescript
import { listen } from '@tauri-apps/api/event';

const unlisten = await listen('traffic:event', (event) => {
  console.log('New traffic event:', event.payload);
});
```

### 4. Sidecar Manager (`src/sidecar.rs`)

**职责**：管理Go后端进程的生命周期。

**功能**：
- 启动sidecar进程，传递端口参数
- 定期健康检查（HTTP /health端点）
- 进程崩溃时自动重启
- 应用退出时优雅关闭

```rust
pub struct SidecarManager {
    child: Option<CommandChild>,
    health_check_handle: Option<JoinHandle<()>>,
}

impl SidecarManager {
    pub async fn start(&mut self, app: AppHandle) -> Result<(), String> {
        // 启动sidecar
        let (mut rx, child) = app.shell()
            .sidecar("prismproxy-server")
            .args(["--port", "9090", "--proxy-port", "8080"])
            .spawn()
            .map_err(|e| e.to_string())?;
        
        self.child = Some(child);
        
        // 启动健康检查
        self.start_health_check(app);
        
        Ok(())
    }
    
    fn start_health_check(&mut self, app: AppHandle) {
        let handle = tokio::spawn(async move {
            loop {
                tokio::time::sleep(Duration::from_secs(5)).await;
                if let Err(_) = reqwest::get("http://localhost:8080/health").await {
                    // 健康检查失败，尝试重启
                    app.emit("sidecar:health_check_failed", ()).unwrap();
                }
            }
        });
        self.health_check_handle = Some(handle);
    }
}
```

### 5. Config Manager (`src/config.rs`)

**职责**：管理桌面端特有的配置。

**存储内容**：
- 窗口位置和大小
- 主题设置（暗色/亮色）
- 语言设置
- 最近打开的项目

**存储方式**：JSON文件，位于 `~/.prismproxy/config.json`

```rust
#[derive(Serialize, Deserialize)]
pub struct AppConfig {
    pub window: WindowConfig,
    pub theme: Theme,
    pub language: String,
}

#[derive(Serialize, Deserialize)]
pub struct WindowConfig {
    pub x: i32,
    pub y: i32,
    pub width: u32,
    pub height: u32,
    pub maximized: bool,
}
```

---

## Proto代码生成

### 构建流程

在 `desktop/src-tauri/build.rs` 中添加proto编译：

```rust
fn main() {
    // 编译proto文件，生成Rust客户端代码
    tonic_build::configure()
        .build_server(false)  // 只生成客户端，不生成服务端
        .build_client(true)
        .compile(
            &[
                "../../proto/traffic.proto",
                "../../proto/rules.proto",
                "../../proto/ai.proto",
                // ... 其他proto文件
            ],
            &["../../proto"],
        )
        .unwrap();
    
    tauri_build::run()
}
```

### 生成的代码结构

```
desktop/src-tauri/src/gen/
  └── prismproxy.rs  // 所有生成的客户端代码
```

### 依赖配置 (`Cargo.toml`)

```toml
[dependencies]
tonic = "0.12"
prost = "0.13"
tokio = { version = "1", features = ["full"] }
serde = { version = "1", features = ["derive"] }
serde_json = "1"
reqwest = { version = "0.12", features = ["json"] }

[build-dependencies]
tonic-build = "0.12"
```

---

## 前端迁移

### Services层改造

**Before**（当前）：
```typescript
// services/traffic.ts
import api from './api';

export async function listTraffic(page: number = 1) {
  const response = await api.get('/traffic/list', { params: { page } });
  return response.data;
}
```

**After**（迁移后）：
```typescript
// services/traffic.ts
import { invoke } from '@tauri-apps/api/core';

export async function listTraffic(page: number = 1): Promise<TrafficListResponse> {
  return invoke('list_traffic', { page, pageSize: 20 });
}
```

### 流式数据处理

```typescript
// services/traffic.ts
import { listen } from '@tauri-apps/api/event';

export async function subscribeTraffic(callback: (event: TrafficEvent) => void) {
  // 启动订阅
  await invoke('subscribe_traffic');
  
  // 监听事件
  const unlisten = await listen('traffic:event', (event) => {
    callback(event.payload as TrafficEvent);
  });
  
  return unlisten;
}
```

---

## 错误处理

### 错误类型定义

```rust
// src/error.rs
#[derive(Debug, thiserror::Error)]
pub enum AppError {
    #[error("gRPC error: {0}")]
    Grpc(#[from] tonic::Status),
    
    #[error("Connection error: {0}")]
    Connection(String),
    
    #[error("Sidecar error: {0}")]
    Sidecar(String),
    
    #[error("Config error: {0}")]
    Config(String),
}

// 实现Serialize，以便通过IPC返回给前端
impl Serialize for AppError {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        serializer.serialize_str(&self.to_string())
    }
}
```

### 错误传递流程

```
Go后端返回错误
    ↓ tonic::Status
Rust层捕获错误
    ↓ AppError
转换为字符串
    ↓ String
通过IPC返回前端
    ↓ invoke() reject
前端捕获异常
```

---

## 迁移计划

### 阶段1：Rust基础设施（2天）

1. **配置Cargo.toml**：添加tonic、prost等依赖
2. **创建build.rs**：配置proto编译
3. **实现gRPC客户端**：`src/grpc_client.rs`
4. **实现Sidecar管理器**：`src/sidecar.rs`
5. **实现配置管理器**：`src/config.rs`

### 阶段2：IPC Handlers（3天）

按服务逐个实现IPC handler：

1. **SystemService**（优先，用于验证架构）
2. **TrafficService**（核心功能）
3. **RulesService**
4. **AIService**（包含流式处理）
5. **BreakpointsService**（包含流式处理）
6. 其他服务（Rewrites、Collections、Environments、CodeGen、Scripts、Diff、Perf、Cert、Search）

### 阶段3：前端迁移（2天）

1. **更新services层**：将所有REST调用改为invoke
2. **处理流式数据**：实现事件监听
3. **移除旧代码**：删除api.ts、grpc.ts等
4. **更新类型定义**：使用生成的TypeScript类型（可选）

### 阶段4：测试和优化（1天）

1. **单元测试**：Rust层的gRPC客户端测试
2. **集成测试**：端到端流程测试
3. **性能测试**：IPC调用延迟、流式数据吞吐量
4. **错误处理测试**：网络断开、sidecar崩溃等场景

---

## 验证方案

### 功能验证

1. **普通请求**：调用TrafficService.List，验证返回数据正确
2. **流式请求**：调用TrafficService.Subscribe，验证实时事件推送
3. **错误处理**：模拟Go后端崩溃，验证错误正确传递到前端
4. **Sidecar管理**：验证进程启动、健康检查、自动重启

### 性能验证

1. **IPC延迟**：单次invoke调用 < 5ms
2. **流式吞吐量**：每秒处理100+事件
3. **内存占用**：Rust层内存占用 < 50MB
4. **启动时间**：应用启动 < 3秒

### 兼容性验证

1. **macOS**：Apple Silicon + Intel
2. **Windows**：x86_64
3. **Linux**：x86_64 + aarch64

---

## 关键文件清单

### Rust层（新增/修改）

- `desktop/src-tauri/Cargo.toml` - 添加依赖
- `desktop/src-tauri/build.rs` - Proto编译配置
- `desktop/src-tauri/src/main.rs` - 应用入口，注册命令
- `desktop/src-tauri/src/grpc_client.rs` - gRPC客户端
- `desktop/src-tauri/src/sidecar.rs` - Sidecar管理
- `desktop/src-tauri/src/config.rs` - 配置管理
- `desktop/src-tauri/src/error.rs` - 错误处理
- `desktop/src-tauri/src/commands/*.rs` - IPC handlers（14个文件）

### 前端（修改）

- `desktop/src/services/traffic.ts` - 改用invoke
- `desktop/src/services/rules.ts` - 改用invoke
- `desktop/src/services/ai.ts` - 改用invoke + 事件监听
- `desktop/src/services/*.ts` - 其他服务
- `desktop/src/services/api.ts` - 删除
- `desktop/src/services/grpc.ts` - 删除

### Go后端（保持不变）

- `proto/` - Proto定义不变
- `internal/grpc/` - gRPC服务实现不变
- `cmd/server/` - 服务器入口不变

---

## 风险和缓解

### 风险1：tonic版本兼容性

**问题**：tonic版本与proto语法不兼容
**缓解**：使用tonic 0.12+，支持proto3所有特性

### 风险2：流式RPC性能

**问题**：大量流式事件可能导致IPC拥塞
**缓解**：实现事件批处理，减少IPC调用次数

### 风险3：Sidecar进程管理

**问题**：进程崩溃检测不及时
**缓解**：健康检查间隔设为5秒，崩溃后立即重启

### 风险4：跨平台兼容性

**问题**：不同平台的进程管理差异
**缓解**：使用Tauri的shell插件，屏蔽平台差异

---

## 总结

本设计文档描述了PrismProxy从"前端直连Go"到"前端→Rust→Go"的架构迁移方案。

**核心原则**：
- Rust层只做转发，业务逻辑留在Go后端
- 保持现有proto定义不变
- 一次性迁移所有服务
- 仅支持Tauri模式

**预期收益**：
- 统一通信协议
- 类型安全的IPC调用
- 更好的进程管理
- 支持系统级操作

**工作量估算**：8天（2+3+2+1）

**下一步**：确认设计文档后，开始编写实现计划。
