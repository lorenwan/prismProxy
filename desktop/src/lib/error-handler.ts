import { useToast } from '../components/ui/Toast'

/**
 * 共享错误处理 hook
 * 统一处理异步操作的错误：记录日志 + 显示 Toast 通知
 * 替代散落在各页面的 alert() 和 console.error()
 */
export function useErrorHandler() {
  const { toast } = useToast()

  return function handleError(err: unknown, fallbackMessage: string): void {
    const message = err instanceof Error ? err.message : fallbackMessage
    console.error(fallbackMessage, err)
    toast('error', message)
  }
}
