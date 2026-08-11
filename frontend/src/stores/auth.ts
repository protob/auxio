import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import * as authApi from '@/api/auth'
import { onUnauthorized } from '@/api/client'
import { clearToken, readToken, writeToken } from '@/lib/token'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(readToken())
  const isAuthenticated = computed(() => token.value !== '')

  function clear() {
    clearToken()
    token.value = ''
  }

  async function login(username: string, password: string) {
    const next = await authApi.login(username, password)
    writeToken(next)
    token.value = next
  }

  // Sessions live in server memory, so a restart 401s everything at once. The
  // dashboard treats that as routine and drops back to the login form.
  onUnauthorized(clear)

  return { token, isAuthenticated, login, logout: clear }
})
