<template>
  <div class="page">
    <PageHeader title="工作台" desc="订单、商品与内容的全景概览，从这里开始今天的工作。">
      <template #actions>
        <el-button :icon="Refresh" :loading="loading" @click="load">刷新数据</el-button>
      </template>
    </PageHeader>

    <!-- 概览指标 -->
    <div class="stat-grid">
      <div
        v-for="c in cards"
        :key="c.key"
        class="stat-card"
        :class="c.tone"
        :role="c.to ? 'button' : undefined"
        @click="c.to && router.push(c.to)"
      >
        <div class="stat-icon">
          <el-icon><component :is="c.icon" /></el-icon>
        </div>
        <div class="stat-body">
          <span class="stat-label">{{ c.label }}</span>
          <div class="stat-value">
            <span v-if="c.money" class="cny">¥</span>{{ c.value }}<i v-if="c.unit">{{ c.unit }}</i>
          </div>
          <span class="stat-extra">{{ c.extra }}</span>
        </div>
      </div>
    </div>

    <div class="two-col">
      <!-- 订单进度 -->
      <div class="panel">
        <div class="panel-head">
          <div class="panel-title">订单进度</div>
          <el-button v-if="can('order:view')" link type="primary" @click="router.push('/orders')">
            进入订单管理
          </el-button>
        </div>
        <div class="panel-body">
          <div v-if="can('order:view')" class="flow">
            <button v-for="(f, i) in flow" :key="f.status" class="flow-step" @click="goStatus(f.status)">
              <span class="step-index">{{ i + 1 }}</span>
              <span class="step-label">{{ f.label }}</span>
              <b class="step-num">{{ f.num }}</b>
              <span class="step-desc">{{ f.desc }}</span>
            </button>
          </div>
          <p v-else class="empty-hint">当前角色没有订单查看权限。</p>
        </div>
      </div>

      <!-- 快捷入口 -->
      <div class="panel">
        <div class="panel-head">
          <div class="panel-title">快捷入口</div>
        </div>
        <div class="panel-body">
          <div v-if="quick.length" class="quick-grid">
            <button v-for="q in quick" :key="q.to" class="quick-item" @click="router.push(q.to)">
              <el-icon><component :is="q.icon" /></el-icon>
              <span>{{ q.title }}</span>
            </button>
          </div>
          <p v-else class="empty-hint">当前角色为只读账号，暂无快捷操作。</p>
        </div>
      </div>
    </div>

    <!-- 最近订单 -->
    <div v-if="can('order:view')" class="panel">
      <div class="panel-head">
        <div class="panel-title">
          最近订单
          <span class="count">共 {{ orders.length }} 单</span>
        </div>
        <el-button link type="primary" @click="router.push('/orders')">查看全部</el-button>
      </div>
      <el-table v-loading="loading" :data="recent">
        <el-table-column prop="order.order_no" label="订单号" width="200">
          <template #default="{ row }">
            <span class="mono">{{ row.order.order_no }}</span>
          </template>
        </el-table-column>
        <el-table-column label="买家" width="150">
          <template #default="{ row }">
            <div class="buyer">
              <el-avatar :size="22" :src="row.user?.avatar">{{ (row.user?.nickname || '?')[0] }}</el-avatar>
              <span class="buyer-name">{{ row.user?.nickname || `用户#${row.order.user_id}` }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="金额" width="110" align="right">
          <template #default="{ row }"><span class="price">¥{{ yuan(row.order.total_cents) }}</span></template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }"><StatusTag :status="row.order.status" /></template>
        </el-table-column>
        <el-table-column label="下单时间" min-width="160">
          <template #default="{ row }">{{ fmtTime(row.order.created_at) }}</template>
        </el-table-column>
        <el-table-column label="" width="90" align="right">
          <template #default>
            <el-button link type="primary" @click="router.push('/orders')">处理</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, type Component } from 'vue'
import { useRouter } from 'vue-router'
import {
  Document,
  Goods,
  MagicStick,
  Picture,
  Refresh,
  Setting,
  Tickets,
  TrendCharts,
  User,
} from '@element-plus/icons-vue'
import PageHeader from '@/components/PageHeader.vue'
import StatusTag from '@/components/StatusTag.vue'
import { listOrders } from '@/api/modules/orders'
import { listProducts } from '@/api/modules/products'
import { listWorks, type AdminWorkRow } from '@/api/modules/works'
import { listUsers, type AdminUserRow } from '@/api/modules/users'
import { useAuthStore } from '@/stores/auth'
import { yuan } from '@/utils/money'
import type { AdminOrderRow, Product } from '@/types'

