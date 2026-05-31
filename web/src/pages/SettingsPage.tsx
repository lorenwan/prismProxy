import { useState, useEffect } from 'react'
import type { Settings } from '../types'
import { getSettings, updateSettings, downloadCaCert } from '../services/ai'

// 默认设置
const defaultSettings: Settings = {
  proxy: { port: 8081, mitmEnabled: false, caCertPath: '' },
  ai: { provider: 'openai', apiKey: '', baseUrl: '', model: '' },
}

export default function SettingsPage() {
  const [settings, setSettings] = useState<Settings>(defaultSettings)
  const [saved, setSaved] = useState(false)

  useEffect(() => {
    getSettings().then(setSettings).catch(console.error)
  }, [])

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
