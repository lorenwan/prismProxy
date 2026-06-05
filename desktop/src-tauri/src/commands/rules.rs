use tauri::State;

use crate::error::AppResult;
use crate::grpc_client::Empty;
use crate::grpc_client::Rule;
use crate::grpc_client::RuleDeleteRequest;
use crate::grpc_client::RuleGetRequest;
use crate::state::AppState;

/// 获取规则列表
#[tauri::command]
pub async fn list_rules(state: State<'_, AppState>) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .rules
        .list(Empty {})
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}

/// 获取单条规则
#[tauri::command]
pub async fn get_rule(state: State<'_, AppState>, id: String) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .rules
        .get(RuleGetRequest { id })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}

/// 创建规则
#[tauri::command]
pub async fn create_rule(state: State<'_, AppState>, rule: String) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let request: Rule = serde_json::from_str(&rule)
        .map_err(|e| crate::error::AppError::Connection(format!("JSON 解析失败: {}", e)))?;
    let response = client
        .rules
        .create(request)
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}

/// 更新规则
#[tauri::command]
pub async fn update_rule(state: State<'_, AppState>, rule: String) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let request: Rule = serde_json::from_str(&rule)
        .map_err(|e| crate::error::AppError::Connection(format!("JSON 解析失败: {}", e)))?;
    let response = client
        .rules
        .update(request)
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}

/// 删除规则
#[tauri::command]
pub async fn delete_rule(state: State<'_, AppState>, id: String) -> AppResult<()> {
    let mut client = state.get_grpc_client().await?;
    client
        .rules
        .delete(RuleDeleteRequest { id })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(())
}
