import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { clearToken, getToken, setToken } from '@/utils/auth'
import { getMe } from '@/api/modules/rbac'
import { MENU } from '@/config/menu'
import type { AdminMe } from '@/types'

/** 登录态唯一来源：token + 当前管理员与权限。
 *  权限每次启动从后端拉取、不做本地缓存，避免改权限后前端残留旧权限。 */
export const useAuthStore = defineStore('auth', () => {
  const token = ref(getToken())
  const me = ref<AdminMe | null>(null)
  const isLoggedIn = computed(() => !!token.value)

  let loading: Promise<void> | null = null

  function login(t: string) {
    token.value = t
    setToken(t)
    me.value = null
    loading = null
  }

  function logout() {
    token.value = ''
    me.value = null
    loading = null
    clearToken()
  }

  /** 拉取当前账号与权限；并发调用复用同一请求 */
  async function loadMe(force = false): Promise<void> {
    if (me.value && !force) return
    if (!token.value) throw new Error('未登录')
    if (!loading) {
      loading = getMe()
        .then((data) => {
          me.value = data
        })
        .finally(() => {
          loading = null
        })
    }
    await loading
  }

  /** 权限判定：超级管理员返回 ["*"]；前端仅用于隐藏入口，真正拦截在后端 */
  function hasPerm(code: string): boolean {
    const perms = me.value?.perms ?? []
    return perms.includes('*') || perms.includes(code)
  }

  const menus = computed(() => MENU.filter((m) => hasPerm(m.perm)))

  /** 无权访问时的落地页：第一个有权限的菜单 */
  const firstPath = computed(() => menus.value[0]?.path ?? '/login')

  return { token, me, isLoggedIn, menus, firstPath, login, logout, loadMe, hasPerm }
})
