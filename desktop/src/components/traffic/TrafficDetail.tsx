import { useState } from 'react'
import { useTrafficStore } from '../../stores/trafficStore'

type TabType = 'request' | 'response' | 'overview'

export default function TrafficDetail() {
  const { trafficList, selectedId } = useTrafficStore()
  const [activeTab, setActiveTab] = useState<TabType>('overview')

  const selected = trafficList.find((t) => t.id === selectedId)

  if (!selected) {
    return (
      <div className="flex items-center justify-center h-full text-[#565f89]">
        选择一条流量查看详情
      </div>
    )
  }

  return (
    <div className="flex flex-col h-full">
      {/* Tab 栏 */}
      <div className="h-10 bg-[#16161e] border-b border-[#3b4261] flex items-center px-2 gap-1">
        {(['overview', 'request', 'response'] as TabType[]).map((tab) => (
          <button
            key={tab}
            onClick={() => setActiveTab(tab)}
            className={`px-3 py-1.5 text-xs rounded transition-colors ${
              activeTab === tab
                ? 'bg-[#283457] text-[#7aa2f7]'
                : 'text-[#565f89] hover:text-[#a9b1d6]'
            }`}
          >
            {tab === 'overview' ? '概览' : tab === 'request' ? '请求' : '响应'}
          </button>
        ))}
      </div>

      {/* 内容区 */}
      <div className="flex-1 overflow-auto p-4 text-sm">
        {activeTab === 'overview' && (
          <div className="space-y-3">
            <div>
              <span className="text-[#565f89]">URL: </span>
              <span className="text-[#c0caf5] break-all">{selected.url}</span>
            </div>
            <div>
              <span className="text-[#565f89]">状态码: </span>
              <span className={getStatusColor(selected.response?.status_code ?? 0)}>{selected.response?.status_code ?? '-'}</span>
            </div>
            <div>
              <span className="text-[#565f89]">方法: </span>
              <span className={getMethodColor(selected.method)}>{selected.method}</span>
            </div>
            <div>
              <span className="text-[#565f89]">耗时: </span>
              <span className="text-[#c0caf5]">{selected.duration_ms}ms</span>
            </div>
            <div>
              <span className="text-[#565f89]">时间: </span>
              <span className="text-[#c0caf5]">{new Date(selected.timestamp).toLocaleString()}</span>
            </div>
            <div>
              <span className="text-[#565f89]">Content-Type: </span>
              <span className="text-[#c0caf5]">{selected.response?.content_type ?? '-'}</span>
            </div>
            <div>
              <span className="text-[#565f89]">大小: </span>
              <span className="text-[#c0caf5]">{formatSize(selected.response?.body_size ?? 0)}</span>
            </div>
          </div>
        )}

        {activeTab === 'request' && (
          <div className="space-y-4">
            <div>
              <h3 className="text-xs text-[#565f89] mb-2">请求头</h3>
              <div className="bg-[#1a1b26] rounded p-2 font-mono text-xs">
                {Object.entries(selected.request.headers).map(([key, val]) => (
                  <div key={key} className="flex">
                    <span className="text-[#7aa2f7] w-40 shrink-0">{key}:</span>
                    <span className="text-[#a9b1d6] break-all">{val.values?.join(', ') ?? ''}</span>
                  </div>
                ))}
              </div>
            </div>
            {selected.request.body && (
              <div>
                <h3 className="text-xs text-[#565f89] mb-2">请求体</h3>
                <pre className="bg-[#1a1b26] rounded p-2 font-mono text-xs text-[#a9b1d6] overflow-auto max-h-60">
                  {selected.request.body}
                </pre>
              </div>
            )}
          </div>
        )}

        {activeTab === 'response' && (
          <div className="space-y-4">
            <div>
              <h3 className="text-xs text-[#565f89] mb-2">响应头</h3>
              <div className="bg-[#1a1b26] rounded p-2 font-mono text-xs">
                {Object.entries(selected.response.headers).map(([key, val]) => (
                  <div key={key} className="flex">
                    <span className="text-[#7aa2f7] w-40 shrink-0">{key}:</span>
                    <span className="text-[#a9b1d6] break-all">{val.values?.join(', ') ?? ''}</span>
                  </div>
                ))}
              </div>
            </div>
            {selected.response.body && (
              <div>
                <h3 className="text-xs text-[#565f89] mb-2">响应体</h3>
                <pre className="bg-[#1a1b26] rounded p-2 font-mono text-xs text-[#a9b1d6] overflow-auto max-h-60">
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

// 格式化大小
function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

// 格式化响应体
function formatBody(body: string, contentType: string): string {
  if (contentType.includes('json')) {
    try {
      return JSON.stringify(JSON.parse(body), null, 2)
    } catch {
      return body
    }
  }
  return body
}