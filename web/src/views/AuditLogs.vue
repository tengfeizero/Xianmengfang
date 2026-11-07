<template>
  <div class="page">
    <PageHeader title="操作日志" desc="记录后台全部写操作、登录与越权拦截；只追加、不可删除，密码与密钥已脱敏。" />

    <div class="panel">
      <div class="toolbar">
        <el-input
          v-model="filters.keyword"
          placeholder="关键字（操作人 / 路径 / 目标 / 明细）"
          clearable
          :prefix-icon="Search"
          class="w280"
          @keyup.enter="search"
        />
        <el-input v-model="filters.action" placeholder="动作，如 order.ship" clearable class="w200" @keyup.enter="search" />
        <el-select v-model="filters.result" placeholder="全部结果" clearable class="w140" @change="search">
          <el-option label="成功" value="ok" />
          <el-option label="失败" value="fail" />
          <el-option label="越权拦截" value="deny" />
        </el-select>
        <el-button type="primary" @click="search">查询</el-button>
        <el-button @click="reset">重置</el-button>
      </div>

      <el-table v-loading="loading" :data="rows">
        <el-table-column type="expand">
        <template #default="{ row }">
          <div class="detail">
            <p><b>路径：</b><span class="mono">{{ row.method }} {{ row.path }}</span></p>
            <p><b>目标：</b>{{ row.target_type }} {{ row.target_id ? '#' + row.target_id : '—' }}</p>
            <p><b>请求摘要：</b></p>
            <pre class="json">{{ pretty(row.detail) }}</pre>
            <p><b>User-Agent：</b><span class="muted">{{ row.user_agent || '—' }}</span></p>
            <p v-if="row.err_msg"><b>失败原因：</b><span class="err">{{ row.err_msg }}</span></p>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="时间" width="170">
        <template #default="{ row }">{{ fmtTime(row.created_at) }}</template>
      </el-table-column>
      <el-table-column label="操作人" width="130">
        <template #default="{ row }">{{ row.admin_name || '—' }}<span v-if="row.admin_id" class="muted"> #{{ row.admin_id }}</span></template>
      </el-table-column>
      <el-table-column label="动作" width="170">
        <template #default="{ row }">
          <span>{{ actionText(row.action) }}</span>
          <div class="mono sub">{{ row.action }}</div>
        </template>
      </el-table-column>
      <el-table-column label="目标" width="120">
        <template #default="{ row }">
          <span v-if="row.target_id">{{ row.target_type }} #{{ row.target_id }}</span>
          <span v-else class="muted">{{ row.target_type || '—' }}</span>
        </template>
      </el-table-column>
      <el-table-column label="结果" width="100">
        <template #default="{ row }">
          <el-tag :type="RESULT_TAG[row.result] || 'info'" size="small">{{ RESULT_TEXT[row.result] || row.result }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="状态码" width="80">
        <template #default="{ row }">{{ row.status_code }}</template>
      </el-table-column>
      <el-table-column prop="ip" label="IP" width="130" />
    </el-table>

      <div class="panel-foot">
        <span class="hint">共 {{ total }} 条记录</span>
        <el-pagination
          layout="sizes, prev, pager, next"
          :total="total"
          :page-size="filters.size"
          :current-page="filters.page"
          :page-sizes="[20, 50, 100]"
          @current-change="onPage"
          @size-change="onSize"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { Search } from '@element-plus/icons-vue'
import PageHeader from '@/components/PageHeader.vue'
import { listAuditLogs } from '@/api/modules/rbac'
import type { AuditLogRow } from '@/types'

const rows = ref<AuditLogRow[]>([])
const total = ref(0)
const loading = ref(false)
const filters = reactive({ keyword: '', action: '', result: '', page: 1, size: 20 })

const ACTION_TEXT: Record<string, string> = {
  'auth.login': '登录',
  'order.accept': '接单',
  'order.pack': '生产完成',
  'order.ready_ship': '打包完成',
  'order.ship': '发货',
  'order.cancel': '取消订单',
  'order.refund': '标记退款',
  'config.update': '修改系统配置',
  'user.status': '启用/禁用用户',
  'work.status': '作品上下架',
  'admin.create': '新增管理员',
  'admin.update': '修改管理员',
  'admin.delete': '删除管理员',
  'role.create': '新增角色',
  'role.update': '修改角色',
  'role.delete': '删除角色',
  'asset.upload': '上传文件',
  deny: '越权访问被拦截',
}

const RESULT_TAG: Record<string, 'success' | 'danger' | 'warning' | 'info'> = {
  ok: 'success',
  fail: 'danger',
  deny: 'warning',
}
const RESULT_TEXT: Record<string, string> = { ok: '成功', fail: '失败', deny: '越权拦截' }

function actionText(a: string): string {
  if (ACTION_TEXT[a]) return ACTION_TEXT[a]
  if (a.startsWith('asset.')) return '素材操作'
  if (a.startsWith('product.')) return '商品操作'
  return a
}

function pretty(raw: string): string {
  if (!raw) return '—'
  try {
    return JSON.stringify(JSON.parse(raw), null, 2)
  } catch {
    return raw
  }
}

async function load() {
  loading.value = true
  try {
    const page = await listAuditLogs({ ...filters })
    rows.value = page.list
    total.value = page.total
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

function search() {
  filters.page = 1
  load()
}

function reset() {
  filters.keyword = ''
  filters.action = ''
  filters.result = ''
  filters.page = 1
  load()
}

function onPage(p: number) {
  filters.page = p
  load()
}

function onSize(s: number) {
  filters.size = s
  filters.page = 1
  load()
}

function fmtTime(v: string): string {
  if (!v) return '—'
  const d = new Date(v)
  if (isNaN(d.getTime())) return '—'
  const p = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}:${p(d.getSeconds())}`
}

onMounted(load)
</script>

<style scoped>
.w280 {
  width: 280px;
}
.w200 {
  width: 200px;
}
.w140 {
  width: 140px;
}
.detail {
  padding: 8px 24px;
  font-size: 13px;
}
.detail p {
  margin: 6px 0;
}
.json {
  background: var(--el-fill-color-light);
  padding: 8px 12px;
  border-radius: 4px;
  max-height: 220px;
  overflow: auto;
  font-size: 12px;
  white-space: pre-wrap;
  word-break: break-all;
}
.mono {
  font-family: ui-monospace, Menlo, Consolas, monospace;
}
.sub {
  color: var(--el-text-color-placeholder);
  font-size: 12px;
}
.muted {
  color: var(--el-text-color-secondary);
}
.err {
  color: var(--el-color-danger);
}
.pager {
  display: flex;
  justify-content: flex-end;
  margin-top: 12px;
}
</style>
