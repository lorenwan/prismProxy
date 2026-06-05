import { invoke } from '@tauri-apps/api/core'
import type { Transaction } from '../types'

export interface SearchQuery {
  keyword: string
  scope?: 'url' | 'headers' | 'body' | 'all'
  method?: string
  status?: number
  host?: string
  contentType?: string
  startTime?: string
  endTime?: string
  minSize?: number
  maxSize?: number
  page?: number
  pageSize?: number
}

export interface SearchResult {
  transactions: Transaction[]
  total: number
  highlights: Record<string, string[]>
  took: number
}

export interface SearchSuggestion {
  type: 'host' | 'path' | 'header' | 'recent'
  value: string
  count: number
}

// 全文搜索
export async function search(query: SearchQuery): Promise<SearchResult> {
  const result = await invoke<string>('search', {
    query: query.keyword,
    sort: '',
    page: query.page,
    pageSize: query.pageSize,
  })
  return JSON.parse(result)
}

// 按方法搜索
export async function searchByMethod(method: string, page?: number, pageSize?: number): Promise<SearchResult> {
  const result = await invoke<string>('search_by_method', { method, page, pageSize })
  return JSON.parse(result)
}

// 按主机搜索
export async function searchByHost(host: string, page?: number, pageSize?: number): Promise<SearchResult> {
  const result = await invoke<string>('search_by_host', { host, page, pageSize })
  return JSON.parse(result)
}

// 按状态码搜索
export async function searchByStatusCode(statusCode: number, page?: number, pageSize?: number): Promise<SearchResult> {
  const result = await invoke<string>('search_by_status_code', { statusCode, page, pageSize })
  return JSON.parse(result)
}

// 搜索慢请求
export async function searchSlowRequests(thresholdMs?: number, page?: number, pageSize?: number): Promise<SearchResult> {
  const result = await invoke<string>('search_slow_requests', { thresholdMs, page, pageSize })
  return JSON.parse(result)
}

// 获取搜索建议（TODO: Rust 层暂未实现搜索建议 IPC 命令）
export async function getSearchSuggestions(keyword: string): Promise<SearchSuggestion[]> {
  return []
}

// 获取最近搜索（TODO: Rust 层暂未实现最近搜索 IPC 命令）
export async function getRecentSearches(): Promise<string[]> {
  return []
}

// 保存搜索（映射为保存过滤器）
export async function saveSearch(name: string, query: SearchQuery): Promise<{ id: string; name: string }> {
  const result = await invoke<string>('save_filter', {
    filter: JSON.stringify({ name, query: JSON.stringify(query) }),
  })
  return JSON.parse(result)
}

// 获取已保存的搜索（映射为过滤器列表）
export async function getSavedSearches(): Promise<Array<{ id: string; name: string; query: SearchQuery }>> {
  const result = await invoke<string>('list_filters')
  return JSON.parse(result)
}

// 删除已保存的搜索（映射为删除过滤器）
export async function deleteSavedSearch(id: string): Promise<void> {
  await invoke('delete_filter', { id })
}
