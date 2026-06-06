use tauri::{AppHandle, Emitter, State};

use crate::error::AppResult;
use crate::grpc_client::Empty;
use crate::grpc_client::Pagination;
use crate::grpc_client::TrafficDeleteRequest;
use crate::grpc_client::TrafficGetRequest;
use crate::grpc_client::TrafficListRequest;
use crate::grpc_client::TrafficStatsRequest;
use crate::grpc_client::TrafficFilter;
use crate::grpc_client::TrafficUpdateBookmarkRequest;
use crate::grpc_client::TrafficUpdateNotesRequest;
use crate::grpc_client::TrafficUpdateColorRequest;
use crate::grpc_client::TrafficUpdateTagsRequest;
use crate::state::AppState;

/// 获取流量列表
#[tauri::command]
pub async fn list_traffic(
    state: State<'_, AppState>,
    page: Option<i32>,
    page_size: Option<i32>,
    method: Option<String>,
    status: Option<i32>,
    host: Option<String>,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;

    // 构建过滤器
    let mut filter = TrafficFilter::default();
    if let Some(ref m) = method {
        if !m.is_empty() {
            filter.method = vec![m.clone()];
        }
    }
    if let Some(s) = status {
        if s != 0 {
            filter.status_code = vec![s];
        }
    }
    if let Some(ref h) = host {
        if !h.is_empty() {
            filter.host = vec![h.clone()];
        }
    }

    let request = TrafficListRequest {
        pagination: Some(Pagination {
            page: page.unwrap_or(1),
            page_size: page_size.unwrap_or(20),
            sort_by: String::new(),
            sort_desc: false,
        }),
        filter: Some(filter),
    };
    let response = client
        .traffic
        .list(request)
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
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
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
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

/// 获取流量统计
#[tauri::command]
pub async fn get_traffic_stats(
    state: State<'_, AppState>,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .traffic
        .stats(TrafficStatsRequest { ..Default::default() })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}

/// 更新书签状态
#[tauri::command]
pub async fn update_traffic_bookmark(
    state: State<'_, AppState>,
    id: i64,
    bookmarked: bool,
) -> AppResult<()> {
    let mut client = state.get_grpc_client().await?;
    client
        .traffic
        .update_bookmark(TrafficUpdateBookmarkRequest { id, bookmarked })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(())
}

/// 更新备注
#[tauri::command]
pub async fn update_traffic_notes(
    state: State<'_, AppState>,
    id: i64,
    notes: String,
) -> AppResult<()> {
    let mut client = state.get_grpc_client().await?;
    client
        .traffic
        .update_notes(TrafficUpdateNotesRequest { id, notes })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(())
}

/// 更新颜色标记
#[tauri::command]
pub async fn update_traffic_color(
    state: State<'_, AppState>,
    id: i64,
    color: String,
) -> AppResult<()> {
    let mut client = state.get_grpc_client().await?;
    client
        .traffic
        .update_color(TrafficUpdateColorRequest { id, color })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(())
}

/// 更新标签
#[tauri::command]
pub async fn update_traffic_tags(
    state: State<'_, AppState>,
    id: i64,
    tags: Vec<String>,
) -> AppResult<()> {
    let mut client = state.get_grpc_client().await?;
    client
        .traffic
        .update_tags(TrafficUpdateTagsRequest { id, tags })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(())
}
