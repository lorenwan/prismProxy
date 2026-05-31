import { NavLink } from 'react-router-dom'

// 侧边栏导航项
const navItems = [
  { path: '/', label: '流量', icon: '📊' },
  { path: '/rules', label: '规则', icon: '📋' },
  { path: '/breakpoints', label: '断点', icon: '⏸️' },
  { path: '/ai', label: 'AI', icon: '🤖' },
  { path: '/settings', label: '设置', icon: '⚙️' },
]

export default function Sidebar() {
  return (
    <aside className="w-14 bg-[#16161e] flex flex-col items-center py-4 border-r border-[#3b4261]">
      {/* Logo */}
      <div className="w-10 h-10 bg-[#7aa2f7] rounded-lg flex items-center justify-center mb-6">
        <span className="text-white font-bold text-lg">P</span>
      </div>

      {/* 导航菜单 */}
      <nav className="flex-1 flex flex-col gap-2">
        {navItems.map((item) => (
          <NavLink
            key={item.path}
            to={item.path}
            className={({ isActive }) =>
              `w-10 h-10 flex items-center justify-center rounded-lg transition-colors ${
                isActive
                  ? 'bg-[#283457] text-[#7aa2f7]'
                  : 'text-[#565f89] hover:bg-[#24283b] hover:text-[#a9b1d6]'
              }`
            }
            title={item.label}
          >
            <span className="text-xl">{item.icon}</span>
          </NavLink>
        ))}
      </nav>
    </aside>
  )
}