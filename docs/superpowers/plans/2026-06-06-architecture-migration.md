# PrismProxy 架构迁移实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将PrismProxy从"前端直连Go"迁移到"前端→Rust→Go"架构，Rust层作为Tauri IPC到gRPC的桥梁。

**Architecture:** 前端通过Tauri IPC调用Rust层，Rust层通过tonic gRPC客户端转发请求到Go后端。Rust层只做转发，业务逻辑留在Go后端。流式RPC通过Tauri事件推送。

**Tech Stack:** Rust (tonic, prost), Tauri 2.x, React, TypeScript, Go (gRPC)

---

## 文件结构

### Rust层（新增/修改）

```
desktop/src-tauri/
├── Cargo.toml                    # 添加tonic、prost等依赖
├── build.rs                      # Proto编译配置
└── src/
    ├── main.rs                   # 应用入口，注册命令，管理状态
    ├── lib.rs                    # 模块导出
    ├── error.rs                  # 错误类型定义
    ├── grpc_client.rs            # gRPC客户端封装
    ├── sidecar.rs                # Sidecar进程管理
    ├── config.rs                 # 本地配置管理
    ├── state.rs                  # 应用状态定义
    └── commands/
        ├── mod.rs                # 命令模块导出
        ├── system.rs             # SystemService IPC handlers
        ├── traffic.rs            # TrafficService IPC handlers
        ├── rules.rs              # RulesService IPC handlers
        ├── ai.rs                 # AIService IPC handlers
        ├── breakpoints.rs        # BreakpointsService IPC handlers
        ├── rewrites.rs           # RewritesService IPC handlers
        ├── collections.rs        # CollectionsService IPC handlers
        ├── environments.rs       # EnvironmentsService IPC handlers
        ├── codegen.rs            # CodeGenService IPC handlers
        ├── scripts.rs            # ScriptsService IPC handlers
        ├── diff.rs               # DiffService IPC handlers
        ├── perf.rs               # PerfService IPC handlers
        ├── cert.rs               # CertService IPC handlers
        └── search.rs             # SearchService IPC handlers
```

### 前端（修改）

```
desktop/src/services/
├── api.ts                        # 删除
├── grpc.ts                       # 删除
├── traffic.ts                    # 改用invoke
├── rules.ts                      # 改用invoke
├── ai.ts                         # 改用invoke + 事件监听
├── breakpoints.ts                # 改用invoke + 事件监听
├── rewrites.ts                   # 改用invoke
├── collections.ts                # 改用invoke
├── environments.ts               # 改用invoke
├── cert.ts                       # 改用invoke
├── scripts.ts                    # 改用invoke
├── search.ts                     # 改用invoke
├── diff.ts                       # 改用invoke
├── perf.ts                       # 改用invoke
└── proxy.ts                      # 改用invoke
```

### Go后端（保持不变）

```
proto/                            # Proto定义不变
internal/grpc/                    # gRPC服务实现不变
cmd/server/                       # 服务器入口不变
```

---

## 任务分解

### Task 1: 配置Rust项目依赖

**Files:**
- Modify: `desktop/src-tauri/Cargo.toml`
- Create: `desktop/src-tauri/build.rs`

- [ ] **Step 1: 更新Cargo.toml添加依赖**

```toml
[package]
name = "prismproxy-desktop"
version = "0.1.0"
edition = "2021"

[dependencies]
tauri = { version = "2", features = ["devtools"] }
tauri-plugin-shell = "2"
serde = { version = "1", features = ["derive"] }
serde_json = "1"
tonic = "0.12"
prost = "0.13"
tokio = { version = "1", features = ["full"] }
thiserror = "1"
reqwest = { version = "0.12", features = ["json"] }
dirs = "5"

[build-dependencies]
tauri-build = { version = "2", features = [] }
tonic-build = "0.12"
```

- [ ] **Step 2: 创建build.rs配置proto编译**

```rust
fn main() {
    // 编译proto文件，生成Rust客户端代码
    tonic_build::configure()
        .build_server(false)  // 只生成客户端，不生成服务端
        .build_client(true)
        .compile(
            &[
                "../../proto/common.proto",
                "../../proto/traffic.proto",
                "../../proto/rules.proto",
                "../../proto/breakpoints.proto",
                "../../proto/rewrites.proto",
                "../../proto/collections.proto",
                "../../proto/environments.proto",
                "../../proto/ai.proto",
                "../../proto/system.proto",
                "../../proto/codegen.proto",
                "../../proto/scripts.proto",
                "../../proto/diff.proto",
                "../../proto/perf.proto",
                "../../proto/cert.proto",
                "../../proto/search.proto",
            ],
            &["../../proto"],
        )
        .unwrap();

    tauri_build::run()
}
```

