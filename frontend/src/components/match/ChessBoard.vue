<template>
  <div class="chess-board-wrapper" ref="wrapperRef">
    <div class="fog-layer" aria-hidden="true">
      <div v-for="sq in foggedSquares" :key="sq" class="fog-square" :style="squareStyle(sq)" />
    </div>
    <div ref="boardRef" class="cg-wrap" />
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { Chessground } from 'chessground'
import 'chessground/assets/chessground.base.css'
import 'chessground/assets/chessground.brown.css'
import 'chessground/assets/chessground.cburnett.css'

const props = defineProps({
  // 'white' | 'black'
  orientation: { type: String, default: 'white' },
  // FEN строка только с видимыми фигурами
  fen: { type: String, default: 'rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1' },
  // 'white' | 'black' — чей ход
  turnColor: { type: String, default: 'white' },
  // Map<square, square[]> — допустимые ходы для текущего игрока
  dests: { type: Object, default: () => new Map() },
  // Последний ход
  lastMove: { type: Array, default: null },
  // Клетки с туманом (массив: ['a7', 'b8', ...])
  foggedSquares: { type: Array, default: () => [] },
  viewOnly: { type: Boolean, default: false },
  // Цвет текущего игрока
  playerColor: { type: String, default: 'white' },
  // синхронизация доски с сервером
  syncNonce: { type: Number, default: 0 },
})

const emit = defineEmits(['move'])

const boardRef = ref(null)
const wrapperRef = ref(null)
let ground = null

onMounted(() => {
  ground = Chessground(boardRef.value, buildConfig())
})

onUnmounted(() => {
  ground?.destroy()
})

watch(
  () => [
    props.fen,
    props.turnColor,
    props.dests,
    props.lastMove,
    props.viewOnly,
    props.orientation,
    props.syncNonce,
  ],
  () => {
    ground?.set(buildConfig())
  },
  { deep: true },
)

function buildConfig() {
  return {
    orientation: props.orientation,
    fen: props.fen,
    turnColor: props.turnColor,
    lastMove: props.lastMove ?? undefined,
    viewOnly: props.viewOnly,
    highlight: {
      lastMove: true,
      check: true,
    },
    animation: {
      enabled: true,
      duration: 150,
    },
    movable: {
      free: false,
      color: props.viewOnly ? undefined : props.playerColor,
      dests: props.dests,
      events: {
        after: (orig, dest) => {
          emit('move', { from: orig, to: dest })
        },
      },
    },
    draggable: {
      enabled: !props.viewOnly,
      showGhost: true,
    },
    selectable: {
      enabled: true,
    },
  }
}

function squareStyle(sq) {
  const file = sq.charCodeAt(0) - 97 // a=0 ... h=7
  const rank = parseInt(sq[1]) - 1 // 1=0 ... 8=7

  let left, bottom
  if (props.orientation === 'white') {
    left = file * 12.5
    bottom = rank * 12.5
  } else {
    left = (7 - file) * 12.5
    bottom = (7 - rank) * 12.5
  }

  return {
    left: `${left}%`,
    bottom: `${bottom}%`,
    width: '12.5%',
    height: '12.5%',
  }
}
</script>

<style scoped>
.chess-board-wrapper {
  position: relative;
  width: 100%;
  aspect-ratio: 1 / 1;
}

.cg-wrap {
  width: 100%;
  height: 100%;
}

.fog-layer {
  position: absolute;
  inset: 0;
  pointer-events: none;
  z-index: 10;
}

.fog-square {
  position: absolute;
  background: rgba(20, 16, 14, 0.82);
  transition: opacity 0.3s ease;
}

:deep(.cg-wrap) {
  border-radius: 2px;
  overflow: hidden;
}
</style>
