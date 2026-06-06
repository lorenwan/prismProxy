import { invoke } from '@tauri-apps/api/core'

export interface CaInfo {
  subject: string
  serial_number: string
  not_before: string
  not_after: string
  fingerprint: string
  is_loaded: boolean
}

export interface CertCheckResult {
  domain: string
  exists: boolean
  is_valid: boolean
  is_expired: boolean
  fingerprint: string
}

export interface CertInfo {
  domain: string
  serial_number: string
  not_before: string
  not_after: string
  issuer: string
  is_valid: boolean
}

// 获取 CA 证书信息
export async function getCaInfo(): Promise<CaInfo> {
  const result = await invoke<string>('get_ca_info')
  return JSON.parse(result)
}

// 下载 CA 证书（TODO: 需要通过 Tauri 文件对话框实现）
export function downloadCaCert() {
  console.warn('downloadCaCert: 暂未实现 Tauri 文件保存')
}

// 重新生成 CA 证书
export async function regenerateCa(): Promise<CaInfo> {
  const result = await invoke<string>('generate_ca')
  return JSON.parse(result)
}

// 导出 CA 证书（返回 cert PEM 和 key PEM）
export async function exportCa(): Promise<{ cert_pem: string; key_pem: string }> {
  const result = await invoke<string>('export_ca')
  return JSON.parse(result)
}

// 签发域名证书
export async function issueCert(domain: string): Promise<CertInfo> {
  const result = await invoke<string>('issue_cert', { domain })
  return JSON.parse(result)
}

// 获取已签发的域名证书列表
export async function getIssuedCerts(): Promise<CertInfo[]> {
  const result = await invoke<string>('list_certs')
  return JSON.parse(result)
}

// 删除域名证书
export async function deleteCert(domain: string): Promise<void> {
  await invoke('delete_cert', { domain })
}

// 检查证书状态
export async function checkCert(domain: string): Promise<CertCheckResult> {
  const result = await invoke<string>('check_cert', { domain })
  return JSON.parse(result)
}

// 清除域名证书缓存
export async function clearCertCache(): Promise<void> {
  await invoke('clear_certs')
}

// 获取证书信任状态
export async function getTrustStatus(): Promise<{ trusted: boolean; platform: string; ca_fingerprint?: string }> {
  const result = await invoke<string>('get_trust_status')
  return JSON.parse(result)
}
