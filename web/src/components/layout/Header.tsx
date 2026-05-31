import { useTrafficStore } from '../../stores/trafficStore'
import { clearTraffic } from '../../services/traffic'

export default function Header() {
  const { filters, setFilters, clearTraffic: clearLocal } = useTrafficStore()

  // 清空流量
  const handleClear = async () => {
    if (confirm('确定要清空所有流量记录吗？')) {
      await clearTraffic()
      clearLocal()
    }
  }

  return (
    <header className="h-12 bg-[#24283b] border-b border-[#3b4261] flex items-center px-4 gap-4">
      {/* 搜索框 */}
      <div className="flex-1 max-w-md">
        <input
          type="text"
          placeholder="搜索 Host、Path..."
          value={filters.host || ''}
          onChange={(e) => setFilters({ host: e.target.value || undefined })}
          className="w-full h-8 px-3 bg-[#1a1b26] border border-[#3b4261] rounded text-sm text-[#c0caf5] placeholder-[#565f89] focus:outline-none focus:border-[#7aa2f7]"
        />
      </div>

      {/* 过滤器 */}
      <div className="flex gap-2">
        <select
          value={filters.method || ''}
          onChange={(e) => setFilters({ method: e.target.value || undefined })}
          className="h-8 px-2 bg-[#1a1b26] border border-[#3b4261] rounded text-sm text-[#a9b1d6] focus:outline-none focus:border-[#7aa2f7]"
        >
          <option value="">全部方法</option>
          <option value="GET">GET</option>
          <option value="POST">POST</option>
          <option value="PUT">PUT</option>
          <option value="DELETE">DELETE</option>
        </select>

        <select
          value={filters.status || ''}
          onChange={(e) => setFilters({ status: e.target.value ? Number(e.target.value) : undefined })}
          className="h-8 px-2 bg-[#1a1b26] border border-[#3b4261] rounded text-sm text-[#a9b1d6] focus:outline-none focus:border-[#7aa2f7]"
        >
          <option value="">全部状态</option>
          <option value="200">2xx 成功</option>
          <option value="300">3xx 重定向</option>
          <option value="400">4xx 客户端错误</option>
          <option value="500">5xx 服务端错误</option>
        </select>
      </div>

      {/* 工具按钮 */}
      <button
        onClick={handleClear}
        className="h-8 px-3 bg-[#f7768e] hover:bg-[#ff899d] text-white rounded text-sm transition-colors"
      >
        清空
      </button>
    </header>
  )
}