- [ ] **Step 3: 验证proto编译**

Run: `cd desktop/src-tauri && cargo build 2>&1 | head -50`
Expected: 编译成功，生成proto客户端代码

- [ ] **Step 4: 提交**

```bash
git add desktop/src-tauri/Cargo.toml desktop/src-tauri/build.rs
git commit -m "feat: 配置Rust项目依赖和proto编译"
```

---

### Task 2: 实现错误处理模块

**Files:**
- Create: `desktop/src-tauri/src/error.rs`

- [ ] **Step 1: 创建错误类型定义**

```rust
use serde::Serialize;
use std::fmt;

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

    #[error("IO error: {0}")]
    Io(#[from] std::io::Error),
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

// 为Tauri命令返回Result类型
pub type AppResult<T> = Result<T, AppError>;
```

- [ ] **Step 2: 验证编译**

Run: `cd desktop/src-tauri && cargo check`
Expected: 编译通过

- [ ] **Step 3: 提交**

```bash
git add desktop/src-tauri/src/error.rs
git commit -m "feat: 添加错误处理模块"
```

---

### Task 3: 实现应用状态管理

**Files:**
- Create: `desktop/src-tauri/src/state.rs`

- [ ] **Step 1: 创建应用状态定义**

```rust
use std::sync::Arc;
use tokio::sync::Mutex;
use crate::grpc_client::GrpcClient;
use crate::sidecar::SidecarManager;
use crate::config::ConfigManager;

#[derive(Clone)]
pub struct AppState {
    pub grpc_client: Arc<Mutex<Option<GrpcClient>>>,
    pub sidecar_manager: Arc<Mutex<SidecarManager>>,
    pub config_manager: Arc<Mutex<ConfigManager>>,
}

impl Default for AppState {
    fn default() -> Self {
        Self {
            grpc_client: Arc::new(Mutex::new(None)),
            sidecar_manager: Arc::new(Mutex::new(SidecarManager::new())),
            config_manager: Arc::new(Mutex::new(ConfigManager::new())),
        }
    }
}

impl AppState {
    pub async fn init_grpc_client(&self, addr: &str) -> Result<(), tonic::Status> {
        let client = GrpcClient::new(addr).await?;
        let mut grpc_client = self.grpc_client.lock().await;
        *grpc_client = Some(client);
        Ok(())
    }

    pub async fn get_grpc_client(&self) -> Result<GrpcClient, crate::error::AppError> {
        let grpc_client = self.grpc_client.lock().await;
        grpc_client.clone().ok_or_else(|| {
            crate::error::AppError::Connection("gRPC client not initialized".to_string())
        })
    }
}
```

- [ ] **Step 2: 验证编译**

Run: `cd desktop/src-tauri && cargo check`
Expected: 编译通过（可能有未定义的类型错误，后续修复）

- [ ] **Step 3: 提交**

```bash
git add desktop/src-tauri/src/state.rs
git commit -m "feat: 添加应用状态管理"
```

---

### Task 4: 实现gRPC客户端封装

**Files:**
- Create: `desktop/src-tauri/src/grpc_client.rs`

- [ ] **Step 1: 创建gRPC客户端封装**

