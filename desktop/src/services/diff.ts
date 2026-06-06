import { invoke } from '@tauri-apps/api/core'
import { getTraffic } from './traffic'

export interface DiffResult {
  leftId: string
  rightId: string
  requestDiff: DiffSection[]
  responseDiff: DiffSection[]
  summary: {
    requestChanges: number
    responseChanges: number
  }
}

export interface DiffSection {
  type: 'equal' | 'added' | 'removed' | 'modified'
  path: string
  left?: string
  right?: string
}

export interface DiffRequest {
  leftId: string
  rightId: string
  compareHeaders?: boolean
  compareBody?: boolean
  ignoreHeaders?: string[]
}

// Proto DiffStatus 枚举值映射
// DIFF_STATUS_UNSPECIFIED = 0, ADDED = 1, REMOVED = 2, MODIFIED = 3, UNCHANGED = 4
const DIFF_STATUS_MAP: Record<number, DiffSection['type']> = {
  0: 'equal',      // UNSPECIFIED 视为 equal
  1: 'added',
  2: 'removed',
  3: 'modified',
  4: 'equal',      // UNCHANGED
}

function mapDiffStatus(status: number): DiffSection['type'] {
  return DIFF_STATUS_MAP[status] ?? 'equal'
}

// 从 DiffResultProto {type, entries} 提取并映射 entries
function mapDiffEntries(result: any): DiffSection[] {
  const entries = result?.entries ?? []
  return entries.map((e: any) => ({
    type: mapDiffStatus(e.status),
    path: e.path ?? '',
    left: e.left ?? '',
    right: e.right ?? '',
  }))
}

// 比较 Headers
export async function compareHeaders(
  left: Record<string, string>,
  right: Record<string, string>
): Promise<DiffSection[]> {
  const result = await invoke<string>('compare_headers', { left, right })
  return mapDiffEntries(JSON.parse(result))
}

// 比较 Body
export async function compareBody(left: string, right: string): Promise<DiffSection[]> {
  const result = await invoke<string>('compare_body', { left, right })
  return mapDiffEntries(JSON.parse(result))
}

// 比较 JSON - 返回 JSONDiffResultProto {diffs, summary}
export async function compareJson(left: string, right: string): Promise<DiffSection[]> {
  const result = await invoke<string>('compare_json', { left, right })
  const parsed = JSON.parse(result)
  const diffs = parsed?.diffs ?? []
  return diffs.map((e: any) => ({
    type: mapDiffStatus(e.status),
    path: e.path ?? '',
    left: e.left ?? '',
    right: e.right ?? '',
  }))
}

// 比较 Query 参数
export async function compareQuery(
  left: Record<string, string>,
  right: Record<string, string>
): Promise<DiffSection[]> {
  const result = await invoke<string>('compare_query', { left, right })
  return mapDiffEntries(JSON.parse(result))
}

// 将 StringList headers 转换为 Record<string, string>（取第一个值）
function flattenHeaders(headers: Record<string, string | { values: string[] }>): Record<string, string> {
  const result: Record<string, string> = {}
  for (const [key, value] of Object.entries(headers)) {
    if (typeof value === 'string') {
      result[key] = value
    } else if (value?.values?.length) {
      result[key] = value.values[0]
    }
  }
  return result
}

// 比较两个请求（高级接口，组合多个比较结果）
// 需要先获取两个请求的详细数据再调用底层比较函数
export async function compareRequests(req: DiffRequest): Promise<DiffResult> {
  const shouldCompareHeaders = req.compareHeaders ?? true
  const shouldCompareBody = req.compareBody ?? true

  // 获取左右请求的详细数据
  const [leftTraffic, rightTraffic] = await Promise.all([
    getTraffic(req.leftId),
    getTraffic(req.rightId),
  ])

  // 过滤掉需要忽略的 headers
  const filterHeaders = (headers: Record<string, string>) => {
    if (!req.ignoreHeaders?.length) return headers
    const filtered = { ...headers }
    for (const h of req.ignoreHeaders) {
      delete filtered[h]
    }
    return filtered
  }

  const requestDiff: DiffSection[] = []
  const responseDiff: DiffSection[] = []

  // 比较请求 headers（traffic 的 headers 是 StringList 格式，需先展平）
  if (shouldCompareHeaders) {
    const leftReqHeaders = filterHeaders(flattenHeaders(leftTraffic.request?.headers ?? {}))
    const rightReqHeaders = filterHeaders(flattenHeaders(rightTraffic.request?.headers ?? {}))
    const reqHeaderDiff = await compareHeaders(leftReqHeaders, rightReqHeaders)
    requestDiff.push(...reqHeaderDiff)
  }

  // 比较请求 body
  if (shouldCompareBody) {
    const leftReqBody = leftTraffic.request?.body ?? ''
    const rightReqBody = rightTraffic.request?.body ?? ''
    if (leftReqBody !== rightReqBody) {
      const reqBodyDiff = await compareBody(leftReqBody, rightReqBody)
      requestDiff.push(...reqBodyDiff)
    }
  }

  // 比较响应 headers
  if (shouldCompareHeaders) {
    const leftResHeaders = filterHeaders(flattenHeaders(leftTraffic.response?.headers ?? {}))
    const rightResHeaders = filterHeaders(flattenHeaders(rightTraffic.response?.headers ?? {}))
    const resHeaderDiff = await compareHeaders(leftResHeaders, rightResHeaders)
    responseDiff.push(...resHeaderDiff)
  }

  // 比较响应 body
  if (shouldCompareBody) {
    const leftResBody = leftTraffic.response?.body ?? ''
    const rightResBody = rightTraffic.response?.body ?? ''
    if (leftResBody !== rightResBody) {
      const resBodyDiff = await compareBody(leftResBody, rightResBody)
      responseDiff.push(...resBodyDiff)
    }
  }

  return {
    leftId: req.leftId,
    rightId: req.rightId,
    requestDiff,
    responseDiff,
    summary: {
      requestChanges: requestDiff.filter((d) => d.type !== 'equal').length,
      responseChanges: responseDiff.filter((d) => d.type !== 'equal').length,
    },
  }
}

// 获取请求的变更历史（TODO: Rust 层暂未实现变更历史 IPC 命令）
export async function getChangeHistory(transactionId: string): Promise<{
  transactionId: string
  changes: Array<{
    timestamp: string
    field: string
    oldValue: string
    newValue: string
  }>
}> {
  return { transactionId, changes: [] }
}
