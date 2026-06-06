# 代码仓库完整扫描报告

扫描时间：2026-06-06

---

## 问题汇总

| 类别 | 问题 | 数量 | 优先级 |
|------|------|------|--------|
| 架构 | WebSocket 直连 Go 后端 | 4 文件 | P0 |
| 可访问性 | 缺少 ARIA 的表单 | 48 | P1 |
| 可访问性 | 缺少 ARIA 的按钮 | 71 | P1 |
| 可访问性 | 缺少 role 的 toggle | 8 | P1 |
| 可访问性 | 缺少 htmlFor 的 label | 18 | P1 |
| 可访问性 | 缺少 aria-live 的 Toast | 1 | P1 |
| 可访问性 | 缺少焦点管理的模态框 | 5 | P1 |
| 可访问性 | 缺少键盘导航的可点击元素 | 100 | P2 |
| 可访问性 | 缺少 aria-hidden 的图标 | 2 | P2 |
| 可访问性 | 缺少语义化的可点击元素 | 6 | P2 |
| 代码质量 | 硬编码颜色 | 10 | P1 |
| 代码质量 | any 类型使用 | 8 | P2 |
| 代码质量 | console.log 语句 | 24 | P3 |
| 性能 | 缺少 useMemo/useCallback | 82 | P2 |
| 性能 | 缺少代码分割的页面 | 11 | P2 |
| 性能 | 缺少防抖的搜索 | 3 | P2 |
| 性能 | 缺少容量限制的 Store | 14 | P2 |
| 错误处理 | 缺少错误处理的 API 调用 | 94 | P1 |
| 错误处理 | 缺少错误边界的路由 | 11 | P1 |
| 错误处理 | 缺少 loading 状态 | 83 | P2 |
| 错误处理 | 缺少错误状态 | 83 | P2 |
| 响应式 | 固定尺寸 | 11 | P3 |
| 动画 | 缺少 reduced-motion | 39 | P2 |

---

## 详细问题列表

### 1. 架构问题 (P0)

#### 1.1 WebSocket 直连 Go 后端
**文件**：
- `hooks/useWebSocket.ts`
- `pages/TrafficPage.tsx`
- `components/layout/StatusBar.tsx`
- `types/index.ts`

**问题**：WebSocket 直接连接 Go 后端，绕过 Rust 层，与统一架构不一致。

**解决方案**：迁移到 Tauri Event 模式。

---

### 2. 可访问性问题 (P1)

#### 2.1 缺少 ARIA 的表单
**数量**：48
**文件**：所有页面的 input、select、textarea

**问题**：表单元素缺少 aria-label 或 aria-labelledby。

**解决方案**：为所有表单添加 aria-label。

#### 2.2 缺少 ARIA 的按钮
**数量**：71
**文件**：所有页面的 button

**问题**：按钮缺少 aria-label。

**解决方案**：为所有按钮添加 aria-label。

#### 2.3 缺少 role 的 toggle
**数量**：8
**文件**：
- `RulesPage.tsx`
- `BreakpointsPage.tsx`
- `RewritePage.tsx`
- `ScriptsPage.tsx`
- `SettingsPage.tsx`

**问题**：自定义 toggle 缺少 role="switch" 和 aria-checked。

**解决方案**：添加 ARIA 属性。

#### 2.4 缺少 htmlFor 的 label
**数量**：18
**文件**：所有页面的 label

**问题**：label 未通过 htmlFor 关联到表单。

**解决方案**：添加 htmlFor 和 id。

#### 2.5 缺少 aria-live 的 Toast
**数量**：1
**文件**：`components/ui/Toast.tsx`

**问题**：Toast 容器缺少 aria-live="polite"。

**解决方案**：添加 ARIA 属性。

#### 2.6 缺少焦点管理的模态框
**数量**：5
**文件**：多个页面的删除确认对话框

**问题**：模态框缺少焦点捕获和 Escape 键关闭。

**解决方案**：添加焦点管理。

---

### 3. 代码质量问题 (P1-P3)

#### 3.1 硬编码颜色
**数量**：10
**文件**：
- `ScriptsPage.tsx`
- `CollectionsPage.tsx`
- `EnvironmentsPage.tsx`
- `BreakpointsPage.tsx`
- `PerformancePage.tsx`

**问题**：使用硬编码颜色值而非 CSS 变量。

**解决方案**：替换为 CSS 变量。

#### 3.2 any 类型使用
**数量**：8
**文件**：多个文件

**问题**：使用 any 类型，降低类型安全。

**解决方案**：定义具体类型。

#### 3.3 console.log 语句
**数量**：24
**文件**：多个文件

**问题**：生产代码中包含调试日志。

**解决方案**：移除或使用日志库。

---

### 4. 性能问题 (P2)

#### 4.1 缺少 useMemo/useCallback
**数量**：82
**文件**：所有页面

**问题**：未优化的计算和回调。

**解决方案**：使用 useMemo/useCallback 包裹。

#### 4.2 缺少代码分割
**数量**：11
**文件**：`App.tsx`

**问题**：所有页面直接导入，无懒加载。

**解决方案**：使用 React.lazy + Suspense。

#### 4.3 缺少防抖的搜索
**数量**：3
**文件**：多个页面

**问题**：搜索输入未防抖。

**解决方案**：添加 debounce。

#### 4.4 缺少容量限制的 Store
**数量**：14
**文件**：多个 Store

**问题**：Store 无容量限制，可能导致内存溢出。

**解决方案**：添加最大长度限制。

---

### 5. 错误处理问题 (P1-P2)

#### 5.1 缺少错误处理的 API 调用
**数量**：94
**文件**：所有页面

**问题**：API 调用缺少 try-catch。

**解决方案**：添加错误处理。

#### 5.2 缺少错误边界
**数量**：11
**文件**：`App.tsx`

**问题**：路由缺少 ErrorBoundary。

**解决方案**：添加错误边界组件。

#### 5.3 缺少 loading 状态
**数量**：83
**文件**：所有页面

**问题**：异步操作缺少 loading 状态。

**解决方案**：添加 loading 状态。

#### 5.4 缺少错误状态
**数量**：83
**文件**：所有页面

**问题**：异步操作缺少错误状态。

**解决方案**：添加错误状态。

---

### 6. 响应式问题 (P3)

#### 6.1 固定尺寸
**数量**：11
**文件**：多个页面

**问题**：使用固定像素值。

**解决方案**：使用响应式单位。

---

### 7. 动画问题 (P2)

#### 7.1 缺少 reduced-motion
**数量**：39
**文件**：多个页面

**问题**：动画缺少 prefers-reduced-motion 支持。

**解决方案**：添加媒体查询。

---

## 修复计划

### 阶段 1：架构和可访问性 (P0-P1)
- 迁移 WebSocket 到 Tauri Event
- 添加 ARIA 属性
- 修复焦点管理

### 阶段 2：代码质量和错误处理 (P1)
- 替换硬编码颜色
- 添加错误处理
- 添加错误边界

### 阶段 3：性能优化 (P2)
- 优化 useMemo/useCallback
- 添加代码分割
- 添加防抖

### 阶段 4：其他优化 (P2-P3)
- 移除 console.log
- 添加 reduced-motion
- 优化响应式
