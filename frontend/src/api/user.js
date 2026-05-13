import http from './http'

export const userApi = {
  me() {
    return http.get('/api/v1/users/me')
  },
  getById(id) {
    return http.get(`/api/v1/users/${id}`)
  },
  stats() {
    return http.get('/api/v1/users/stats')
  },
  activeMatch() {
    return http.get('/api/v1/matches/active')
  },
  history(page = 1, limit = 10) {
    return http.get('/api/v1/users/history', { params: { page, limit } })
  },
  adminList({ page = 1, limit = 10, username = '' } = {}) {
    return http.get('/api/v1/admin/users', { params: { page, limit, username } })
  },
  deleteById(id) {
    return http.delete(`/api/v1/admin/users/${id}`)
  },
}
