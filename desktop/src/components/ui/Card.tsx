import type { HTMLAttributes, ReactNode } from 'react'
import { clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

interface CardProps extends HTMLAttributes<HTMLDivElement> {
  header?: ReactNode
  footer?: ReactNode
  noPadding?: boolean
}

export function Card({ header, footer, noPadding, className, children, ...props }: CardProps) {
  return (
    <div
      className={twMerge(
        clsx(
          'rounded-lg border border-[#30363d] bg-[#161b22]',
          className
        )
      )}
      {...props}
    >
      {header && (
        <div className="border-b border-[#30363d] px-4 py-2.5 text-sm font-medium text-[#e6edf3]">
          {header}
        </div>
      )}
      <div className={clsx(!noPadding && 'p-4')}>{children}</div>
      {footer && (
        <div className="border-t border-[#30363d] px-4 py-2.5">
          {footer}
        </div>
      )}
    </div>
  )
}

interface CardGroupProps extends HTMLAttributes<HTMLDivElement> {
  direction?: 'horizontal' | 'vertical'
}

export function CardGroup({ direction = 'vertical', className, children, ...props }: CardGroupProps) {
  return (
    <div
      className={twMerge(
        clsx(
          'flex gap-3',
          direction === 'vertical' ? 'flex-col' : 'flex-row',
          className
        )
      )}
      {...props}
    >
      {children}
    </div>
  )
}
