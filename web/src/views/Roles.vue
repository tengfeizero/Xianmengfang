<template>
  <div class="page">
    <PageHeader title="角色权限" desc="角色决定后台可见菜单与可执行操作；账号在「管理员账号」里分配角色。">
      <template #actions>
        <el-button v-if="auth.hasPerm('role:edit')" type="primary" :icon="Plus" @click="openCreate">
          新增角色
        </el-button>
      </template>
    </PageHeader>

    <div class="panel">
      <el-table v-loading="loading" :data="roles">
      <el-table-column prop="name" label="角色名称" width="160" />
      <el-table-column label="标识" width="110">
        <template #default="{ row }"><span class="mono">{{ row.code }}</span></template>
      </el-table-column>
      <el-table-column prop="remark" label="说明" min-width="220" show-overflow-tooltip />
      <el-table-column label="权限" min-width="220">
        <template #default="{ row }">
          <el-tag v-if="row.perms.includes('*')" type="danger" size="small">全部权限</el-tag>
          <span v-else class="perm-count">{{ row.perms.length }} 项</span>
          <span v-if="!row.perms.includes('*')" class="perm-list">{{ permNames(row.perms) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="账号数" width="80">
        <template #default="{ row }">{{ row.admin_count }}</template>
      </el-table-column>
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">{{ row.status === 1 ? '启用' : '停用' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="类型" width="90">
        <template #default="{ row }">
          <el-tag v-if="row.builtin" type="warning" size="small">内置</el-tag>
          <span v-else>自定义</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="140" align="right">
        <template #default="{ row }">
          <el-button v-if="auth.hasPerm('role:edit')" link type="primary" @click="openEdit(row as RoleRow)">
            编辑
          </el-button>
          <el-button
            v-if="auth.hasPerm('role:edit') && !row.builtin"
            link
            type="danger"
            @click="remove(row as RoleRow)"
          >
            删除
          </el-button>
          <span v-if="!auth.hasPerm('role:edit')" class="readonly-hint">只读</span>
        </template>
      </el-table-column>
    </el-table>
    </div>

    <el-dialog v-model="dialogVisible" :title="editing ? '编辑角色' : '新增角色'" width="720px" top="6vh">
      <el-form label-width="90px">
        <el-form-item label="角色名称" required>
          <el-input v-model="form.name" placeholder="如：仓库发货员" maxlength="20" />
        </el-form-item>
        <el-form-item label="说明">
          <el-input v-model="form.remark" placeholder="这个角色负责什么" maxlength="120" />
        </el-form-item>
        <el-form-item v-if="editing" label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" active-text="启用" inactive-text="停用" />
        </el-form-item>
        <el-form-item label="权限">
          <el-alert
            v-if="form.isSuper"
            type="warning"
            :closable="false"
            show-icon
            title="超级管理员拥有全部权限，权限不可修改"
            class="mb"
          />
          <div class="perm-box">
            <div v-for="g in groups" :key="g.key" class="perm-group">
              <div class="group-head">
                <el-checkbox
                  :model-value="groupChecked(g)"
                  :indeterminate="groupIndeterminate(g)"
                  :disabled="form.isSuper"
                  @change="(v: string | number | boolean) => toggleGroup(g, !!v)"
                >
                  <b>{{ g.name }}</b>
                </el-checkbox>
              </div>
              <div class="group-items">
                <el-checkbox
                  v-for="it in g.items"
                  :key="it.code"
                  v-model="checked[it.code]"
                  :disabled="form.isSuper"
                >
                  {{ it.name }}
                  <span class="code">{{ it.code }}</span>
                </el-checkbox>
              </div>
            </div>
          </div>
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
import { createRole, deleteRole, getPermCatalog, listRoles, updateRole } from '@/api/modules/rbac'
import { useAuthStore } from '@/stores/auth'
import type { PermGroup, RoleRow } from '@/types'

const auth = useAuthStore()
const roles = ref<RoleRow[]>([])
const groups = ref<PermGroup[]>([])
const loading = ref(false)
const saving = ref(false)
const dialogVisible = ref(false)
const editing = ref<RoleRow | null>(null)

const form = reactive({ name: '', remark: '', status: 1, isSuper: false })
const checked = reactive<Record<string, boolean>>({})

// 权限码 → 中文名，表格里展示可读性更好
const permNameMap = ref<Record<string, string>>({})

async function load() {
  loading.value = true
  try {
    roles.value = await listRoles()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

async function loadGroups() {
  try {
    groups.value = await getPermCatalog()
    const m: Record<string, string> = {}
    for (const g of groups.value) for (const it of g.items) m[it.code] = it.name
    permNameMap.value = m
  } catch {
    // 权限清单拉取失败不阻断列表展示
  }
}

function permNames(codes: string[]): string {
  return codes
    .slice(0, 4)
    .map((c) => permNameMap.value[c] ?? c)
    .join('、') + (codes.length > 4 ? ` 等 ${codes.length} 项` : '')
}

function resetChecked() {
  for (const k of Object.keys(checked)) delete checked[k]
}

function openCreate() {
  editing.value = null
  form.name = ''
  form.remark = ''
  form.status = 1
  form.isSuper = false
  resetChecked()
  dialogVisible.value = true
}

function openEdit(row: RoleRow) {
  editing.value = row
  form.name = row.name
  form.remark = row.remark
  form.status = row.status
  form.isSuper = row.perms.includes('*')
  resetChecked()
  for (const c of row.perms) if (c !== '*') checked[c] = true
  dialogVisible.value = true
}

function groupChecked(g: PermGroup): boolean {
  return g.items.every((it) => checked[it.code])
}

function groupIndeterminate(g: PermGroup): boolean {
  const n = g.items.filter((it) => checked[it.code]).length
  return n > 0 && n < g.items.length
}

function toggleGroup(g: PermGroup, v: boolean) {
  for (const it of g.items) checked[it.code] = v
}

async function save() {
  if (!form.name.trim()) {
    ElMessage.warning('请填写角色名称')
    return
  }
  const perms = form.isSuper ? ['*'] : Object.keys(checked).filter((k) => checked[k])
  if (!perms.length) {
    ElMessage.warning('请至少勾选一项权限')
    return
  }
  saving.value = true
  try {
    if (editing.value) {
      await updateRole(editing.value.id, { name: form.name, remark: form.remark, perms, status: form.status })
    } else {
      await createRole({ name: form.name, remark: form.remark, perms })
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

async function remove(row: RoleRow) {
  try {
    await ElMessageBox.confirm(`确定删除角色「${row.name}」？`, '删除角色', { type: 'warning' })
  } catch {
    return
  }
  try {
    await deleteRole(row.id)
    ElMessage.success('已删除')
    load()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '删除失败')
  }
}

onMounted(() => {
  load()
  loadGroups()
})
</script>

<style scoped>
.mono {
  font-family: ui-monospace, Menlo, Consolas, monospace;
  color: var(--text-3);
}
.perm-count {
  margin-right: 8px;
}
.perm-list {
  color: var(--text-3);
  font-size: 12px;
}
.perm-box {
  width: 100%;
  max-height: 46vh;
  overflow-y: auto;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  padding: 8px 12px;
}
.perm-group + .perm-group {
  margin-top: 10px;
  border-top: 1px dashed var(--el-border-color-lighter);
  padding-top: 10px;
}
.group-items {
  display: flex;
  flex-wrap: wrap;
  gap: 2px 20px;
  padding-left: 24px;
}
.group-items :deep(.el-checkbox) {
  margin-right: 0;
}
.code {
  margin-left: 4px;
  color: var(--el-text-color-placeholder);
  font-size: 12px;
}
.mb {
  margin-bottom: 8px;
}
</style>
