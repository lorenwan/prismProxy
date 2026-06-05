use tauri::{AppHandle, Emitter};
use tauri_plugin_shell::process::CommandChild;
use tauri_plugin_shell::ShellExt;
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

    /// 启动 sidecar 进程并开始健康检查
    pub async fn start(&mut self, app: AppHandle) -> Result<(), String> {
        // 启动 sidecar
        let (_rx, child) = app
            .shell()
            .sidecar("prismproxy-server")
            .map_err(|e| e.to_string())?
            .args(&["--port", "9090", "--proxy-port", "8888"])
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
                tokio::time::sleep(std::time::Duration::from_secs(5)).await;
                if reqwest::get("http://localhost:8080/health").await.is_err() {
                    // 健康检查失败，通知前端
                    let _ = app.emit("sidecar:health_check_failed", ());
                }
            }
        });
        self.health_check_handle = Some(handle);
    }

    /// 停止 sidecar 进程并取消健康检查
    pub async fn stop(&mut self) {
        if let Some(child) = self.child.take() {
            let _ = child.kill();
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
