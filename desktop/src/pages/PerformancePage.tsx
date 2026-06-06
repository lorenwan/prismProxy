import { useState, useEffect, useMemo } from 'react'
import { getPerfReports, deletePerfReport } from '../services/perf'
import type { PerformanceReport } from '../services/perf'

export default function PerformancePage() {
  const [reports, setReports] = useState<PerformanceReport[]>([])
  const [selected, setSelected] = useState<PerformanceReport | null>(null)

  useEffect(() => {
    getPerfReports().then(setReports).catch(console.error)
  }, [])

  async function handleDelete(id: string) {
    await deletePerfReport(id)
    setReports(reports.filter((r) => r.id !== id))
    if (selected?.id === id) setSelected(null)
  }

  // 从报告中提取概览数据 - 使用 useMemo 缓存计算结果
  const completedReports = useMemo(
    () => reports.filter((r) => r.status === 'completed' && r.results),
    [reports]
  )

  const totalRequests = useMemo(
    () => completedReports.reduce((sum, r) => sum + (r.results?.totalRequests || 0), 0),
    [completedReports]
  )

  const avgDuration = useMemo(
    () => completedReports.length > 0
      ? Math.round(completedReports.reduce((sum, r) => sum + (r.results?.avgDurationMs || 0), 0) / completedReports.length)
      : 0,
    [completedReports]
  )

  const p50 = selected?.results?.p50Ms ?? 0
  const p90 = selected?.results?.p90Ms ?? 0
  const p99 = selected?.results?.p99Ms ?? 0
  const errorRate = selected?.results?.slowRequests ?? 0

  // 慢请求：取所有报告中 p90 以上的请求 - 缓存排序结果
  const slowRequests = useMemo(
    () => completedReports
      .filter((r) => r.results && r.results.avgDurationMs > 500)
      .sort((a, b) => (b.results?.avgDurationMs || 0) - (a.results?.avgDurationMs || 0)),
    [completedReports]
  )

  // 域名统计（从报告 config 中提取）- 缓存复杂聚合计算
  const domainStatsList = useMemo(() => {
    const domainStats = new Map<string, { count: number; totalDuration: number; errors: number }>()
    completedReports.forEach((r) => {
      try {
        const url = new URL(r.config.targetUrl)
        const domain = url.hostname
        const existing = domainStats.get(domain) || { count: 0, totalDuration: 0, errors: 0 }
        existing.count++
        existing.totalDuration += r.results?.avgDurationMs || 0
        existing.errors += r.results?.slowRequests || 0
        domainStats.set(domain, existing)
      } catch {}
    })
    return Array.from(domainStats.entries())
      .map(([domain, stats]) => ({
        domain,
        requests: stats.count,
        avgDuration: Math.round(stats.totalDuration / stats.count),
        errorRate: stats.count > 0 ? ((stats.errors / stats.count) * 100).toFixed(1) : '0',
      }))
      .sort((a, b) => b.requests - a.requests)
  }, [completedReports])

  // 时间线数据（PerfResults 中无 timeline 字段，暂用空数组）
  const timeline: Array<{ rps: number; avgLatency: number; errorRate: number }> = []
  const maxRps = Math.max(...timeline.map((t) => t.rps), 1)

  function getStatusColor(report: PerformanceReport) {
    if (report.status === 'running') return 'text-[var(--blue)]'
    if (report.status === 'failed') return 'text-[var(--red)]'
    return 'text-[var(--green)]'
  }

  function getStatusLabel(report: PerformanceReport) {
    if (report.status === 'running') return '运行中'
    if (report.status === 'failed') return '失败'
    return '完成'
  }

  return (
    <div className="flex flex-col h-full bg-[var(--hover-bg)] overflow-y-auto">
      <div className="p-4 space-y-4">
        <h1 className="text-lg font-semibold">性能分析</h1>

        {/* 概览卡片 */}
        <div className="grid grid-cols-5 gap-3">
          {[
            { label: '总请求数', value: totalRequests.toLocaleString(), color: 'var(--blue)' },
            { label: '平均耗时', value: `${avgDuration}ms`, color: 'var(--green)' },
            { label: 'P50', value: `${p50}ms`, color: 'var(--yellow)' },
            { label: 'P90', value: `${p90}ms`, color: 'var(--yellow)' },
            { label: '慢请求', value: `${errorRate}`, color: 'var(--red)' },
          ].map((card) => (
            <div key={card.label} className="bg-[var(--bg-inset)] border border-[var(--border)] rounded p-3">
              <div className="text-xs text-[var(--text-tertiary)] mb-1">{card.label}</div>
              <div className="text-xl font-semibold" style={{ color: card.color }}>{card.value}</div>
            </div>
          ))}
        </div>

        {/* 选中报告的详细指标 */}
        {selected?.results && (
          <div className="grid grid-cols-4 gap-3">
            <div className="bg-[var(--bg-inset)] border border-[var(--border)] rounded p-3">
              <div className="text-xs text-[var(--text-tertiary)] mb-1">P99</div>
              <div className="text-lg font-semibold text-[var(--red)]">{selected.results.p99Ms}ms</div>
            </div>
            <div className="bg-[var(--bg-inset)] border border-[var(--border)] rounded p-3">
              <div className="text-xs text-[var(--text-tertiary)] mb-1">慢请求数</div>
              <div className="text-lg font-semibold text-[var(--blue)]">{selected.results.slowRequests}</div>
            </div>
            <div className="bg-[var(--bg-inset)] border border-[var(--border)] rounded p-3">
              <div className="text-xs text-[var(--text-tertiary)] mb-1">最大耗时</div>
              <div className="text-lg font-semibold text-[var(--green)]">{selected.results.maxDurationMs}ms</div>
            </div>
            <div className="bg-[var(--bg-inset)] border border-[var(--border)] rounded p-3">
              <div className="text-xs text-[var(--text-tertiary)] mb-1">总耗时</div>
              <div className="text-lg font-semibold text-[var(--yellow)]">{(selected.results.totalDurationMs / 1000).toFixed(1)}s</div>
            </div>
          </div>
        )}

        <div className="grid grid-cols-2 gap-4">
          {/* 慢请求列表 */}
          <div className="bg-[var(--bg-inset)] border border-[var(--border)] rounded overflow-hidden">
            <div className="px-3 py-2 border-b border-[var(--border)] text-sm font-medium text-[var(--blue)]">慢请求列表</div>
            <div className="max-h-64 overflow-y-auto">
              <table className="w-full text-xs">
                <thead>
                  <tr className="text-[var(--text-tertiary)] border-b border-[var(--border-subtle)]">
                    <th className="px-3 py-1.5 text-left font-medium">名称</th>
                    <th className="px-3 py-1.5 text-left font-medium">状态</th>
                    <th className="px-3 py-1.5 text-right font-medium">平均耗时</th>
                    <th className="px-3 py-1.5 text-right font-medium">请求数</th>
                    <th className="px-3 py-1.5 text-right font-medium">操作</th>
                  </tr>
                </thead>
                <tbody>
                  {slowRequests.map((report) => (
                    <tr
                      key={report.id}
                      onClick={() => setSelected(report)}
                      className={`cursor-pointer border-b border-[var(--border-subtle)] ${selected?.id === report.id ? 'bg-[var(--selected-bg)]' : 'hover:bg-[var(--hover-bg)]'}`}
                    >
                      <td className="px-3 py-1.5 truncate max-w-[200px]">{report.name}</td>
                      <td className={`px-3 py-1.5 ${getStatusColor(report)}`}>{getStatusLabel(report)}</td>
                      <td className="px-3 py-1.5 text-right text-[var(--red)]">{report.results?.avgDurationMs}ms</td>
                      <td className="px-3 py-1.5 text-right">{report.results?.totalRequests}</td>
                      <td className="px-3 py-1.5 text-right">
                        <button
                          onClick={(e) => { e.stopPropagation(); handleDelete(report.id) }}
                          className="text-[var(--red)] hover:text-[var(--red)]/90"
                        >
                          ✕
                        </button>
                      </td>
                    </tr>
                  ))}
                  {slowRequests.length === 0 && (
                    <tr><td colSpan={5} className="px-3 py-4 text-center text-[var(--text-tertiary)]">暂无慢请求</td></tr>
                  )}
                </tbody>
              </table>
            </div>
          </div>

          {/* 域名统计 */}
          <div className="bg-[var(--bg-inset)] border border-[var(--border)] rounded overflow-hidden">
            <div className="px-3 py-2 border-b border-[var(--border)] text-sm font-medium text-[var(--blue)]">域名统计</div>
            <div className="max-h-64 overflow-y-auto">
              <table className="w-full text-xs">
                <thead>
                  <tr className="text-[var(--text-tertiary)] border-b border-[var(--border-subtle)]">
                    <th className="px-3 py-1.5 text-left font-medium">域名</th>
                    <th className="px-3 py-1.5 text-right font-medium">请求数</th>
                    <th className="px-3 py-1.5 text-right font-medium">平均耗时</th>
                    <th className="px-3 py-1.5 text-right font-medium">错误率</th>
                  </tr>
                </thead>
                <tbody>
                  {domainStatsList.map((stat) => (
                    <tr key={stat.domain} className="border-b border-[var(--border-subtle)] hover:bg-[var(--hover-bg)]">
                      <td className="px-3 py-1.5 font-mono">{stat.domain}</td>
                      <td className="px-3 py-1.5 text-right">{stat.requests}</td>
                      <td className="px-3 py-1.5 text-right">{stat.avgDuration}ms</td>
                      <td className={`px-3 py-1.5 text-right ${parseFloat(stat.errorRate) > 5 ? 'text-[var(--red)]' : 'text-[var(--text-tertiary)]'}`}>
                        {stat.errorRate}%
                      </td>
                    </tr>
                  ))}
                  {domainStatsList.length === 0 && (
                    <tr><td colSpan={4} className="px-3 py-4 text-center text-[var(--text-tertiary)]">暂无数据</td></tr>
                  )}
                </tbody>
              </table>
            </div>
          </div>
        </div>

        {/* 时间线图表 */}
        {timeline.length > 0 && (
          <div className="bg-[var(--bg-inset)] border border-[var(--border)] rounded overflow-hidden">
            <div className="px-3 py-2 border-b border-[var(--border)] text-sm font-medium text-[var(--blue)]">
              RPS 时间线 — {selected?.name}
            </div>
            <div className="p-3">
              <div className="flex items-end gap-0.5 h-32">
                {timeline.map((point, i) => {
                  const height = Math.max((point.rps / maxRps) * 100, 2)
                  const hasError = point.errorRate > 0.05
                  return (
                    <div
                      key={i}
                      className="flex-1 group relative"
                      style={{ height: '100%' }}
                    >
                      <div
                        className={`absolute bottom-0 w-full rounded-t transition-all ${
                          hasError ? 'bg-[var(--red)]' : 'bg-[var(--blue)]'
                        }`}
                        style={{ height: `${height}%` }}
                      />
                      <div className="absolute bottom-full left-1/2 -translate-x-1/2 mb-1 hidden group-hover:block z-10">
                        <div className="bg-[var(--hover-bg)] border border-[var(--border)] rounded px-2 py-1 text-[10px] whitespace-nowrap">
                          <div>RPS: {point.rps.toFixed(1)}</div>
                          <div>延迟: {point.avgLatency.toFixed(0)}ms</div>
                          <div>错误: {(point.errorRate * 100).toFixed(1)}%</div>
                        </div>
                      </div>
                    </div>
                  )
                })}
              </div>
              <div className="flex justify-between mt-1 text-[10px] text-[var(--text-tertiary)]">
                <span>开始</span>
                <span>结束</span>
              </div>
            </div>
          </div>
        )}

        {/* 所有报告列表 */}
        <div className="bg-[var(--bg-inset)] border border-[var(--border)] rounded overflow-hidden">
          <div className="px-3 py-2 border-b border-[var(--border)] text-sm font-medium text-[var(--blue)]">所有报告</div>
          <div className="max-h-48 overflow-y-auto">
            <table className="w-full text-xs">
              <thead>
                <tr className="text-[var(--text-tertiary)] border-b border-[var(--border-subtle)]">
                  <th className="px-3 py-1.5 text-left font-medium">名称</th>
                  <th className="px-3 py-1.5 text-left font-medium">目标</th>
                  <th className="px-3 py-1.5 text-left font-medium">状态</th>
                  <th className="px-3 py-1.5 text-right font-medium">并发</th>
                  <th className="px-3 py-1.5 text-right font-medium">请求数</th>
                  <th className="px-3 py-1.5 text-right font-medium">操作</th>
                </tr>
              </thead>
              <tbody>
                {reports.map((report) => (
                  <tr
                    key={report.id}
                    onClick={() => setSelected(report)}
                    className={`cursor-pointer border-b border-[var(--border-subtle)] ${selected?.id === report.id ? 'bg-[var(--selected-bg)]' : 'hover:bg-[var(--hover-bg)]'}`}
                  >
                    <td className="px-3 py-1.5 truncate max-w-[150px]">{report.name}</td>
                    <td className="px-3 py-1.5 font-mono truncate max-w-[200px]">{report.config.targetUrl}</td>
                    <td className={`px-3 py-1.5 ${getStatusColor(report)}`}>{getStatusLabel(report)}</td>
                    <td className="px-3 py-1.5 text-right">{report.config.concurrency}</td>
                    <td className="px-3 py-1.5 text-right">{report.config.totalRequests}</td>
                    <td className="px-3 py-1.5 text-right">
                      <button
                        onClick={(e) => { e.stopPropagation(); handleDelete(report.id) }}
                        className="text-[var(--red)] hover:text-[var(--red)]/90"
                      >
                        ✕
                      </button>
                    </td>
                  </tr>
                ))}
                {reports.length === 0 && (
                  <tr><td colSpan={6} className="px-3 py-4 text-center text-[var(--text-tertiary)]">暂无报告</td></tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  )
}
