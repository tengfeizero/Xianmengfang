<template>
  <span class="status-pill" :style="{ color: meta.color, background: meta.bg }">
    <i class="dot" :style="{ background: meta.color }" />
    {{ meta.text }}
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { OrderStatus } from '@/types'

const props = defineProps<{ status: OrderStatus | string }>()

interface StatusMeta {
  text: string
  color: string
  bg: string
}

const MAP: Record<string, StatusMeta> = {
  pending_accept: { text: '待接单', color: '#e08600', bg: '#fdf4e6' },
  producing: { text: '生产中', color: '#2b6ef2', bg: '#eef3ff' },
  packing: { text: '打包中', color: '#7c5cf0', bg: '#f3efff' },
  pending_ship: { text: '待发货', color: '#0d8f8f', bg: '#e6f7f7' },
  shipped: { text: '已发货', color: '#12a150', bg: '#e9f8ef' },
  done: { text: '已完成', color: '#7c8496', bg: '#f2f4f7' },
  canceled: { text: '已取消', color: '#b0b6c0', bg: '#f5f6f8' },
}

const meta = computed<StatusMeta>(
  () => MAP[props.status] ?? { text: String(props.status), color: '#7c8496', bg: '#f2f4f7' },
)
</script>

<style scoped>
.status-pill {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  height: 22px;
  padding: 0 8px;
  border-radius: 11px;
  font-size: 12px;
  font-weight: 500;
  line-height: 1;
  white-space: nowrap;
}
.dot {
  width: 5px;
  height: 5px;
  border-radius: 50%;
  flex-shrink: 0;
}
</style>
