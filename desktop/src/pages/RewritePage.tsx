import { useState, useEffect } from 'react'
import { getRewrites, createRewrite, updateRewrite, deleteRewrite, toggleRewrite, batchToggleRewrites } from '../services/rewrites'
import type { RewriteRule } from '../types'

const rewriteTypes = [
  { value: 'add_header', label: '添加请求头' },
  { value: 'remove_header', label: '删除请求头' },
  { value: 'replace_header', label: '替换请求头' },
  { value: 'replace_body', label: '替换响应体' },
  { value: 'replace_url', label: '替换 URL' },
  { value: 'map_local', label: 'Map Local' },
  { value: 'map_remote', label: 'Map Remote' },
]

const matchTypes = [
  { value: 'host', label: 'Host' },
  { value: 'path', label: 'Path' },
  { value: 'url', label: 'URL' },
  { value: 'method', label: 'Method' },
]

const emptyRule: Partial<RewriteRule> = {
  name: '',
  enabled: true,
  type: 'add_header',
  matchType: 'host',
  matchValue: '',
  actionKey: '',
  actionValue: '',
  priority: 0,
}

export default function RewritePage() {
  const [rules, setRules] = useState<RewriteRule[]>([])
  const [selected, setSelected] = useState<RewriteRule | null>(null)
  const [editing, setEditing] = useState<Partial<RewriteRule>>(emptyRule)
  const [isNew, setIsNew] = useState(false)
  const [deleteConfirm, setDeleteConfirm] = useState<RewriteRule | null>(null)

  useEffect(() => {
    getRewrites().then(setRules).catch(console.error)
  }, [])

  function handleSelect(rule: RewriteRule) {
    setSelected(rule)
    setEditing(rule)
    setIsNew(false)
  }

  function handleNew() {
    setSelected(null)
    setEditing({ ...emptyRule })
    setIsNew(true)
  }

  async function handleSave() {
    if (isNew) {
      const created = await createRewrite(editing)
      setRules([...rules, created])
      setSelected(created)
      setIsNew(false)
    } else if (selected) {
      const updated = await updateRewrite(selected.id, editing)
      setRules(rules.map((r) => (r.id === selected.id ? updated : r)))
      setSelected(updated)
    }
  }

  async function handleDelete(rule: RewriteRule) {
    await deleteRewrite(rule.id)
    setRules(rules.filter((r) => r.id !== rule.id))
    if (selected?.id === rule.id) {
      setSelected(null)
      setEditing(emptyRule)
      setIsNew(false)
    }
    setDeleteConfirm(null)
  }

  async function handleToggle(rule: RewriteRule) {
    await toggleRewrite(rule.id, !rule.enabled)
    setRules(rules.map((r) => (r.id === rule.id ? { ...r, enabled: !r.enabled } : r)))
  }

  async function handleBatchToggle(enabled: boolean) {
    const ids = rules.map((r) => r.id)
    await batchToggleRewrites(ids, enabled)
    setRules(rules.map((r) => ({ ...r, enabled })))
  }

  function getTypeLabel(type: string) {
    return rewriteTypes.find((t) => t.value === type)?.label || type
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
              <span className="text-sm">
                {rule.type === 'add_header' && '➕'}
                {rule.type === 'remove_header' && '➖'}
                {rule.type === 'replace_header' && '🔄'}
                {rule.type === 'replace_body' && '📝'}
                {rule.type === 'replace_url' && '🔗'}
                {rule.type === 'map_local' && '📂'}
                {rule.type === 'map_remote' && '🌐'}
              </span>
              <div className="flex-1 min-w-0">
                <div className="text-sm truncate">{rule.name || '未命名规则'}</div>
                <div className="text-xs text-[var(--text-tertiary)]">{getTypeLabel(rule.type)}</div>
              </div>
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
            <div className="p-4 text-center text-[var(--text-tertiary)] text-sm">暂无重写规则</div>
          )}
        </div>
      </div>

      {/* 右侧编辑器 */}
      <div className="flex-1 flex flex-col bg-[var(--hover-bg)] overflow-y-auto">
        <div className="p-4 space-y-4 max-w-2xl">
          <h2 className="text-lg font-semibold">{isNew ? '新增重写规则' : '编辑重写规则'}</h2>

          <div>
            <label className="block text-sm text-[var(--text-tertiary)] mb-1">规则名称</label>
            <input
              value={editing.name || ''}
              onChange={(e) => setEditing({ ...editing, name: e.target.value })}
              className="w-full px-3 py-2 bg-[var(--bg-inset)] border border-[var(--border)] rounded text-sm focus:border-[var(--blue)] focus:outline-none"
              placeholder="输入规则名称"
            />
          </div>

          <div>
            <label className="block text-sm text-[var(--text-tertiary)] mb-1">优先级</label>
            <input
              type="number"
              value={editing.priority || 0}
              onChange={(e) => setEditing({ ...editing, priority: Number(e.target.value) })}
              className="w-32 px-3 py-2 bg-[var(--bg-inset)] border border-[var(--border)] rounded text-sm focus:border-[var(--blue)] focus:outline-none"
            />
          </div>

          {/* 规则类型 */}
          <div className="space-y-2">
            <h3 className="text-sm font-medium text-[var(--blue)]">规则类型</h3>
            <select
              value={editing.type || 'add_header'}
              onChange={(e) => setEditing({ ...editing, type: e.target.value as RewriteRule['type'] })}
              className="w-full px-3 py-2 bg-[var(--bg-inset)] border border-[var(--border)] rounded text-sm focus:border-[var(--blue)] focus:outline-none"
            >
              {rewriteTypes.map((t) => (
                <option key={t.value} value={t.value}>{t.label}</option>
              ))}
            </select>
          </div>

          {/* 匹配条件 */}
          <div className="space-y-2">
            <h3 className="text-sm font-medium text-[var(--blue)]">匹配条件</h3>
            <div className="flex gap-2">
              <select
                value={editing.matchType || 'host'}
                onChange={(e) => setEditing({ ...editing, matchType: e.target.value as RewriteRule['matchType'] })}
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
            {(editing.type === 'add_header' || editing.type === 'replace_header') && (
              <input
                value={editing.actionKey || ''}
                onChange={(e) => setEditing({ ...editing, actionKey: e.target.value })}
                className="w-full px-3 py-2 bg-[var(--bg-inset)] border border-[var(--border)] rounded text-sm focus:border-[var(--blue)] focus:outline-none"
                placeholder="Header Name"
              />
            )}
            {editing.type === 'remove_header' && (
              <input
                value={editing.actionKey || ''}
                onChange={(e) => setEditing({ ...editing, actionKey: e.target.value })}
                className="w-full px-3 py-2 bg-[var(--bg-inset)] border border-[var(--border)] rounded text-sm focus:border-[var(--blue)] focus:outline-none"
                placeholder="要删除的 Header Name"
              />
            )}
            <textarea
              value={editing.actionValue || ''}
              onChange={(e) => setEditing({ ...editing, actionValue: e.target.value })}
              className="w-full px-3 py-2 bg-[var(--bg-inset)] border border-[var(--border)] rounded text-sm h-32 resize-none focus:border-[var(--blue)] focus:outline-none font-mono"
              placeholder={
                editing.type === 'add_header' ? 'Header Value' :
                editing.type === 'replace_header' ? '新的 Header Value' :
                editing.type === 'replace_body' ? '新的响应体内容 (JSON)' :
                editing.type === 'replace_url' ? '新的 URL' :
                editing.type === 'map_local' ? '本地文件路径' :
                '远程 URL'
              }
            />
          </div>

          {/* 操作按钮 */}
          <div className="flex gap-2 pt-2">
            <button onClick={handleSave} className="px-4 py-2 bg-[var(--blue)] text-white rounded text-sm hover:bg-[var(--blue)]/90">
              保存
            </button>
            {!isNew && selected && (
              <button onClick={() => setDeleteConfirm(selected)} className="px-4 py-2 bg-[var(--red)] text-white rounded text-sm hover:bg-[var(--red)]/90">
                删除
              </button>
            )}
          </div>
        </div>
      </div>

      {/* 删除确认对话框 */}
      {deleteConfirm && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50" onClick={() => setDeleteConfirm(null)}>
          <div className="bg-[var(--hover-bg)] border border-[var(--border)] rounded-lg p-4 w-80" onClick={(e) => e.stopPropagation()}>
            <h3 className="text-base font-semibold mb-2">确认删除</h3>
            <p className="text-sm text-[var(--text-tertiary)] mb-4">确定要删除规则 "{deleteConfirm.name}" 吗？此操作不可撤销。</p>
            <div className="flex gap-2 justify-end">
              <button onClick={() => setDeleteConfirm(null)} className="px-3 py-1.5 text-sm bg-[var(--bg-inset)] rounded hover:bg-[var(--hover-bg)]">
                取消
              </button>
              <button onClick={() => handleDelete(deleteConfirm)} className="px-3 py-1.5 text-sm bg-[var(--red)] text-white rounded hover:bg-[var(--red)]/90">
                删除
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
