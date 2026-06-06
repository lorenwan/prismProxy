use tauri::{AppHandle, Emitter, Manager};
use tauri_plugin_shell::process::CommandChild;
use tauri_plugin_shell::ShellExt;
use tokio::task::JoinHandle;

/// 端口占用检查
fn is_port_available(port: u16) -> bool {
    std::net::TcpListener::bind(("127.0.0.1", port)).is_ok()
}

/// 检查所有需要的端口是否可用，返回被占用的端口列表
fn check_ports() -> Vec<u16> {
    let ports = [9090, 8080, 8888];
    ports
        .iter()
        .filter(|&&port| !is_port_available(port))
        .copied()
        .collect()
}

pub struct SidecarManager {
    child: Option<CommandChild>,
    health_check_handle: Option<JoinHandle<()>>,
    app_handle: Option<AppHandle>,
    restart_count: u32,
}

/// 最大自动重启次数，防止无限重启循环
const MAX_RESTART_ATTEMPTS: u32 = 3;

impl SidecarManager {
    pub fn new() -> Self {
        Self {
            child: None,
            health_check_handle: None,
            app_handle: None,
            restart_count: 0,
        }
    }

    /// 启动 sidecar 进程并开始健康检查
    pub async fn start(&mut self, app: AppHandle) -> Result<(), String> {
        self.app_handle = Some(app.clone());

        // 端口占用检查
        let occupied_ports = check_ports();
        if !occupied_ports.is_empty() {
            return Err(format!(
                "以下端口被占用，无法启动服务: {:?}。请关闭占用这些端口的程序后重试。",
                occupied_ports
            ));
        }

        self.spawn_sidecar(&app).await
    }

    /// 生成 sidecar 子进程
    async fn spawn_sidecar(&mut self, app: &AppHandle) -> Result<(), String> {
        // 获取应用数据目录，数据库文件存放在此处
        let db_path = app
            .path()
            .app_data_dir()
            .map(|p| p.join("prismproxy.db").to_string_lossy().to_string())
            .unwrap_or_else(|_| "./data/prismproxy.db".to_string());

        let (rx, child) = app
            .shell()
            .sidecar("prismproxy-server")
            .map_err(|e| e.to_string())?
            .args(&[
                "--port", "9090",
                "--http-port", "8080",
                "--proxy-port", "8888",
                "--db-path", &db_path,
            ])
            .spawn()
            .map_err(|e| e.to_string())?;

        self.child = Some(child);

        // 消费 stdout/stderr 输出，打印到控制台便于调试
        tokio::spawn(async move {
            let mut rx = rx;
            while let Some(event) = rx.recv().await {
                use tauri_plugin_shell::process::CommandEvent;
                match event {
                    CommandEvent::Stdout(line) => {
                        eprintln!("[Sidecar stdout] {}", String::from_utf8_lossy(&line));
                    }
                    CommandEvent::Stderr(line) => {
                        eprintln!("[Sidecar stderr] {}", String::from_utf8_lossy(&line));
                    }
                    CommandEvent::Error(err) => {
                        eprintln!("[Sidecar error] {}", err);
                    }
                    CommandEvent::Terminated(status) => {
                        eprintln!("[Sidecar] Process terminated with status: {:?}", status);
                    }
                    _ => {}
                }
            }
        });

        // 启动健康检查
        self.start_health_check(app.clone());

        Ok(())
    }

    fn start_health_check(&mut self, app: AppHandle) {
        let app_handle = app.clone();
        let handle = tokio::spawn(async move {
            // 等待 sidecar 启动
            tokio::time::sleep(std::time::Duration::from_secs(3)).await;

            loop {
                tokio::time::sleep(std::time::Duration::from_secs(5)).await;
                if reqwest::get("http://localhost:8080/health").await.is_err() {
                    // 健康检查失败，通知前端
                    let _ = app_handle.emit("sidecar:health_check_failed", ());
                }
            }
        });
        self.health_check_handle = Some(handle);
    }

    /// 重启 sidecar 进程（用于崩溃恢复）
    pub async fn restart(&mut self) -> Result<(), String> {
        if self.restart_count >= MAX_RESTART_ATTEMPTS {
            return Err(format!(
                "sidecar 已自动重启 {} 次仍失败，请手动检查",
                MAX_RESTART_ATTEMPTS
            ));
        }

        self.restart_count += 1;
        eprintln!(
            "[Sidecar] Attempting restart ({}/{})...",
            self.restart_count, MAX_RESTART_ATTEMPTS
        );

        // 停止现有进程和健康检查
        self.stop_internal().await;

        // 等待端口释放
        tokio::time::sleep(std::time::Duration::from_secs(2)).await;

        // 重新启动
        if let Some(app) = self.app_handle.clone() {
            self.spawn_sidecar(&app).await
        } else {
            Err("无法重启: AppHandle 未初始化".to_string())
        }
    }

    /// 重置重启计数（成功连接后调用）
    pub fn reset_restart_count(&mut self) {
        self.restart_count = 0;
    }

    /// 获取当前重启次数
    pub fn restart_count(&self) -> u32 {
        self.restart_count
    }

    /// 内部停止逻辑
    async fn stop_internal(&mut self) {
        if let Some(child) = self.child.take() {
            let _ = child.kill();
        }
        if let Some(handle) = self.health_check_handle.take() {
            handle.abort();
        }
    }

    /// 停止 sidecar 进程并取消健康检查
    pub async fn stop(&mut self) {
        self.stop_internal().await;
    }

    /// 同步停止 sidecar（用于窗口关闭时，确保在进程退出前完成）
    pub fn stop_sync(&mut self) {
        if let Some(child) = self.child.take() {
            let _ = child.kill();
        }
        if let Some(handle) = self.health_check_handle.take() {
            handle.abort();
        }
    }

    /// 检查进程是否仍在运行
    pub fn is_running(&self) -> bool {
        self.child.is_some()
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
