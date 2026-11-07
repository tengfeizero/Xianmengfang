<template>
  <div class="page">
    <PageHeader title="拼装元素" desc="元素挂在具体底板下，只会在该底板的拼装页出现；换底板后原来的元素不会再被混用。">
      <template #actions>
        <el-button
          v-if="auth.hasPerm('asset:edit')"
          type="primary"
          :icon="Plus"
          :disabled="!currentBoard || currentBoard.status !== 1"
          @click="openCreate"
        >
          新增元素
        </el-button>
      </template>
    </PageHeader>

    <div v-if="!loading && !boards.length" class="panel">
      <div class="empty">
        <p>还没有底板，元素必须挂在底板上。</p>
        <el-button type="primary" @click="router.push('/boards')">去新增底板</el-button>
      </div>
    </div>

    <div v-else class="panel">
      <div class="toolbar">
        <el-select v-model="boardId" class="w240" placeholder="选择底板" @change="onBoardChange">
          <el-option v-for="b in boards" :key="b.id" :label="b.name" :value="b.id">
            <span>{{ b.name }}</span>
            <span class="opt-desc">{{ b.status === 1 ? `${b.element_count || 0} 个元素` : '已下架' }}</span>
          </el-option>
        </el-select>
        <el-select v-model="catFilter" class="w140" clearable placeholder="全部分类" @change="load" @clear="load">
          <el-option v-for="c in CATEGORIES" :key="c.value" :label="c.label" :value="c.value" />
        </el-select>
        <el-button :icon="Refresh" :loading="loading" @click="load">刷新</el-button>
        <span class="spacer" />
        <span v-if="currentBoard" class="board-meta">
          画布 {{ currentBoard.width }} × {{ currentBoard.height }}
          <el-tag v-if="currentBoard.status !== 1" type="info" size="small" effect="light">底板已下架</el-tag>
        </span>
      </div>

      <el-table v-loading="loading" :data="pageRows">
        <el-table-column label="预览" width="80">
          <template #default="{ row }">
            <el-image :src="row.image_url" class="preview" fit="contain" />
          </template>
        </el-table-column>
        <el-table-column prop="name" label="名称" width="150" />
        <el-table-column label="分类" width="110">
          <template #default="{ row }">
            <el-tag size="small" type="info" effect="plain">{{ catLabel(row.category) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="默认锚点" min-width="200">
          <template #default="{ row }"><span class="mono muted">{{ row.anchor }}</span></template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" effect="light" size="small">
              {{ row.status === 1 ? '上架' : '下架' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="140" align="right">
          <template #default="{ row }">
            <template v-if="auth.hasPerm('asset:edit')">
              <el-button link type="primary" @click="openEdit(row as ElementItem)">编辑</el-button>
              <el-button link :type="row.status === 1 ? 'warning' : 'success'" @click="toggleStatus(row as ElementItem)">
                {{ row.status === 1 ? '下架' : '上架' }}
              </el-button>
            </template>
            <span v-else class="readonly-hint">只读</span>
          </template>
        </el-table-column>
      </el-table>

      <TablePager v-model:page="page" v-model:size="size" :total="total" />
    </div>

    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑元素' : '新增元素'" width="460px">
      <el-form label-width="82px">
        <el-form-item label="所属底板" required>
          <el-select v-model="form.board_id" class="w-full">
            <el-option v-for="b in boards" :key="b.id" :label="b.name" :value="b.id" :disabled="b.status !== 1" />
          </el-select>
          <span class="form-tip">元素只在该底板的拼装页出现</span>
        </el-form-item>
        <el-form-item label="分类" required>
          <el-select v-model="form.category">
            <el-option v-for="c in CATEGORIES" :key="c.value" :label="c.label" :value="c.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="名称" required>
          <el-input v-model="form.name" maxlength="32" />
        </el-form-item>
        <el-form-item v-if="!form.id" label="图片">
          <el-upload :show-file-list="false" :auto-upload="false" accept="image/png" :on-change="onFile">
            <el-button>选择 PNG（透明背景）</el-button>
          </el-upload>
          <span v-if="form.image_url" class="uploaded">已上传</span>
        </el-form-item>
        <el-form-item label="锚点X">
          <el-slider v-model="form.ax" :min="0" :max="1" :step="0.01" class="slider" />
        </el-form-item>
        <el-form-item label="锚点Y">
          <el-slider v-model="form.ay" :min="0" :max="1" :step="0.01" class="slider" />
        </el-form-item>
        <el-form-item label="默认倍率">
          <el-input-number v-model="form.scale" :min="0.1" :max="4" :step="0.1" />
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
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import type { UploadFile } from 'element-plus'
import { Plus, Refresh } from '@element-plus/icons-vue'
import PageHeader from '@/components/PageHeader.vue'
import TablePager from '@/components/TablePager.vue'
import { createElement, listBoards, listElements, toggleElement, updateElement, uploadImage } from '@/api/modules/assets'
import { usePager } from '@/utils/pager'
import { useAuthStore } from '@/stores/auth'
import type { Board, ElementItem } from '@/types'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()

const CATEGORIES = [
  { value: 'hair', label: '头发' },
  { value: 'hat', label: '帽子' },
  { value: 'lash', label: '睫毛' },
  { value: 'expression', label: '表情' },
  { value: 'outfit', label: '服饰' },
  { value: 'accessory', label: '配饰' },
]

const boards = ref<Board[]>([])
const boardId = ref<number>(Number(route.query.board_id) || 0)
const catFilter = ref('')
const elements = ref<ElementItem[]>([])
const loading = ref(false)
const saving = ref(false)
const dialogVisible = ref(false)
const form = reactive({
  id: 0,
  board_id: 0,
  category: 'hair',
  name: '',
  image_url: '',
  ax: 0.5,
  ay: 0.5,
  scale: 1,
})

const { page, size, total, rows: pageRows } = usePager(elements, () => [boardId.value, catFilter.value])

const currentBoard = computed(() => boards.value.find((b) => b.id === boardId.value) ?? null)

function catLabel(v: string): string {
  return CATEGORIES.find((c) => c.value === v)?.label ?? v
}

async function loadBoards() {
  loading.value = true
  try {
    boards.value = await listBoards()
    // 默认落在第一块底板上；链接带 board_id 时优先用链接里的
    if (!boards.value.some((b) => b.id === boardId.value)) {
      boardId.value = boards.value[0]?.id ?? 0
    }
    if (boardId.value) router.replace({ query: { ...route.query, board_id: String(boardId.value) } })
    await load()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

async function load() {
  if (!boardId.value) {
    elements.value = []
    return
  }
  loading.value = true
  try {
    elements.value = await listElements(boardId.value, catFilter.value)
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

function onBoardChange(v: number) {
  // 底板写进地址栏，刷新/分享后仍停在同一块底板
  router.replace({ query: { ...route.query, board_id: String(v) } })
  load()
}

function openCreate() {
  if (!boardId.value) return
  form.id = 0
  form.board_id = boardId.value
  form.category = 'hair'
  form.name = ''
  form.image_url = ''
  form.ax = 0.5
  form.ay = 0.5
  form.scale = 1
  dialogVisible.value = true
}

function openEdit(el: ElementItem) {
  form.id = el.id
  form.board_id = el.board_id
  form.category = el.category
  form.name = el.name
  form.image_url = el.image_url
  const anchor = safeAnchor(el.anchor)
  form.ax = anchor.x
  form.ay = anchor.y
  form.scale = anchor.scale
  dialogVisible.value = true
}

function safeAnchor(raw: string): { x: number; y: number; scale: number } {
  try {
    const a = JSON.parse(raw || '{}') as { x?: number; y?: number; scale?: number }
    return { x: a.x ?? 0.5, y: a.y ?? 0.5, scale: a.scale ?? 1 }
  } catch {
    return { x: 0.5, y: 0.5, scale: 1 }
  }
}

async function onFile(f: UploadFile) {
  if (!f.raw) return
  try {
    form.image_url = await uploadImage(f.raw)
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '上传失败')
  }
}

async function save() {
  if (!form.name.trim() || (!form.id && !form.image_url)) {
    ElMessage.warning('请填写名称并上传图片')
    return
  }
  const anchor = JSON.stringify({ x: form.ax, y: form.ay, scale: form.scale })
  const payload = {
    board_id: form.board_id,
    category: form.category,
    name: form.name,
    image_url: form.image_url,
    anchor,
    sort: 0,
  }
  saving.value = true
  try {
    if (form.id) await updateElement(form.id, payload)
    else await createElement(payload)
    ElMessage.success('已保存')
    dialogVisible.value = false
    // 新增/改归属后跳到该底板，保证能看到刚保存的元素
    boardId.value = form.board_id
    await loadBoards()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '保存失败')
  } finally {
    saving.value = false
  }
}

async function toggleStatus(el: ElementItem) {
  try {
    await toggleElement(el.id)
    ElMessage.success('已切换')
    await loadBoards()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '操作失败')
  }
}

onMounted(loadBoards)
</script>

<style scoped>
.w240 {
  width: 240px;
}
.w140 {
  width: 140px;
}
.w-full {
  width: 100%;
}
.preview {
  width: 44px;
  height: 44px;
  border-radius: 6px;
  background: var(--fill-1);
  display: block;
}
.board-meta {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--text-3);
}
.mono {
  font-family: ui-monospace, Menlo, Consolas, monospace;
}
.muted {
  color: var(--text-3);
}
.slider {
  width: 200px;
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
.empty {
  padding: 48px 0;
  text-align: center;
  color: var(--text-3);
  font-size: 13px;
}
.empty p {
  margin: 0 0 12px;
}
</style>
