import { useState, useEffect } from 'react'
import { getScripts, createScript, updateScript, deleteScript, toggleScript, testScript, batchToggleScripts } from '../services/scripts'
import type { Script } from '../services/scripts'

const triggers = [
  { value: 'request', label: '请求阶段' },
  { value: 'response', label: '响应阶段' },
  { value: 'error', label: '错误阶段' },
]

const scriptTemplates = [
  {
    name: '添加自定义 Header',
    trigger: 'request' as const,
    code: `// 在请求阶段添加自定义 Header\nrequest.headers['X-Custom-Header'] = 'my-value';`,
  },
  {
    name: '修改响应体',
    trigger: 'response' as const,
    code: `// 修改响应体\nconst body = JSON.parse(response.body);\nbody.injected = true;\nresponse.body = JSON.stringify(body);`,
  },
  {
    name: '请求日志',
    trigger: 'request' as const,
    code: `// 记录请求日志\nconsole.log(\`[\${request.method}] \${request.url}\`);`,
  },
  {
    name: '延迟模拟',
    trigger: 'request' as const,
    code: `// 模拟网络延迟\nawait new Promise(r => setTimeout(r, 1000));`,
  },
  {
    name: '错误重试',
    trigger: 'error' as const,
    code: `// 错误时重试\nif (retryCount < 3) {\n  return { retry: true };\n}`,
  },
  {
    name: '响应断言',
    trigger: 'response' as const,
    code: `// 检查响应状态\nif (response.status !== 200) {\n  console.error(\`异常状态码: \${response.status}\`);\n}`,
  },
]

const emptyScript: Partial<Script> = {
  name: '',
  description: '',
  enabled: true,
  language: 'javascript',
  trigger: 'request',
  code: '',
  priority: 0,
}

