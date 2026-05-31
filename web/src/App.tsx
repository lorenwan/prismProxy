import { BrowserRouter, Routes, Route } from 'react-router-dom'
import Sidebar from './components/layout/Sidebar'
import Header from './components/layout/Header'
import StatusBar from './components/layout/StatusBar'
import TrafficPage from './pages/TrafficPage'
import RulesPage from './pages/RulesPage'
import BreakpointsPage from './pages/BreakpointsPage'
import AiPage from './pages/AiPage'
import SettingsPage from './pages/SettingsPage'

export default function App() {
  return (
    <BrowserRouter>
      <div className="flex h-screen bg-[#1a1b26] text-[#a9b1d6]">
        {/* 侧边栏 */}
        <Sidebar />

        {/* 主内容区 */}
        <div className="flex flex-col flex-1 overflow-hidden">
          {/* 头部 */}
          <Header />

          {/* 路由内容 */}
          <main className="flex-1 overflow-hidden">
            <Routes>
              <Route path="/" element={<TrafficPage />} />
              <Route path="/rules" element={<RulesPage />} />
              <Route path="/breakpoints" element={<BreakpointsPage />} />
              <Route path="/ai" element={<AiPage />} />
              <Route path="/settings" element={<SettingsPage />} />
            </Routes>
          </main>

          {/* 状态栏 */}
          <StatusBar />
        </div>
      </div>
    </BrowserRouter>
  )
}