```rust
use tonic::transport::Channel;
use tonic::Status;

// 导入生成的proto代码
include!(concat!(env!("OUT_DIR"), "/prismproxy.rs"));

#[derive(Clone)]
pub struct GrpcClient {
    pub traffic: traffic_service_client::TrafficServiceClient<Channel>,
    pub rules: rules_service_client::RulesServiceClient<Channel>,
    pub breakpoints: breakpoints_service_client::BreakpointsServiceClient<Channel>,
    pub rewrites: rewrites_service_client::RewritesServiceClient<Channel>,
    pub collections: collections_service_client::CollectionsServiceClient<Channel>,
    pub environments: environments_service_client::EnvironmentsServiceClient<Channel>,
    pub ai: ai_service_client::AIServiceClient<Channel>,
    pub system: system_service_client::SystemServiceClient<Channel>,
    pub codegen: codegen_service_client::CodeGenServiceClient<Channel>,
    pub scripts: scripts_service_client::ScriptsServiceClient<Channel>,
    pub diff: diff_service_client::DiffServiceClient<Channel>,
    pub perf: perf_service_client::PerfServiceClient<Channel>,
    pub cert: cert_service_client::CertServiceClient<Channel>,
    pub search: search_service_client::SearchServiceClient<Channel>,
}

impl GrpcClient {
    pub async fn new(addr: &str) -> Result<Self, Status> {
        let channel = Channel::from_shared(addr.to_string())
            .map_err(|e| Status::internal(format!("Failed to create channel: {}", e)))?
            .connect()
            .await
            .map_err(|e| Status::internal(format!("Failed to connect: {}", e)))?;

        Ok(Self {
            traffic: traffic_service_client::TrafficServiceClient::new(channel.clone()),
            rules: rules_service_client::RulesServiceClient::new(channel.clone()),
            breakpoints: breakpoints_service_client::BreakpointsServiceClient::new(channel.clone()),
            rewrites: rewrites_service_client::RewritesServiceClient::new(channel.clone()),
            collections: collections_service_client::CollectionsServiceClient::new(channel.clone()),
            environments: environments_service_client::EnvironmentsServiceClient::new(channel.clone()),
            ai: ai_service_client::AIServiceClient::new(channel.clone()),
            system: system_service_client::SystemServiceClient::new(channel.clone()),
            codegen: codegen_service_client::CodeGenServiceClient::new(channel.clone()),
            scripts: scripts_service_client::ScriptsServiceClient::new(channel.clone()),
            diff: diff_service_client::DiffServiceClient::new(channel.clone()),
            perf: perf_service_client::PerfServiceClient::new(channel.clone()),
            cert: cert_service_client::CertServiceClient::new(channel.clone()),
            search: search_service_client::SearchServiceClient::new(channel),
        })
    }
}
```

- [ ] **Step 2: 验证编译**

Run: `cd desktop/src-tauri && cargo check`
Expected: 编译通过（需要先运行Task 1的proto编译）

- [ ] **Step 3: 提交**

```bash
git add desktop/src-tauri/src/grpc_client.rs
git commit -m "feat: 实现gRPC客户端封装"
```

---

### Task 5: 实现Sidecar进程管理

**Files:**
- Create: `desktop/src-tauri/src/sidecar.rs`

- [ ] **Step 1: 创建Sidecar管理器**

```rust
use tauri::{AppHandle, Manager};
use tauri_plugin_shell::ShellExt;
use tauri_plugin_shell::process::CommandChild;
use std::time::Duration;
use tokio::task::JoinHandle;

pub struct SidecarManager {
    child: Option<CommandChild>,
    health_check_handle: Option<JoinHandle<()>>,
}

impl SidecarManager {
    pub fn new() -> Self {
        Self {
            child: None,
            health_check_handle: None,
        }
    }

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

    pub async fn stop(&mut self) {
        if let Some(child) = self.child.take() {
            child.kill().unwrap();
        }
        if let Some(handle) = self.health_check_handle.take() {
            handle.abort();
        }
    }
}

impl Drop for SidecarManager {
    fn drop(&mut self) {
        if let Some(child) = self.child.take() {
            let _ = child.kill();
        }
        if let Some(handle) = self.health_check_handle.take() {
            handle.abort();
        }
    }
}
```

- [ ] **Step 2: 验证编译**

Run: `cd desktop/src-tauri && cargo check`
Expected: 编译通过

- [ ] **Step 3: 提交**

```bash
git add desktop/src-tauri/src/sidecar.rs
git commit -m "feat: 实现Sidecar进程管理"
```

---

### Task 6: 实现配置管理器

**Files:**
- Create: `desktop/src-tauri/src/config.rs`

- [ ] **Step 1: 创建配置管理器**

```rust
use serde::{Deserialize, Serialize};
use std::path::PathBuf;
use dirs::config_dir;

#[derive(Debug, Serialize, Deserialize)]
pub struct AppConfig {
    pub window: WindowConfig,
    pub theme: Theme,
    pub language: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct WindowConfig {
    pub x: i32,
    pub y: i32,
    pub width: u32,
    pub height: u32,
    pub maximized: bool,
}

#[derive(Debug, Serialize, Deserialize)]
pub enum Theme {
    Light,
    Dark,
    System,
}

impl Default for AppConfig {
    fn default() -> Self {
        Self {
            window: WindowConfig {
                x: 100,
                y: 100,
                width: 1400,
                height: 900,
                maximized: false,
            },
            theme: Theme::System,
            language: "zh-CN".to_string(),
        }
    }
}

pub struct ConfigManager {
    config: AppConfig,
    config_path: PathBuf,
}

impl ConfigManager {
    pub fn new() -> Self {
        let config_path = config_dir()
            .unwrap_or_else(|| PathBuf::from("."))
            .join("prismproxy")
            .join("config.json");

        let config = if config_path.exists() {
            let content = std::fs::read_to_string(&config_path).unwrap_or_default();
            serde_json::from_str(&content).unwrap_or_default()
        } else {
            AppConfig::default()
        };

        Self {
            config,
            config_path,
        }
    }

    pub fn get_config(&self) -> &AppConfig {
        &self.config
    }

    pub fn update_config(&mut self, config: AppConfig) -> Result<(), std::io::Error> {
        self.config = config;
        self.save()
    }

    fn save(&self) -> Result<(), std::io::Error> {
        if let Some(parent) = self.config_path.parent() {
            std::fs::create_dir_all(parent)?;
        }
        let content = serde_json::to_string_pretty(&self.config).unwrap();
        std::fs::write(&self.config_path, content)
    }
}
```

