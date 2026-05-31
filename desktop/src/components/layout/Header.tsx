import { Search, Trash2, Play, Square, Filter } from 'lucide-react'
import { useTrafficStore } from '../../stores/trafficStore'
import { clearTraffic } from '../../services/traffic'
import Button from '../ui/Button'
import Input from '../ui/Input'
import Select from '../ui/Select'

export default function Header() {
  const { filters, setFilters, clearTraffic: clearLocal } = useTrafficStore()

  const handleClear = async () => {
    if (confirm('确定要清空所有流量记录吗？')) {
      await clearTraffic()
      clearLocal()
    }
  }

  return (
    <header className="h-11 bg-[#161b22] border-b border-[#30363d] flex items-center px-3 gap-2 shrink-0">
      {/* 搜索框 */}
      <div className="flex-1 max-w-sm">
        <Input
          icon={<Search size={14} />}
          placeholder="搜索 Host、Path、URL..."
          value={filters.host || ''}
          onChange={(e) => setFilters({ host: e.target.value || undefined })}
        />
      </div>

      {/* 过滤器 */}
      <div className="flex items-center gap-1.5">
        <Filter size={14} className="text-[#8b949e]" />
        <Select
          value={filters.method || ''}
          onChange={(e) => setFilters({ method: e.target.value || undefined })}
        >
          <option value="">全部方法</option>
          <option value="GET">GET</option>
          <option value="POST">POST</option>
          <option value="PUT">PUT</option>
          <option value="DELETE">DELETE</option>
          <option value="PATCH">PATCH</option>
        </Select>

        <Select
          value={filters.status || ''}
          onChange={(e) => setFilters({ status: e.target.value ? Number(e.target.value) : undefined })}
        >
          <option value="">全部状态</option>
          <option value="200">2xx 成功</option>
          <option value="300">3xx 重定向</option>
          <option value="400">4xx 客户端错误</option>
          <option value="500">5xx 服务端错误</option>
        </Select>
      </div>

      {/* 分隔线 */}
      <div className="w-px h-5 bg-[#30363d]" />

      {/* 工具按钮 */}
      <Button variant="ghost" size="sm" icon={<Play size={14} />} title="开始抓包">
        抓包
      </Button>
      <Button variant="ghost" size="sm" icon={<Square size={14} />} title="停止抓包">
        停止
      </Button>
      <Button variant="danger" size="sm" icon={<Trash2 size={14} />} onClick={handleClear}>
        清空
      </Button>
    </header>
  )
}
