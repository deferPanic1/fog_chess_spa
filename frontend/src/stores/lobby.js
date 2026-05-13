import { defineStore } from 'pinia'
import { lobbyApi } from '@/api/lobby'
import { useAuthStore } from './auth'

export const useLobbyStore = defineStore('lobby', {
  state: () => ({
    lobbies: [],
    pagination: {
      page: 1,
      limit: 10,
      total: 0,
      totalPages: 0,
      hasPrev: false,
      hasNext: false,
    },
    currentLobby: null,
    loading: false,
    error: null,
    ws: null,
    lobbyClosed: false,
    matchStart: null,
    startedMatch: null,
  }),

  getters: {
    isHost: (state) => {
      const auth = useAuthStore()
      return state.currentLobby?.hostUsername === auth.username
    },
    isGuest: (state) => {
      const auth = useAuthStore()
      return state.currentLobby?.guestUsername === auth.username
    },
    isParticipant: (getters) => getters.isHost || getters.isGuest,
    statusClass: (state) => (status) => {
      const s = status || state.currentLobby?.status
      return (
        {
          waiting: 'text-gruvbox-green',
          full: 'text-gruvbox-yellow',
          ready: 'text-gruvbox-orange',
          closed: 'text-gruvbox-red',
        }[s] ?? 'text-gruvbox-fg4'
      )
    },
  },

  actions: {
    async fetchLobbies(page = this.pagination.page) {
      this.loading = true
      this.error = null
      try {
        const { data } = await lobbyApi.list(page, this.pagination.limit)
        this.lobbies = data.lobbies ?? []
        this.pagination = {
          ...this.pagination,
          ...(data.pagination ?? {}),
        }
      } catch {
        this.error = 'Failed to load lobbies'
      } finally {
        this.loading = false
      }
    },

    async nextPage() {
      if (!this.pagination.hasNext) return
      await this.fetchLobbies(this.pagination.page + 1)
    },

    async prevPage() {
      if (!this.pagination.hasPrev) return
      await this.fetchLobbies(this.pagination.page - 1)
    },

    async fetchActiveLobby() {
      try {
        const { data } = await lobbyApi.getActive()
        this.currentLobby = data.lobby
      } catch {
        this.currentLobby = null
      }
    },

    async fetchLobbyByCode(code) {
      this.loading = true
      try {
        const { data } = await lobbyApi.getByCode(code)
        this.currentLobby = data.lobby
      } catch {
        this.currentLobby = null
        this.error = 'Lobby not found'
      } finally {
        this.loading = false
      }
    },

    async createLobby(timeControl) {
      this.error = null
      const { data } = await lobbyApi.create(timeControl)
      this.currentLobby = data.lobby
      return data.lobby
    },

    async joinLobby(code) {
      this.error = null
      try {
        const { data } = await lobbyApi.getActive()
        if (data.lobby && data.lobby.code !== code) {
          await lobbyApi.leave(data.lobby.code)
        }
      } catch {
        // нет активного лобби
      }
      const { data } = await lobbyApi.join(code)
      this.currentLobby = data.lobby
    },

    async leaveLobby() {
      if (!this.currentLobby) return
      this.disconnectWS()
      try {
        await lobbyApi.leave(this.currentLobby.code)
      } catch (err) {
        if (err.response?.status !== 409) {
          throw err
        }
      } finally {
        this.currentLobby = null
      }
    },

    connectWS(lobbyCode) {
      if (this.ws) return

      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'

      const url = `${protocol}//${window.location.host}` + `/ws/api/v1/ws?lobbyCode=${lobbyCode}`

      const socket = new WebSocket(url)

      socket.onopen = () => {
        console.log('[ws] connected to lobby', lobbyCode)
      }

      socket.onmessage = (event) => {
        console.log('[ws] received:', event.data)

        try {
          const msg = JSON.parse(event.data)
          console.log('[ws] parsed:', msg.type, msg.payload)
          this._handleMessage(msg)
        } catch (e) {
          console.error('[ws] parse error', e)
        }
      }

      socket.onclose = () => {
        console.log('[ws] disconnected')
        this.ws = null
      }

      socket.onerror = (e) => {
        console.error('[ws] error', e)
      }

      this.ws = socket
    },

    disconnectWS() {
      if (this.ws) {
        this.ws.close()
        this.ws = null
      }
    },

    sendWS(type, payload = {}) {
      if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
        console.warn('[ws] not connected')
        return
      }
      console.log('[ws] sending:', type, payload)
      this.ws.send(JSON.stringify({ type, payload }))
    },

    toggleReady() {
      if (!this.currentLobby) return
      const currentReady = this.isHost ? this.currentLobby.hostReady : this.currentLobby.guestReady
      this.sendWS('ready', { isReady: !currentReady })
    },

    startMatch() {
      this.sendWS('start', {})
    },

    _handleMessage(msg) {
      console.log('[ws] handling:', msg.type)
      switch (msg.type) {
        case 'player_sync':
          this._applyPlayerSync(msg.payload)
          break
        case 'lobby_status':
          if (this.currentLobby) {
            this.currentLobby.status = msg.payload.status
            if (msg.payload.status === 'closed') {
              this.disconnectWS()
              this.lobbyClosed = true
            }
          }
          break
        case 'ready':
          this._applyReady(msg.payload)
          break
        case 'match_start':
          this.startedMatch = {
            ...msg.payload,
            lobbyCode: this.currentLobby?.code ?? null,
            hostId: this.currentLobby?.hostId ?? this.currentLobby?.hostID ?? null,
            guestId: this.currentLobby?.guestId ?? this.currentLobby?.guestID ?? null,
            hostUsername: this.currentLobby?.hostUsername ?? '',
            guestUsername: this.currentLobby?.guestUsername ?? '',
          }
          this.matchStart = this.startedMatch
          break
        case 'error':
          this.error = msg.payload.message
          break
      }
    },

    _applyPlayerSync(payload) {
      if (!this.currentLobby) return
      console.log('[ws] player_sync:', payload)
      if (payload.role === 'host') {
        this.currentLobby.hostUsername = payload.username
        this.currentLobby.hostReady = payload.isReady
        this.currentLobby.hostId = payload.playerId
      } else if (payload.role === 'guest') {
        this.currentLobby.guestUsername = payload.isPresent ? payload.username : null
        this.currentLobby.guestReady = payload.isReady
        this.currentLobby.guestId = payload.isPresent ? payload.playerId : null
      }
    },

    _applyReady(payload) {
      if (!this.currentLobby) return
      console.log('[ws] ready:', payload, 'hostId:', this.currentLobby.hostId)
      if (payload.playerId === this.currentLobby.hostId) {
        this.currentLobby.hostReady = payload.isReady
      } else {
        this.currentLobby.guestReady = payload.isReady
      }
    },
  },
})
