import { computed, ref, watch, type ComputedRef, type Ref } from 'vue'

/** 全站列表统一分页规格：默认 10 条/页，可选 10 / 20 / 50 / 100 */
export const PAGE_SIZES = [10, 20, 50, 100]
export const DEFAULT_PAGE_SIZE = PAGE_SIZES[0]

/**
 * 前端分页：数据已全量在内存时使用（后台多数接口一次返回全量，便于页签计数/本地筛选）。
 * - rows：当前页要渲染的切片
 * - total：全量条数（分页组件与「共 N 条」都用它）
 * - resetOn：筛选条件变化时自动回到第 1 页，避免停留在越界页
 */
export function usePager<T>(source: Ref<T[]> | ComputedRef<T[]>, resetOn?: () => unknown) {
  const page = ref(1)
  const size = ref(DEFAULT_PAGE_SIZE)

  const total = computed(() => source.value.length)
  const maxPage = computed(() => Math.max(1, Math.ceil(total.value / size.value)))
  const rows = computed(() => {
    const start = (page.value - 1) * size.value
    return source.value.slice(start, start + size.value)
  })

  // 删除/筛选后总页数变少，把页码收敛回有效范围
  watch([total, size], () => {
    if (page.value > maxPage.value) page.value = maxPage.value
  })

  if (resetOn) watch(resetOn, () => (page.value = 1), { deep: true })

  return { page, size, total, maxPage, rows }
}
