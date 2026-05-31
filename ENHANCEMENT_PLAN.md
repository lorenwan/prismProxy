# PrismProxy 功能增强 + UI 美化方案

## 一、补齐缺失模块

### 1. 脚本引擎 (Script Engine)
```
internal/script/
├── engine.go          # 脚本引擎主控
├── runtime.go         # 沙箱运行时 (expr-lang)
├── scripts.go         # 脚本存储
└── models.go          # 数据模型
```

功能：
- 使用 expr-lang 表达式引擎（轻量、安全、无需嵌入 Lua）
- 支持请求阶段脚本：修改请求、添加 Header、注入变量
- 支持响应阶段脚本：修改响应、记录日志、触发告警
- 内置函数：json_path, base64, md5, sha256, timestamp, uuid, regex_match
- 脚本库：预置常用脚本模板

Proto: `proto/scripts.proto`

### 2. Diff 对比 (Diff Compare)
```
internal/diff/
├── engine.go          # Diff 引擎
├── compare.go         # 对比算法
└── models.go          # 数据模型
```

功能：
- Headers 逐字段对比
- Body 文本逐行 Diff
- JSON 结构化 Diff（深度对比，高亮差异）
- Query 参数对比
- 支持忽略顺序、忽略大小写等选项

Proto: `proto/diff.proto`

### 3. 前端缺失页面
- RewritePage — 请求重写规则管理
- CollectionsPage — API 集合管理（类似 Postman Collection）
- EnvironmentsPage — 环境变量管理
- CodeGenPage — 代码生成（集成在集合页面中）
- ScriptsPage — 脚本管理
- DiffPage — Diff 对比页面

---

## 二、新增功能模块

### 4. 请求构建器 (Request Builder)
独立的 HTTP 请求发送工具（类似 Postman 的 Send 功能）：
- 方法选择、URL 输入、Headers 编辑
- Body 编辑器（JSON/Form/Raw/Binary）
- Auth 配置（Basic/Bearer/API Key/OAuth2）
- 发送请求并显示响应
- 保存到集合

### 5. 证书管理 (Certificate Manager)
```
internal/cert/
├── ca.go              # CA 证书管理
├── store.go           # 证书存储
└── models.go          # 数据模型
```

功能：
- 自动生成根 CA 证书
- 导出 CA 证书供客户端安装
- 查看已签发的域名证书列表
- 证书过期提醒

### 6. WebSocket 调试器
增强 WebSocket 支持：
- 拦截 WebSocket 消息
- 查看消息时间线
- 手动发送 WebSocket 消息
- 消息过滤和搜索

### 7. 性能分析 (Performance)
```
internal/perf/
├── analyzer.go        # 性能分析器
├── stats.go           # 统计数据
└── models.go          # 数据模型
```

功能：
- 请求耗时统计（DNS/TCP/TLS/首字节/下载）
- 慢请求告警
- 按域名/路径统计平均耗时
- 吞吐量统计

### 8. 搜索与过滤增强
- 全文搜索（URL + Headers + Body）
- 正则表达式过滤
- 保存过滤器（Saved Filters）
- 按状态码/方法/Content-Type 快速过滤

---

## 三、UI 美化方案

### 设计风格：类 Reqable 深色主题

#### 配色方案
```
背景色:
  - 主背景: #0d1117 (GitHub Dark)
  - 次背景: #161b22
  - 面板背景: #1c2128
  - 悬浮: #21262d
  - 边框: #30363d

文字色:
  - 主文字: #e6edf3
  - 次文字: #8b949e
  - 链接: #58a6ff
  - 成功: #3fb950
  - 警告: #d29922
  - 错误: #f85149
  - 信息: #58a6ff

强调色:
  - 主色: #58a6ff (蓝色)
  - 次色: #bc8cff (紫色)
  - GET: #3fb950
  - POST: #58a6ff
  - PUT: #d29922
  - DELETE: #f85149
  - PATCH: #bc8cff
```

#### 布局设计
```
┌─────────────────────────────────────────────────────────────┐
│ 顶部工具栏: Logo | 代理状态 | 搜索栏 | 过滤器 | 设置 | AI   │
├──────┬──────────────────────────────────────────────────────┤
│      │ Tab 栏: 流量 | 规则 | 断点 | 重写 | 集合 | 脚本 | 设置 │
│ 侧   ├──────────────────────────────────────────────────────┤
│ 栏   │                                                      │
│      │ 主内容区域                                            │
│ 快   │                                                      │
│ 捷   │                                                      │
│ 导   │                                                      │
│ 航   │                                                      │
│      │                                                      │
│      ├──────────────────────────────────────────────────────┤
│      │ 底部状态栏: 请求数 | 流量速率 | 内存使用 | 版本         │
├──────┴──────────────────────────────────────────────────────┤
```

