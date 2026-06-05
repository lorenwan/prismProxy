use std::sync::Arc;
use tokio::sync::Mutex;
use crate::grpc_client::GrpcClient;
use crate::sidecar::SidecarManager;
use crate::config::ConfigManager;

#[derive(Clone)]
pub struct AppState {
    pub grpc_client: Arc<Mutex<Option<GrpcClient>>>,
    pub sidecar_manager: Arc<Mutex<SidecarManager>>,
    pub config_manager: Arc<Mutex<ConfigManager>>,
}

impl Default for AppState {
    fn default() -> Self {
        Self {
            grpc_client: Arc::new(Mutex::new(None)),
            sidecar_manager: Arc::new(Mutex::new(SidecarManager::new())),
            config_manager: Arc::new(Mutex::new(ConfigManager::new())),
        }
    }
}

impl AppState {
    pub async fn init_grpc_client(&self, addr: &str) -> Result<(), tonic::Status> {
        let client = GrpcClient::new(addr).await?;
        let mut grpc_client = self.grpc_client.lock().await;
        *grpc_client = Some(client);
        Ok(())
    }

    pub async fn get_grpc_client(&self) -> Result<GrpcClient, crate::error::AppError> {
        let grpc_client = self.grpc_client.lock().await;
        grpc_client.clone().ok_or_else(|| {
            crate::error::AppError::Connection("gRPC client not initialized".to_string())
        })
    }
}
