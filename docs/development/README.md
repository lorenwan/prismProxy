# PrismProxy 开发规范

本目录包含 PrismProxy 项目的开发规范文档。

## 快速参考

### 常用命令

```bash
# 前端开发
cd desktop && npm run dev          # 启动 Tauri 桌面应用
cd desktop && npm run frontend:dev # 仅启动前端
cd desktop && npm run build        # 构建前端

# 后端开发
go build -o prismproxy-server ./cmd/server  # 构建后端
go run ./cmd/server --port 9090             # 运行后端
go test ./...                                # 运行测试
go vet ./...                                 # 代码检查

# Proto 代码生成
bash scripts/gen_proto.sh go
```

### 命名规范

| 类型 | 规范 | 示例 |
|------|------|------|
| 前端文件 | PascalCase / kebab-case | `TrafficList.tsx`, `button.tsx` |
| 前端变量 | camelCase | `trafficList`, `selectedId` |
| 前端组件 | PascalCase | `TrafficList`, `Header` |
| 后端包 | snake_case | `traffic`, `rules` |
| 后端函数 | PascalCase (导出) / camelCase (私有) | `List`, `newService` |
| 后端变量 | camelCase | `pageSize`, `totalCount` |

### 提交规范

```
<type>(<scope>): <description>

feat(traffic): 添加流量过滤功能
fix(auth): 修复登录超时问题
docs(readme): 更新安装说明
```

## 文档目录

### 通用规范

| 文档 | 内容 |
|------|------|
| [git-workflow.md](git-workflow.md) | Git 工作流和提交规范 |
| [architecture.md](architecture.md) | 项目架构说明 |

### 前端规范

| 文档 | 内容 |
|------|------|
| [frontend/code-style.md](frontend/code-style.md) | 前端代码风格规范 |
| [frontend/testing.md](frontend/testing.md) | 前端测试规范 |

### 后端规范

| 文档 | 内容 |
|------|------|
| [backend/code-style.md](backend/code-style.md) | 后端代码风格规范 |
| [backend/testing.md](backend/testing.md) | 后端测试规范 |

## 相关文档

- [DESIGN.md](../design/DESIGN.md) - 设计规范
- [PRODUCT.md](../../PRODUCT.md) - 产品说明
- [CLAUDE.md](../../CLAUDE.md) - Claude Code 指南

## 适用范围

本规范适用于：
- PrismProxy 项目的所有代码贡献
- 前端（React + TypeScript）
- 后端（Go）
- 桌面端（Rust + Tauri）

## 更新日志

| 日期 | 变更 |
|------|------|
| 2026-06-06 | 初始版本 |
