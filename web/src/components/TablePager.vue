<template>
  <div class="panel-foot">
    <span class="hint">
      共 {{ total }} 条<template v-if="total">，当前 {{ rangeText }}</template>
    </span>
    <el-pagination
      layout="sizes, prev, pager, next"
      :total="total"
      :page-size="size"
      :current-page="page"
      :page-sizes="PAGE_SIZES"
      background
      @current-change="onPage"
      @size-change="onSize"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { PAGE_SIZES } from '@/utils/pager'

/** 列表底部统一分页条：默认 10 条/页，可选 10/20/50/100；数据由各页自行分页（前端或后端） */
const props = defineProps<{ total: number; page: number; size: number }>()
const emit = defineEmits<{ 'update:page': [number]; 'update:size': [number] }>()

const rangeText = computed(() => {
  const from = (props.page - 1) * props.size + 1
  const to = Math.min(props.page * props.size, props.total)
  return `${from}-${to}`
})

function onPage(p: number) {
  emit('update:page', p)
}

/** 改每页条数后回到第 1 页，避免停在越界页 */
function onSize(s: number) {
  emit('update:size', s)
  emit('update:page', 1)
}
</script>
