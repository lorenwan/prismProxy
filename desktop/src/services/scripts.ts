import { invoke } from '@tauri-apps/api/core'

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
  const result = await invoke<string>('list_scripts')
  return JSON.parse(result)
}

// 获取脚本详情
export async function getScript(id: string): Promise<Script> {
  const result = await invoke<string>('get_script', { id })
  return JSON.parse(result)
}

// 创建脚本
export async function createScript(script: Partial<Script>): Promise<Script> {
  const result = await invoke<string>('create_script', { script: JSON.stringify(script) })
  return JSON.parse(result)
}

// 更新脚本
export async function updateScript(id: string, script: Partial<Script>): Promise<Script> {
  const result = await invoke<string>('update_script', { script: JSON.stringify({ ...script, id }) })
  return JSON.parse(result)
}

// 删除脚本
export async function deleteScript(id: string): Promise<void> {
  await invoke('delete_script', { id })
}

// 切换脚本启用状态
export async function toggleScript(id: string, enabled: boolean): Promise<void> {
  await invoke('toggle_script', { id })
}

// 测试运行脚本
export async function testScript(id: string, transactionId: string): Promise<{ output: string; error?: string }> {
  const result = await invoke<string>('execute_script', {
    scriptId: id,
    transactionId,
    data: null,
  })
  return JSON.parse(result)
}

// 批量启用/禁用（逐个调用）
export async function batchToggleScripts(ids: string[], enabled: boolean): Promise<void> {
  for (const id of ids) {
    await invoke('toggle_script', { id })
  }
}
