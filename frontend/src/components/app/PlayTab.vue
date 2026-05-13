<template>
  <div class="border border-gruvbox-bg2 bg-gruvbox-bg1 rounded text-sm w-full max-w-2xl">
    <div v-if="lobbyStore.error" class="px-4 py-2 text-gruvbox-red border-b border-gruvbox-bg2">
      {{ lobbyStore.error }}
    </div>

    <div
      v-if="activeMatch"
      class="px-4 py-4 border-b border-gruvbox-bg2 bg-[rgba(184,187,38,0.08)] space-y-3"
    >
      <div class="text-gruvbox-yellow">[notice] you have an active match in progress</div>
      <div class="text-gruvbox-fg text-sm space-y-1">
        <div>
          <span class="text-gruvbox-fg4">opponent:</span>
          <span class="text-gruvbox-orange"> {{ activeMatch.opponentUsername }}</span>
        </div>
        <div>
          <span class="text-gruvbox-fg4">your color:</span>
          <span class="text-gruvbox-fg"> {{ activeMatch.yourColor }}</span>
        </div>
      </div>
      <router-link
        :to="`/match/${activeMatch.matchId}`"
        class="btn-cta btn-cta--primary no-underline block text-center"
      >
        [Resume active match]
      </router-link>
    </div>

    <div class="text-gruvbox-fg4 px-4 py-2 border-b border-gruvbox-bg2">
      [user@fog-chess]$ lobby --create
    </div>
    <div class="px-4 py-4 space-y-4">
      <div v-if="lobbyStore.currentLobby" class="text-gruvbox-yellow">
        [notice] you already have an active lobby:
        <router-link
          :to="`/lobbies/${lobbyStore.currentLobby.code}`"
          class="text-gruvbox-orange no-underline"
        >
          {{ lobbyStore.currentLobby.code }}
        </router-link>
      </div>
      <div v-else>
        <div class="text-gruvbox-fg mb-3">Select time control:</div>
        <div class="flex gap-2">
          <button
            v-for="time in [300, 600, 900]"
            :key="time"
            @click="handleCreate(time)"
            class="btn-cta btn-cta--primary flex-1 cursor-pointer"
          >
            [{{ time / 60 }} min]
          </button>
        </div>
      </div>
    </div>

    <div class="text-gruvbox-fg4 px-4 py-2 border-t border-b border-gruvbox-bg2">
      [user@fog-chess]$ lobby --list
    </div>
    <div class="px-4 py-4 space-y-3">
      <div class="flex items-center justify-between gap-3">
        <button @click="lobbyStore.fetchLobbies()" class="text-gruvbox-orange cursor-pointer">
        [Refresh]
        </button>
        <div class="text-gruvbox-fg4 text-xs">
          page {{ lobbyStore.pagination.page }}
          <span v-if="lobbyStore.pagination.totalPages">/ {{ lobbyStore.pagination.totalPages }}</span>
          · {{ lobbyStore.pagination.total }} total
        </div>
      </div>
      <div v-if="lobbyStore.loading" class="text-gruvbox-fg4">Loading...</div>
      <div v-else-if="lobbyStore.lobbies.length">
        <div
          v-for="lobby in lobbyStore.lobbies"
          :key="lobby.code"
          class="border border-gruvbox-bg2 rounded px-3 py-3 flex flex-col space-y-2 mb-2"
        >
          <div class="flex items-center justify-between">
            <div>
              <span class="text-gruvbox-fg4">code:</span>
              <span class="text-gruvbox-orange"> {{ lobby.code }}</span>
            </div>
            <div :class="lobbyStore.statusClass(lobby.status)">[{{ lobby.status }}]</div>
          </div>
          <div>
            <span class="text-gruvbox-fg4">host:</span>
            <span class="text-gruvbox-orange"> {{ lobby.hostUsername }}</span>
          </div>
          <div>
            <span class="text-gruvbox-fg4">guest:</span>
            <span :class="lobby.guestUsername ? 'text-gruvbox-orange' : 'text-gruvbox-fg4'">
              {{ lobby.guestUsername || '[waiting]' }}
            </span>
          </div>
          <div>
            <span class="text-gruvbox-fg4">time: {{ lobby.timeControl / 60 }} min</span>
          </div>
          <button
            @click="handleJoin(lobby.code)"
            class="w-full mt-2 px-4 py-2 rounded text-center text-gruvbox-orange border border-gruvbox-bg2 bg-transparent hover:bg-gruvbox-bg2 transition-colors duration-150 text-sm cursor-pointer"
          >
            [Open]
          </button>
        </div>
      </div>
      <div v-else class="text-gruvbox-fg4">No open lobbies yet.</div>

      <div class="flex items-center justify-between pt-2 border-t border-gruvbox-bg2">
        <button
          @click="lobbyStore.prevPage()"
          :disabled="!lobbyStore.pagination.hasPrev || lobbyStore.loading"
          class="text-gruvbox-orange cursor-pointer disabled:text-gruvbox-fg4 disabled:cursor-not-allowed"
        >
          [Prev]
        </button>
        <button
          @click="lobbyStore.nextPage()"
          :disabled="!lobbyStore.pagination.hasNext || lobbyStore.loading"
          class="text-gruvbox-orange cursor-pointer disabled:text-gruvbox-fg4 disabled:cursor-not-allowed"
        >
          [Next]
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useLobbyStore } from '@/stores/lobby'
import { userApi } from '@/api/user'

const router = useRouter()
const lobbyStore = useLobbyStore()
const activeMatch = ref(null)

async function handleCreate(time) {
  try {
    const lobby = await lobbyStore.createLobby(time)
    lobbyStore.connectWS(lobby.code)
    router.push(`/lobbies/${lobby.code}`)
  } catch {
    // ошибка в lobbyStore.error
  }
}

async function handleJoin(code) {
  try {
    await lobbyStore.joinLobby(code)
    lobbyStore.connectWS(code)
    router.push(`/lobbies/${code}`)
  } catch {
    // ошибка в lobbyStore.error
  }
}

async function fetchActiveMatch() {
  try {
    const { data } = await userApi.activeMatch()
    activeMatch.value = data.match ?? null
  } catch {
    activeMatch.value = null
  }
}

onMounted(() => {
  lobbyStore.fetchLobbies()
  lobbyStore.fetchActiveLobby()
  fetchActiveMatch()
})
</script>