- [ ] **Step 2: 验证编译**

Run: `cd desktop/src-tauri && cargo check`
Expected: 编译通过

- [ ] **Step 3: 提交**

```bash
git add desktop/src-tauri/src/config.rs
git commit -m "feat: 实现配置管理器"
```

---

### Task 7: 实现SystemService IPC Handlers

**Files:**
- Create: `desktop/src-tauri/src/commands/mod.rs`
- Create: `desktop/src-tauri/src/commands/system.rs`
- Modify: `desktop/src-tauri/src/main.rs`

- [ ] **Step 1: 创建commands模块**

```rust
// desktop/src-tauri/src/commands/mod.rs
pub mod system;
pub mod traffic;
pub mod rules;
pub mod ai;
pub mod breakpoints;
pub mod rewrites;
pub mod collections;
pub mod environments;
pub mod codegen;
pub mod scripts;
pub mod diff;
pub mod perf;
pub mod cert;
pub mod search;
```

- [ ] **Step 2: 创建SystemService命令**

```rust
// desktop/src-tauri/src/commands/system.rs
use tauri::State;
use crate::state::AppState;
use crate::error::AppResult;
use crate::grpc_client::common::Empty;

#[tauri::command]
pub async fn get_system_status(
    state: State<'_, AppState>,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client.system.get_status(Empty {}).await
        .map_err(|e| crate::error::AppError::Grpc(e))?;
    Ok(serde_json::to_string(&response.into_inner()).unwrap())
}

#[tauri::command]
pub async fn start_proxy(
    state: State<'_, AppState>,
) -> AppResult<()> {
    let mut client = state.get_grpc_client().await?;
    client.system.start_proxy(Empty {}).await
        .map_err(|e| crate::error::AppError::Grpc(e))?;
    Ok(())
}

#[tauri::command]
pub async fn stop_proxy(
    state: State<'_, AppState>,
) -> AppResult<()> {
    let mut client = state.get_grpc_client().await?;
    client.system.stop_proxy(Empty {}).await
        .map_err(|e| crate::error::AppError::Grpc(e))?;
    Ok(())
}
```

- [ ] **Step 3: 更新main.rs注册命令**

```rust
// desktop/src-tauri/src/main.rs
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

mod error;
mod grpc_client;
mod sidecar;
mod config;
mod state;
mod commands;

use state::AppState;

fn main() {
    tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .manage(AppState::default())
        .setup(|app| {
            let app_handle = app.handle().clone();
            let state = app_handle.state::<AppState>();

            // 启动sidecar
            let state_clone = state.clone();
            tokio::spawn(async move {
                let mut sidecar = state_clone.sidecar_manager.lock().await;
                sidecar.start(app_handle.clone()).await.unwrap();
            });

            // 初始化gRPC客户端
            let state_clone = state.clone();
            tokio::spawn(async move {
                tokio::time::sleep(std::time::Duration::from_secs(2)).await;
                state_clone.init_grpc_client("http://localhost:9090").await.unwrap();
            });

            Ok(())
        })
        .invoke_handler(tauri::generate_handler![
            commands::system::get_system_status,
            commands::system::start_proxy,
            commands::system::stop_proxy,
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
```

- [ ] **Step 4: 创建lib.rs导出模块**

```rust
// desktop/src-tauri/src/lib.rs
pub mod error;
pub mod grpc_client;
pub mod sidecar;
pub mod config;
pub mod state;
pub mod commands;
```

- [ ] **Step 5: 验证编译**

Run: `cd desktop/src-tauri && cargo check`
Expected: 编译通过

- [ ] **Step 6: 提交**

```bash
git add desktop/src-tauri/src/commands/mod.rs desktop/src-tauri/src/commands/system.rs desktop/src-tauri/src/main.rs desktop/src-tauri/src/lib.rs
git commit -m "feat: 实现SystemService IPC Handlers"
```

---

### Task 8: 实现TrafficService IPC Handlers

