// 后台导航与路由共用一份声明，避免菜单/守卫/权限三处不同步
export interface MenuItem {
  path: string
  title: string
  /** 访问该页需要的权限码 */
  perm: string
}

export const MENU: MenuItem[] = [
  { path: '/orders', title: '订单管理', perm: 'order:view' },
  { path: '/products', title: '商品管理', perm: 'product:view' },
  { path: '/assets', title: '素材管理', perm: 'asset:view' },
  { path: '/users', title: '用户管理', perm: 'user:view' },
  { path: '/works', title: '作品管理', perm: 'work:view' },
  { path: '/configs', title: '系统配置', perm: 'config:view' },
  { path: '/roles', title: '角色权限', perm: 'role:view' },
  { path: '/admins', title: '管理员账号', perm: 'admin:view' },
  { path: '/audit', title: '操作日志', perm: 'audit:view' },
]
