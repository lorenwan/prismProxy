import { useTrafficStore } from '../../stores/trafficStore'

// 状态码颜色
function getStatusColor(status: number): string {
  if (status >= 200 && status < 300) return 'text-[#9ece6a]'
  if (status >= 300 && status < 400) return 'text-[#7aa2f7]'
  if (status >= 400 && status < 500) return 'text-[#e0af68]'
  return 'text-[#f7768e]'
}

// 方法颜色
function getMethodColor(method: string): string {
  switch (method) {
    case 'GET': return 'text-[#9ece6a]'
    case 'POST': return 'text-[#7aa2f7]'
    case 'PUT': return 'text-[#e0af68]'
    case 'DELETE': return 'text-[#f7768e]'
    default: return 'text-[#a9b1d6]'
  }
}

export default function TrafficList() {
  const { trafficList, selectedId, setSelectedId, filters } = useTrafficStore()

  // 过滤流量
  const filteredList = trafficList.filter((item) => {
    if (filters.method && item.method !== filters.method) return false
    if (filters.status && !(item.response?.status_code >= filters.status && item.response?.status_code < filters.status + 100)) return false
    if (filters.host && !item.host.includes(filters.host)) return false
    return true
  })

  return (
    <div className="flex flex-col h-full overflow-hidden">
      {/* 表头 */}
      <div className="h-8 bg-[#16161e] border-b border-[#3b4261] flex items-center px-2 text-xs text-[#565f89]">
        <span className="w-16">状态码</span>
        <span className="w-16">方法</span>
        <span className="flex-1 min-w-0">Host</span>
        <span className="flex-1 min-w-0">Path</span>
        <span className="w-20 text-right">耗时</span>
        <span className="w-20 text-right">时间</span>
      </div>

      {/* 列表 */}
      <div className="flex-1 overflow-y-auto">
        {filteredList.length === 0 ? (
          <div className="flex items-center justify-center h-full text-[#565f89]">
            暂无流量数据
          </div>
        ) : (
          filteredList.map((item) => (
            <div
              key={item.id}
              onClick={() => setSelectedId(item.id)}
              className={`h-7 flex items-center px-2 text-xs cursor-pointer border-b border-[#1a1b26] hover:bg-[#24283b] ${
                selectedId === item.id ? 'traffic-row-selected' : ''
              }`}
            >
              <span className={`w-16 font-mono ${getStatusColor(item.response?.status_code ?? 0)}`}>
                {item.response?.status_code ?? '-'}
              </span>
              <span className={`w-16 font-mono ${getMethodColor(item.method)}`}>
                {item.method}
              </span>
              <span className="flex-1 min-w-0 truncate text-[#c0caf5]">
                {item.host}
              </span>
              <span className="flex-1 min-w-0 truncate">
                {item.path}
              </span>
              <span className="w-20 text-right text-[#565f89]">
                {item.duration_ms}ms
              </span>
              <span className="w-20 text-right text-[#565f89]">
                {new Date(item.timestamp).toLocaleTimeString()}
              </span>
            </div>
          ))
        )}
      </div>
    </div>
  )
}