import { useState, useEffect } from 'react'
import type { Rule } from '../types'
import { getRules, createRule, updateRule, deleteRule, toggleRule, batchToggleRules } from '../services/rules'

// 匹配类型选项
const matchTypes = [
  { value: 'host', label: 'Host' },
  { value: 'path', label: 'Path' },
  { value: 'method', label: 'Method' },
  { value: 'url', label: 'URL' },
  { value: 'header', label: 'Header' },
  { value: 'body', label: 'Body' },
]

// 动作类型选项
const actionTypes = [
  { value: 'block', label: '阻止' },
  { value: 'redirect', label: '重定向' },
  { value: 'modify_request', label: '修改请求' },
  { value: 'modify_response', label: '修改响应' },
  { value: 'delay', label: '延迟' },
  { value: 'mock', label: 'Mock 响应' },
]

// 空规则模板
const emptyRule: Partial<Rule> = {
  name: '',
  enabled: true,
  priority: 0,
  matchType: 'host',
  matchValue: '',
  actionType: 'block',
  actionValue: '',
}

export default function RulesPage() {
  const [rules, setRules] = useState<Rule[]>([])
  const [selected, setSelected] = useState<Rule | null>(null)
  const [editing, setEditing] = useState<Partial<Rule>>(emptyRule)
  const [isNew, setIsNew] = useState(false)

  useEffect(() => {
    getRules().then(setRules).catch(console.error)
  }, [])

  // 选中规则
  function handleSelect(rule: Rule) {
    setSelected(rule)
    setEditing(rule)
    setIsNew(false)
  }

  // 新增规则
  function handleNew() {
    setSelected(null)
    setEditing({ ...emptyRule })
    setIsNew(true)
  }

  // 保存规则
  async function handleSave() {
    if (isNew) {
      const created = await createRule(editing)
      setRules([...rules, created])
      setSelected(created)
      setIsNew(false)
    } else if (selected) {
      const updated = await updateRule(selected.id, editing)
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
    setEditing(emptyRule)
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
      <div className="w-80 border-r border-[#3b4261] flex flex-col bg-[#1a1b26]">
        {/* 工具栏 */}
        <div className="flex items-center gap-1 p-2 border-b border-[#3b4261]">
          <button onClick={handleNew} className="px-2 py-1 text-xs bg-[#7aa2f7] text-[#1a1b26] rounded hover:bg-[#89b4fa]">
            新增
          </button>
          <button onClick={() => handleBatchToggle(true)} className="px-2 py-1 text-xs bg-[#24283b] rounded hover:bg-[#3b4261]">
            全部启用
          </button>
          <button onClick={() => handleBatchToggle(false)} className="px-2 py-1 text-xs bg-[#24283b] rounded hover:bg-[#3b4261]">
            全部禁用
          </button>
        </div>

        {/* 规则列表 */}
        <div className="flex-1 overflow-y-auto">
          {rules.map((rule) => (
            <div
              key={rule.id}
              onClick={() => handleSelect(rule)}
              className={`flex items-center gap-2 px-3 py-2 cursor-pointer border-b border-[#24283b] ${
                selected?.id === rule.id ? 'bg-[#283457]' : 'hover:bg-[#24283b]'
              }`}
            >
              {/* 类型图标 */}
              <span className="text-sm">
                {rule.actionType === 'block' && '🚫'}
                {rule.actionType === 'redirect' && '↩️'}
                {rule.actionType === 'mock' && '📄'}
                {rule.actionType === 'delay' && '⏱️'}
                {rule.actionType === 'modify_request' && '✏️'}
                {rule.actionType === 'modify_response' && '✏️'}
              </span>

              {/* 名称和优先级 */}
              <div className="flex-1 min-w-0">
                <div className="text-sm truncate">{rule.name || '未命名规则'}</div>
                <div className="text-xs text-[#565f89]">优先级: {rule.priority}</div>
              </div>

              {/* 启用开关 */}
              <button
                onClick={(e) => {
                  e.stopPropagation()
                  handleToggle(rule)
                }}
                className={`w-8 h-4 rounded-full transition-colors ${
                  rule.enabled ? 'bg-[#9ece6a]' : 'bg-[#3b4261]'
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
            <div className="p-4 text-center text-[#565f89] text-sm">暂无规则</div>
          )}
        </div>
      </div>

      {/* 右侧规则编辑器 */}
      <div className="flex-1 flex flex-col bg-[#24283b] overflow-y-auto">
        <div className="p-4 space-y-4 max-w-2xl">
          <h2 className="text-lg font-semibold">{isNew ? '新增规则' : '编辑规则'}</h2>

          {/* 名称 */}
          <div>
            <label className="block text-sm text-[#565f89] mb-1">规则名称</label>
            <input
              value={editing.name || ''}
              onChange={(e) => setEditing({ ...editing, name: e.target.value })}
              className="w-full px-3 py-2 bg-[#1a1b26] border border-[#3b4261] rounded text-sm focus:border-[#7aa2f7] focus:outline-none"
              placeholder="输入规则名称"
            />
          </div>

          {/* 优先级 */}
          <div>
            <label className="block text-sm text-[#565f89] mb-1">优先级</label>
            <input
              type="number"
              value={editing.priority || 0}
              onChange={(e) => setEditing({ ...editing, priority: Number(e.target.value) })}
              className="w-32 px-3 py-2 bg-[#1a1b26] border border-[#3b4261] rounded text-sm focus:border-[#7aa2f7] focus:outline-none"
            />
          </div>

          {/* 匹配条件 */}
          <div className="space-y-2">
            <h3 className="text-sm font-medium text-[#7aa2f7]">匹配条件</h3>
            <div className="flex gap-2">
              <select
                value={editing.matchType || 'host'}
                onChange={(e) => setEditing({ ...editing, matchType: e.target.value as Rule['matchType'] })}
                className="px-3 py-2 bg-[#1a1b26] border border-[#3b4261] rounded text-sm focus:border-[#7aa2f7] focus:outline-none"
              >
                {matchTypes.map((t) => (
                  <option key={t.value} value={t.value}>{t.label}</option>
                ))}
              </select>
              <input
                value={editing.matchValue || ''}
                onChange={(e) => setEditing({ ...editing, matchValue: e.target.value })}
                className="flex-1 px-3 py-2 bg-[#1a1b26] border border-[#3b4261] rounded text-sm focus:border-[#7aa2f7] focus:outline-none"
                placeholder="匹配值（支持正则）"
              />
            </div>
          </div>

          {/* 动作配置 */}
          <div className="space-y-2">
            <h3 className="text-sm font-medium text-[#7aa2f7]">动作配置</h3>
            <select
              value={editing.actionType || 'block'}
              onChange={(e) => setEditing({ ...editing, actionType: e.target.value as Rule['actionType'] })}
              className="w-full px-3 py-2 bg-[#1a1b26] border border-[#3b4261] rounded text-sm focus:border-[#7aa2f7] focus:outline-none"
            >
              {actionTypes.map((t) => (
                <option key={t.value} value={t.value}>{t.label}</option>
              ))}
            </select>
            <textarea
              value={editing.actionValue || ''}
              onChange={(e) => setEditing({ ...editing, actionValue: e.target.value })}
              className="w-full px-3 py-2 bg-[#1a1b26] border border-[#3b4261] rounded text-sm h-32 resize-none focus:border-[#7aa2f7] focus:outline-none"
              placeholder={
                editing.actionType === 'redirect' ? '重定向 URL' :
                editing.actionType === 'mock' ? 'Mock 响应内容 (JSON)' :
                editing.actionType === 'delay' ? '延迟时间 (ms)' :
                '动作参数'
              }
            />
          </div>

          {/* 操作按钮 */}
          <div className="flex gap-2 pt-2">
            <button onClick={handleSave} className="px-4 py-2 bg-[#7aa2f7] text-[#1a1b26] rounded text-sm hover:bg-[#89b4fa]">
              保存
            </button>
            {!isNew && (
              <button onClick={handleDelete} className="px-4 py-2 bg-[#f7768e] text-[#1a1b26] rounded text-sm hover:bg-[#ff9eaf]">
                删除
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
