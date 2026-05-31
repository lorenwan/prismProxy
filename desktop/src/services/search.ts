import api from './api'
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
  return api.post('/search', query) as any
}

// 获取搜索建议
export async function getSearchSuggestions(keyword: string): Promise<SearchSuggestion[]> {
  return api.get('/search/suggestions', { params: { keyword } }) as any
}

// 获取最近搜索
export async function getRecentSearches(): Promise<string[]> {
  return api.get('/search/recent') as any
}

// 保存搜索
export async function saveSearch(name: string, query: SearchQuery): Promise<{ id: string; name: string }> {
  return api.post('/search/saved', { name, query }) as any
}

// 获取已保存的搜索
export async function getSavedSearches(): Promise<Array<{ id: string; name: string; query: SearchQuery }>> {
  return api.get('/search/saved') as any
}

// 删除已保存的搜索
export async function deleteSavedSearch(id: string): Promise<void> {
  return api.delete(`/search/saved/${id}`) as any
}
