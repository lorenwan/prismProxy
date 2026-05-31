import { createGrpcClient } from './grpc'
import { SystemService } from '../proto/gen/ts/system_service_pb'

// 创建 gRPC 客户端
const systemClient = createGrpcClient(SystemService)

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

// 启动代理
export async function startProxy(): Promise<ProxyStatus> {
  const response = await systemClient.startProxy({})
  return {
    running: response.running,
    addr: response.addr || '',
    error: response.error || undefined,
  }
}

// 停止代理
export async function stopProxy(): Promise<ProxyStatus> {
  const response = await systemClient.stopProxy({})
  return {
    running: response.running,
    addr: '',
    error: response.error || undefined,
  }
}

// 获取代理状态
export async function getProxyStatus(): Promise<ProxyStatus> {
  const response = await systemClient.getStatus({})
  return {
    running: response.proxyRunning || false,
    addr: response.proxyAddr || '',
  }
}

// 启用系统代理
export async function enableSystemProxy(): Promise<SystemProxyStatus> {
  const response = await systemClient.enableSystemProxy({})
  return {
    enabled: response.enabled,
    proxyAddr: response.proxyAddr || '',
    error: response.error || undefined,
  }
}

// 禁用系统代理
export async function disableSystemProxy(): Promise<SystemProxyStatus> {
  const response = await systemClient.disableSystemProxy({})
  return {
    enabled: response.enabled,
    proxyAddr: '',
    error: response.error || undefined,
  }
}

// 获取系统代理状态
export async function getSystemProxyStatus(): Promise<SystemProxyStatus> {
  const response = await systemClient.getSystemProxyStatus({})
  return {
    enabled: response.enabled,
    proxyAddr: response.proxyAddr || '',
    error: response.error || undefined,
  }
}
