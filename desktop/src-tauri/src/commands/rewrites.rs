use tauri::State;

use crate::error::AppResult;
use crate::grpc_client::Empty;
use crate::grpc_client::RewriteDeleteRequest;
use crate::grpc_client::RewriteGetRequest;
use crate::grpc_client::RewriteRule;
use crate::grpc_client::RewriteToggleRequest;
use crate::state::AppState;

/// 获取重写规则列表
#[tauri::command]
pub async fn list_rewrites(state: State<'_, AppState>) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .rewrites
        .list(Empty {})
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}

/// 获取单条重写规则
#[tauri::command]
pub async fn get_rewrite(
    state: State<'_, AppState>,
    id: String,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .rewrites
        .get(RewriteGetRequest { id })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}

/// 创建重写规则
#[tauri::command]
pub async fn create_rewrite(
    state: State<'_, AppState>,
    rewrite: String,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let request: RewriteRule = serde_json::from_str(&rewrite)
        .map_err(|e| crate::error::AppError::Connection(format!("JSON 解析失败: {}", e)))?;
    let response = client
        .rewrites
        .create(request)
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}

/// 更新重写规则
#[tauri::command]
pub async fn update_rewrite(
    state: State<'_, AppState>,
    rewrite: String,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let request: RewriteRule = serde_json::from_str(&rewrite)
        .map_err(|e| crate::error::AppError::Connection(format!("JSON 解析失败: {}", e)))?;
    let response = client
        .rewrites
        .update(request)
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}

/// 删除重写规则
#[tauri::command]
pub async fn delete_rewrite(
    state: State<'_, AppState>,
    id: String,
) -> AppResult<()> {
    let mut client = state.get_grpc_client().await?;
    client
        .rewrites
        .delete(RewriteDeleteRequest { id })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(())
}

/// 切换重写规则启用状态
#[tauri::command]
pub async fn toggle_rewrite(
    state: State<'_, AppState>,
    id: String,
    enabled: bool,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .rewrites
        .toggle(RewriteToggleRequest { id, enabled })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}
