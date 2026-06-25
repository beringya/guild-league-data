import { defineStore } from 'pinia'
import { api } from '../api/client'
import type { User } from '../types'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: null as User | null,
    loading: false,
    csrfToken: ''
  }),
  getters: {
    authenticated: (state) => Boolean(state.user)
  },
  actions: {
    async loadMe() {
      this.loading = true
      try {
        const res = await api.get<{ user: User; csrf_token: string }>('/api/auth/me')
        this.user = res.user
        this.csrfToken = res.csrf_token
      } catch {
        this.user = null
      } finally {
        this.loading = false
      }
    },
    async login(username: string, password: string) {
      const res = await api.post<{ user: User; csrf_token: string }>('/api/auth/login', { username, password })
      this.user = res.user
      this.csrfToken = res.csrf_token
    },
    async logout() {
      await api.post('/api/auth/logout')
      this.user = null
      this.csrfToken = ''
    },
    async changePassword(currentPassword: string, newPassword: string) {
      await api.post('/api/auth/change-password', { current_password: currentPassword, new_password: newPassword })
      await this.loadMe()
    }
  }
})
