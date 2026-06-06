import { useState, useEffect } from 'react'
import type { Rule, RuleMatch, RuleAction } from '../types'
import { getRules, createRule, updateRule, deleteRule, toggleRule, batchToggleRules } from '../services/rules'

// 匹配类型选项
const matchTypes = [
  { value: 'host', label: 'Host' },
  { value: 'path', label: 'Path (URL 正则)' },
  { value: 'method', label: 'Method' },
  { value: 'url', label: 'URL 通配符' },
]

// 动作类型选项
const actionTypes = [
  { value: 'block', label: '阻止' },
  { value: 'redirect', label: '重定向' },
  { value: 'modify', label: '修改' },
  { value: 'delay', label: '延迟' },
]

// 简化表单状态（用于 UI 编辑）
interface RuleFormState {
  name: string
  enabled: boolean
  priority: number
  matchType: string
  matchValue: string
  actionType: string
  actionValue: string
}

const emptyForm: RuleFormState = {
  name: '',
  enabled: true,
  priority: 0,
  matchType: 'host',
  matchValue: '',
  actionType: 'block',
  actionValue: '',
}

// 从 Rule 转换为表单状态
function ruleToForm(rule: Rule): RuleFormState {
  let matchType = 'host'
  let matchValue = ''
  if (rule.match?.host_pattern) {
    matchType = 'host'
    matchValue = rule.match.host_pattern
  } else if (rule.match?.url_wildcard) {
    matchType = 'url'
    matchValue = rule.match.url_wildcard
  } else if (rule.match?.url_pattern) {
    matchType = 'path'
    matchValue = rule.match.url_pattern
  } else if (rule.match?.methods?.length) {
    matchType = 'method'
    matchValue = rule.match.methods.join(',')
  }

  let actionType = 'block'
  let actionValue = ''
  if (rule.action?.type) {
    actionType = rule.action.type
    if (rule.action.type === 'redirect') {
      actionValue = rule.action.remote_url || rule.action.local_path || ''
    } else if (rule.action.type === 'block') {
      actionValue = rule.action.block_response?.body || ''
    } else if (rule.action.type === 'modify') {
      actionValue = rule.action.modify?.body_replace || ''
    } else if (rule.action.type === 'delay') {
      actionValue = String(rule.action.delay_ms || 0)
    }
  }

  return {
    name: rule.name || '',
    enabled: rule.enabled ?? true,
    priority: rule.priority || 0,
    matchType,
    matchValue,
    actionType,
    actionValue,
  }
}

// 从表单状态构建 Rule 的 match 和 action
function formToRuleMatch(form: RuleFormState): RuleMatch {
  const match: RuleMatch = {}
  switch (form.matchType) {
    case 'host':
      match.host_pattern = form.matchValue
      break
    case 'path':
      match.url_pattern = form.matchValue
      break
    case 'url':
      match.url_wildcard = form.matchValue
      break
    case 'method':
      match.methods = form.matchValue.split(',').map((s) => s.trim()).filter(Boolean)
      break
  }
  return match
}

function formToRuleAction(form: RuleFormState): RuleAction {
  const action: RuleAction = { type: form.actionType as RuleAction['type'] }
  switch (form.actionType) {
    case 'redirect':
      action.remote_url = form.actionValue
      break
    case 'block':
      action.block_response = { body: form.actionValue, status_code: 200 }
      break
    case 'modify':
      action.modify = { body_replace: form.actionValue }
      break
    case 'delay':
      action.delay_ms = parseInt(form.actionValue) || 0
      break
  }
  return action
}

