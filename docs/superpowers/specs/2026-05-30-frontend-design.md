# PrismProxy 前端设计文档

## 概述

为 PrismProxy HTTP 调试工具创建前端项目，使用 React + TypeScript + TailwindCSS，深色专业工具风格，支持实时流量推送。

## 技术栈

- **构建工具**: Vite 6
- **框架**: React 19 + TypeScript 5
- **样式**: TailwindCSS 4
- **状态管理**: Zustand 5
- **路由**: React Router 7
- **HTTP 客户端**: Axios
- **代码编辑器**: Monaco Editor (@monaco-editor/react)

## 项目结构

```
web/
├── index.html
├── package.json
├── vite.config.ts
├── tsconfig.json
└── src/
    ├── main.tsx
    ├── App.tsx
    ├── index.css
    ├── types/
    │   └── index.ts
    ├── services/
    │   ├── api.ts
    │   ├── traffic.ts
    │   ├── rules.ts
    │   └── ai.ts
    ├── stores/
    │   ├── trafficStore.ts
    │   ├── rulesStore.ts
    │   └── uiStore.ts
    ├── hooks/
    │   └── useWebSocket.ts
    ├── components/
    │   ├── layout/
    │   │   ├── Sidebar.tsx
    │   │   ├── Header.tsx
    │   │   └── StatusBar.tsx
    │   ├── traffic/
    │   │   ├── TrafficList.tsx
    │   │   ├── TrafficDetail.tsx
    │   │   └── TrafficItem.tsx
    │   ├── rules/
    │   │   ├── RuleList.tsx
    │   │   └── RuleEditor.tsx
    │   └── ai/
    │       └── AiAssistant.tsx
    └── pages/
        ├── TrafficPage.tsx
        ├── RulesPage.tsx
        └── AiPage.tsx
```

## 路由设计

| 路径 | 页面 | 说明 |
|------|------|------|
| `/` | TrafficPage | 流量列表（默认主页） |
| `/rules` | RulesPage | 规则配置 |
| `/ai` | AiPage | AI 助手 |

## 全局布局

```
┌─────────────────────────────────────────────┐
│  Header (Logo + 搜索 + 工具栏)               │
├────┬────────────────────────────────────────┤
│    │                                        │
│ S  │   主内容区                              │
│ i  │   (TrafficPage: 左列表 + 右详情)        │
│ d  │   (RulesPage: 列表 + 编辑器)            │
│ e  │   (AiPage: 对话界面)                    │
│ b  │                                        │
│ a  │                                        │
│ r  │                                        │
├────┴────────────────────────────────────────┤
│  StatusBar (连接状态 + 流量统计)              │
└─────────────────────────────────────────────┘
```

- **Sidebar**: 固定宽度 56px，图标导航，当前页高亮
- **Header**: 全局搜索框、清空流量、设置按钮
- **StatusBar**: WebSocket 连接状态、总请求数、代理端口

## 页面设计

### 流量页面 (TrafficPage)

左右分栏，可拖拽调整宽度。

**左侧 - TrafficList**:
- 表头：状态码 | 方法 | Host | Path | 耗时 | 时间
- 颜色编码：2xx 绿色、3xx 蓝色、4xx 橙色、5xx 红色
- 支持按方法/状态码/Host 过滤
- 实时追加新流量（WebSocket `traffic:new` 事件）
- 选中行高亮

**右侧 - TrafficDetail**:
- Tab 切换：请求 | 响应 | 概览
- 请求 Tab：Headers 表格 + Body（JSON/HTML/原始格式）
- 响应 Tab：Headers 表格 + Body（JSON 美化、图片预览、HTML 渲染）
- 概览 Tab：URL、IP、耗时、时间线等元信息
- Body 使用 Monaco Editor 只读展示

### 规则页面 (RulesPage)

