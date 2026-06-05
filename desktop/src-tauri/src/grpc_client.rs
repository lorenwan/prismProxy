use tonic::transport::Channel;
use tonic::Status;

// 导入生成的proto代码
include!(concat!(env!("OUT_DIR"), "/prismproxy.rs"));

#[derive(Clone)]
pub struct GrpcClient {
    pub traffic: traffic_service_client::TrafficServiceClient<Channel>,
    pub rules: rules_service_client::RulesServiceClient<Channel>,
    pub breakpoints: breakpoints_service_client::BreakpointsServiceClient<Channel>,
    pub rewrites: rewrites_service_client::RewritesServiceClient<Channel>,
    pub collections: collections_service_client::CollectionsServiceClient<Channel>,
    pub environments: environments_service_client::EnvironmentsServiceClient<Channel>,
    pub ai: ai_service_client::AiServiceClient<Channel>,
    pub system: system_service_client::SystemServiceClient<Channel>,
    pub codegen: code_gen_service_client::CodeGenServiceClient<Channel>,
    pub scripts: scripts_service_client::ScriptsServiceClient<Channel>,
    pub diff: diff_service_client::DiffServiceClient<Channel>,
    pub perf: perf_service_client::PerfServiceClient<Channel>,
    pub cert: cert_service_client::CertServiceClient<Channel>,
    pub search: search_service_client::SearchServiceClient<Channel>,
}

impl GrpcClient {
    pub async fn new(addr: &str) -> Result<Self, Status> {
        let channel = Channel::from_shared(addr.to_string())
            .map_err(|e| Status::internal(format!("Failed to create channel: {}", e)))?
            .connect()
            .await
            .map_err(|e| Status::internal(format!("Failed to connect: {}", e)))?;

        Ok(Self {
            traffic: traffic_service_client::TrafficServiceClient::new(channel.clone()),
            rules: rules_service_client::RulesServiceClient::new(channel.clone()),
            breakpoints: breakpoints_service_client::BreakpointsServiceClient::new(channel.clone()),
            rewrites: rewrites_service_client::RewritesServiceClient::new(channel.clone()),
            collections: collections_service_client::CollectionsServiceClient::new(channel.clone()),
            environments: environments_service_client::EnvironmentsServiceClient::new(channel.clone()),
            ai: ai_service_client::AiServiceClient::new(channel.clone()),
            system: system_service_client::SystemServiceClient::new(channel.clone()),
            codegen: code_gen_service_client::CodeGenServiceClient::new(channel.clone()),
            scripts: scripts_service_client::ScriptsServiceClient::new(channel.clone()),
            diff: diff_service_client::DiffServiceClient::new(channel.clone()),
            perf: perf_service_client::PerfServiceClient::new(channel.clone()),
            cert: cert_service_client::CertServiceClient::new(channel.clone()),
            search: search_service_client::SearchServiceClient::new(channel),
        })
    }
}
