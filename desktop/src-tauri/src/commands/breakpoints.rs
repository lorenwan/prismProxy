use tauri::{AppHandle, Emitter, State};

use crate::error::AppResult;
use crate::grpc_client::Breakpoint;
use crate::grpc_client::BreakpointDeleteRequest;
use crate::grpc_client::BreakpointGetRequest;
use crate::grpc_client::BreakpointToggleRequest;
use crate::grpc_client::Empty;
use crate::grpc_client::ResolveSessionRequest;
use crate::state::AppState;

/// 获取断点列表
#[tauri::command]
pub async fn list_breakpoints(state: State<'_, AppState>) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .breakpoints
        .list(Empty {})
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}

/// 获取单条断点
#[tauri::command]
pub async fn get_breakpoint(
    state: State<'_, AppState>,
    id: String,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .breakpoints
        .get(BreakpointGetRequest { id })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}

/// 创建断点
#[tauri::command]
pub async fn create_breakpoint(
    state: State<'_, AppState>,
    breakpoint: String,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let request: Breakpoint = serde_json::from_str(&breakpoint)
        .map_err(|e| crate::error::AppError::Connection(format!("JSON 解析失败: {}", e)))?;
    let response = client
        .breakpoints
        .create(request)
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}

/// 更新断点
#[tauri::command]
pub async fn update_breakpoint(
    state: State<'_, AppState>,
    breakpoint: String,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let request: Breakpoint = serde_json::from_str(&breakpoint)
        .map_err(|e| crate::error::AppError::Connection(format!("JSON 解析失败: {}", e)))?;
    let response = client
        .breakpoints
        .update(request)
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}

/// 删除断点
#[tauri::command]
pub async fn delete_breakpoint(
    state: State<'_, AppState>,
    id: String,
) -> AppResult<()> {
    let mut client = state.get_grpc_client().await?;
    client
        .breakpoints
        .delete(BreakpointDeleteRequest { id })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(())
}

/// 切换断点启用状态
#[tauri::command]
pub async fn toggle_breakpoint(
    state: State<'_, AppState>,
    id: String,
    enabled: bool,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .breakpoints
        .toggle(BreakpointToggleRequest { id, enabled })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}

/// 获取活跃会话列表
#[tauri::command]
pub async fn list_breakpoint_sessions(
    state: State<'_, AppState>,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .breakpoints
        .list_sessions(Empty {})
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}

/// 处理会话（继续/丢弃/修改后继续）
#[tauri::command]
pub async fn resolve_breakpoint_session(
    state: State<'_, AppState>,
    session_id: String,
    action: String,
    modified_data: Option<String>,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let modified = modified_data
        .map(|d| {
            serde_json::from_str(&d)
                .map_err(|e| crate::error::AppError::Connection(format!("JSON 解析失败: {}", e)))
        })
        .transpose()?;
    let response = client
        .breakpoints
        .resolve_session(ResolveSessionRequest {
            session_id,
            action,
            modified_data: modified,
        })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}

/// 订阅断点事件（服务端流式推送）
#[tauri::command]
pub async fn subscribe_breakpoints(
    app: AppHandle,
    state: State<'_, AppState>,
) -> AppResult<()> {
    let mut client = state.get_grpc_client().await?;
    let mut stream = client
        .breakpoints
        .subscribe(Empty {})
        .await
        .map_err(crate::error::AppError::Grpc)?
        .into_inner();

    tokio::spawn(async move {
        while let Ok(Some(event)) = stream.message().await {
            if let Ok(payload) = serde_json::to_string(&event) {
                let _ = app.emit("breakpoints:event", payload);
            }
        }
    });

    Ok(())
}