**Files:**
- Create: `desktop/src-tauri/src/commands/traffic.rs`
- Modify: `desktop/src-tauri/src/main.rs`

- [ ] **Step 1: 创建TrafficService命令**

```rust
// desktop/src-tauri/src/commands/traffic.rs
use tauri::{State, AppHandle, Emitter};
use crate::state::AppState;
use crate::error::AppResult;
use crate::grpc_client::common::{Empty, Pagination};
use crate::grpc_client::traffic::*;

#[tauri::command]
pub async fn list_traffic(
    state: State<'_, AppState>,
    page: Option<i32>,
    page_size: Option<i32>,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let request = TrafficListRequest {
        pagination: Some(Pagination {
            page: page.unwrap_or(1),
            page_size: page_size.unwrap_or(20),
            sort_by: None,
            sort_desc: None,
        }),
        ..Default::default()
    };
    let response = client.traffic.list(request).await
        .map_err(|e| crate::error::AppError::Grpc(e))?;
    Ok(serde_json::to_string(&response.into_inner()).unwrap())
}

#[tauri::command]
pub async fn get_traffic(
    state: State<'_, AppState>,
    id: String,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let request = TrafficGetRequest { id };
    let response = client.traffic.get(request).await
        .map_err(|e| crate::error::AppError::Grpc(e))?;
    Ok(serde_json::to_string(&response.into_inner()).unwrap())
}

#[tauri::command]
pub async fn delete_traffic(
    state: State<'_, AppState>,
    ids: Vec<String>,
) -> AppResult<()> {
    let mut client = state.get_grpc_client().await?;
    let request = TrafficDeleteRequest { ids };
    client.traffic.delete(request).await
        .map_err(|e| crate::error::AppError::Grpc(e))?;
    Ok(())
}

#[tauri::command]
pub async fn clear_traffic(
    state: State<'_, AppState>,
) -> AppResult<()> {
    let mut client = state.get_grpc_client().await?;
    client.traffic.clear(Empty {}).await
        .map_err(|e| crate::error::AppError::Grpc(e))?;
    Ok(())
}

#[tauri::command]
pub async fn subscribe_traffic(
    app: AppHandle,
    state: State<'_, AppState>,
) -> AppResult<()> {
    let mut client = state.get_grpc_client().await?;
    let mut stream = client.traffic.subscribe(Empty {}).await
        .map_err(|e| crate::error::AppError::Grpc(e))?
        .into_inner();

    // 在后台任务中处理流
    tokio::spawn(async move {
        while let Ok(Some(event)) = stream.message().await {
            // 将gRPC事件转换为Tauri事件
            app.emit("traffic:event", serde_json::to_string(&event).unwrap()).unwrap();
        }
    });

    Ok(())
}
```

- [ ] **Step 2: 更新main.rs注册命令**

在 `invoke_handler` 中添加：
```rust
commands::traffic::list_traffic,
commands::traffic::get_traffic,
commands::traffic::delete_traffic,
commands::traffic::clear_traffic,
commands::traffic::subscribe_traffic,
```

- [ ] **Step 3: 验证编译**

Run: `cd desktop/src-tauri && cargo check`
Expected: 编译通过

- [ ] **Step 4: 提交**

```bash
git add desktop/src-tauri/src/commands/traffic.rs desktop/src-tauri/src/main.rs
git commit -m "feat: 实现TrafficService IPC Handlers"
```

---

### Task 9: 实现RulesService IPC Handlers

**Files:**
- Create: `desktop/src-tauri/src/commands/rules.rs`
- Modify: `desktop/src-tauri/src/main.rs`

- [ ] **Step 1: 创建RulesService命令**

```rust
// desktop/src-tauri/src/commands/rules.rs
use tauri::State;
use crate::state::AppState;
use crate::error::AppResult;
use crate::grpc_client::common::Empty;
use crate::grpc_client::rules::*;

#[tauri::command]
pub async fn list_rules(
    state: State<'_, AppState>,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client.rules.list(Empty {}).await
        .map_err(|e| crate::error::AppError::Grpc(e))?;
    Ok(serde_json::to_string(&response.into_inner()).unwrap())
}

#[tauri::command]
pub async fn get_rule(
    state: State<'_, AppState>,
    id: String,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let request = RuleGetRequest { id };
    let response = client.rules.get(request).await
        .map_err(|e| crate::error::AppError::Grpc(e))?;
    Ok(serde_json::to_string(&response.into_inner()).unwrap())
}

#[tauri::command]
pub async fn create_rule(
    state: State<'_, AppState>,
    rule: String,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let request: RuleCreateRequest = serde_json::from_str(&rule).unwrap();
    let response = client.rules.create(request).await
        .map_err(|e| crate::error::AppError::Grpc(e))?;
    Ok(serde_json::to_string(&response.into_inner()).unwrap())
}

#[tauri::command]
pub async fn update_rule(
    state: State<'_, AppState>,
    rule: String,
) -> AppResult<()> {
    let mut client = state.get_grpc_client().await?;
    let request: RuleUpdateRequest = serde_json::from_str(&rule).unwrap();
    client.rules.update(request).await
        .map_err(|e| crate::error::AppError::Grpc(e))?;
    Ok(())
}

#[tauri::command]
pub async fn delete_rule(
    state: State<'_, AppState>,
    id: String,
) -> AppResult<()> {
    let mut client = state.get_grpc_client().await?;
    let request = RuleDeleteRequest { id };
    client.rules.delete(request).await
        .map_err(|e| crate::error::AppError::Grpc(e))?;
    Ok(())
}
```

