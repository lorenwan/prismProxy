import { invoke } from '@tauri-apps/api/core'
import { listen } from '@tauri-apps/api/event'
import type { Breakpoint, BreakpointSession, RequestData, ResponseData } from '../types'

// 获取断点列表
export async function getBreakpoints(): Promise<Breakpoint[]> {
  const result = await invoke<string>('list_breakpoints')
  return JSON.parse(result)
}

// 获取单条断点
export async function getBreakpoint(id: string): Promise<Breakpoint> {
  const result = await invoke<string>('get_breakpoint', { id })
  return JSON.parse(result)
}

// 创建断点
export async function createBreakpoint(bp: Partial<Breakpoint>): Promise<Breakpoint> {
  const result = await invoke<string>('create_breakpoint', { breakpoint: JSON.stringify(bp) })
  return JSON.parse(result)
}

// 更新断点
export async function updateBreakpoint(id: string, bp: Partial<Breakpoint>): Promise<Breakpoint> {
  const result = await invoke<string>('update_breakpoint', { breakpoint: JSON.stringify({ ...bp, id }) })
  return JSON.parse(result)
}

// 删除断点
export async function deleteBreakpoint(id: string): Promise<void> {
  await invoke('delete_breakpoint', { id })
}

// 切换断点启用状态
export async function toggleBreakpoint(id: string, enabled: boolean): Promise<void> {
  await invoke('toggle_breakpoint', { id, enabled })
}

// 获取活跃断点会话
export async function getActiveSessions(): Promise<BreakpointSession[]> {
  const result = await invoke<string>('list_breakpoint_sessions')
  return JSON.parse(result)
}

// 恢复断点会话（继续执行）
export async function resumeSession(sessionId: string): Promise<void> {
  await invoke('resolve_breakpoint_session', {
    sessionId,
    action: 'resume',
    modifiedData: null,
  })
}

// 修改并恢复断点会话
export async function modifyAndResume(sessionId: string, data: RequestData | ResponseData): Promise<void> {
  await invoke('resolve_breakpoint_session', {
    sessionId,
    action: 'modify',
    modifiedData: JSON.stringify(data),
  })
}

// 订阅断点事件（服务端流式推送）
export async function subscribeBreakpoints(callback: (event: any) => void): Promise<() => void> {
  await invoke('subscribe_breakpoints')
  const unlisten = await listen('breakpoints:event', (event) => {
    callback(JSON.parse(event.payload as string))
  })
  return unlisten
}
