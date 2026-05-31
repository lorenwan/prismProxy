import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { ToastProvider } from './components/ui/Toast'
import Sidebar from './components/layout/Sidebar'
import Header from './components/layout/Header'
import StatusBar from './components/layout/StatusBar'
import TrafficPage from './pages/TrafficPage'
import RulesPage from './pages/RulesPage'
import BreakpointsPage from './pages/BreakpointsPage'
import AiPage from './pages/AiPage'
import SettingsPage from './pages/SettingsPage'
import RewritePage from './pages/RewritePage'
import CollectionsPage from './pages/CollectionsPage'
import EnvironmentsPage from './pages/EnvironmentsPage'
import ScriptsPage from './pages/ScriptsPage'
import DiffPage from './pages/DiffPage'
import PerformancePage from './pages/PerformancePage'

export default function App() {
  return (
    <ToastProvider>
      <BrowserRouter>
        <div className="flex h-screen bg-[#0d1117] text-[#e6edf3]">
          {/* 侧边栏 */}
          <Sidebar />

          {/* 主内容区 */}
          <div className="flex flex-col flex-1 overflow-hidden">
            {/* 头部工具栏 */}
            <Header />

            {/* 路由内容 */}
            <main className="flex-1 overflow-hidden bg-[#0d1117]">
              <Routes>
                <Route path="/" element={<TrafficPage />} />
                <Route path="/rules" element={<RulesPage />} />
                <Route path="/breakpoints" element={<BreakpointsPage />} />
                <Route path="/ai" element={<AiPage />} />
                <Route path="/settings" element={<SettingsPage />} />
                <Route path="/rewrite" element={<RewritePage />} />
                <Route path="/collections" element={<CollectionsPage />} />
                <Route path="/environments" element={<EnvironmentsPage />} />
                <Route path="/scripts" element={<ScriptsPage />} />
                <Route path="/diff" element={<DiffPage />} />
                <Route path="/performance" element={<PerformancePage />} />
              </Routes>
            </main>

            {/* 状态栏 */}
            <StatusBar />
          </div>
        </div>
      </BrowserRouter>
    </ToastProvider>
  )
}
