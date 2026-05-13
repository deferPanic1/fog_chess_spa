import http from './http'

export const authApi = {
  login(username, password) {
    return http.post('/api/v1/auth/login', { username, password })
  },
  register(email, username, password) {
    return http.post('/api/v1/auth/register', { email, username, password })
  },
  logout() {
    return http.post('/api/v1/auth/logout')
  },
  me() {
    return http.get('/api/v1/auth/me')
  },
}
