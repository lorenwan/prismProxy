import { useState, useEffect, useRef } from 'react'
import {
  getCollections,
  createCollection,
  deleteCollection,
  addRequest,
  updateRequest,
  deleteRequest,
  sendRequest,
  importCollection,
} from '../services/collections'
import type { Collection, CollectionRequest } from '../types'
import { useErrorHandler } from '../lib/error-handler'

const methods = ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'HEAD', 'OPTIONS']

export default function CollectionsPage() {
  const handleError = useErrorHandler()
  const [collections, setCollections] = useState<Collection[]>([])
  const [selectedCollection, setSelectedCollection] = useState<Collection | null>(null)
  const [selectedRequest, setSelectedRequest] = useState<CollectionRequest | null>(null)
  const [editingRequest, setEditingRequest] = useState<Partial<CollectionRequest>>({})
  const [isNewRequest, setIsNewRequest] = useState(false)
  const [newCollectionName, setNewCollectionName] = useState('')
  const [showNewCollection, setShowNewCollection] = useState(false)
  const [headers, setHeaders] = useState<Array<{ key: string; value: string }>>([])
  const [response, setResponse] = useState<{
    status: number
    statusText: string
    headers: Record<string, string>
    body: string
    duration: number
  } | null>(null)
  const [sending, setSending] = useState(false)
  const fileInputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    getCollections().then(setCollections).catch(console.error)
  }, [])

  function handleSelectRequest(req: CollectionRequest) {
    setSelectedRequest(req)
    setEditingRequest(req)
    setIsNewRequest(false)
    setHeaders(Object.entries(req.headers || {}).map(([key, value]) => ({ key, value })))
    setResponse(null)
  }

  function handleNewRequest(collectionId: string) {
    setSelectedRequest(null)
    setEditingRequest({
      name: '',
      method: 'GET',
      url: '',
      headers: {},
      body: '',
      contentType: 'application/json',
      collectionId,
    })
    setHeaders([{ key: '', value: '' }])
    setIsNewRequest(true)
    setResponse(null)
  }

  async function handleSaveRequest() {
    try {
      const headersObj: Record<string, string> = {}
      headers.forEach((h) => {
        if (h.key.trim()) headersObj[h.key.trim()] = h.value
      })

      const reqData = { ...editingRequest, headers: headersObj }

      if (isNewRequest && selectedCollection) {
        const created = await addRequest(selectedCollection.id, reqData)
        const updatedCollections = collections.map((c) =>
          c.id === selectedCollection.id
            ? { ...c, requests: [...c.requests, created] }
            : c
        )
        setCollections(updatedCollections)
        setSelectedCollection(updatedCollections.find((c) => c.id === selectedCollection.id) || null)
        setSelectedRequest(created)
        setEditingRequest(created)
        setIsNewRequest(false)
      } else if (selectedRequest && selectedCollection) {
        const updated = await updateRequest(selectedCollection.id, selectedRequest.id, reqData)
        const updatedCollections = collections.map((c) =>
          c.id === selectedCollection.id
            ? { ...c, requests: c.requests.map((r) => (r.id === updated.id ? updated : r)) }
            : c
        )
        setCollections(updatedCollections)
        setSelectedCollection(updatedCollections.find((c) => c.id === selectedCollection.id) || null)
        setSelectedRequest(updated)
        setEditingRequest(updated)
      }
    } catch (err) {
      handleError(err, '保存请求失败')
    }
  }

  async function handleDeleteRequest() {
    if (!selectedRequest || !selectedCollection) return
    try {
      await deleteRequest(selectedCollection.id, selectedRequest.id)
      const updatedCollections = collections.map((c) =>
        c.id === selectedCollection.id
          ? { ...c, requests: c.requests.filter((r) => r.id !== selectedRequest.id) }
          : c
      )
      setCollections(updatedCollections)
      setSelectedCollection(updatedCollections.find((c) => c.id === selectedCollection.id) || null)
      setSelectedRequest(null)
      setEditingRequest({})
      setIsNewRequest(false)
    } catch (err) {
      handleError(err, '删除请求失败')
    }
  }

  async function handleNewCollection() {
    if (!newCollectionName.trim()) return
    try {
      const created = await createCollection({ name: newCollectionName.trim() })
      setCollections([...collections, created])
      setSelectedCollection(created)
      setNewCollectionName('')
      setShowNewCollection(false)
    } catch (err) {
      handleError(err, '创建集合失败')
    }
  }

  async function handleDeleteCollection(id: string) {
    try {
      await deleteCollection(id)
      const filtered = collections.filter((c) => c.id !== id)
      setCollections(filtered)
      if (selectedCollection?.id === id) {
        setSelectedCollection(filtered[0] || null)
        setSelectedRequest(null)
        setEditingRequest({})
      }
    } catch (err) {
      handleError(err, '删除集合失败')
    }
  }

  async function handleSend() {
    setSending(true)
    setResponse(null)
    try {
      const headersObj: Record<string, string> = {}
      headers.forEach((h) => {
        if (h.key.trim()) headersObj[h.key.trim()] = h.value
      })
      const result = await sendRequest({
        method: editingRequest.method || 'GET',
        url: editingRequest.url || '',
        headers: headersObj,
        body: editingRequest.body || '',
      })
      setResponse(result)
    } catch (err) {
      handleError(err, '发送请求失败')
    } finally {
      setSending(false)
    }
  }

  function handleImport() {
    fileInputRef.current?.click()
  }

  async function handleFileImport(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file) return
    try {
      const imported = await importCollection(file)
      setCollections([...collections, imported])
    } catch (err) {
      handleError(err, '导入集合失败')
    }
    e.target.value = ''
  }

  function addHeaderRow() {
    setHeaders([...headers, { key: '', value: '' }])
  }

  function removeHeaderRow(index: number) {
    setHeaders(headers.filter((_, i) => i !== index))
  }

  function updateHeader(index: number, field: 'key' | 'value', value: string) {
    const updated = [...headers]
    updated[index][field] = value
    setHeaders(updated)
  }

  function formatBody(body: string) {
    try {
      return JSON.stringify(JSON.parse(body), null, 2)
    } catch {
      return body
    }
  }

  return (
    <div className="flex h-full">
      {/* 左侧集合树 */}
      <div className="w-72 border-r border-[var(--border)] flex flex-col bg-[var(--bg-inset)]">
        <div className="flex items-center gap-1 p-2 border-b border-[var(--border)]">
          <button onClick={() => setShowNewCollection(true)} className="px-2 py-1 text-xs bg-[var(--blue)] text-white rounded hover:bg-[var(--blue)]/90" aria-label="新建集合">
            新建集合
          </button>
          <button onClick={handleImport} className="px-2 py-1 text-xs bg-[var(--hover-bg)] rounded hover:bg-[var(--hover-bg)]" aria-label="导入集合">
            导入
          </button>
          <input ref={fileInputRef} type="file" accept=".json" className="hidden" onChange={handleFileImport} />
        </div>

        {showNewCollection && (
          <div className="p-2 border-b border-[var(--border)] flex gap-1">
            <input
              value={newCollectionName}
              onChange={(e) => setNewCollectionName(e.target.value)}
              className="flex-1 px-2 py-1 bg-[var(--hover-bg)] border border-[var(--border)] rounded text-xs focus:border-[var(--blue)] focus:outline-none"
              placeholder="集合名称"
              onKeyDown={(e) => e.key === 'Enter' && handleNewCollection()}
              autoFocus
              aria-label="集合名称"
            />
            <button onClick={handleNewCollection} className="px-2 py-1 text-xs bg-[var(--green)] text-white rounded" aria-label="确认创建集合">OK</button>
          </div>
        )}

        <div className="flex-1 overflow-y-auto">
          {collections.map((col) => (
            <div key={col.id}>
              <div
                onClick={() => { setSelectedCollection(col); setSelectedRequest(null); setEditingRequest({}) }}
                className={`group flex items-center gap-2 px-3 py-2 cursor-pointer border-b border-[var(--border-subtle)] ${
                  selectedCollection?.id === col.id && !selectedRequest ? 'bg-[var(--selected-bg)]' : 'hover:bg-[var(--hover-bg)]'
                }`}
              >
                <span className="text-sm">📁</span>
                <span className="flex-1 text-sm truncate">{col.name}</span>
                <span className="text-xs text-[var(--text-tertiary)]">{col.requests?.length || 0}</span>
                <button
                  onClick={(e) => { e.stopPropagation(); handleDeleteCollection(col.id) }}
                  className="text-xs text-[var(--red)] hover:text-[var(--red)]/90 opacity-0 group-hover:opacity-100"
                  aria-label={`删除集合 ${col.name}`}
                >
                  ✕
                </button>
              </div>
              {selectedCollection?.id === col.id && col.requests?.map((req) => (
                <div
                  key={req.id}
                  onClick={() => handleSelectRequest(req)}
                  className={`flex items-center gap-2 pl-6 pr-3 py-1.5 cursor-pointer border-b border-[var(--border-subtle)] ${
                    selectedRequest?.id === req.id ? 'bg-[var(--selected-bg)]' : 'hover:bg-[var(--hover-bg)]'
                  }`}
                >
                  <span className={`text-xs font-mono px-1 rounded ${
                    req.method === 'GET' ? 'bg-[var(--green)]/20 text-[var(--green)]' :
                    req.method === 'POST' ? 'bg-[var(--blue)]/20 text-[var(--blue)]' :
                    req.method === 'PUT' ? 'bg-[var(--yellow)]/20 text-[var(--yellow)]' :
                    req.method === 'DELETE' ? 'bg-[var(--red)]/20 text-[var(--red)]' :
                    'bg-[var(--text-tertiary)]/20 text-[var(--text-tertiary)]'
                  }`}>{req.method}</span>
                  <span className="text-xs truncate">{req.name || req.url}</span>
                </div>
              ))}
              {selectedCollection?.id === col.id && (
                <div
                  onClick={() => handleNewRequest(col.id)}
                  className="pl-6 pr-3 py-1.5 cursor-pointer border-b border-[var(--border-subtle)] hover:bg-[var(--hover-bg)] text-xs text-[var(--text-tertiary)]"
                >
                  + 新建请求
                </div>
              )}
            </div>
          ))}
          {collections.length === 0 && (
            <div className="p-4 text-center text-[var(--text-tertiary)] text-sm">暂无集合</div>
          )}
        </div>
      </div>

      {/* 右侧请求编辑器 */}
      <div className="flex-1 flex flex-col bg-[var(--hover-bg)] overflow-hidden">
        {selectedCollection ? (
          <>
            {/* 请求构建器 */}
            <div className="p-4 border-b border-[var(--border)] space-y-3">
              <div className="flex gap-2">
                <select
                  value={editingRequest.method || 'GET'}
                  onChange={(e) => setEditingRequest({ ...editingRequest, method: e.target.value })}
                  className="px-2 py-2 bg-[var(--bg-inset)] border border-[var(--border)] rounded text-sm font-mono focus:border-[var(--blue)] focus:outline-none"
                  aria-label="请求方法"
                >
                  {methods.map((m) => <option key={m} value={m}>{m}</option>)}
                </select>
                <input
                  value={editingRequest.url || ''}
                  onChange={(e) => setEditingRequest({ ...editingRequest, url: e.target.value })}
                  className="flex-1 px-3 py-2 bg-[var(--bg-inset)] border border-[var(--border)] rounded text-sm font-mono focus:border-[var(--blue)] focus:outline-none"
                  placeholder="输入请求 URL，支持 {{variable}} 变量"
                  aria-label="请求 URL"
                />
                <button
                  onClick={handleSend}
                  disabled={sending || !editingRequest.url}
                  className="px-4 py-2 bg-[var(--green)] text-white rounded text-sm hover:bg-[var(--green)]/90 disabled:opacity-50"
                  aria-label="发送请求"
                >
                  {sending ? '发送中...' : '发送'}
                </button>
                <button onClick={handleSaveRequest} className="px-4 py-2 bg-[var(--blue)] text-white rounded text-sm hover:bg-[var(--blue)]/90" aria-label="保存请求">
                  保存
                </button>
                {!isNewRequest && selectedRequest && (
                  <button onClick={handleDeleteRequest} className="px-3 py-2 bg-[var(--red)] text-white rounded text-sm hover:bg-[var(--red)]/90" aria-label="删除请求">
                    删除
                  </button>
                )}
              </div>

              <input
                value={editingRequest.name || ''}
                onChange={(e) => setEditingRequest({ ...editingRequest, name: e.target.value })}
                className="w-full px-3 py-2 bg-[var(--bg-inset)] border border-[var(--border)] rounded text-sm focus:border-[var(--blue)] focus:outline-none"
                placeholder="请求名称"
                aria-label="请求名称"
              />
            </div>

            {/* 标签页区域 */}
            <div className="flex-1 flex overflow-hidden">
              {/* Headers */}
              <div className="w-1/2 border-r border-[var(--border)] flex flex-col overflow-hidden">
                <div className="flex items-center justify-between px-3 py-2 border-b border-[var(--border)]">
                  <span className="text-xs font-medium text-[var(--blue)]">Headers</span>
                  <button onClick={addHeaderRow} className="text-xs text-[var(--text-tertiary)] hover:text-[var(--blue)]" aria-label="添加 Header">+ 添加</button>
                </div>
                <div className="flex-1 overflow-y-auto">
                  {headers.map((h, i) => (
                    <div key={i} className="flex items-center gap-1 px-2 py-1 border-b border-[var(--border-subtle)]">
                      <input
                        value={h.key}
                        onChange={(e) => updateHeader(i, 'key', e.target.value)}
                        className="flex-1 px-2 py-1 bg-transparent text-xs font-mono focus:outline-none"
                        placeholder="Key"
                        aria-label={`Header ${i + 1} 名称`}
                      />
                      <input
                        value={h.value}
                        onChange={(e) => updateHeader(i, 'value', e.target.value)}
                        className="flex-1 px-2 py-1 bg-transparent text-xs font-mono focus:outline-none"
                        placeholder="Value"
                        aria-label={`Header ${i + 1} 值`}
                      />
                      <button onClick={() => removeHeaderRow(i)} className="text-xs text-[var(--red)] hover:text-[var(--red)]/90 px-1" aria-label={`删除 Header ${i + 1}`}>✕</button>
                    </div>
                  ))}
                </div>

                {/* Body */}
                <div className="border-t border-[var(--border)] flex-1 flex flex-col min-h-0">
                  <div className="px-3 py-2 border-b border-[var(--border)]">
                    <span className="text-xs font-medium text-[var(--blue)]">Body</span>
                  </div>
                  <textarea
                    value={editingRequest.body || ''}
                    onChange={(e) => setEditingRequest({ ...editingRequest, body: e.target.value })}
                    className="flex-1 px-3 py-2 bg-[var(--bg-inset)] text-xs font-mono resize-none focus:outline-none"
                    placeholder='{"key": "value"}'
                    aria-label="请求 Body"
                  />
                </div>
              </div>

              {/* 响应面板 */}
              <div className="w-1/2 flex flex-col overflow-hidden">
                <div className="px-3 py-2 border-b border-[var(--border)]">
                  <span className="text-xs font-medium text-[var(--blue)]">响应</span>
                </div>
                {response ? (
                  <div className="flex-1 overflow-y-auto">
                    <div className="px-3 py-2 flex items-center gap-3 border-b border-[var(--border-subtle)]">
                      <span className={`text-sm font-mono px-2 py-0.5 rounded ${
                        response.status < 300 ? 'bg-[var(--green)]/20 text-[var(--green)]' :
                        response.status < 400 ? 'bg-[var(--yellow)]/20 text-[var(--yellow)]' :
                        'bg-[var(--red)]/20 text-[var(--red)]'
                      }`}>{response.status} {response.statusText}</span>
                      <span className="text-xs text-[var(--text-tertiary)]">{response.duration}ms</span>
                    </div>
                    <div className="px-3 py-2 border-b border-[var(--border-subtle)]">
                      <div className="text-xs text-[var(--text-tertiary)] mb-1">响应头</div>
                      {Object.entries(response.headers).map(([key, value]) => (
                        <div key={key} className="text-xs font-mono">
                          <span className="text-[var(--blue)]">{key}</span>: {value}
                        </div>
                      ))}
                    </div>
                    <pre className="px-3 py-2 text-xs font-mono whitespace-pre-wrap break-all">
                      {formatBody(response.body)}
                    </pre>
                  </div>
                ) : (
                  <div className="flex-1 flex items-center justify-center text-[var(--text-tertiary)] text-sm">
                    {sending ? '请求发送中...' : '点击"发送"查看响应'}
                  </div>
                )}
              </div>
            </div>
          </>
        ) : (
          <div className="flex-1 flex items-center justify-center text-[var(--text-tertiary)] text-sm">
            选择或创建一个集合开始
          </div>
        )}
      </div>
    </div>
  )
}
