<template>
  <div class="border border-gruvbox-bg2 bg-gruvbox-bg1 rounded text-sm w-full max-w-2xl">
    <div class="text-gruvbox-fg4 px-4 py-2 border-b border-gruvbox-bg2">
      [user@fog-chess]$ cat profile.txt
    </div>

    <div v-if="loading" class="px-4 py-3 text-gruvbox-fg4">Loading...</div>

    <template v-else>
      <div class="px-4 py-3 space-y-1 text-gruvbox-fg">
        <div>
          <span class="text-gruvbox-fg4">username:</span>
          <span class="text-gruvbox-orange"> {{ profile.username }}</span>
        </div>
        <div>
          <span class="text-gruvbox-fg4">rating:</span>
          <span class="text-gruvbox-yellow"> {{ profile.rating }} ELO</span>
        </div>
        <div>
          <span class="text-gruvbox-fg4">joined:</span>
          {{ formatDate(profile.createdAt) }}
        </div>
      </div>

      <div class="text-gruvbox-fg4 px-4 py-2 border-t border-b border-gruvbox-bg2">
        [user@fog-chess]$ cat stats.txt
      </div>

      <div v-if="statsLoading" class="px-4 py-3 text-gruvbox-fg4">Loading stats...</div>
      <div v-else class="px-4 py-3 space-y-1 text-gruvbox-fg">
        <div><span class="text-gruvbox-fg4">games:</span> {{ stats.games }}</div>
        <div>
          <span class="text-gruvbox-fg4">wins:</span>
          <span class="text-gruvbox-green"> {{ stats.wins }} ({{ pct(stats.wins) }}%)</span>
        </div>
        <div>
          <span class="text-gruvbox-fg4">losses:</span>
          <span class="text-gruvbox-red"> {{ stats.losses }} ({{ pct(stats.losses) }}%)</span>
        </div>
        <div>
          <span class="text-gruvbox-fg4">draws:</span>
          <span class="text-gruvbox-blue"> {{ stats.draws }} ({{ pct(stats.draws) }}%)</span>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { userApi } from '@/api/user'

const loading = ref(true)
const statsLoading = ref(true)

const profile = ref({ username: '', rating: 0, createdAt: null })
const stats = ref({ games: 0, wins: 0, losses: 0, draws: 0 })

function formatDate(date) {
  if (!date) return '—'
  return new Date(date).toLocaleDateString()
}

function pct(value) {
  if (!stats.value.games) return 0
  return Math.round((value / stats.value.games) * 100)
}

async function loadProfile() {
  try {
    const { data } = await userApi.me()
    profile.value = data.user
  } finally {
    loading.value = false
  }
}

async function loadStats() {
  try {
    const { data } = await userApi.stats()
    stats.value = data.stats
  } finally {
    statsLoading.value = false
  }
}

onMounted(() => {
  loadProfile()
  loadStats()
})
</script>
