import { invoke } from '@tauri-apps/api/core'
import type { Collection, CollectionRequest } from '../types'

// === 后端数据映射 ===
// Proto 使用 snake_case 和嵌套结构，前端使用 camelCase 和扁平结构
// 在此服务层做数据转换，使前端组件可以正确访问数据

interface BackendKeyValue {
  key: string
  value: string
  description?: string
  enabled?: boolean
}

interface BackendRequestBody {
  type?: string
  content?: string
  binary?: string
  graphql?: { query: string; variables?: string }
}

interface BackendApiRequest {
  id?: string
  name?: string
  method?: string
  url?: string
  headers?: BackendKeyValue[]
  query_params?: BackendKeyValue[]
  body?: BackendRequestBody
  auth?: { type?: string; config?: Record<string, string> }
  description?: string
  created_at?: string
  updated_at?: string
}

interface BackendCollectionItem {
  id?: string
  type?: string
  name?: string
  request?: BackendApiRequest
  items?: BackendCollectionItem[]
  created_at?: string
  updated_at?: string
}

interface BackendCollection {
  id?: string
  name?: string
  description?: string
  parent_id?: string
  items?: BackendCollectionItem[]
  created_at?: string
  updated_at?: string
}

interface BackendExecutionResult {
  request_id?: string
  status?: number
  status_text?: string
  headers?: BackendKeyValue[]
  body?: string
  content_type?: string
  duration_ms?: number
  size?: number
  error?: string
}

/** 将 KeyValue[] 转换为 Record<string, string> */
function keyValueArrayToRecord(arr?: BackendKeyValue[]): Record<string, string> {
  if (!arr) return {}
  const result: Record<string, string> = {}
  for (const kv of arr) {
    if (kv.enabled !== false) {
      result[kv.key] = kv.value
    }
  }
  return result
}

/** 将后端 CollectionItem 转换为前端 CollectionRequest */
function mapItemToRequest(item: BackendCollectionItem, collectionId: string): CollectionRequest {
  const req = item.request
  return {
    id: item.id ?? '',
    name: item.name ?? req?.name ?? '',
    method: req?.method ?? 'GET',
    url: req?.url ?? '',
    headers: keyValueArrayToRecord(req?.headers),
    body: req?.body?.content ?? '',
    contentType: req?.body?.type ?? 'none',
    collectionId,
    createdAt: item.created_at ?? req?.created_at ?? '',
    updatedAt: item.updated_at ?? req?.updated_at ?? '',
  }
}

/** 将后端 Collection 转换为前端 Collection */
function mapCollection(col: BackendCollection): Collection {
  const requests: CollectionRequest[] = []
  if (col.items) {
    for (const item of col.items) {
      // 只处理 request 类型的条目（跳过 folder）
      if (item.type === 'request') {
        requests.push(mapItemToRequest(item, col.id ?? ''))
      }
      // folder 类型的子条目中可能也包含 request，递归提取
      if (item.items) {
        for (const child of item.items) {
          if (child.type === 'request') {
            requests.push(mapItemToRequest(child, col.id ?? ''))
          }
        }
      }
    }
  }

  return {
    id: col.id ?? '',
    name: col.name ?? '',
    description: col.description ?? '',
    requests,
    createdAt: col.created_at ?? '',
    updatedAt: col.updated_at ?? '',
  }
}

/** 将前端 headers Record 转换为后端 KeyValue[] */
function recordToKeyValueArray(record?: Record<string, string>): BackendKeyValue[] {
  if (!record) return []
  return Object.entries(record).map(([key, value]) => ({ key, value, enabled: true }))
}

/** 将前端简单请求结构转换为后端 APIRequest 结构 */
function buildApiRequest(request: {
  method: string
  url: string
  headers: Record<string, string>
  body: string
  name?: string
}): BackendApiRequest {
  return {
    name: request.name ?? '',
    method: request.method,
    url: request.url,
    headers: recordToKeyValueArray(request.headers),
    body: {
      type: request.body ? 'raw' : 'none',
      content: request.body,
    },
  }
}

// === 集合 CRUD ===

// 获取集合列表
export async function getCollections(): Promise<Collection[]> {
  const result = await invoke<string>('list_collections')
  const parsed = JSON.parse(result) as { collections?: BackendCollection[] }
  const collections = parsed.collections ?? []
  return collections.map(mapCollection)
}

// 获取集合详情
export async function getCollection(id: string): Promise<Collection> {
  const result = await invoke<string>('get_collection', { id })
  return mapCollection(JSON.parse(result) as BackendCollection)
}

