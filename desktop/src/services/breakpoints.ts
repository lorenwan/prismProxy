import api from './api'
import type { Breakpoint, BreakpointSession, RequestData, ResponseData } from '../types'

// 获取断点列表
export async function getBreakpoints(): Promise<Breakpoint[]> {
  return api.get('/breakpoints') as any
}

// 创建断点
export async function createBreakpoint(bp: Partial<Breakpoint>): Promise<Breakpoint> {
  return api.post('/breakpoints', bp) as any
}

// 更新断点
export async function updateBreakpoint(id: string, bp: Partial<Breakpoint>): Promise<Breakpoint> {
  return api.put(`/breakpoints/${id}`, bp) as any
}

// 删除断点
export async function deleteBreakpoint(id: string): Promise<void> {
  return api.delete(`/breakpoints/${id}`) as any
}

// 切换断点启用状态
export async function toggleBreakpoint(id: string, enabled: boolean): Promise<void> {
  return api.put(`/breakpoints/${id}`, { enabled }) as any
}

// 获取活跃断点会话
export async function getActiveSessions(): Promise<BreakpointSession[]> {
  return api.get('/breakpoints/sessions') as any
}

// 恢复断点会话
export async function resumeSession(sessionId: string): Promise<void> {
  return api.post(`/breakpoints/sessions/${sessionId}/resume`) as any
}

// 修改并恢复断点会话
export async function modifyAndResume(sessionId: string, data: RequestData | ResponseData): Promise<void> {
  return api.post(`/breakpoints/sessions/${sessionId}/modify`, data) as any
}
