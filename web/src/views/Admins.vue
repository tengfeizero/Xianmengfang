<template>
  <div class="page">
    <PageHeader title="管理员账号" desc="按岗位分配角色；禁用账号或调整角色权限会立即生效，无需等待登录过期。">
      <template #actions>
        <el-button v-if="auth.hasPerm('admin:edit')" type="primary" :icon="Plus" @click="openCreate">
          新增管理员
        </el-button>
      </template>
    </PageHeader>

    <div class="panel">
      <el-table v-loading="loading" :data="admins">
      <el-table-column prop="username" label="用户名" width="140">
        <template #default="{ row }">
          <span class="mono">{{ row.username }}</span>
          <el-tag v-if="row.id === auth.me?.id" size="small" class="me">当前</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="nickname" label="昵称" width="140" />
      <el-table-column label="角色" width="150">
        <template #default="{ row }">
          <el-tag :type="row.role_code === 'super' ? 'danger' : 'primary'" size="small">{{ row.role_name || '未分配' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">{{ row.status === 1 ? '正常' : '已禁用' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="最近登录" width="170">
        <template #default="{ row }">{{ fmtTime(row.last_login_at) }}</template>
      </el-table-column>
      <el-table-column label="创建时间" width="170">
        <template #default="{ row }">{{ fmtTime(row.created_at) }}</template>
      </el-table-column>
      <el-table-column label="操作" min-width="180" align="right" fixed="right">
        <template #default="{ row }">
          <template v-if="auth.hasPerm('admin:edit')">
            <el-button link type="primary" @click="openEdit(row as AdminRow)">编辑</el-button>
            <el-button
              v-if="row.status === 1"
              link
              type="warning"
              :disabled="row.id === auth.me?.id"
              @click="toggleStatus(row as AdminRow)"
            >
              禁用
            </el-button>
            <el-button v-else link type="success" @click="toggleStatus(row as AdminRow)">启用</el-button>
            <el-button link type="danger" :disabled="row.id === auth.me?.id" @click="remove(row as AdminRow)">
              删除
            </el-button>
          </template>
          <span v-else class="readonly-hint">只读</span>
        </template>
      </el-table-column>
    </el-table>
    </div>

    <el-dialog v-model="dialogVisible" :title="editing ? '编辑管理员' : '新增管理员'" width="480px">
      <el-form label-width="90px">
        <el-form-item label="用户名" required>
          <el-input v-model="form.username" :disabled="!!editing" placeholder="登录名，创建后不可改" maxlength="32" />
        </el-form-item>
        <el-form-item label="昵称">
          <el-input v-model="form.nickname" placeholder="显示名称" maxlength="20" />
        </el-form-item>
        <el-form-item label="角色" required>
          <el-select v-model="form.role_id" placeholder="选择角色" class="w-full">
            <el-option v-for="r in roles" :key="r.id" :label="r.name" :value="r.id">
              <span>{{ r.name }}</span>
              <span class="opt-desc">{{ r.perms.includes('*') ? '全部权限' : `${r.perms.length} 项权限` }}</span>
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item v-if="editing" label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" active-text="正常" inactive-text="禁用" />
        </el-form-item>
        <el-form-item :label="editing ? '重置密码' : '初始密码'">
          <el-input
            v-model="form.password"
            type="password"
            show-password
            :placeholder="editing ? '留空则不修改' : '至少 6 位'"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { Plus } from '@element-plus/icons-vue'
import PageHeader from '@/components/PageHeader.vue'
import { createAdmin, deleteAdmin, listAdmins, listRoles, updateAdmin } from '@/api/modules/rbac'
import { useAuthStore } from '@/stores/auth'
import type { AdminRow, RoleRow } from '@/types'

const auth = useAuthStore()
const admins = ref<AdminRow[]>([])
const roles = ref<RoleRow[]>([])
const loading = ref(false)
const saving = ref(false)
const dialogVisible = ref(false)
const editing = ref<AdminRow | null>(null)
const form = reactive({ username: '', nickname: '', role_id: 0, status: 1, password: '' })

async function load() {
  loading.value = true
  try {
    ;[admins.value, roles.value] = await Promise.all([listAdmins(), listRoles()])
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editing.value = null
  form.username = ''
  form.nickname = ''
  form.role_id = roles.value.find((r) => r.code !== 'super')?.id ?? roles.value[0]?.id ?? 0
  form.status = 1
  form.password = ''
  dialogVisible.value = true
}

function openEdit(row: AdminRow) {
  editing.value = row
  form.username = row.username
  form.nickname = row.nickname
  form.role_id = row.role_id
  form.status = row.status
  form.password = ''
  dialogVisible.value = true
}

async function save() {
  if (!editing.value) {
    if (!form.username.trim()) {
      ElMessage.warning('请填写用户名')
      return
    }
    if (form.password.length < 6) {
      ElMessage.warning('初始密码至少 6 位')
      return
    }
  }
  if (!form.role_id) {
    ElMessage.warning('请选择角色')
    return
  }
  saving.value = true
  try {
    if (editing.value) {
      await updateAdmin(editing.value.id, {
        nickname: form.nickname,
        role_id: form.role_id,
        status: form.status,
        password: form.password || undefined,
      })
    } else {
      await createAdmin({
        username: form.username.trim(),
        password: form.password,
        nickname: form.nickname,
        role_id: form.role_id,
      })
    }
    ElMessage.success('已保存')
    dialogVisible.value = false
    load()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '保存失败')
  } finally {
    saving.value = false
  }
}

async function toggleStatus(row: AdminRow) {
  const to = row.status === 1 ? 0 : 1
  try {
    await updateAdmin(row.id, { nickname: row.nickname, role_id: row.role_id, status: to })
    ElMessage.success(to === 1 ? '已启用' : '已禁用')
    load()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '操作失败')
  }
}

async function remove(row: AdminRow) {
  try {
    await ElMessageBox.confirm(`确定删除管理员「${row.username}」？操作不可恢复。`, '删除管理员', { type: 'warning' })
  } catch {
    return
  }
  try {
    await deleteAdmin(row.id)
    ElMessage.success('已删除')
    load()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '删除失败')
  }
}

function fmtTime(v: string | null | undefined): string {
  if (!v) return '—'
  const d = new Date(v)
  if (isNaN(d.getTime())) return '—'
  const p = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}`
}

onMounted(load)
</script>

<style scoped>
.mono {
  font-family: ui-monospace, Menlo, Consolas, monospace;
}
.me {
  margin-left: 6px;
}
.w-full {
  width: 100%;
}
.opt-desc {
  float: right;
  color: var(--el-text-color-placeholder);
  font-size: 12px;
}
</style>
