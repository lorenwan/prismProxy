import { useState, useMemo } from 'react'
import { useTrafficStore } from '../trafficStore'
import { getStatusColor, getMethodColor, formatSize, formatBody } from '../../../lib/traffic-utils'
import type { Transaction } from '../../../types'

type TabType = 'request' | 'response' | 'overview'

// 概览字段配置
const overviewFields = [
  { label: 'URL', key: 'url' as const, full: true },
  { label: '状态码', key: 'status' as const },
  { label: '方法', key: 'method' as const },
  { label: '耗时', key: 'duration' as const },
  { label: '时间', key: 'time' as const },
  { label: 'Content-Type', key: 'contentType' as const },
  { label: '大小', key: 'size' as const },
]

function getFieldValue(item: Transaction, key: string) {
  switch (key) {
    case 'url': return item.url
    case 'status': return { value: item.response?.status_code ?? '-', className: getStatusColor(item.response?.status_code ?? 0) }
    case 'method': return { value: item.method, className: getMethodColor(item.method) }
    case 'duration': return `${item.duration_ms >= 1000 ? `${(item.duration_ms / 1000).toFixed(1)}s` : `${item.duration_ms}ms`}`
    case 'time': return new Date(item.timestamp).toLocaleString()
    case 'contentType': return item.response?.content_type ?? '-'
    case 'size': return formatSize(item.response?.body_size ?? 0)
    default: return '-'
  }
}

// Tab 配置
const tabs: { id: TabType; label: string }[] = [
  { id: 'overview', label: '概览' },
  { id: 'request', label: '请求' },
  { id: 'response', label: '响应' },
]

export default function TrafficDetail() {
  const { trafficList, selectedId } = useTrafficStore()
  const [activeTab, setActiveTab] = useState<TabType>('overview')

  const selected = useMemo(() => trafficList.find((t) => t.id === selectedId), [trafficList, selectedId])

  if (!selected) {
    return (
      <div className="flex flex-col items-center justify-center h-full gap-4 text-[var(--text-tertiary)] animate-fade-in">
        <div className="w-16 h-16 rounded-2xl bg-[var(--bg-inset)] flex items-center justify-center">
          <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
            <path d="M9 5H2v7l6.29 6.29c.94.94 2.48.94 3.42 0l3.58-3.58c.94-.94.94-2.48 0-3.42L9 5Z" />
            <path d="M6 9.01V9" />
            <path d="m15 5 6.3 6.3a2.4 2.4 0 0 1 0 3.4L17 19" />
          </svg>
        </div>
        <div className="text-center">
          <p className="text-sm font-medium text-[var(--text-secondary)]">选择一条流量查看详情</p>
          <p className="text-xs mt-1.5 text-[var(--text-tertiary)]">从左侧列表中点击任意请求</p>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col h-full">
      {/* Tab 栏 */}
      <div className="h-10 bg-[var(--bg-secondary)] border-b border-[var(--border)] flex items-center px-1" role="tablist" aria-label="请求详情">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={`
              relative px-3 py-1.5 text-xs font-medium rounded transition-colors
              ${activeTab === tab.id
                ? 'text-[var(--blue)]'
                : 'text-[var(--text-tertiary)] hover:text-[var(--text-primary)]'
              }
            `}
            role="tab"
            aria-selected={activeTab === tab.id}
            aria-controls={`tab-panel-${tab.id}`}
          >
            {tab.label}
            {activeTab === tab.id && (
              <span className="absolute bottom-0 left-1/2 -translate-x-1/2 w-4 h-0.5 bg-[var(--blue)] rounded-full" />
            )}
          </button>
        ))}
      </div>

      {/* 内容区 */}
      <div className="flex-1 overflow-auto p-4">
        {activeTab === 'overview' && (
          <div className="space-y-1 animate-fade-in" id="tab-panel-overview" role="tabpanel" aria-label="概览">
            {overviewFields.map((field) => {
              const fieldValue = getFieldValue(selected, field.key)
              const isObject = typeof fieldValue === 'object' && fieldValue !== null
              return (
                <div key={field.key} className={`flex items-start py-2 ${field.full ? 'flex-col gap-1' : ''}`}>
                  <span className="text-xs text-[var(--text-tertiary)] shrink-0 w-24">{field.label}</span>
                  <span className={`text-sm break-all ${isObject ? fieldValue.className : 'text-[var(--text-primary)]'}`}>
                    {isObject ? fieldValue.value : fieldValue}
                  </span>
                </div>
              )
            })}
          </div>
        )}

        {activeTab === 'request' && (
          <div className="space-y-4 animate-fade-in" id="tab-panel-request" role="tabpanel" aria-label="请求">
            <div>
              <h3 className="text-xs font-medium text-[var(--text-tertiary)] mb-2 uppercase tracking-wider">请求头</h3>
              <div className="bg-[var(--bg-inset)] rounded-lg p-3 font-mono text-xs divide-y divide-[var(--border-subtle)]">
                {Object.entries(selected.request.headers).map(([key, val]) => (
                  <div key={key} className="flex py-1.5 first:pt-0 last:pb-0">
                    <span className="text-[var(--blue)] w-44 shrink-0 font-medium">{key}</span>
                    <span className="text-[var(--text-secondary)] break-all">{val.values?.join(', ') ?? ''}</span>
                  </div>
                ))}
              </div>
            </div>
            {selected.request.body && (
              <div>
                <h3 className="text-xs font-medium text-[var(--text-tertiary)] mb-2 uppercase tracking-wider">请求体</h3>
                <pre className="bg-[var(--bg-inset)] rounded-lg p-3 font-mono text-xs text-[var(--text-secondary)] overflow-auto max-h-60 whitespace-pre-wrap break-all">
                  {selected.request.body}
                </pre>
              </div>
            )}
          </div>
        )}

        {activeTab === 'response' && (
          <div className="space-y-4 animate-fade-in" id="tab-panel-response" role="tabpanel" aria-label="响应">
            <div>
              <h3 className="text-xs font-medium text-[var(--text-tertiary)] mb-2 uppercase tracking-wider">响应头</h3>
              <div className="bg-[var(--bg-inset)] rounded-lg p-3 font-mono text-xs divide-y divide-[var(--border-subtle)]">
                {Object.entries(selected.response.headers).map(([key, val]) => (
                  <div key={key} className="flex py-1.5 first:pt-0 last:pb-0">
                    <span className="text-[var(--blue)] w-44 shrink-0 font-medium">{key}</span>
                    <span className="text-[var(--text-secondary)] break-all">{val.values?.join(', ') ?? ''}</span>
                  </div>
                ))}
              </div>
            </div>
            {selected.response.body && (
              <div>
                <h3 className="text-xs font-medium text-[var(--text-tertiary)] mb-2 uppercase tracking-wider">响应体</h3>
                <pre className="bg-[var(--bg-inset)] rounded-lg p-3 font-mono text-xs text-[var(--text-secondary)] overflow-auto max-h-60 whitespace-pre-wrap break-all">
                  {formatBody(selected.response.body, selected.response.content_type)}
                </pre>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
