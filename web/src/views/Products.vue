<template>
  <div>
    <el-card shadow="never" class="card-block">
      <template #header><b>商品分类</b></template>
      <div class="page-toolbar">
        <el-input v-model="newCat" placeholder="新分类名称" class="cat-input" />
        <el-button v-if="auth.hasPerm('product:edit')" type="primary" @click="addCategory">添加分类</el-button>
      </div>
      <el-tag
        v-for="c in categories"
        :key="c.id"
        class="cat-tag"
        :type="c.status === 1 ? 'primary' : 'info'"
        :closable="auth.hasPerm('product:edit')"
        @close="delCategory(c)"
        @click="auth.hasPerm('product:edit') && editCategory(c)"
      >
        {{ c.name }}
      </el-tag>
    </el-card>

    <el-card shadow="never">
      <template #header>
        <div class="flex-between">
          <b>商品列表</b>
          <el-button v-if="auth.hasPerm('product:edit')" type="primary" @click="openProduct(null)">新增商品</el-button>
        </div>
      </template>
      <el-table v-loading="loading" :data="products" border>
        <el-table-column label="商品" min-width="200">
          <template #default="{ row }">
            <div class="p-cell">
              <el-image v-if="firstImg(row.images)" :src="firstImg(row.images)" fit="cover" class="p-thumb" />
              <div>
                <div class="p-title">{{ row.title }}</div>
                <div v-if="row.subtitle" class="p-subtitle">{{ row.subtitle }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="分类" width="100">
          <template #default="{ row }">{{ catName(row.category_id) }}</template>
        </el-table-column>
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'">{{ row.status === 1 ? '上架' : '下架' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="规格/SKU" min-width="320">
          <template #default="{ row }">
            <div v-for="s in row.skus" :key="s.id" class="sku-row">
              <el-tag size="small">{{ s.specs }}</el-tag>
              <span><AmountText :cents="s.price_cents" /> / 库存 {{ s.stock }}</span>
              <template v-if="auth.hasPerm('product:edit')">
                <el-button size="small" link type="primary" @click="openSKU(row as Product, s)">改价/库存</el-button>
                <el-button size="small" link type="danger" @click="delSKU(s)">删除</el-button>
              </template>
            </div>
            <el-button v-if="auth.hasPerm('product:edit')" size="small" @click="openSKU(row as Product, null)">
              + 添加规格
            </el-button>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160">
          <template #default="{ row }">
            <el-button v-if="auth.hasPerm('product:edit')" size="small" @click="openProduct(row as Product)">编辑</el-button>
            <el-button
              v-if="auth.hasPerm('product:edit')"
              size="small"
              type="warning"
              @click="toggleShelf(row as Product)"
            >
              {{ row.status === 1 ? '下架' : '上架' }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="productVisible" :title="productForm.id ? '编辑商品' : '新增商品'" width="640px">
      <el-form label-width="80px">
        <el-form-item label="分类">
          <el-select v-model="productForm.category_id">
            <el-option v-for="c in categories.filter((x) => x.status === 1)" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="名称" required>
          <el-input v-model="productForm.title" maxlength="64" show-word-limit />
        </el-form-item>
        <el-form-item label="副标题">
          <el-input v-model="productForm.subtitle" maxlength="128" show-word-limit placeholder="一句话卖点，如：220g 精梳棉 · 不透不塌" />
        </el-form-item>
        <el-form-item label="划线价">
          <div class="price-row">
            <el-input-number v-model="productForm.originalYuan" :min="0" :precision="2" :step="1" />
            <span class="price-hint">元。仅用于展示优惠（现价取各规格最低价），0 表示不展示</span>
          </div>
        </el-form-item>
        <el-form-item label="主图">
          <div class="img-list">
            <div v-for="(u, i) in productForm.images" :key="u" class="img-item">
              <el-image :src="u" fit="cover" class="img-thumb" />
              <div class="img-ops">
                <el-button size="small" link :disabled="i === 0" @click="moveImg(i, -1)">‹</el-button>
                <el-button size="small" link type="danger" @click="productForm.images.splice(i, 1)">删</el-button>
              </div>
              <el-tag v-if="i === 0" size="small" class="img-first">封面</el-tag>
            </div>
            <el-upload
              :show-file-list="false"
              :auto-upload="false"
              multiple
              accept="image/jpeg,image/png,image/webp"
              :on-change="onImageFile"
            >
              <div class="img-add">＋ 上传图片</div>
            </el-upload>
          </div>
          <div class="img-tip">建议 5 张：正面 / 背面 / 细节 / 场景 / 尺寸；第一张为列表封面</div>
        </el-form-item>
        <el-form-item label="单位/排序">
          <div class="price-row">
            <el-input v-model="productForm.unit" class="unit-input" placeholder="件" maxlength="8" />
            <span class="price-hint">计价单位</span>
            <el-input-number v-model="productForm.sort" :min="0" :step="10" />
            <span class="price-hint">排序权重，大的靠前</span>
          </div>
        </el-form-item>
        <el-form-item label="文字详情">
          <el-input
            v-model="productForm.detail"
            type="textarea"
            :rows="5"
            placeholder="空行分段。常用结构：&#10;【面料】…&#10;【工艺】…&#10;【尺码】…&#10;【注意】屏幕显示与成品存在轻微色差，以实际印制效果为准。"
          />
        </el-form-item>
        <el-form-item label="详情长图">
          <div class="img-list">
            <div v-for="(u, i) in productForm.detailImages" :key="u" class="img-item">
              <el-image :src="u" fit="cover" class="img-thumb tall" />
              <div class="img-ops">
                <el-button size="small" link type="danger" @click="productForm.detailImages.splice(i, 1)">删</el-button>
              </div>
            </div>
            <el-upload
              :show-file-list="false"
              :auto-upload="false"
              multiple
              accept="image/jpeg,image/png,image/webp"
              :on-change="onDetailImageFile"
            >
              <div class="img-add tall">＋ 上传</div>
            </el-upload>
          </div>
          <div class="img-tip">详情页自上而下拼接的说明长图（材质/工艺/尺码表等）</div>
        </el-form-item>
        <el-form-item label="服务承诺">
          <el-input
            v-model="productForm.service_notes"
            maxlength="255"
            placeholder="如：48 小时内发货 · 定制商品不支持七天无理由退货"
          />
        </el-form-item>
        <el-form-item label="简介">
          <el-input v-model="productForm.description" placeholder="一句话简介（列表页兜底文案，可留空）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="productVisible = false">取消</el-button>
        <el-button type="primary" @click="saveProduct">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="skuVisible" :title="skuForm.id ? '改价/库存' : '添加规格'" width="440px">
      <el-form label-width="80px">
        <el-form-item label="规格">
          <el-input v-model="skuForm.specs" placeholder='如：{"尺码":"M","颜色":"白"}' :disabled="!!skuForm.id" />
        </el-form-item>
        <el-form-item label="价格(元)">
          <el-input-number v-model="skuForm.priceYuan" :min="0.01" :precision="2" :step="1" />
        </el-form-item>
        <el-form-item label="库存">
          <el-input-number v-model="skuForm.stock" :min="0" :step="1" />
        </el-form-item>
        <el-form-item label="编码">
          <el-input v-model="skuForm.code" placeholder="可选" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="skuVisible = false">取消</el-button>
        <el-button type="primary" @click="saveSKU">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { UploadFile } from 'element-plus'
import { uploadImage } from '@/api/http'
import AmountText from '@/components/AmountText.vue'
import {
  createCategory,
  createProduct,
  createSku,
  deleteCategory,
  deleteSku,
  listCategories,
  listProducts,
  updateCategory,
  updateProduct,
  updateSku,
} from '@/api/modules/products'
import type { ProductInput } from '@/api/modules/products'
import type { Category, Product, Sku } from '@/types'
import { cents } from '@/utils/money'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()

const categories = ref<Category[]>([])
const products = ref<Product[]>([])
const loading = ref(false)
const newCat = ref('')

const productVisible = ref(false)
const productForm = reactive({
  id: 0,
  category_id: 0,
  title: '',
  subtitle: '',
  description: '',
  images: [] as string[],
  originalYuan: 0,
  unit: '件',
  detail: '',
  detailImages: [] as string[],
  service_notes: '',
  sort: 0,
})

const skuVisible = ref(false)
const skuForm = reactive({ id: 0, productId: 0, specs: '', priceYuan: 9.9, stock: 10, code: '' })

async function load() {
  loading.value = true
  try {
    const [cats, prods] = await Promise.all([listCategories(), listProducts()])
    categories.value = cats
    products.value = prods
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

function catName(id: number): string {
  return categories.value.find((c) => c.id === id)?.name ?? String(id)
}

async function addCategory() {
  if (!newCat.value) return
  try {
    await createCategory(newCat.value)
    newCat.value = ''
    ElMessage.success('分类已添加')
    load()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '添加失败')
  }
}

async function editCategory(c: Category) {
  try {
    const { value } = await ElMessageBox.prompt('修改分类名', '编辑', { inputValue: c.name })
    if (value && value !== c.name) {
      await updateCategory(c.id, value, c.sort)
      load()
    }
  } catch (e) {
    if (e instanceof Error) ElMessage.error(e.message)
  }
}

async function delCategory(c: Category) {
  try {
    await deleteCategory(c.id)
    ElMessage.success('已删除')
    load()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '删除失败')
  }
}

function firstImg(imagesJson: string): string {
  try {
    return (JSON.parse(imagesJson || '[]') as string[])[0] || ''
  } catch {
    return ''
  }
}

function parseImgList(imagesJson: string): string[] {
  try {
    return JSON.parse(imagesJson || '[]') as string[]
  } catch {
    return []
  }
}

function productInput(p: Product, status?: number): ProductInput {
  return {
    category_id: p.category_id,
    title: p.title,
    subtitle: p.subtitle,
    description: p.description,
    images: p.images,
    original_price_cents: p.original_price_cents,
    unit: p.unit,
    detail: p.detail,
    detail_images: p.detail_images,
    service_notes: p.service_notes,
    sort: p.sort,
    status,
  }
}

function openProduct(p: Product | null) {
  productForm.id = p?.id ?? 0
  productForm.category_id = p?.category_id ?? categories.value.find((c) => c.status === 1)?.id ?? 0
  productForm.title = p?.title ?? ''
  productForm.subtitle = p?.subtitle ?? ''
  productForm.description = p?.description ?? ''
  productForm.images = p ? parseImgList(p.images) : []
  productForm.originalYuan = p ? p.original_price_cents / 100 : 0
  productForm.unit = p?.unit || '件'
  productForm.detail = p?.detail ?? ''
  productForm.detailImages = p ? parseImgList(p.detail_images) : []
  productForm.service_notes = p?.service_notes ?? ''
  productForm.sort = p?.sort ?? 0
  productVisible.value = true
}

async function onImageFile(f: UploadFile) {
  if (!f.raw) return
  try {
    productForm.images.push(await uploadImage(f.raw))
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '上传失败')
  }
}

async function onDetailImageFile(f: UploadFile) {
  if (!f.raw) return
  try {
    productForm.detailImages.push(await uploadImage(f.raw))
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '上传失败')
  }
}

