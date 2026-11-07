<template>
  <div class="page">
    <PageHeader title="订单管理" desc="从接单到发货的全流程跟进：生产完成 → 打包完成 → 填写物流发货。">
      <template #actions>
        <el-button :icon="Refresh" :loading="loading" @click="load">刷新</el-button>
      </template>
    </PageHeader>

    <div class="panel">
      <!-- 状态页签：带数量，替代下拉筛选 -->
      <div class="tabs">
        <button
          v-for="t in tabs"
          :key="t.value"
          class="tab"
          :class="{ active: statusFilter === t.value }"
          @click="pick(t.value)"
        >
          {{ t.label }}
          <span class="tab-badge">{{ t.count }}</span>
        </button>
      </div>

      <div class="toolbar">
        <el-input
          v-model="keyword"
          placeholder="搜索订单号 / 买家昵称"
          clearable
          :prefix-icon="Search"
          class="w280"
        />
        <span class="spacer" />
        <span class="count-hint">共 {{ filtered.length }} 条</span>
      </div>

      <el-table v-loading="loading" :data="filtered">
        <el-table-column label="订单号" width="190">
          <template #default="{ row }"><span class="mono">{{ row.order.order_no }}</span></template>
        </el-table-column>
        <el-table-column label="买家" width="160">
          <template #default="{ row }">
            <div v-if="row.user" class="buyer">
              <el-avatar :size="24" :src="row.user.avatar">{{ (row.user.nickname || '?')[0] }}</el-avatar>
              <span class="buyer-name">{{ row.user.nickname || `用户#${row.order.user_id}` }}</span>
            </div>
            <span v-else class="muted">#{{ row.order.user_id }}</span>
          </template>
        </el-table-column>
        <el-table-column label="实付金额" width="120" align="right">
          <template #default="{ row }"><span class="price">¥{{ yuan(row.order.total_cents) }}</span></template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }"><StatusTag :status="row.order.status" /></template>
        </el-table-column>
        <el-table-column label="下单时间" width="150">
          <template #default="{ row }"><span class="muted">{{ fmtTime(row.order.created_at) }}</span></template>
        </el-table-column>
        <el-table-column label="备注" min-width="110">
          <template #default="{ row }">
            <el-tag v-if="row.order.refund_flag" type="danger" size="small" effect="light">退款标记</el-tag>
            <span v-else class="muted">—</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="230" fixed="right" align="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDetail(row.order.id)">详情</el-button>

            <el-button
              v-if="auth.hasPerm('order:process') && row.order.status === 'pending_accept'"
              type="primary"
              size="small"
              @click="act('accept', row as AdminOrderRow)"
            >
              接单
            </el-button>
            <el-button
              v-else-if="auth.hasPerm('order:process') && row.order.status === 'producing'"
              type="primary"
              size="small"
              @click="act('pack', row as AdminOrderRow)"
            >
              生产完成
            </el-button>
            <el-button
              v-else-if="auth.hasPerm('order:process') && row.order.status === 'packing'"
              type="primary"
              size="small"
              @click="act('ready_ship', row as AdminOrderRow)"
            >
              打包完成
            </el-button>
            <el-button
              v-else-if="auth.hasPerm('order:ship') && row.order.status === 'pending_ship'"
              type="success"
              size="small"
              @click="openShip(row as AdminOrderRow)"
            >
              发货
            </el-button>

            <el-dropdown
              v-if="
                (auth.hasPerm('order:cancel') && cancellable.includes(row.order.status)) ||
                (auth.hasPerm('order:refund') && ['shipped', 'done'].includes(row.order.status) && !row.order.refund_flag)
              "
              trigger="click"
              @command="(cmd: string) => act(cmd as OrderAction, row as AdminOrderRow)"
            >
              <el-button link class="more">更多<el-icon><ArrowDown /></el-icon></el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item v-if="auth.hasPerm('order:cancel') && cancellable.includes(row.order.status)" command="cancel">
                    取消订单
                  </el-dropdown-item>
                  <el-dropdown-item
                    v-if="
                      auth.hasPerm('order:refund') &&
                      ['shipped', 'done'].includes(row.order.status) &&
                      !row.order.refund_flag
                    "
                    command="refund_mark"
                  >
                    标记退款
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 订单详情 -->
    <el-dialog v-model="detailVisible" title="订单详情" width="760px" top="6vh">
      <div v-loading="detailLoading" class="detail-body">
        <template v-if="detail">
          <div class="hero">
            <div>
              <StatusTag :status="detail.order.status" />
              <p class="hero-no mono">{{ detail.order.order_no }}</p>
              <p class="hero-time">下单于 {{ fmtTime(detail.order.created_at) }}</p>
            </div>
            <div class="hero-amount">
              <span class="cny">¥</span>{{ yuan(detail.amounts.total_cents) }}
              <em>{{ detail.order.pay_type === 'wxpay' ? '微信支付' : '模拟支付' }}</em>
            </div>
          </div>

          <p class="section">基础信息</p>
          <el-descriptions :column="3" border size="small">
            <el-descriptions-item label="买家">
              <span v-if="detail.user">{{ detail.user.nickname }}（#{{ detail.user.id }}）</span>
              <span v-else>#{{ detail.order.user_id }}</span>
            </el-descriptions-item>
            <el-descriptions-item label="支付时间">{{ fmtTime(detail.order.paid_at) }}</el-descriptions-item>
            <el-descriptions-item label="收货人">{{ detail.order.receiver_name || '—' }}</el-descriptions-item>
            <el-descriptions-item label="联系电话">{{ detail.order.receiver_phone || '—' }}</el-descriptions-item>
            <el-descriptions-item label="收货地址" :span="2">{{ detail.order.receiver_address || '—' }}</el-descriptions-item>
            <el-descriptions-item v-if="detail.order.cancel_reason" label="取消原因" :span="3">
              <span class="danger-text">{{ detail.order.cancel_reason }}</span>
            </el-descriptions-item>
          </el-descriptions>

          <p class="section">金额明细</p>
          <el-descriptions :column="4" border size="small">
            <el-descriptions-item label="商品合计">¥{{ yuan(detail.amounts.items_cents) }}</el-descriptions-item>
            <el-descriptions-item label="运费">¥{{ yuan(detail.amounts.shipping_cents) }}</el-descriptions-item>
            <el-descriptions-item label="优惠">¥{{ yuan(detail.amounts.discount_cents) }}</el-descriptions-item>
            <el-descriptions-item label="实付"><b class="price">¥{{ yuan(detail.amounts.total_cents) }}</b></el-descriptions-item>
          </el-descriptions>

          <p class="section">商品明细</p>
          <el-table :data="detail.items" size="small">
            <el-table-column label="图片" width="66">
              <template #default="{ row }">
                <el-image
                  v-if="row.image_url || row.work_image_url"
                  :src="row.work_image_url || row.image_url"
                  :preview-src-list="[row.work_image_url || row.image_url]"
                  preview-teleported
                  fit="cover"
                  class="thumb"
                />
                <span v-else class="muted">—</span>
              </template>
            </el-table-column>
            <el-table-column prop="title" label="商品" min-width="140" />
            <el-table-column prop="specs" label="规格" min-width="100" />
            <el-table-column label="单价" width="88" align="right">
              <template #default="{ row }">¥{{ yuan(row.price_cents) }}</template>
            </el-table-column>
            <el-table-column prop="qty" label="数量" width="60" align="center" />
            <el-table-column label="小计" width="92" align="right">
              <template #default="{ row }"><b>¥{{ yuan(row.price_cents * row.qty) }}</b></template>
            </el-table-column>
            <el-table-column label="作品" width="86">
              <template #default="{ row }">
                <el-tag v-if="row.work_id" size="small" type="warning" effect="light">#{{ row.work_id }}</el-tag>
                <span v-else class="muted">—</span>
              </template>
            </el-table-column>
          </el-table>

          <p class="section">物流信息</p>
          <p v-if="detail.shipment" class="plain">
            <b>{{ detail.shipment.company }}</b>
            <span class="mono tracking">{{ detail.shipment.tracking_no }}</span>
            <span class="muted">发货于 {{ fmtTime(detail.shipment.created_at) }}</span>
          </p>
          <p v-else class="plain muted">暂无发货记录</p>

          <p class="section">状态流水</p>
          <el-timeline v-if="detail.events.length" class="timeline">
            <el-timeline-item
              v-for="ev in detail.events"
              :key="ev.id"
              :timestamp="fmtTime(ev.created_at)"
              placement="top"
              size="normal"
            >
              {{ eventText(ev) }}
            </el-timeline-item>
          </el-timeline>
          <p v-else class="plain muted">暂无流水</p>
        </template>
      </div>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="shipVisible" title="填写物流信息" width="440px">
      <el-form label-width="82px">
        <el-form-item label="物流公司">
          <el-input v-model="shipForm.company" placeholder="如：顺丰速运" />
        </el-form-item>
        <el-form-item label="运单号">
          <el-input v-model="shipForm.tracking_no" placeholder="快递单号" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="shipVisible = false">取消</el-button>
        <el-button type="primary" @click="doShip">确认发货</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import { ArrowDown, Refresh, Search } from '@element-plus/icons-vue'
