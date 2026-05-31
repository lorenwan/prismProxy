import { useEffect, useRef, useCallback } from 'react'
import { useTrafficStore } from '../stores/trafficStore'
import type { WsMessage, Transaction } from '../types'

// WebSocket 连接状态
type WsStatus = 'connecting' | 'connected' | 'disconnected'

export function useWebSocket() {
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimerRef = useRef<number | null>(null)
  const reconnectDelayRef = useRef(1000)
  const statusRef = useRef<WsStatus>('disconnected')

  const { addTraffic, removeTraffic, clearTraffic } = useTrafficStore()

  // 连接 WebSocket
  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) return

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${window.location.host}/ws`

    statusRef.current = 'connecting'
    const ws = new WebSocket(wsUrl)

    ws.onopen = () => {
      console.log('WebSocket 已连接')
      statusRef.current = 'connected'
      reconnectDelayRef.current = 1000 // 重置重连延迟
    }

    ws.onmessage = (event) => {
      try {
        const msg: WsMessage = JSON.parse(event.data)
        handleMessage(msg)
      } catch (err) {
        console.error('解析 WebSocket 消息失败:', err)
      }
    }

    ws.onclose = () => {
      console.log('WebSocket 已断开')
      statusRef.current = 'disconnected'
      wsRef.current = null

      // 指数退避重连
      reconnectTimerRef.current = window.setTimeout(() => {
        reconnectDelayRef.current = Math.min(reconnectDelayRef.current * 2, 30000)
        connect()
      }, reconnectDelayRef.current)
    }

    ws.onerror = (err) => {
      console.error('WebSocket 错误:', err)
    }

    wsRef.current = ws
  }, [addTraffic, removeTraffic, clearTraffic])

  // 处理消息
  const handleMessage = useCallback((msg: WsMessage) => {
    switch (msg.type) {
      case 'traffic:new':
        addTraffic(msg.payload as Transaction)
        break
      case 'traffic:delete':
        removeTraffic(msg.payload.id)
        break
      case 'traffic:clear':
        clearTraffic()
        break
      default:
        console.log('未知消息类型:', msg.type)
    }
  }, [addTraffic, removeTraffic, clearTraffic])

  // 断开连接
  const disconnect = useCallback(() => {
    if (reconnectTimerRef.current) {
      clearTimeout(reconnectTimerRef.current)
      reconnectTimerRef.current = null
    }
    if (wsRef.current) {
      wsRef.current.close()
      wsRef.current = null
    }
    statusRef.current = 'disconnected'
  }, [])

  useEffect(() => {
    connect()
    return () => disconnect()
  }, [connect, disconnect])

  return {
    status: statusRef.current,
    connect,
    disconnect,
  }
}