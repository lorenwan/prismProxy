import { invoke } from '@tauri-apps/api/core'
import { listen } from '@tauri-apps/api/event'

// --- 类型定义 ---

export interface TrafficListResponse {
  entries: TrafficEntry[]
  pagination: PageMeta
}

export interface TrafficEntry {
  id: string
  method: string
  url: string
  host: string
  path: string
  scheme: string
  port: number
  status: number
  statusCode: number
  contentType: string
  size: number
  duration: number
  requestTime: string
  source: string
  bookmarked: boolean
  notes: string
  color: string
  tags: string[]
  request: RequestData
  response: ResponseData
}

export interface RequestData {
  method: string
  url: string
  headers: Record<string, string>
  body: string
  contentType: string
  size: number
}

export interface ResponseData {
  status: number
  statusText: string
  headers: Record<string, string>
  body: string
  contentType: string
  size: number
}

export interface PageMeta {
  page: number
  pageSize: number
  total: number
}

export interface TrafficStats {
  totalRequests: number
  successRequests: number
  failedRequests: number
  totalSize: number
  avgDuration: number
}

// --- Tauri IPC 函数 ---

/**
 * 获取流量列表
 * @param page 页码，默认 1
 * @param pageSize 每页数量，默认 20
 */
export async function listTraffic(page: number = 1, pageSize: number = 20): Promise<TrafficListResponse> {
  return invoke('list_traffic', { page, pageSize })
}

/**
 * 获取单条流量详情
 * @param id 流量记录 ID
 */
export async function getTraffic(id: string): Promise<TrafficEntry> {
  return invoke('get_traffic', { id })
}

/**
 * 删除流量记录
 * @param ids 要删除的流量记录 ID 数组
 */
export async function deleteTraffic(ids: string[]): Promise<void> {
  return invoke('delete_traffic', { ids })
}

/**
 * 清空所有流量记录
 */
export async function clearTraffic(): Promise<void> {
  return invoke('clear_traffic')
}

/**
 * 订阅流量事件（服务端流式推送）
 * @param callback 事件回调函数
 * @returns 取消订阅的函数
 */
export async function subscribeTraffic(callback: (event: any) => void): Promise<() => void> {
  // 启动订阅
  await invoke('subscribe_traffic')

  // 监听事件
  const unlisten = await listen('traffic:event', (event) => {
    callback(JSON.parse(event.payload as string))
  })

  return unlisten
}

// --- 向后兼容的别名函数 ---
// 保持与现有调用方（TrafficPage、StatusBar、Header）兼容

/**
 * 获取流量列表（向后兼容，包装 listTraffic）
 */
export async function getTrafficList(params?: {
  page?: number
  pageSize?: number
  method?: string
  status?: number
  host?: string
}): Promise<{ data: { data: any[]; total: number } }> {
  const page = params?.page ?? 1
  const pageSize = params?.pageSize ?? 20
  const result = await listTraffic(page, pageSize)
  return { data: { data: result.entries, total: result.pagination.total } }
}

/**
 * 获取单条流量详情（向后兼容，包装 getTraffic）
 */
export async function getTrafficDetail(id: string): Promise<any> {
  return getTraffic(id)
}

/**
 * 获取流量统计（向后兼容，返回模拟数据）
 * 注意：Rust 层暂未实现 get_traffic_stats IPC 命令
 */
export async function getTrafficStats(): Promise<{ data: TrafficStats }> {
  // 当 Rust 层 IPC 命令就绪后替换为 invoke 调用
  return { data: { totalRequests: 0, successRequests: 0, failedRequests: 0, totalSize: 0, avgDuration: 0 } }
}

/**
 * 更新书签（向后兼容）
 * 注意：Rust 层暂未实现 update_bookmark IPC 命令
 */
export async function updateBookmark(id: string, bookmarked: boolean): Promise<void> {
  // 当 Rust 层 IPC 命令就绪后替换为 invoke 调用
}

/**
 * 更新备注（向后兼容）
 * 注意：Rust 层暂未实现 update_notes IPC 命令
 */
export async function updateNotes(id: string, notes: string): Promise<void> {
  // 当 Rust 层 IPC 命令就绪后替换为 invoke 调用
}

/**
 * 更新颜色标记（向后兼容）
 * 注意：Rust 层暂未实现 update_color IPC 命令
 */
export async function updateColor(id: string, color: string): Promise<void> {
  // 当 Rust 层 IPC 命令就绪后替换为 invoke 调用
}

/**
 * 更新标签（向后兼容）
 * 注意：Rust 层暂未实现 update_tags IPC 命令
 */
export async function updateTags(id: string, tags: string[]): Promise<void> {
  // 当 Rust 层 IPC 命令就绪后替换为 invoke 调用
}