// 创建集合
export async function createCollection(collection: Partial<Collection>): Promise<Collection> {
  // 将前端格式转换为后端格式
  const backendData: Partial<BackendCollection> = {
    name: collection.name,
    description: collection.description,
  }
  const result = await invoke<string>('create_collection', { collection: JSON.stringify(backendData) })
  return mapCollection(JSON.parse(result) as BackendCollection)
}

// 更新集合
export async function updateCollection(id: string, collection: Partial<Collection>): Promise<Collection> {
  const backendData: Partial<BackendCollection> = {
    id,
    name: collection.name,
    description: collection.description,
  }
  const result = await invoke<string>('update_collection', { collection: JSON.stringify(backendData) })
  return mapCollection(JSON.parse(result) as BackendCollection)
}

// 删除集合
export async function deleteCollection(id: string): Promise<void> {
  await invoke('delete_collection', { id })
}

// === 请求 CRUD ===

// 添加请求到集合
export async function addRequest(collectionId: string, request: Partial<CollectionRequest>): Promise<CollectionRequest> {
  const apiRequest: BackendApiRequest = {
    name: request.name ?? '',
    method: request.method ?? 'GET',
    url: request.url ?? '',
    headers: recordToKeyValueArray(request.headers),
    body: {
      type: request.contentType ?? (request.body ? 'raw' : 'none'),
      content: request.body ?? '',
    },
  }
  const result = await invoke<string>('add_collection_request', {
    collectionId,
    parentItemId: null,
    request: JSON.stringify(apiRequest),
  })
  // 返回的是 CollectionItem，需要转换
  const item = JSON.parse(result) as BackendCollectionItem
  return mapItemToRequest(item, collectionId)
}

// 更新请求
export async function updateRequest(collectionId: string, requestId: string, request: Partial<CollectionRequest>): Promise<CollectionRequest> {
  const apiRequest: BackendApiRequest = {
    id: requestId,
    name: request.name ?? '',
    method: request.method ?? 'GET',
    url: request.url ?? '',
    headers: recordToKeyValueArray(request.headers),
    body: {
      type: request.contentType ?? (request.body ? 'raw' : 'none'),
      content: request.body ?? '',
    },
  }
  const result = await invoke<string>('update_collection_request', {
    collectionId,
    itemId: requestId,
    request: JSON.stringify(apiRequest),
  })
  // 返回的是 APIRequest，构造 CollectionRequest
  const returned = JSON.parse(result) as BackendApiRequest
  return {
    id: requestId,
    name: returned.name ?? request.name ?? '',
    method: returned.method ?? request.method ?? '',
    url: returned.url ?? request.url ?? '',
    headers: keyValueArrayToRecord(returned.headers) ?? request.headers ?? {},
    body: returned.body?.content ?? request.body ?? '',
    contentType: returned.body?.type ?? request.contentType ?? 'none',
    collectionId,
    createdAt: returned.created_at ?? '',
    updatedAt: returned.updated_at ?? '',
  }
}

// 删除请求
export async function deleteRequest(collectionId: string, requestId: string): Promise<void> {
  await invoke('delete_collection_request', { collectionId, itemId: requestId })
}

// 发送/执行请求
export async function sendRequest(request: {
  method: string
  url: string
  headers: Record<string, string>
  body: string
  name?: string
}): Promise<{
  status: number
  statusText: string
  headers: Record<string, string>
  body: string
  duration: number
}> {
  const apiRequest = buildApiRequest(request)
  const result = await invoke<string>('execute_collection_request', {
    request: JSON.stringify(apiRequest),
    environmentId: null,
  })
  const parsed = JSON.parse(result) as BackendExecutionResult
  return {
    status: parsed.status ?? 0,
    statusText: parsed.status_text ?? '',
    headers: keyValueArrayToRecord(parsed.headers),
    body: parsed.body ?? '',
    duration: parsed.duration_ms ?? 0,
  }
}

// === 导入导出（TODO: 待完整实现） ===

// 导出集合
// TODO: 当前实现可能不完整，需要验证后端 ExportCollection 的返回格式
export async function exportCollection(id: string): Promise<Blob> {
  const result = await invoke<string>('export_collection', { id })
  const parsed = JSON.parse(result)
  const data = parsed.data ?? result
  if (typeof data === 'string') {
    return new Blob([data], { type: 'application/json' })
  }
  return new Blob([JSON.stringify(parsed)], { type: 'application/json' })
}

// 导入集合
// TODO: 当前实现可能不完整，需要验证后端 ImportCollection 的请求格式
export async function importCollection(file: File): Promise<Collection> {
  const text = await file.text()
  const result = await invoke<string>('import_collection', {
    data: text,
    filename: file.name,
  })
  return mapCollection(JSON.parse(result) as BackendCollection)
}
