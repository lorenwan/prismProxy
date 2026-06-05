import { invoke } from '@tauri-apps/api/core'
import { listen } from '@tauri-apps/api/event'
import type { ChatMessage, Settings } from '../types'

// 发送聊天消息（非流式）
export async function sendMessage(messages: ChatMessage[], provider?: string): Promise<{ message: ChatMessage }> {
  const result = await invoke<string>('chat', {
    messages: messages.map((m) => ({ role: m.role, content: m.content })),
    model: provider,
  })
  return JSON.parse(result)
}

// 发送流式聊天消息
// 返回兼容 AbortController 的对象，支持 abort() 取消
export function sendMessageStream(
  messages: ChatMessage[],
  provider: string | undefined,
  onChunk: (text: string) => void,
  onDone: () => void,
  onError: (err: Error) => void
): AbortController {
  const controller = new AbortController()
  let unlistenChunk: (() => void) | null = null
  let unlistenEnd: (() => void) | null = null

  // 启动流式聊天
  invoke('stream_chat', {
    messages: messages.map((m) => ({ role: m.role, content: m.content })),
    model: provider,
  }).catch((err) => {
    if (!controller.signal.aborted) onError(err)
  })

  // 监听 chunk 事件
  listen('ai:chat_chunk', (event) => {
    if (controller.signal.aborted) return
    try {
      const chunk = JSON.parse(event.payload as string)
      onChunk(chunk.content || '')
    } catch {
      onChunk(event.payload as string)
    }
  }).then((unlisten) => {
    unlistenChunk = unlisten
  })

  // 监听结束事件
  listen('ai:chat_end', () => {
    if (controller.signal.aborted) return
    onDone()
    unlistenChunk?.()
    unlistenEnd?.()
  }).then((unlisten) => {
    unlistenEnd = unlisten
  })

  // 包装 abort 方法以同时清理事件监听
  const originalAbort = controller.abort.bind(controller)
  controller.abort = () => {
    originalAbort()
    unlistenChunk?.()
    unlistenEnd?.()
  }

  return controller
}

// 检查 AI 服务可用性
export async function checkAiAvailability(): Promise<boolean> {
  const result = await invoke<string>('check_ai_availability')
  return JSON.parse(result)
}

// 获取设置（TODO: Rust 层暂未实现 settings IPC 命令）
export async function getSettings(): Promise<Settings> {
  // 当 Rust 层 settings IPC 命令就绪后替换为 invoke 调用
  return {
    proxy: { port: 8888, mitmEnabled: false, caCertPath: '' },
    ai: { provider: 'openai', apiKey: '', baseUrl: '', model: '' },
  }
}

// 更新设置（TODO: Rust 层暂未实现 settings IPC 命令）
export async function updateSettings(settings: Partial<Settings>): Promise<Settings> {
  // 当 Rust 层 settings IPC 命令就绪后替换为 invoke 调用
  return settings as Settings
}

// 下载 CA 证书（TODO: 需要通过 Tauri 文件对话框实现）
export function downloadCaCert() {
  // 当 Rust 层 export_ca IPC 命令支持文件保存后替换
  console.warn('downloadCaCert: 暂未实现 Tauri 文件保存')
}
