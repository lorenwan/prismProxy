import { useState, useEffect } from 'react'
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

  // 从报告中提取概览数据
  const completedReports = reports.filter((r) => r.status === 'completed' && r.results)
  const totalRequests = completedReports.reduce((sum, r) => sum + (r.results?.totalRequests || 0), 0)
  const avgDuration = completedReports.length > 0
    ? Math.round(completedReports.reduce((sum, r) => sum + (r.results?.avgResponseTime || 0), 0) / completedReports.length)
    : 0
  const p50 = selected?.results?.p50ResponseTime ?? 0
  const p90 = selected?.results?.p90ResponseTime ?? 0
  const p99 = selected?.results?.p99ResponseTime ?? 0
  const errorRate = selected?.results?.errorRate ?? 0

  // 慢请求：取所有报告中 p90 以上的请求
  const slowRequests = completedReports
    .filter((r) => r.results && r.results.avgResponseTime > 500)
    .sort((a, b) => (b.results?.avgResponseTime || 0) - (a.results?.avgResponseTime || 0))

  // 域名统计（从报告 config 中提取）
  const domainStats = new Map<string, { count: number; totalDuration: number; errors: number }>()
  completedReports.forEach((r) => {
    try {
      const url = new URL(r.config.targetUrl)
      const domain = url.hostname
      const existing = domainStats.get(domain) || { count: 0, totalDuration: 0, errors: 0 }
      existing.count++
      existing.totalDuration += r.results?.avgResponseTime || 0
      existing.errors += r.results?.failedRequests || 0
      domainStats.set(domain, existing)
    } catch {}
  })
  const domainStatsList = Array.from(domainStats.entries())
    .map(([domain, stats]) => ({
      domain,
      requests: stats.count,
      avgDuration: Math.round(stats.totalDuration / stats.count),
      errorRate: stats.count > 0 ? ((stats.errors / stats.count) * 100).toFixed(1) : '0',
    }))
    .sort((a, b) => b.requests - a.requests)

  // 时间线数据
  const timeline = selected?.results?.timeline || []
  const maxRps = Math.max(...timeline.map((t) => t.rps), 1)

  function getStatusColor(report: PerformanceReport) {
    if (report.status === 'running') return 'text-[#7aa2f7]'
    if (report.status === 'failed') return 'text-[#f7768e]'
    return 'text-[#9ece6a]'
  }

  function getStatusLabel(report: PerformanceReport) {
    if (report.status === 'running') return '运行中'
    if (report.status === 'failed') return '失败'
    return '完成'
  }

  return (
    <div className="flex flex-col h-full bg-[#24283b] overflow-y-auto">
      <div className="p-4 space-y-4">
        <h1 className="text-lg font-semibold">性能分析</h1>

        {/* 概览卡片 */}
        <div className="grid grid-cols-5 gap-3">
          {[
            { label: '总请求数', value: totalRequests.toLocaleString(), color: '#7aa2f7' },
            { label: '平均耗时', value: `${avgDuration}ms`, color: '#9ece6a' },
            { label: 'P50', value: `${p50}ms`, color: '#e0af68' },
            { label: 'P90', value: `${p90}ms`, color: '#ff9e6d' },
            { label: '错误率', value: `${(errorRate * 100).toFixed(1)}%`, color: '#f7768e' },
          ].map((card) => (
            <div key={card.label} className="bg-[#1a1b26] border border-[#3b4261] rounded p-3">
              <div className="text-xs text-[#565f89] mb-1">{card.label}</div>
              <div className="text-xl font-semibold" style={{ color: card.color }}>{card.value}</div>
            </div>
          ))}
        </div>

        {/* 选中报告的详细指标 */}
        {selected?.results && (
          <div className="grid grid-cols-4 gap-3">
            <div className="bg-[#1a1b26] border border-[#3b4261] rounded p-3">
              <div className="text-xs text-[#565f89] mb-1">P99</div>
              <div className="text-lg font-semibold text-[#f7768e]">{selected.results.p99ResponseTime}ms</div>
            </div>
            <div className="bg-[#1a1b26] border border-[#3b4261] rounded p-3">
              <div className="text-xs text-[#565f89] mb-1">RPS</div>
              <div className="text-lg font-semibold text-[#7aa2f7]">{selected.results.requestsPerSecond.toFixed(1)}</div>
            </div>
            <div className="bg-[#1a1b26] border border-[#3b4261] rounded p-3">
              <div className="text-xs text-[#565f89] mb-1">吞吐量</div>
              <div className="text-lg font-semibold text-[#9ece6a]">{(selected.results.throughput / 1024).toFixed(1)} KB/s</div>
            </div>
            <div className="bg-[#1a1b26] border border-[#3b4261] rounded p-3">
              <div className="text-xs text-[#565f89] mb-1">总耗时</div>
              <div className="text-lg font-semibold text-[#e0af68]">{(selected.results.totalDuration / 1000).toFixed(1)}s</div>
            </div>
          </div>
        )}

        <div className="grid grid-cols-2 gap-4">
          {/* 慢请求列表 */}
          <div className="bg-[#1a1b26] border border-[#3b4261] rounded overflow-hidden">
            <div className="px-3 py-2 border-b border-[#3b4261] text-sm font-medium text-[#7aa2f7]">慢请求列表</div>
            <div className="max-h-64 overflow-y-auto">
              <table className="w-full text-xs">
                <thead>
                  <tr className="text-[#565f89] border-b border-[#24283b]">
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
                      className={`cursor-pointer border-b border-[#24283b] ${selected?.id === report.id ? 'bg-[#283457]' : 'hover:bg-[#24283b]'}`}
                    >
                      <td className="px-3 py-1.5 truncate max-w-[200px]">{report.name}</td>
                      <td className={`px-3 py-1.5 ${getStatusColor(report)}`}>{getStatusLabel(report)}</td>
                      <td className="px-3 py-1.5 text-right text-[#f7768e]">{report.results?.avgResponseTime}ms</td>
                      <td className="px-3 py-1.5 text-right">{report.results?.totalRequests}</td>
                      <td className="px-3 py-1.5 text-right">
                        <button
                          onClick={(e) => { e.stopPropagation(); handleDelete(report.id) }}
                          className="text-[#f7768e] hover:text-[#ff9eaf]"
                        >
                          ✕
                        </button>
                      </td>
                    </tr>
                  ))}
                  {slowRequests.length === 0 && (
                    <tr><td colSpan={5} className="px-3 py-4 text-center text-[#565f89]">暂无慢请求</td></tr>
                  )}
                </tbody>
              </table>
            </div>
          </div>

          {/* 域名统计 */}
          <div className="bg-[#1a1b26] border border-[#3b4261] rounded overflow-hidden">
            <div className="px-3 py-2 border-b border-[#3b4261] text-sm font-medium text-[#7aa2f7]">域名统计</div>
            <div className="max-h-64 overflow-y-auto">
              <table className="w-full text-xs">
                <thead>
                  <tr className="text-[#565f89] border-b border-[#24283b]">
                    <th className="px-3 py-1.5 text-left font-medium">域名</th>
                    <th className="px-3 py-1.5 text-right font-medium">请求数</th>
                    <th className="px-3 py-1.5 text-right font-medium">平均耗时</th>
                    <th className="px-3 py-1.5 text-right font-medium">错误率</th>
                  </tr>
                </thead>
                <tbody>
                  {domainStatsList.map((stat) => (
                    <tr key={stat.domain} className="border-b border-[#24283b] hover:bg-[#24283b]">
                      <td className="px-3 py-1.5 font-mono">{stat.domain}</td>
                      <td className="px-3 py-1.5 text-right">{stat.requests}</td>
                      <td className="px-3 py-1.5 text-right">{stat.avgDuration}ms</td>
                      <td className={`px-3 py-1.5 text-right ${parseFloat(stat.errorRate) > 5 ? 'text-[#f7768e]' : 'text-[#565f89]'}`}>
                        {stat.errorRate}%
                      </td>
                    </tr>
                  ))}
                  {domainStatsList.length === 0 && (
                    <tr><td colSpan={4} className="px-3 py-4 text-center text-[#565f89]">暂无数据</td></tr>
                  )}
                </tbody>
              </table>
            </div>
          </div>
        </div>

        {/* 时间线图表 */}
        {timeline.length > 0 && (
          <div className="bg-[#1a1b26] border border-[#3b4261] rounded overflow-hidden">
            <div className="px-3 py-2 border-b border-[#3b4261] text-sm font-medium text-[#7aa2f7]">
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
                          hasError ? 'bg-[#f7768e]' : 'bg-[#7aa2f7]'
                        }`}
                        style={{ height: `${height}%` }}
                      />
                      <div className="absolute bottom-full left-1/2 -translate-x-1/2 mb-1 hidden group-hover:block z-10">
                        <div className="bg-[#24283b] border border-[#3b4261] rounded px-2 py-1 text-[10px] whitespace-nowrap">
                          <div>RPS: {point.rps.toFixed(1)}</div>
                          <div>延迟: {point.avgLatency.toFixed(0)}ms</div>
                          <div>错误: {(point.errorRate * 100).toFixed(1)}%</div>
                        </div>
                      </div>
                    </div>
                  )
                })}
              </div>
              <div className="flex justify-between mt-1 text-[10px] text-[#565f89]">
                <span>开始</span>
                <span>结束</span>
              </div>
            </div>
          </div>
        )}

        {/* 所有报告列表 */}
        <div className="bg-[#1a1b26] border border-[#3b4261] rounded overflow-hidden">
          <div className="px-3 py-2 border-b border-[#3b4261] text-sm font-medium text-[#7aa2f7]">所有报告</div>
          <div className="max-h-48 overflow-y-auto">
            <table className="w-full text-xs">
              <thead>
                <tr className="text-[#565f89] border-b border-[#24283b]">
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
                    className={`cursor-pointer border-b border-[#24283b] ${selected?.id === report.id ? 'bg-[#283457]' : 'hover:bg-[#24283b]'}`}
                  >
                    <td className="px-3 py-1.5 truncate max-w-[150px]">{report.name}</td>
                    <td className="px-3 py-1.5 font-mono truncate max-w-[200px]">{report.config.targetUrl}</td>
                    <td className={`px-3 py-1.5 ${getStatusColor(report)}`}>{getStatusLabel(report)}</td>
                    <td className="px-3 py-1.5 text-right">{report.config.concurrency}</td>
                    <td className="px-3 py-1.5 text-right">{report.config.totalRequests}</td>
                    <td className="px-3 py-1.5 text-right">
                      <button
                        onClick={(e) => { e.stopPropagation(); handleDelete(report.id) }}
                        className="text-[#f7768e] hover:text-[#ff9eaf]"
                      >
                        ✕
                      </button>
                    </td>
                  </tr>
                ))}
                {reports.length === 0 && (
                  <tr><td colSpan={6} className="px-3 py-4 text-center text-[#565f89]">暂无报告</td></tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  )
}
