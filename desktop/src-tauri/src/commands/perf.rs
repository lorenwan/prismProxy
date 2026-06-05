use tauri::State;

use crate::error::AppResult;
use crate::grpc_client::Empty;
use crate::grpc_client::GetRecentStatsRequest;
use crate::grpc_client::GetSlowRequestsRequest;
use crate::grpc_client::GetTimelineRequest;
use crate::grpc_client::PerfStatsRequest;
use crate::state::AppState;

/// 获取性能统计
#[tauri::command]
pub async fn get_perf_stats(
    state: State<'_, AppState>,
    since: Option<String>,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .perf
        .get_stats(PerfStatsRequest {
            since: since.unwrap_or_default(),
        })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(serde_json::to_string(&response.into_inner()).unwrap())
}

/// 获取慢请求
#[tauri::command]
pub async fn get_slow_requests(
    state: State<'_, AppState>,
    threshold_ms: Option<i64>,
    limit: Option<i32>,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .perf
        .get_slow_requests(GetSlowRequestsRequest {
            threshold_ms: threshold_ms.unwrap_or(1000),
            limit: limit.unwrap_or(50),
        })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(serde_json::to_string(&response.into_inner()).unwrap())
}

/// 获取域名统计
#[tauri::command]
pub async fn get_domain_stats(
    state: State<'_, AppState>,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .perf
        .get_domain_stats(Empty {})
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(serde_json::to_string(&response.into_inner()).unwrap())
}

/// 获取时间线数据
#[tauri::command]
pub async fn get_perf_timeline(
    state: State<'_, AppState>,
    since: Option<String>,
    interval_seconds: Option<i64>,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .perf
        .get_timeline(GetTimelineRequest {
            since: since.unwrap_or_default(),
            interval_seconds: interval_seconds.unwrap_or(60),
        })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(serde_json::to_string(&response.into_inner()).unwrap())
}

/// 获取状态码统计
#[tauri::command]
pub async fn get_status_code_stats(
    state: State<'_, AppState>,
    since: Option<String>,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .perf
        .get_status_code_stats(PerfStatsRequest {
            since: since.unwrap_or_default(),
        })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(serde_json::to_string(&response.into_inner()).unwrap())
}

/// 获取请求方法统计
#[tauri::command]
pub async fn get_method_stats(
    state: State<'_, AppState>,
    since: Option<String>,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .perf
        .get_method_stats(PerfStatsRequest {
            since: since.unwrap_or_default(),
        })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(serde_json::to_string(&response.into_inner()).unwrap())
}

/// 获取最近 N 分钟统计
#[tauri::command]
pub async fn get_recent_stats(
    state: State<'_, AppState>,
    minutes: i32,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .perf
        .get_recent_stats(GetRecentStatsRequest { minutes })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(serde_json::to_string(&response.into_inner()).unwrap())
}