function moveImg(i: number, delta: number) {
  const list = productForm.images
  const j = i + delta
  if (j < 0 || j >= list.length) return
  ;[list[i], list[j]] = [list[j], list[i]]
}

async function saveProduct() {
  if (!productForm.title) {
    ElMessage.warning('请填写商品名称')
    return
  }
  if (productForm.originalYuan > 0 && productForm.originalYuan * 100 % 1 !== 0) {
    ElMessage.warning('划线价最多两位小数')
    return
  }
  const input: ProductInput = {
    category_id: productForm.category_id,
    title: productForm.title,
    subtitle: productForm.subtitle,
    description: productForm.description,
    images: JSON.stringify(productForm.images),
    original_price_cents: Math.round(productForm.originalYuan * 100),
    unit: productForm.unit || '件',
    detail: productForm.detail,
    detail_images: JSON.stringify(productForm.detailImages),
    service_notes: productForm.service_notes,
    sort: productForm.sort,
  }
  try {
    if (productForm.id) await updateProduct(productForm.id, input)
    else await createProduct(input)
    ElMessage.success('已保存')
    productVisible.value = false
    load()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '保存失败')
  }
}

async function toggleShelf(p: Product) {
  // 后台更新接口是全量覆盖语义，上下架必须带全所有字段，否则会把新字段清空
  try {
    await updateProduct(p.id, productInput(p, p.status === 1 ? 0 : 1))
    ElMessage.success(p.status === 1 ? '已下架' : '已上架')
    load()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '操作失败')
  }
}

