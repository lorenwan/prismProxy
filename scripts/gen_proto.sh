#!/bin/bash
# 生成 Go 和 TypeScript 的 Protobuf 代码
# 依赖: protoc, protoc-gen-go, protoc-gen-go-grpc, protoc-gen-ts

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
PROTO_DIR="$PROJECT_ROOT/proto"
GO_OUT="$PROJECT_ROOT/proto/gen/go"
TS_OUT="$PROJECT_ROOT/proto/gen/ts"

# Proto 文件列表（按依赖顺序）
PROTO_FILES=(
  common.proto
  traffic.proto
  rules.proto
  breakpoints.proto
  rewrites.proto
  collections.proto
  environments.proto
  ai.proto
  system.proto
  codegen.proto
  scripts.proto
  diff.proto
  perf.proto
  cert.proto
  search.proto
)

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info()  { echo -e "${GREEN}[INFO]${NC} $1"; }
warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; }

# 检查依赖
check_deps() {
  local missing=0

  if ! command -v protoc &>/dev/null; then
    error "protoc 未安装。请安装 Protocol Buffers 编译器。"
    missing=1
  fi

  if ! command -v protoc-gen-go &>/dev/null; then
    warn "protoc-gen-go 未安装，将跳过 Go 代码生成。"
    warn "安装: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest"
  fi

  if ! command -v protoc-gen-go-grpc &>/dev/null; then
    warn "protoc-gen-go-grpc 未安装，将跳过 Go gRPC 代码生成。"
    warn "安装: go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"
  fi

  if ! command -v protoc-gen-ts &>/dev/null && ! command -v protoc-gen-es &>/dev/null; then
    warn "protoc-gen-ts / protoc-gen-es 未安装，将跳过 TypeScript 代码生成。"
  fi

  if [ $missing -eq 1 ]; then
    exit 1
  fi
}

# 创建输出目录
mkdir -p "$GO_OUT" "$TS_OUT"

# 生成 Go 代码
gen_go() {
  if ! command -v protoc-gen-go &>/dev/null; then
    warn "跳过 Go 代码生成 (protoc-gen-go 未安装)"
    return
  fi

  info "生成 Go 代码..."

  local grpc_opt=""
  if command -v protoc-gen-go-grpc &>/dev/null; then
    grpc_opt="--go-grpc_out=$GO_OUT --go-grpc_opt=paths=source_relative"
  fi

  for proto in "${PROTO_FILES[@]}"; do
    info "  编译 $proto"
    protoc \
      --proto_path="$PROTO_DIR" \
      --go_out="$GO_OUT" \
      --go_opt=paths=source_relative \
      $grpc_opt \
      "$PROTO_DIR/$proto"
  done

  info "Go 代码生成完成: $GO_OUT"
}

# 生成 TypeScript 代码 (使用 @connectrpc/protoc-gen-es 或 protoc-gen-ts)
gen_ts() {
  local ts_plugin=""
  local ts_out_opt=""

  if command -v protoc-gen-es &>/dev/null; then
    ts_plugin="protoc-gen-es"
    ts_out_opt="--es_out=$TS_OUT --es_opt=target=ts"
  elif command -v protoc-gen-ts &>/dev/null; then
    ts_plugin="protoc-gen-ts"
    ts_out_opt="--ts_out=$TS_OUT"
  else
    warn "跳过 TypeScript 代码生成 (protoc-gen-es / protoc-gen-ts 未安装)"
    warn "安装: npm install -g @connectrpc/protoc-gen-es"
    return
  fi

  info "生成 TypeScript 代码 (使用 $ts_plugin)..."

  for proto in "${PROTO_FILES[@]}"; do
    info "  编译 $proto"
    protoc \
      --proto_path="$PROTO_DIR" \
      $ts_out_opt \
      "$PROTO_DIR/$proto"
  done

  info "TypeScript 代码生成完成: $TS_OUT"
}

# 验证 proto 文件
validate() {
  info "验证 Proto 文件..."
  for proto in "${PROTO_FILES[@]}"; do
    if ! protoc --proto_path="$PROTO_DIR" /dev/null "$PROTO_DIR/$proto" 2>/dev/null; then
      # protoc 不能直接验证，尝试用 --descriptor_set_out
      protoc --proto_path="$PROTO_DIR" --descriptor_set_out=/dev/null "$PROTO_DIR/$proto" 2>/dev/null && \
        info "  $proto ✓" || error "  $proto ✗"
    fi
  done
}

# 清理生成文件
clean() {
  info "清理生成文件..."
  rm -rf "$GO_OUT"/*.go "$GO_OUT"/*.pb.go
  rm -rf "$TS_OUT"/*.ts
  info "清理完成"
}

# 使用说明
usage() {
  echo "用法: $0 [命令]"
  echo ""
  echo "命令:"
  echo "  all      生成所有代码 (默认)"
  echo "  go       仅生成 Go 代码"
  echo "  ts       仅生成 TypeScript 代码"
  echo "  validate 验证 Proto 文件"
  echo "  clean    清理生成文件"
  echo "  deps     安装 Go 依赖"
  echo "  help     显示此帮助"
}

# 安装 Go protobuf 依赖
install_deps() {
  info "安装 Go protobuf 依赖..."
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
  info "安装完成"
}

# 主入口
main() {
  local cmd="${1:-all}"

  case "$cmd" in
    all)
      check_deps
      gen_go
      gen_ts
      ;;
    go)
      check_deps
      gen_go
      ;;
    ts)
      check_deps
      gen_ts
      ;;
    validate)
      check_deps
      validate
      ;;
    clean)
      clean
      ;;
    deps)
      install_deps
      ;;
    help|--help|-h)
      usage
      ;;
    *)
      error "未知命令: $cmd"
      usage
      exit 1
      ;;
  esac
}

main "$@"
