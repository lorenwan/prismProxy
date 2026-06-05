use tauri::State;

use crate::error::AppResult;
use crate::grpc_client::AddRequestRequest;
use crate::grpc_client::Collection;
use crate::grpc_client::CollectionDeleteRequest;
use crate::grpc_client::CollectionGetRequest;
use crate::grpc_client::DeleteRequestRequest;
use crate::grpc_client::Empty;
use crate::grpc_client::ExecuteRequestRequest;
use crate::grpc_client::UpdateRequestRequest;
use crate::state::AppState;

/// 获取集合列表
#[tauri::command]
pub async fn list_collections(state: State<'_, AppState>) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .collections
        .list(Empty {})
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(serde_json::to_string(&response.into_inner()).unwrap())
}

/// 获取单个集合
#[tauri::command]
pub async fn get_collection(
    state: State<'_, AppState>,
    id: String,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .collections
        .get(CollectionGetRequest { id })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(serde_json::to_string(&response.into_inner()).unwrap())
}

/// 创建集合
#[tauri::command]
pub async fn create_collection(
    state: State<'_, AppState>,
    collection: String,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let request: Collection = serde_json::from_str(&collection)
        .map_err(|e| crate::error::AppError::Connection(format!("JSON 解析失败: {}", e)))?;
    let response = client
        .collections
        .create(request)
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(serde_json::to_string(&response.into_inner()).unwrap())
}

/// 更新集合
#[tauri::command]
pub async fn update_collection(
    state: State<'_, AppState>,
    collection: String,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let request: Collection = serde_json::from_str(&collection)
        .map_err(|e| crate::error::AppError::Connection(format!("JSON 解析失败: {}", e)))?;
    let response = client
        .collections
        .update(request)
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(serde_json::to_string(&response.into_inner()).unwrap())
}

/// 删除集合
#[tauri::command]
pub async fn delete_collection(
    state: State<'_, AppState>,
    id: String,
) -> AppResult<()> {
    let mut client = state.get_grpc_client().await?;
    client
        .collections
        .delete(CollectionDeleteRequest { id })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(())
}

/// 添加请求到集合
#[tauri::command]
pub async fn add_collection_request(
    state: State<'_, AppState>,
    collection_id: String,
    parent_item_id: Option<String>,
    request: String,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let api_request: crate::grpc_client::ApiRequest = serde_json::from_str(&request)
        .map_err(|e| crate::error::AppError::Connection(format!("JSON 解析失败: {}", e)))?;
    let response = client
        .collections
        .add_request(AddRequestRequest {
            collection_id,
            parent_item_id: parent_item_id.unwrap_or_default(),
            request: Some(api_request),
        })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(serde_json::to_string(&response.into_inner()).unwrap())
}

/// 更新集合中的请求
#[tauri::command]
pub async fn update_collection_request(
    state: State<'_, AppState>,
    collection_id: String,
    item_id: String,
    request: String,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let api_request: crate::grpc_client::ApiRequest = serde_json::from_str(&request)
        .map_err(|e| crate::error::AppError::Connection(format!("JSON 解析失败: {}", e)))?;
    let response = client
        .collections
        .update_request(UpdateRequestRequest {
            collection_id,
            item_id,
            request: Some(api_request),
        })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(serde_json::to_string(&response.into_inner()).unwrap())
}

/// 删除集合中的请求
#[tauri::command]
pub async fn delete_collection_request(
    state: State<'_, AppState>,
    collection_id: String,
    item_id: String,
) -> AppResult<()> {
    let mut client = state.get_grpc_client().await?;
    client
        .collections
        .delete_request(DeleteRequestRequest {
            collection_id,
            item_id,
        })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(())
}

/// 执行集合中的请求
#[tauri::command]
pub async fn execute_collection_request(
    state: State<'_, AppState>,
    request: String,
    environment_id: Option<String>,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let api_request: crate::grpc_client::ApiRequest = serde_json::from_str(&request)
        .map_err(|e| crate::error::AppError::Connection(format!("JSON 解析失败: {}", e)))?;
    let response = client
        .collections
        .execute_request(ExecuteRequestRequest {
            request: Some(api_request),
            environment_id: environment_id.unwrap_or_default(),
        })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(serde_json::to_string(&response.into_inner()).unwrap())
}
