import { invoke } from '@tauri-apps/api/core'
import type { Environment, EnvironmentVariable } from '../types'

// === 后端数据映射 ===
// Proto 使用 snake_case，前端使用 camelCase
// 在此服务层做数据转换

interface BackendVariable {
  id?: string
  key?: string
  value?: string
  description?: string
  enabled?: boolean
  is_secret?: boolean
}

interface BackendEnvironment {
  id?: string
  name?: string
  variables?: BackendVariable[]
  is_active?: boolean
  is_default?: boolean
  base_url?: string
  created_at?: string
  updated_at?: string
}

/** 将后端 Variable 转换为前端 EnvironmentVariable */
function mapVariable(v: BackendVariable): EnvironmentVariable {
  return {
    key: v.key ?? '',
    value: v.value ?? '',
    enabled: v.enabled ?? true,
  }
}

/** 将后端 Environment 转换为前端 Environment */
function mapEnvironment(env: BackendEnvironment): Environment {
  return {
    id: env.id ?? '',
    name: env.name ?? '',
    active: env.is_active ?? false,
    variables: (env.variables ?? []).map(mapVariable),
    createdAt: env.created_at ?? '',
    updatedAt: env.updated_at ?? '',
  }
}

// === 环境 CRUD ===

// 获取环境列表
export async function getEnvironments(): Promise<Environment[]> {
  const result = await invoke<string>('list_environments')
  const parsed = JSON.parse(result) as { environments?: BackendEnvironment[] }
  const envs = parsed.environments ?? []
  return envs.map(mapEnvironment)
}

// 获取单个环境
export async function getEnvironment(id: string): Promise<Environment> {
  const result = await invoke<string>('get_environment', { id })
  return mapEnvironment(JSON.parse(result) as BackendEnvironment)
}

// 创建环境
export async function createEnvironment(env: Partial<Environment>): Promise<Environment> {
  // 将前端格式转换为后端格式
  const backendData: Partial<BackendEnvironment> = {
    name: env.name,
    variables: env.variables?.map((v) => ({
      key: v.key,
      value: v.value,
      enabled: v.enabled,
    })),
  }
  const result = await invoke<string>('create_environment', { environment: JSON.stringify(backendData) })
  return mapEnvironment(JSON.parse(result) as BackendEnvironment)
}

// 更新环境
export async function updateEnvironment(id: string, env: Partial<Environment>): Promise<Environment> {
  const backendData: Partial<BackendEnvironment> = {
    id,
    name: env.name,
    variables: env.variables?.map((v) => ({
      key: v.key,
      value: v.value,
      enabled: v.enabled,
    })),
  }
  const result = await invoke<string>('update_environment', { environment: JSON.stringify(backendData) })
  return mapEnvironment(JSON.parse(result) as BackendEnvironment)
}

// 删除环境
export async function deleteEnvironment(id: string): Promise<void> {
  await invoke('delete_environment', { id })
}

// 激活环境 - 使用后端返回的环境数据更新前端状态
export async function activateEnvironment(id: string): Promise<Environment> {
  const result = await invoke<string>('activate_environment', { id })
  return mapEnvironment(JSON.parse(result) as BackendEnvironment)
}

// 导出环境
export async function exportEnvironment(id: string): Promise<string> {
  const result = await invoke<string>('export_environment', { id })
  return result
}

// 导入环境
export async function importEnvironment(data: string): Promise<Environment> {
  const result = await invoke<string>('import_environment', { environmentExport: data })
  return mapEnvironment(JSON.parse(result) as BackendEnvironment)
}
