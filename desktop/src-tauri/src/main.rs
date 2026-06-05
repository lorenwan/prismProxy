mod error;
mod state;

use std::sync::Mutex;
use tauri::Manager;
use tauri_plugin_shell::{process::CommandChild, ShellExt};

fn main() {
    tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .setup(|app| {
            // 启动 Go sidecar (gRPC 服务器)
            let sidecar_command = app.shell().sidecar("prismproxy-server").unwrap();
            let (_rx, child) = sidecar_command
                .args(&["--port", "9090", "--proxy-port", "8080"])
                .spawn()
                .expect("Failed to spawn PrismProxy server sidecar");

            // 存储子进程句柄，关闭窗口时清理
            app.manage(Mutex::new(Some(child)));

            println!("[Tauri] PrismProxy sidecar started");
            Ok(())
        })
        .on_window_event(|window, event| {
            if let tauri::WindowEvent::CloseRequested { .. } = event {
                // 关闭时停止 sidecar
                if let Some(state) = window.try_state::<Mutex<Option<CommandChild>>>() {
                    if let Ok(mut guard) = state.lock() {
                        if let Some(child) = guard.take() {
                            let _ = child.kill();
                            println!("[Tauri] PrismProxy sidecar stopped");
                        }
                    }
                }
            }
        })
        .run(tauri::generate_context!())
        .expect("error while running PrismProxy");
}
