import { createRouter, createWebHashHistory } from 'vue-router'

const router = createRouter({
  history: createWebHashHistory('/ui/'),
  routes: [
    { path: '/login', component: () => import('./views/Login.vue') },
    {
      path: '/',
      component: () => import('./views/Layout.vue'),
      redirect: '/orders',
      children: [
        { path: 'orders', component: () => import('./views/Orders.vue'), meta: { title: '订单管理' } },
        { path: 'products', component: () => import('./views/Products.vue'), meta: { title: '商品管理' } },
        { path: 'assets', component: () => import('./views/Assets.vue'), meta: { title: '素材管理' } },
        { path: 'configs', component: () => import('./views/Configs.vue'), meta: { title: '系统配置' } },
      ],
    },
  ],
})

router.beforeEach((to) => {
  if (to.path !== '/login' && !localStorage.getItem('xmf_token')) {
    return '/login'
  }
})

export default router
