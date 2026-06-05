use tauri_plugin_shell::process::CommandChild;

pub struct SidecarManager {
    child: Option<CommandChild>,
}

impl SidecarManager {
    pub fn new() -> Self {
        Self { child: None }
    }
}
