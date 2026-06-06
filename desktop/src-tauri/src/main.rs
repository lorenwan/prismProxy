#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

use prismproxy_desktop::*;
use state::AppState;
use tauri::{Emitter, Listener, Manager};

fn main() {
    tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .manage(AppState::default())
        .setup(|app| {
            let app_handle = app.handle().clone();

            // 获取 AppState（Clone 产生独立副本，不持有引用）
            let state: AppState = app.state::<AppState>().inner().clone();

            // 启动 sidecar（使用 tauri async_runtime，因为 setup 回调不在 Tokio runtime 上下文中）
            {
                let state_clone = state.clone();
                let app_clone = app_handle.clone();
                tauri::async_runtime::spawn(async move {
                    let mut sidecar = state_clone.sidecar_manager.lock().await;
                    if let Err(e) = sidecar.start(app_clone).await {
                        eprintln!("[Tauri] Failed to start sidecar: {}", e);
                    }
                });
            }

            // 初始化 gRPC 客户端（带重试机制，20次 * 1.5秒 = 30秒窗口）
            {
                let state_clone = state.clone();
                tauri::async_runtime::spawn(async move {
                    let max_retries = 20;
                    let retry_interval = std::time::Duration::from_millis(1500);

                    for attempt in 1..=max_retries {
                        tokio::time::sleep(retry_interval).await;
                        eprintln!("[Tauri] Attempting to connect gRPC client ({}/{})...", attempt, max_retries);
                        match state_clone.init_grpc_client("http://localhost:9090").await {
                            Ok(()) => {
                                eprintln!("[Tauri] gRPC client connected successfully");
                                // 连接成功，重置 sidecar 重启计数
                                let mut sidecar = state_clone.sidecar_manager.lock().await;
                                sidecar.reset_restart_count();
                                return;
                            }
                            Err(e) => {
                                eprintln!("[Tauri] gRPC client connection attempt {} failed: {}", attempt, e);
                            }
                        }
                    }
                    eprintln!("[Tauri] Failed to init gRPC client after {} retries", max_retries);
                });
            }

            // 监听 sidecar 健康检查失败事件，尝试自动重启
            {
                let state_clone = state.clone();
                let app_clone = app_handle.clone();
                app_handle.listen("sidecar:health_check_failed", move |_event| {
                    let state = state_clone.clone();
                    let app = app_clone.clone();
                    tauri::async_runtime::spawn(async move {
                        eprintln!("[Tauri] Health check failed, attempting sidecar restart...");
                        let mut sidecar = state.sidecar_manager.lock().await;
                        match sidecar.restart().await {
                            Ok(()) => {
                                eprintln!("[Tauri] Sidecar restarted successfully");
                                // 重启后重新初始化 gRPC 客户端
                                drop(sidecar); // 释放锁
                                if let Err(e) = state.init_grpc_client("http://localhost:9090").await {
                                    eprintln!("[Tauri] Failed to reinit gRPC client after restart: {}", e);
                                }
                            }
                            Err(e) => {
                                eprintln!("[Tauri] Sidecar restart failed: {}", e);
                                let _ = app.emit("sidecar:restart_failed", e);
                            }
                        }
                    });
                });
            }

            Ok(())
        })
        .on_window_event(|window, event| {
            if let tauri::WindowEvent::CloseRequested { .. } = event {
                if let Some(state) = window.try_state::<AppState>() {
                    let sidecar_mgr = state.sidecar_manager.clone();
                    let lock_result = sidecar_mgr.try_lock();
                    if let Ok(mut sidecar) = lock_result {
                        sidecar.stop_sync();
                        println!("[Tauri] PrismProxy sidecar stopped");
                    }
                }
            }
        })
        .invoke_handler(tauri::generate_handler![
            // System
            commands::system::get_system_status,
            commands::system::start_proxy,
            commands::system::stop_proxy,
            commands::system::enable_system_proxy,
            commands::system::disable_system_proxy,
            commands::system::get_system_proxy_status,
            commands::system::get_settings,
            commands::system::update_settings,
            // Traffic
            commands::traffic::list_traffic,
            commands::traffic::get_traffic,
            commands::traffic::delete_traffic,
            commands::traffic::clear_traffic,
            commands::traffic::subscribe_traffic,
            commands::traffic::get_traffic_stats,
            commands::traffic::update_traffic_bookmark,
            commands::traffic::update_traffic_notes,
            commands::traffic::update_traffic_color,
            commands::traffic::update_traffic_tags,
            // Rules
            commands::rules::list_rules,
            commands::rules::get_rule,
            commands::rules::create_rule,
            commands::rules::update_rule,
            commands::rules::delete_rule,
            commands::rules::toggle_rule,
            commands::rules::get_rule_stats,
            // AI
            commands::ai::chat,
            commands::ai::stream_chat,
            commands::ai::check_ai_availability,
            // Breakpoints
            commands::breakpoints::list_breakpoints,
            commands::breakpoints::get_breakpoint,
            commands::breakpoints::create_breakpoint,
            commands::breakpoints::update_breakpoint,
            commands::breakpoints::delete_breakpoint,
            commands::breakpoints::toggle_breakpoint,
            commands::breakpoints::list_breakpoint_sessions,
            commands::breakpoints::resolve_breakpoint_session,
            commands::breakpoints::subscribe_breakpoints,
            // Rewrites
            commands::rewrites::list_rewrites,
            commands::rewrites::get_rewrite,
            commands::rewrites::create_rewrite,
            commands::rewrites::update_rewrite,
            commands::rewrites::delete_rewrite,
            commands::rewrites::toggle_rewrite,
            // Collections
            commands::collections::list_collections,
            commands::collections::get_collection,
            commands::collections::create_collection,
            commands::collections::update_collection,
            commands::collections::delete_collection,
            commands::collections::add_collection_request,
            commands::collections::update_collection_request,
            commands::collections::delete_collection_request,
            commands::collections::execute_collection_request,
            commands::collections::export_collection,
            commands::collections::import_collection,
            // Environments
            commands::environments::list_environments,
            commands::environments::get_environment,
            commands::environments::create_environment,
            commands::environments::update_environment,
            commands::environments::delete_environment,
            commands::environments::activate_environment,
            commands::environments::export_environment,
            commands::environments::import_environment,
            // CodeGen
            commands::codegen::generate_code,
            commands::codegen::list_codegen_languages,
            // Scripts
            commands::scripts::list_scripts,
            commands::scripts::get_script,
            commands::scripts::create_script,
            commands::scripts::update_script,
            commands::scripts::delete_script,
            commands::scripts::toggle_script,
            commands::scripts::execute_script,
            // Diff
            commands::diff::compare_headers,
            commands::diff::compare_body,
            commands::diff::compare_json,
            commands::diff::compare_query,
            // Perf
            commands::perf::get_perf_stats,
            commands::perf::get_slow_requests,
            commands::perf::get_domain_stats,
            commands::perf::get_perf_timeline,
            commands::perf::get_status_code_stats,
            commands::perf::get_method_stats,
            commands::perf::get_recent_stats,
            // Cert
            commands::cert::get_ca_info,
            commands::cert::generate_ca,
            commands::cert::export_ca,
            commands::cert::issue_cert,
            commands::cert::list_certs,
            commands::cert::delete_cert,
            commands::cert::check_cert,
            commands::cert::clear_certs,
            commands::cert::get_trust_status,
            // Search
            commands::search::search,
            commands::search::search_by_method,
            commands::search::search_by_host,
            commands::search::search_by_status_code,
            commands::search::search_slow_requests,
            commands::search::get_search_stats,
            commands::search::save_filter,
            commands::search::list_filters,
            commands::search::delete_filter,
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
