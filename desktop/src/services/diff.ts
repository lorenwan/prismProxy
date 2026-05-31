import api from './api'

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

// 比较两个请求
export async function compareRequests(req: DiffRequest): Promise<DiffResult> {
  return api.post('/diff/compare', req) as any
}

// 获取请求的变更历史
export async function getChangeHistory(transactionId: string): Promise<{
  transactionId: string
  changes: Array<{
    timestamp: string
    field: string
    oldValue: string
    newValue: string
  }>
}> {
  return api.get(`/diff/history/${transactionId}`) as any
}
