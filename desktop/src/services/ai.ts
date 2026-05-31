import api from './api'
import type { ChatMessage, Settings } from '../types'

// 发送聊天消息（非流式）
export async function sendMessage(messages: ChatMessage[], provider?: string): Promise<{ message: ChatMessage }> {
  return api.post('/ai/chat', { messages, provider }) as any
}

// 发送流式聊天消息
export function sendMessageStream(
  messages: ChatMessage[],
  provider: string | undefined,
  onChunk: (text: string) => void,
  onDone: () => void,
  onError: (err: Error) => void
) {
  const controller = new AbortController()

  fetch('/api/ai/chat/stream', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ messages, provider }),
    signal: controller.signal,
  })
    .then(async (res) => {
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      const reader = res.body?.getReader()
      if (!reader) return
      const decoder = new TextDecoder()
      while (true) {
        const { done, value } = await reader.read()
        if (done) break
        onChunk(decoder.decode(value, { stream: true }))
      }
      onDone()
    })
    .catch((err) => {
      if (err.name !== 'AbortError') onError(err)
    })

  return controller
}

// 获取设置
export async function getSettings(): Promise<Settings> {
  return api.get('/settings') as any
}

// 更新设置
export async function updateSettings(settings: Partial<Settings>): Promise<Settings> {
  return api.put('/settings', settings) as any
}

// 下载 CA 证书
export function downloadCaCert() {
  window.open('/api/ca/cert', '_blank')
}