import PageHeader from '@/components/PageHeader.vue'
import StatusTag from '@/components/StatusTag.vue'
import { getOrderDetail, listOrders, orderAction, shipOrder, type OrderAction } from '@/api/modules/orders'
import { useAuthStore } from '@/stores/auth'
import { yuan } from '@/utils/money'
import type { AdminOrderDetail, AdminOrderRow, OrderEvent } from '@/types'

const route = useRoute()
const auth = useAuthStore()

const orders = ref<AdminOrderRow[]>([])
const loading = ref(false)
const statusFilter = ref((route.query.status as string) || '')
const keyword = ref('')
const shipVisible = ref(false)
const shipForm = reactive({ id: 0, company: '', tracking_no: '' })

const detailVisible = ref(false)
const detailLoading = ref(false)
const detail = ref<AdminOrderDetail | null>(null)

/** 可取消状态：发货前均可（生产/打包中的库存尚未发出，取消自动回补） */
const cancellable = ['pending_accept', 'producing', 'packing', 'pending_ship']

const STATUS_TABS = [
  { value: '', label: '全部' },
  { value: 'pending_accept', label: '待接单' },
  { value: 'producing', label: '生产中' },
  { value: 'packing', label: '打包中' },
  { value: 'pending_ship', label: '待发货' },
  { value: 'shipped', label: '已发货' },
  { value: 'done', label: '已完成' },
  { value: 'canceled', label: '已取消' },
]