export default function RulesPage() {
  const [rules, setRules] = useState<Rule[]>([])
  const [selected, setSelected] = useState<Rule | null>(null)
  const [editing, setEditing] = useState<RuleFormState>(emptyForm)
  const [isNew, setIsNew] = useState(false)

  useEffect(() => {
    getRules().then(setRules).catch(console.error)
  }, [])

  // 选中规则
  function handleSelect(rule: Rule) {
    setSelected(rule)
    setEditing(ruleToForm(rule))
    setIsNew(false)
  }

  // 新增规则
  function handleNew() {
    setSelected(null)
    setEditing({ ...emptyForm })
    setIsNew(true)
  }

  // 保存规则
  async function handleSave() {
    const ruleData: Partial<Rule> = {
      name: editing.name,
      enabled: editing.enabled,
      priority: editing.priority,
      match: formToRuleMatch(editing),
      action: formToRuleAction(editing),
    }
    if (isNew) {
      const created = await createRule(ruleData)
      setRules([...rules, created])
      setSelected(created)
      setIsNew(false)
    } else if (selected) {
      const updated = await updateRule(selected.id, ruleData)
      setRules(rules.map((r) => (r.id === selected.id ? updated : r)))
      setSelected(updated)
    }
  }

  // 删除规则
  async function handleDelete() {
    if (!selected) return
    await deleteRule(selected.id)
    setRules(rules.filter((r) => r.id !== selected.id))
    setSelected(null)
    setEditing(emptyForm)
    setIsNew(false)
  }

  // 切换启用
  async function handleToggle(rule: Rule) {
    await toggleRule(rule.id, !rule.enabled)
    setRules(rules.map((r) => (r.id === rule.id ? { ...r, enabled: !r.enabled } : r)))
  }

  // 批量操作
  async function handleBatchToggle(enabled: boolean) {
    const ids = rules.map((r) => r.id)
    await batchToggleRules(ids, enabled)
    setRules(rules.map((r) => ({ ...r, enabled })))
  }

  return (
    <div className="flex h-full">
      {/* 左侧规则列表 */}
      <div className="w-80 border-r border-[var(--border)] flex flex-col bg-[var(--bg-inset)]">
        {/* 工具栏 */}
        <div className="flex items-center gap-1 p-2 border-b border-[var(--border)]">
          <button onClick={handleNew} className="px-2 py-1 text-xs bg-[var(--blue)] text-white rounded hover:bg-[var(--blue)]/90">
            新增
          </button>
          <button onClick={() => handleBatchToggle(true)} className="px-2 py-1 text-xs bg-[var(--hover-bg)] rounded hover:bg-[var(--hover-bg)]">
            全部启用
          </button>
          <button onClick={() => handleBatchToggle(false)} className="px-2 py-1 text-xs bg-[var(--hover-bg)] rounded hover:bg-[var(--hover-bg)]">
            全部禁用
          </button>
        </div>

        {/* 规则列表 */}
        <div className="flex-1 overflow-y-auto">
          {rules.map((rule) => (
            <div
              key={rule.id}
              onClick={() => handleSelect(rule)}
              className={`flex items-center gap-2 px-3 py-2 cursor-pointer border-b border-[var(--border-subtle)] ${
                selected?.id === rule.id ? 'bg-[var(--selected-bg)]' : 'hover:bg-[var(--hover-bg)]'
              }`}
            >
              {/* 类型图标 */}
              <span className="text-sm">
                {rule.action?.type === 'block' && '🚫'}
                {rule.action?.type === 'redirect' && '↩️'}
                {rule.action?.type === 'delay' && '⏱️'}
                {rule.action?.type === 'modify' && '✏️'}
              </span>

              {/* 名称和优先级 */}
              <div className="flex-1 min-w-0">
                <div className="text-sm truncate">{rule.name || '未命名规则'}</div>
                <div className="text-xs text-[var(--text-tertiary)]">优先级: {rule.priority}</div>
              </div>

              {/* 启用开关 */}
              <button
                onClick={(e) => {
                  e.stopPropagation()
                  handleToggle(rule)
                }}
                className={`w-8 h-4 rounded-full transition-colors ${
                  rule.enabled ? 'bg-[var(--green)]' : 'bg-[var(--border)]'
                }`}
              >
                <div
                  className={`w-3 h-3 rounded-full bg-white transition-transform ${
                    rule.enabled ? 'translate-x-4' : 'translate-x-0.5'
                  }`}
                />
              </button>
            </div>
          ))}
          {rules.length === 0 && (
            <div className="p-4 text-center text-[var(--text-tertiary)] text-sm">暂无规则</div>
          )}
        </div>
      </div>

      {/* 右侧规则编辑器 */}
      <div className="flex-1 flex flex-col bg-[var(--hover-bg)] overflow-y-auto">
        <div className="p-4 space-y-4 max-w-2xl">
          <h2 className="text-lg font-semibold">{isNew ? '新增规则' : '编辑规则'}</h2>

          {/* 名称 */}
          <div>
            <label className="block text-sm text-[var(--text-tertiary)] mb-1">规则名称</label>
            <input
              value={editing.name || ''}
              onChange={(e) => setEditing({ ...editing, name: e.target.value })}
              className="w-full px-3 py-2 bg-[var(--bg-inset)] border border-[var(--border)] rounded text-sm focus:border-[var(--blue)] focus:outline-none"
              placeholder="输入规则名称"
            />
          </div>

          {/* 优先级 */}
          <div>
            <label className="block text-sm text-[var(--text-tertiary)] mb-1">优先级</label>
            <input
              type="number"
              value={editing.priority || 0}
              onChange={(e) => setEditing({ ...editing, priority: Number(e.target.value) })}
              className="w-32 px-3 py-2 bg-[var(--bg-inset)] border border-[var(--border)] rounded text-sm focus:border-[var(--blue)] focus:outline-none"
            />
          </div>

          {/* 匹配条件 */}
          <div className="space-y-2">
            <h3 className="text-sm font-medium text-[var(--blue)]">匹配条件</h3>
            <div className="flex gap-2">
              <select
                value={editing.matchType || 'host'}
                onChange={(e) => setEditing({ ...editing, matchType: e.target.value })}
                className="px-3 py-2 bg-[var(--bg-inset)] border border-[var(--border)] rounded text-sm focus:border-[var(--blue)] focus:outline-none"
              >
                {matchTypes.map((t) => (
                  <option key={t.value} value={t.value}>{t.label}</option>
                ))}
              </select>
              <input
                value={editing.matchValue || ''}
                onChange={(e) => setEditing({ ...editing, matchValue: e.target.value })}
                className="flex-1 px-3 py-2 bg-[var(--bg-inset)] border border-[var(--border)] rounded text-sm focus:border-[var(--blue)] focus:outline-none"
                placeholder="匹配值（支持正则）"
              />
            </div>
          </div>

          {/* 动作配置 */}
          <div className="space-y-2">
            <h3 className="text-sm font-medium text-[var(--blue)]">动作配置</h3>
            <select
              value={editing.actionType || 'block'}
              onChange={(e) => setEditing({ ...editing, actionType: e.target.value })}
              className="w-full px-3 py-2 bg-[var(--bg-inset)] border border-[var(--border)] rounded text-sm focus:border-[var(--blue)] focus:outline-none"
            >
              {actionTypes.map((t) => (
                <option key={t.value} value={t.value}>{t.label}</option>
              ))}
            </select>
            <textarea
              value={editing.actionValue || ''}
              onChange={(e) => setEditing({ ...editing, actionValue: e.target.value })}
              className="w-full px-3 py-2 bg-[var(--bg-inset)] border border-[var(--border)] rounded text-sm h-32 resize-none focus:border-[var(--blue)] focus:outline-none"
              placeholder={
                editing.actionType === 'redirect' ? '重定向 URL' :
                editing.actionType === 'block' ? '拦截响应内容 (JSON)' :
                editing.actionType === 'delay' ? '延迟时间 (ms)' :
                editing.actionType === 'modify' ? '修改内容 (JSON)' :
                '动作参数'
              }
            />
          </div>

          {/* 操作按钮 */}
          <div className="flex gap-2 pt-2">
            <button onClick={handleSave} className="px-4 py-2 bg-[var(--blue)] text-white rounded text-sm hover:bg-[var(--blue)]/90">
              保存
            </button>
            {!isNew && (
              <button onClick={handleDelete} className="px-4 py-2 bg-[var(--red)] text-white rounded text-sm hover:bg-[var(--red)]/90">
                删除
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
