use tauri::State;

use crate::error::AppResult;
use crate::grpc_client::DeleteFilterRequest;
use crate::grpc_client::Empty;
use crate::grpc_client::Pagination;
use crate::grpc_client::SavedFilter;
use crate::grpc_client::SearchByHostRequest;
use crate::grpc_client::SearchByMethodRequest;
use crate::grpc_client::SearchByStatusCodeRequest;
use crate::grpc_client::SearchFilter;
use crate::grpc_client::SearchRequest;
use crate::grpc_client::SearchSlowRequestsRequest;
use crate::state::AppState;

/// 全文搜索
#[tauri::command]
pub async fn search(
    state: State<'_, AppState>,
    query: String,
    sort: Option<String>,
    page: Option<i32>,
    page_size: Option<i32>,
    filters: Option<Vec<serde_json::Value>>,
) -> AppResult<String> {
    let search_filters: Vec<SearchFilter> = filters
        .unwrap_or_default()
        .into_iter()
        .filter_map(|f| {
            let field = f.get("field")?.as_str()?.to_string();
            let operator = f.get("operator")?.as_i64().unwrap_or(0) as i32;
            let value = f.get("value")?.as_str()?.to_string();
            Some(SearchFilter {
                field,
                operator,
                value,
            })
        })
        .collect();

    let mut client = state.get_grpc_client().await?;
    let response = client
        .search
        .search(SearchRequest {
            query,
            filters: search_filters,
            sort: sort.unwrap_or_default(),
            pagination: Some(Pagination {
                page: page.unwrap_or(1),
                page_size: page_size.unwrap_or(20),
                sort_by: String::new(),
                sort_desc: false,
            }),
        })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}

/// 按方法搜索
#[tauri::command]
pub async fn search_by_method(
    state: State<'_, AppState>,
    method: String,
    page: Option<i32>,
    page_size: Option<i32>,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .search
        .search_by_method(SearchByMethodRequest {
            method,
            pagination: Some(Pagination {
                page: page.unwrap_or(1),
                page_size: page_size.unwrap_or(20),
                sort_by: String::new(),
                sort_desc: false,
            }),
        })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}

/// 按主机搜索
#[tauri::command]
pub async fn search_by_host(
    state: State<'_, AppState>,
    host: String,
    page: Option<i32>,
    page_size: Option<i32>,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .search
        .search_by_host(SearchByHostRequest {
            host,
            pagination: Some(Pagination {
                page: page.unwrap_or(1),
                page_size: page_size.unwrap_or(20),
                sort_by: String::new(),
                sort_desc: false,
            }),
        })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}

/// 按状态码搜索
#[tauri::command]
pub async fn search_by_status_code(
    state: State<'_, AppState>,
    status_code: i32,
    page: Option<i32>,
    page_size: Option<i32>,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .search
        .search_by_status_code(SearchByStatusCodeRequest {
            status_code,
            pagination: Some(Pagination {
                page: page.unwrap_or(1),
                page_size: page_size.unwrap_or(20),
                sort_by: String::new(),
                sort_desc: false,
            }),
        })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}

/// 搜索慢请求
#[tauri::command]
pub async fn search_slow_requests(
    state: State<'_, AppState>,
    threshold_ms: Option<i64>,
    page: Option<i32>,
    page_size: Option<i32>,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .search
        .search_slow_requests(SearchSlowRequestsRequest {
            threshold_ms: threshold_ms.unwrap_or(1000),
            pagination: Some(Pagination {
                page: page.unwrap_or(1),
                page_size: page_size.unwrap_or(20),
                sort_by: String::new(),
                sort_desc: false,
            }),
        })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}

/// 获取搜索统计
#[tauri::command]
pub async fn get_search_stats(
    state: State<'_, AppState>,
    query: String,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .search
        .get_search_stats(SearchRequest {
            query,
            filters: vec![],
            sort: String::new(),
            pagination: None,
        })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}

/// 保存过滤器
#[tauri::command]
pub async fn save_filter(
    state: State<'_, AppState>,
    filter: String,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let request: SavedFilter = serde_json::from_str(&filter)
        .map_err(|e| crate::error::AppError::Serialize(format!("JSON 解析失败: {}", e)))?;
    let response = client
        .search
        .save_filter(request)
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}

/// 获取保存的过滤器列表
#[tauri::command]
pub async fn list_filters(state: State<'_, AppState>) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .search
        .list_filters(Empty {})
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}

/// 删除过滤器
#[tauri::command]
pub async fn delete_filter(
    state: State<'_, AppState>,
    id: String,
) -> AppResult<()> {
    let mut client = state.get_grpc_client().await?;
    client
        .search
        .delete_filter(DeleteFilterRequest { id })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(())
}
