import { invoke } from '@tauri-apps/api/core'

// 认证数据
export interface AuthData {
  type: 'none' | 'basic' | 'bearer' | 'api_key'
  username?: string
  password?: string
  token?: string
  apiKey?: string
  apiValue?: string
  location?: 'header' | 'query'
}

// 代码生成请求（对应 Proto CodeGenRequest）
export interface CodeGenRequest {
  method: string
  url: string
  headers?: Record<string, string>
  queryParams?: Record<string, string>
  body?: string
  bodyType?: 'none' | 'raw' | 'json' | 'xml' | 'form'
  auth?: AuthData
  language: string
}

// 代码生成结果（对应 Proto CodeGenResult）
export interface CodeGenResult {
  language: string
  code: string
  filename: string
}

// 语言信息（对应 Proto LanguageInfo）
export interface LanguageInfo {
  id: string
  name: string
  description: string
}

// 生成代码
export async function generateCode(request: CodeGenRequest): Promise<CodeGenResult> {
  const result = await invoke<string>('generate_code', {
    request: JSON.stringify(request),
  })
  return JSON.parse(result)
}

// 获取支持的语言列表
export async function listCodegenLanguages(): Promise<LanguageInfo[]> {
  const result = await invoke<string>('list_codegen_languages')
  const parsed = JSON.parse(result)
  return parsed?.languages ?? []
}
