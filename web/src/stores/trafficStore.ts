import { create } from 'zustand'
import type { Transaction } from '../types'

interface TrafficState {
  // 流量列表
  trafficList: Transaction[]
  // 选中的流量ID
  selectedId: string | null
  // 过滤条件
  filters: {
    method?: string
    status?: number
    host?: string
  }
  // 加载状态
  loading: boolean

  // Actions
  setTrafficList: (list: Transaction[]) => void
  addTraffic: (item: Transaction) => void
  removeTraffic: (id: string) => void
  clearTraffic: () => void
  setSelectedId: (id: string | null) => void
  setFilters: (filters: Partial<TrafficState['filters']>) => void
  setLoading: (loading: boolean) => void
}

export const useTrafficStore = create<TrafficState>((set) => ({
  trafficList: [],
  selectedId: null,
  filters: {},
  loading: false,

  setTrafficList: (list) => set({ trafficList: list }),

  addTraffic: (item) => set((state) => ({
    trafficList: [item, ...state.trafficList]
  })),

  removeTraffic: (id) => set((state) => ({
    trafficList: state.trafficList.filter((t) => t.id !== id),
    selectedId: state.selectedId === id ? null : state.selectedId
  })),

  clearTraffic: () => set({ trafficList: [], selectedId: null }),

  setSelectedId: (id) => set({ selectedId: id }),

  setFilters: (filters) => set((state) => ({
    filters: { ...state.filters, ...filters }
  })),

  setLoading: (loading) => set({ loading }),
}))