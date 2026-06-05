import { invoke } from '@tauri-apps/api/core'
import type { Environment } from '../types'

// 获取环境列表
export async function getEnvironments(): Promise<Environment[]> {
  const result = await invoke<string>('list_environments')
  return JSON.parse(result)
}

// 获取单个环境
export async function getEnvironment(id: string): Promise<Environment> {
  const result = await invoke<string>('get_environment', { id })
  return JSON.parse(result)
}

// 创建环境
export async function createEnvironment(env: Partial<Environment>): Promise<Environment> {
  const result = await invoke<string>('create_environment', { environment: JSON.stringify(env) })
  return JSON.parse(result)
}

// 更新环境
export async function updateEnvironment(id: string, env: Partial<Environment>): Promise<Environment> {
  const result = await invoke<string>('update_environment', { environment: JSON.stringify({ ...env, id }) })
  return JSON.parse(result)
}

// 删除环境
export async function deleteEnvironment(id: string): Promise<void> {
  await invoke('delete_environment', { id })
}

// 激活环境
export async function activateEnvironment(id: string): Promise<void> {
  await invoke('activate_environment', { id })
}

// 导出环境
export async function exportEnvironment(id: string): Promise<string> {
  const result = await invoke<string>('export_environment', { id })
  return result
}

// 导入环境
export async function importEnvironment(data: string): Promise<Environment> {
  const result = await invoke<string>('import_environment', { environmentExport: data })
  return JSON.parse(result)
}