- [ ] **Step 2: 更新main.rs注册命令**

在 `invoke_handler` 中添加：
```rust
commands::rules::list_rules,
commands::rules::get_rule,
commands::rules::create_rule,
commands::rules::update_rule,
commands::rules::delete_rule,
```

- [ ] **Step 3: 验证编译**

Run: `cd desktop/src-tauri && cargo check`
Expected: 编译通过

- [ ] **Step 4: 提交**

```bash
git add desktop/src-tauri/src/commands/rules.rs desktop/src-tauri/src/main.rs
git commit -m "feat: 实现RulesService IPC Handlers"
```

---

### Task 10: 实现AIService IPC Handlers（包含流式处理）

**Files:**
- Create: `desktop/src-tauri/src/commands/ai.rs`
- Modify: `desktop/src-tauri/src/main.rs`

- [ ] **Step 1: 创建AIService命令**

```rust
// desktop/src-tauri/src/commands/ai.rs
use tauri::{State, AppHandle, Emitter};
use crate::state::AppState;
use crate::error::AppResult;
use crate::grpc_client::common::Empty;
use crate::grpc_client::ai::*;

#[tauri::command]
pub async fn chat(
    state: State<'_, AppState>,
    message: String,
    conversation_id: Option<String>,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let request = ChatRequest {
        message,
        conversation_id: conversation_id.unwrap_or_default(),
        ..Default::default()
    };
    let response = client.ai.chat(request).await
        .map_err(|e| crate::error::AppError::Grpc(e))?;
    Ok(serde_json::to_string(&response.into_inner()).unwrap())
}

#[tauri::command]
pub async fn stream_chat(
    app: AppHandle,
    state: State<'_, AppState>,
    message: String,
    conversation_id: Option<String>,
) -> AppResult<()> {
    let mut client = state.get_grpc_client().await?;
    let request = ChatRequest {
        message,
        conversation_id: conversation_id.unwrap_or_default(),
        ..Default::default()
    };
    let mut stream = client.ai.stream_chat(request).await
        .map_err(|e| crate::error::AppError::Grpc(e))?
        .into_inner();

    // 在后台任务中处理流
    tokio::spawn(async move {
        while let Ok(Some(chunk)) = stream.message().await {
            // 将gRPC流转换为Tauri事件
            app.emit("ai:chat_chunk", serde_json::to_string(&chunk).unwrap()).unwrap();
        }
        // 流结束
        app.emit("ai:chat_end", ()).unwrap();
    });

    Ok(())
}

#[tauri::command]
pub async fn check_ai_availability(
    state: State<'_, AppState>,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client.ai.check_availability(Empty {}).await
        .map_err(|e| crate::error::AppError::Grpc(e))?;
    Ok(serde_json::to_string(&response.into_inner()).unwrap())
}
```

- [ ] **Step 2: 更新main.rs注册命令**

在 `invoke_handler` 中添加：
```rust
commands::ai::chat,
commands::ai::stream_chat,
commands::ai::check_ai_availability,
```

- [ ] **Step 3: 验证编译**

Run: `cd desktop/src-tauri && cargo check`
Expected: 编译通过

- [ ] **Step 4: 提交**

```bash
git add desktop/src-tauri/src/commands/ai.rs desktop/src-tauri/src/main.rs
git commit -m "feat: 实现AIService IPC Handlers"
```

---

### Task 11: 实现其他Service IPC Handlers

