import { NavLink } from 'react-router-dom'
import {
  Activity,
  FileText,
  PauseCircle,
  Bot,
  Settings,
  ChevronLeft,
  ChevronRight,
  Repeat,
  FolderOpen,
  Globe,
  Code,
  GitCompare,
  BarChart3,
} from 'lucide-react'
import { useState } from 'react'
import { clsx } from 'clsx'

const navItems = [
  { path: '/', label: '流量', icon: Activity },
  { path: '/rules', label: '规则', icon: FileText },
  { path: '/rewrite', label: '重写', icon: Repeat },
  { path: '/breakpoints', label: '断点', icon: PauseCircle },
  { path: '/collections', label: '集合', icon: FolderOpen },
  { path: '/environments', label: '环境', icon: Globe },
  { path: '/scripts', label: '脚本', icon: Code },
  { path: '/diff', label: 'Diff', icon: GitCompare },
  { path: '/performance', label: '性能', icon: BarChart3 },
  { path: '/ai', label: 'AI', icon: Bot },
  { path: '/settings', label: '设置', icon: Settings },
]

export default function Sidebar() {
  const [collapsed, setCollapsed] = useState(false)

  return (
    <aside
      className={clsx(
        'flex flex-col bg-[#161b22] border-r border-[#30363d] transition-all duration-200',
        collapsed ? 'w-[52px]' : 'w-[180px]'
      )}
    >
      {/* Logo */}
      <div className="flex items-center h-12 px-3 border-b border-[#30363d]">
        <div className="w-7 h-7 rounded-md bg-[#58a6ff] flex items-center justify-center shrink-0">
          <span className="text-white font-bold text-sm">P</span>
        </div>
        {!collapsed && (
          <span className="ml-2.5 text-sm font-semibold text-[#e6edf3] whitespace-nowrap overflow-hidden">
            PrismProxy
          </span>
        )}
      </div>

      {/* 导航菜单 */}
      <nav className="flex-1 flex flex-col gap-0.5 p-2 overflow-y-auto">
        {navItems.map((item) => {
          const Icon = item.icon
          return (
            <NavLink
              key={item.path}
              to={item.path}
              end={item.path === '/'}
              className={({ isActive }) =>
                clsx(
                  'flex items-center gap-2.5 rounded-md transition-colors group',
                  collapsed ? 'justify-center h-9 px-0' : 'h-8 px-2.5',
                  isActive
                    ? 'bg-[#58a6ff]/15 text-[#58a6ff]'
                    : 'text-[#8b949e] hover:bg-[#21262d] hover:text-[#e6edf3]'
                )
              }
              title={collapsed ? item.label : undefined}
            >
              <Icon size={18} className="shrink-0" />
              {!collapsed && (
                <span className="text-sm whitespace-nowrap overflow-hidden">{item.label}</span>
              )}
            </NavLink>
          )
        })}
      </nav>

      {/* 折叠按钮 */}
      <button
        onClick={() => setCollapsed(!collapsed)}
        className="flex items-center justify-center h-9 border-t border-[#30363d] text-[#8b949e] hover:text-[#e6edf3] hover:bg-[#21262d] transition-colors"
      >
        {collapsed ? <ChevronRight size={16} /> : <ChevronLeft size={16} />}
      </button>
    </aside>
  )
}
