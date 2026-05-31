import { useState, useEffect } from 'react'
import { getEnvironments, createEnvironment, updateEnvironment, deleteEnvironment, activateEnvironment } from '../services/environments'
import type { Environment, EnvironmentVariable } from '../types'

export default function EnvironmentsPage() {
  const [environments, setEnvironments] = useState<Environment[]>([])
  const [selected, setSelected] = useState<Environment | null>(null)
  const [editingName, setEditingName] = useState('')
  const [variables, setVariables] = useState<EnvironmentVariable[]>([])
  const [isNew, setIsNew] = useState(false)
  const [deleteConfirm, setDeleteConfirm] = useState<Environment | null>(null)

  useEffect(() => {
    getEnvironments().then(setEnvironments).catch(console.error)
  }, [])

  function handleSelect(env: Environment) {
    setSelected(env)
    setEditingName(env.name)
    setVariables(env.variables || [])
    setIsNew(false)
  }

  function handleNew() {
    setSelected(null)
    setEditingName('')
    setVariables([{ key: '', value: '', enabled: true }])
    setIsNew(true)
  }

  async function handleSave() {
    const validVars = variables.filter((v) => v.key.trim())
    const envData = { name: editingName, variables: validVars }

    if (isNew) {
      const created = await createEnvironment(envData)
      setEnvironments([...environments, created])
      setSelected(created)
      setIsNew(false)
    } else if (selected) {
      const updated = await updateEnvironment(selected.id, envData)
      setEnvironments(environments.map((e) => (e.id === selected.id ? updated : e)))
      setSelected(updated)
    }
  }

  async function handleDelete(env: Environment) {
    await deleteEnvironment(env.id)
    setEnvironments(environments.filter((e) => e.id !== env.id))
    if (selected?.id === env.id) {
      setSelected(null)
      setEditingName('')
      setVariables([])
      setIsNew(false)
    }
    setDeleteConfirm(null)
  }

  async function handleActivate(env: Environment) {
    await activateEnvironment(env.id)
    setEnvironments(environments.map((e) => ({ ...e, active: e.id === env.id })))
    if (selected?.id === env.id) {
      setSelected({ ...selected, active: true })
    }
  }

  function addVariable() {
    setVariables([...variables, { key: '', value: '', enabled: true }])
  }

  function removeVariable(index: number) {
    setVariables(variables.filter((_, i) => i !== index))
  }

  function updateVariable(index: number, field: keyof EnvironmentVariable, value: string | boolean) {
    const updated = [...variables]
    updated[index] = { ...updated[index], [field]: value }
    setVariables(updated)
  }

  return (
    <div className="flex h-full">
      {/* 左侧环境列表 */}
      <div className="w-72 border-r border-[#3b4261] flex flex-col bg-[#1a1b26]">
        <div className="flex items-center gap-1 p-2 border-b border-[#3b4261]">
          <button onClick={handleNew} className="px-2 py-1 text-xs bg-[#7aa2f7] text-[#1a1b26] rounded hover:bg-[#89b4fa]">
            新增环境
          </button>
        </div>

        <div className="flex-1 overflow-y-auto">
          {environments.map((env) => (
            <div
              key={env.id}
              onClick={() => handleSelect(env)}
              className={`flex items-center gap-2 px-3 py-2 cursor-pointer border-b border-[#24283b] ${
                selected?.id === env.id ? 'bg-[#283457]' : 'hover:bg-[#24283b]'
              }`}
            >
              <span className="text-sm">{env.active ? '✅' : '📦'}</span>
              <div className="flex-1 min-w-0">
                <div className="text-sm truncate">{env.name}</div>
                <div className="text-xs text-[#565f89]">{env.variables?.length || 0} 个变量</div>
              </div>
              {!env.active && (
                <button
                  onClick={(e) => { e.stopPropagation(); handleActivate(env) }}
                  className="px-1.5 py-0.5 text-xs bg-[#24283b] rounded hover:bg-[#3b4261]"
                >
                  激活
                </button>
              )}
            </div>
          ))}
          {environments.length === 0 && (
            <div className="p-4 text-center text-[#565f89] text-sm">暂无环境</div>
          )}
        </div>
      </div>

      {/* 右侧变量编辑器 */}
      <div className="flex-1 flex flex-col bg-[#24283b] overflow-y-auto">
        {selected || isNew ? (
          <div className="p-4 space-y-4 max-w-3xl">
            <h2 className="text-lg font-semibold">{isNew ? '新增环境' : '编辑环境'}</h2>

            <div>
              <label className="block text-sm text-[#565f89] mb-1">环境名称</label>
              <input
                value={editingName}
                onChange={(e) => setEditingName(e.target.value)}
                className="w-full px-3 py-2 bg-[#1a1b26] border border-[#3b4261] rounded text-sm focus:border-[#7aa2f7] focus:outline-none"
                placeholder="如：开发环境、测试环境、生产环境"
              />
            </div>

            {/* 变量表格 */}
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <h3 className="text-sm font-medium text-[#7aa2f7]">环境变量</h3>
                <button onClick={addVariable} className="text-xs text-[#565f89] hover:text-[#7aa2f7]">+ 添加变量</button>
              </div>

              <div className="border border-[#3b4261] rounded overflow-hidden">
                {/* 表头 */}
                <div className="flex items-center gap-2 px-3 py-2 bg-[#1a1b26] border-b border-[#3b4261] text-xs text-[#565f89]">
                  <div className="w-8">启用</div>
                  <div className="flex-1">Key</div>
                  <div className="flex-1">Value</div>
                  <div className="w-8" />
                </div>

                {/* 变量行 */}
                {variables.map((v, i) => (
                  <div key={i} className="flex items-center gap-2 px-3 py-1.5 border-b border-[#24283b]">
                    <input
                      type="checkbox"
                      checked={v.enabled}
                      onChange={(e) => updateVariable(i, 'enabled', e.target.checked)}
                      className="w-4 h-4 accent-[#7aa2f7]"
                    />
                    <input
                      value={v.key}
                      onChange={(e) => updateVariable(i, 'key', e.target.value)}
                      className="flex-1 px-2 py-1 bg-transparent text-sm font-mono focus:outline-none"
                      placeholder="variable_name"
                    />
                    <input
                      value={v.value}
                      onChange={(e) => updateVariable(i, 'value', e.target.value)}
                      className="flex-1 px-2 py-1 bg-transparent text-sm font-mono focus:outline-none"
                      placeholder="value"
                    />
                    <button onClick={() => removeVariable(i)} className="text-xs text-[#f7768e] hover:text-[#ff9eaf] px-1">✕</button>
                  </div>
                ))}

                {variables.length === 0 && (
                  <div className="px-3 py-4 text-center text-[#565f89] text-sm">点击"添加变量"开始</div>
                )}
              </div>
            </div>

            {/* 变量引用说明 */}
            <div className="bg-[#1a1b26] border border-[#3b4261] rounded p-3">
              <h4 className="text-sm font-medium text-[#e0af68] mb-2">变量引用说明</h4>
              <div className="text-xs text-[#565f89] space-y-1">
                <p>在请求 URL、Headers、Body 中使用 <code className="px-1 py-0.5 bg-[#24283b] rounded text-[#7aa2f7]">{'{{variable_name}}'}</code> 引用变量。</p>
                <p>例如：<code className="px-1 py-0.5 bg-[#24283b] rounded text-[#7aa2f7]">{'{{base_url}}'}</code>/api/users</p>
                <p>激活环境后，所有请求中的变量会自动替换为对应的值。</p>
              </div>
            </div>

            {/* 操作按钮 */}
            <div className="flex gap-2 pt-2">
              <button onClick={handleSave} className="px-4 py-2 bg-[#7aa2f7] text-[#1a1b26] rounded text-sm hover:bg-[#89b4fa]">
                保存
              </button>
              {!isNew && selected && (
                <>
                  <button onClick={() => handleActivate(selected)} className="px-4 py-2 bg-[#9ece6a] text-[#1a1b26] rounded text-sm hover:bg-[#a9d882]">
                    激活
                  </button>
                  <button onClick={() => setDeleteConfirm(selected)} className="px-4 py-2 bg-[#f7768e] text-[#1a1b26] rounded text-sm hover:bg-[#ff9eaf]">
                    删除
                  </button>
                </>
              )}
            </div>
          </div>
        ) : (
          <div className="flex-1 flex items-center justify-center text-[#565f89] text-sm">
            选择或创建一个环境开始
          </div>
        )}
      </div>

      {/* 删除确认 */}
      {deleteConfirm && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50" onClick={() => setDeleteConfirm(null)}>
          <div className="bg-[#24283b] border border-[#3b4261] rounded-lg p-4 w-80" onClick={(e) => e.stopPropagation()}>
            <h3 className="text-base font-semibold mb-2">确认删除</h3>
            <p className="text-sm text-[#565f89] mb-4">确定要删除环境 "{deleteConfirm.name}" 吗？此操作不可撤销。</p>
            <div className="flex gap-2 justify-end">
              <button onClick={() => setDeleteConfirm(null)} className="px-3 py-1.5 text-sm bg-[#1a1b26] rounded hover:bg-[#3b4261]">取消</button>
              <button onClick={() => handleDelete(deleteConfirm)} className="px-3 py-1.5 text-sm bg-[#f7768e] text-[#1a1b26] rounded hover:bg-[#ff9eaf]">删除</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
