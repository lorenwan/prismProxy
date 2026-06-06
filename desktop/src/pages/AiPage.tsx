import { useState, useRef, useEffect } from 'react'
import type { ChatMessage } from '../types'
import { sendMessageStream } from '../services/ai'

// 快捷操作
const quickActions = [
  { label: '分析流量', prompt: '请分析最近的 HTTP 流量，找出异常请求和性能问题。' },
  { label: '安全检测', prompt: '请检查流量中是否存在安全风险，如 SQL 注入、XSS 攻击等。' },
  { label: '生成测试', prompt: '请根据当前流量数据生成 API 测试用例。' },
]

// Provider 选项
const providers = [
  { value: 'openai', label: 'OpenAI' },
  { value: 'claude', label: 'Claude' },
  { value: 'ollama', label: 'Ollama' },
]

export default function AiPage() {
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [input, setInput] = useState('')
  const [provider, setProvider] = useState('openai')
  const [loading, setLoading] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const controllerRef = useRef<AbortController | null>(null)

  // 滚动到底部
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  // 发送消息
  function handleSend(text?: string) {
    const content = text || input.trim()
    if (!content || loading) return

    const userMsg: ChatMessage = { role: 'user', content, timestamp: new Date().toISOString() }
    const newMessages = [...messages, userMsg]
    setMessages(newMessages)
    setInput('')
    setLoading(true)

    // 添加空的 assistant 消息用于流式填充
    const assistantMsg: ChatMessage = { role: 'assistant', content: '', timestamp: new Date().toISOString() }
    setMessages([...newMessages, assistantMsg])

    controllerRef.current = sendMessageStream(
      newMessages,
      provider,
      // onChunk
      (chunk) => {
        setMessages((prev) => {
          const last = prev[prev.length - 1]
          if (last.role !== 'assistant') return prev
          return [...prev.slice(0, -1), { ...last, content: last.content + chunk }]
        })
      },
      // onDone
      () => setLoading(false),
      // onError
      (err) => {
        setMessages((prev) => {
          const last = prev[prev.length - 1]
          if (last.role !== 'assistant') return prev
          return [...prev.slice(0, -1), { ...last, content: `错误: ${err.message}` }]
        })
        setLoading(false)
      }
    )
  }

  // 停止生成
  function handleStop() {
    controllerRef.current?.abort()
    setLoading(false)
  }

  // 清空对话
  function handleClear() {
    setMessages([])
  }

  return (
    <div className="flex flex-col h-full bg-[var(--bg-inset)]">
      {/* 顶部栏 */}
      <div className="flex items-center justify-between px-4 py-2 border-b border-[var(--border)]">
        <div className="flex items-center gap-3">
          <h2 className="text-sm font-semibold">AI 助手</h2>
          <select
            value={provider}
            onChange={(e) => setProvider(e.target.value)}
            className="px-2 py-1 bg-[var(--hover-bg)] border border-[var(--border)] rounded text-xs focus:border-[var(--blue)] focus:outline-none"
          >
            {providers.map((p) => (
              <option key={p.value} value={p.value}>{p.label}</option>
            ))}
          </select>
        </div>
        <button onClick={handleClear} className="px-2 py-1 text-xs bg-[var(--hover-bg)] rounded hover:bg-[var(--hover-bg)]">
          清空对话
        </button>
      </div>

      {/* 快捷操作 */}
      <div className="flex gap-2 px-4 py-2 border-b border-[var(--border)]">
        {quickActions.map((action) => (
          <button
            key={action.label}
            onClick={() => handleSend(action.prompt)}
            disabled={loading}
            className="px-3 py-1 text-xs bg-[var(--hover-bg)] border border-[var(--border)] rounded hover:bg-[var(--selected-bg)] hover:border-[var(--blue)] disabled:opacity-50"
          >
            {action.label}
          </button>
        ))}
      </div>

      {/* 消息列表 */}
      <div className="flex-1 overflow-y-auto px-4 py-3 space-y-4">
        {messages.length === 0 && (
          <div className="flex items-center justify-center h-full text-[var(--text-tertiary)] text-sm">
            选择快捷操作或输入问题开始对话
          </div>
        )}
        {messages.map((msg, i) => (
          <div key={i} className={`flex gap-3 ${msg.role === 'user' ? 'justify-end' : ''}`}>
            {msg.role === 'assistant' && (
              <div className="w-8 h-8 rounded-lg bg-[var(--blue)] flex items-center justify-center flex-shrink-0">
                <span className="text-sm text-white font-bold">AI</span>
              </div>
            )}
            <div
              className={`max-w-[70%] px-3 py-2 rounded-lg text-sm whitespace-pre-wrap ${
                msg.role === 'user'
                  ? 'bg-[var(--blue)] text-white'
                  : 'bg-[var(--hover-bg)]'
              }`}
            >
              {msg.content || (loading && i === messages.length - 1 ? '思考中...' : '')}
            </div>
          </div>
        ))}
        <div ref={messagesEndRef} />
      </div>

      {/* 输入框 */}
      <div className="px-4 py-3 border-t border-[var(--border)]">
        <div className="flex gap-2">
          <input
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && !e.shiftKey && handleSend()}
            className="flex-1 px-3 py-2 bg-[var(--hover-bg)] border border-[var(--border)] rounded text-sm focus:border-[var(--blue)] focus:outline-none"
            placeholder="输入消息..."
            disabled={loading}
          />
          {loading ? (
            <button onClick={handleStop} className="px-4 py-2 bg-[var(--red)] text-white rounded text-sm hover:bg-[var(--red)]/90">
              停止
            </button>
          ) : (
            <button onClick={() => handleSend()} className="px-4 py-2 bg-[var(--blue)] text-white rounded text-sm hover:bg-[var(--blue)]/90">
              发送
            </button>
          )}
        </div>
      </div>
    </div>
  )
}
