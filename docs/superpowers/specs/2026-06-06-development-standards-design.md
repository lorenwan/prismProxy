# PrismProxy 开发规范设计文档

## Context

**问题**：项目缺少统一的开发规范文档，导致：
- 代码风格不一致
- 缺少测试规范
- 新贡献者难以快速上手
- 没有明确的 Git 工作流

**目标**：建立一套完整的开发规范，覆盖代码风格、测试、Git 工作流和项目架构。

**约束**：
- 仅文档规范，不强制代码格式化工具
- 使用 Conventional Commits 提交规范
- 前端测试使用 Vitest
- 多文件组织，按主题拆分

**审查修复**：根据 spec review 反馈，补充错误处理、安全规范、完善示例代码、修正目录结构描述。

---

## 设计概览

### 文档结构

```
docs/development/
├── README.md              # 规范总览和快速参考
├── code-style.md          # 代码风格（前端 + 后端）
├── git-workflow.md        # Git 工作流和提交规范
├── testing.md             # 测试规范（前端 + 后端）
└── architecture.md        # 项目架构
```

### 设计原则

1. **简洁实用** - 规范应易于理解和执行
2. **前后端统一** - 同一文档覆盖前后端规范
3. **渐进式** - 可根据需要扩展
4. **引用现有文档** - 引用 DESIGN.md 和 PRODUCT.md

---

## 文档详细设计

### 1. README.md - 规范总览

**目的**：提供规范的快速参考和使用指南

**内容**：
- 规范目的和适用范围
- 快速参考卡片（常用命令、命名规则）
- 文档目录和使用指南
- 相关资源链接

**字数限制**：500-800 字

---

### 2. code-style.md - 代码风格

**目的**：统一前后端代码风格

**结构**：

```
# 代码风格规范

## 前端规范

### TypeScript
- 类型定义规范
- 泛型使用
- 错误处理
- 空值处理

### React
- 函数组件规范
- Hooks 使用规范
- 状态管理规范
- 组件结构

### TailwindCSS
- 类名顺序
- 自定义样式规范
- 响应式设计

### 命名规范
- 文件命名（PascalCase / kebab-case）
- 变量命名
- 函数命名
- 组件命名

## 后端规范

### Go
- 命名规范（包、函数、变量）
- 错误处理
- 包结构
- 接口设计

### gRPC
- 服务定义规范
- Proto 文件规范
- 错误码设计

### 数据库
- SQL 规范
- 迁移规范
- 索引规范

## 错误处理规范

### 三层架构错误传递
- Go → Rust：gRPC Status Code
- Rust → 前端：字符串错误信息
- 前端：统一错误展示

### 错误类型定义
```typescript
// 前端错误类型
interface AppError {
  code: string
  message: string
  details?: unknown
}
```

## 安全规范

### 证书处理
- CA 证书存储在应用数据目录
- 域名证书按域名分组
- 证书私钥不得日志输出

### 敏感数据
- 代理捕获的认证信息脱敏存储
- API Key 不写入日志
- 请求体中的密码字段掩码显示
```

**关键规范**：

**TypeScript**：
```typescript
// 类型定义
interface TrafficItem {
  id: string
  method: 'GET' | 'POST' | 'PUT' | 'DELETE'
  url: string
  status: number
}

// Result 类型定义
type Result<T, E = Error> =
  | { ok: true; data: T }
  | { ok: false; error: E }

// 错误处理
async function fetchData(): Promise<Result<TrafficItem[], Error>> {
  try {
    const data = await invoke('list_traffic')
    return { ok: true, data }
  } catch (error) {
    return { ok: false, error: new Error(String(error)) }
  }
}
```

**Go**：
```go
// 命名规范
type TrafficService struct {
    repo *storage.TrafficRepo
}

// 错误处理 - 使用 gRPC Status Code
func (s *TrafficService) List(ctx context.Context, req *pb.TrafficListRequest) (*pb.TrafficListResponse, error) {
    items, err := s.repo.List(ctx, req.PageSize)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to list traffic: %v", err)
    }
    return &pb.TrafficListResponse{Items: items}, nil
}
```

