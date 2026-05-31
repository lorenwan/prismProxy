import { useState, useEffect } from 'react'
import type { Settings } from '../types'
import { getSettings, updateSettings, downloadCaCert } from '../services/ai'
import { startProxy, stopProxy, getProxyStatus, enableSystemProxy, disableSystemProxy, getSystemProxyStatus } from '../services/proxy'

export default function SettingsPage() {
  const [settings, setSettings] = useState<Settings>(defaultSettings)
  const [saved, setSaved] = useState(false)
  const [proxyRunning, setProxyRunning] = useState(false)
  const [proxyAddr, setProxyAddr] = useState('')
  const [proxyLoading, setProxyLoading] = useState(false)
  const [systemProxyEnabled, setSystemProxyEnabled] = useState(false)
  const [systemProxyLoading, setSystemProxyLoading] = useState(false)

  useEffect(() => {
    getSettings().then(setSettings).catch(console.error)
    refreshProxyStatus()
    refreshSystemProxyStatus()
  }, [])

  // 刷新代理状态
  async function refreshProxyStatus() {
    try {
      const status = await getProxyStatus()
      setProxyRunning(status.running)
      setProxyAddr(status.addr || '')
    } catch (err) {
      console.error('获取代理状态失败:', err)
    }
  }

  // 刷新系统代理状态
  async function refreshSystemProxyStatus() {
    try {
      const status = await getSystemProxyStatus()
      setSystemProxyEnabled(status.enabled)
    } catch (err) {
      console.error('获取系统代理状态失败:', err)
    }
  }

  // 启动/停止代理
  async function toggleProxy() {
    setProxyLoading(true)
    try {
      if (proxyRunning) {
        await stopProxy()
      } else {
        await startProxy()
      }
      await refreshProxyStatus()
    } catch (err) {
      console.error('切换代理状态失败:', err)
    } finally {
      setProxyLoading(false)
    }
  }

  // 启用/禁用系统代理
  async function toggleSystemProxy() {
    setSystemProxyLoading(true)
    try {
      if (systemProxyEnabled) {
        await disableSystemProxy()
      } else {
        await enableSystemProxy()
      }
      await refreshSystemProxyStatus()
    } catch (err) {
      console.error('切换系统代理失败:', err)
    } finally {
      setSystemProxyLoading(false)
    }
  }

  // 保存设置
  async function handleSave() {
    await updateSettings(settings)
    setSaved(true)
    setTimeout(() => setSaved(false), 2000)
  }

  // 更新代理设置
  function updateProxy(key: keyof Settings['proxy'], value: any) {
    setSettings({ ...settings, proxy: { ...settings.proxy, [key]: value } })
  }

  // 更新 AI 设置
  function updateAi(key: keyof Settings['ai'], value: string) {
    setSettings({ ...settings, ai: { ...settings.ai, [key]: value } })
  }

  return (
    <div className="h-full overflow-y-auto bg-[#1a1b26] p-6">
      <div className="max-w-2xl space-y-8">
        {/* 代理控制 */}
        <section className="space-y-4">
          <h2 className="text-lg font-semibold text-[#7aa2f7]">代理控制</h2>
          <div className="bg-[#24283b] rounded-lg p-4 space-y-4">
            {/* 代理开关 */}
            <div className="flex items-center justify-between">
              <div>
                <div className="text-sm font-medium">HTTP 代理服务</div>
                <div className="text-xs text-[#565f89]">
                  {proxyRunning 
                    ? `运行中 - ${proxyAddr}` 
                    : '已停止 - 点击启动代理服务'}
                </div>
              </div>
              <button
                onClick={toggleProxy}
                disabled={proxyLoading}
                className={`relative w-14 h-7 rounded-full transition-colors ${
                  proxyRunning ? 'bg-[#9ece6a]' : 'bg-[#3b4261]'
                } ${proxyLoading ? 'opacity-50' : ''}`}
              >
                <div
                  className={`absolute top-0.5 w-6 h-6 rounded-full bg-white transition-all ${
                    proxyRunning ? 'left-7' : 'left-0.5'
                  }`}
                />
                {proxyLoading && (
                  <div className="absolute inset-0 flex items-center justify-center">
                    <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                  </div>
                )}
              </button>
            </div>

            {/* 系统代理开关 */}
            <div className="flex items-center justify-between">
              <div>
                <div className="text-sm font-medium">系统代理</div>
                <div className="text-xs text-[#565f89]">
                  {systemProxyEnabled 
                    ? '已启用 - 所有流量经过 PrismProxy' 
                    : '已禁用 - 仅手动代理'}
                </div>
              </div>
              <button
                onClick={toggleSystemProxy}
                disabled={systemProxyLoading || !proxyRunning}
                className={`relative w-14 h-7 rounded-full transition-colors ${
                  systemProxyEnabled ? 'bg-[#9ece6a]' : 'bg-[#3b4261]'
                } ${systemProxyLoading || !proxyRunning ? 'opacity-50' : ''}`}
              >
                <div
                  className={`absolute top-0.5 w-6 h-6 rounded-full bg-white transition-all ${
                    systemProxyEnabled ? 'left-7' : 'left-0.5'
                  }`}
                />
                {systemProxyLoading && (
                  <div className="absolute inset-0 flex items-center justify-center">
                    <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                  </div>
                )}
              </button>
            </div>

            {/* 状态指示 */}
            <div className="flex items-center gap-2">
              <div className={`w-2 h-2 rounded-full ${proxyRunning ? 'bg-[#9ece6a]' : 'bg-[#f7768e]'}`} />
              <span className="text-xs text-[#565f89]">
                {proxyRunning ? '代理已就绪' : '代理未启动'}
              </span>
              {systemProxyEnabled && (
                <>
                  <span className="text-xs text-[#565f89]">•</span>
                  <div className="w-2 h-2 rounded-full bg-[#7aa2f7]" />
                  <span className="text-xs text-[#565f89]">系统代理已启用</span>
                </>
              )}
            </div>

            {/* 使用说明 */}
            {proxyRunning && !systemProxyEnabled && (
              <div className="text-xs text-[#565f89] bg-[#1a1b26] p-3 rounded">
                <p className="font-medium mb-1">使用方法：</p>
                <p>1. 设置浏览器代理为 <code className="text-[#7aa2f7]">{proxyAddr}</code></p>
                <p>2. 或启用系统代理（上方开关）</p>
                <p>3. 开始抓包调试</p>
              </div>
            )}

            {systemProxyEnabled && (
              <div className="text-xs text-[#565f89] bg-[#1a1b26] p-3 rounded">
                <p className="font-medium mb-1">系统代理已启用：</p>
                <p>• 所有 HTTP/HTTPS 流量将经过 PrismProxy</p>
                <p>• 无需手动配置浏览器代理</p>
                <p>• 关闭 PrismProxy 时会自动恢复原设置</p>
              </div>
            )}
          </div>
        </section>

        {/* 代理配置 */}
        <section className="space-y-4">
          <h2 className="text-lg font-semibold text-[#7aa2f7]">代理配置</h2>
          <div className="bg-[#24283b] rounded-lg p-4 space-y-4">
            {/* 端口 */}
            <div>
              <label className="block text-sm text-[#565f89] mb-1">代理端口</label>
              <input
                type="number"
                value={settings.proxy.port}
                onChange={(e) => updateProxy('port', Number(e.target.value))}
                className="w-40 px-3 py-2 bg-[#1a1b26] border border-[#3b4261] rounded text-sm focus:border-[#7aa2f7] focus:outline-none"
              />
            </div>

            {/* MITM 开关 */}
            <div className="flex items-center justify-between">
              <div>
                <div className="text-sm">MITM (中间人攻击)</div>
                <div className="text-xs text-[#565f89]">启用后可查看 HTTPS 请求内容</div>
              </div>
              <button
                onClick={() => updateProxy('mitmEnabled', !settings.proxy.mitmEnabled)}
                className={`w-10 h-5 rounded-full transition-colors ${
                  settings.proxy.mitmEnabled ? 'bg-[#9ece6a]' : 'bg-[#3b4261]'
                }`}
              >
                <div
                  className={`w-4 h-4 rounded-full bg-white transition-transform ${
                    settings.proxy.mitmEnabled ? 'translate-x-5' : 'translate-x-0.5'
                  }`}
                />
              </button>
            </div>

            {/* 证书下载 */}
            <div>
              <button
                onClick={downloadCaCert}
                className="px-4 py-2 bg-[#283457] text-[#7aa2f7] rounded text-sm hover:bg-[#3b4261]"
              >
                下载 CA 证书
              </button>
              <span className="ml-2 text-xs text-[#565f89]">安装证书后可解密 HTTPS 流量</span>
            </div>
          </div>
        </section>

        {/* AI 配置 */}
        <section className="space-y-4">
          <h2 className="text-lg font-semibold text-[#7aa2f7]">AI 配置</h2>
          <div className="bg-[#24283b] rounded-lg p-4 space-y-4">
            {/* Provider */}
            <div>
              <label className="block text-sm text-[#565f89] mb-1">AI Provider</label>
              <select
                value={settings.ai.provider}
                onChange={(e) => updateAi('provider', e.target.value)}
                className="w-full px-3 py-2 bg-[#1a1b26] border border-[#3b4261] rounded text-sm focus:border-[#7aa2f7] focus:outline-none"
              >
                <option value="openai">OpenAI</option>
                <option value="claude">Claude</option>
                <option value="ollama">Ollama</option>
              </select>
            </div>

            {/* API Key */}
            <div>
              <label className="block text-sm text-[#565f89] mb-1">API Key</label>
              <input
                type="password"
                value={settings.ai.apiKey}
                onChange={(e) => updateAi('apiKey', e.target.value)}
                className="w-full px-3 py-2 bg-[#1a1b26] border border-[#3b4261] rounded text-sm focus:border-[#7aa2f7] focus:outline-none"
                placeholder="sk-..."
              />
            </div>

            {/* Base URL */}
            <div>
              <label className="block text-sm text-[#565f89] mb-1">Base URL</label>
              <input
                value={settings.ai.baseUrl}
                onChange={(e) => updateAi('baseUrl', e.target.value)}
                className="w-full px-3 py-2 bg-[#1a1b26] border border-[#3b4261] rounded text-sm focus:border-[#7aa2f7] focus:outline-none"
                placeholder="https://api.openai.com/v1"
              />
            </div>

            {/* Model */}
            <div>
              <label className="block text-sm text-[#565f89] mb-1">模型</label>
              <input
                value={settings.ai.model}
                onChange={(e) => updateAi('model', e.target.value)}
                className="w-full px-3 py-2 bg-[#1a1b26] border border-[#3b4261] rounded text-sm focus:border-[#7aa2f7] focus:outline-none"
                placeholder="gpt-4o"
              />
            </div>
          </div>
        </section>

        {/* 保存按钮 */}
        <div className="flex items-center gap-3">
          <button onClick={handleSave} className="px-6 py-2 bg-[#7aa2f7] text-[#1a1b26] rounded text-sm hover:bg-[#89b4fa]">
            保存设置
          </button>
          {saved && <span className="text-sm text-[#9ece6a]">保存成功</span>}
        </div>
      </div>
    </div>
  )
}

// 默认设置
const defaultSettings: Settings = {
  proxy: { port: 8081, mitmEnabled: false, caCertPath: '' },
  ai: { provider: 'openai', apiKey: '', baseUrl: '', model: '' },
}