const router = useRouter()
const auth = useAuthStore()
const can = (p: string) => auth.hasPerm(p)

const loading = ref(false)
const orders = ref<AdminOrderRow[]>([])
const products = ref<Product[]>([])
const works = ref<AdminWorkRow[]>([])
const users = ref<AdminUserRow[]>([])

const PENDING: string[] = ['pending_accept', 'producing', 'packing', 'pending_ship']

function byStatus(s: string): number {
  return orders.value.filter((r) => r.order.status === s).length
}

const pendingCount = computed(() => orders.value.filter((r) => PENDING.includes(r.order.status)).length)
const todayOrders = computed(() =>
  orders.value.filter((r) => isToday(r.order.created_at) && r.order.status !== 'canceled'),
)
const todayAmount = computed(() => todayOrders.value.reduce((s, r) => s + r.order.total_cents, 0))
const onSaleCount = computed(() => products.value.filter((p) => p.status === 1).length)
const workCount = computed(() => works.value.filter((w) => w.status === 1).length)

interface StatCard {
  key: string
  label: string
  value: number | string
  unit?: string
  money?: boolean
  extra: string
  tone: 'red' | 'blue' | 'violet' | 'green'
  icon: Component
  to?: string
}

const cards = computed<StatCard[]>(() => {
  const out: StatCard[] = []
  if (can('order:view')) {
    out.push({
      key: 'pending',
      label: '待处理订单',
      value: pendingCount.value,
      unit: ' 单',
      extra: `待接单 ${byStatus('pending_accept')} · 待发货 ${byStatus('pending_ship')}`,
      tone: 'red',
      icon: Tickets,
      to: '/orders',
    })
    out.push({
      key: 'today',
      label: '今日成交',
      value: yuan(todayAmount.value),
      money: true,
      extra: `今日 ${todayOrders.value.length} 笔订单`,
      tone: 'blue',
      icon: TrendCharts,
      to: '/orders',
    })
  }
  if (can('product:view')) {
    out.push({
      key: 'products',
      label: '在售商品',
      value: onSaleCount.value,
      unit: ' 款',
      extra: `共 ${products.value.length} 款，含下架`,
      tone: 'violet',
      icon: Goods,
      to: '/products',
    })
  }
  if (can('work:view')) {
    out.push({
      key: 'works',
      label: '用户作品',
      value: workCount.value,
      unit: ' 件',
      extra: `共 ${works.value.length} 件，含已下架`,
      tone: 'green',
      icon: MagicStick,
      to: '/works',
    })
  }
  if (can('user:view')) {
    out.push({
      key: 'users',
      label: '用户总数',
      value: users.value.length,
      unit: ' 人',
      extra: '来自小程序端注册',
      tone: 'blue',
      icon: User,
      to: '/users',
    })
  }
  return out
})

const flow = computed(() => [
  { status: 'pending_accept', label: '待接单', num: byStatus('pending_accept'), desc: '等待确认' },
  { status: 'producing', label: '生产中', num: byStatus('producing'), desc: '工厂制作' },
  { status: 'packing', label: '打包中', num: byStatus('packing'), desc: '等待打包' },
  { status: 'pending_ship', label: '待发货', num: byStatus('pending_ship'), desc: '待填物流' },
])

const quick = computed(() => {
  const list: { title: string; to: string; icon: Component }[] = []
  if (can('product:edit')) list.push({ title: '新增商品', to: '/products', icon: Goods })
  if (can('asset:view')) list.push({ title: '拼装素材', to: '/assets', icon: Picture })
  if (can('work:view')) list.push({ title: '用户作品', to: '/works', icon: MagicStick })
  if (can('config:edit')) list.push({ title: 'AI 配置', to: '/configs', icon: Setting })
  if (can('audit:view')) list.push({ title: '操作日志', to: '/audit', icon: Document })
  if (can('user:view')) list.push({ title: '用户管理', to: '/users', icon: User })
  return list
})

const recent = computed(() => orders.value.slice(0, 8))

async function load() {
  loading.value = true
  const jobs: Promise<unknown>[] = []
  if (can('order:view')) jobs.push(listOrders().then((d) => (orders.value = d)))
  if (can('product:view')) jobs.push(listProducts().then((d) => (products.value = d)))
  if (can('work:view')) jobs.push(listWorks().then((d) => (works.value = d)))
  if (can('user:view')) jobs.push(listUsers().then((d) => (users.value = d)))
  // 并行加载，单个接口失败不影响其他区块
  await Promise.allSettled(jobs)
  loading.value = false
}

