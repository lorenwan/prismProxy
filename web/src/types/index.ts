// 流量记录
export interface Transaction {
  id: string
  method: string
  host: string
  path: string
  url: string
  status: number
  statusCode: number
  contentType: string
  size: number
  duration: number
  requestTime: string
  source: string
  bookmarked: boolean
  notes: string
  color: string
  tags: string[]
  request: RequestData
  response: ResponseData
}

// 请求数据
export interface RequestData {
  method: string
  url: string
  headers: Record<string, string>
  body: string
  contentType: string
  size: number
}

// 响应数据
export interface ResponseData {
  status: number
  statusText: string
  headers: Record<string, string>
  body: string
  contentType: string
  size: number
}

// 流量统计
export interface TrafficStats {
  totalRequests: number
  successRequests: number
  failedRequests: number
  totalSize: number
  avgDuration: number
}

// WebSocket 消息
export interface WsMessage {
  type: string
  payload: any
  time: string
}

// 规则
export interface Rule {
  id: string
  name: string
  enabled: boolean
  priority: number
  matchType: 'host' | 'path' | 'method' | 'header' | 'body' | 'url'
  matchValue: string
  actionType: 'block' | 'redirect' | 'modify_request' | 'modify_response' | 'delay' | 'mock'
  actionValue: string
  hitCount: number
  createdAt: string
  updatedAt: string
}

// 断点
export interface Breakpoint {
  id: string
  name: string
  enabled: boolean
  phase: 'request' | 'response'
  matchType: 'host' | 'path' | 'url'
  matchValue: string
  hitCount: number
  createdAt: string
  updatedAt: string
}

// 断点会话
export interface BreakpointSession {
  id: string
  breakpointId: string
  transactionId: string
  phase: 'request' | 'response'
  status: 'paused' | 'resumed' | 'modified'
  originalData: RequestData | ResponseData
  modifiedData?: RequestData | ResponseData
  createdAt: string
}

// AI 聊天消息
export interface ChatMessage {
  role: 'user' | 'assistant' | 'system'
  content: string
  timestamp: string
}

// AI 聊天请求
export interface ChatRequest {
  messages: ChatMessage[]
  provider?: string
  stream?: boolean
}

// 代理设置
export interface ProxySettings {
  port: number
  mitmEnabled: boolean
  caCertPath: string
}

// AI 设置
export interface AiSettings {
  provider: 'openai' | 'claude' | 'ollama'
  apiKey: string
  baseUrl: string
  model: string
}

// 系统设置
export interface Settings {
  proxy: ProxySettings
  ai: AiSettings
}