- 左侧规则列表：名称、类型图标、启用开关、优先级、命中次数
- 右侧规则编辑器：表单式编辑匹配条件和动作
- 支持操作类型：Map Local、Map Remote、修改请求/响应、阻止、延迟、Mock
- 顶部工具栏：新增、导入/导出、启用全部/禁用全部

### AI 助手页面 (AiPage)

- 聊天界面，支持流式输出
- 预设快捷操作：分析选中流量、安全检测、生成测试用例
- 支持选择 AI 提供商（OpenAI/Claude/Ollama）
- 对话历史上下文

## WebSocket 实时推送

- 连接地址：`ws://localhost:8081/ws`
- 后端需在 Gin 路由中注册 `/ws` 端点，升级 HTTP 连接为 WebSocket
- 消息格式：`{ type: string, payload: any, time: string }`
- 事件类型：
  - `traffic:new` → 追加到列表顶部
  - `traffic:delete` → 从列表移除
  - `traffic:clear` → 清空列表
- 断线自动重连（指数退避：1s, 2s, 4s, 8s, 最大 30s）

## 后端 API 补充

### 规则 API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/rules | 获取规则列表 |
| POST | /api/rules | 创建规则 |
| PUT | /api/rules/:id | 更新规则 |
| DELETE | /api/rules/:id | 删除规则 |
| POST | /api/rules/:id/toggle | 切换启用状态 |
| POST | /api/rules/reorder | 重排序 |
| POST | /api/rules/import | 导入规则 |
| GET | /api/rules/export | 导出规则 |

### AI API

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/ai/chat | AI 聊天（支持流式） |
| POST | /api/ai/analyze | 流量分析 |
| POST | /api/ai/security | 安全检测 |
| POST | /api/ai/testgen | 测试生成 |
| GET | /api/ai/providers | 可用提供商列表 |

### 流量 API 增强

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/traffic/stats | 流量统计 |
| PUT | /api/traffic/:id/bookmark | 更新书签 |
| PUT | /api/traffic/:id/notes | 更新备注 |
| PUT | /api/traffic/:id/color | 更新颜色标记 |
| PUT | /api/traffic/:id/tags | 更新标签 |

## 类型定义

前端 TypeScript 类型与后端 Go 结构体一一对应：

- `Transaction` — 流量记录
- `RequestData` / `ResponseData` — 请求/响应数据
- `TrafficStats` — 流量统计
- `Rule` / `RuleMatch` / `RuleAction` — 规则相关
- `ChatMessage` / `ChatRequest` / `ChatResponse` — AI 聊天
- `AnalysisResult` / `SecurityReport` / `TestCase` — AI 功能

## 开发配置

Vite 开发服务器配置代理：

```ts
// vite.config.ts
export default defineConfig({
  server: {
    port: 3000,
    proxy: {
      '/api': 'http://localhost:8081',
      '/ws': {
        target: 'ws://localhost:8081',
        ws: true,
      },
    },
  },
})
```

## 深色主题配色

- 背景：`#1a1b26` (主背景) / `#24283b` (面板背景) / `#16161e` (侧边栏)
- 文字：`#a9b1d6` (常规) / `#c0caf5` (高亮) / `#565f89` (次要)
- 强调色：`#7aa2f7` (蓝色) / `#9ece6a` (绿色) / `#f7768e` (红色) / `#e0af68` (黄色)
- 边框：`#3b4261`

## 实施步骤

1. **Phase 1**: 初始化 Vite 项目，安装依赖，配置 TailwindCSS 4（使用 @tailwindcss/vite 插件）
2. **Phase 2**: 实现类型定义、API 服务层、WebSocket hook
3. **Phase 3**: 实现全局布局组件（Sidebar, Header, StatusBar）
4. **Phase 4**: 实现流量页面（TrafficList, TrafficDetail, 实时推送）
5. **Phase 5**: 补充后端规则和 AI API
6. **Phase 6**: 实现规则页面
7. **Phase 7**: 实现 AI 助手页面
