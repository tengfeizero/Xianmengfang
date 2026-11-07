import { createRouter, createWebHashHistory, type RouteRecordRaw } from 'vue-router'
import { getToken } from '@/utils/auth'
import { MENU } from '@/config/menu'
import { useAuthStore } from '@/stores/auth'

/** 菜单路径 → 页面组件（与 config/menu.ts 一一对应） */
const PAGES: Record<string, () => Promise<unknown>> = {
  '/dashboard': () => import('@/views/Dashboard.vue'),
  '/orders': () => import('@/views/Orders.vue'),
  '/products': () => import('@/views/Products.vue'),
  '/boards': () => import('@/views/Boards.vue'),
  '/elements': () => import('@/views/Elements.vue'),
  '/users': () => import('@/views/Users.vue'),
  '/works': () => import('@/views/Works.vue'),
  '/configs': () => import('@/views/Configs.vue'),
  '/roles': () => import('@/views/Roles.vue'),
  '/admins': () => import('@/views/Admins.vue'),
  '/audit': () => import('@/views/AuditLogs.vue'),
}

const routes: RouteRecordRaw[] = [
  { path: '/login', name: 'login', component: () => import('@/views/Login.vue'), meta: { title: '登录' } },
  {
    path: '/',
    component: () => import('@/layouts/Layout.vue'),
    redirect: '/dashboard',
    children: MENU.map((m) => ({
      path: m.path.replace(/^\//, ''),
      name: m.path.replace(/^\//, ''),
      component: PAGES[m.path] as RouteRecordRaw['component'],
      meta: { title: m.title, perm: m.perm },
    })) as RouteRecordRaw[],
  },
]

const router = createRouter({
  history: createWebHashHistory('/ui/'),
  routes,
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()

  if (to.path === '/login') {
    if (getToken()) {
      try {
        await auth.loadMe()
        return auth.firstPath
      } catch {
        auth.logout()
      }
    }
    return true
  }

  if (!getToken()) return { path: '/login' }

  // 每次启动拉一次权限；失败（401/账号被禁用/角色停用）直接回登录
  try {
    await auth.loadMe()
  } catch {
    auth.logout()
    return { path: '/login' }
  }

  const perm = to.meta.perm as string | undefined
  if (perm && !auth.hasPerm(perm)) {
    if (auth.firstPath === '/login') {
      auth.logout()
      return { path: '/login' }
    }
    return auth.firstPath
  }
  return true
})

router.afterEach((to) => {
  document.title = to.meta.title ? `${String(to.meta.title)} · 现梦坊` : '现梦坊管理后台'
})

export default router