**Files:**
- Create: `desktop/src-tauri/src/commands/breakpoints.rs`
- Create: `desktop/src-tauri/src/commands/rewrites.rs`
- Create: `desktop/src-tauri/src/commands/collections.rs`
- Create: `desktop/src-tauri/src/commands/environments.rs`
- Create: `desktop/src-tauri/src/commands/codegen.rs`
- Create: `desktop/src-tauri/src/commands/scripts.rs`
- Create: `desktop/src-tauri/src/commands/diff.rs`
- Create: `desktop/src-tauri/src/commands/perf.rs`
- Create: `desktop/src-tauri/src/commands/cert.rs`
- Create: `desktop/src-tauri/src/commands/search.rs`
- Modify: `desktop/src-tauri/src/main.rs`

- [ ] **Step 1: 创建BreakpointsService命令**

```rust
// desktop/src-tauri/src/commands/breakpoints.rs
use tauri::{State, AppHandle, Emitter};
use crate::state::AppState;
use crate::error::AppResult;
use crate::grpc_client::common::Empty;
use crate::grpc_client::breakpoints::*;

#[tauri::command]
pub async fn list_breakpoints(
    state: State<'_, AppState>,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client.breakpoints.list(Empty {}).await
        .map_err(|e| crate::error::AppError::Grpc(e))?;
    Ok(serde_json::to_string(&response.into_inner()).unwrap())
}

#[tauri::command]
pub async fn subscribe_breakpoints(
    app: AppHandle,
    state: State<'_, AppState>,
) -> AppResult<()> {
    let mut client = state.get_grpc_client().await?;
    let mut stream = client.breakpoints.subscribe(Empty {}).await
        .map_err(|e| crate::error::AppError::Grpc(e))?
        .into_inner();

    tokio::spawn(async move {
        while let Ok(Some(event)) = stream.message().await {
            app.emit("breakpoints:event", serde_json::to_string(&event).unwrap()).unwrap();
        }
    });

    Ok(())
}
```

- [ ] **Step 2: 创建其他Service命令（模式相同）**

为每个服务创建类似的命令文件，包含：
- `list_*` - 列表查询
- `get_*` - 单个查询
- `create_*` - 创建
- `update_*` - 更新
- `delete_*` - 删除
- `subscribe_*` - 订阅（如果服务支持）

- [ ] **Step 3: 更新main.rs注册所有命令**

在 `invoke_handler` 中添加所有新命令。

- [ ] **Step 4: 验证编译**

Run: `cd desktop/src-tauri && cargo check`
Expected: 编译通过

- [ ] **Step 5: 提交**

```bash
git add desktop/src-tauri/src/commands/*.rs desktop/src-tauri/src/main.rs
git commit -m "feat: 实现所有Service IPC Handlers"
```

---

### Task 12: 前端迁移 - 更新traffic服务

**Files:**
- Modify: `desktop/src/services/traffic.ts`

- [ ] **Step 1: 更新traffic服务**

```typescript
// desktop/src/services/traffic.ts
import { invoke } from '@tauri-apps/api/core';
import { listen } from '@tauri-apps/api/event';

export interface TrafficListResponse {
  entries: TrafficEntry[];
  pagination: PageMeta;
}

export interface TrafficEntry {
  id: string;
  method: string;
  url: string;
  host: string;
  path: string;
  scheme: string;
  port: number;
  // ... 其他字段
}

export interface PageMeta {
  page: number;
  pageSize: number;
  total: number;
}

export async function listTraffic(page: number = 1, pageSize: number = 20): Promise<TrafficListResponse> {
  return invoke('list_traffic', { page, pageSize });
}

export async function getTraffic(id: string): Promise<TrafficEntry> {
  return invoke('get_traffic', { id });
}

export async function deleteTraffic(ids: string[]): Promise<void> {
  return invoke('delete_traffic', { ids });
}

export async function clearTraffic(): Promise<void> {
  return invoke('clear_traffic');
}

export async function subscribeTraffic(callback: (event: any) => void): Promise<() => void> {
  // 启动订阅
  await invoke('subscribe_traffic');

  // 监听事件
  const unlisten = await listen('traffic:event', (event) => {
    callback(JSON.parse(event.payload as string));
  });

  return unlisten;
}
```

- [ ] **Step 2: 验证前端编译**

Run: `cd desktop && npm run build`
Expected: 编译通过

- [ ] **Step 3: 提交**

```bash
git add desktop/src/services/traffic.ts
git commit -m "feat: 迁移traffic服务到Tauri IPC"
```

---

### Task 13: 前端迁移 - 更新其他服务

