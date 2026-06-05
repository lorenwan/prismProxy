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
    Ok(serde_json::to_string(&response.into_inner()).unwrap())
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
