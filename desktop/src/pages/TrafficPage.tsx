import { useEffect, useState, useCallback } from 'react'
import { invoke } from '@tauri-apps/api/core'
import { listen } from '@tauri-apps/api/event'
import TrafficList from '../features/traffic/components/TrafficList'
import TrafficDetail from '../features/traffic/components/TrafficDetail'
import { useTrafficStore } from '../features/traffic/trafficStore'
import { getTrafficList } from '../features/traffic/trafficService'
import type { TrafficEvent, Transaction } from '../types'

export default function TrafficPage() {
  const { setTrafficList, addTraffic, updateTraffic, removeTraffic, setLoading } = useTrafficStore()
  const [error, setError] = useState<string | null>(null)

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

  // 启动事件订阅
  useEffect(() => {
    invoke('subscribe_traffic').catch(console.error)

    const unlisten = listen('traffic:event', (event) => {
      try {
        const msg = JSON.parse(event.payload as string) as TrafficEvent
        switch (msg.type) {
          case 'new':
            addTraffic(msg.entry as Transaction)
            break
          case 'update':
            updateTraffic(msg.entry as Transaction)
            break
          case 'delete':
            removeTraffic(String(msg.entry.id))
            break
        }
      } catch (err) {
        console.error('处理流量事件失败:', err)
      }
    })

    return () => { unlisten.then(fn => fn()) }
  }, [addTraffic, updateTraffic, removeTraffic])

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