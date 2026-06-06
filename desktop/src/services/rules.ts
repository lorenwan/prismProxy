import { invoke } from '@tauri-apps/api/core'
import type { Rule, RuleStats } from '../types'

// 获取规则列表
export async function getRules(): Promise<Rule[]> {
  const result = await invoke<string>('list_rules')
  return JSON.parse(result)
}

// 获取规则详情
export async function getRule(id: string): Promise<Rule> {
  const result = await invoke<string>('get_rule', { id })
  return JSON.parse(result)
}

// 创建规则
export async function createRule(rule: Partial<Rule>): Promise<Rule> {
  const result = await invoke<string>('create_rule', { rule: JSON.stringify(rule) })
  return JSON.parse(result)
}

// 更新规则
export async function updateRule(id: string, rule: Partial<Rule>): Promise<Rule> {
  const result = await invoke<string>('update_rule', { rule: JSON.stringify({ ...rule, id }) })
  return JSON.parse(result)
}

// 删除规则
export async function deleteRule(id: string): Promise<void> {
  await invoke('delete_rule', { id })
}

// 切换规则启用状态（使用专用 Toggle RPC，避免反序列化完整 Rule）
export async function toggleRule(id: string, enabled: boolean): Promise<void> {
  await invoke('toggle_rule', { id, enabled })
}

// 批量启用/禁用
export async function batchToggleRules(ids: string[], enabled: boolean): Promise<void> {
  for (const id of ids) {
    await invoke('toggle_rule', { id, enabled })
  }
}

// 获取规则统计
export async function getRuleStats(): Promise<RuleStats> {
  const result = await invoke<string>('get_rule_stats')
  return JSON.parse(result)
}
