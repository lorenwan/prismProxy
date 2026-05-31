import { useEffect, useState } from 'react'
import { Wifi, WifiOff, Globe, Clock, Server } from 'lucide-react'
import { useTrafficStore } from '../../stores/trafficStore'
import { getTrafficStats } from '../../services/traffic'
import type { TrafficStats } from '../../types'

export default function StatusBar() {
  const { trafficList } = useTrafficStore()
  const [stats, setStats] = useState<TrafficStats | null>(null)
  const [wsStatus] = useState<'connected' | 'disconnected' | 'connecting'>('disconnected')

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const data = await getTrafficStats()
        setStats(data.data)
      } catch {
        // 静默处理
      }
    }
    fetchStats()
    const timer = setInterval(fetchStats, 5000)
    return () => clearInterval(timer)
  }, [])

  const wsIcon = wsStatus === 'connected' ? <Wifi size={12} /> : <WifiOff size={12} />
  const wsColor = wsStatus === 'connected'
    ? 'text-[#3fb950]'
    : wsStatus === 'connecting'
      ? 'text-[#d29922]'
      : 'text-[#f85149]'

  return (
    <footer className="h-6 bg-[#161b22] border-t border-[#30363d] flex items-center px-3 text-[11px] text-[#8b949e] shrink-0 select-none">
      {/* WebSocket 状态 */}
      <div className={`flex items-center gap-1 ${wsColor}`}>
        {wsIcon}
        <span>{wsStatus === 'connected' ? '已连接' : wsStatus === 'connecting' ? '连接中...' : '未连接'}</span>
      </div>

      <span className="mx-3 text-[#30363d]">|</span>

      {/* 流量统计 */}
      <div className="flex items-center gap-1">
        <Globe size={12} />
        <span>请求: {trafficList.length}</span>
      </div>

      {stats && (
        <>
          <span className="mx-3 text-[#30363d]">|</span>
          <div className="flex items-center gap-1">
            <Clock size={12} />
            <span>
              平均耗时: {stats.avgDuration.toFixed(0)}ms
            </span>
          </div>
          <span className="mx-3 text-[#30363d]">|</span>
          <span>
            成功率: {stats.totalRequests > 0 ? Math.round((stats.successRequests / stats.totalRequests) * 100) : 0}%
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
