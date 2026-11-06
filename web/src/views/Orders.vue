<template>
  <div>
    <el-select v-model="statusFilter" placeholder="全部状态" clearable style="width: 160px; margin-bottom: 12px" @change="load">
      <el-option label="待接单" value="pending_accept" />
      <el-option label="待发货" value="pending_ship" />
      <el-option label="已发货" value="shipped" />
      <el-option label="已完成" value="done" />
      <el-option label="已取消" value="canceled" />
    </el-select>

    <el-table :data="orders" v-loading="loading" border>
      <el-table-column type="expand">
        <template #default="{ row }">
          <div style="padding: 12px 24px">
            <p><b>明细：</b></p>
            <p v-for="it in row.items" :key="it.id" style="margin: 4px 0">
              {{ it.title }}（{{ it.specs }}）× {{ it.qty }} = {{ yuan(it.price_cents * it.qty) }} 元
              <el-tag v-if="it.work_id" size="small" type="warning">绑定作品 #{{ it.work_id }}</el-tag>
            </p>
            <p v-if="row.shipment"><b>物流：</b>{{ row.shipment.company }} {{ row.shipment.tracking_no }}</p>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="order.order_no" label="单号" width="200" />
      <el-table-column prop="order.user_id" label="用户ID" width="80" />
      <el-table-column label="金额" width="100">
        <template #default="{ row }">{{ yuan(row.order.total_cents) }} 元</template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="statusType(row.order.status)">{{ statusText(row.order.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="退款标记" width="90">
        <template #default="{ row }">
          <el-tag v-if="row.order.refund_flag" type="danger">已标记</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" min-width="240">
        <template #default="{ row }">
          <el-button v-if="row.order.status === 'pending_accept'" type="primary" size="small" @click="act('accept', row)">接单</el-button>
          <el-button v-if="row.order.status === 'pending_ship'" type="success" size="small" @click="openShip(row)">发货</el-button>
          <el-button
            v-if="['pending_accept', 'pending_ship'].includes(row.order.status)"
            type="warning"
            size="small"
            @click="act('cancel', row)"
          >取消</el-button>
          <el-button
            v-if="['shipped', 'done'].includes(row.order.status) && !row.order.refund_flag"
            type="danger"
            size="small"
            @click="act('refund_mark', row)"
          >标记退款</el-button>
        </template>
      </el-table-column>
    </el-table>

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
import { ElMessage, ElMessageBox } from 'element-plus'
import { api, yuan } from '../api'

interface OrderRow {
  order: {
    id: number
    order_no: string
    status: string
    total_cents: number
    user_id: number
    refund_flag: number
    cancel_reason: string
  }
  items: { id: number; title: string; specs: string; qty: number; price_cents: number; work_id: number | null }[]
  shipment: { company: string; tracking_no: string } | null
}

const orders = ref<OrderRow[]>([])
const loading = ref(false)
const statusFilter = ref('')
const shipVisible = ref(false)
const shipForm = reactive({ id: 0, company: '', tracking_no: '' })

async function load() {
  loading.value = true
  try {
    orders.value = await api<OrderRow[]>('GET', '/admin/orders' + (statusFilter.value ? '?status=' + statusFilter.value : ''))
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

function statusText(s: string): string {
  return { pending_accept: '待接单', pending_ship: '待发货', shipped: '已发货', done: '已完成', canceled: '已取消' }[s] ?? s
}
function statusType(s: string): string {
  return { pending_accept: 'warning', pending_ship: 'primary', shipped: 'success', done: 'info', canceled: 'danger' }[s] ?? 'info'
}

async function act(action: string, row: OrderRow) {
  let body: unknown = undefined
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
    await api('POST', `/admin/orders/${row.order.id}/${action}`, body)
    ElMessage.success('操作成功')
    load()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '操作失败')
  }
}

function openShip(row: OrderRow) {
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
    await api('POST', `/admin/orders/${shipForm.id}/ship`, { company: shipForm.company, tracking_no: shipForm.tracking_no })
    ElMessage.success('发货成功')
    shipVisible.value = false
    load()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '发货失败')
  }
}

onMounted(load)
</script>
