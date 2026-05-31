import api from './api'
import type { Transaction, TrafficStats } from '../types'

// 获取流量列表
export const getTrafficList = (params?: {
  page?: number
  pageSize?: number
  method?: string
  status?: number
  host?: string
}) => {
  return api.get<{ data: Transaction[]; total: number }>('/traffic', { params })
}

// 获取流量详情
export const getTrafficDetail = (id: string) => {
  return api.get<Transaction>(`/traffic/${id}`)
}

// 删除流量记录
export const deleteTraffic = (id: string) => {
  return api.delete(`/traffic/${id}`)
}

// 清空流量
export const clearTraffic = () => {
  return api.delete('/traffic')
}

// 获取流量统计
export const getTrafficStats = () => {
  return api.get<TrafficStats>('/traffic/stats')
}

// 更新书签
export const updateBookmark = (id: string, bookmarked: boolean) => {
  return api.put(`/traffic/${id}/bookmark`, { bookmarked })
}

// 更新备注
export const updateNotes = (id: string, notes: string) => {
  return api.put(`/traffic/${id}/notes`, { notes })
}

// 更新颜色标记
export const updateColor = (id: string, color: string) => {
  return api.put(`/traffic/${id}/color`, { color })
}

// 更新标签
export const updateTags = (id: string, tags: string[]) => {
  return api.put(`/traffic/${id}/tags`, { tags })
}