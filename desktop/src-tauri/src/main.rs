#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

use prismproxy_desktop::*;
use state::AppState;
use tauri::Manager;

fn main() {
    tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .manage(AppState::default())
        .setup(|app| {
            let app_handle = app.handle().clone();

            // 获取 AppState（Clone 产生独立副本，不持有引用）
            let state: AppState = app.state::<AppState>().inner().clone();

            // 启动 sidecar
            {
                let state_clone = state.clone();
                let app_clone = app_handle.clone();
                tokio::spawn(async move {
                    let mut sidecar = state_clone.sidecar_manager.lock().await;
                    if let Err(e) = sidecar.start(app_clone).await {
                        eprintln!("[Tauri] Failed to start sidecar: {}", e);
                    }
                });
            }

            // 初始化 gRPC 客户端
            {
                let state_clone = state.clone();
                tokio::spawn(async move {
                    tokio::time::sleep(std::time::Duration::from_secs(2)).await;
                    if let Err(e) = state_clone.init_grpc_client("http://localhost:9090").await {
                        eprintln!("[Tauri] Failed to init gRPC client: {}", e);
                    }
                });
            }

            Ok(())
        })
        .on_window_event(|window, event| {
            if let tauri::WindowEvent::CloseRequested { .. } = event {
                if let Some(state) = window.try_state::<AppState>() {
                    let state = (*state).clone();
                    tokio::spawn(async move {
                        let mut sidecar = state.sidecar_manager.lock().await;
                        sidecar.stop().await;
                        println!("[Tauri] PrismProxy sidecar stopped");
                    });
                }
            }
        })
        .invoke_handler(tauri::generate_handler![
            commands::system::get_system_status,
            commands::system::start_proxy,
            commands::system::stop_proxy,
            commands::traffic::list_traffic,
            commands::traffic::get_traffic,
            commands::traffic::delete_traffic,
            commands::traffic::clear_traffic,
            commands::traffic::subscribe_traffic,
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
