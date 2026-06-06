import { useState, useEffect } from 'react'
import type { Breakpoint, BreakpointSession, RuleMatch, BreakAction } from '../types'
import {
  getBreakpoints,
  createBreakpoint,
  updateBreakpoint,
  deleteBreakpoint,
  toggleBreakpoint,
  getActiveSessions,
  resumeSession,
} from '../services/breakpoints'

// 简化表单状态（用于 UI 编辑）
interface BreakpointFormState {
  name: string
  enabled: boolean
  phase: 'request' | 'response'
  matchType: string
  matchValue: string
}

const emptyForm: BreakpointFormState = {
  name: '',
  enabled: true,
  phase: 'request',
  matchType: 'host',
  matchValue: '',
}

// 从 Breakpoint 转换为表单状态
function bpToForm(bp: Breakpoint): BreakpointFormState {
  let matchType = 'host'
  let matchValue = ''
  if (bp.match?.host_pattern) {
    matchType = 'host'
    matchValue = bp.match.host_pattern
  } else if (bp.match?.url_wildcard) {
    matchType = 'url'
    matchValue = bp.match.url_wildcard
  } else if (bp.match?.url_pattern) {
    matchType = 'path'
    matchValue = bp.match.url_pattern
  }
  return {
    name: bp.name || '',
    enabled: bp.enabled ?? true,
    phase: bp.phase || 'request',
    matchType,
    matchValue,
  }
}

// 从表单状态构建 RuleMatch
function formToMatch(form: BreakpointFormState): RuleMatch {
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
  }
  return match
}

