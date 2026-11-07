<template>
  <div class="page">
    <PageHeader
      title="使用明细"
      desc="作品被下到哪一笔订单：下单时绑定的定制品作品会在这里留痕。使用次数在订单完成时按件数结算并计入积分，因此本页也用于核对积分发放。"
    />

    <StatStrip :items="stats" />

    <div class="panel">
      <div class="toolbar">
        <el-input
          v-model="keyword"
          placeholder="按订单号或买家昵称搜索"
          clearable
          :prefix-icon="Search"
          class="w240"
          @keyup.enter="load"
          @clear="load"
        />
        <el-input v-model="workID" placeholder="作品 ID" clearable class="w120" @keyup.enter="load" />
        <el-input v-model="userID" placeholder="买家 ID" clearable class="w120" @keyup.enter="load" />
        <el-select v-model="statusFilter" class="w140" @change="load">
          <el-option v-for="s in STATUS_OPTIONS" :key="s.value" :label="s.label" :value="s.value" />
        </el-select>
        <el-button type="primary" @click="load">查询</el-button>
        <el-button v-if="hasFilter" @click="reset">重置</el-button>
        <span class="spacer" />
        <span class="muted">共 {{ total }} 条使用记录{{ capped ? '，仅展示最近 2000 条' : '' }}</span>
      </div>

      <el-table v-loading="loading" :data="pageRows" empty-text="还没有作品被用于下单">
        <el-table-column label="时间" width="160">
          <template #default="{ row }"><span class="muted">{{ fmtTime(row.created_at) }}</span></template>
        </el-table-column>
        <el-table-column label="订单号" width="180">
          <template #default="{ row }"><span class="mono">{{ row.order_no }}</span></template>
        </el-table-column>
        <el-table-column label="作品" min-width="260">
          <template #default="{ row }">
            <div class="work-cell">
              <el-image
                v-if="row.work_image_url"
                :src="row.work_image || row.work_image_url"
                :preview-src-list="[row.work_image_url]"
                preview-teleported
                class="thumb"
                fit="cover"
              />
              <div v-else class="thumb thumb-empty" />
              <div class="work-meta">
                <span class="work-prompt">{{ row.work_prompt || '（无提示词）' }}</span>
                <span class="work-sub">#{{ row.work_id }} · {{ typeText(row.work_type) }}</span>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="定制商品" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">
            <span>{{ row.title }}</span>
            <span v-if="row.specs" class="spec">{{ specText(row.specs) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="件数" width="80" align="right">
          <template #default="{ row }">{{ row.qty }}</template>
        </el-table-column>
        <el-table-column label="买家" width="150">
          <template #default="{ row }">
            <span>{{ row.nickname || '—' }}</span>
            <span class="uid">#{{ row.user_id }}</span>
          </template>
        </el-table-column>
        <el-table-column label="订单状态" width="110">
          <template #default="{ row }">
            <el-tag :type="statusTag(row.status)" size="small" effect="light">
              {{ statusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>

      <TablePager v-model:page="page" v-model:size="size" :total="pageTotal" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import PageHeader from '@/components/PageHeader.vue'
import StatStrip, { type StatItem } from '@/components/StatStrip.vue'
import TablePager from '@/components/TablePager.vue'
import { usePager } from '@/utils/pager'
import { listWorkUsages, type WorkUsageRow } from '@/api/modules/insights'

const STATUS_OPTIONS = [
  { value: '', label: '全部状态' },
  { value: 'pending_accept', label: '待接单' },
  { value: 'producing', label: '生产中' },
  { value: 'packing', label: '打包中' },
  { value: 'pending_ship', label: '待发货' },
  { value: 'shipped', label: '已发货' },
  { value: 'done', label: '已完成' },
  { value: 'canceled', label: '已取消' },
]

const STATUS_LABEL: Record<string, string> = {
  pending_accept: '待接单',
  producing: '生产中',
  packing: '打包中',
  pending_ship: '待发货',
  shipped: '已发货',
  done: '已完成',
  canceled: '已取消',
}

const route = useRoute()

const rows = ref<WorkUsageRow[]>([])
const keyword = ref('')
const workID = ref(String(route.query.work_id ?? ''))
const userID = ref(String(route.query.user_id ?? ''))
const statusFilter = ref('')
const total = ref(0)
const qtySum = ref(0)
const loading = ref(false)

const { page, size, total: pageTotal, rows: pageRows } = usePager(rows, () => [
  keyword.value,
  workID.value,
  userID.value,
  statusFilter.value,
])

const hasFilter = computed(() => !!(keyword.value || workID.value || userID.value || statusFilter.value))
const capped = computed(() => total.value > rows.value.length)

const stats = computed<StatItem[]>(() => {
  const works = new Set(rows.value.map((r) => r.work_id))
  return [
    { label: '使用记录', value: total.value, unit: '条' },
    { label: '累计件数', value: qtySum.value, unit: '件' },
    { label: '涉及作品', value: works.size, unit: '件' },
  ]
})

onMounted(load)

async function load() {
  loading.value = true
  try {
    const data = await listWorkUsages({
      keyword: keyword.value.trim(),
      work_id: workID.value.trim(),
      user_id: userID.value.trim(),
      status: statusFilter.value,
    })
    rows.value = data.items ?? []
    total.value = data.total ?? rows.value.length
    qtySum.value = data.summary?.qty ?? 0
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

function reset() {
  keyword.value = ''
  workID.value = ''
  userID.value = ''
  statusFilter.value = ''
  load()
}

function typeText(t: string) {
  return t === 'collage' ? '拼装' : t === 'ai_text' ? '文生图' : t === 'ai_photo' ? '照片转风格' : t || '—'
}
function statusText(s: string) {
  return STATUS_LABEL[s] ?? s
}
function statusTag(s: string) {
  if (s === 'done') return 'success'
  if (s === 'canceled') return 'info'
  if (s === 'shipped') return 'primary'
  return 'warning'
}

/** specs 是 JSON 字符串（{"尺码":"M"}），拍平成「尺码 M」 */
function specText(raw: string) {
  try {
    const obj = JSON.parse(raw) as Record<string, string>
    const parts = Object.entries(obj).map(([k, v]) => `${k} ${v}`)
    return parts.length ? ` · ${parts.join(' / ')}` : ''
  } catch {
    return raw ? ` · ${raw}` : ''
  }
}

function fmtTime(v: string | null | undefined): string {
  if (!v) return '—'
  const d = new Date(v)
  if (isNaN(d.getTime())) return '—'
  const p = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}`
}
</script>

<style scoped>
.w240 {
  width: 240px;
}
.w140 {
  width: 140px;
}
.w120 {
  width: 120px;
}
.uid {
  margin-left: 6px;
  font-size: 12px;
  color: var(--text-4);
}
.spec {
  margin-left: 6px;
  font-size: 12px;
  color: var(--text-3);
}
.work-cell {
  display: flex;
  align-items: center;
  gap: 10px;
}
.thumb {
  width: 40px;
  height: 40px;
  border-radius: 6px;
  background: var(--fill-1);
  flex-shrink: 0;
  display: block;
}
.thumb-empty {
  border: 1px dashed var(--line-1);
}
.work-meta {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}
.work-prompt {
  font-size: 13px;
  color: var(--text-1);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.work-sub {
  font-size: 12px;
  color: var(--text-3);
}
</style>
