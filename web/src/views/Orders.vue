<template>
  <div>
    <div class="page-toolbar">
      <el-select v-model="statusFilter" placeholder="全部状态" clearable class="filter" @change="load">
        <el-option label="待接单" value="pending_accept" />
        <el-option label="生产中" value="producing" />
        <el-option label="打包中" value="packing" />
        <el-option label="待发货" value="pending_ship" />
        <el-option label="已发货" value="shipped" />
        <el-option label="已完成" value="done" />
        <el-option label="已取消" value="canceled" />
      </el-select>
      <el-input
        v-model="orderNoFilter"
        placeholder="订单号搜索"
        clearable
        class="filter wide"
        @keyup.enter="load"
        @clear="load"
      />
      <el-button :loading="loading" @click="load">刷新</el-button>
    </div>

    <el-table v-loading="loading" :data="orders" border>
      <el-table-column prop="order.order_no" label="单号" width="200" />
      <el-table-column label="买家" width="140">
        <template #default="{ row }">
          <div v-if="row.user" class="buyer">
            <el-avatar :size="24" :src="row.user.avatar">{{ (row.user.nickname || '?')[0] }}</el-avatar>
            <span class="buyer-name">{{ row.user.nickname || `用户#${row.order.user_id}` }}</span>
          </div>
          <span v-else>#{{ row.order.user_id }}</span>
        </template>
      </el-table-column>
      <el-table-column label="金额" width="110">
        <template #default="{ row }"><AmountText :cents="row.order.total_cents" /></template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }"><StatusTag :status="row.order.status" /></template>
      </el-table-column>
      <el-table-column label="下单时间" width="170">
        <template #default="{ row }">{{ fmtTime(row.order.created_at) }}</template>
      </el-table-column>
      <el-table-column label="退款标记" width="90">
        <template #default="{ row }">
          <el-tag v-if="row.order.refund_flag" type="danger">已标记</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" min-width="320">
        <template #default="{ row }">
          <el-button size="small" @click="openDetail(row.order.id)">详情</el-button>
          <el-button v-if="row.order.status === 'pending_accept'" type="primary" size="small" @click="act('accept', row as AdminOrderRow)">
            接单
          </el-button>
          <el-button v-if="row.order.status === 'producing'" type="primary" size="small" @click="act('pack', row as AdminOrderRow)">
            生产完成
          </el-button>
          <el-button v-if="row.order.status === 'packing'" type="primary" size="small" @click="act('ready_ship', row as AdminOrderRow)">
            打包完成
          </el-button>
          <el-button v-if="row.order.status === 'pending_ship'" type="success" size="small" @click="openShip(row as AdminOrderRow)">
            发货
          </el-button>
          <el-button
            v-if="cancellable.includes(row.order.status)"
            type="warning"
            size="small"
            @click="act('cancel', row as AdminOrderRow)"
          >
            取消
          </el-button>
          <el-button
            v-if="['shipped', 'done'].includes(row.order.status) && !row.order.refund_flag"
            type="danger"
            size="small"
            @click="act('refund_mark', row as AdminOrderRow)"
          >
            标记退款
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 订单详情弹窗 -->
    <el-dialog v-model="detailVisible" title="订单详情" width="720px" top="6vh">
      <div v-loading="detailLoading">
        <template v-if="detail">
          <el-descriptions title="基础信息" :column="2" border size="small">
            <el-descriptions-item label="订单号">{{ detail.order.order_no }}</el-descriptions-item>
            <el-descriptions-item label="状态"><StatusTag :status="detail.order.status" /></el-descriptions-item>
            <el-descriptions-item label="买家">
              <span v-if="detail.user">{{ detail.user.nickname }}（ID {{ detail.user.id }}）</span>
              <span v-else>ID {{ detail.order.user_id }}</span>
            </el-descriptions-item>
            <el-descriptions-item label="支付方式">{{ detail.order.pay_type === 'wxpay' ? '微信支付' : '模拟支付' }}</el-descriptions-item>
            <el-descriptions-item label="下单时间">{{ fmtTime(detail.order.created_at) }}</el-descriptions-item>
            <el-descriptions-item label="支付时间">{{ fmtTime(detail.order.paid_at) }}</el-descriptions-item>
            <el-descriptions-item label="收货人">{{ detail.order.receiver_name }}</el-descriptions-item>
            <el-descriptions-item label="电话">{{ detail.order.receiver_phone }}</el-descriptions-item>
            <el-descriptions-item label="收货地址" :span="2">{{ detail.order.receiver_address || '—' }}</el-descriptions-item>
            <el-descriptions-item v-if="detail.order.cancel_reason" label="取消原因" :span="2">
              {{ detail.order.cancel_reason }}
            </el-descriptions-item>
          </el-descriptions>

          <el-descriptions title="金额明细" :column="4" border size="small" class="mt">
            <el-descriptions-item label="商品合计"><AmountText :cents="detail.amounts.items_cents" /></el-descriptions-item>
            <el-descriptions-item label="运费"><AmountText :cents="detail.amounts.shipping_cents" /></el-descriptions-item>
            <el-descriptions-item label="优惠"><AmountText :cents="detail.amounts.discount_cents" /></el-descriptions-item>
            <el-descriptions-item label="实付"><AmountText :cents="detail.amounts.total_cents" /></el-descriptions-item>
          </el-descriptions>

          <p class="section-title">商品明细</p>
          <el-table :data="detail.items" border size="small">
            <el-table-column label="图片" width="72">
              <template #default="{ row }">
                <el-image
                  v-if="row.image_url || row.work_image_url"
                  :src="row.work_image_url || row.image_url"
                  :preview-src-list="[row.work_image_url || row.image_url]"
                  preview-teleported
                  fit="cover"
                  class="thumb"
                />
                <span v-else>—</span>
              </template>
            </el-table-column>
            <el-table-column prop="title" label="商品" min-width="140" />
            <el-table-column prop="specs" label="规格" min-width="110" />
            <el-table-column label="单价" width="90">
              <template #default="{ row }"><AmountText :cents="row.price_cents" /></template>
            </el-table-column>
            <el-table-column prop="qty" label="数量" width="60" />
            <el-table-column label="小计" width="95">
              <template #default="{ row }"><AmountText :cents="row.price_cents * row.qty" /></template>
            </el-table-column>
            <el-table-column label="作品" width="110">
              <template #default="{ row }">
                <el-tag v-if="row.work_id" size="small" type="warning">#{{ row.work_id }}</el-tag>
                <span v-else>—</span>
              </template>
            </el-table-column>
          </el-table>

          <p class="section-title">物流</p>
          <p v-if="detail.shipment" class="plain">
            {{ detail.shipment.company }}　{{ detail.shipment.tracking_no }}
            <span class="muted">（{{ fmtTime(detail.shipment.created_at) }}）</span>
          </p>
          <p v-else class="plain muted">暂无发货记录</p>

          <p class="section-title">状态流水</p>
          <el-timeline v-if="detail.events.length">
            <el-timeline-item
              v-for="ev in detail.events"
              :key="ev.id"
              :timestamp="fmtTime(ev.created_at)"
              placement="top"
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

    <el-dialog v-model="shipVisible" title="发货" width="420px">
      <el-form label-width="80px">
        <el-form-item label="物流公司">
          <el-input v-model="shipForm.company" placeholder="如：顺丰速运" />
        </el-form-item>
        <el-form-item label="运单号">
          <el-input v-model="shipForm.tracking_no" />
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
import { onMounted, reactive, ref } from 'vue'
import AmountText from '@/components/AmountText.vue'
import StatusTag from '@/components/StatusTag.vue'
import { listOrders, getOrderDetail, orderAction, shipOrder, type OrderAction } from '@/api/modules/orders'
import type { AdminOrderDetail, AdminOrderRow, OrderEvent } from '@/types'

const orders = ref<AdminOrderRow[]>([])
const loading = ref(false)
const statusFilter = ref('')
const orderNoFilter = ref('')
const shipVisible = ref(false)
const shipForm = reactive({ id: 0, company: '', tracking_no: '' })

const detailVisible = ref(false)
const detailLoading = ref(false)
const detail = ref<AdminOrderDetail | null>(null)

/** 可取消状态：发货前均可（生产/打包中的库存尚未发出，取消自动回补） */
const cancellable = ['pending_accept', 'producing', 'packing', 'pending_ship']

async function load() {
  loading.value = true
  try {
    orders.value = await listOrders({ status: statusFilter.value, order_no: orderNoFilter.value.trim() })
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
.filter {
  width: 160px;
}
.filter.wide {
  width: 220px;
}
.buyer {
  display: flex;
  align-items: center;
  gap: 6px;
}
.buyer-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.mt {
  margin-top: 16px;
}
.section-title {
  font-weight: 600;
  margin: 18px 0 8px;
}
.plain {
  margin: 4px 0;
}
.muted {
  color: var(--el-text-color-secondary);
}
.thumb {
  width: 48px;
  height: 48px;
  border-radius: 4px;
  display: block;
}
</style>
