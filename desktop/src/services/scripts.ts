import api from './api'

export interface Script {
  id: string
  name: string
  description: string
  enabled: boolean
  language: 'javascript' | 'lua'
  trigger: 'request' | 'response' | 'error'
  code: string
  priority: number
  hitCount: number
  createdAt: string
  updatedAt: string
}

// 获取脚本列表
export async function getScripts(): Promise<Script[]> {
  return api.get('/scripts') as any
}

// 获取脚本详情
export async function getScript(id: string): Promise<Script> {
  return api.get(`/scripts/${id}`) as any
}

// 创建脚本
export async function createScript(script: Partial<Script>): Promise<Script> {
  return api.post('/scripts', script) as any
}

// 更新脚本
export async function updateScript(id: string, script: Partial<Script>): Promise<Script> {
  return api.put(`/scripts/${id}`, script) as any
}

// 删除脚本
export async function deleteScript(id: string): Promise<void> {
  return api.delete(`/scripts/${id}`) as any
}

// 切换脚本启用状态
export async function toggleScript(id: string, enabled: boolean): Promise<void> {
  return api.put(`/scripts/${id}`, { enabled }) as any
}

// 测试运行脚本
export async function testScript(id: string, transactionId: string): Promise<{ output: string; error?: string }> {
  return api.post(`/scripts/${id}/test`, { transactionId }) as any
}

// 批量启用/禁用
export async function batchToggleScripts(ids: string[], enabled: boolean): Promise<void> {
  return api.put('/scripts/batch', { ids, enabled }) as any
}
