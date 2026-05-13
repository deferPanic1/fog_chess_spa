<template>
  <section class="flex-1 flex flex-col">
    <!-- Main layout -->
    <div class="flex-1 flex flex-col lg:flex-row items-center justify-center gap-6 p-4">
      <!-- Board area -->
      <div class="flex items-center justify-center w-full lg:w-auto">
        <div class="board-container relative">
          <ChessBoard
            :orientation="matchStore.playerColor"
            :fen="matchStore.visibleFen"
            :turn-color="matchStore.turnColor"
            :dests="matchStore.legalDests"
            :last-move="visibleLastMove"
            :fogged-squares="matchStore.foggedSquares"
            :view-only="!matchStore.isMyTurn || matchStore.status !== 'active'"
            :player-color="matchStore.playerColor"
            :sync-nonce="matchStore.boardSyncNonce"
            @move="handleBoardMove"
          />

          <div
            v-if="promotionDialog"
            class="absolute inset-0 z-20 flex items-center justify-center bg-gruvbox-bg/85 p-4"
          >
            <div
              class="w-full max-w-xs rounded border border-gruvbox-bg3 bg-gruvbox-bg1 p-4 shadow-2xl"
            >
              <div class="mb-3 text-sm font-bold text-gruvbox-fg">Choose promotion</div>
              <div class="grid grid-cols-2 gap-2">
                <button
                  v-for="option in promotionOptions"
                  :key="option.value"
                  class="btn-cta btn-cta--secondary promotion-option"
                  @click="confirmPromotion(option.value)"
                >
                  <span class="promotion-glyph">{{ option.glyph }}</span>
                  <span class="text-xs">{{ option.label }}</span>
                </button>
              </div>
              <button class="mt-3 w-full text-xs text-gruvbox-fg4 hover:text-gruvbox-fg" @click="cancelPromotion">
                Cancel
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Sidebar -->
      <div class="w-full lg:w-72 shrink-0 flex items-center">
        <MatchInfo
          class="w-full"
          :match-id="$route.params.id"
          :status="matchStore.status"
          :result="matchStore.result"
          :player-username="matchStore.playerUsername"
          :opponent-username="matchStore.opponentUsername"
          :player-rating="matchStore.playerRating"
          :opponent-rating="matchStore.opponentRating"
          :player-rating-delta="matchStore.playerRatingDelta"
          :opponent-rating-delta="matchStore.opponentRatingDelta"
          :player-color="matchStore.playerColor"
          :opponent-color="matchStore.opponentColor"
          :turn-color="matchStore.turnColor"
          :player-time="matchStore.playerTime"
          :opponent-time="matchStore.opponentTime"
          :moves="matchStore.moves"
          :visible-squares="matchStore.visibleSquares"
          :lobby-code="lobbyCode"
          @resign="matchStore.resign"
          @return-to-lobby="handleReturnToLobby"
        />
      </div>
    </div>
  </section>
</template>

<script setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useLobbyStore } from '@/stores/lobby'
import { useMatchStore } from '@/stores/match'
import ChessBoard from '@/components/match/ChessBoard.vue'
import MatchInfo from '@/components/match/MatchInfo.vue'

const route = useRoute()
const router = useRouter()
const lobbyStore = useLobbyStore()
const matchStore = useMatchStore()
const promotionDialog = ref(null)

const visibleLastMove = computed(() => {
  if (!matchStore.lastMove?.length) return null
  return matchStore.lastMove.some((sq) => matchStore.foggedSquares.includes(sq))
    ? null
    : matchStore.lastMove
})

const lobbyCode = computed(
  () => lobbyStore.startedMatch?.lobbyCode ?? lobbyStore.currentLobby?.code ?? null,
)

const promotionOptions = computed(() => {
  const isWhite = matchStore.playerColor === 'white'
  return [
    { value: 'queen', label: 'Queen', glyph: isWhite ? 'Q' : 'q' },
    { value: 'rook', label: 'Rook', glyph: isWhite ? 'R' : 'r' },
    { value: 'bishop', label: 'Bishop', glyph: isWhite ? 'B' : 'b' },
    { value: 'knight', label: 'Knight', glyph: isWhite ? 'N' : 'n' },
  ]
})

function handleBoardMove(move) {
  if (!move || !requiresPromotion(move.from, move.to)) {
    matchStore.sendMove(move)
    return
  }

  promotionDialog.value = move
  matchStore.boardSyncNonce += 1
}

function confirmPromotion(promotion) {
  if (!promotionDialog.value) return

  matchStore.sendMove({
    ...promotionDialog.value,
    promotion,
  })
  promotionDialog.value = null
}

function cancelPromotion() {
  promotionDialog.value = null
  matchStore.boardSyncNonce += 1
}

function requiresPromotion(from, to) {
  if (!from || !to) return false

  const piece = pieceAtSquare(matchStore.visibleFen, from)
  if (piece?.toLowerCase() !== 'p') return false

  const targetRank = Number.parseInt(to[1], 10)
  return (
    (matchStore.playerColor === 'white' && targetRank === 8) ||
    (matchStore.playerColor === 'black' && targetRank === 1)
  )
}

function pieceAtSquare(fen, square) {
  const [placement] = String(fen || '').split(' ')
  const ranks = placement?.split('/') ?? []
  if (ranks.length !== 8 || !square || square.length !== 2) return null

  const file = square.charCodeAt(0) - 97
  const rank = 8 - Number.parseInt(square[1], 10)
  if (file < 0 || file > 7 || rank < 0 || rank > 7) return null

  let fileIndex = 0
  for (const char of ranks[rank]) {
    const empty = Number.parseInt(char, 10)
    if (Number.isFinite(empty)) {
      fileIndex += empty
      continue
    }
    if (fileIndex === file) return char
    fileIndex += 1
  }

  return null
}

function handleReturnToLobby() {
  router.push(lobbyCode.value ? `/lobbies/${lobbyCode.value}` : '/app')
}

onMounted(() => {
  matchStore.init(route.params.id)
})

onUnmounted(() => {
  matchStore.cleanup()
})
</script>

<style scoped>
.board-container {
  width: min(calc(100vw - 2rem), 560px);
  height: min(calc(100vw - 2rem), 560px);
  flex-shrink: 0;
}

.promotion-option {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  align-items: center;
  justify-content: center;
  min-height: 4.5rem;
}

.promotion-glyph {
  font-size: 1.4rem;
  line-height: 1;
  color: var(--color-gruvbox-yellow);
}
</style>
