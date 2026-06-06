# 前端代码风格规范

## TypeScript

### 类型定义

- 使用 `interface` 定义对象类型
- 使用 `type` 定义联合类型或工具类型
- 泛型参数使用有意义的名称（如 `TItem` 而非 `T`）

### 错误处理

- 使用 `Result<T, E>` 类型封装可能失败的操作
- 异步函数统一使用 try-catch 捕获错误
- 错误信息应包含上下文

### 空值处理

- 使用可选链 `?.` 访问可能为空的属性
- 使用空值合并 `??` 提供默认值
- 使用类型守卫缩小类型范围

---

## React

### 函数组件

- 使用函数声明而非箭头函数
- Props 类型使用 `interface` 定义
- 组件名与文件名保持一致

### Hooks 使用

- 自定义 Hook 以 `use` 开头
- `useEffect` 必须返回清理函数
- 避免在循环或条件中调用 Hook

---

## TailwindCSS

### 类名顺序

- 布局 → 尺寸 → 间距 → 外观 → 交互
- 响应式类名放在最后

### 响应式设计

- 使用移动端优先策略
- 使用 `md:` 和 `lg:` 前缀

---

## 命名规范

| 类型 | 规范 | 示例 |
|------|------|------|
| 组件文件 | PascalCase | `TrafficList.tsx` |
| 工具文件 | kebab-case | `traffic-utils.ts` |
| Hook 文件 | camelCase (use 前缀) | `useWebSocket.ts` |
| 组件 | PascalCase | `TrafficList` |
| 变量 | camelCase | `trafficList` |
| 常量 | UPPER_SNAKE_CASE | `MAX_PAGE_SIZE` |

---

## 目录结构

```
desktop/src/
├── components/          # 通用组件
│   ├── layout/          # 布局组件
│   └── ui/              # UI 基础组件（shadcn/ui）
├── features/            # 功能模块
│   └── [module]/
│       ├── components/  # 模块专属组件
│       ├── index.ts     # 统一导出
│       ├── [module]Store.ts
│       └── [module]Service.ts
├── hooks/               # 自定义 hooks
├── lib/                 # 工具函数
├── pages/               # 页面组件
├── services/            # API 服务（Tauri IPC）
└── types/               # 类型定义
```

---

## 安全规范

- API Key 不写入日志
- 敏感信息（密码、Token）必须脱敏显示
