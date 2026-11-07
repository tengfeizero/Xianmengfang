// 后台导航与路由共用一份声明，避免菜单/守卫/权限三处不同步
export interface MenuItem {
  path: string
  title: string
  /** 访问该页需要的权限码；空字符串表示登录即可访问 */
  perm: string
  /** Element Plus 图标组件名（在 layouts/Layout.vue 的 ICONS 表里注册） */
  icon: string
  /** 分组 key，对应 MENU_GROUPS */
  group: string
}

export interface MenuGroup {
  key: string
  /** 分组标题；空字符串表示不显示标题（置顶概览） */
  name: string
}

export const MENU_GROUPS: MenuGroup[] = [
  { key: 'overview', name: '' },
  { key: 'trade', name: '交易' },
  { key: 'goods', name: '商品' },
  { key: 'content', name: '内容' },
  { key: 'customer', name: '客户' },
  { key: 'system', name: '系统' },
]

export const MENU: MenuItem[] = [
  { path: '/dashboard', title: '工作台', perm: '', icon: 'Odometer', group: 'overview' },

  { path: '/orders', title: '订单管理', perm: 'order:view', icon: 'Tickets', group: 'trade' },

  { path: '/products', title: '商品管理', perm: 'product:view', icon: 'Goods', group: 'goods' },

  { path: '/boards', title: '拼装底板', perm: 'asset:view', icon: 'Picture', group: 'content' },
  { path: '/elements', title: '拼装元素', perm: 'asset:view', icon: 'Brush', group: 'content' },
  { path: '/works', title: '用户作品', perm: 'work:view', icon: 'MagicStick', group: 'content' },

  { path: '/users', title: '用户管理', perm: 'user:view', icon: 'User', group: 'customer' },

  { path: '/admins', title: '管理员账号', perm: 'admin:view', icon: 'Avatar', group: 'system' },
  { path: '/roles', title: '角色权限', perm: 'role:view', icon: 'Key', group: 'system' },
  { path: '/audit', title: '操作日志', perm: 'audit:view', icon: 'Document', group: 'system' },
  { path: '/configs', title: '系统配置', perm: 'config:view', icon: 'Setting', group: 'system' },
]
