import { invoke } from '@tauri-apps/api/core'

// --- 类型定义 ---

export interface TrafficListResponse {
  entries: TrafficEntry[]
  pagination: PageMeta
}

export interface TrafficEntry {
  id: string
  timestamp: string
  duration_ms: number
  method: string
  url: string
  host: string
  path: string
  scheme: string
  port: number
  request: RequestData
  response: ResponseData
  client_addr: string
  server_ip: string
  bookmarked: boolean
  color: string
  notes: string
  tags: string[]
}

export interface RequestData {
  headers: Record<string, { values: string[] }>
  body: string
  body_size: number
  content_type: string
}

export interface ResponseData {
  status_code: number
  status_text: string
  headers: Record<string, { values: string[] }>
  body: string
  body_size: number
  content_type: string
}

export interface PageMeta {
  page: number
  pageSize: number
  total: number
}

export interface TrafficStats {
  total_requests: number
  total_responses: number
  avg_duration_ms: number
  max_duration_ms: number
  min_duration_ms: number
  error_count: number
  success_count: number
  host_stats: Array<{ host: string; count: number; avg_time_ms: number }>
  method_stats: Array<{ method: string; count: number }>
  status_stats: Array<{ status_code: number; count: number }>
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
  return invoke('get_traffic', { id: parseInt(id, 10) })
}

/**
 * 删除流量记录
 * @param ids 要删除的流量记录 ID 数组
 */
export async function deleteTraffic(ids: string[]): Promise<void> {
  return invoke('delete_traffic', { ids: ids.map(id => parseInt(id, 10)) })
}

/**
 * 清空所有流量记录
 */
export async function clearTraffic(): Promise<void> {
  return invoke('clear_traffic')
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
}): Promise<{ data: { data: TrafficEntry[]; total: number } }> {
  const result = await listTraffic(params?.page, params?.pageSize)
  return { data: { data: result.entries, total: result.pagination.total } }
}

/**
 * 获取单条流量详情（向后兼容，包装 getTraffic）
 */
export async function getTrafficDetail(id: string): Promise<any> {
  return getTraffic(id)
}

/**
 * 获取流量统计
 */
export async function getTrafficStats(): Promise<{ data: TrafficStats }> {
  const result = await invoke<TrafficStats>('get_traffic_stats')
  return { data: result }
}

/**
 * 更新书签
 */
export async function updateBookmark(id: string, bookmarked: boolean): Promise<void> {
  return invoke('update_traffic_bookmark', { id: parseInt(id, 10), bookmarked })
}

/**
 * 更新备注
 */
export async function updateNotes(id: string, notes: string): Promise<void> {
  return invoke('update_traffic_notes', { id: parseInt(id, 10), notes })
}

/**
 * 更新颜色标记
 */
export async function updateColor(id: string, color: string): Promise<void> {
  return invoke('update_traffic_color', { id: parseInt(id, 10), color })
}

/**
 * 更新标签
 */
export async function updateTags(id: string, tags: string[]): Promise<void> {
  return invoke('update_traffic_tags', { id: parseInt(id, 10), tags })
}
