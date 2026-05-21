import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { ElMessage } from 'element-plus'
import { mockserverApi } from '@/api'
import type { UserInfo } from '@/types'
import { clearAuthToken, getAuthToken, setAuthToken } from '@/utils/request'

const USER_KEY = 'mockserver_auth_user'

function loadStoredUser(): UserInfo | null {
  const raw = localStorage.getItem(USER_KEY)
  if (!raw) return null
  try {
    const user = JSON.parse(raw) as UserInfo
    return user?.email ? user : null
  } catch {
    return null
  }
}

function storeUser(user: UserInfo | null) {
  if (!user) {
    localStorage.removeItem(USER_KEY)
    return
  }
  localStorage.setItem(USER_KEY, JSON.stringify(user))
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref(getAuthToken())
  const user = ref<UserInfo | null>(loadStoredUser())
  const loading = ref(false)
  const isAuthenticated = computed(() => Boolean(token.value))
  const userEmail = computed(() => user.value?.email || '')

  async function debugLogin(email: string) {
    loading.value = true
    try {
      const result = await mockserverApi.debugLogin({ email })
      token.value = result.token
      user.value = result.user
      setAuthToken(result.token)
      storeUser(result.user)
      ElMessage.success('Logged in')
      return result.user
    } finally {
      loading.value = false
    }
  }

  async function fetchCurrentUser() {
    if (!token.value) return null
    const current = await mockserverApi.currentUser()
    user.value = current
    storeUser(current)
    return current
  }

  function logout() {
    token.value = ''
    user.value = null
    clearAuthToken()
    storeUser(null)
  }

  window.addEventListener('mockserver-auth-cleared', () => {
    token.value = ''
    user.value = null
    storeUser(null)
  })

  return {
    token,
    user,
    loading,
    isAuthenticated,
    userEmail,
    debugLogin,
    fetchCurrentUser,
    logout,
  }
})
