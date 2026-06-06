import { useEffect, useState, useCallback } from 'react'
import TrafficList from '../features/traffic/components/TrafficList'
import TrafficDetail from '../features/traffic/components/TrafficDetail'
import { useWebSocket } from '../hooks/useWebSocket'
import { useTrafficStore } from '../features/traffic/trafficStore'
import { getTrafficList } from '../features/traffic/trafficService'

export default function TrafficPage() {
  const { setTrafficList, setLoading } = useTrafficStore()
  const [error, setError] = useState<string | null>(null)

  // 连接 WebSocket
  useWebSocket()

  // 加载初始数据
  const loadTraffic = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const res = await getTrafficList({ pageSize: 100 })
      setTrafficList(res.data.data)
    } catch (err) {
      const message = err instanceof Error ? err.message : '加载流量失败'
      console.error('加载流量失败:', err)
      setError(message)
    } finally {
      setLoading(false)
    }
  }, [setTrafficList, setLoading])

  useEffect(() => {
    loadTraffic()
  }, [loadTraffic])

  // 错误状态
  if (error) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-center space-y-3">
          <p className="text-sm text-[var(--red)]">{error}</p>
          <button
            onClick={loadTraffic}
            className="px-3 py-1.5 text-xs bg-[var(--bg-secondary)] border border-[var(--border)] rounded hover:bg-[var(--hover-bg)] transition-colors"
          >
            重试
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="flex h-full min-w-0">
      {/* 左侧列表 - 使用 min-w-0 防止溢出 */}
      <div className="w-1/2 min-w-[320px] border-r border-[var(--border)]">
        <TrafficList />
      </div>

      {/* 右侧详情 */}
      <div className="w-1/2 min-w-[320px]">
        <TrafficDetail />
      </div>
    </div>
  )
}