export default function ScriptsPage() {
  const [scripts, setScripts] = useState<Script[]>([])
  const [selected, setSelected] = useState<Script | null>(null)
  const [editing, setEditing] = useState<Partial<Script>>(emptyScript)
  const [isNew, setIsNew] = useState(false)
  const [testResult, setTestResult] = useState<{ output: string; error?: string } | null>(null)
  const [testing, setTesting] = useState(false)
  const [showTemplates, setShowTemplates] = useState(false)
  const [deleteConfirm, setDeleteConfirm] = useState<Script | null>(null)

  useEffect(() => {
    getScripts().then(setScripts).catch(console.error)
  }, [])

  function handleSelect(script: Script) {
    setSelected(script)
    setEditing(script)
    setIsNew(false)
    setTestResult(null)
  }

  function handleNew() {
    setSelected(null)
    setEditing({ ...emptyScript })
    setIsNew(true)
    setTestResult(null)
  }

  async function handleSave() {
    try {
      if (isNew) {
        const created = await createScript(editing)
        setScripts([...scripts, created])
        setSelected(created)
        setIsNew(false)
      } else if (selected) {
        const updated = await updateScript(selected.id, editing)
        setScripts(scripts.map((s) => (s.id === selected.id ? updated : s)))
        setSelected(updated)
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : '保存脚本失败'
      console.error('保存脚本失败:', err)
      alert(message)
    }
  }

  async function handleDelete(script: Script) {
    try {
      await deleteScript(script.id)
      setScripts(scripts.filter((s) => s.id !== script.id))
      if (selected?.id === script.id) {
        setSelected(null)
        setEditing(emptyScript)
        setIsNew(false)
      }
      setDeleteConfirm(null)
    } catch (err) {
      const message = err instanceof Error ? err.message : '删除脚本失败'
      console.error('删除脚本失败:', err)
      alert(message)
    }
  }

  async function handleToggle(script: Script) {
    try {
      await toggleScript(script.id, !script.enabled)
      setScripts(scripts.map((s) => (s.id === script.id ? { ...s, enabled: !s.enabled } : s)))
    } catch (err) {
      console.error('切换脚本状态失败:', err)
    }
  }

  async function handleTest() {
    if (!selected) return
    setTesting(true)
    setTestResult(null)
    try {
      const result = await testScript(selected.id, 'test-transaction')
      setTestResult(result)
    } catch (err) {
      setTestResult({ output: '', error: String(err) })
    } finally {
      setTesting(false)
    }
  }

  async function handleBatchToggle(enabled: boolean) {
    try {
      const ids = scripts.map((s) => s.id)
      await batchToggleScripts(ids, enabled)
      setScripts(scripts.map((s) => ({ ...s, enabled })))
    } catch (err) {
      console.error('批量操作失败:', err)
    }
  }

  function useTemplate(template: typeof scriptTemplates[0]) {
    setEditing({
      ...editing,
      name: template.name,
      trigger: template.trigger,
      code: template.code,
    })
    setShowTemplates(false)
  }

  return (
    <div className="flex h-full">
      {/* 左侧脚本列表 */}
      <div className="w-72 border-r border-[var(--border)] flex flex-col bg-[var(--bg-inset)]">
        <div className="flex items-center gap-1 p-2 border-b border-[var(--border)]">
          <button onClick={handleNew} className="px-2 py-1 text-xs bg-[var(--blue)] text-white rounded hover:bg-[var(--blue)]/90" aria-label="新增脚本">
            新增
          </button>
          <button onClick={() => handleBatchToggle(true)} className="px-2 py-1 text-xs bg-[var(--hover-bg)] rounded hover:bg-[var(--hover-bg)]" aria-label="全部启用脚本">
            全部启用
          </button>
          <button onClick={() => handleBatchToggle(false)} className="px-2 py-1 text-xs bg-[var(--hover-bg)] rounded hover:bg-[var(--hover-bg)]" aria-label="全部禁用脚本">
            全部禁用
          </button>
        </div>

        <div className="flex-1 overflow-y-auto">
          {scripts.map((script) => (
            <div
              key={script.id}
              onClick={() => handleSelect(script)}
              className={`flex items-center gap-2 px-3 py-2 cursor-pointer border-b border-[var(--border-subtle)] ${
                selected?.id === script.id ? 'bg-[var(--selected-bg)]' : 'hover:bg-[var(--hover-bg)]'
              }`}
            >
              <span className="text-sm">
                {script.trigger === 'request' && '📤'}
                {script.trigger === 'response' && '📥'}
                {script.trigger === 'error' && '⚠️'}
              </span>
              <div className="flex-1 min-w-0">
                <div className="text-sm truncate">{script.name || '未命名脚本'}</div>
                <div className="text-xs text-[var(--text-tertiary)]">{triggers.find((t) => t.value === script.trigger)?.label}</div>
              </div>
              <button
                onClick={(e) => { e.stopPropagation(); handleToggle(script) }}
                className={`w-8 h-4 rounded-full transition-colors ${script.enabled ? 'bg-[var(--green)]' : 'bg-[var(--border)]'}`}
                role="switch"
                aria-checked={script.enabled}
                aria-label={script.enabled ? '禁用脚本' : '启用脚本'}
              >
                <div className={`w-3 h-3 rounded-full bg-white transition-transform ${script.enabled ? 'translate-x-4' : 'translate-x-0.5'}`} />
              </button>
            </div>
          ))}
          {scripts.length === 0 && (
            <div className="p-4 text-center text-[var(--text-tertiary)] text-sm">暂无脚本</div>
          )}
        </div>
      </div>

      {/* 右侧编辑器 */}
      <div className="flex-1 flex flex-col bg-[var(--hover-bg)] overflow-hidden">
        <div className="flex-1 flex flex-col overflow-y-auto">
          <div className="p-4 space-y-4">
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-semibold">{isNew ? '新增脚本' : '编辑脚本'}</h2>
              <button
                onClick={() => setShowTemplates(!showTemplates)}
                className="px-2 py-1 text-xs bg-[var(--purple)] text-white rounded hover:bg-[var(--purple)]/90"
              >
                模板库
              </button>
            </div>

            {/* 模板库面板 */}
            {showTemplates && (
              <div className="bg-[var(--bg-inset)] border border-[var(--border)] rounded p-3">
                <h4 className="text-sm font-medium text-[var(--purple)] mb-2">脚本模板</h4>
                <div className="grid grid-cols-2 gap-2">
                  {scriptTemplates.map((tpl, i) => (
                    <button
                      key={i}
                      onClick={() => useTemplate(tpl)}
                      className="text-left px-3 py-2 bg-[var(--hover-bg)] rounded hover:bg-[var(--selected-bg)] text-xs"
                    >
                      <div className="font-medium">{tpl.name}</div>
                      <div className="text-[var(--text-tertiary)] mt-0.5">{triggers.find((t) => t.value === tpl.trigger)?.label}</div>
                    </button>
                  ))}
                </div>
              </div>
            )}

            <div className="grid grid-cols-2 gap-4">
              <div>
                <label htmlFor="script-name" className="block text-sm text-[var(--text-tertiary)] mb-1">脚本名称</label>
                <input
                  id="script-name"
                  value={editing.name || ''}
                  onChange={(e) => setEditing({ ...editing, name: e.target.value })}
                  className="w-full px-3 py-2 bg-[var(--bg-inset)] border border-[var(--border)] rounded text-sm focus:border-[var(--blue)] focus:outline-none"
                  placeholder="输入脚本名称"
                  aria-label="脚本名称"
                />
              </div>
              <div>
                <label htmlFor="script-trigger" className="block text-sm text-[var(--text-tertiary)] mb-1">触发阶段</label>
                <select
                  id="script-trigger"
                  value={editing.trigger || 'request'}
                  onChange={(e) => setEditing({ ...editing, trigger: e.target.value as Script['trigger'] })}
                  className="w-full px-3 py-2 bg-[var(--bg-inset)] border border-[var(--border)] rounded text-sm focus:border-[var(--blue)] focus:outline-none"
                  aria-label="触发阶段"
                >
                  {triggers.map((t) => <option key={t.value} value={t.value}>{t.label}</option>)}
                </select>
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <label htmlFor="script-priority" className="block text-sm text-[var(--text-tertiary)] mb-1">优先级</label>
                <input
                  id="script-priority"
                  type="number"
                  value={editing.priority || 0}
                  onChange={(e) => setEditing({ ...editing, priority: Number(e.target.value) })}
                  className="w-32 px-3 py-2 bg-[var(--bg-inset)] border border-[var(--border)] rounded text-sm focus:border-[var(--blue)] focus:outline-none"
                  aria-label="脚本优先级"
                />
              </div>
              <div>
                <label htmlFor="script-desc" className="block text-sm text-[var(--text-tertiary)] mb-1">描述</label>
                <input
                  id="script-desc"
                  value={editing.description || ''}
                  onChange={(e) => setEditing({ ...editing, description: e.target.value })}
                  className="w-full px-3 py-2 bg-[var(--bg-inset)] border border-[var(--border)] rounded text-sm focus:border-[var(--blue)] focus:outline-none"
                  placeholder="脚本描述（可选）"
                  aria-label="脚本描述"
                />
              </div>
            </div>

            {/* 代码编辑器 */}
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <label htmlFor="script-code" className="text-sm text-[var(--text-tertiary)]">脚本代码</label>
                <span className="text-xs text-[var(--text-tertiary)]">JavaScript</span>
              </div>
              <textarea
                id="script-code"
                value={editing.code || ''}
                onChange={(e) => setEditing({ ...editing, code: e.target.value })}
                className="w-full px-3 py-2 bg-[var(--bg-inset)] border border-[var(--border)] rounded text-sm font-mono h-64 resize-none focus:border-[var(--blue)] focus:outline-none"
                placeholder="// 编写脚本代码&#10;// request: 请求对象&#10;// response: 响应对象&#10;// console.log(): 日志输出"
                spellCheck={false}
                aria-label="脚本代码"
              />
            </div>

            {/* 操作按钮 */}
            <div className="flex gap-2 pt-2">
              <button onClick={handleSave} className="px-4 py-2 bg-[var(--blue)] text-white rounded text-sm hover:bg-[var(--blue)]/90" aria-label="保存脚本">
                保存
              </button>
              {!isNew && selected && (
                <>
                  <button
                    onClick={handleTest}
                    disabled={testing}
                    className="px-4 py-2 bg-[var(--yellow)] text-white rounded text-sm hover:bg-[var(--yellow)]/90 disabled:opacity-50"
                    aria-label="测试脚本"
                  >
                    {testing ? '测试中...' : '测试脚本'}
                  </button>
                  <button onClick={() => setDeleteConfirm(selected)} className="px-4 py-2 bg-[var(--red)] text-white rounded text-sm hover:bg-[var(--red)]/90" aria-label="删除脚本">
                    删除
                  </button>
                </>
              )}
            </div>

            {/* 测试结果 */}
            {testResult && (
              <div className="bg-[var(--bg-inset)] border border-[var(--border)] rounded p-3">
                <h4 className="text-sm font-medium text-[var(--green)] mb-2">测试结果</h4>
                {testResult.error ? (
                  <pre className="text-xs font-mono text-[var(--red)] whitespace-pre-wrap">{testResult.error}</pre>
                ) : (
                  <pre className="text-xs font-mono text-[var(--text-primary)] whitespace-pre-wrap">{testResult.output || '（无输出）'}</pre>
                )}
              </div>
            )}
          </div>
        </div>
      </div>

      {/* 删除确认 */}
      {deleteConfirm && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50" onClick={() => setDeleteConfirm(null)}>
          <div className="bg-[var(--hover-bg)] border border-[var(--border)] rounded-lg p-4 w-80" onClick={(e) => e.stopPropagation()}>
            <h3 className="text-base font-semibold mb-2">确认删除</h3>
            <p className="text-sm text-[var(--text-tertiary)] mb-4">确定要删除脚本 "{deleteConfirm.name}" 吗？</p>
            <div className="flex gap-2 justify-end">
              <button onClick={() => setDeleteConfirm(null)} className="px-3 py-1.5 text-sm bg-[var(--bg-inset)] rounded hover:bg-[var(--hover-bg)]" aria-label="取消删除">取消</button>
              <button onClick={() => handleDelete(deleteConfirm)} className="px-3 py-1.5 text-sm bg-[var(--red)] text-white rounded hover:bg-[var(--red)]/90" aria-label="确认删除">删除</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