function openSKU(p: Product, s: Sku | null) {
  skuForm.id = s?.id ?? 0
  skuForm.productId = p.id
  skuForm.specs = s?.specs ?? ''
  skuForm.priceYuan = s ? s.price_cents / 100 : 9.9
  skuForm.stock = s?.stock ?? 10
  skuForm.code = s?.code ?? ''
  skuVisible.value = true
}

async function saveSKU() {
  const input = {
    product_id: skuForm.productId,
    specs: skuForm.specs || '{}',
    price_cents: cents(skuForm.priceYuan),
    stock: skuForm.stock,
    code: skuForm.code,
  }
  try {
    if (skuForm.id) await updateSku(skuForm.id, input)
    else await createSku(input)
    ElMessage.success('已保存')
    skuVisible.value = false
    load()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '保存失败')
  }
}

async function delSKU(s: Sku) {
  try {
    await ElMessageBox.confirm(`确认删除规格 ${s.specs}？`, '确认')
  } catch {
    return
  }
  try {
    await deleteSku(s.id)
    ElMessage.success('已删除')
    load()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '删除失败')
  }
}

onMounted(load)
</script>

<style scoped>
.cat-input {
  width: 200px;
}
.cat-tag {
  margin: 0 8px 8px 0;
  cursor: pointer;
}
.sku-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 4px 0;
}
.p-cell {
  display: flex;
  align-items: center;
  gap: 10px;
}
.p-thumb {
  width: 44px;
  height: 44px;
  border-radius: 6px;
  flex-shrink: 0;
  background: var(--el-fill-color-light);
}
.p-title {
  font-weight: 500;
}
.p-subtitle {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
.price-row {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}
.price-hint {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
.unit-input {
  width: 90px;
}
.img-list {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}
.img-item {
  position: relative;
  width: 76px;
}
.img-thumb {
  width: 76px;
  height: 76px;
  border-radius: 6px;
  border: 1px solid var(--el-border-color-lighter);
  display: block;
}
.img-thumb.tall {
  height: 100px;
}
.img-add {
  width: 76px;
  height: 76px;
  border: 1px dashed var(--el-border-color);
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  cursor: pointer;
}
.img-add.tall {
  height: 100px;
}
.img-ops {
  display: flex;
  justify-content: center;
  margin-top: 2px;
}
.img-first {
  position: absolute;
  top: 2px;
  left: 2px;
}
.img-tip {
  width: 100%;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
</style>