function goStatus(status: string) {
  router.push({ path: '/orders', query: { status } })
}

function isToday(v: string): boolean {
  if (!v) return false
  const d = new Date(v)
  const n = new Date()
  return d.getFullYear() === n.getFullYear() && d.getMonth() === n.getMonth() && d.getDate() === n.getDate()
}

function fmtTime(v: string | null | undefined): string {
  if (!v) return '—'
  const d = new Date(v)
  if (isNaN(d.getTime())) return '—'
  const p = (n: number) => String(n).padStart(2, '0')
  return `${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}`
}

onMounted(load)
</script>

<style scoped>
.stat-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(210px, 1fr));
  gap: 14px;
  margin-bottom: 16px;
}
.stat-card {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  padding: 18px;
  background: var(--panel-bg);
  border: 1px solid var(--line-1);
  border-radius: var(--radius);
  box-shadow: var(--shadow-panel);
  transition: transform 0.16s ease, box-shadow 0.16s ease;
}
.stat-card[role='button'] {
  cursor: pointer;
}
.stat-card[role='button']:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(16, 24, 40, 0.08);
}
.stat-icon {
  display: grid;
  place-items: center;
  width: 40px;
  height: 40px;
  border-radius: 10px;
  font-size: 19px;
  flex-shrink: 0;
}
.stat-card.red .stat-icon {
  color: var(--brand);
  background: var(--brand-soft);
}
.stat-card.blue .stat-icon {
  color: #2b6ef2;
  background: #eef3ff;
}
.stat-card.violet .stat-icon {
  color: #7c5cf0;
  background: #f3efff;
}
.stat-card.green .stat-icon {
  color: #12a150;
  background: #e9f8ef;
}
.stat-body {
  display: flex;
  flex-direction: column;
  min-width: 0;
}
.stat-label {
  font-size: 12.5px;
  color: var(--text-3);
}
.stat-value {
  margin-top: 2px;
  font-size: 25px;
  font-weight: 600;
  line-height: 1.25;
  letter-spacing: -0.5px;
  color: var(--text-1);
  font-variant-numeric: tabular-nums;
}
.stat-value .cny {
  margin-right: 2px;
  font-size: 15px;
  font-weight: 500;
}
.stat-value i {
  font-size: 12px;
  font-style: normal;
  font-weight: 400;
  color: var(--text-3);
}
.stat-extra {
  margin-top: 3px;
  font-size: 11.5px;
  color: var(--text-4);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.two-col {
  display: grid;
  grid-template-columns: minmax(0, 1.35fr) minmax(0, 1fr);
  gap: 16px;
  margin-bottom: 16px;
}
@media (max-width: 1100px) {
  .two-col {
    grid-template-columns: minmax(0, 1fr);
  }
}

/* 订单进度：四步流水 */
.flow {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 10px;
}
.flow-step {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 2px;
  padding: 14px 14px 12px;
  text-align: left;
  background: var(--fill-1);
  border: 1px solid var(--line-2);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: border-color 0.15s ease, background-color 0.15s ease;
  font: inherit;
}
.flow-step:hover {
  background: #fff;
  border-color: var(--brand-line);
}
.step-index {
  position: absolute;
  top: 10px;
  right: 12px;
  font-size: 11px;
  color: var(--text-4);
}
.step-label {
  font-size: 12px;
  color: var(--text-3);
}
.step-num {
  font-size: 22px;
  font-weight: 600;
  line-height: 1.3;
  color: var(--text-1);
  font-variant-numeric: tabular-nums;
}
.step-desc {
  font-size: 11.5px;
  color: var(--text-4);
}

/* 快捷入口 */
.quick-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}
.quick-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 7px;
  padding: 16px 8px;
  background: var(--fill-1);
  border: 1px solid transparent;
  border-radius: var(--radius-sm);
  color: var(--text-2);
  font-size: 12.5px;
  cursor: pointer;
  transition: all 0.15s ease;
}
.quick-item .el-icon {
  font-size: 19px;
  color: var(--text-3);
  transition: color 0.15s ease;
}
.quick-item:hover {
  background: var(--brand-soft);
  border-color: var(--brand-line);
  color: var(--brand);
}
.quick-item:hover .el-icon {
  color: var(--brand);
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
.empty-hint {
  margin: 6px 0;
  font-size: 12.5px;
  color: var(--text-4);
}
</style>
