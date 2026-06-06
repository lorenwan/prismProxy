import { lazy, Suspense } from 'react'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { ToastProvider } from './components/ui/Toast'
import { TooltipProvider } from './components/ui/tooltip'
import { ErrorBoundary } from './components/ui/ErrorBoundary'
import Sidebar from './components/layout/Sidebar'
import Header from './components/layout/Header'
import StatusBar from './components/layout/StatusBar'

// 懒加载页面组件 - 实现代码分割，减少初始包大小
const TrafficPage = lazy(() => import('./pages/TrafficPage'))
const RulesPage = lazy(() => import('./pages/RulesPage'))
const BreakpointsPage = lazy(() => import('./pages/BreakpointsPage'))
const AiPage = lazy(() => import('./pages/AiPage'))
const SettingsPage = lazy(() => import('./pages/SettingsPage'))
const RewritePage = lazy(() => import('./pages/RewritePage'))
const CollectionsPage = lazy(() => import('./pages/CollectionsPage'))
const EnvironmentsPage = lazy(() => import('./pages/EnvironmentsPage'))
const ScriptsPage = lazy(() => import('./pages/ScriptsPage'))
const DiffPage = lazy(() => import('./pages/DiffPage'))
const PerformancePage = lazy(() => import('./pages/PerformancePage'))

// 页面加载指示器
function PageLoading() {
  return (
    <div className="flex items-center justify-center h-full">
      <div className="animate-spin h-6 w-6 border-2 border-[var(--blue)] border-t-transparent rounded-full" />
    </div>
  )
}

export default function App() {
  return (
    <TooltipProvider>
      <ToastProvider>
        <BrowserRouter>
          <div className="flex h-screen bg-[var(--bg-primary)] text-[var(--text-primary)]">
            {/* 侧边栏 */}
            <Sidebar />

            {/* 主内容区 */}
            <div className="flex flex-col flex-1 overflow-hidden">
              {/* 头部工具栏 */}
              <Header />

              {/* 路由内容 - ErrorBoundary 包裹 Suspense，正确捕获懒加载失败 */}
              <main className="flex-1 overflow-hidden bg-[var(--bg-primary)]" role="main">
                <ErrorBoundary>
                  <Suspense fallback={<PageLoading />}>
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
                  </Suspense>
                </ErrorBoundary>
              </main>

              {/* 状态栏 */}
              <StatusBar />
            </div>
          </div>
        </BrowserRouter>
      </ToastProvider>
    </TooltipProvider>
  )
}
