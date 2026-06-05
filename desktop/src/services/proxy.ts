import { invoke } from '@tauri-apps/api/core'

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
  await invoke('start_proxy')
  const status = await getSystemStatus()
  return {
    running: status?.proxyRunning || true,
    addr: status?.proxyAddr || '',
  }
}

// 停止代理
export async function stopProxy(): Promise<ProxyStatus> {
  await invoke('stop_proxy')
  return {
    running: false,
    addr: '',
  }
}

// 获取代理状态
export async function getProxyStatus(): Promise<ProxyStatus> {
  const status = await getSystemStatus()
  return {
    running: status?.proxyRunning || false,
    addr: status?.proxyAddr || '',
  }
}

// 启用系统代理（TODO: Rust 层暂未实现系统代理开关 IPC 命令）
export async function enableSystemProxy(): Promise<SystemProxyStatus> {
  console.warn('enableSystemProxy: 暂未实现 Tauri IPC')
  return { enabled: false, proxyAddr: '', error: '暂未实现' }
}

// 禁用系统代理（TODO: Rust 层暂未实现系统代理开关 IPC 命令）
export async function disableSystemProxy(): Promise<SystemProxyStatus> {
  console.warn('disableSystemProxy: 暂未实现 Tauri IPC')
  return { enabled: false, proxyAddr: '', error: '暂未实现' }
}

// 获取系统代理状态（TODO: Rust 层暂未实现系统代理状态 IPC 命令）
export async function getSystemProxyStatus(): Promise<SystemProxyStatus> {
  console.warn('getSystemProxyStatus: 暂未实现 Tauri IPC')
  return { enabled: false, proxyAddr: '' }
}
