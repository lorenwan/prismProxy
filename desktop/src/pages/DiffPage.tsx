import { useState } from 'react'
import { compareJson, compareBody, type DiffSection, type DiffResult } from '../services/diff'

export default function DiffPage() {
  const [leftInput, setLeftInput] = useState('')
  const [rightInput, setRightInput] = useState('')
  const [result, setResult] = useState<DiffResult | null>(null)
  const [comparing, setComparing] = useState(false)
  const [error, setError] = useState('')
  const [mode, setMode] = useState<'text' | 'json'>('json')

  async function handleCompare() {
    if (!leftInput.trim() || !rightInput.trim()) {
      setError('请输入两侧内容')
      return
    }
    setError('')
    setComparing(true)
    setResult(null)

    try {
      let leftContent = leftInput
      let rightContent = rightInput

      if (mode === 'json') {
        // 校验并格式化 JSON
        leftContent = JSON.stringify(JSON.parse(leftInput))
        rightContent = JSON.stringify(JSON.parse(rightInput))
      }

      let diffSections: DiffSection[]
      if (mode === 'json') {
        diffSections = await compareJson(leftContent, rightContent)
      } else {
        diffSections = await compareBody(leftContent, rightContent)
      }

      setResult({
        leftId: 'inline-left',
        rightId: 'inline-right',
        requestDiff: diffSections,
        responseDiff: [],
        summary: {
          requestChanges: diffSections.filter(d => d.type !== 'equal').length,
          responseChanges: 0,
        },
      })
    } catch (err) {
      if (err instanceof SyntaxError) {
        setError('JSON 格式错误，请检查输入')
      } else {
        setError(err instanceof Error ? err.message : '对比失败')
      }
    } finally {
      setComparing(false)
    }
  }

  function handleSwap() {
    const temp = leftInput
    setLeftInput(rightInput)
    setRightInput(temp)
    setResult(null)
  }

  function handleClear() {
    setLeftInput('')
    setRightInput('')
    setResult(null)
    setError('')
  }

  function getDiffColor(type: DiffSection['type']) {
    switch (type) {
      case 'added': return 'bg-[var(--green)]/15 text-[var(--green)] border-l-2 border-[var(--green)]'
      case 'removed': return 'bg-[var(--red)]/15 text-[var(--red)] border-l-2 border-[var(--red)]'
      case 'modified': return 'bg-[var(--yellow)]/15 text-[var(--yellow)] border-l-2 border-[var(--yellow)]'
      default: return 'text-[var(--text-tertiary)]'
    }
  }

  function getDiffLabel(type: DiffSection['type']) {
    switch (type) {
      case 'added': return '+'
      case 'removed': return '-'
      case 'modified': return '~'
      default: return ' '
    }
  }

  const displaySections = result
    ? [...result.requestDiff, ...result.responseDiff]
    : []

  return (
    <div className="flex flex-col h-full bg-[var(--hover-bg)]">
      {/* 工具栏 */}
      <div className="flex items-center gap-2 px-4 py-2 border-b border-[var(--border)]">
        <div className="flex items-center gap-1 bg-[var(--bg-inset)] rounded p-0.5">
          <button
            onClick={() => setMode('json')}
            className={`px-2 py-1 text-xs rounded ${mode === 'json' ? 'bg-[var(--blue)] text-white' : 'text-[var(--text-tertiary)] hover:text-[var(--text-primary)]'}`}
            aria-label="JSON 模式"
          >
            JSON
          </button>
          <button
            onClick={() => setMode('text')}
            className={`px-2 py-1 text-xs rounded ${mode === 'text' ? 'bg-[var(--blue)] text-white' : 'text-[var(--text-tertiary)] hover:text-[var(--text-primary)]'}`}
            aria-label="Text 模式"
          >
            Text
          </button>
        </div>
        <button onClick={handleCompare} disabled={comparing} className="px-3 py-1 text-xs bg-[var(--blue)] text-white rounded hover:bg-[var(--blue)]/90 disabled:opacity-50" aria-label="对比">
          {comparing ? '对比中...' : '对比'}
        </button>
        <button onClick={handleSwap} className="px-3 py-1 text-xs bg-[var(--hover-bg)] border border-[var(--border)] rounded hover:bg-[var(--hover-bg)]" aria-label="交换左右内容">
          交换
        </button>
        <button onClick={handleClear} className="px-3 py-1 text-xs bg-[var(--hover-bg)] border border-[var(--border)] rounded hover:bg-[var(--hover-bg)]" aria-label="清空输入">
          清空
        </button>
        {result && (
          <div className="ml-auto flex gap-3 text-xs text-[var(--text-tertiary)]">
            <span>请求变更: <span className="text-[var(--yellow)]">{result.summary.requestChanges}</span></span>
            <span>响应变更: <span className="text-[var(--yellow)]">{result.summary.responseChanges}</span></span>
          </div>
        )}
      </div>

      {error && (
        <div className="px-4 py-2 bg-[var(--red)]/10 border-b border-[var(--red)]/30 text-sm text-[var(--red)]">{error}</div>
      )}

      {/* 输入区域 */}
      <div className="flex flex-1 overflow-hidden">
        {/* 左侧输入 */}
        <div className="w-1/2 flex flex-col border-r border-[var(--border)]">
          <div className="px-3 py-1.5 border-b border-[var(--border)] text-xs text-[var(--text-tertiary)]">左侧（原始）</div>
          <textarea
            value={leftInput}
            onChange={(e) => setLeftInput(e.target.value)}
            className="flex-1 px-3 py-2 bg-[var(--bg-inset)] text-xs font-mono resize-none focus:outline-none"
            placeholder={mode === 'json' ? '{\n  "key": "value"\n}' : '输入文本内容...'}
            spellCheck={false}
            aria-label="左侧内容（原始）"
          />
        </div>

        {/* 右侧输入 */}
        <div className="w-1/2 flex flex-col">
          <div className="px-3 py-1.5 border-b border-[var(--border)] text-xs text-[var(--text-tertiary)]">右侧（修改）</div>
          <textarea
            value={rightInput}
            onChange={(e) => setRightInput(e.target.value)}
            className="flex-1 px-3 py-2 bg-[var(--bg-inset)] text-xs font-mono resize-none focus:outline-none"
            placeholder={mode === 'json' ? '{\n  "key": "new_value"\n}' : '输入文本内容...'}
            spellCheck={false}
            aria-label="右侧内容（修改）"
          />
        </div>
      </div>

      {/* 对比结果 */}
      {displaySections.length > 0 && (
        <div className="h-64 border-t border-[var(--border)] overflow-y-auto">
          <div className="px-3 py-1.5 border-b border-[var(--border)] text-xs text-[var(--blue)] sticky top-0 bg-[var(--hover-bg)]">
            差异结果
          </div>
          {displaySections.map((section, i) => (
            <div key={i} className={`flex items-start px-3 py-1 ${getDiffColor(section.type)}`}>
              <span className="w-4 text-xs font-mono shrink-0">{getDiffLabel(section.type)}</span>
              <span className="w-20 text-xs text-[var(--text-tertiary)] shrink-0">{section.path}</span>
              <div className="flex-1 min-w-0">
                {section.left !== undefined && (
                  <div className={`text-xs font-mono ${section.type === 'removed' || section.type === 'modified' ? 'line-through opacity-60' : ''}`}>
                    {section.left}
                  </div>
                )}
                {section.right !== undefined && section.type !== 'equal' && (
                  <div className="text-xs font-mono">{section.right}</div>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
