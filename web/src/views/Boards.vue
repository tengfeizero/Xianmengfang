<template>
  <div class="page">
    <PageHeader title="拼装底板" desc="底板是小程序拼装页的底图；元素按底板归属，一块底板对应一套元素，互不串用。">
      <template #actions>
        <el-button v-if="auth.hasPerm('asset:edit')" type="primary" :icon="Plus" @click="openCreate">
          新增底板
        </el-button>
      </template>
    </PageHeader>

    <div class="panel">
      <el-table v-loading="loading" :data="pageRows">
        <el-table-column label="预览" width="90">
          <template #default="{ row }">
            <el-image :src="row.url" :preview-src-list="[row.url]" preview-teleported class="preview" fit="contain" />
          </template>
        </el-table-column>
        <el-table-column prop="name" label="名称" min-width="160" />
        <el-table-column label="画布尺寸" width="140">
          <template #default="{ row }">
            <span class="mono muted">{{ row.width }} × {{ row.height }}</span>
          </template>
        </el-table-column>
        <el-table-column label="元素" width="220">
          <template #default="{ row }">
            <div class="el-cell">
              <el-tag :type="row.element_count ? 'primary' : 'info'" size="small" effect="light">
                {{ row.element_count || 0 }} 个
              </el-tag>
              <el-button link type="primary" @click="goElements(row as Board)">管理元素</el-button>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" effect="light" size="small">
              {{ row.status === 1 ? '上架' : '下架' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160" align="right">
          <template #default="{ row }">
            <template v-if="auth.hasPerm('asset:edit')">
              <el-button link type="primary" @click="openEdit(row as Board)">编辑</el-button>
              <el-popconfirm
                v-if="row.status === 1"
                title="下架后小程序不再可选该底板；底板下还有启用元素时不允许下架"
                @confirm="offline(row as Board)"
              >
                <template #reference><el-button link type="warning">下架</el-button></template>
              </el-popconfirm>
            </template>
            <span v-else class="readonly-hint">只读</span>
          </template>
        </el-table-column>
      </el-table>

      <TablePager v-model:page="page" v-model:size="size" :total="total" />
    </div>

    <el-dialog v-model="dialogVisible" :title="editing ? '编辑底板' : '新增底板'" width="460px">
      <el-form label-width="82px">
        <el-form-item label="名称" required>
          <el-input v-model="form.name" maxlength="32" placeholder="如：底板·小姑娘" />
        </el-form-item>
        <el-form-item label="底图" :required="!editing">
          <el-upload :show-file-list="false" :auto-upload="false" accept="image/png" :on-change="onFile">
            <el-button>选择 PNG（透明背景）</el-button>
          </el-upload>
          <span v-if="form.url" class="uploaded">{{ editing ? '已保留原图（可重传替换）' : '已上传' }}</span>
        </el-form-item>
        <el-form-item label="画布宽">
          <el-input-number v-model="form.width" :min="100" :max="2000" />
          <span class="form-tip">元素锚点按该尺寸计算</span>
        </el-form-item>
        <el-form-item label="画布高">
          <el-input-number v-model="form.height" :min="100" :max="2000" />
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
import { useRouter } from 'vue-router'
import type { UploadFile } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import PageHeader from '@/components/PageHeader.vue'
import TablePager from '@/components/TablePager.vue'
import { createBoard, listBoards, toggleBoard, updateBoard, uploadImage } from '@/api/modules/assets'
import { usePager } from '@/utils/pager'
import { useAuthStore } from '@/stores/auth'
import type { Board } from '@/types'

const auth = useAuthStore()
const router = useRouter()

const boards = ref<Board[]>([])
const loading = ref(false)
const saving = ref(false)
const dialogVisible = ref(false)
const editing = ref<Board | null>(null)
const form = reactive({ name: '', url: '', width: 400, height: 500 })

const { page, size, total, rows: pageRows } = usePager(boards)

async function load() {
  loading.value = true
  try {
    boards.value = await listBoards()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editing.value = null
  form.name = ''
  form.url = ''
  form.width = 400
  form.height = 500
  dialogVisible.value = true
}

function openEdit(b: Board) {
  editing.value = b
  form.name = b.name
  form.url = b.url
  form.width = b.width || 400
  form.height = b.height || 500
  dialogVisible.value = true
}

async function onFile(f: UploadFile) {
  if (!f.raw) return
  try {
    form.url = await uploadImage(f.raw)
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '上传失败')
  }
}

async function save() {
  if (!form.name.trim() || !form.url) {
    ElMessage.warning('请填写名称并上传底图')
    return
  }
  saving.value = true
  try {
    if (editing.value) await updateBoard(editing.value.id, { ...form })
    else await createBoard({ ...form })
    ElMessage.success('已保存')
    dialogVisible.value = false
    load()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '保存失败')
  } finally {
    saving.value = false
  }
}

async function offline(b: Board) {
  try {
    await toggleBoard(b.id)
    ElMessage.success('底板已下架')
    load()
  } catch (e) {
    // 409：底板下还有启用元素，直接展示后端提示
    ElMessage.error(e instanceof Error ? e.message : '下架失败')
  }
}

function goElements(b: Board) {
  router.push({ path: '/elements', query: { board_id: String(b.id) } })
}

onMounted(load)
</script>

<style scoped>
.preview {
  width: 56px;
  height: 70px;
  border-radius: 6px;
  background: var(--fill-1);
  display: block;
}
.el-cell {
  display: flex;
  align-items: center;
  gap: 10px;
}
.mono {
  font-family: ui-monospace, Menlo, Consolas, monospace;
}
.muted {
  color: var(--text-3);
}
.uploaded {
  margin-left: 8px;
  font-size: 12px;
  color: var(--el-color-success);
}
.form-tip {
  margin-left: 8px;
  font-size: 12px;
  color: var(--text-3);
}
</style>
