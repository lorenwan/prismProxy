// 状态码颜色
export function getStatusColor(status: number): string {
  if (status >= 200 && status < 300) return 'text-[var(--green)]'
  if (status >= 300 && status < 400) return 'text-[var(--blue)]'
  if (status >= 400 && status < 500) return 'text-[var(--yellow)]'
  return 'text-[var(--red)]'
}

// 方法颜色
export function getMethodColor(method: string): string {
  switch (method) {
    case 'GET': return 'text-[var(--green)]'
    case 'POST': return 'text-[var(--blue)]'
    case 'PUT': return 'text-[var(--yellow)]'
    case 'DELETE': return 'text-[var(--red)]'
    default: return 'text-[var(--text-secondary)]'
  }
}

// 格式化大小
export function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

// 格式化响应体
export function formatBody(body: string, contentType: string): string {
  if (contentType.includes('json')) {
    try {
      return JSON.stringify(JSON.parse(body), null, 2)
    } catch {
      return body
    }
  }
  return body
}

// HTTP 方法 Badge 变体
export function getMethodBadgeVariant(method: string): string {
  switch (method) {
    case 'GET': return 'green'
    case 'POST': return 'blue'
    case 'PUT': return 'yellow'
    case 'DELETE': return 'red'
    case 'PATCH': return 'purple'
    default: return 'default'
  }
}

// 状态码 Badge 变体
export function getStatusBadgeVariant(status: number): string {
  if (status >= 200 && status < 300) return 'green'
  if (status >= 300 && status < 400) return 'blue'
  if (status >= 400 && status < 500) return 'yellow'
  return 'red'
}
