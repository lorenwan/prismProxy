fn main() {
    // 编译proto文件，生成Rust客户端代码
    tonic_build::configure()
        .build_server(false) // 只生成客户端，不生成服务端
        .build_client(true)
        .compile(
            &[
                "../../proto/common.proto",
                "../../proto/traffic.proto",
                "../../proto/rules.proto",
                "../../proto/breakpoints.proto",
                "../../proto/rewrites.proto",
                "../../proto/collections.proto",
                "../../proto/environments.proto",
                "../../proto/ai.proto",
                "../../proto/system.proto",
                "../../proto/codegen.proto",
                "../../proto/scripts.proto",
                "../../proto/diff.proto",
                "../../proto/perf.proto",
                "../../proto/cert.proto",
                "../../proto/search.proto",
            ],
            &["../../proto"],
        )
        .unwrap();

    tauri_build::run()
}