#### 流量页面增强
```
┌──────────────────────────────────────────────────────────────┐
│ 过滤工具栏: [方法▼] [状态码▼] [类型▼] [时间范围] [保存过滤器]  │
├──────────────────────────────────────────────────────────────┤
│ # │ 方法 │ 状态 │ URL              │ 大小  │ 时间  │ 标记     │
│───┼──────┼──────┼──────────────────┼───────┼───────┼─────────│
│ 1 │ GET  │ 200  │ /api/users       │ 1.2KB │ 45ms  │         │
│ 2 │ POST │ 201  │ /api/login       │ 0.5KB │ 120ms │ ⚡慢     │
│ 3 │ GET  │ 404  │ /api/missing     │ 0.1KB │ 12ms  │ ❌错误   │
│ 4 │ PUT  │ 200  │ /api/users/1     │ 0.8KB │ 67ms  │ 🔒HTTPS │
├──────────────────────────────────────────────────────────────┤
│ 详情面板 (可拖拽调整高度):                                     │
│ ┌────────┬────────┬────────┬────────┬────────┐              │
│ │ 概览   │ 请求   │ 响应   │ 头部   │ Cookies│              │
│ ├────────┴────────┴────────┴────────┴────────┤              │
│ │ [JSON 格式化] [复制] [发送到集合] [Diff] [AI分析]           │
│ │ ┌──────────────────────────────────────────┐│              │
│ │ │ {                                        ││              │
│ │ │   "id": 1,                               ││              │
│ │ │   "name": "test"                         ││              │
│ │ │ }                                        ││              │
│ │ └──────────────────────────────────────────┘│              │
│ └─────────────────────────────────────────────┘              │
└──────────────────────────────────────────────────────────────┘
```

#### 组件设计规范
- 圆角: 6px (小), 8px (中), 12px (大)
- 间距: 4px (紧凑), 8px (标准), 12px (宽松), 16px (分组)
- 阴影: 0 2px 8px rgba(0,0,0,0.3)
- 动画: 150ms ease-in-out
- 字体: Inter / SF Mono (代码)
- 图标: Lucide React

#### 交互增强
- 行悬停高亮
- 右键上下文菜单
- 拖拽调整面板大小
- 键盘快捷键
- 动画过渡效果
- Toast 通知

---

## 四、Proto 文件新增

### scripts.proto
```protobuf
service ScriptService {
  rpc ListScripts(ListScriptsRequest) returns (ListScriptsResponse);
  rpc CreateScript(CreateScriptRequest) returns (Script);
  rpc UpdateScript(UpdateScriptRequest) returns (Script);
  rpc DeleteScript(DeleteScriptRequest) returns (google.protobuf.Empty);
  rpc ToggleScript(ToggleScriptRequest) returns (Script);
  rpc GetScriptLibrary(google.protobuf.Empty) returns (ScriptLibraryResponse);
  rpc TestScript(TestScriptRequest) returns (TestScriptResponse);
}
```

### diff.proto
```protobuf
service DiffService {
  rpc CompareRequests(CompareRequestsRequest) returns (DiffResult);
  rpc CompareResponses(CompareResponsesRequest) returns (DiffResult);
  rpc CompareJSON(CompareJSONRequest) returns (JSONDiffResult);
}
```

### perf.proto
```protobuf
service PerfService {
  rpc GetStats(GetStatsRequest) returns (PerfStats);
  rpc GetSlowRequests(GetSlowRequestsRequest) returns (SlowRequestsResponse);
  rpc GetDomainStats(google.protobuf.Empty) returns (DomainStatsResponse);
  rpc GetTimelineStats(GetTimelineRequest) returns (TimelineStatsResponse);
}
```

### cert.proto
```protobuf
service CertService {
  rpc GetCAInfo(google.protobuf.Empty) returns (CAInfo);
  rpc ExportCACert(google.protobuf.Empty) returns (CertFile);
  rpc ListCerts(ListCertsRequest) returns (ListCertsResponse);
  rpc RegenerateCA(google.protobuf.Empty) returns (CAInfo);
}
```

### search.proto
```protobuf
service SearchService {
  rpc FullTextSearch(FullTextSearchRequest) returns (SearchResults);
  rpc ListSavedFilters(google.protobuf.Empty) returns (SavedFiltersResponse);
  rpc SaveFilter(SaveFilterRequest) returns (SavedFilter);
  rpc DeleteFilter(DeleteFilterRequest) returns (google.protobuf.Empty);
}
```

---

## 五、前端新增依赖

```json
{
  "dependencies": {
    "@radix-ui/react-tabs": "^1.0.0",
    "@radix-ui/react-dialog": "^1.0.0",
    "@radix-ui/react-dropdown-menu": "^1.0.0",
    "@radix-ui/react-context-menu": "^1.0.0",
    "@radix-ui/react-tooltip": "^1.0.0",
    "@radix-ui/react-toast": "^1.0.0",
    "@radix-ui/react-select": "^1.0.0",
    "@radix-ui/react-switch": "^1.0.0",
    "@radix-ui/react-slider": "^1.0.0",
    "@radix-ui/react-popover": "^1.0.0",
    "lucide-react": "^0.300.0",
    "clsx": "^2.0.0",
    "tailwind-merge": "^2.0.0",
    "framer-motion": "^11.0.0",
    "@uiw/react-json-view": "^2.0.0",
    "diff": "^5.0.0"
  }
}
```

---

## 六、实施顺序

1. **安装前端依赖** (Radix UI, Lucide, Framer Motion)
2. **创建全局 UI 组件库** (Button, Input, Card, Badge, Tabs, Dialog, Toast, ContextMenu)
3. **美化 Layout 和导航**
4. **补齐后端模块** (Script, Diff, Perf, Cert, Search)
5. **补齐 Proto + gRPC 服务**
6. **补齐前端页面** (Rewrite, Collection, Environment, Script, Diff, Performance)
7. **增强流量页面** (过滤器、详情面板、右键菜单)
8. **重新构建 Tauri 应用**
