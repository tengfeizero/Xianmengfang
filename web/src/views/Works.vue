<template>
  <el-card shadow="never">
    <div class="toolbar">
      <el-select v-model="typeFilter" style="width: 160px" @change="load">
        <el-option label="全部类型" value="" />
        <el-option label="拼装" value="collage" />
        <el-option label="文生图" value="ai_text" />
        <el-option label="照片转风格" value="ai_photo" />
      </el-select>
      <el-select v-model="statusFilter" style="width: 140px" @change="load">
        <el-option label="全部状态" value="" />
        <el-option label="正常" value="1" />
        <el-option label="已下架" value="2" />
        <el-option label="已删除" value="0" />
      </el-select>
      <el-button type="primary" @click="load">查询</el-button>
    </div>

    <el-table :data="rows" v-loading="loading" stripe>
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
      <el-table-column label="创建时间" width="170">
        <template #default="{ row }">{{ new Date(row.created_at).toLocaleString() }}</template>
      </el-table-column>
      <el-table-column label="操作" width="110" fixed="right">
        <template #default="{ row }">
          <span v-if="!auth.hasPerm('work:edit')" class="deleted-hint">只读</span>
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
          <span v-else class="deleted-hint">用户已删除</span>
        </template>
      </el-table-column>
    </el-table>
  </el-card>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
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
.toolbar {
  display: flex;
  gap: 12px;
  margin-bottom: 14px;
}
.preview {
  width: 60px;
  height: 60px;
  border-radius: 4px;
}
.deleted-hint {
  color: #909399;
  font-size: 12px;
}
</style>
