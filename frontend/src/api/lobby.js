import http from './http'

export const lobbyApi = {
  list(page = 1, limit = 10) {
    return http.get('/api/v1/lobbies', { params: { page, limit } })
  },
  getByCode(code) {
    return http.get(`/api/v1/lobbies/${code}`)
  },
  getActive() {
    return http.get('/api/v1/lobbies/active')
  },
  create(timeControl) {
    return http.post('/api/v1/lobbies', { timeControl })
  },
  join(code) {
    return http.post(`/api/v1/lobbies/${code}/join`)
  },
  leave(code) {
    return http.post(`/api/v1/lobbies/${code}/leave`)
  },
}