**Files:**
- Modify: `desktop/src/services/rules.ts`
- Modify: `desktop/src/services/ai.ts`
- Modify: `desktop/src/services/breakpoints.ts`
- Modify: `desktop/src/services/rewrites.ts`
- Modify: `desktop/src/services/collections.ts`
- Modify: `desktop/src/services/environments.ts`
- Modify: `desktop/src/services/cert.ts`
- Modify: `desktop/src/services/scripts.ts`
- Modify: `desktop/src/services/search.ts`
- Modify: `desktop/src/services/diff.ts`
- Modify: `desktop/src/services/perf.ts`
- Modify: `desktop/src/services/proxy.ts`
- Delete: `desktop/src/services/api.ts`
- Delete: `desktop/src/services/grpc.ts`

- [ ] **Step 1: 更新rules服务**

```typescript
// desktop/src/services/rules.ts
import { invoke } from '@tauri-apps/api/core';

export interface Rule {
  id: string;
  name: string;
  enabled: boolean;
  // ... 其他字段
}

export async function listRules(): Promise<Rule[]> {
  return invoke('list_rules');
}

export async function getRule(id: string): Promise<Rule> {
  return invoke('get_rule', { id });
}

export async function createRule(rule: Partial<Rule>): Promise<Rule> {
  return invoke('create_rule', { rule: JSON.stringify(rule) });
}

export async function updateRule(rule: Rule): Promise<void> {
  return invoke('update_rule', { rule: JSON.stringify(rule) });
}

export async function deleteRule(id: string): Promise<void> {
  return invoke('delete_rule', { id });
}
```

- [ ] **Step 2: 更新ai服务（包含流式处理）**

```typescript
// desktop/src/services/ai.ts
import { invoke } from '@tauri-apps/api/core';
import { listen } from '@tauri-apps/api/event';

export interface ChatMessage {
  role: 'user' | 'assistant';
  content: string;
}

export async function chat(message: string, conversationId?: string): Promise<ChatMessage> {
  return invoke('chat', { message, conversationId });
}

export async function streamChat(
  message: string,
  conversationId: string | undefined,
  onChunk: (chunk: string) => void,
  onEnd: () => void
): Promise<() => void> {
  // 启动流式聊天
  await invoke('stream_chat', { message, conversationId });

  // 监听chunk事件
  const unlistenChunk = await listen('ai:chat_chunk', (event) => {
    const chunk = JSON.parse(event.payload as string);
    onChunk(chunk.content);
  });

  // 监听结束事件
  const unlistenEnd = await listen('ai:chat_end', () => {
    onEnd();
  });

  return () => {
    unlistenChunk();
    unlistenEnd();
  };
}

export async function checkAiAvailability(): Promise<boolean> {
  return invoke('check_ai_availability');
}
```

- [ ] **Step 3: 更新其他服务（模式相同）**

为每个服务创建类似的实现，使用 `invoke` 调用对应的IPC命令。

- [ ] **Step 4: 删除旧的api.ts和grpc.ts**

```bash
rm desktop/src/services/api.ts desktop/src/services/grpc.ts
```

- [ ] **Step 5: 验证前端编译**

Run: `cd desktop && npm run build`
Expected: 编译通过

- [ ] **Step 6: 提交**

```bash
git add desktop/src/services/*.ts
git commit -m "feat: 迁移所有服务到Tauri IPC"
```

---

### Task 14: 集成测试

**Files:**
- Test: 手动测试所有功能

- [ ] **Step 1: 启动应用**

Run: `cd desktop && npm run dev`
Expected: 应用启动，Go sidecar进程启动

- [ ] **Step 2: 测试普通请求**

1. 打开流量列表页面
2. 验证能正确加载数据
3. 测试分页功能
4. 测试删除、清空操作

Expected: 所有操作正常工作

- [ ] **Step 3: 测试流式请求**

1. 发送一个HTTP请求通过代理
2. 验证流量列表实时更新
3. 测试AI聊天功能（如果配置了API key）

Expected: 流式数据实时推送

- [ ] **Step 4: 测试错误处理**

1. 停止Go sidecar进程
2. 尝试执行操作
3. 验证错误信息正确显示

Expected: 错误信息友好提示

- [ ] **Step 5: 测试Sidecar管理**

1. 杀死Go sidecar进程
2. 等待5秒健康检查
3. 验证进程自动重启

Expected: 进程自动恢复

- [ ] **Step 6: 提交最终代码**

```bash
git add .
git commit -m "feat: 完成架构迁移 - 前端→Rust→Go"
```

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

本实现计划将PrismProxy从"前端直连Go"迁移到"前端→Rust→Go"架构。

**核心原则**：
- Rust层只做转发，业务逻辑留在Go后端
- 保持现有proto定义不变
- 一次性迁移所有服务
- 仅支持Tauri模式

**工作量估算**：14个任务，约8天工作量

**下一步**：执行计划，逐个完成任务。
