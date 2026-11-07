<template>
  <div class="page">
    <PageHeader title="用户作品" desc="小程序用户的创作结果；下架后用户端与加购均不可见，用户自行删除的作品不可恢复。" />

    <div class="panel">
      <div class="toolbar">
        <el-select v-model="typeFilter" class="w160" @change="load">
          <el-option label="全部类型" value="" />
          <el-option label="拼装" value="collage" />
          <el-option label="文生图" value="ai_text" />
          <el-option label="照片转风格" value="ai_photo" />
        </el-select>
        <el-select v-model="statusFilter" class="w140" @change="load">
          <el-option label="全部状态" value="" />
          <el-option label="正常" value="1" />
          <el-option label="已下架" value="2" />
          <el-option label="已删除" value="0" />
        </el-select>
        <el-button @click="load">查询</el-button>
        <span class="spacer" />
        <span class="count-hint">共 {{ rows.length }} 件</span>
      </div>

      <el-table v-loading="loading" :data="rows">
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column label="预览" width="90">
        <template #default="{ row }">
          <el-image
            :src="row.image_url"
            :preview-src-list="[row.image_url]"
            preview-teleported
            class="preview"
            fit="cover"
          />
        </template>
      </el-table-column>
      <el-table-column label="类型" width="110">
        <template #default="{ row }">
          <el-tag :type="typeTag(row.type)" size="small">{{ typeText(row.type) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="作者" width="140">
        <template #default="{ row }">{{ row.nickname || `用户#${row.user_id}` }}</template>
      </el-table-column>
      <el-table-column label="生成参数" min-width="200" show-overflow-tooltip>
        <template #default="{ row }">{{ promptOf(row as AdminWorkRow) }}</template>
      </el-table-column>
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : row.status === 2 ? 'warning' : 'info'" size="small">
            {{ statusText(row.status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="创建时间" width="160">
        <template #default="{ row }"><span class="muted">{{ fmtTime(row.created_at) }}</span></template>
      </el-table-column>
      <el-table-column label="操作" width="110" fixed="right" align="right">
        <template #default="{ row }">
          <span v-if="!auth.hasPerm('work:edit')" class="readonly-hint">只读</span>
          <template v-else-if="row.status === 1">
            <el-popconfirm title="下架后用户端将不可见，确认？" @confirm="setStatus(row as AdminWorkRow, 2)">
              <template #reference>
                <el-button link type="warning">下架</el-button>
              </template>
            </el-popconfirm>
          </template>
          <el-button v-else-if="row.status === 2" link type="success" @click="setStatus(row as AdminWorkRow, 1)">
            恢复上架
          </el-button>
          <span v-else class="readonly-hint">用户已删除</span>
        </template>
      </el-table-column>
    </el-table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import PageHeader from '@/components/PageHeader.vue'
import { listWorks, setWorkStatus, type AdminWorkRow } from '@/api/modules/works'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()

const rows = ref<AdminWorkRow[]>([])
const typeFilter = ref('')
const statusFilter = ref('')
const loading = ref(false)

onMounted(load)

async function load() {
  loading.value = true
  try {
    rows.value = await listWorks({ type: typeFilter.value, status: statusFilter.value })
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

function promptOf(row: AdminWorkRow): string {
  try {
    const m = JSON.parse(row.meta) as { prompt?: string }
    return m.prompt || ''
  } catch {
    return ''
  }
}

function typeText(t: string) {
  return t === 'collage' ? '拼装' : t === 'ai_text' ? '文生图' : '照片转风格'
}
function typeTag(t: string) {
  return t === 'collage' ? 'primary' : t === 'ai_text' ? 'success' : 'warning'
}
function statusText(s: number) {
  return s === 1 ? '正常' : s === 2 ? '已下架' : '已删除'
}

function fmtTime(v: string | null | undefined): string {
  if (!v) return '—'
  const d = new Date(v)
  if (isNaN(d.getTime())) return '—'
  const p = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}`
}

async function setStatus(row: AdminWorkRow, status: 1 | 2) {
  try {
    await setWorkStatus(row.id, status)
    row.status = status
    ElMessage.success(status === 1 ? '已恢复上架' : '已下架')
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '操作失败')
  }
}
</script>

<style scoped>
.w160 {
  width: 160px;
}
.w140 {
  width: 140px;
}
.preview {
  width: 48px;
  height: 48px;
  border-radius: 6px;
  background: var(--fill-1);
  display: block;
}
.count-hint {
  font-size: 12px;
  color: var(--text-3);
}
</style>
