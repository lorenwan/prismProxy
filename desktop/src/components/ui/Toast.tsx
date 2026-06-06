import { useEffect, useState, useCallback, createContext, useContext, type ReactNode } from 'react'
import { X, CheckCircle, AlertCircle, AlertTriangle, Info } from 'lucide-react'
import { clsx } from 'clsx'

type ToastType = 'success' | 'error' | 'warning' | 'info'

interface ToastItem {
  id: number
  type: ToastType
  message: string
}

interface ToastContextValue {
  toast: (type: ToastType, message: string) => void
}

const ToastContext = createContext<ToastContextValue | null>(null)

let toastId = 0

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<ToastItem[]>([])

  const removeToast = useCallback((id: number) => {
    setToasts((prev) => prev.filter((t) => t.id !== id))
  }, [])

  const toast = useCallback((type: ToastType, message: string) => {
    const id = ++toastId
    setToasts((prev) => [...prev, { id, type, message }])
    setTimeout(() => removeToast(id), 4000)
  }, [removeToast])

  return (
    <ToastContext.Provider value={{ toast }}>
      {children}
      {/* Toast 容器 */}
      <div className="fixed top-4 right-4 z-[100] flex flex-col gap-2 pointer-events-none">
        {toasts.map((t) => (
          <ToastItem key={t.id} item={t} onClose={() => removeToast(t.id)} />
        ))}
      </div>
    </ToastContext.Provider>
  )
}

export function useToast() {
  const ctx = useContext(ToastContext)
  if (!ctx) throw new Error('useToast must be used within ToastProvider')
  return ctx
}

const iconMap = {
  success: CheckCircle,
  error: AlertCircle,
  warning: AlertTriangle,
  info: Info,
}

const colorMap = {
  success: 'border-[var(--green)]/40 bg-[var(--green)]/10 text-[var(--green)]',
  error: 'border-[var(--red)]/40 bg-[var(--red)]/10 text-[var(--red)]',
  warning: 'border-[var(--yellow)]/40 bg-[var(--yellow)]/10 text-[var(--yellow)]',
  info: 'border-[var(--blue)]/40 bg-[var(--blue)]/10 text-[var(--blue)]',
}

function ToastItem({ item, onClose }: { item: ToastItem; onClose: () => void }) {
  const [exiting, setExiting] = useState(false)
  const Icon = iconMap[item.type]

  useEffect(() => {
    const timer = setTimeout(() => setExiting(true), 3600)
    return () => clearTimeout(timer)
  }, [])

  return (
    <div
      className={clsx(
        'pointer-events-auto flex items-center gap-2 px-4 py-2.5 rounded-lg border shadow-lg backdrop-blur-sm',
        'transition-all duration-300 min-w-[280px]',
        exiting ? 'opacity-0 translate-x-4' : 'opacity-100 translate-x-0',
        colorMap[item.type]
      )}
    >
      <Icon size={16} className="shrink-0" />
      <span className="flex-1 text-sm text-[var(--text-primary)]">{item.message}</span>
      <button onClick={onClose} className="shrink-0 p-0.5 hover:bg-white/10 rounded">
        <X size={14} />
      </button>
    </div>
  )
}
