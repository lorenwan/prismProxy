import { invoke } from '@tauri-apps/api/core'

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
  leftValue?: string
  rightValue?: string
}

export interface DiffRequest {
  leftId: string
  rightId: string
  compareHeaders?: boolean
  compareBody?: boolean
  ignoreHeaders?: string[]
}

// 比较 Headers
export async function compareHeaders(
  left: Record<string, string>,
  right: Record<string, string>
): Promise<DiffSection[]> {
  const result = await invoke<string>('compare_headers', { left, right })
  return JSON.parse(result)
}

// 比较 Body
export async function compareBody(left: string, right: string): Promise<DiffSection[]> {
  const result = await invoke<string>('compare_body', { left, right })
  return JSON.parse(result)
}

// 比较 JSON
export async function compareJson(left: string, right: string): Promise<DiffSection[]> {
  const result = await invoke<string>('compare_json', { left, right })
  return JSON.parse(result)
}

// 比较 Query 参数
export async function compareQuery(
  left: Record<string, string>,
  right: Record<string, string>
): Promise<DiffSection[]> {
  const result = await invoke<string>('compare_query', { left, right })
  return JSON.parse(result)
}

// 比较两个请求（高级接口，组合多个比较结果）
// 注意：需要先获取两个请求的详细数据再调用底层比较函数
export async function compareRequests(req: DiffRequest): Promise<DiffResult> {
  // 调用底层比较接口，由调用方提供具体数据
  // 这里提供一个占位实现，实际使用时需要根据具体数据调用 compareHeaders/compareBody 等
  const requestDiff: DiffSection[] = []
  const responseDiff: DiffSection[] = []

  if (req.compareHeaders) {
    // TODO: 需要获取左右请求的 headers 数据
  }
  if (req.compareBody) {
    // TODO: 需要获取左右请求的 body 数据
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
