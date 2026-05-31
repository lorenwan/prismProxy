import { forwardRef, type InputHTMLAttributes, type ReactNode } from 'react'
import { clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  icon?: ReactNode
  suffix?: ReactNode
  error?: string
}

const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ icon, suffix, error, className, ...props }, ref) => {
    return (
      <div className="flex flex-col gap-1">
        <div className="relative flex items-center">
          {icon && (
            <span className="absolute left-2.5 text-[#8b949e] pointer-events-none">
              {icon}
            </span>
          )}
          <input
            ref={ref}
            className={twMerge(
              clsx(
                'w-full h-8 rounded-md border bg-[#0d1117] text-[#e6edf3] text-sm',
                'placeholder:text-[#484f58] transition-colors',
                'focus:outline-none focus:ring-2 focus:ring-[#58a6ff]/40 focus:border-[#58a6ff]',
                error
                  ? 'border-[#f85149]'
                  : 'border-[#30363d] hover:border-[#484f58]',
                icon ? 'pl-8' : 'pl-3',
                suffix ? 'pr-8' : 'pr-3',
                className
              )
            )}
            {...props}
          />
          {suffix && (
            <span className="absolute right-2.5 text-[#8b949e]">
              {suffix}
            </span>
          )}
        </div>
        {error && <p className="text-xs text-[#f85149]">{error}</p>}
      </div>
    )
  }
)

Input.displayName = 'Input'
export default Input
