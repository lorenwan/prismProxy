import { useEffect, useRef, useCallback, useMemo } from 'react'
import { useTrafficStore } from '../trafficStore'
import { getStatusColor, getMethodColor } from '../../../lib/traffic-utils'

export default function TrafficList() {
  const { trafficList, selectedId, setSelectedId, filters, loading } = useTrafficStore()
  const listRef = useRef<HTMLDivElement>(null)

  // 使用 useMemo 缓存过滤结果，避免每次渲染重新计算
  const filteredList = useMemo(() => {
    return trafficList.filter((item) => {
      if (filters.method && item.method !== filters.method) return false
      if (filters.status && !(item.response?.status_code >= filters.status && item.response?.status_code < filters.status + 100)) return false
      if (filters.host && !item.host.includes(filters.host)) return false
      return true
    })
  }, [trafficList, filters])

  // 键盘导航
  const handleKeyDown = useCallback((e: KeyboardEvent) => {
    if (!filteredList.length) return

    const currentIndex = filteredList.findIndex((item) => item.id === selectedId)

    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault()
        if (currentIndex < filteredList.length - 1) {
          setSelectedId(filteredList[currentIndex + 1].id)
        } else if (currentIndex === -1) {
          setSelectedId(filteredList[0].id)
        }
        break
      case 'ArrowUp':
        e.preventDefault()
        if (currentIndex > 0) {
          setSelectedId(filteredList[currentIndex - 1].id)
        } else if (currentIndex === -1) {
          setSelectedId(filteredList[filteredList.length - 1].id)
        }
        break
      case 'Home':
        e.preventDefault()
        setSelectedId(filteredList[0].id)
        break
      case 'End':
        e.preventDefault()
        setSelectedId(filteredList[filteredList.length - 1].id)
        break
      case 'Delete':
        if (selectedId) {
          const store = useTrafficStore.getState()
          store.removeTraffic(selectedId)
        }
        break
    }
  }, [filteredList, selectedId, setSelectedId])

  useEffect(() => {
    const list = listRef.current
    if (!list) return

    list.addEventListener('keydown', handleKeyDown)
    return () => list.removeEventListener('keydown', handleKeyDown)
  }, [handleKeyDown])

  // 滚动到选中项
  useEffect(() => {
    if (!selectedId || !listRef.current) return
    const selectedElement = listRef.current.querySelector(`[data-id="${selectedId}"]`)
    if (selectedElement) {
      selectedElement.scrollIntoView({ block: 'nearest' })
    }
  }, [selectedId])

  // 加载状态
  if (loading && trafficList.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-full gap-3 text-[var(--text-tertiary)]">
        <div className="w-8 h-8 border-2 border-[var(--blue)] border-t-transparent rounded-full animate-spin" />
        <p className="text-sm">加载中...</p>
      </div>
    )
  }

  return (
    <div
      ref={listRef}
      className="flex flex-col h-full overflow-hidden"
      role="grid"
      aria-label="流量列表"
      tabIndex={0}
    >
      {/* 表头 */}
      <div
        className="h-9 bg-[var(--bg-secondary)] border-b border-[var(--border)] flex items-center px-3 text-xs font-medium text-[var(--text-tertiary)] select-none"
        role="row"
      >
        <span className="w-[72px]" role="columnheader">状态码</span>
        <span className="w-[64px]" role="columnheader">方法</span>
        <span className="flex-1 min-w-0" role="columnheader">Host</span>
        <span className="flex-1 min-w-0" role="columnheader">Path</span>
        <span className="w-[72px] text-right" role="columnheader">耗时</span>
        <span className="w-[80px] text-right" role="columnheader">时间</span>
      </div>

      {/* 列表 */}
      <div className="flex-1 overflow-y-auto" role="rowgroup">
        {loading ? (
          <div className="flex items-center justify-center h-full gap-2 text-[var(--text-tertiary)] animate-fade-in">
            <svg className="animate-spin h-5 w-5" viewBox="0 0 24 24">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
            </svg>
            <span className="text-sm">加载中...</span>
          </div>
        ) : filteredList.length === 0 ? (
          <div
            className="flex flex-col items-center justify-center h-full gap-3 text-[var(--text-tertiary)] animate-fade-in"
            role="status"
          >
            <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
              <path d="M22 12h-4l-3 9L9 3l-3 9H2" />
            </svg>
            <div className="text-center">
              <p className="text-sm font-medium text-[var(--text-secondary)]">
                {filters.method || filters.status || filters.host ? '没有匹配的流量' : '暂无流量数据'}
              </p>
              <p className="text-xs mt-1">
                {filters.method || filters.status || filters.host ? '尝试调整过滤条件' : '点击上方"抓包"按钮开始捕获 HTTP 流量'}
              </p>
            </div>
          </div>
        ) : (
          filteredList.map((item) => {
            const isSelected = selectedId === item.id
            return (
              <div
                key={item.id}
                data-id={item.id}
                onClick={() => setSelectedId(item.id)}
                className={`
                  h-8 flex items-center px-3 text-xs cursor-pointer
                  border-l-2 border-b border-b-[var(--border-subtle)]
                  transition-colors duration-75
                  ${isSelected
                    ? 'bg-[var(--selected-bg)] border-l-[var(--blue)] text-[var(--text-primary)]'
                    : 'border-l-transparent hover:bg-[var(--hover-bg)]'
                  }
                `}
                role="row"
                aria-selected={isSelected}
                tabIndex={-1}
              >
                <span className={`w-[72px] font-mono font-medium ${getStatusColor(item.response?.status_code ?? 0)}`} role="cell">
                  {item.response?.status_code ?? '-'}
                </span>
                <span className={`w-[64px] font-mono font-medium ${getMethodColor(item.method)}`} role="cell">
                  {item.method}
                </span>
                <span className="flex-1 min-w-0 truncate text-[var(--text-primary)]" role="cell">
                  {item.host}
                </span>
                <span className="flex-1 min-w-0 truncate text-[var(--text-secondary)]" role="cell">
                  {item.path}
                </span>
                <span className={`w-[72px] text-right font-mono ${item.duration_ms > 1000 ? 'text-[var(--yellow)]' : 'text-[var(--text-tertiary)]'}`} role="cell">
                  {item.duration_ms >= 1000 ? `${(item.duration_ms / 1000).toFixed(1)}s` : `${item.duration_ms}ms`}
                </span>
                <span className="w-[80px] text-right text-[var(--text-tertiary)] font-mono" role="cell">
                  {new Date(item.timestamp).toLocaleTimeString()}
                </span>
              </div>
            )
          })
        )}
      </div>

      {/* 底部状态栏 */}
      {trafficList.length > 0 && (
        <div className="h-7 bg-[var(--bg-secondary)] border-t border-[var(--border)] flex items-center px-3 text-xs text-[var(--text-tertiary)] select-none">
          <span>{filteredList.length} / {trafficList.length} 条记录</span>
        </div>
      )}
    </div>
  )
}
