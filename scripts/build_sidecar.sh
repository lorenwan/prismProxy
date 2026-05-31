#!/bin/bash
# 构建 Go sidecar 用于 Tauri 桌面端
set -e

PROJECT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
TAURI_BIN_DIR="$PROJECT_DIR/desktop/src-tauri/bin"

mkdir -p "$TAURI_BIN_DIR"

# 获取目标平台
PLATFORM="${1:-$(uname -s | tr '[:upper:]' '[:lower:]')}"
ARCH="${2:-$(uname -m)}"

echo "[INFO] 构建 Go sidecar: platform=$PLATFORM arch=$ARCH"

cd "$PROJECT_DIR"

case "$PLATFORM" in
  linux)
    if [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
      GOOS=linux GOARCH=arm64 go build -o "$TAURI_BIN_DIR/prismproxy-server-aarch64-unknown-linux-gnu" ./cmd/server/
    else
      GOOS=linux GOARCH=amd64 go build -o "$TAURI_BIN_DIR/prismproxy-server-x86_64-unknown-linux-gnu" ./cmd/server/
    fi
    ;;
  darwin)
    if [ "$ARCH" = "arm64" ]; then
      GOOS=darwin GOARCH=arm64 go build -o "$TAURI_BIN_DIR/prismproxy-server-aarch64-apple-darwin" ./cmd/server/
    else
      GOOS=darwin GOARCH=amd64 go build -o "$TAURI_BIN_DIR/prismproxy-server-x86_64-apple-darwin" ./cmd/server/
    fi
    ;;
  mingw*|msys*|windows*)
    GOOS=windows GOARCH=amd64 go build -o "$TAURI_BIN_DIR/prismproxy-server-x86_64-pc-windows-msvc.exe" ./cmd/server/
    ;;
  *)
    echo "[ERROR] 不支持的平台: $PLATFORM"
    exit 1
    ;;
esac

echo "[INFO] Go sidecar 构建完成"
ls -lh "$TAURI_BIN_DIR/"
