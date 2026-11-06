<template>
  <el-tag :type="meta.type" size="default">{{ meta.text }}</el-tag>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { OrderStatus } from '@/types'

const props = defineProps<{ status: OrderStatus | string }>()

type TagType = 'primary' | 'success' | 'warning' | 'info' | 'danger'

const MAP: Record<string, { text: string; type: TagType }> = {
  pending_accept: { text: '待接单', type: 'warning' },
  pending_ship: { text: '待发货', type: 'primary' },
  shipped: { text: '已发货', type: 'success' },
  done: { text: '已完成', type: 'info' },
  canceled: { text: '已取消', type: 'danger' },
}

const meta = computed(() => MAP[props.status] ?? { text: String(props.status), type: 'info' as TagType })
</script>
