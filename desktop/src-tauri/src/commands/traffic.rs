use tauri::{AppHandle, Emitter, State};

use crate::error::AppResult;
use crate::grpc_client::Empty;
use crate::grpc_client::Pagination;
use crate::grpc_client::TrafficDeleteRequest;
use crate::grpc_client::TrafficGetRequest;
use crate::grpc_client::TrafficListRequest;
use crate::state::AppState;

/// 获取流量列表
#[tauri::command]
pub async fn list_traffic(
    state: State<'_, AppState>,
    page: Option<i32>,
    page_size: Option<i32>,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let request = TrafficListRequest {
        pagination: Some(Pagination {
            page: page.unwrap_or(1),
            page_size: page_size.unwrap_or(20),
            sort_by: String::new(),
            sort_desc: false,
        }),
        ..Default::default()
    };
    let response = client
        .traffic
        .list(request)
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(serde_json::to_string(&response.into_inner()).unwrap())
}

/// 获取单条流量详情
#[tauri::command]
pub async fn get_traffic(
    state: State<'_, AppState>,
    id: i64,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let request = TrafficGetRequest { id };
    let response = client
        .traffic
        .get(request)
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(serde_json::to_string(&response.into_inner()).unwrap())
}

/// 删除流量记录
#[tauri::command]
pub async fn delete_traffic(
    state: State<'_, AppState>,
    ids: Vec<i64>,
) -> AppResult<()> {
    let mut client = state.get_grpc_client().await?;
    let request = TrafficDeleteRequest { ids };
    client
        .traffic
        .delete(request)
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(())
}

/// 清空所有流量
#[tauri::command]
pub async fn clear_traffic(state: State<'_, AppState>) -> AppResult<()> {
    let mut client = state.get_grpc_client().await?;
    client
        .traffic
        .clear(Empty {})
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(())
}

/// 订阅流量事件（服务端流式推送）
#[tauri::command]
pub async fn subscribe_traffic(
    app: AppHandle,
    state: State<'_, AppState>,
) -> AppResult<()> {
    let mut client = state.get_grpc_client().await?;
    let mut stream = client
        .traffic
        .subscribe(Empty {})
        .await
        .map_err(crate::error::AppError::Grpc)?
        .into_inner();

    // 在后台任务中处理流，将gRPC事件转换为Tauri事件
    tokio::spawn(async move {
        while let Ok(Some(event)) = stream.message().await {
            if let Ok(payload) = serde_json::to_string(&event) {
                let _ = app.emit("traffic:event", payload);
            }
        }
    });

    Ok(())
}
