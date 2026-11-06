<template>
  <div>
    <el-card shadow="never" class="card-block">
      <template #header>
        <div class="flex-between">
          <b>拼装底板</b>
          <el-button type="primary" @click="openBoard">新增底板</el-button>
        </div>
      </template>
      <el-table :data="boards" border>
        <el-table-column label="预览" width="80">
          <template #default="{ row }">
            <el-image :src="row.url" class="preview" fit="contain" />
          </template>
        </el-table-column>
        <el-table-column prop="name" label="名称" width="160" />
        <el-table-column prop="url" label="图片" min-width="200" />
        <el-table-column label="尺寸" width="110">
          <template #default="{ row }">{{ row.width }}×{{ row.height }}</template>
        </el-table-column>
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'">{{ row.status === 1 ? '上架' : '下架' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="90">
          <template #default="{ row }">
            <el-button size="small" type="warning" @click="onToggleBoard(row as Board)">
              {{ row.status === 1 ? '下架' : '上架' }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-card shadow="never">
      <template #header>
        <div class="flex-between">
          <b>拼装元素</b>
          <div class="page-toolbar">
            <el-select v-model="catFilter" placeholder="全部分类" clearable class="filter" @change="loadElements">
              <el-option v-for="c in CATEGORIES" :key="c.value" :label="c.label" :value="c.value" />
            </el-select>
            <el-button type="primary" @click="openElement()">新增元素</el-button>
          </div>
        </div>
      </template>
      <el-table :data="elements" border>
        <el-table-column label="预览" width="80">
          <template #default="{ row }">
            <el-image :src="row.image_url" class="preview square" fit="contain" />
          </template>
        </el-table-column>
        <el-table-column prop="name" label="名称" width="140" />
        <el-table-column label="分类" width="110">
          <template #default="{ row }">{{ catLabel(row.category) }}</template>
        </el-table-column>
        <el-table-column prop="anchor" label="默认锚点" min-width="160" />
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'">{{ row.status === 1 ? '上架' : '下架' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="140">
          <template #default="{ row }">
            <el-button size="small" @click="openElement(row as ElementItem)">编辑</el-button>
            <el-button size="small" type="warning" @click="onToggleElement(row as ElementItem)">
              {{ row.status === 1 ? '下架' : '上架' }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="boardVisible" title="新增底板" width="440px">
      <el-form label-width="70px">
        <el-form-item label="名称"><el-input v-model="boardForm.name" /></el-form-item>
        <el-form-item label="图片">
          <el-upload :show-file-list="false" :auto-upload="false" accept="image/png" :on-change="onBoardFile">
            <el-button>选择 PNG（透明背景）</el-button>
          </el-upload>
          <span v-if="boardForm.url" class="uploaded">已上传</span>
        </el-form-item>
        <el-form-item label="宽"><el-input-number v-model="boardForm.width" :min="100" /></el-form-item>
        <el-form-item label="高"><el-input-number v-model="boardForm.height" :min="100" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="boardVisible = false">取消</el-button>
        <el-button type="primary" @click="saveBoard">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="elementVisible" :title="elementForm.id ? '编辑元素' : '新增元素'" width="440px">
      <el-form label-width="80px">
        <el-form-item label="分类">
          <el-select v-model="elementForm.category" :disabled="!!elementForm.id">
            <el-option v-for="c in CATEGORIES" :key="c.value" :label="c.label" :value="c.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="名称"><el-input v-model="elementForm.name" /></el-form-item>
        <el-form-item v-if="!elementForm.id" label="图片">
          <el-upload :show-file-list="false" :auto-upload="false" accept="image/png" :on-change="onElementFile">
            <el-button>选择 PNG（透明背景）</el-button>
          </el-upload>
          <span v-if="elementForm.image_url" class="uploaded">已上传</span>
        </el-form-item>
        <el-form-item label="锚点X">
          <el-slider v-model="elementForm.ax" :min="0" :max="1" :step="0.01" class="slider" />
        </el-form-item>
        <el-form-item label="锚点Y">
          <el-slider v-model="elementForm.ay" :min="0" :max="1" :step="0.01" class="slider" />
        </el-form-item>
        <el-form-item label="默认倍率">
          <el-input-number v-model="elementForm.scale" :min="0.1" :max="4" :step="0.1" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="elementVisible = false">取消</el-button>
        <el-button type="primary" @click="saveElement">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import type { UploadFile } from 'element-plus'
import {
  createBoard,
  createElement,
  listBoards,
  listElements,
  toggleBoard,
  toggleElement,
  updateElement,
  uploadImage,
} from '@/api/modules/assets'
import type { Board, ElementItem } from '@/types'

const CATEGORIES = [
  { value: 'hair', label: '头发' },
  { value: 'hat', label: '帽子' },
  { value: 'lash', label: '睫毛' },
  { value: 'expression', label: '表情' },
  { value: 'outfit', label: '服饰' },
  { value: 'accessory', label: '配饰' },
]

const boards = ref<Board[]>([])
const elements = ref<ElementItem[]>([])
const catFilter = ref('')

const boardVisible = ref(false)
const boardForm = reactive({ name: '', url: '', width: 400, height: 500 })

const elementVisible = ref(false)
const elementForm = reactive({ id: 0, category: 'hair', name: '', image_url: '', ax: 0.5, ay: 0.5, scale: 1 })

function catLabel(v: string): string {
  return CATEGORIES.find((c) => c.value === v)?.label ?? v
}

async function load() {
  try {
    boards.value = await listBoards()
    await loadElements()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '加载失败')
  }
}

async function loadElements() {
  try {
    elements.value = await listElements(catFilter.value)
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '加载失败')
  }
}

