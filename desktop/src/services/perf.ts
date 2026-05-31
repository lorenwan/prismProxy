import api from './api'

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

// 获取性能报告列表
export async function getPerfReports(): Promise<PerformanceReport[]> {
  return api.get('/perf/reports') as any
}

// 获取性能报告详情
export async function getPerfReport(id: string): Promise<PerformanceReport> {
  return api.get(`/perf/reports/${id}`) as any
}

// 创建性能测试
export async function createPerfTest(config: PerfConfig): Promise<PerformanceReport> {
  return api.post('/perf/tests', config) as any
}

// 停止性能测试
export async function stopPerfTest(id: string): Promise<void> {
  return api.post(`/perf/tests/${id}/stop`) as any
}

// 删除性能报告
export async function deletePerfReport(id: string): Promise<void> {
  return api.delete(`/perf/reports/${id}`) as any
}

// 从流量记录创建性能测试
export async function createPerfFromTransaction(transactionId: string, config: Partial<PerfConfig>): Promise<PerformanceReport> {
  return api.post(`/perf/from-transaction/${transactionId}`, config) as any
}
