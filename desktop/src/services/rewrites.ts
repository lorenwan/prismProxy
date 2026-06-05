import { invoke } from '@tauri-apps/api/core'
import type { RewriteRule } from '../types'

// 获取重写规则列表
export async function getRewrites(): Promise<RewriteRule[]> {
  const result = await invoke<string>('list_rewrites')
  return JSON.parse(result)
}

// 获取单条重写规则
export async function getRewrite(id: string): Promise<RewriteRule> {
  const result = await invoke<string>('get_rewrite', { id })
  return JSON.parse(result)
}

// 创建重写规则
export async function createRewrite(rule: Partial<RewriteRule>): Promise<RewriteRule> {
  const result = await invoke<string>('create_rewrite', { rewrite: JSON.stringify(rule) })
  return JSON.parse(result)
}

// 更新重写规则
export async function updateRewrite(id: string, rule: Partial<RewriteRule>): Promise<RewriteRule> {
  const result = await invoke<string>('update_rewrite', { rewrite: JSON.stringify({ ...rule, id }) })
  return JSON.parse(result)
}

// 删除重写规则
export async function deleteRewrite(id: string): Promise<void> {
  await invoke('delete_rewrite', { id })
}

// 切换启用状态
export async function toggleRewrite(id: string, enabled: boolean): Promise<void> {
  await invoke('toggle_rewrite', { id, enabled })
}

// 批量启用/禁用（逐个调用）
export async function batchToggleRewrites(ids: string[], enabled: boolean): Promise<void> {
  for (const id of ids) {
    await invoke('toggle_rewrite', { id, enabled })
  }
}
