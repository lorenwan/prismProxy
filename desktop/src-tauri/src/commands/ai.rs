use tauri::{AppHandle, Emitter, State};

use crate::error::AppResult;
use crate::grpc_client::ChatRequest;
use crate::grpc_client::ChatMessage;
use crate::grpc_client::Empty;
use crate::state::AppState;

/// 非流式聊天
#[tauri::command]
pub async fn chat(
    state: State<'_, AppState>,
    messages: Vec<ChatMessage>,
    model: Option<String>,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let request = ChatRequest {
        messages,
        model: model.unwrap_or_default(),
        stream: false,
    };
    let response = client
        .ai
        .chat(request)
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(serde_json::to_string(&response.into_inner()).unwrap())
}

/// 流式聊天（通过Tauri事件推送流式响应）
#[tauri::command]
pub async fn stream_chat(
    app: AppHandle,
    state: State<'_, AppState>,
    messages: Vec<ChatMessage>,
    model: Option<String>,
) -> AppResult<()> {
    let mut client = state.get_grpc_client().await?;
    let request = ChatRequest {
        messages,
        model: model.unwrap_or_default(),
        stream: true,
    };
    let mut stream = client
        .ai
        .stream_chat(request)
        .await
        .map_err(crate::error::AppError::Grpc)?
        .into_inner();

    // 在后台任务中处理流，将gRPC流转换为Tauri事件
    tokio::spawn(async move {
        while let Ok(Some(chunk)) = stream.message().await {
            if let Ok(payload) = serde_json::to_string(&chunk) {
                let _ = app.emit("ai:chat_chunk", payload);
            }
        }
        // 流结束
        let _ = app.emit("ai:chat_end", ());
    });

    Ok(())
}

/// 检查AI服务可用性
#[tauri::command]
pub async fn check_ai_availability(
    state: State<'_, AppState>,
) -> AppResult<String> {
    let mut client = state.get_grpc_client().await?;
    let response = client
        .ai
        .check_availability(Empty {})
        .await
        .map_err(crate::error::AppError::Grpc)?;
    Ok(serde_json::to_string(&response.into_inner()).unwrap())
}
