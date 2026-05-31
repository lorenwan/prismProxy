import { useState } from 'react'
import { compareRequests } from '../services/diff'
import type { DiffResult, DiffSection } from '../services/diff'

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
      if (mode === 'json') {
        JSON.parse(leftInput)
        JSON.parse(rightInput)
      }

      const diffResult = await compareRequests({
        leftId: 'inline-left',
        rightId: 'inline-right',
        compareHeaders: true,
        compareBody: true,
      })
      setResult(diffResult)
    } catch (err: any) {
      if (err instanceof SyntaxError) {
        setError('JSON 格式错误，请检查输入')
      } else {
        setError(String(err.message || err))
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

  function formatJson(input: string): string {
    try {
      return JSON.stringify(JSON.parse(input), null, 2)
    } catch {
      return input
    }
  }

  function getDiffColor(type: DiffSection['type']) {
    switch (type) {
      case 'added': return 'bg-[#9ece6a]/15 text-[#9ece6a] border-l-2 border-[#9ece6a]'
      case 'removed': return 'bg-[#f7768e]/15 text-[#f7768e] border-l-2 border-[#f7768e]'
      case 'modified': return 'bg-[#e0af68]/15 text-[#e0af68] border-l-2 border-[#e0af68]'
      default: return 'text-[#565f89]'
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

  // 本地文本对比（当 API 不可用时的后备方案）
  function localTextDiff(left: string, right: string): DiffSection[] {
    const leftLines = left.split('\n')
    const rightLines = right.split('\n')
    const sections: DiffSection[] = []
    const maxLen = Math.max(leftLines.length, rightLines.length)

    for (let i = 0; i < maxLen; i++) {
      const l = leftLines[i]
      const r = rightLines[i]
      if (l === undefined) {
        sections.push({ type: 'added', path: `line ${i + 1}`, rightValue: r })
      } else if (r === undefined) {
        sections.push({ type: 'removed', path: `line ${i + 1}`, leftValue: l })
      } else if (l !== r) {
        sections.push({ type: 'modified', path: `line ${i + 1}`, leftValue: l, rightValue: r })
      } else {
        sections.push({ type: 'equal', path: `line ${i + 1}`, leftValue: l })
      }
    }
    return sections
  }

  const displaySections = result
    ? [...result.requestDiff, ...result.responseDiff]
    : (leftInput && rightInput && !comparing)
      ? localTextDiff(
          mode === 'json' ? formatJson(leftInput) : leftInput,
          mode === 'json' ? formatJson(rightInput) : rightInput
        )
      : []

  return (
    <div className="flex flex-col h-full bg-[#24283b]">
      {/* 工具栏 */}
      <div className="flex items-center gap-2 px-4 py-2 border-b border-[#3b4261]">
        <div className="flex items-center gap-1 bg-[#1a1b26] rounded p-0.5">
          <button
            onClick={() => setMode('json')}
            className={`px-2 py-1 text-xs rounded ${mode === 'json' ? 'bg-[#7aa2f7] text-[#1a1b26]' : 'text-[#565f89] hover:text-[#e6edf3]'}`}
          >
            JSON
          </button>
          <button
            onClick={() => setMode('text')}
            className={`px-2 py-1 text-xs rounded ${mode === 'text' ? 'bg-[#7aa2f7] text-[#1a1b26]' : 'text-[#565f89] hover:text-[#e6edf3]'}`}
          >
            Text
          </button>
        </div>
        <button onClick={handleCompare} disabled={comparing} className="px-3 py-1 text-xs bg-[#7aa2f7] text-[#1a1b26] rounded hover:bg-[#89b4fa] disabled:opacity-50">
          {comparing ? '对比中...' : '对比'}
        </button>
        <button onClick={handleSwap} className="px-3 py-1 text-xs bg-[#24283b] border border-[#3b4261] rounded hover:bg-[#3b4261]">
          交换
        </button>
        <button onClick={handleClear} className="px-3 py-1 text-xs bg-[#24283b] border border-[#3b4261] rounded hover:bg-[#3b4261]">
          清空
        </button>
        {result && (
          <div className="ml-auto flex gap-3 text-xs text-[#565f89]">
            <span>请求变更: <span className="text-[#e0af68]">{result.summary.requestChanges}</span></span>
            <span>响应变更: <span className="text-[#e0af68]">{result.summary.responseChanges}</span></span>
          </div>
        )}
      </div>

      {error && (
        <div className="px-4 py-2 bg-[#f7768e]/10 border-b border-[#f7768e]/30 text-sm text-[#f7768e]">{error}</div>
      )}

      {/* 输入区域 */}
      <div className="flex flex-1 overflow-hidden">
        {/* 左侧输入 */}
        <div className="w-1/2 flex flex-col border-r border-[#3b4261]">
          <div className="px-3 py-1.5 border-b border-[#3b4261] text-xs text-[#565f89]">左侧（原始）</div>
          <textarea
            value={leftInput}
            onChange={(e) => setLeftInput(e.target.value)}
            className="flex-1 px-3 py-2 bg-[#1a1b26] text-xs font-mono resize-none focus:outline-none"
            placeholder={mode === 'json' ? '{\n  "key": "value"\n}' : '输入文本内容...'}
            spellCheck={false}
          />
        </div>

        {/* 右侧输入 */}
        <div className="w-1/2 flex flex-col">
          <div className="px-3 py-1.5 border-b border-[#3b4261] text-xs text-[#565f89]">右侧（修改）</div>
          <textarea
            value={rightInput}
            onChange={(e) => setRightInput(e.target.value)}
            className="flex-1 px-3 py-2 bg-[#1a1b26] text-xs font-mono resize-none focus:outline-none"
            placeholder={mode === 'json' ? '{\n  "key": "new_value"\n}' : '输入文本内容...'}
            spellCheck={false}
          />
        </div>
      </div>

      {/* 对比结果 */}
      {displaySections.length > 0 && (
        <div className="h-64 border-t border-[#3b4261] overflow-y-auto">
          <div className="px-3 py-1.5 border-b border-[#3b4261] text-xs text-[#7aa2f7] sticky top-0 bg-[#24283b]">
            差异结果
          </div>
          {displaySections.map((section, i) => (
            <div key={i} className={`flex items-start px-3 py-1 ${getDiffColor(section.type)}`}>
              <span className="w-4 text-xs font-mono shrink-0">{getDiffLabel(section.type)}</span>
              <span className="w-20 text-xs text-[#565f89] shrink-0">{section.path}</span>
              <div className="flex-1 min-w-0">
                {section.leftValue !== undefined && (
                  <div className={`text-xs font-mono ${section.type === 'removed' || section.type === 'modified' ? 'line-through opacity-60' : ''}`}>
                    {section.leftValue}
                  </div>
                )}
                {section.rightValue !== undefined && section.type !== 'equal' && (
                  <div className="text-xs font-mono">{section.rightValue}</div>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
