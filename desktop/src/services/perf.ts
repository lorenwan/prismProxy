import { invoke } from '@tauri-apps/api/core'

export interface PerformanceReport {
  id: string
  name: string
  status: 'running' | 'completed' | 'failed'
  config: PerfConfig
  results?: PerfResults
  createdAt: string
  completedAt?: string
}

export interface PerfConfig {
  targetUrl: string
  method: string
  headers?: Record<string, string>
  body?: string
  concurrency: number
  totalRequests: number
  duration?: number
}

export interface PerfResults {
  totalRequests: number
  successfulRequests: number
  failedRequests: number
  totalDuration: number
  avgResponseTime: number
  minResponseTime: number
  maxResponseTime: number
  p50ResponseTime: number
  p90ResponseTime: number
  p95ResponseTime: number
  p99ResponseTime: number
  requestsPerSecond: number
  errorRate: number
  throughput: number
  statusCodes: Record<number, number>
  timeline: Array<{
    timestamp: number
    rps: number
    avgLatency: number
    errorRate: number
  }>
}

// 获取性能统计
export async function getPerfStats(since?: string): Promise<any> {
  const result = await invoke<string>('get_perf_stats', { since })
  return JSON.parse(result)
}

// 获取慢请求列表
export async function getSlowRequests(thresholdMs?: number, limit?: number): Promise<any> {
  const result = await invoke<string>('get_slow_requests', { thresholdMs, limit })
  return JSON.parse(result)
}

// 获取域名统计
export async function getDomainStats(): Promise<any> {
  const result = await invoke<string>('get_domain_stats')
  return JSON.parse(result)
}

// 获取时间线数据
export async function getPerfTimeline(since?: string, intervalSeconds?: number): Promise<any> {
  const result = await invoke<string>('get_perf_timeline', { since, intervalSeconds })
  return JSON.parse(result)
}

// 获取状态码统计
export async function getStatusCodeStats(since?: string): Promise<any> {
  const result = await invoke<string>('get_status_code_stats', { since })
  return JSON.parse(result)
}

// 获取请求方法统计
export async function getMethodStats(since?: string): Promise<any> {
  const result = await invoke<string>('get_method_stats', { since })
  return JSON.parse(result)
}

// 获取最近 N 分钟统计
export async function getRecentStats(minutes: number): Promise<any> {
  const result = await invoke<string>('get_recent_stats', { minutes })
  return JSON.parse(result)
}

// --- 向后兼容的别名函数 ---
// 保持与现有调用方兼容

// 获取性能报告列表（TODO: Rust 层暂未实现性能报告列表 IPC 命令）
export async function getPerfReports(): Promise<PerformanceReport[]> {
  return []
}

// 获取性能报告详情（TODO: Rust 层暂未实现性能报告详情 IPC 命令）
export async function getPerfReport(id: string): Promise<PerformanceReport> {
  return {
    id,
    name: '',
    status: 'completed',
    config: { targetUrl: '', method: 'GET', concurrency: 1, totalRequests: 0 },
    createdAt: '',
  }
}

// 创建性能测试（TODO: Rust 层暂未实现性能测试 IPC 命令）
export async function createPerfTest(config: PerfConfig): Promise<PerformanceReport> {
  console.warn('createPerfTest: 暂未实现 Tauri IPC')
  return {
    id: '',
    name: '',
    status: 'failed',
    config,
    createdAt: new Date().toISOString(),
  }
}

// 停止性能测试（TODO: Rust 层暂未实现性能测试 IPC 命令）
export async function stopPerfTest(id: string): Promise<void> {
  console.warn('stopPerfTest: 暂未实现 Tauri IPC')
}

// 删除性能报告（TODO: Rust 层暂未实现性能报告 IPC 命令）
export async function deletePerfReport(id: string): Promise<void> {
  console.warn('deletePerfReport: 暂未实现 Tauri IPC')
}

// 从流量记录创建性能测试（TODO: Rust 层暂未实现）
export async function createPerfFromTransaction(transactionId: string, config: Partial<PerfConfig>): Promise<PerformanceReport> {
  console.warn('createPerfFromTransaction: 暂未实现 Tauri IPC')
  return {
    id: '',
    name: '',
    status: 'failed',
    config: { targetUrl: '', method: 'GET', concurrency: 1, totalRequests: 0, ...config },
    createdAt: new Date().toISOString(),
  }
}