function openBoard() {
  boardForm.name = ''
  boardForm.url = ''
  boardForm.width = 400
  boardForm.height = 500
  boardVisible.value = true
}

async function onBoardFile(f: UploadFile) {
  if (!f.raw) return
  try {
    boardForm.url = await uploadImage(f.raw)
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '上传失败')
  }
}

async function saveBoard() {
  if (!boardForm.name || !boardForm.url) {
    ElMessage.warning('请填写名称并上传图片')
    return
  }
  try {
    await createBoard({ ...boardForm })
    ElMessage.success('底板已添加')
    boardVisible.value = false
    load()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '保存失败')
  }
}

async function onToggleBoard(b: Board) {
  await toggleBoard(b.id)
  ElMessage.success('已切换')
  load()
}

function openElement(el?: ElementItem | null) {
  elementForm.id = el?.id ?? 0
  elementForm.category = el?.category ?? 'hair'
  elementForm.name = el?.name ?? ''
  elementForm.image_url = ''
  const anchor = el ? (JSON.parse(el.anchor) as { x: number; y: number; scale: number }) : { x: 0.5, y: 0.5, scale: 1 }
  elementForm.ax = anchor.x
  elementForm.ay = anchor.y
  elementForm.scale = anchor.scale
  elementVisible.value = true
}

async function onElementFile(f: UploadFile) {
  if (!f.raw) return
  try {
    elementForm.image_url = await uploadImage(f.raw)
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '上传失败')
  }
}

async function saveElement() {
  if (!elementForm.name || (!elementForm.id && !elementForm.image_url)) {
    ElMessage.warning('请填写名称并上传图片')
    return
  }
  const anchor = JSON.stringify({ x: elementForm.ax, y: elementForm.ay, scale: elementForm.scale })
  const base = { category: elementForm.category, name: elementForm.name, anchor, sort: 0 }
  try {
    if (elementForm.id) {
      const imageUrl = elements.value.find((x) => x.id === elementForm.id)?.image_url ?? ''
      await updateElement(elementForm.id, { ...base, image_url: imageUrl })
    } else {
      await createElement({ ...base, image_url: elementForm.image_url })
    }
    ElMessage.success('已保存')
    elementVisible.value = false
    loadElements()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '保存失败')
  }
}

async function onToggleElement(el: ElementItem) {
  try {
    await toggleElement(el.id)
    ElMessage.success('已切换')
    loadElements()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '操作失败')
  }
}

onMounted(load)
</script>

<style scoped>
.preview {
  width: 48px;
  height: 60px;
}
.preview.square {
  height: 48px;
}
.filter {
  width: 140px;
}
.slider {
  width: 200px;
}
.uploaded {
  margin-left: 8px;
  color: #67c23a;
}
</style>
