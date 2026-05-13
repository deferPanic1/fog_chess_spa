<template>
  <div class="border border-gruvbox-bg2 bg-gruvbox-bg1 rounded text-sm w-full max-w-3xl">
    <div class="text-gruvbox-fg4 px-4 py-2 border-b border-gruvbox-bg2">
      [user@fog-chess]$ history --games
    </div>

    <div class="px-4 py-2 border-b border-gruvbox-bg2">
      <div class="flex text-gruvbox-fg4 font-bold">
        <span class="w-8 shrink-0">#</span>
        <span class="w-28 shrink-0">Date</span>
        <span class="w-32 shrink-0">Opponent</span>
        <span class="w-16 shrink-0">Result</span>
        <span>Rating</span>
      </div>
    </div>

    <div class="px-4 py-2 space-y-1">
      <div v-if="loading" class="text-gruvbox-fg4">Loading...</div>
      <template v-else-if="history.length">
        <div v-for="(game, i) in history" :key="i" class="flex text-gruvbox-fg">
          <span class="w-8 text-gruvbox-fg4 shrink-0">{{ rowNumber(i) }}</span>
          <span class="w-28 shrink-0">{{ formatDate(game.date) }}</span>
          <span class="w-32 shrink-0 text-gruvbox-orange">{{ game.opponent }}</span>
          <span class="w-16 shrink-0" :class="resultClass(game.result)">{{ game.result }}</span>
          <span :class="ratingClass(game.ratingChange)">{{ formatRating(game.ratingChange) }}</span>
        </div>
      </template>
      <div v-else class="text-gruvbox-fg4">No games yet.</div>

      <div class="flex items-center justify-between pt-3 mt-2 border-t border-gruvbox-bg2">
        <div class="text-gruvbox-fg4 text-xs">
          page {{ pagination.page }}
          <span v-if="pagination.totalPages">/ {{ pagination.totalPages }}</span>
          · {{ pagination.total }} games
        </div>
        <div class="flex gap-4">
          <button
            @click="loadHistory(pagination.page - 1)"
            :disabled="!pagination.hasPrev || loading"
            class="text-gruvbox-orange cursor-pointer disabled:text-gruvbox-fg4 disabled:cursor-not-allowed"
          >
            [Prev]
          </button>
          <button
            @click="loadHistory(pagination.page + 1)"
            :disabled="!pagination.hasNext || loading"
            class="text-gruvbox-orange cursor-pointer disabled:text-gruvbox-fg4 disabled:cursor-not-allowed"
          >
            [Next]
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { userApi } from '@/api/user'

const loading = ref(true)
const history = ref([])
const pagination = ref({
  page: 1,
  limit: 10,
  total: 0,
  totalPages: 0,
  hasPrev: false,
  hasNext: false,
})

function rowNumber(index) {
  return (pagination.value.page - 1) * pagination.value.limit + index + 1
}

function formatDate(date) {
  if (!date) return '—'
  return new Date(date).toLocaleDateString()
}

function resultClass(result) {
  return (
    {
      win: 'text-gruvbox-green',
      loss: 'text-gruvbox-red',
      draw: 'text-gruvbox-blue',
    }[result] ?? 'text-gruvbox-fg4'
  )
}

function ratingClass(change) {
  if (change > 0) return 'text-gruvbox-green'
  if (change < 0) return 'text-gruvbox-red'
  return 'text-gruvbox-fg4'
}

function formatRating(change) {
  if (change > 0) return `+${change}`
  if (change < 0) return `${change}`
  return '0'
}

async function loadHistory(page = pagination.value.page) {
  loading.value = true
  try {
    const { data } = await userApi.history(page, pagination.value.limit)
    history.value = data.history ?? []
    pagination.value = {
      ...pagination.value,
      ...(data.pagination ?? {}),
    }
  } catch {
    history.value = []
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadHistory()
})
</script>
