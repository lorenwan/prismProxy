# Git 工作流规范

## 分支策略

### 分支类型

| 分支 | 用途 | 命名示例 |
|------|------|----------|
| `main` | 生产分支，始终保持可部署 | - |
| `feature/*` | 新功能开发 | `feature/traffic-filter` |
| `fix/*` | Bug 修复 | `fix/memory-leak` |
| `release/*` | 版本发布准备 | `release/v1.0.0` |
| `hotfix/*` | 紧急修复 | `hotfix/critical-crash` |

### 分支工作流

1. 从 `main` 创建功能分支
2. 在功能分支上开发
3. 提交 PR 并通过 Code Review
4. 合并回 `main`

---

## 提交规范

### Conventional Commits 格式

```
<type>(<scope>): <description>
```

### 类型说明

| 类型 | 说明 | 示例 |
|------|------|------|
| `feat` | 新功能 | `feat(traffic): 添加流量过滤` |
| `fix` | Bug 修复 | `fix(auth): 修复登录超时` |
| `docs` | 文档更新 | `docs(readme): 更新安装说明` |
| `style` | 代码格式（不影响功能） | `style: 格式化代码` |
| `refactor` | 重构 | `refactor(traffic): 提取工具函数` |
| `test` | 测试相关 | `test(traffic): 添加单元测试` |
| `chore` | 构建/工具相关 | `chore: 更新依赖` |
| `perf` | 性能优化 | `perf(list): 优化渲染性能` |

### Scope 说明

- 模块名：`traffic`, `rules`, `ai`, `proxy`
- 层级：`frontend`, `backend`, `rust`
- 工具：`deps`, `ci`, `build`

---

## PR 流程

### PR 模板

- 变更说明：描述本次变更的内容和动机
- 变更类型：新功能 / Bug 修复 / 重构 / 文档更新 / 性能优化
- 测试：单元测试 / 集成测试 / 手动测试
- 截图：如适用
- 相关 Issue：关联 Issue 编号

### PR 规范

1. **标题**：使用 Conventional Commits 格式
2. **描述**：详细说明变更内容和动机
3. **大小**：保持 PR 小而专注，便于 Review
4. **关联**：关联相关 Issue

---

## Code Review 清单

### 代码质量

- 代码是否符合命名规范
- 是否有未使用的变量或导入
- 是否有硬编码的魔法数字
- 错误处理是否完整

### 测试

- 新功能是否有测试覆盖
- 测试是否覆盖正常和异常路径
- Mock 是否合理

### 安全

- 是否有注入风险（SQL、XSS）
- 敏感数据是否脱敏
- 权限检查是否完整

### 性能

- 是否有不必要的渲染
- 是否有内存泄漏风险
- 数据结构是否合理

### 可维护性

- 代码是否易于理解
- 是否有适当的注释
- 是否遵循 DRY 原则

---

## 版本规范

### 语义化版本

```
MAJOR.MINOR.PATCH
```

| 类型 | 说明 | 示例 |
|------|------|------|
| MAJOR | 不兼容的 API 变更 | `2.0.0` |
| MINOR | 向后兼容的功能新增 | `1.1.0` |
| PATCH | 向后兼容的 Bug 修复 | `1.0.1` |

---

## CHANGELOG 规范

### 格式

- 使用 [Keep a Changelog](https://keepachangelog.com/) 格式
- 按 Added / Changed / Deprecated / Removed / Fixed / Security 分类
- 每次发布更新 CHANGELOG
- 链接到对应的 PR 或 Issue
