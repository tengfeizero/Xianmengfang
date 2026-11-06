import { http } from '@/api/http'
import type { ConfigItem } from '@/types'

export const CONFIG_KEYS = {
  aiDailyFreeLimit: 'ai_daily_free_limit',
} as const

export function listConfigs() {
  return http.get<ConfigItem[]>('/admin/configs')
}

export function updateConfig(key: string, value: string) {
  return http.put('/admin/configs', { key, value })
}
