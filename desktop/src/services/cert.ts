import api from './api'

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
  return api.get('/cert/ca') as any
}

// 下载 CA 证书
export function downloadCaCert() {
  window.open('/api/cert/ca/download', '_blank')
}

// 重新生成 CA 证书
export async function regenerateCa(): Promise<CaInfo> {
  return api.post('/cert/ca/regenerate') as any
}

// 获取已签发的域名证书列表
export async function getIssuedCerts(): Promise<CertInfo[]> {
  return api.get('/cert/issued') as any
}

// 清除域名证书缓存
export async function clearCertCache(): Promise<void> {
  return api.delete('/cert/cache') as any
}

// 获取证书信任状态
export async function getTrustStatus(): Promise<{ trusted: boolean; platform: string }> {
  return api.get('/cert/trust-status') as any
}