export default function BreakpointsPage() {
  const [breakpoints, setBreakpoints] = useState<Breakpoint[]>([])
  const [sessions, setSessions] = useState<BreakpointSession[]>([])
  const [selected, setSelected] = useState<Breakpoint | null>(null)
  const [editing, setEditing] = useState<BreakpointFormState>(emptyForm)
  const [isNew, setIsNew] = useState(false)

  useEffect(() => {
    getBreakpoints().then(setBreakpoints).catch(console.error)
    getActiveSessions().then(setSessions).catch(console.error)
  }, [])

  // 选中断点
  function handleSelect(bp: Breakpoint) {
    setSelected(bp)
    setEditing(bpToForm(bp))
    setIsNew(false)
  }

  // 新增
  function handleNew() {
    setSelected(null)
    setEditing({ ...emptyForm })
    setIsNew(true)
  }

  // 保存
  async function handleSave() {
    try {
      const bpData: Partial<Breakpoint> = {
        name: editing.name,
        enabled: editing.enabled,
        phase: editing.phase,
        match: formToMatch(editing),
        action: { type: 'pause' },
      }
      if (isNew) {
        const created = await createBreakpoint(bpData)
        setBreakpoints([...breakpoints, created])
        setSelected(created)
        setIsNew(false)
      } else if (selected) {
        const updated = await updateBreakpoint(selected.id, bpData)
        setBreakpoints(breakpoints.map((b) => (b.id === selected.id ? updated : b)))
        setSelected(updated)
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : '保存断点失败'
      console.error('保存断点失败:', err)
      alert(message)
    }
  }

  // 删除
  async function handleDelete() {
    if (!selected) return
    try {
      await deleteBreakpoint(selected.id)
      setBreakpoints(breakpoints.filter((b) => b.id !== selected.id))
      setSelected(null)
      setEditing(emptyForm)
      setIsNew(false)
    } catch (err) {
      const message = err instanceof Error ? err.message : '删除断点失败'
      console.error('删除断点失败:', err)
      alert(message)
    }
  }

  // 切换启用
  async function handleToggle(bp: Breakpoint) {
    try {
      await toggleBreakpoint(bp.id, !bp.enabled)
      setBreakpoints(breakpoints.map((b) => (b.id === bp.id ? { ...b, enabled: !b.enabled } : b)))
    } catch (err) {
      console.error('切换断点状态失败:', err)
    }
  }

  // 恢复会话
  async function handleResume(sessionId: string) {
    try {
      await resumeSession(sessionId)
      setSessions(sessions.filter((s) => s.id !== sessionId))
    } catch (err) {
      const message = err instanceof Error ? err.message : '恢复会话失败'
      console.error('恢复会话失败:', err)
      alert(message)
    }
  }

  return (
    <div className="flex h-full">
      {/* 左侧断点列表 */}
      <div className="w-80 border-r border-[var(--border)] flex flex-col bg-[var(--bg-inset)]">
        {/* 工具栏 */}
        <div className="flex items-center gap-1 p-2 border-b border-[var(--border)]">
          <button onClick={handleNew} className="px-2 py-1 text-xs bg-[var(--blue)] text-white rounded hover:bg-[var(--blue)]/90" aria-label="新增断点">
            新增
          </button>
        </div>

        {/* 断点列表 */}
        <div className="flex-1 overflow-y-auto">
          {breakpoints.map((bp) => (
            <div
              key={bp.id}
              onClick={() => handleSelect(bp)}
              className={`flex items-center gap-2 px-3 py-2 cursor-pointer border-b border-[var(--border-subtle)] ${
                selected?.id === bp.id ? 'bg-[var(--selected-bg)]' : 'hover:bg-[var(--hover-bg)]'
              }`}
            >
              {/* 阶段图标 */}
              <span className="text-sm">{bp.phase === 'request' ? '📥' : '📤'}</span>

              {/* 信息 */}
              <div className="flex-1 min-w-0">
                <div className="text-sm truncate">{bp.name || '未命名断点'}</div>
                <div className="text-xs text-[var(--text-tertiary)]">
                  {bp.match?.host_pattern ? `host: ${bp.match.host_pattern}` :
                   bp.match?.url_pattern ? `path: ${bp.match.url_pattern}` :
                   bp.match?.url_wildcard ? `url: ${bp.match.url_wildcard}` :
                   '未设置匹配'} · 命中 {bp.hitCount} 次
                </div>
              </div>

              {/* 启用开关 */}
              <button
                onClick={(e) => {
                  e.stopPropagation()
                  handleToggle(bp)
                }}
                className={`w-8 h-4 rounded-full transition-colors ${
                  bp.enabled ? 'bg-[var(--green)]' : 'bg-[var(--border)]'
                }`}
                role="switch"
                aria-checked={bp.enabled}
                aria-label={bp.enabled ? '禁用断点' : '启用断点'}
              >
                <div
                  className={`w-3 h-3 rounded-full bg-white transition-transform ${
                    bp.enabled ? 'translate-x-4' : 'translate-x-0.5'
                  }`}
                />
              </button>
            </div>
          ))}
          {breakpoints.length === 0 && (
            <div className="p-4 text-center text-[var(--text-tertiary)] text-sm">暂无断点</div>
          )}
        </div>
      </div>

      {/* 右侧编辑器和会话 */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* 断点编辑器 */}
        <div className="flex-1 overflow-y-auto bg-[var(--hover-bg)] p-4">
          <div className="max-w-2xl space-y-4">
            <h2 className="text-lg font-semibold">{isNew ? '新增断点' : '编辑断点'}</h2>

            {/* 名称 */}
            <div>
              <label htmlFor="bp-name" className="block text-sm text-[var(--text-tertiary)] mb-1">断点名称</label>
              <input
                id="bp-name"
                value={editing.name || ''}
                onChange={(e) => setEditing({ ...editing, name: e.target.value })}
                className="w-full px-3 py-2 bg-[var(--bg-inset)] border border-[var(--border)] rounded text-sm focus:border-[var(--blue)] focus:outline-none"
                placeholder="输入断点名称"
                aria-label="断点名称"
              />
            </div>

            {/* 阶段 */}
            <div>
              <label htmlFor="bp-phase" className="block text-sm text-[var(--text-tertiary)] mb-1">断点阶段</label>
              <select
                id="bp-phase"
                value={editing.phase || 'request'}
                onChange={(e) => setEditing({ ...editing, phase: e.target.value as Breakpoint['phase'] })}
                className="w-full px-3 py-2 bg-[var(--bg-inset)] border border-[var(--border)] rounded text-sm focus:border-[var(--blue)] focus:outline-none"
                aria-label="断点阶段"
              >
                <option value="request">请求阶段</option>
                <option value="response">响应阶段</option>
              </select>
            </div>

            {/* 匹配条件 */}
            <div className="space-y-2">
              <h3 className="text-sm font-medium text-[var(--blue)]">匹配条件</h3>
              <div className="flex gap-2">
                <select
                  value={editing.matchType || 'path'}
                  onChange={(e) => setEditing({ ...editing, matchType: e.target.value })}
                  className="px-3 py-2 bg-[var(--bg-inset)] border border-[var(--border)] rounded text-sm focus:border-[var(--blue)] focus:outline-none"
                  aria-label="匹配类型"
                >
                  <option value="host">Host</option>
                  <option value="path">Path</option>
                  <option value="url">URL</option>
                </select>
                <input
                  value={editing.matchValue || ''}
                  onChange={(e) => setEditing({ ...editing, matchValue: e.target.value })}
                  className="flex-1 px-3 py-2 bg-[var(--bg-inset)] border border-[var(--border)] rounded text-sm focus:border-[var(--blue)] focus:outline-none"
                  placeholder="匹配值（支持正则）"
                  aria-label="匹配值"
                />
              </div>
            </div>

            {/* 操作按钮 */}
            <div className="flex gap-2 pt-2">
              <button onClick={handleSave} className="px-4 py-2 bg-[var(--blue)] text-white rounded text-sm hover:bg-[var(--blue)]/90" aria-label="保存断点">
                保存
              </button>
              {!isNew && (
                <button onClick={handleDelete} className="px-4 py-2 bg-[var(--red)] text-white rounded text-sm hover:bg-[var(--red)]/90" aria-label="删除断点">
                  删除
                </button>
              )}
            </div>
          </div>
        </div>

        {/* 活跃会话面板 */}
        {sessions.length > 0 && (
          <div className="h-48 border-t border-[var(--border)] bg-[var(--bg-inset)] flex flex-col">
            <div className="px-3 py-2 border-b border-[var(--border)] text-sm font-medium">
              活跃断点会话 ({sessions.length})
            </div>
            <div className="flex-1 overflow-y-auto">
              {sessions.map((s) => (
                <div key={s.id} className="flex items-center gap-3 px-3 py-2 border-b border-[var(--border-subtle)]">
                  <span className="text-sm">{s.phase === 'request' ? '📥' : '📤'}</span>
                  <div className="flex-1 min-w-0">
                    <div className="text-sm truncate">会话 {s.id.slice(0, 8)}</div>
                    <div className="text-xs text-[var(--text-tertiary)]">状态: {s.status}</div>
                  </div>
                  <button
                    onClick={() => handleResume(s.id)}
                    className="px-2 py-1 text-xs bg-[var(--green)] text-white rounded hover:bg-[var(--green)]/90"
                    aria-label={`恢复会话 ${s.id.slice(0, 8)}`}
                  >
                    恢复
                  </button>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
