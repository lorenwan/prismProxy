import { invoke } from '@tauri-apps/api/core'
import type { Collection, CollectionRequest } from '../types'

// 获取集合列表
export async function getCollections(): Promise<Collection[]> {
  const result = await invoke<string>('list_collections')
  return JSON.parse(result)
}

// 获取集合详情
export async function getCollection(id: string): Promise<Collection> {
  const result = await invoke<string>('get_collection', { id })
  return JSON.parse(result)
}

// 创建集合
export async function createCollection(collection: Partial<Collection>): Promise<Collection> {
  const result = await invoke<string>('create_collection', { collection: JSON.stringify(collection) })
  return JSON.parse(result)
}

// 更新集合
export async function updateCollection(id: string, collection: Partial<Collection>): Promise<Collection> {
  const result = await invoke<string>('update_collection', { collection: JSON.stringify({ ...collection, id }) })
  return JSON.parse(result)
}

// 删除集合
export async function deleteCollection(id: string): Promise<void> {
  await invoke('delete_collection', { id })
}

// 添加请求到集合
export async function addRequest(collectionId: string, request: Partial<CollectionRequest>): Promise<CollectionRequest> {
  const result = await invoke<string>('add_collection_request', {
    collectionId,
    parentItemId: null,
    request: JSON.stringify(request),
  })
  return JSON.parse(result)
}

// 更新请求
export async function updateRequest(collectionId: string, requestId: string, request: Partial<CollectionRequest>): Promise<CollectionRequest> {
  const result = await invoke<string>('update_collection_request', {
    collectionId,
    itemId: requestId,
    request: JSON.stringify(request),
  })
  return JSON.parse(result)
}

// 删除请求
export async function deleteRequest(collectionId: string, requestId: string): Promise<void> {
  await invoke('delete_collection_request', { collectionId, itemId: requestId })
}

// 发送请求
export async function sendRequest(request: { method: string; url: string; headers: Record<string, string>; body: string }): Promise<{
  status: number
  statusText: string
  headers: Record<string, string>
  body: string
  duration: number
}> {
  const result = await invoke<string>('execute_collection_request', {
    request: JSON.stringify(request),
    environmentId: null,
  })
  return JSON.parse(result)
}

// 导出集合（TODO: Rust 层暂未实现集合导出 IPC 命令）
export async function exportCollection(id: string): Promise<Blob> {
  console.warn('exportCollection: 暂未实现 Tauri IPC')
  return new Blob([])
}

// 导入集合（TODO: Rust 层暂未实现集合导入 IPC 命令）
export async function importCollection(file: File): Promise<Collection> {
  console.warn('importCollection: 暂未实现 Tauri IPC')
  return { id: '', name: file.name, description: '', requests: [], createdAt: '', updatedAt: '' }
}
