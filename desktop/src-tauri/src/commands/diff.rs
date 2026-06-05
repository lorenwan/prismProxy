use tauri::State;

use crate::error::AppResult;
use crate::grpc_client::CompareBodyRequest;
use crate::grpc_client::CompareHeadersRequest;
use crate::grpc_client::CompareJsonRequest;
use crate::grpc_client::CompareQueryRequest;
use crate::grpc_client::StringList;
use crate::state::AppState;

/// 对比 Headers
#[tauri::command]
pub async fn compare_headers(
    state: State<'_, AppState>,
    left: std::collections::HashMap<String, String>,
    right: std::collections::HashMap<String, String>,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let left_map: std::collections::HashMap<String, StringList> = left
        .into_iter()
        .map(|(k, v)| (k, StringList { values: vec![v] }))
        .collect();
    let right_map: std::collections::HashMap<String, StringList> = right
        .into_iter()
        .map(|(k, v)| (k, StringList { values: vec![v] }))
        .collect();
    let response = client
        .diff
        .compare_headers(CompareHeadersRequest {
            left: left_map,
            right: right_map,
        })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(serde_json::to_string(&response.into_inner()).unwrap())
}

/// 对比 Body
#[tauri::command]
pub async fn compare_body(
    state: State<'_, AppState>,
    left: String,
    right: String,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .diff
        .compare_body(CompareBodyRequest {
            left: left.into_bytes(),
            right: right.into_bytes(),
        })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(serde_json::to_string(&response.into_inner()).unwrap())
}

/// 对比 JSON
#[tauri::command]
pub async fn compare_json(
    state: State<'_, AppState>,
    left: String,
    right: String,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .diff
        .compare_json(CompareJsonRequest { left, right })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(serde_json::to_string(&response.into_inner()).unwrap())
}

/// 对比 Query 参数
#[tauri::command]
pub async fn compare_query(
    state: State<'_, AppState>,
    left: std::collections::HashMap<String, String>,
    right: std::collections::HashMap<String, String>,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let left_map: std::collections::HashMap<String, StringList> = left
        .into_iter()
        .map(|(k, v)| (k, StringList { values: vec![v] }))
        .collect();
    let right_map: std::collections::HashMap<String, StringList> = right
        .into_iter()
        .map(|(k, v)| (k, StringList { values: vec![v] }))
        .collect();
    let response = client
        .diff
        .compare_query(CompareQueryRequest {
            left: left_map,
            right: right_map,
        })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(serde_json::to_string(&response.into_inner()).unwrap())
}
