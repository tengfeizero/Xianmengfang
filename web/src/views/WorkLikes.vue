<template>
  <div class="page">
    <PageHeader
      title="点赞明细"
      desc="谁在什么时候赞了哪件作品。点赞是「借用作品下单」的前置动作，同一用户对同一作品只记一条，重复点赞不会重复计数。"
    />

    <StatStrip :items="stats" />

    <div class="panel">
      <div class="toolbar">
        <el-input
          v-model="keyword"
          placeholder="按昵称或作品提示词搜索"
          clearable
          :prefix-icon="Search"
          class="w240"
          @keyup.enter="load"
          @clear="load"
        />
        <el-input v-model="workID" placeholder="作品 ID" clearable class="w120" @keyup.enter="load" />
        <el-input v-model="userID" placeholder="用户 ID" clearable class="w120" @keyup.enter="load" />
        <el-button type="primary" @click="load">查询</el-button>
        <el-button v-if="hasFilter" @click="reset">重置</el-button>
        <span class="spacer" />
        <span class="muted">共 {{ total }} 条点赞{{ capped ? '，仅展示最近 2000 条' : '' }}</span>
      </div>

      <el-table v-loading="loading" :data="pageRows" empty-text="还没有人点赞">
        <el-table-column label="时间" width="160">
          <template #default="{ row }"><span class="muted">{{ fmtTime(row.created_at) }}</span></template>
        </el-table-column>
        <el-table-column label="作品" min-width="280">
          <template #default="{ row }">
            <div class="work-cell">
              <el-image
                v-if="row.work.image_url"
                :src="row.work.thumbnail_url || row.work.image_url"
                :preview-src-list="[row.work.image_url]"
                preview-teleported
                class="thumb"
                fit="cover"
              />
              <div v-else class="thumb thumb-empty" />
              <div class="work-meta">
                <span class="work-prompt">{{ row.work.prompt || '（无提示词）' }}</span>
                <span class="work-sub">#{{ row.work.id }} · {{ typeText(row.work.type) }}</span>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="作者" width="140">
          <template #default="{ row }">{{ row.work.author_name || `用户#${row.work.author_id}` }}</template>
        </el-table-column>
        <el-table-column label="点赞人" width="160">
          <template #default="{ row }">
            <span>{{ row.nickname || '—' }}</span>
            <span class="uid">#{{ row.user_id }}</span>
          </template>
        </el-table-column>
        <el-table-column label="广场" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.work.is_public === 1" type="primary" effect="light" size="small">已发布</el-tag>
            <span v-else class="muted">私密</span>
          </template>
        </el-table-column>
        <el-table-column label="作品状态" width="110">
          <template #default="{ row }">
            <el-tag :type="statusTag(row.work.status)" size="small">
              {{ statusText(row.work.status) }}
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
import { listWorkLikes, type WorkLikeRow } from '@/api/modules/insights'

const route = useRoute()

const rows = ref<WorkLikeRow[]>([])
const keyword = ref('')
const workID = ref(String(route.query.work_id ?? ''))
const userID = ref(String(route.query.user_id ?? ''))
const total = ref(0)
const loading = ref(false)

const { page, size, total: pageTotal, rows: pageRows } = usePager(rows, () => [
  keyword.value,
  workID.value,
  userID.value,
])

const hasFilter = computed(() => !!(keyword.value || workID.value || userID.value))
const capped = computed(() => total.value > rows.value.length)

const stats = computed<StatItem[]>(() => {
  const works = new Set(rows.value.map((r) => r.work.id))
  const users = new Set(rows.value.map((r) => r.user_id))
  return [
    { label: '点赞总数', value: total.value, unit: '次' },
    { label: '涉及作品', value: works.size, unit: '件' },
    { label: '参与用户', value: users.size, unit: '人' },
  ]
})

onMounted(load)

async function load() {
  loading.value = true
  try {
    const data = await listWorkLikes({
      keyword: keyword.value.trim(),
      work_id: workID.value.trim(),
      user_id: userID.value.trim(),
    })
    rows.value = data.items ?? []
    total.value = data.total ?? rows.value.length
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
  load()
}

function typeText(t: string) {
  return t === 'collage' ? '拼装' : t === 'ai_text' ? '文生图' : t === 'ai_photo' ? '照片转风格' : t || '—'
}
function statusText(s: number) {
  return s === 1 ? '正常' : s === 2 ? '已下架' : '已删除'
}
function statusTag(s: number) {
  return s === 1 ? 'success' : s === 2 ? 'warning' : 'info'
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
.w120 {
  width: 120px;
}
.uid {
  margin-left: 6px;
  font-size: 12px;
  color: var(--text-4);
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
