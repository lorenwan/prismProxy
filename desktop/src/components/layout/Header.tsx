import { useState, useEffect, useCallback } from 'react'
import { Search, Trash2, Play, Square, Filter, X } from 'lucide-react'
import { useTrafficStore } from '../../features/traffic/trafficStore'
import { clearTraffic } from '../../features/traffic/trafficService'
import { Button } from '../ui/button'
import { Select } from '../ui/select'
import { Tooltip } from '../ui/tooltip'

export default function Header() {
  const { filters, setFilters, clearTraffic: clearLocal } = useTrafficStore()
  const [searchValue, setSearchValue] = useState(filters.host || '')

  // 搜索防抖
  useEffect(() => {
    const timer = setTimeout(() => {
      setFilters({ host: searchValue || undefined })
    }, 300)
    return () => clearTimeout(timer)
  }, [searchValue, setFilters])

  // 同步外部 filter 变化
  useEffect(() => {
    if (filters.host !== (searchValue || undefined)) {
      setSearchValue(filters.host || '')
    }
  }, [filters.host])

  const handleClear = useCallback(async () => {
    if (confirm('确定要清空所有流量记录吗？')) {
      await clearTraffic()
      clearLocal()
    }
  }, [clearLocal])

  const clearSearch = useCallback(() => {
    setSearchValue('')
    setFilters({ host: undefined })
  }, [setFilters])

  const hasActiveFilters = filters.method || filters.status || filters.host

  return (
    <header className="h-11 bg-[var(--bg-secondary)] border-b border-[var(--border)] flex items-center px-3 gap-2 shrink-0" role="banner">
      {/* 搜索框 */}
      <div className="flex-1 max-w-sm relative">
        <Search size={14} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-[var(--text-tertiary)] pointer-events-none" />
        <input
          type="text"
          placeholder="搜索 Host、Path、URL..."
          value={searchValue}
          onChange={(e) => setSearchValue(e.target.value)}
          className="w-full h-7 pl-8 pr-7 text-xs bg-[var(--bg-inset)] border border-[var(--border)] rounded focus:outline-none focus:border-[var(--blue)] focus:ring-1 focus:ring-[var(--blue)]/30 placeholder:text-[var(--text-tertiary)] transition-colors"
          aria-label="搜索流量"
        />
        {searchValue && (
          <button
            onClick={clearSearch}
            className="absolute right-1.5 top-1/2 -translate-y-1/2 p-0.5 text-[var(--text-tertiary)] hover:text-[var(--text-primary)] transition-colors"
            aria-label="清除搜索"
          >
            <X size={12} />
          </button>
        )}
      </div>

      {/* 过滤器 */}
      <div className="flex items-center gap-1.5">
        <Tooltip content={hasActiveFilters ? '已启用过滤' : '过滤请求'}>
          <Filter size={14} className={hasActiveFilters ? 'text-[var(--blue)]' : 'text-[var(--text-secondary)]'} />
        </Tooltip>

        <Select
          value={filters.method || ''}
          onChange={(e) => setFilters({ method: e.target.value || undefined })}
          aria-label="按方法过滤"
          className="h-7 text-xs w-24"
        >
          <option value="">全部方法</option>
          <option value="GET">GET</option>
          <option value="POST">POST</option>
          <option value="PUT">PUT</option>
          <option value="DELETE">DELETE</option>
          <option value="PATCH">PATCH</option>
        </Select>

        <Select
          value={filters.status?.toString() || ''}
          onChange={(e) => setFilters({ status: e.target.value ? Number(e.target.value) : undefined })}
          aria-label="按状态码过滤"
          className="h-7 text-xs w-28"
        >
          <option value="">全部状态</option>
          <option value="200">2xx 成功</option>
          <option value="300">3xx 重定向</option>
          <option value="400">4xx 客户端错误</option>
          <option value="500">5xx 服务端错误</option>
        </Select>
      </div>

      {/* 分隔线 */}
      <div className="w-px h-5 bg-[var(--border)]" aria-hidden="true" />

      {/* 工具按钮 */}
      <Tooltip content="开始捕获 HTTP 流量">
        <Button variant="ghost" size="sm" className="h-7 text-xs gap-1" aria-label="开始抓包">
          <Play size={12} />
          抓包
        </Button>
      </Tooltip>

      <Tooltip content="停止捕获 HTTP 流量">
        <Button variant="ghost" size="sm" className="h-7 text-xs gap-1" aria-label="停止抓包">
          <Square size={12} />
          停止
        </Button>
      </Tooltip>

      <Tooltip content="清空所有流量记录">
        <Button variant="destructive" size="sm" className="h-7 text-xs gap-1" onClick={handleClear} aria-label="清空流量">
          <Trash2 size={12} />
          清空
        </Button>
      </Tooltip>
    </header>
  )
}
