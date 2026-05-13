<template>
  <div
    class="match-info border border-gruvbox-bg2 bg-gruvbox-bg1 rounded text-sm flex flex-col gap-0 overflow-hidden"
  >
    <!-- Header -->
    <div class="px-4 py-3 border-b border-gruvbox-bg2 flex items-center justify-between">
      <span class="text-gruvbox-fg font-bold">Match</span>
      <span class="text-xs px-2 py-0.5 rounded border" :class="statusBadge">{{ status }}</span>
    </div>

    <!-- Opponent block -->
    <div
      class="player-block px-4 py-3 border-b border-gruvbox-bg2"
      :class="{ 'active-turn': isOpponentTurn }"
    >
      <div class="flex items-center justify-between gap-2">
        <div class="flex items-center gap-2">
          <div class="flex flex-col">
            <div class="flex items-center gap-2">
              <span class="text-gruvbox-orange font-bold">{{ opponentUsername }}</span>
              <span class="text-gruvbox-fg4 text-xs">({{ opponentColor }})</span>
            </div>
            <div class="text-gruvbox-fg4 text-xs">
              rating:
              <span class="text-gruvbox-yellow">{{ formatRating(opponentRating) }}</span>
              <span
                v-if="showRatingDelta(opponentRatingDelta)"
                class="ml-1"
                :class="ratingDeltaClass(opponentRatingDelta)"
              >
                {{ formatRatingDelta(opponentRatingDelta) }}
              </span>
            </div>
          </div>
        </div>
        <div class="text-gruvbox-yellow font-bold tabular-nums">
          {{ formatTime(opponentTime) }}
        </div>
      </div>
      <!-- Thinking  -->
      <div
        v-if="isOpponentTurn && status === 'active'"
        class="mt-1 flex items-center gap-1 text-gruvbox-fg4 text-xs"
      >
        <span class="thinking-dot" /><span class="thinking-dot delay-1" /><span
          class="thinking-dot delay-2"
        />
        <span class="ml-1">thinking</span>
      </div>
    </div>

    <!-- Current player block -->
    <div
      class="player-block px-4 py-3 border-b border-gruvbox-bg2"
      :class="{ 'active-turn': isMyTurn }"
    >
      <div class="flex items-center justify-between gap-2">
        <div class="flex items-center gap-2">
          <div class="flex flex-col">
            <div class="flex items-center gap-2">
              <span class="text-gruvbox-green font-bold">{{ playerUsername }}</span>
              <span class="text-gruvbox-yellow text-xs">(you)</span>
              <span class="text-gruvbox-fg4 text-xs">({{ playerColor }})</span>
            </div>
            <div class="text-gruvbox-fg4 text-xs">
              rating:
              <span class="text-gruvbox-yellow">{{ formatRating(playerRating) }}</span>
              <span
                v-if="showRatingDelta(playerRatingDelta)"
                class="ml-1"
                :class="ratingDeltaClass(playerRatingDelta)"
              >
                {{ formatRatingDelta(playerRatingDelta) }}
              </span>
            </div>
          </div>
        </div>
        <div class="font-bold tabular-nums" :class="myTimeColor">
          {{ formatTime(playerTime) }}
        </div>
      </div>
      <div v-if="isMyTurn && status === 'active'" class="mt-1 text-gruvbox-green text-xs">
        ▶ your turn
      </div>
    </div>

    <!-- Actions -->
    <div class="px-4 py-3 flex flex-col gap-2">
      <button
        v-if="status === 'active'"
        @click="$emit('resign')"
        class="btn-cta btn-cta--secondary text-xs w-full"
      >
        Resign
      </button>
      <div v-else class="text-center text-gruvbox-fg4 text-xs py-1">
        <span v-if="result === 'draw'" class="text-gruvbox-yellow">½–½ Draw</span>
        <span v-else-if="result === 'win'" class="text-gruvbox-green">1–0 You win</span>
        <span v-else-if="result === 'loss'" class="text-gruvbox-red">0–1 You lose</span>
        <span v-else class="text-gruvbox-fg4">{{ status }}</span>
      </div>
      <button
        v-if="status !== 'active'"
        @click="$emit('return-to-lobby')"
        class="btn-cta btn-cta--primary text-xs w-full"
      >
        {{ lobbyCode ? 'Return to lobby' : 'Back to dashboard' }}
      </button>
    </div>

    <div class="px-4 py-3 border-t border-gruvbox-bg2 text-xs space-y-2">
      <div class="flex items-center justify-between gap-3">
        <span class="text-gruvbox-fg4">match</span>
        <span class="text-gruvbox-fg">{{ shortMatchId }}</span>
      </div>
      <div class="flex items-center justify-between gap-3">
        <span class="text-gruvbox-fg4">visible squares</span>
        <span class="text-gruvbox-yellow">{{ visibleSquares }}/64</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  matchId: { type: String, default: '' },
  status: { type: String, default: 'active' },
  result: { type: String, default: null },
  playerUsername: { type: String, default: '' },
  opponentUsername: { type: String, default: '' },
  playerRating: { type: Number, default: null },
  opponentRating: { type: Number, default: null },
  playerRatingDelta: { type: Number, default: 0 },
  opponentRatingDelta: { type: Number, default: 0 },
  playerColor: { type: String, default: 'white' },
  opponentColor: { type: String, default: 'black' },
  turnColor: { type: String, default: 'white' },
  playerTime: { type: Number, default: 600 },
  opponentTime: { type: Number, default: 600 },
  moves: { type: Array, default: () => [] },
  visibleSquares: { type: Number, default: 0 },
  lobbyCode: { type: String, default: '' },
})

