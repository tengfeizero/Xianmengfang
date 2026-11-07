<template>
  <div class="page">
    <PageHeader title="用户管理" desc="小程序端注册用户；禁用后无法登录，已登录会话在下次请求时失效。" />

    <div class="panel">
      <div class="toolbar">
        <el-input
          v-model="keyword"
          placeholder="按昵称搜索"
          clearable
          :prefix-icon="Search"
          class="w240"
          @keyup.enter="load"
          @clear="load"
        />
        <el-select v-model="statusFilter" class="w140" @change="load">
          <el-option label="全部状态" value="" />
          <el-option label="正常" value="1" />
          <el-option label="禁用" value="0" />
        </el-select>
        <el-button @click="load">查询</el-button>
      </div>

      <el-table v-loading="loading" :data="pageRows">
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column label="头像" width="70">
        <template #default="{ row }">
          <el-avatar :size="36" :src="row.avatar">{{ (row.nickname || '?').slice(0, 1) }}</el-avatar>
        </template>
      </el-table-column>
      <el-table-column prop="nickname" label="昵称" min-width="140" show-overflow-tooltip />
      <el-table-column prop="order_count" label="订单数" width="90" />
      <el-table-column prop="work_count" label="作品数" width="90" />
      <el-table-column label="积分" width="90">
        <template #default="{ row }">{{ row.points || 0 }}</template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'" effect="light" size="small">
            {{ row.status === 1 ? '正常' : '禁用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="注册时间" width="170">
        <template #default="{ row }"><span class="muted">{{ fmtTime(row.created_at) }}</span></template>
      </el-table-column>
      <el-table-column label="操作" width="170" align="right" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="goLogs(row as AdminUserRow)">积分</el-button>
          <span v-if="!auth.hasPerm('user:edit')" class="readonly-hint">只读</span>
          <el-popconfirm
            v-else
            :title="row.status === 1 ? '禁用后该用户将无法登录，确认？' : '恢复该用户为正常状态？'"
            @confirm="toggle(row as AdminUserRow)"
          >
            <template #reference>
              <el-button link :type="row.status === 1 ? 'danger' : 'success'">
                {{ row.status === 1 ? '禁用' : '启用' }}
              </el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>

      <TablePager v-model:page="page" v-model:size="size" :total="total" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import PageHeader from '@/components/PageHeader.vue'
import TablePager from '@/components/TablePager.vue'
import { usePager } from '@/utils/pager'
import { listUsers, setUserStatus, type AdminUserRow } from '@/api/modules/users'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()

const rows = ref<AdminUserRow[]>([])
const keyword = ref('')
const statusFilter = ref('')
const loading = ref(false)

/** 前端分页：改搜索/状态筛选回到第 1 页 */
const { page, size, total, rows: pageRows } = usePager(rows, () => [keyword.value, statusFilter.value])

onMounted(load)

async function load() {
  loading.value = true
  try {
    rows.value = await listUsers({ keyword: keyword.value.trim(), status: statusFilter.value })
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

/** 下钻：只看该用户的积分流水 */
function goLogs(row: AdminUserRow) {
  router.push({ path: '/point-logs', query: { user_id: row.id } })
}

async function toggle(row: AdminUserRow) {
  const next: 0 | 1 = row.status === 1 ? 0 : 1
  try {
    await setUserStatus(row.id, next)
    row.status = next
    ElMessage.success(next === 1 ? '已启用' : '已禁用')
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '操作失败')
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
</style>