/** 一次拉全量，前端做状态/关键词过滤：页签计数才能准确 */
const tabs = computed(() =>
  STATUS_TABS.map((t) => ({
    ...t,
    count: t.value ? orders.value.filter((o) => o.order.status === t.value).length : orders.value.length,
  })),
)

const filtered = computed(() => {
  const kw = keyword.value.trim().toLowerCase()
  return orders.value.filter((o) => {
    if (statusFilter.value && o.order.status !== statusFilter.value) return false
    if (!kw) return true
    return (
      o.order.order_no.toLowerCase().includes(kw) ||
      (o.user?.nickname ?? '').toLowerCase().includes(kw)
    )
  })
})

function pick(v: string) {
  statusFilter.value = v
}

async function load() {
  loading.value = true
  try {
    orders.value = await listOrders()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

async function openDetail(id: number) {
  detailVisible.value = true
  detailLoading.value = true
  detail.value = null
  try {
    detail.value = await getOrderDetail(id)
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '获取详情失败')
    detailVisible.value = false
  } finally {
    detailLoading.value = false
  }
}

const STATUS_LABEL: Record<string, string> = {
  pending_accept: '待接单',
  producing: '生产中',
  packing: '打包中',
  pending_ship: '待发货',
  shipped: '已发货',
  done: '已完成',
  canceled: '已取消',
}

function eventText(ev: OrderEvent): string {
  const who = ev.operator.startsWith('admin:') ? '管理员' : '用户'
  if (ev.note) return `${who} ${ev.note}`
  return `${who} → ${STATUS_LABEL[ev.to] ?? ev.to}`
}

function fmtTime(v: string | null | undefined): string {
  if (!v) return '—'
  const d = new Date(v)
  if (isNaN(d.getTime())) return '—'
  const p = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}`
}

async function act(action: OrderAction, row: AdminOrderRow) {
  let body: unknown
  if (action === 'cancel') {
    try {
      const { value } = await ElMessageBox.prompt('请输入取消原因', '取消订单', {
        confirmButtonText: '确定',
        cancelButtonText: '返回',
        inputValue: '用户申请取消',
      })
      body = { reason: value || '未填写原因' }
    } catch {
      return
    }
  }
  try {
    await orderAction(row.order.id, action, body)
    ElMessage.success('操作成功')
    load()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '操作失败')
  }
}

function openShip(row: AdminOrderRow) {
  shipForm.id = row.order.id
  shipForm.company = ''
  shipForm.tracking_no = ''
  shipVisible.value = true
}

async function doShip() {
  if (!shipForm.company || !shipForm.tracking_no) {
    ElMessage.warning('请填写物流公司和运单号')
    return
  }
  try {
    await shipOrder(shipForm.id, shipForm.company, shipForm.tracking_no)
    ElMessage.success('发货成功')
    shipVisible.value = false
    load()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '发货失败')
  }
}

onMounted(load)
</script>

<style scoped>
.w280 {
  width: 280px;
}

/* 状态页签 */
.tabs {
  display: flex;
  gap: 4px;
  padding: 10px 18px 0;
  border-bottom: 1px solid var(--line-2);
  overflow-x: auto;
}
.tab {
  position: relative;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 9px 10px 11px;
  border: none;
  background: none;
  color: var(--text-2);
  font-size: 13px;
  font-family: inherit;
  cursor: pointer;
  white-space: nowrap;
  transition: color 0.15s ease;
}
.tab:hover {
  color: var(--text-1);
}
.tab.active {
  color: var(--brand);
  font-weight: 600;
}
.tab.active::after {
  content: '';
  position: absolute;
  left: 10px;
  right: 10px;
  bottom: -1px;
  height: 2px;
  border-radius: 2px 2px 0 0;
  background: var(--brand);
}
.tab-badge {
  min-width: 18px;
  padding: 0 5px;
  height: 17px;
  line-height: 17px;
  text-align: center;
  border-radius: 9px;
  background: var(--fill-1);
  color: var(--text-3);
  font-size: 11px;
  font-weight: 500;
}
.tab.active .tab-badge {
  background: var(--brand-soft);
  color: var(--brand);
}
.count-hint {
  font-size: 12px;
  color: var(--text-3);
}

.buyer {
  display: flex;
  align-items: center;
  gap: 7px;
}
.buyer-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.more {
  color: var(--text-3);
}
.more .el-icon {
  margin-left: 2px;
  font-size: 11px;
}

/* 详情弹窗 */
.detail-body {
  min-height: 120px;
}
.hero {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 4px 0 16px;
  border-bottom: 1px dashed var(--line-1);
}
.hero-no {
  margin: 8px 0 2px;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-1);
}
.hero-time {
  margin: 0;
  font-size: 12px;
  color: var(--text-3);
}
.hero-amount {
  text-align: right;
  font-size: 26px;
  font-weight: 600;
  letter-spacing: -0.5px;
  color: var(--text-1);
  font-variant-numeric: tabular-nums;
}
.hero-amount .cny {
  margin-right: 2px;
  font-size: 16px;
  font-weight: 500;
}
.hero-amount em {
  display: block;
  margin-top: 2px;
  font-size: 11.5px;
  font-style: normal;
  font-weight: 400;
  color: var(--text-3);
}
.section {
  margin: 18px 0 8px;
  padding-left: 9px;
  border-left: 3px solid var(--brand);
  font-size: 13px;
  font-weight: 600;
  color: var(--text-1);
  line-height: 1.2;
}
.plain {
  margin: 4px 0;
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 13px;
}
.tracking {
  padding: 1px 6px;
  border-radius: 4px;
  background: var(--fill-1);
}
.danger-text {
  color: var(--el-color-danger);
}
.thumb {
  width: 42px;
  height: 42px;
  border-radius: 6px;
  display: block;
  background: var(--fill-1);
}
.timeline {
  padding-left: 4px;
}
</style>
