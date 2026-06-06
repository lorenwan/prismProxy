import { invoke } from '@tauri-apps/api/core'
import type { RewriteRule } from '../types'

// Proto 嵌套结构类型
interface ProtoRuleMatch {
  url_pattern?: string
  url_wildcard?: string
  host_pattern?: string
  methods?: string[]
  header_match?: { name: string; value: string; match_type: string }
  content_type?: string[]
}

interface ProtoRewriteAction {
  type: number
  where: number
  key?: string
  value?: string
  target?: string
}

interface ProtoRewriteRule {
  id?: string
  name?: string
  enabled?: boolean
  priority?: number
  match?: ProtoRuleMatch
  actions?: ProtoRewriteAction[]
  created_at?: string
  updated_at?: string
}

// RewriteType 枚举映射
const REWRITE_TYPE_TO_PROTO: Record<string, number> = {
  'add_header': 1,
  'remove_header': 2,
  'replace_header': 5,
  'replace_body': 6,
  'replace_url': 7,
  'map_local': 8,
  'map_remote': 9,
}

const PROTO_TO_REWRITE_TYPE: Record<number, string> = {
  1: 'add_header',
  2: 'remove_header',
  3: 'replace_header',  // REPLACE (兼容)
  4: 'replace_header',  // SET (兼容)
  5: 'replace_header',
  6: 'replace_body',
  7: 'replace_url',
  8: 'map_local',
  9: 'map_remote',
}

// RewriteWhere 枚举映射
const MATCH_TYPE_TO_WHERE: Record<string, number> = {
  'host': 1,      // REQUEST_HEADER (用于主机匹配)
  'path': 6,      // URL_PATH
  'url': 1,       // REQUEST_HEADER (用于 URL 匹配)
  'method': 1,    // REQUEST_HEADER (用于方法匹配)
}

// 将前端扁平结构转换为 Proto 嵌套结构
function toProtoRewrite(rule: Partial<RewriteRule>): ProtoRewriteRule {
  const proto: ProtoRewriteRule = {
    id: rule.id,
    name: rule.name,
    enabled: rule.enabled,
    priority: rule.priority,
  }

  // 转换 match
  if (rule.matchType && rule.matchValue) {
    const match: ProtoRuleMatch = {}
    switch (rule.matchType) {
      case 'host':
        match.host_pattern = rule.matchValue
        break
      case 'path':
        match.url_pattern = rule.matchValue
        break
      case 'url':
        match.url_pattern = rule.matchValue
        break
      case 'method':
        match.methods = [rule.matchValue]
        break
    }
    proto.match = match
  }

  // 转换 actions
  if (rule.type) {
    const action: ProtoRewriteAction = {
      type: REWRITE_TYPE_TO_PROTO[rule.type] || 0,
      where: MATCH_TYPE_TO_WHERE[rule.matchType || 'url'] || 1,
      key: rule.actionKey,
      value: rule.actionValue,
    }
    proto.actions = [action]
  }

  return proto
}

// 将 Proto 嵌套结构转换为前端扁平结构
function fromProtoRewrite(proto: ProtoRewriteRule): RewriteRule {
  const rule: Partial<RewriteRule> = {
    id: proto.id || '',
    name: proto.name || '',
    enabled: proto.enabled ?? false,
    priority: proto.priority ?? 0,
    createdAt: proto.created_at || '',
    updatedAt: proto.updated_at || '',
  }

  // 转换 match
  if (proto.match) {
    if (proto.match.host_pattern) {
      rule.matchType = 'host'
      rule.matchValue = proto.match.host_pattern
    } else if (proto.match.url_pattern) {
      rule.matchType = 'url'
      rule.matchValue = proto.match.url_pattern
    } else if (proto.match.url_wildcard) {
      rule.matchType = 'url'
      rule.matchValue = proto.match.url_wildcard
    } else if (proto.match.methods && proto.match.methods.length > 0) {
      rule.matchType = 'method'
      rule.matchValue = proto.match.methods[0]
    }
  }

  // 转换 actions (取第一个)
  if (proto.actions && proto.actions.length > 0) {
    const action = proto.actions[0]
    rule.type = PROTO_TO_REWRITE_TYPE[action.type] as RewriteRule['type'] || 'replace_header'
    rule.actionKey = action.key || ''
    rule.actionValue = action.value || action.target || ''
  }

  return rule as RewriteRule
}

// 获取重写规则列表
export async function getRewrites(): Promise<RewriteRule[]> {
  const result = await invoke<string>('list_rewrites')
  const response = JSON.parse(result)
  const rules = response.rules || []
  return rules.map(fromProtoRewrite)
}

// 获取单条重写规则
export async function getRewrite(id: string): Promise<RewriteRule> {
  const result = await invoke<string>('get_rewrite', { id })
  return fromProtoRewrite(JSON.parse(result))
}

// 创建重写规则
export async function createRewrite(rule: Partial<RewriteRule>): Promise<RewriteRule> {
  const protoRule = toProtoRewrite(rule)
  const result = await invoke<string>('create_rewrite', { rewrite: JSON.stringify(protoRule) })
  return fromProtoRewrite(JSON.parse(result))
}

// 更新重写规则
export async function updateRewrite(id: string, rule: Partial<RewriteRule>): Promise<RewriteRule> {
  const protoRule = toProtoRewrite({ ...rule, id })
  const result = await invoke<string>('update_rewrite', { rewrite: JSON.stringify(protoRule) })
  return fromProtoRewrite(JSON.parse(result))
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
