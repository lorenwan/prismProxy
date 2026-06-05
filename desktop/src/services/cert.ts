import { invoke } from '@tauri-apps/api/core'

export interface CaInfo {
  subject: string
  issuer: string
  serialNumber: string
  notBefore: string
  notAfter: string
  fingerprint: string
  isInstalled: boolean
}

export interface CertInfo {
  host: string
  subject: string
  issuer: string
  serialNumber: string
  notBefore: string
  notAfter: string
  fingerprint: string
  isValid: boolean
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

// 导出 CA 证书
export async function exportCa(): Promise<string> {
  const result = await invoke<string>('export_ca')
  return result
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
export async function checkCert(domain: string): Promise<CertInfo> {
  const result = await invoke<string>('check_cert', { domain })
  return JSON.parse(result)
}

// 清除域名证书缓存
export async function clearCertCache(): Promise<void> {
  await invoke('clear_certs')
}

// 获取证书信任状态（TODO: 暂未实现对应的 Rust IPC 命令）
export async function getTrustStatus(): Promise<{ trusted: boolean; platform: string }> {
  return { trusted: false, platform: 'unknown' }
}
