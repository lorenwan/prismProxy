import { create } from 'zustand'
import type { Transaction } from '../../types'

// 流量列表最大容量，防止长时间抓包后内存溢出
const MAX_TRAFFIC_ENTRIES = 5000

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
  updateTraffic: (item: Transaction) => void
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

  addTraffic: (item) => set((state) => {
    const newList = [item, ...state.trafficList]
    return {
      trafficList: newList.length > MAX_TRAFFIC_ENTRIES
        ? newList.slice(0, MAX_TRAFFIC_ENTRIES)
        : newList
    }
  }),

  updateTraffic: (item) => set((state) => ({
    trafficList: state.trafficList.map(t => t.id === item.id ? item : t)
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