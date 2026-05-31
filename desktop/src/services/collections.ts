import api from './api'
import type { Collection, CollectionRequest } from '../types'

// 获取集合列表
export async function getCollections(): Promise<Collection[]> {
  return api.get('/collections') as any
}

// 获取集合详情
export async function getCollection(id: string): Promise<Collection> {
  return api.get(`/collections/${id}`) as any
}

// 创建集合
export async function createCollection(collection: Partial<Collection>): Promise<Collection> {
  return api.post('/collections', collection) as any
}

// 更新集合
export async function updateCollection(id: string, collection: Partial<Collection>): Promise<Collection> {
  return api.put(`/collections/${id}`, collection) as any
}

// 删除集合
export async function deleteCollection(id: string): Promise<void> {
  return api.delete(`/collections/${id}`) as any
}

// 添加请求到集合
export async function addRequest(collectionId: string, request: Partial<CollectionRequest>): Promise<CollectionRequest> {
  return api.post(`/collections/${collectionId}/requests`, request) as any
}

// 更新请求
export async function updateRequest(collectionId: string, requestId: string, request: Partial<CollectionRequest>): Promise<CollectionRequest> {
  return api.put(`/collections/${collectionId}/requests/${requestId}`, request) as any
}

// 删除请求
export async function deleteRequest(collectionId: string, requestId: string): Promise<void> {
  return api.delete(`/collections/${collectionId}/requests/${requestId}`) as any
}

// 发送请求
export async function sendRequest(request: { method: string; url: string; headers: Record<string, string>; body: string }): Promise<{
  status: number
  statusText: string
  headers: Record<string, string>
  body: string
  duration: number
}> {
  return api.post('/collections/send', request) as any
}

// 导出集合
export async function exportCollection(id: string): Promise<Blob> {
  return api.get(`/collections/${id}/export`, { responseType: 'blob' }) as any
}

// 导入集合
export async function importCollection(file: File): Promise<Collection> {
  const formData = new FormData()
  formData.append('file', file)
  return api.post('/collections/import', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  }) as any
}
