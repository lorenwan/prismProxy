import { useEffect, useState } from 'react'
import { useTrafficStore } from '../../stores/trafficStore'
import { getTrafficStats } from '../../services/traffic'
import type { TrafficStats } from '../../types'

export default function StatusBar() {
  const { trafficList } = useTrafficStore()
  const [stats, setStats] = useState<TrafficStats | null>(null)
  const [wsStatus] = useState<'connected' | 'disconnected' | 'connecting'>('disconnected')

  // 获取统计数据
  useEffect(() => {
    const fetchStats = async () => {
      try {
        const data = await getTrafficStats()
        setStats(data.data)
      } catch (err) {
        console.error('获取统计失败:', err)
      }
    }
    fetchStats()
    const timer = setInterval(fetchStats, 5000)
    return () => clearInterval(timer)
  }, [])

  return (
    <footer className="h-6 bg-[#16161e] border-t border-[#3b4261] flex items-center px-4 text-xs text-[#565f89]">
      {/* WebSocket 状态 */}
      <div className="flex items-center gap-1">
        <span
          className={`w-2 h-2 rounded-full ${
            wsStatus === 'connected'
              ? 'bg-[#9ece6a]'
              : wsStatus === 'connecting'
                ? 'bg-[#e0af68]'
                : 'bg-[#f7768e]'
          }`}
        />
        <span>{wsStatus === 'connected' ? '已连接' : wsStatus === 'connecting' ? '连接中...' : '未连接'}</span>
      </div>

      <span className="mx-4">|</span>

      {/* 流量统计 */}
      <span>总请求数: {trafficList.length}</span>

      {stats && (
        <>
          <span className="mx-4">|</span>
          <span>成功率: {stats.totalRequests > 0 ? Math.round((stats.successRequests / stats.totalRequests) * 100) : 0}%</span>
        </>
      )}

      <span className="mx-4">|</span>

      {/* 代理端口 */}
      <span>代理端口: 8081</span>
    </footer>
  )
}