**Proto 文件规范**：
```protobuf
// proto/traffic.proto
syntax = "proto3";

package prismproxy;

option go_package = "prismproxy/proto/gen/go";

service TrafficService {
  rpc List(TrafficListRequest) returns (TrafficListResponse);
  rpc Get(TrafficGetRequest) returns (TrafficItem);
  rpc Subscribe(google.protobuf.Empty) returns (stream TrafficEvent);
}

message TrafficItem {
  string id = 1;
  string method = 2;
  string url = 3;
  int32 status_code = 4;
  int64 duration_ms = 5;
}
```

---

### 3. git-workflow.md - Git 工作流

**目的**：规范 Git 使用和协作流程

**内容**：

```
# Git 工作流规范

## 分支策略

### 分支类型
- `main` - 生产分支，始终保持可部署状态
- `feature/*` - 功能分支，从 main 创建
- `fix/*` - 修复分支，从 main 创建
- `release/*` - 发布分支，用于版本准备

### 分支命名
- `feature/traffic-filter` - 功能分支
- `fix/memory-leak` - 修复分支
- `release/v1.0.0` - 发布分支

## 提交规范

### Conventional Commits 格式
```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

### 类型说明
- `feat` - 新功能
- `fix` - Bug 修复
- `docs` - 文档更新
- `style` - 代码格式（不影响功能）
- `refactor` - 重构
- `test` - 测试相关
- `chore` - 构建/工具相关

### 提交示例
```
feat(traffic): 添加流量过滤功能

- 支持按方法过滤
- 支持按状态码过滤
- 支持按 Host 过滤

Closes #123
```

## PR 流程

### PR 模板
```markdown
## 变更说明
[描述本次变更的内容]

## 变更类型
- [ ] 新功能
- [ ] Bug 修复
- [ ] 重构
- [ ] 文档更新

## 测试
- [ ] 单元测试通过
- [ ] 集成测试通过
- [ ] 手动测试通过

## 截图（如适用）
[添加截图]

## 相关 Issue
Closes #[issue_number]
```

## 版本规范

### 语义化版本
```
MAJOR.MINOR.PATCH
```

- **MAJOR** - 不兼容的 API 变更
- **MINOR** - 向后兼容的功能新增
- **PATCH** - 向后兼容的 Bug 修复

### 版本示例
- `1.0.0` - 初始发布
- `1.1.0` - 添加新功能
- `1.1.1` - Bug 修复
- `2.0.0` - 不兼容变更

## CHANGELOG 规范

### 格式
```markdown
# Changelog

## [1.1.0] - 2026-06-06

### Added
- 流量过滤功能
- 支持按方法、状态码、Host 过滤

### Fixed
- 修复内存泄漏问题

### Changed
- 优化列表渲染性能
```

### 维护规则
- 每次发布更新 CHANGELOG
- 使用 Keep a Changelog 格式
- 链接到对应的 PR 或 Issue

### Code Review 清单

#### 代码质量
- [ ] 代码是否符合命名规范
- [ ] 是否有未使用的变量或导入
- [ ] 是否有硬编码的魔法数字
- [ ] 错误处理是否完整

#### 测试
- [ ] 新功能是否有测试覆盖
- [ ] 测试是否覆盖正常和异常路径
- [ ] Mock 是否合理

#### 安全
- [ ] 是否有注入风险（SQL、XSS）
- [ ] 敏感数据是否脱敏
- [ ] 权限检查是否完整

#### 性能
- [ ] 是否有不必要的渲染
- [ ] 是否有内存泄漏风险
- [ ] 数据结构是否合理

#### 可维护性
- [ ] 代码是否易于理解
- [ ] 是否有适当的注释
- [ ] 是否遵循 DRY 原则
```

---

### 4. testing.md - 测试规范

**目的**：统一前后端测试规范

**结构**：

```
# 测试规范

## 测试策略

### 测试金字塔
- 单元测试（70%）- 快速、独立、可重复
- 集成测试（20%）- 验证模块间交互
- E2E 测试（10%）- 验证完整流程

### 测试原则
- 测试应该快速
- 测试应该独立
- 测试应该可重复
- 测试应该自验证

## 前端测试

### Vitest 配置
```typescript
// vitest.config.ts
import { defineConfig } from 'vitest/config'

export default defineConfig({
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.ts'],
  },
})
```

### 组件测试
```typescript
// TrafficList.test.tsx
import { render, screen, fireEvent } from '@testing-library/react'
import { TrafficList } from './TrafficList'

describe('TrafficList', () => {
  it('renders empty state when no data', () => {
    render(<TrafficList items={[]} />)
    expect(screen.getByText('暂无流量数据')).toBeInTheDocument()
  })

  it('selects item on click', () => {
    const onSelect = vi.fn()
    render(<TrafficList items={mockItems} onSelect={onSelect} />)
    fireEvent.click(screen.getByText('GET'))
    expect(onSelect).toHaveBeenCalledWith('1')
  })
})
```

### Store 测试
```typescript
// trafficStore.test.ts
import { useTrafficStore } from './trafficStore'

describe('trafficStore', () => {
  beforeEach(() => {
    useTrafficStore.setState({ items: [], selectedId: null })
  })

  it('adds traffic item', () => {
    const { addTraffic } = useTrafficStore.getState()
    addTraffic(mockItem)
    expect(useTrafficStore.getState().items).toHaveLength(1)
  })
})
```

## 后端测试

### Go 测试规范
```go
// traffic_service_test.go
func TestTrafficService_List(t *testing.T) {
    tests := []struct {
        name    string
        req     *pb.TrafficListRequest
        want    *pb.TrafficListResponse
        wantErr bool
    }{
        {
            name: "success",
            req:  &pb.TrafficListRequest{PageSize: 10},
            want: &pb.TrafficListResponse{Items: mockItems},
        },
        {
            name:    "invalid request",
            req:     &pb.TrafficListRequest{PageSize: -1},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            svc := NewTrafficService(mockRepo)
            got, err := svc.List(context.Background(), tt.req)
            if (err != nil) != tt.wantErr {
                t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("List() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Mock 策略
```go
// mock_traffic_repo.go
type MockTrafficRepo struct {
    items []*TrafficItem
    err   error
}

func (m *MockTrafficRepo) List(ctx context.Context, pageSize int32) ([]*TrafficItem, error) {
    if m.err != nil {
        return nil, m.err
    }
    return m.items, nil
}
```

## 测试覆盖率

### 覆盖率要求
- 新代码单元测试覆盖率：≥ 80%
- 现有代码覆盖率：逐步提升
- 关键路径覆盖率：100%

### 覆盖率检查
```bash
# 前端
npm run test:coverage

# 后端
go test -cover ./...
```

## 测试工具配置

### Vitest 完整配置
```typescript
// vitest.config.ts
import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.ts'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
      exclude: ['node_modules/', 'src/test/'],
    },
  },
})
```

### 测试依赖
```json
{
  "devDependencies": {
    "vitest": "^1.0.0",
    "@testing-library/react": "^14.0.0",
    "@testing-library/jest-dom": "^6.0.0",
    "@testing-library/user-event": "^14.0.0",
    "jsdom": "^22.0.0"
  }
}
```
```

---

### 5. architecture.md - 项目架构

**目的**：描述项目整体架构

**内容**：

```
# 项目架构

## 概述

PrismProxy 采用 **前端 → Rust → Go** 三层架构。

[架构图]

## 前端架构

### 目录结构
[引用 CLAUDE.md 中的前端目录结构]

### 组件规范
- 使用 shadcn/ui 组件库
- 按功能模块组织组件
- 同类组件出现超 2 次就封装复用

### 状态管理
- 使用 Zustand
- 每个功能模块独立 Store
- Store 与 Service 分离

### 路由
- 使用 React Router v7
- 页面组件作为路由入口

## 后端架构

### 目录结构
```
internal/
├── traffic/         # 流量模块
├── rules/           # 规则模块
├── ai/              # AI 模块
├── grpc/            # gRPC 服务实现（14 个服务）
├── storage/         # 存储层（SQLite）
├── proxy/           # HTTP 代理引擎
├── cert/            # 证书管理
├── codegen/         # 代码生成
├── collection/      # API 集合管理
├── debugger/        # 断点调试
├── diff/            # Diff 对比
├── environment/     # 环境变量
├── perf/            # 性能分析
├── rewrite/         # 请求重写
├── script/          # 脚本引擎
├── search/          # 搜索增强
└── websocket/       # WebSocket
```

### 模块化设计
- 每个模块独立目录
- 模块核心逻辑 + gRPC 服务实现
- 部分模块有独立 Store 层

### gRPC 服务
- 14 个服务实现（Traffic、Rules、AI、Breakpoints、Rewrites、Collections、Environments、CodeGen、Scripts、Diff、Perf、Cert、Search、System）
- 统一错误处理（gRPC Status Code）
- 流式数据支持（Server Streaming）

## 通信协议

### Tauri IPC (前端 → Rust)
```typescript
// 前端调用
const result = await invoke('list_traffic', { page: 1 })
```

### gRPC (Rust → Go)
```rust
// Rust 调用
let response = client.list_traffic(request).await?;
```

### Tauri Event (Rust → 前端)
```typescript
// 前端监听
listen('traffic:event', (event) => {
  console.log(event.payload)
})
```

## 设计规范

引用 [DESIGN.md](../design/DESIGN.md)

## 产品说明

引用 [PRODUCT.md](../../PRODUCT.md)
```

---

## 验证方案

### 功能验证

1. **文档完整性** - 检查所有规范是否覆盖
2. **示例代码** - 验证示例代码是否可运行
3. **链接有效性** - 检查所有引用链接

### 质量验证

1. **可读性** - 文档是否易于理解
2. **可执行性** - 规范是否易于执行
3. **一致性** - 文档间是否一致

---

## 实施计划

### 阶段 1：创建文档结构（1 小时）

1. 创建 `docs/development/` 目录
2. 创建 README.md 骨架
3. 创建其他文档骨架

### 阶段 2：编写核心规范（2 小时）

1. 编写 code-style.md
2. 编写 git-workflow.md
3. 编写 testing.md

### 阶段 3：编写架构文档（1 小时）

1. 编写 architecture.md
2. 引用 DESIGN.md 和 PRODUCT.md

### 阶段 4：审查和优化（1 小时）

1. 内部审查
2. 优化文档结构
3. 添加示例代码

---

## 风险和缓解

### 风险 1：规范过于复杂

**问题**：规范文档太多，难以执行
**缓解**：保持文档简洁，提供快速参考

### 风险 2：规范过时

**问题**：项目演进后规范过时
**缓解**：定期审查，建立更新机制

### 风险 3：执行力度不够

**问题**：规范存在但无人遵守
**缓解**：Code Review 检查，PR 模板强制

---

## 总结

本设计文档描述了 PrismProxy 开发规范的完整方案：

**核心原则**：
- 简洁实用
- 前后端统一
- 渐进式扩展
- 引用现有文档

**文档结构**：
- README.md - 规范总览
- code-style.md - 代码风格
- git-workflow.md - Git 工作流
- testing.md - 测试规范
- architecture.md - 项目架构

**下一步**：确认设计文档后，开始编写实现计划。
