// 流量记录（对应 Proto TrafficEntry）
export interface Transaction {
  id: string
  timestamp: string
  duration_ms: number
  method: string
  url: string
  host: string
  path: string
  scheme: string
  port: number
  request: RequestData
  response: ResponseData
  client_addr: string
  server_ip: string
  bookmarked: boolean
  color: string
  notes: string
  tags: string[]
}

// 请求数据（对应 Proto RequestData）
export interface RequestData {
  headers: Record<string, { values: string[] }>
  body: string
  body_size: number
  content_type: string
}

// 响应数据（对应 Proto ResponseData）
export interface ResponseData {
  status_code: number
  status_text: string
  headers: Record<string, { values: string[] }>
  body: string
  body_size: number
  content_type: string
}

// 流量统计（对应 Proto TrafficStats）
export interface TrafficStats {
  total_requests: number
  total_responses: number
  avg_duration_ms: number
  max_duration_ms: number
  min_duration_ms: number
  error_count: number
  success_count: number
  host_stats: Array<{ host: string; count: number; avg_time_ms: number }>
  method_stats: Array<{ method: string; count: number }>
  status_stats: Array<{ status_code: number; count: number }>
}

// WebSocket 消息
export interface WsMessage {
  type: string
  payload: any
  time: string
}

// 规则匹配条件 (与 rules.proto RuleMatch 对应)
export interface RuleMatch {
  url_pattern?: string       // URL 正则匹配
  url_wildcard?: string      // URL 通配符匹配
  host_pattern?: string      // 主机名匹配
  methods?: string[]        // HTTP 方法过滤
  header_match?: {
    name: string
    value: string
    match_type: string       // exact / regex / contains
  }
  content_type?: string[]    // Content-Type 过滤
}

// 拦截响应规范 (与 rules.proto BlockSpec 对应)
export interface BlockSpec {
  status_code?: number
  headers?: Record<string, string>
  body?: string
}

// 修改规范 (与 rules.proto ModifySpec 对应)
export interface ModifySpec {
  add_headers?: Record<string, string>
  remove_headers?: string[]
  set_headers?: Record<string, string>
  add_query?: Record<string, string>
  remove_query?: string[]
  set_query?: Record<string, string>
  body_replace?: string
}

// 规则动作 (与 rules.proto RuleAction 对应)
export interface RuleAction {
  type: 'block' | 'redirect' | 'modify' | 'delay'
  local_path?: string       // 本地文件路径 (redirect)
  remote_url?: string       // 远程 URL (redirect)
  modify?: ModifySpec       // 修改规范
  block_response?: BlockSpec // 拦截响应规范
  delay_ms?: number         // 延迟毫秒数
}

// 规则 (与 rules.proto Rule 对应)
export interface Rule {
  id: string
  name: string
  enabled: boolean
  priority: number
  match: RuleMatch
  action: RuleAction
  created_at: string
  updated_at: string
}

// 规则统计 (与 rules.proto RuleStats 对应)
export interface RuleStats {
  total_rules: number
  enabled_rules: number
  disabled_rules: number
  hit_counts: Record<string, number>
}

// 断点动作类型
export type BreakActionType = 'pause' | 'auto_modify' | 'drop'

// 断点动作
export interface BreakAction {
  type: BreakActionType
  modifications?: ModifySpec
}

// 流量记录 (用于断点会话)
export interface TrafficEntry {
  id: string
  method: string
  host: string
  path: string
  url: string
  statusCode: number
  requestHeaders: Record<string, string>
  requestBody: string
  responseHeaders: Record<string, string>
  responseBody: string
  requestTime: string
  responseTime?: string
  duration?: number
  source?: string
}

// 断点 (与 breakpoints.proto Breakpoint 对应)
export interface Breakpoint {
  id: string
  name: string
  enabled: boolean
  phase: 'request' | 'response'
  match: RuleMatch
  action: BreakAction
  hitCount: number
  createdAt: string
  updatedAt: string
}

// 断点会话状态
export type SessionStatus = 'waiting' | 'resolved' | 'dropped' | 'timeout'

// 断点会话 (与 breakpoints.proto BreakpointSession 对应)
export interface BreakpointSession {
  id: string
  breakpointId: string
  transactionId: number
  phase: 'request' | 'response'
  status: SessionStatus
  original?: TrafficEntry
  modified?: TrafficEntry
  createdAt: string
  resolvedAt?: string
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

// 重写规则
export interface RewriteRule {
  id: string
  name: string
  enabled: boolean
  type: 'add_header' | 'remove_header' | 'replace_header' | 'replace_body' | 'replace_url' | 'map_local' | 'map_remote'
  matchType: 'host' | 'path' | 'url' | 'method'
  matchValue: string
  actionKey: string
  actionValue: string
  priority: number
  createdAt: string
  updatedAt: string
}

// API 集合请求
export interface CollectionRequest {
  id: string
  name: string
  method: string
  url: string
  headers: Record<string, string>
  body: string
  contentType: string
  collectionId: string
  createdAt: string
  updatedAt: string
}

// API 集合
export interface Collection {
  id: string
  name: string
  description: string
  requests: CollectionRequest[]
  createdAt: string
  updatedAt: string
}

// 环境变量
export interface EnvironmentVariable {
  key: string
  value: string
  enabled: boolean
}

// 环境
export interface Environment {
  id: string
  name: string
  active: boolean
  variables: EnvironmentVariable[]
  createdAt: string
  updatedAt: string
}