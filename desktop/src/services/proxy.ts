import { invoke } from '@tauri-apps/api/core'

// TODO: 前端 IPC 调用缺少集中错误处理
// 所有 invoke 调用应包装在 try/catch 中，统一处理 Tauri IPC 错误
// 建议创建通用的 invokeWithRetry / invokeWithErrorHandling 工具函数

// 代理状态
export interface ProxyStatus {
  running: boolean
  addr: string
  error?: string
}

// 系统代理状态
export interface SystemProxyStatus {
  enabled: boolean
  proxyAddr: string
  error?: string
}

// 获取系统状态（包含代理信息）
export async function getSystemStatus(): Promise<any> {
  const result = await invoke<string>('get_system_status')
  return JSON.parse(result)
}

// 启动代理
export async function startProxy(): Promise<ProxyStatus> {
  const result = await invoke<string>('start_proxy')
  return JSON.parse(result)
}

// 停止代理
export async function stopProxy(): Promise<ProxyStatus> {
  const result = await invoke<string>('stop_proxy')
  return JSON.parse(result)
}

// 获取代理状态
export async function getProxyStatus(): Promise<ProxyStatus> {
  const status = await getSystemStatus()
  return {
    running: status?.proxyRunning ?? false,
    addr: status?.proxyAddr || '',
  }
}

// 启用系统代理
export async function enableSystemProxy(): Promise<SystemProxyStatus> {
  const result = await invoke<string>('enable_system_proxy')
  return JSON.parse(result)
}

// 禁用系统代理
export async function disableSystemProxy(): Promise<SystemProxyStatus> {
  const result = await invoke<string>('disable_system_proxy')
  return JSON.parse(result)
}

// 获取系统代理状态
export async function getSystemProxyStatus(): Promise<SystemProxyStatus> {
  const result = await invoke<string>('get_system_proxy_status')
  return JSON.parse(result)
}