defineEmits(['resign', 'return-to-lobby'])

const isMyTurn = computed(() => props.turnColor === props.playerColor)
const isOpponentTurn = computed(() => !isMyTurn.value)
const shortMatchId = computed(() => (props.matchId ? String(props.matchId).slice(0, 8) : 'pending'))

const statusBadge = computed(
  () =>
    ({
      active: 'text-gruvbox-green border-gruvbox-green',
      finished: 'text-gruvbox-fg4 border-gruvbox-fg4',
      aborted: 'text-gruvbox-red border-gruvbox-red',
    })[props.status] ?? 'text-gruvbox-fg4 border-gruvbox-fg4',
)

const myTimeColor = computed(() =>
  props.playerTime < 30 ? 'text-gruvbox-red' : 'text-gruvbox-yellow',
)

function formatTime(seconds) {
  if (seconds == null) return '--:--'
  const m = Math.floor(seconds / 60)
  const s = seconds % 60
  return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
}

function formatRating(rating) {
  return rating == null ? '---' : `${rating} ELO`
}

function showRatingDelta(delta) {
  return props.status === 'finished' && Number.isFinite(delta) && delta !== 0
}

function formatRatingDelta(delta) {
  return delta > 0 ? `(+${delta})` : `(${delta})`
}

function ratingDeltaClass(delta) {
  if (delta > 0) return 'text-gruvbox-green'
  if (delta < 0) return 'text-gruvbox-red'
  return 'text-gruvbox-fg4'
}
</script>

<style scoped>
.player-block {
  transition: background-color 0.2s;
}
.active-turn {
  background-color: rgba(80, 73, 69, 0.4); /* gruvbox-bg2 с прозрачностью */
}

.thinking-dot {
  display: inline-block;
  width: 4px;
  height: 4px;
  border-radius: 50%;
  background: var(--color-gruvbox-fg4);
  animation: dot-pulse 1.2s ease-in-out infinite;
}
.thinking-dot.delay-1 {
  animation-delay: 0.2s;
}
.thinking-dot.delay-2 {
  animation-delay: 0.4s;
}

.move-log {
  display: grid;
  gap: 0.35rem;
  max-height: 12rem;
  overflow-y: auto;
}

.move-row {
  display: grid;
  grid-template-columns: 2rem 3.5rem 1fr;
  gap: 0.5rem;
  align-items: center;
}

@keyframes dot-pulse {
  0%,
  80%,
  100% {
    opacity: 0.2;
    transform: scale(0.8);
  }
  40% {
    opacity: 1;
    transform: scale(1);
  }
}
</style>
