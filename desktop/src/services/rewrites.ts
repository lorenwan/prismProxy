import api from './api'
import type { RewriteRule } from '../types'

// 获取重写规则列表
export async function getRewrites(): Promise<RewriteRule[]> {
  return api.get('/rewrites') as any
}

// 创建重写规则
export async function createRewrite(rule: Partial<RewriteRule>): Promise<RewriteRule> {
  return api.post('/rewrites', rule) as any
}

// 更新重写规则
export async function updateRewrite(id: string, rule: Partial<RewriteRule>): Promise<RewriteRule> {
  return api.put(`/rewrites/${id}`, rule) as any
}

// 删除重写规则
export async function deleteRewrite(id: string): Promise<void> {
  return api.delete(`/rewrites/${id}`) as any
}

// 切换启用状态
export async function toggleRewrite(id: string, enabled: boolean): Promise<void> {
  return api.put(`/rewrites/${id}`, { enabled }) as any
}

// 批量启用/禁用
export async function batchToggleRewrites(ids: string[], enabled: boolean): Promise<void> {
  return api.put('/rewrites/batch', { ids, enabled }) as any
}
