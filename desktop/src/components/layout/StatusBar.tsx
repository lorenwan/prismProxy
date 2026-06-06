import { useEffect, useState } from 'react'
import { Globe, Clock, Server } from 'lucide-react'
import { useTrafficStore } from '../../features/traffic/trafficStore'
import { getTrafficStats } from '../../features/traffic/trafficService'
import { getProxyStatus } from '../../services/proxy'
import type { TrafficStats } from '../../types'

export default function StatusBar() {
  const { trafficList } = useTrafficStore()
  const [stats, setStats] = useState<TrafficStats | null>(null)
  const [proxyRunning, setProxyRunning] = useState(false)

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const data = await getTrafficStats()
        setStats(data.data)
      } catch {
        // 静默处理
      }
    }

    const fetchProxyStatus = async () => {
      try {
        const status = await getProxyStatus()
        setProxyRunning(status.running)
      } catch {
        // 静默处理
      }
    }

    fetchStats()
    fetchProxyStatus()
    const timer = setInterval(() => {
      fetchStats()
      fetchProxyStatus()
    }, 5000)
    return () => clearInterval(timer)
  }, [])

  return (
    <footer className="h-6 bg-[var(--bg-secondary)] border-t border-[var(--border)] flex items-center px-3 text-[11px] text-[var(--text-secondary)] shrink-0 select-none">
      {/* 连接状态 */}
      <div className="flex items-center gap-1">
        <div className={`w-2 h-2 rounded-full ${proxyRunning ? 'bg-[var(--green)]' : 'bg-[var(--red)]'}`} />
        <span>{proxyRunning ? '已连接' : '未连接'}</span>
      </div>

      <span className="mx-3 text-[var(--border)]">|</span>

      {/* 流量统计 */}
      <div className="flex items-center gap-1">
        <Globe size={12} />
        <span>请求: {trafficList.length}</span>
      </div>

      {stats && (
        <>
          <span className="mx-3 text-[var(--border)]">|</span>
          <div className="flex items-center gap-1">
            <Clock size={12} />
            <span>
              平均耗时: {stats.avg_duration_ms.toFixed(0)}ms
            </span>
          </div>
          <span className="mx-3 text-[var(--border)]">|</span>
          <span>
            成功率: {stats.total_requests > 0 ? Math.round((stats.success_count / stats.total_requests) * 100) : 0}%
          </span>
        </>
      )}

      {/* 右侧信息 */}
      <div className="ml-auto flex items-center gap-1">
        <Server size={12} />
        <span>代理端口: 8081</span>
      </div>
    </footer>
  )
}
