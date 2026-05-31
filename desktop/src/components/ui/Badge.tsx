import type { HTMLAttributes, ReactNode } from 'react'
import { clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

type BadgeVariant = 'default' | 'blue' | 'green' | 'yellow' | 'red' | 'purple'

interface BadgeProps extends HTMLAttributes<HTMLSpanElement> {
  variant?: BadgeVariant
  icon?: ReactNode
}

const variantStyles: Record<BadgeVariant, string> = {
  default: 'bg-[#21262d] text-[#8b949e] border-[#30363d]',
  blue: 'bg-[#58a6ff]/15 text-[#58a6ff] border-[#58a6ff]/30',
  green: 'bg-[#3fb950]/15 text-[#3fb950] border-[#3fb950]/30',
  yellow: 'bg-[#d29922]/15 text-[#d29922] border-[#d29922]/30',
  red: 'bg-[#f85149]/15 text-[#f85149] border-[#f85149]/30',
  purple: 'bg-[#bc8cff]/15 text-[#bc8cff] border-[#bc8cff]/30',
}

// HTTP 方法 -> Badge 配色
export const methodBadgeVariant: Record<string, BadgeVariant> = {
  GET: 'green',
  POST: 'blue',
  PUT: 'yellow',
  DELETE: 'red',
  PATCH: 'purple',
}

export default function Badge({ variant = 'default', icon, className, children, ...props }: BadgeProps) {
  return (
    <span
      className={twMerge(
        clsx(
          'inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-xs font-medium',
          variantStyles[variant],
          className
        )
      )}
      {...props}
    >
      {icon}
      {children}
    </span>
  )
}
