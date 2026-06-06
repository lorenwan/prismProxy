# 开发规范实现计划

## 目标

创建 `docs/development/` 目录下的开发规范文档，包括：
- README.md - 规范总览
- code-style.md - 代码风格（前端 + 后端）
- git-workflow.md - Git 工作流
- testing.md - 测试规范（前端 + 后端）
- architecture.md - 项目架构

## 实施步骤

### 步骤 1：创建目录结构

```bash
mkdir -p docs/development
```

### 步骤 2：创建 README.md

**文件**：`docs/development/README.md`

**内容**：
- 规范目的和适用范围
- 快速参考卡片（常用命令、命名规则）
- 文档目录和使用指南

### 步骤 3：创建 code-style.md

**文件**：`docs/development/code-style.md`

**内容**：
- 前端规范（TypeScript、React、TailwindCSS、命名）
- 后端规范（Go、gRPC、数据库）
- 错误处理规范
- 安全规范

### 步骤 4：创建 git-workflow.md

**文件**：`docs/development/git-workflow.md`

**内容**：
- 分支策略
- Conventional Commits 规范
- PR 流程和模板
- Code Review 清单
- 版本规范
- CHANGELOG 规范

### 步骤 5：创建 testing.md

**文件**：`docs/development/testing.md`

**内容**：
- 测试策略（测试金字塔）
- 前端测试（Vitest 配置、组件测试、Store 测试）
- 后端测试（Go testing、表驱动测试、Mock）
- 测试覆盖率要求
- 测试工具配置

### 步骤 6：创建 architecture.md

**文件**：`docs/development/architecture.md`

**内容**：
- 整体架构（前端 → Rust → Go）
- 前端架构（目录结构、组件规范、状态管理）
- 后端架构（模块化设计、gRPC 服务）
- 通信协议（Tauri IPC、gRPC、事件流）
- 引用 DESIGN.md 和 PRODUCT.md

## 验证清单

- [ ] 所有文档创建完成
- [ ] 文档内容与设计文档一致
- [ ] 示例代码准确可用
- [ ] 链接引用正确
- [ ] CLAUDE.md 更新文档路径

## 预计时间

- 步骤 1：5 分钟
- 步骤 2：15 分钟
- 步骤 3：30 分钟
- 步骤 4：20 分钟
- 步骤 5：30 分钟
- 步骤 6：20 分钟
- 验证：10 分钟

**总计**：约 2 小时
