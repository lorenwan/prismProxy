use tauri::State;

use crate::error::AppResult;
use crate::grpc_client::Empty;
use crate::grpc_client::ExecuteScriptRequest;
use crate::grpc_client::Script;
use crate::grpc_client::ScriptDeleteRequest;
use crate::grpc_client::ScriptGetRequest;
use crate::grpc_client::ScriptToggleRequest;
use crate::state::AppState;

/// 获取脚本列表
#[tauri::command]
pub async fn list_scripts(state: State<'_, AppState>) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .scripts
        .list(Empty {})
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}

/// 获取单个脚本
#[tauri::command]
pub async fn get_script(
    state: State<'_, AppState>,
    id: String,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .scripts
        .get(ScriptGetRequest { id })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}

/// 创建脚本
#[tauri::command]
pub async fn create_script(
    state: State<'_, AppState>,
    script: String,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let request: Script = serde_json::from_str(&script)
        .map_err(|e| crate::error::AppError::Connection(format!("JSON 解析失败: {}", e)))?;
    let response = client
        .scripts
        .create(request)
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}

/// 更新脚本
#[tauri::command]
pub async fn update_script(
    state: State<'_, AppState>,
    script: String,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let request: Script = serde_json::from_str(&script)
        .map_err(|e| crate::error::AppError::Connection(format!("JSON 解析失败: {}", e)))?;
    let response = client
        .scripts
        .update(request)
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}

/// 删除脚本
#[tauri::command]
pub async fn delete_script(
    state: State<'_, AppState>,
    id: String,
) -> AppResult<()> {
    let mut client = state.get_grpc_client().await?;
    client
        .scripts
        .delete(ScriptDeleteRequest { id })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(())
}

/// 启用/禁用脚本
#[tauri::command]
pub async fn toggle_script(
    state: State<'_, AppState>,
    id: String,
    enabled: Option<bool>,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .scripts
        .toggle(ScriptToggleRequest {
            id,
            enabled: enabled.unwrap_or(false),
        })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}

/// 执行脚本
#[tauri::command]
pub async fn execute_script(
    state: State<'_, AppState>,
    script_id: String,
    transaction_id: String,
    data: Option<std::collections::HashMap<String, String>>,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .scripts
        .execute(ExecuteScriptRequest {
            script_id,
            transaction_id,
            data: data.unwrap_or_default(),
        })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}
