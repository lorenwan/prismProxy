import { invoke } from '@tauri-apps/api/core'
import { listen } from '@tauri-apps/api/event'
import type { ChatMessage, Settings } from '../types'

// 发送聊天消息（非流式）
// 注意: model 字段被重载为 provider 选择器，传入 provider 名称而非具体模型名
export async function sendMessage(messages: ChatMessage[], provider?: string): Promise<{ content: string; provider: string; model: string; usage?: { promptTokens: number; completionTokens: number; totalTokens: number } }> {
  const result = await invoke<string>('chat', {
    messages: messages.map((m) => ({ role: m.role, content: m.content })),
    model: provider,
  })
  return JSON.parse(result)
}

// 发送流式聊天消息
// 注意: model 字段被重载为 provider 选择器，传入 provider 名称而非具体模型名
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
  let unlistenError: (() => void) | null = null

  const cleanup = () => {
    unlistenChunk?.()
    unlistenEnd?.()
    unlistenError?.()
  }

  // 先注册所有 listener，再启动流
  Promise.all([
    listen('ai:chat_chunk', (event) => {
      if (controller.signal.aborted) return
      try {
        const chunk = JSON.parse(event.payload as string)
        onChunk(chunk.content || '')
      } catch {
        onChunk(event.payload as string)
      }
    }),
    listen('ai:chat_end', () => {
      if (controller.signal.aborted) return
      onDone()
      cleanup()
    }),
    listen<string>('ai:chat_error', (event) => {
      if (controller.signal.aborted) return
      onError(new Error(event.payload))
      cleanup()
    }),
  ]).then(([unlistenC, unlistenE, unlistenErr]) => {
    unlistenChunk = unlistenC
    unlistenEnd = unlistenE
    unlistenError = unlistenErr

    // listener 注册完成后再启动流
    invoke('stream_chat', {
      messages: messages.map((m) => ({ role: m.role, content: m.content })),
      model: provider,
    }).catch((err) => {
      if (!controller.signal.aborted) {
        onError(err)
        cleanup()
      }
    })
  })

  const originalAbort = controller.abort.bind(controller)
  controller.abort = () => {
    originalAbort()
    cleanup()
  }

  return controller
}

// 检查 AI 服务可用性
export async function checkAiAvailability(): Promise<boolean> {
  const result = await invoke<string>('check_ai_availability')
  const response = JSON.parse(result) as { available: boolean; providers: string[] }
  return response.available
}

// 获取设置
export async function getSettings(): Promise<Settings> {
  const result = await invoke<string>('get_settings')
  return JSON.parse(result)
}

// 更新设置
export async function updateSettings(settings: Partial<Settings>): Promise<Settings> {
  const result = await invoke<string>('update_settings', {
    settings: JSON.stringify(settings),
  })
  return JSON.parse(result)
}

// 下载 CA 证书（TODO: 需要通过 Tauri 文件对话框实现）
export function downloadCaCert() {
  // 当 Rust 层 export_ca IPC 命令支持文件保存后替换
  console.warn('downloadCaCert: 暂未实现 Tauri 文件保存')
}
