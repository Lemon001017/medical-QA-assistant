import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '../api'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const user = ref(JSON.parse(localStorage.getItem('user') || 'null'))

  const isAuthenticated = computed(() => !!token.value)

  function setAuth(authToken, userData) {
    token.value = authToken
    user.value = userData
    localStorage.setItem('token', authToken)
    localStorage.setItem('user', JSON.stringify(userData))
  }

  function clearAuth() {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  async function register(username, email, password) {
    try {
      const response = await api.post('/auth/register', {
        username,
        email,
        password
      })
      setAuth(response.data.token, response.data.user)
      return response.data
    } catch (error) {
      throw error.response?.data?.error || 'Registration failed'
    }
  }

  async function login(username, password) {
    try {
      const response = await api.post('/auth/login', {
        username,
        password
      })
      setAuth(response.data.token, response.data.user)
      return response.data
    } catch (error) {
      throw error.response?.data?.error || 'Login failed'
    }
  }

  function logout() {
    clearAuth()
  }

  return {
    token,
    user,
    isAuthenticated,
    register,
    login,
    logout
  }
})
