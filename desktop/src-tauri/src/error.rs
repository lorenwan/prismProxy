use serde::Serialize;

#[derive(Debug, thiserror::Error)]
pub enum AppError {
    #[error("gRPC error: {0}")]
    Grpc(#[from] tonic::Status),

    #[error("Connection error: {0}")]
    Connection(String),

    #[error("Sidecar error: {0}")]
    Sidecar(String),

    #[error("Config error: {0}")]
    Config(String),

    #[error("IO error: {0}")]
    Io(#[from] std::io::Error),

    #[error("Serialization error: {0}")]
    Serialize(String),
}

impl AppError {
    /// 将 Result<T, serde_json::Error> 转换为 AppResult<T>
    pub fn from_json_result<T>(result: Result<T, serde_json::Error>) -> AppResult<T> {
        result.map_err(|e| AppError::Serialize(format!("JSON 序列化失败: {}", e)))
    }
}

// 实现Serialize，以便通过IPC返回给前端
impl Serialize for AppError {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        serializer.serialize_str(&self.to_string())
    }
}

// 为Tauri命令返回Result类型
pub type AppResult<T> = Result<T, AppError>;
