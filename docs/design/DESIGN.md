# PrismProxy 详细设计文档

> 对标 Reqable 的 HTTP 调试工具，集成 AI 能力

## 技术栈
- **后端**: Go + Gin + SQLite + WebSocket
- **前端**: React + TypeScript + TailwindCSS
- **桌面端**: Tauri
- **AI**: OpenAI/Claude/Ollama

---

## 模块总览与进度

| 模块 | 状态 | 说明 |
|------|------|------|
| 代理引擎 (proxy/) | ✅ | HTTP/HTTPS MITM |
| 存储层 (storage/) | ✅ | SQLite + 迁移 |
| REST API (api/) | ✅ | 基础 CRUD |
| 流量管理 (traffic/) | ✅ | 数据模型、过滤、统计 |
| 规则引擎 (rules/) | ✅ | 匹配、动作、存储 |
| WebSocket (websocket/) | ✅ | 实时推送 |
| AI 服务 (ai/) | ✅ | Provider、分析、安全、测试 |
| 前端 UI (web/) | ✅ | 流量列表、详情、布局 |
| **断点调试器** | ❌ | Phase 2 核心缺失 |
| **请求重写 (Rewrite)** | ❌ | Phase 2 核心缺失 |
| **API 集合管理** | ❌ | Phase 3 |
| **环境变量** | ❌ | Phase 3 |
| **脚本引擎** | ❌ | Phase 4 |
| **代码生成** | ❌ | Phase 4 |
| **Diff 对比** | ❌ | Phase 4 |

---

## 模块 1: 断点调试器 (Breakpoint Debugger) — 待实现

### 目录结构
```
internal/debugger/
├── debugger.go        # 调试器主控
├── breakpoint.go      # 断点匹配和触发
├── session.go         # 断点会话管理
└── models.go          # 数据模型
```

### 数据模型
```go
type Breakpoint struct {
    ID        string      `json:"id"`
    Enabled   bool        `json:"enabled"`
    Phase     Phase       `json:"phase"`   // request / response
    Match     RuleMatch   `json:"match"`   // 复用 rules 匹配
    Action    BreakAction `json:"action"`
    HitCount  int         `json:"hit_count"`
}

type BreakAction struct {
    Type string `json:"type"` // pause, auto_modify, drop
    Modifications *ModifySpec `json:"modifications,omitempty"`
}

type BreakpointSession struct {
    ID            string       `json:"id"`
    BreakpointID  string       `json:"breakpoint_id"`
    TransactionID int64        `json:"transaction_id"`
    Phase         Phase        `json:"phase"`
    Status        string       `json:"status"` // paused, modified, released, dropped
    Original      *Transaction `json:"original"`
    Modified      *Transaction `json:"modified,omitempty"`
    CreatedAt     time.Time    `json:"created_at"`
    ResolvedAt    *time.Time   `json:"resolved_at,omitempty"`
}
```

### 工作流
```
请求到达 → 匹配断点 → 创建会话 → 暂停 → WebSocket 通知前端
    ↓
用户操作: 查看 / 修改 / 释放 / 丢弃
    ↓
执行动作 → 记录结果
```

### API
```
GET    /api/breakpoints              # 列表
POST   /api/breakpoints              # 创建
PUT    /api/breakpoints/:id          # 更新
DELETE /api/breakpoints/:id          # 删除
PATCH  /api/breakpoints/:id/toggle   # 启用/禁用
GET    /api/breakpoint-sessions      # 活跃会话列表
POST   /api/breakpoint-sessions/:id/release  # 释放
POST   /api/breakpoint-sessions/:id/modify   # 修改后释放
POST   /api/breakpoint-sessions/:id/drop     # 丢弃
```

---

## 模块 2: 请求重写 (Rewrite) — 待实现

### 目录结构
```
internal/rewrite/
├── engine.go          # 重写引擎
├── rules.go           # 重写规则存储
└── models.go          # 数据模型
```

