use tauri::State;

use crate::error::AppResult;
use crate::grpc_client::Empty;
use crate::state::AppState;

/// 获取系统状态
#[tauri::command]
pub async fn get_system_status(
    state: State<'_, AppState>,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .system
        .get_status(Empty {})
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}

/// 启动代理
#[tauri::command]
pub async fn start_proxy(
    state: State<'_, AppState>,
) -> AppResult<()> {
    let mut client = state.get_grpc_client().await?;
    client
        .system
        .start_proxy(Empty {})
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(())
}

/// 停止代理
#[tauri::command]
pub async fn stop_proxy(
    state: State<'_, AppState>,
) -> AppResult<()> {
    let mut client = state.get_grpc_client().await?;
    client
        .system
        .stop_proxy(Empty {})
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(())
}

/// 启用系统代理
#[tauri::command]
pub async fn enable_system_proxy(
    state: State<'_, AppState>,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .system
        .enable_system_proxy(Empty {})
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}

/// 禁用系统代理
#[tauri::command]
pub async fn disable_system_proxy(
    state: State<'_, AppState>,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .system
        .disable_system_proxy(Empty {})
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}

/// 获取系统代理状态
#[tauri::command]
pub async fn get_system_proxy_status(
    state: State<'_, AppState>,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .system
        .get_system_proxy_status(Empty {})
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}

/// 获取应用设置
#[tauri::command]
pub async fn get_settings(
    state: State<'_, AppState>,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .system
        .get_settings(Empty {})
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}

/// 更新应用设置
#[tauri::command]
pub async fn update_settings(
    state: State<'_, AppState>,
    settings: String,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let app_settings: crate::grpc_client::Settings = serde_json::from_str(&settings)
        .map_err(|e| crate::error::AppError::Connection(format!("JSON 解析失败: {}", e)))?;
    let response = client
        .system
        .update_settings(app_settings)
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}
