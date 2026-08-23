import { defineStore } from 'pinia'
import { api } from '../services/api'
import type { EventSummary } from '../types/domain'
export const useEventStore = defineStore('events', { state: () => ({ items: [] as EventSummary[], total: 0, loading: false, error: '' }), actions: { async load(page = 0) { this.loading = true; this.error = ''; try { const result = await api.events(page); this.items = result.items; this.total = result.total } catch (error) { this.error = error instanceof Error ? error.message : '无法加载舞会' } finally { this.loading = false } } } })
