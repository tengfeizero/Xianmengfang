import { createRouter, createWebHashHistory, type RouteRecordRaw } from 'vue-router'
import { getToken } from '@/utils/auth'

const routes: RouteRecordRaw[] = [
  { path: '/login', name: 'login', component: () => import('@/views/Login.vue'), meta: { title: '登录' } },
  {
    path: '/',
    component: () => import('@/layouts/Layout.vue'),
    redirect: '/orders',
    children: [
      { path: 'orders', name: 'orders', component: () => import('@/views/Orders.vue'), meta: { title: '订单管理' } },
      {
        path: 'products',
        name: 'products',
        component: () => import('@/views/Products.vue'),
        meta: { title: '商品管理' },
      },
      { path: 'assets', name: 'assets', component: () => import('@/views/Assets.vue'), meta: { title: '素材管理' } },
      { path: 'users', name: 'users', component: () => import('@/views/Users.vue'), meta: { title: '用户管理' } },
      {
        path: 'configs',
        name: 'configs',
        component: () => import('@/views/Configs.vue'),
        meta: { title: '系统配置' },
      },
    ],
  },
]

const router = createRouter({
  history: createWebHashHistory('/ui/'),
  routes,
})

router.beforeEach((to) => {
  if (to.path !== '/login' && !getToken()) return { path: '/login' }
  if (to.path === '/login' && getToken()) return { path: '/orders' }
  return true
})

router.afterEach((to) => {
  document.title = to.meta.title ? `${String(to.meta.title)} · 现梦坊` : '现梦坊管理后台'
})

export default router
