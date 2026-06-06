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
import { Tooltip } from '../ui/tooltip'

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
        'flex flex-col bg-[var(--bg-secondary)] border-r border-[var(--border)] transition-all duration-200 ease-in-out',
        collapsed ? 'w-[52px]' : 'w-[180px]'
      )}
      aria-label="主导航"
    >
      {/* Logo */}
      <div className="flex items-center h-12 px-3 border-b border-[var(--border)]">
        <div className="w-7 h-7 rounded-lg bg-gradient-to-br from-[var(--blue)] to-[var(--purple)] flex items-center justify-center shrink-0 shadow-sm">
          <span className="text-white font-bold text-sm">P</span>
        </div>
        {!collapsed && (
          <span className="ml-2.5 text-sm font-semibold text-[var(--text-primary)] whitespace-nowrap overflow-hidden animate-fade-in">
            PrismProxy
          </span>
        )}
      </div>

      {/* 导航菜单 */}
      <nav className="flex-1 flex flex-col gap-0.5 p-2 overflow-y-auto" aria-label="功能导航">
        {navItems.map((item) => {
          const Icon = item.icon
          const navLink = (
            <NavLink
              key={item.path}
              to={item.path}
              end={item.path === '/'}
              className={({ isActive }) =>
                clsx(
                  'flex items-center gap-2.5 rounded-lg transition-all duration-150',
                  collapsed ? 'justify-center h-9 px-0' : 'h-8 px-2.5',
                  isActive
                    ? 'bg-[var(--blue)]/15 text-[var(--blue)] shadow-sm'
                    : 'text-[var(--text-secondary)] hover:bg-[var(--hover-bg)] hover:text-[var(--text-primary)]'
                )
              }
              aria-label={item.label}
            >
              <Icon size={18} className="shrink-0" aria-hidden="true" />
              {!collapsed && (
                <span className="text-sm whitespace-nowrap overflow-hidden">{item.label}</span>
              )}
            </NavLink>
          )

          if (collapsed) {
            return (
              <Tooltip key={item.path} content={item.label} side="right">
                {navLink}
              </Tooltip>
            )
          }

          return navLink
        })}
      </nav>

      {/* 折叠按钮 */}
      <Tooltip content={collapsed ? '展开侧边栏' : '折叠侧边栏'} side="right">
        <button
          onClick={() => setCollapsed(!collapsed)}
          className="flex items-center justify-center h-9 border-t border-[var(--border)] text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-[var(--hover-bg)] transition-colors"
          aria-label={collapsed ? '展开侧边栏' : '折叠侧边栏'}
        >
          {collapsed ? <ChevronRight size={16} /> : <ChevronLeft size={16} />}
        </button>
      </Tooltip>
    </aside>
  )
}
