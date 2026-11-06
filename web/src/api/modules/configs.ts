import { http } from '@/api/http'
import type { ConfigItem } from '@/types'

export const CONFIG_KEYS = {
  aiDailyFreeLimit: 'ai_daily_free_limit',
  aiApiKey: 'ai_api_key',
  aiBaseUrl: 'ai_base_url',
  aiModel: 'ai_model',
  aiStyle: 'ai_style',
} as const

export function listConfigs() {
  return http.get<ConfigItem[]>('/admin/configs')
}

export function updateConfig(key: string, value: string) {
  return http.put('/admin/configs', { key, value })
}
