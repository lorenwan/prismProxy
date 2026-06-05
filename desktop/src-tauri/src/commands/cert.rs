use tauri::State;

use crate::error::AppResult;
use crate::grpc_client::CertCheckRequest;
use crate::grpc_client::CertDeleteRequest;
use crate::grpc_client::Empty;
use crate::grpc_client::IssueCertRequest;
use crate::state::AppState;

/// 获取 CA 信息
#[tauri::command]
pub async fn get_ca_info(state: State<'_, AppState>) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .cert
        .get_ca_info(Empty {})
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}

/// 生成 CA 证书
#[tauri::command]
pub async fn generate_ca(state: State<'_, AppState>) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .cert
        .generate_ca(Empty {})
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}

/// 导出 CA 证书
#[tauri::command]
pub async fn export_ca(state: State<'_, AppState>) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .cert
        .export_ca(Empty {})
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}

/// 签发域名证书
#[tauri::command]
pub async fn issue_cert(
    state: State<'_, AppState>,
    domain: String,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .cert
        .issue_cert(IssueCertRequest { domain })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}

/// 列出所有域名证书
#[tauri::command]
pub async fn list_certs(state: State<'_, AppState>) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .cert
        .list_certs(Empty {})
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}

/// 删除域名证书
#[tauri::command]
pub async fn delete_cert(
    state: State<'_, AppState>,
    domain: String,
) -> AppResult<()> {
    let mut client = state.get_grpc_client().await?;
    client
        .cert
        .delete_cert(CertDeleteRequest { domain })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(())
}

/// 检查证书状态
#[tauri::command]
pub async fn check_cert(
    state: State<'_, AppState>,
    domain: String,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .cert
        .check_cert(CertCheckRequest { domain })
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}

/// 清除所有证书缓存
#[tauri::command]
pub async fn clear_certs(state: State<'_, AppState>) -> AppResult<()> {
    let mut client = state.get_grpc_client().await?;
    client
        .cert
        .clear_certs(Empty {})
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(())
}

/// 获取 CA 证书信任状态
#[tauri::command]
pub async fn get_trust_status(state: State<'_, AppState>) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .cert
        .get_trust_status(Empty {})
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}
