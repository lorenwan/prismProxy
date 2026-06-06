import { Component, type ErrorInfo, type ReactNode } from 'react'

interface Props {
  children: ReactNode
  fallback?: ReactNode
}

interface State {
  hasError: boolean
  error?: Error
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = { hasError: false }
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('ErrorBoundary caught error:', error, errorInfo)
  }

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback
      }
      return (
        <div className="flex items-center justify-center h-full">
          <div className="text-center space-y-3">
            <p className="text-sm text-[var(--red)]">页面出错了</p>
            <p className="text-xs text-[var(--text-tertiary)]">{this.state.error?.message}</p>
            <button
              onClick={() => this.setState({ hasError: false })}
              className="px-3 py-1.5 text-xs bg-[var(--bg-secondary)] border border-[var(--border)] rounded hover:bg-[var(--hover-bg)]"
            >
              重试
            </button>
          </div>
        </div>
      )
    }

    return this.props.children
  }
}
