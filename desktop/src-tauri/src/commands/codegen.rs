use tauri::State;

use crate::error::AppResult;
use crate::grpc_client::CodeGenRequest;
use crate::grpc_client::Empty;
use crate::state::AppState;

/// 生成代码
#[tauri::command]
pub async fn generate_code(
    state: State<'_, AppState>,
    request: String,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let code_request: CodeGenRequest = serde_json::from_str(&request)
        .map_err(|e| crate::error::AppError::Connection(format!("JSON 解析失败: {}", e)))?;
    let response = client
        .codegen
        .generate(code_request)
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}

/// 获取支持的语言列表
#[tauri::command]
pub async fn list_codegen_languages(
    state: State<'_, AppState>,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .codegen
        .list_languages(Empty {})
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(crate::error::AppError::from_json_result(serde_json::to_string(&response.into_inner()))?)
}
