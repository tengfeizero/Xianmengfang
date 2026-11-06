import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { clearToken, getToken, setToken } from '@/utils/auth'

/** 登录态唯一来源：组件读 store，不直接碰 localStorage */
export const useAuthStore = defineStore('auth', () => {
  const token = ref(getToken())
  const isLoggedIn = computed(() => !!token.value)

  function login(t: string) {
    token.value = t
    setToken(t)
  }

  function logout() {
    token.value = ''
    clearToken()
  }

  return { token, isLoggedIn, login, logout }
})
