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
            // System
            commands::system::get_system_status,
            commands::system::start_proxy,
            commands::system::stop_proxy,
            // Traffic
            commands::traffic::list_traffic,
            commands::traffic::get_traffic,
            commands::traffic::delete_traffic,
            commands::traffic::clear_traffic,
            commands::traffic::subscribe_traffic,
            // Rules
            commands::rules::list_rules,
            commands::rules::get_rule,
            commands::rules::create_rule,
            commands::rules::update_rule,
            commands::rules::delete_rule,
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
