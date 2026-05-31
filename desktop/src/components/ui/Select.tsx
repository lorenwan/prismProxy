import { forwardRef, type SelectHTMLAttributes } from 'react'
import { clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

interface SelectProps extends SelectHTMLAttributes<HTMLSelectElement> {
  error?: string
}

const Select = forwardRef<HTMLSelectElement, SelectProps>(
  ({ error, className, children, ...props }, ref) => {
    return (
      <div className="flex flex-col gap-1">
        <select
          ref={ref}
          className={twMerge(
            clsx(
              'w-full h-8 rounded-md border bg-[#0d1117] text-[#e6edf3] text-sm px-3',
              'transition-colors appearance-none cursor-pointer',
              'focus:outline-none focus:ring-2 focus:ring-[#58a6ff]/40 focus:border-[#58a6ff]',
              'bg-[url("data:image/svg+xml,%3Csvg%20xmlns%3D%22http%3A//www.w3.org/2000/svg%22%20width%3D%2212%22%20height%3D%2212%22%20viewBox%3D%220%200%2012%2012%22%3E%3Cpath%20fill%3D%22%238b949e%22%20d%3D%22M6%208L1%203h10z%22/%3E%3C/svg%3E")]',
              'bg-no-repeat bg-[right_8px_center]',
              error
                ? 'border-[#f85149]'
                : 'border-[#30363d] hover:border-[#484f58]',
              'pr-7',
              className
            )
          )}
          {...props}
        >
          {children}
        </select>
        {error && <p className="text-xs text-[#f85149]">{error}</p>}
      </div>
    )
  }
)

Select.displayName = 'Select'
export default Select
