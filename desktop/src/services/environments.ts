import api from './api'
import type { Environment } from '../types'

// 获取环境列表
export async function getEnvironments(): Promise<Environment[]> {
  return api.get('/environments') as any
}

// 创建环境
export async function createEnvironment(env: Partial<Environment>): Promise<Environment> {
  return api.post('/environments', env) as any
}

// 更新环境
export async function updateEnvironment(id: string, env: Partial<Environment>): Promise<Environment> {
  return api.put(`/environments/${id}`, env) as any
}

// 删除环境
export async function deleteEnvironment(id: string): Promise<void> {
  return api.delete(`/environments/${id}`) as any
}

// 激活环境
export async function activateEnvironment(id: string): Promise<void> {
  return api.put(`/environments/${id}/activate`) as any
}
