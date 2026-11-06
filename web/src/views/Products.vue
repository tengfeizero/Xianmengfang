<template>
  <div>
    <el-card shadow="never" class="card-block">
      <template #header><b>商品分类</b></template>
      <div class="page-toolbar">
        <el-input v-model="newCat" placeholder="新分类名称" class="cat-input" />
        <el-button type="primary" @click="addCategory">添加分类</el-button>
      </div>
      <el-tag
        v-for="c in categories"
        :key="c.id"
        class="cat-tag"
        :type="c.status === 1 ? 'primary' : 'info'"
        closable
        @close="delCategory(c)"
        @click="editCategory(c)"
      >
        {{ c.name }}
      </el-tag>
    </el-card>

    <el-card shadow="never">
      <template #header>
        <div class="flex-between">
          <b>商品列表</b>
          <el-button type="primary" @click="openProduct(null)">新增商品</el-button>
        </div>
      </template>
      <el-table v-loading="loading" :data="products" border>
        <el-table-column prop="title" label="商品" min-width="160" />
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
              <el-button size="small" link type="primary" @click="openSKU(row as Product, s)">改价/库存</el-button>
              <el-button size="small" link type="danger" @click="delSKU(s)">删除</el-button>
            </div>
            <el-button size="small" @click="openSKU(row as Product, null)">+ 添加规格</el-button>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160">
          <template #default="{ row }">
            <el-button size="small" @click="openProduct(row as Product)">编辑</el-button>
            <el-button size="small" type="warning" @click="toggleShelf(row as Product)">
              {{ row.status === 1 ? '下架' : '上架' }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="productVisible" :title="productForm.id ? '编辑商品' : '新增商品'" width="440px">
      <el-form label-width="60px">
        <el-form-item label="分类">
          <el-select v-model="productForm.category_id">
            <el-option v-for="c in categories.filter((x) => x.status === 1)" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="名称">
          <el-input v-model="productForm.title" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="productForm.description" type="textarea" />
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
import type { Category, Product, Sku } from '@/types'
import { cents } from '@/utils/money'

const categories = ref<Category[]>([])
const products = ref<Product[]>([])
const loading = ref(false)
const newCat = ref('')

const productVisible = ref(false)
const productForm = reactive({ id: 0, category_id: 0, title: '', description: '' })

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

function openProduct(p: Product | null) {
  productForm.id = p?.id ?? 0
  productForm.category_id = p?.category_id ?? categories.value.find((c) => c.status === 1)?.id ?? 0
  productForm.title = p?.title ?? ''
  productForm.description = p?.description ?? ''
  productVisible.value = true
}

async function saveProduct() {
  if (!productForm.title) {
    ElMessage.warning('请填写商品名称')
    return
  }
  const input = {
    category_id: productForm.category_id,
    title: productForm.title,
    description: productForm.description,
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
  try {
    await updateProduct(p.id, {
      category_id: p.category_id,
      title: p.title,
      description: p.description,
      status: p.status === 1 ? 0 : 1,
    })
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
</style>
