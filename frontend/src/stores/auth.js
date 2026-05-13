import { defineStore } from 'pinia'
import { authApi } from '@/api/auth'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: null,
    errors: {},
  }),

  getters: {
    isAuthenticated: (state) => !!state.user,
    username: (state) => state.user?.username ?? '',
  },

  actions: {
    clearErrors() {
      this.errors = {}
    },

    async loginAction({ username, password }) {
      this.clearErrors()
      try {
        const { data } = await authApi.login(username, password)
        this.user = data.user
        if (data.accessToken) {
          localStorage.setItem('token', data.accessToken)
        }
      } catch (err) {
        this._handleError(err)
        throw err
      }
    },

    async registerAction({ username, email, password }) {
      this.clearErrors()
      try {
        await authApi.register(email, username, password)
        await this.loginAction({ username, password })
      } catch (err) {
        this._handleError(err)
        throw err
      }
    },

    async logout() {
      try {
        await authApi.logout()
      } finally {
        this.user = null
        this.errors = {}
        localStorage.removeItem('token')
      }
    },

    async loadFromSession() {
      try {
        const { data } = await authApi.me()
        this.user = data.user
      } catch (err) {
        this.user = null
        if (err.response?.status === 401) {
          localStorage.removeItem('token')
        }
      }
    },

    _handleError(err) {
      const status = err.response?.status
      const message = err.response?.data?.error

      switch (status) {
        case 400:
          this.errors = { General: message ?? 'Invalid request' }
          break
        case 401:
          this.errors = { General: 'Invalid username or password' }
          break
        case 409:
          if (message?.includes('email')) {
            this.errors = { Email: message }
          } else {
            this.errors = { Username: message }
          }
          break
        default:
          this.errors = { General: 'Something went wrong. Try again.' }
      }
    },
  },
})
