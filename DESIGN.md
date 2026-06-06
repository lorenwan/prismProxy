# Design System

## Theme

深色主题，GitHub 风格配色。开发者工具，减少长时间使用的眼疲劳。

## Color Palette

### Backgrounds

| Token | Hex | Usage |
|-------|-----|-------|
| `--bg-primary` | `#0d1117` | 页面背景、主内容区 |
| `--bg-secondary` | `#161b22` | 卡片、面板、侧边栏 |
| `--bg-panel` | `#1c2128` | 输入框、下拉菜单 |

### Borders

| Token | Hex | Usage |
|-------|-----|-------|
| `--border` | `#30363d` | 分割线、卡片边框、输入框边框 |

### Text

| Token | Hex | Usage |
|-------|-----|-------|
| `--text-primary` | `#e6edf3` | 主文本、标题 |
| `--text-secondary` | `#8b949e` | 次要文本、标签、描述 |

### Accent Colors

| Token | Hex | Usage |
|-------|-----|-------|
| `--blue` | `#58a6ff` | 链接、主按钮、选中状态、焦点环 |
| `--green` | `#3fb950` | 成功状态、GET 方法 |
| `--yellow` | `#d29922` | 警告状态、PUT 方法、4xx 状态码 |
| `--red` | `#f85149` | 错误状态、DELETE 方法、5xx 状态码 |
| `--purple` | `#bc8cff` | PATCH 方法、特殊标记 |

## Typography

### Font Stack

```css
font-family: 'Geist Variable', -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Noto Sans SC', sans-serif;
```

系统字体栈，优先使用平台原生字体，支持中文（Noto Sans SC）。

### Scale

| Size | Tailwind | Usage |
|------|----------|-------|
| xs | `text-xs` (12px) | Badge、小标签 |
| sm | `text-sm` (14px) | 次要文本、按钮 |
| base | `text-base` (16px) | 正文、对话框标题 |

## Spacing

| Token | Value | Usage |
|-------|-------|-------|
| `p-2.5` | 10px | 卡片头部/底部内边距 |
| `p-4` | 16px | 卡片内容区内边距 |
| `px-5` | 20px | 对话框水平内边距 |
| `gap-2` | 8px | 按钮间距 |
| `gap-3` | 12px | 卡片组间距 |

## Border Radius

| Token | Value | Usage |
|-------|-------|-------|
| `rounded-md` | 6px | 按钮、输入框 |
| `rounded-lg` | 8px | 卡片、对话框 |
| `rounded-full` | 9999px | Badge |

## Animations

| Name | Keyframes | Duration |
|------|-----------|----------|
| fadeIn | opacity 0→1 | 150ms |
| slideInRight | opacity 0→1, translateX 8px→0 | 150ms |
| slideInUp | opacity 0→1, translateY 8px→0 | 150ms |
| zoomIn | opacity 0→1, scale 0.95→1 | 150ms |

缓动函数: `cubic-bezier(0.4, 0, 0.2, 1)`

## HTTP Method Colors

| Method | Color | CSS Class |
|--------|-------|-----------|
| GET | green | `.method-get` |
| POST | blue | `.method-post` |
| PUT | yellow | `.method-put` |
| DELETE | red | `.method-delete` |
| PATCH | purple | `.method-patch` |

## Status Code Colors

| Range | Color | CSS Class |
|-------|-------|-----------|
| 2xx | green | `.status-2xx` |
| 3xx | blue | `.status-3xx` |
| 4xx | yellow | `.status-4xx` |
| 5xx | red | `.status-5xx` |
