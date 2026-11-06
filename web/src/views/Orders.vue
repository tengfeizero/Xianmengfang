<template>
  <div>
    <div class="page-toolbar">
      <el-select v-model="statusFilter" placeholder="全部状态" clearable class="filter" @change="load">
        <el-option label="待接单" value="pending_accept" />
        <el-option label="待发货" value="pending_ship" />
        <el-option label="已发货" value="shipped" />
        <el-option label="已完成" value="done" />
        <el-option label="已取消" value="canceled" />
      </el-select>
      <el-button :loading="loading" @click="load">刷新</el-button>
    </div>

    <el-table v-loading="loading" :data="orders" border>
      <el-table-column type="expand">
        <template #default="{ row }">
          <div class="detail">
            <p><b>明细：</b></p>
            <p v-for="it in row.items" :key="it.id" class="line">
              {{ it.title }}（{{ it.specs }}）× {{ it.qty }} =
              <AmountText :cents="it.price_cents * it.qty" />
              <el-tag v-if="it.work_id" size="small" type="warning">绑定作品 #{{ it.work_id }}</el-tag>
            </p>
            <p v-if="row.shipment"><b>物流：</b>{{ row.shipment.company }} {{ row.shipment.tracking_no }}</p>
            <p v-if="row.order.receiver_name">
              <b>收货：</b>{{ row.order.receiver_name }} {{ row.order.receiver_phone }} {{ row.order.receiver_address }}
            </p>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="order.order_no" label="单号" width="200" />
      <el-table-column prop="order.user_id" label="用户ID" width="80" />
      <el-table-column label="金额" width="110">
        <template #default="{ row }"><AmountText :cents="row.order.total_cents" /></template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }"><StatusTag :status="row.order.status" /></template>
      </el-table-column>
      <el-table-column label="退款标记" width="90">
        <template #default="{ row }">
          <el-tag v-if="row.order.refund_flag" type="danger">已标记</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" min-width="240">
        <template #default="{ row }">
          <el-button v-if="row.order.status === 'pending_accept'" type="primary" size="small" @click="act('accept', row as AdminOrderRow)">
            接单
          </el-button>
          <el-button v-if="row.order.status === 'pending_ship'" type="success" size="small" @click="openShip(row as AdminOrderRow)">
            发货
          </el-button>
          <el-button
            v-if="['pending_accept', 'pending_ship'].includes(row.order.status)"
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
import { listOrders, orderAction, shipOrder, type OrderAction } from '@/api/modules/orders'
import type { AdminOrderRow } from '@/types'

const orders = ref<AdminOrderRow[]>([])
const loading = ref(false)
const statusFilter = ref('')
const shipVisible = ref(false)
const shipForm = reactive({ id: 0, company: '', tracking_no: '' })

async function load() {
  loading.value = true
  try {
    orders.value = await listOrders(statusFilter.value)
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
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
.detail {
  padding: 12px 24px;
}
.line {
  margin: 4px 0;
}
</style>
