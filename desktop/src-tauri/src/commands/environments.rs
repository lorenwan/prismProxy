use tauri::State;

use crate::error::AppResult;
use crate::grpc_client::Empty;
use crate::grpc_client::Environment;
use crate::grpc_client::EnvironmentActivateRequest;
use crate::grpc_client::EnvironmentDeleteRequest;
use crate::grpc_client::EnvironmentExport;
use crate::grpc_client::EnvironmentGetRequest;
use crate::state::AppState;

/// 获取环境列表
#[tauri::command]
pub async fn list_environments(state: State<'_, AppState>) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .environments
        .list(Empty {})
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(serde_json::to_string(&response.into_inner()).unwrap())
}

/// 获取单个环境
#[tauri::command]
pub async fn get_environment(
    state: State<'_, AppState>,
    id: String,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .environments
        .get(EnvironmentGetRequest { id })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(serde_json::to_string(&response.into_inner()).unwrap())
}

/// 创建环境
#[tauri::command]
pub async fn create_environment(
    state: State<'_, AppState>,
    environment: String,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let request: Environment = serde_json::from_str(&environment)
        .map_err(|e| crate::error::AppError::Connection(format!("JSON 解析失败: {}", e)))?;
    let response = client
        .environments
        .create(request)
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(serde_json::to_string(&response.into_inner()).unwrap())
}

/// 更新环境
#[tauri::command]
pub async fn update_environment(
    state: State<'_, AppState>,
    environment: String,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let request: Environment = serde_json::from_str(&environment)
        .map_err(|e| crate::error::AppError::Connection(format!("JSON 解析失败: {}", e)))?;
    let response = client
        .environments
        .update(request)
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(serde_json::to_string(&response.into_inner()).unwrap())
}

/// 删除环境
#[tauri::command]
pub async fn delete_environment(
    state: State<'_, AppState>,
    id: String,
) -> AppResult<()> {
    let mut client = state.get_grpc_client().await?;
    client
        .environments
        .delete(EnvironmentDeleteRequest { id })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(())
}

/// 激活环境
#[tauri::command]
pub async fn activate_environment(
    state: State<'_, AppState>,
    id: String,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .environments
        .activate(EnvironmentActivateRequest { id })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(serde_json::to_string(&response.into_inner()).unwrap())
}

/// 导出环境
#[tauri::command]
pub async fn export_environment(
    state: State<'_, AppState>,
    id: String,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .environments
        .export(EnvironmentGetRequest { id })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(serde_json::to_string(&response.into_inner()).unwrap())
}

/// 导入环境
#[tauri::command]
pub async fn import_environment(
    state: State<'_, AppState>,
    environment_export: String,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let request: EnvironmentExport = serde_json::from_str(&environment_export)
        .map_err(|e| crate::error::AppError::Connection(format!("JSON 解析失败: {}", e)))?;
    let response = client
        .environments
        .import(request)
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(serde_json::to_string(&response.into_inner()).unwrap())
}
