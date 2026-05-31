import { useEffect } from 'react'
import TrafficList from '../components/traffic/TrafficList'
import TrafficDetail from '../components/traffic/TrafficDetail'
import { useWebSocket } from '../hooks/useWebSocket'
import { useTrafficStore } from '../stores/trafficStore'
import { getTrafficList } from '../services/traffic'

export default function TrafficPage() {
  const { setTrafficList, setLoading } = useTrafficStore()

  // 连接 WebSocket
  useWebSocket()

  // 加载初始数据
  useEffect(() => {
    const loadTraffic = async () => {
      setLoading(true)
      try {
        const res = await getTrafficList({ pageSize: 100 })
        setTrafficList(res.data.data)
      } catch (err) {
        console.error('加载流量失败:', err)
      } finally {
        setLoading(false)
      }
    }
    loadTraffic()
  }, [setTrafficList, setLoading])

  return (
    <div className="flex h-full">
      {/* 左侧列表 */}
      <div className="w-1/2 border-r border-[#3b4261]">
        <TrafficList />
      </div>

      {/* 右侧详情 */}
      <div className="w-1/2">
        <TrafficDetail />
      </div>
    </div>
  )
}