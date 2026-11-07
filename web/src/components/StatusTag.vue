<template>
  <el-tag size="default" :style="tagStyle">{{ meta.text }}</el-tag>
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
  pending_accept: { text: '待接单', color: '#e6a23c', bg: '#fdf6ec' },
  producing: { text: '生产中', color: '#409eff', bg: '#ecf5ff' },
  packing: { text: '打包中', color: '#8b5cf6', bg: '#f4f0ff' },
  pending_ship: { text: '待发货', color: '#14b8a6', bg: '#e8fbf8' },
  shipped: { text: '已发货', color: '#67c23a', bg: '#f0f9eb' },
  done: { text: '已完成', color: '#909399', bg: '#f4f4f5' },
  canceled: { text: '已取消', color: '#f56c6c', bg: '#fef0f0' },
}

const meta = computed<StatusMeta>(
  () => MAP[props.status] ?? { text: String(props.status), color: '#909399', bg: '#f4f4f5' },
)

const tagStyle = computed(() => ({
  color: meta.value.color,
  backgroundColor: meta.value.bg,
  borderColor: meta.value.color,
}))
</script>
