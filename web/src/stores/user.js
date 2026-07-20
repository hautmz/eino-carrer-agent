import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { login as loginApi, register as registerApi } from '../api/auth'
import { setToken, setUser, getToken, getUser, clearAuth } from '../utils/auth'

export const useUserStore = defineStore('user', () => {
  const token = ref(getToken() || '')
  const user = ref(getUser() || null)

  const isLoggedIn = computed(() => !!token.value)
  const username = computed(() => user.value?.username || '')

  async function login(usernameVal, password) {
    const res = await loginApi({ username: usernameVal, password })
    const data = res.data
    token.value = data.token
    user.value = { user_id: data.user_id, username: data.username }
    setToken(data.token)
    setUser({ user_id: data.user_id, username: data.username })
    return data
  }

  async function register(usernameVal, password) {
    const res = await registerApi({ username: usernameVal, password })
    const data = res.data
    token.value = data.token
    user.value = { user_id: data.user_id, username: data.username }
    setToken(data.token)
    setUser({ user_id: data.user_id, username: data.username })
    return data
  }

  function logout() {
    token.value = ''
    user.value = null
    clearAuth()
  }

  return { token, user, isLoggedIn, username, login, register, logout }
})
