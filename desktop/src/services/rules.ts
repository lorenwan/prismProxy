import api from './api'
import type { Rule } from '../types'

// 获取规则列表
export async function getRules(): Promise<Rule[]> {
  return api.get('/rules') as any
}

// 获取规则详情
export async function getRule(id: string): Promise<Rule> {
  return api.get(`/rules/${id}`) as any
}

// 创建规则
export async function createRule(rule: Partial<Rule>): Promise<Rule> {
  return api.post('/rules', rule) as any
}

// 更新规则
export async function updateRule(id: string, rule: Partial<Rule>): Promise<Rule> {
  return api.put(`/rules/${id}`, rule) as any
}

// 删除规则
export async function deleteRule(id: string): Promise<void> {
  return api.delete(`/rules/${id}`) as any
}

// 切换规则启用状态
export async function toggleRule(id: string, enabled: boolean): Promise<void> {
  return api.put(`/rules/${id}`, { enabled }) as any
}

// 批量启用/禁用
export async function batchToggleRules(ids: string[], enabled: boolean): Promise<void> {
  return api.put('/rules/batch', { ids, enabled }) as any
}
