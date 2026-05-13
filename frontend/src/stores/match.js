import { defineStore } from 'pinia'
import { useAuthStore } from './auth'
import { useLobbyStore } from './lobby'
import { userApi } from '@/api/user'

export const useMatchStore = defineStore('match', {
  state: () => ({
    matchId: null,
    status: 'active', // 'active' | 'finished' | 'aborted'
    result: null, // 'win' | 'loss' | 'draw' | null

    playerColor: 'white', // цвет текущего юзера
    opponentColor: 'black',
    playerUsername: '',
    opponentUsername: '',
    playerRating: null,
    opponentRating: null,
    playerRatingDelta: 0,
    opponentRatingDelta: 0,

    playerTime: 600,
    opponentTime: 600,

    // Доска
    visibleFen: 'rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1',
    turnColor: 'white',
    lastMove: null,
    legalDests: new Map(), // Map<square, square[]>
    foggedSquares: [],

    // История ходов
    moves: [],
    boardSyncNonce: 0,

    ws: null,
  }),

  getters: {
    isMyTurn: (state) => state.turnColor === state.playerColor,
    visibleSquares: (state) => 64 - state.foggedSquares.length,
  },

  actions: {
    async init(matchId) {
      this.matchId = matchId
      const auth = useAuthStore()
      const lobbyStore = useLobbyStore()
      this.playerUsername = auth.username ?? ''
      this.playerRating = auth.user?.rating ?? null
      this._hydratePlayersFromLobbyContext(lobbyStore.startedMatch, auth.user?.id)
      await this._loadMissingPlayerData(auth.user?.id)
      this._connectWS(matchId)
    },

    cleanup() {
      this._disconnectWS()
      this.$reset()
    },

    _connectWS(matchId) {
      if (this.ws) return

      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'

      const url = `${protocol}//${window.location.host}` + `/ws/api/v1/ws/match?matchId=${matchId}`

      const socket = new WebSocket(url)

      socket.onopen = () => {
        console.log('[match ws] connected', matchId)
      }

      socket.onmessage = (event) => {
        try {
          const msg = JSON.parse(event.data)
          console.log('[match ws]', msg.type, msg.payload)
          this._handleMessage(msg)
        } catch (e) {
          console.error('[match ws] parse error', e)
        }
      }

      socket.onclose = () => {
        console.log('[match ws] disconnected')
        this.ws = null
      }

      socket.onerror = (e) => {
        console.error('[match ws] error', e)
      }

      this.ws = socket
    },

    _disconnectWS() {
      if (this.ws) {
        this.ws.close()
        this.ws = null
      }
    },

    _sendWS(type, payload = {}) {
      if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
        console.warn('[match ws] not connected')
        return
      }
      this.ws.send(JSON.stringify({ type, payload }))
    },

    // Входящие msg
    _handleMessage(msg) {
      switch (msg.type) {
        // Начальное состояние матча
        case 'match_init':
          this._applyMatchInit(msg.payload)
          break

        // Обновление доски после хода
        case 'board_update':
          this._applyBoardUpdate(msg.payload)
          break

        // Обновление таймеров
        case 'clock_update':
          this.playerTime = msg.payload.playerTime
          this.opponentTime = msg.payload.opponentTime
          break

        // Конец игры
        case 'game_over':
          this._applyBoardUpdate(msg.payload)
          this.status = 'finished'
          this.result = msg.payload.result
          this.legalDests = new Map()
          break

        case 'error':
          console.error('[match] server error:', msg.payload.message)
          this.boardSyncNonce += 1
          break
      }
    },

    _applyMatchInit(payload) {
      this.status = payload.status ?? this.status
      this.result = payload.result ?? this.result
      this.playerColor = payload.yourColor // 'white' | 'black'
      this.opponentColor = payload.yourColor === 'white' ? 'black' : 'white'
      this.opponentUsername = payload.opponentUsername ?? ''
      this.playerUsername = payload.playerUsername ?? this.playerUsername
      this.playerRating = payload.playerRating ?? this.playerRating
      this.opponentRating = payload.opponentRating ?? this.opponentRating
      this.playerRatingDelta = payload.playerRatingDelta ?? this.playerRatingDelta
      this.opponentRatingDelta = payload.opponentRatingDelta ?? this.opponentRatingDelta
      this.visibleFen = payload.fen
      this.turnColor = payload.turnColor
      this.playerTime = payload.playerTime
      this.opponentTime = payload.opponentTime
      this.foggedSquares = payload.foggedSquares ?? []
      this.legalDests = this._parseDests(payload.legalMoves ?? {})
    },

    _applyBoardUpdate(payload) {
      this.status = payload.status ?? this.status
      this.result = payload.result ?? this.result
      this.playerRating = payload.playerRating ?? this.playerRating
      this.opponentRating = payload.opponentRating ?? this.opponentRating
      this.playerRatingDelta = payload.playerRatingDelta ?? this.playerRatingDelta
      this.opponentRatingDelta = payload.opponentRatingDelta ?? this.opponentRatingDelta
      this.visibleFen = payload.fen
      this.turnColor = payload.turnColor
      this.lastMove = payload.lastMove ?? null
      this.foggedSquares = payload.foggedSquares ?? []
      this.legalDests = this._parseDests(payload.legalMoves ?? {})
      if (payload.move) {
        this.moves.push({ san: payload.move.san, color: payload.move.color })
      }
    },

    // Moves

    sendMove({ from, to, promotion }) {
      const payload = { from, to }
      if (promotion) {
        payload.promotion = promotion
      }
      this._sendWS('move', payload)
    },

    resign() {
      this._sendWS('resign', {})
    },

    _parseDests(obj) {
      const map = new Map()
      for (const [from, targets] of Object.entries(obj)) {
        map.set(from, targets)
      }
      return map
    },

    _hydratePlayersFromLobbyContext(startedMatch, currentUserId) {
      if (!startedMatch || startedMatch.matchId !== this.matchId || !currentUserId) return

      const isWhite = startedMatch.whitePlayer === currentUserId
      this.playerColor = isWhite ? 'white' : 'black'
      this.opponentColor = isWhite ? 'black' : 'white'

      const hostId = startedMatch.hostId
      const guestId = startedMatch.guestId
      const playerIsHost = hostId === currentUserId

      this.playerUsername = playerIsHost
        ? startedMatch.hostUsername || this.playerUsername
        : startedMatch.guestUsername || this.playerUsername

      this.opponentUsername = playerIsHost
        ? startedMatch.guestUsername || this.opponentUsername
        : startedMatch.hostUsername || this.opponentUsername
    },

    async _loadMissingPlayerData(currentUserId) {
      const lobbyStore = useLobbyStore()
      const startedMatch = lobbyStore.startedMatch
      if (!startedMatch || startedMatch.matchId !== this.matchId || !currentUserId) return

      const opponentId =
        startedMatch.whitePlayer === currentUserId
          ? startedMatch.blackPlayer
          : startedMatch.whitePlayer

      if (!opponentId) return

      try {
        const { data } = await userApi.getById(opponentId)
        this.opponentUsername = data.user?.username ?? this.opponentUsername
        this.opponentRating = data.user?.rating ?? this.opponentRating
      } catch (err) {
        console.error('[match] failed to load opponent profile', err)
      }
    },
  },
})