### 数据模型
```go
type RewriteRule struct {
    ID       string       `json:"id"`
    Name     string       `json:"name"`
    Enabled  bool         `json:"enabled"`
    Priority int          `json:"priority"`
    Match    RuleMatch    `json:"match"`
    Actions  []RewriteAction `json:"actions"`
}

type RewriteAction struct {
    Type   RewriteType    `json:"type"`
    Target string         `json:"target"`   // header_name, body_field, url_path
    Where  string         `json:"where"`    // request / response
    Key    string         `json:"key"`      // 字段名 (header name 等)
    Value  string         `json:"value"`    // 替换值
}

type RewriteType string
const (
    RewriteAddHeader     RewriteType = "add_header"
    RewriteRemoveHeader  RewriteType = "remove_header"
    RewriteReplaceHeader RewriteType = "replace_header"
    RewriteReplaceBody   RewriteType = "replace_body"
    RewriteReplaceURL    RewriteType = "replace_url"
    RewriteMapLocal      RewriteType = "map_local"
    RewriteMapRemote     RewriteType = "map_remote"
)
```

### API
```
GET    /api/rewrites
POST   /api/rewrites
PUT    /api/rewrites/:id
DELETE /api/rewrites/:id
PATCH  /api/rewrites/:id/toggle
POST   /api/rewrites/reorder
```

---

## 模块 3: API 集合管理 (Collection) — 待实现

### 目录结构
```
internal/collection/
├── manager.go
├── collection.go
├── request.go
├── environment.go
├── runner.go
└── models.go
```

### 数据模型
```go
type Collection struct {
    ID          string            `json:"id"`
    Name        string            `json:"name"`
    Description string            `json:"description"`
    ParentID    string            `json:"parent_id,omitempty"`
    Variables   []Variable        `json:"variables"`
    Items       []CollectionItem  `json:"items"`
}

type APIRequest struct {
    ID      string        `json:"id"`
    Name    string        `json:"name"`
    Method  string        `json:"method"`
    URL     string        `json:"url"`
    Headers []KeyValue    `json:"headers"`
    Body    *RequestBody  `json:"body,omitempty"`
    Auth    *AuthConfig   `json:"auth,omitempty"`
    Tests   []Test        `json:"tests"`
}

type Environment struct {
    ID        string     `json:"id"`
    Name      string     `json:"name"`
    Variables []Variable `json:"variables"`
    IsActive  bool       `json:"is_active"`
}
```

### API
```
GET/POST        /api/collections
GET/PUT/DELETE  /api/collections/:id
GET/POST        /api/collections/:id/requests
POST            /api/collections/:id/run
GET/POST        /api/environments
POST            /api/collections/import    # Postman/Insomnia 导入
POST            /api/collections/export
```

---

## 模块 4: 脚本引擎 — 待实现

### 设计
- 嵌入式 Lua 或 Expr-lang 表达式引擎
- 请求阶段脚本: 修改请求、添加 Header、注入变量
- 响应阶段脚本: 修改响应、记录日志、触发告警
- 沙箱执行，超时限制

---

## 模块 5: 代码生成 — 待实现

### 支持语言
- cURL
- Python (requests)
- JavaScript (fetch/axios)
- Go (net/http)
- Java (OkHttp)
- PHP (cURL)

### API
```
POST /api/codegen/:id   # 根据流量 ID 生成代码
POST /api/codegen       # 根据请求定义生成代码
```

---

## 模块 6: Diff 对比 — 待实现

### 功能
- 两个请求的 Headers 对比
- 两个响应的 Body 对比
- JSON 结构化 Diff
- 文本逐行 Diff

### API
```
POST /api/diff   # 比较两个流量记录
```

---

## 前端页面规划

| 页面 | 路由 | 状态 |
|------|------|------|
| 流量页面 | `/` | ✅ 已实现 |
| 规则页面 | `/rules` | ❌ 待实现 |
| 断点页面 | `/breakpoints` | ❌ 待实现 |
| 重写页面 | `/rewrites` | ❌ 待实现 |
| 集合页面 | `/collections` | ❌ 待实现 |
| 环境页面 | `/environments` | ❌ 待实现 |
| AI 助手 | `/ai` | ❌ 待实现 |
| 设置页面 | `/settings` | ❌ 待实现 |

---

## 开发规范

1. 中文注释，简洁明了，专业名词用英文
2. 功能分支开发，PR review
3. 提交前格式化代码
4. 不要过度设计，保持代码简洁易读
