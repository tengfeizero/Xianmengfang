<template>
  <div class="page">
    <PageHeader
      title="积分明细"
      desc="小程序端的积分流水：作品被使用、AI 生成抵扣、后台调整全部留痕，只追加不可改。余额与流水同事务写入，可按用户或原因追溯。"
    />

    <StatStrip :items="stats" />

    <div class="panel">
      <div class="toolbar">
        <el-input
          v-model="keyword"
          placeholder="按昵称搜索"
          clearable
          :prefix-icon="Search"
          class="w200"
          @keyup.enter="load"
          @clear="load"
        />
        <el-select v-model="reason" class="w160" @change="load">
          <el-option label="全部原因" value="" />
          <el-option label="作品被使用" value="work_used" />
          <el-option label="AI 生成抵扣" value="ai_spend" />
          <el-option label="后台调整" value="admin_adjust" />
        </el-select>
        <el-input v-model="userID" placeholder="用户 ID" clearable class="w120" @keyup.enter="load" />
        <el-button type="primary" @click="load">查询</el-button>
        <el-button v-if="hasFilter" @click="reset">重置</el-button>
        <span class="spacer" />
        <span class="muted">共 {{ total }} 笔流水{{ capped ? '，仅展示最近 2000 条' : '' }}</span>
      </div>

      <el-table v-loading="loading" :data="pageRows" empty-text="没有符合条件的积分流水">
        <el-table-column label="时间" width="160">
          <template #default="{ row }"><span class="muted">{{ fmtTime(row.created_at) }}</span></template>
        </el-table-column>
        <el-table-column label="用户" min-width="150">
          <template #default="{ row }">
            <span>{{ row.nickname || '—' }}</span>
            <span class="uid">#{{ row.user_id }}</span>
          </template>
        </el-table-column>
        <el-table-column label="变动" width="100" align="right">
          <template #default="{ row }">
            <span class="delta" :class="row.delta > 0 ? 'up' : 'down'">
              {{ row.delta > 0 ? `+${row.delta}` : row.delta }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="变动后余额" width="110" align="right">
          <template #default="{ row }">{{ row.balance }}</template>
        </el-table-column>
        <el-table-column label="原因" width="130">
          <template #default="{ row }">
            <el-tag :type="reasonTag(row.reason)" size="small" effect="light">{{ row.reason_text }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="关联单据" width="130">
          <template #default="{ row }">
            <span v-if="row.ref_id" class="muted">{{ refText(row as PointLogRow) }} #{{ row.ref_id }}</span>
            <span v-else class="muted">—</span>
          </template>
        </el-table-column>
        <el-table-column label="备注" min-width="220" show-overflow-tooltip>
          <template #default="{ row }">{{ row.remark || '—' }}</template>
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
import { listPointLogs, type PointLogRow } from '@/api/modules/insights'

const route = useRoute()

const rows = ref<PointLogRow[]>([])
const keyword = ref('')
const reason = ref('')
const userID = ref(String(route.query.user_id ?? ''))
const total = ref(0)
const summary = ref({ in: 0, out: 0 })
const loading = ref(false)

const { page, size, total: pageTotal, rows: pageRows } = usePager(rows, () => [
  keyword.value,
  reason.value,
  userID.value,
])

const hasFilter = computed(() => !!(keyword.value || reason.value || userID.value))
/** 后端单次最多回 2000 条，超出时提示口径 */
const capped = computed(() => total.value > rows.value.length)

const stats = computed<StatItem[]>(() => [
  { label: '流水笔数', value: total.value, unit: '笔' },
  { label: '累计获得', value: summary.value.in, prefix: '+', unit: '分', tone: 'up' },
  { label: '累计消耗', value: summary.value.out, prefix: '−', unit: '分', tone: 'down' },
])

onMounted(load)

async function load() {
  loading.value = true
  try {
    const data = await listPointLogs({
      keyword: keyword.value.trim(),
      reason: reason.value,
      user_id: userID.value.trim(),
    })
    rows.value = data.items ?? []
    total.value = data.total ?? rows.value.length
    summary.value = data.summary ?? { in: 0, out: 0 }
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

function reset() {
  keyword.value = ''
  reason.value = ''
  userID.value = ''
  load()
}

function reasonTag(r: string) {
  return r === 'work_used' ? 'success' : r === 'ai_spend' ? 'warning' : 'info'
}

function refText(row: PointLogRow) {
  if (row.ref_type === 'order') return '订单'
  if (row.ref_type === 'ai_task') return 'AI 任务'
  if (row.ref_type === 'work') return '作品'
  return row.ref_type || '单据'
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
.w200 {
  width: 200px;
}
.w160 {
  width: 160px;
}
.w120 {
  width: 120px;
}
.uid {
  margin-left: 6px;
  font-size: 12px;
  color: var(--text-4);
}
.delta {
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}
.delta.up {
  color: var(--el-color-success);
}
.delta.down {
  color: var(--el-color-danger);
}
</style>
