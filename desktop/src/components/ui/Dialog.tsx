import { useEffect, useRef, type ReactNode } from 'react'
import { X } from 'lucide-react'
import { clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'
import Button from './Button'

interface DialogProps {
  open: boolean
  onClose: () => void
  title?: ReactNode
  description?: string
  children: ReactNode
  footer?: ReactNode
  size?: 'sm' | 'md' | 'lg'
  className?: string
}

const sizeStyles = {
  sm: 'max-w-sm',
  md: 'max-w-lg',
  lg: 'max-w-2xl',
}

export default function Dialog({
  open,
  onClose,
  title,
  description,
  children,
  footer,
  size = 'md',
  className,
}: DialogProps) {
  const overlayRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!open) return
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
    }
    document.addEventListener('keydown', handleEscape)
    return () => document.removeEventListener('keydown', handleEscape)
  }, [open, onClose])

  if (!open) return null

  return (
    <div
      ref={overlayRef}
      className="fixed inset-0 z-50 flex items-center justify-center"
      onClick={(e) => {
        if (e.target === overlayRef.current) onClose()
      }}
    >
      {/* 遮罩 */}
      <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" />

      {/* 对话框 */}
      <div
        className={twMerge(
          clsx(
            'relative w-full mx-4 rounded-lg border border-[#30363d] bg-[#161b22] shadow-2xl',
            'animate-in fade-in zoom-in-95 duration-150',
            sizeStyles[size],
            className
          )
        )}
      >
        {/* 头部 */}
        {(title || description) && (
          <div className="flex items-start justify-between px-5 pt-5 pb-3">
            <div>
              {title && <h2 className="text-base font-semibold text-[#e6edf3]">{title}</h2>}
              {description && <p className="mt-1 text-sm text-[#8b949e]">{description}</p>}
            </div>
            <button
              onClick={onClose}
              className="p-1 rounded-md text-[#8b949e] hover:text-[#e6edf3] hover:bg-[#21262d] transition-colors"
            >
              <X size={16} />
            </button>
          </div>
        )}

        {/* 内容 */}
        <div className="px-5 py-3">{children}</div>

        {/* 底部 */}
        {footer && (
          <div className="flex items-center justify-end gap-2 px-5 py-3 border-t border-[#30363d]">
            <Button variant="ghost" onClick={onClose}>取消</Button>
            {footer}
          </div>
        )}
      </div>
    </div>
  )
}
