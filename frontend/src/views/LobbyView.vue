<template>
  <section class="flex-1 flex items-center justify-center px-4 py-8">
    <div class="border border-gruvbox-bg2 bg-gruvbox-bg1 rounded text-sm w-full max-w-3xl">
      <div v-if="lobbyStore.loading" class="px-4 py-8 text-gruvbox-fg4 text-center">Loading...</div>

      <template v-else-if="lobbyStore.currentLobby">
        <div
          class="px-4 py-2 border-b border-gruvbox-bg2 text-gruvbox-fg4 flex flex-wrap items-center justify-between gap-3"
        >
          <div>[user@fog-chess]$ lobby --open {{ lobbyStore.currentLobby.code }}</div>
          <div :class="lobbyStore.statusClass()">[{{ lobbyStore.currentLobby.status }}]</div>
        </div>

        <div class="px-4 py-4 space-y-6">
          <div class="space-y-1 text-gruvbox-fg">
            <div>
              <span class="text-gruvbox-fg4">code:</span>
              <span class="text-gruvbox-orange"> {{ lobbyStore.currentLobby.code }}</span>
            </div>
            <div>
              <span class="text-gruvbox-fg4">time:</span>
              {{ lobbyStore.currentLobby.timeControl / 60 }} min
            </div>
          </div>

          <div class="border border-gruvbox-bg2 rounded px-4 py-3 space-y-3">
            <!-- Host -->
            <div class="border border-gruvbox-bg2 rounded px-3 py-3">
              <div class="flex items-center justify-between gap-3">
                <div class="text-gruvbox-fg">
                  <span class="text-gruvbox-fg4">Host:</span>
                  <span class="text-gruvbox-orange">
                    {{ lobbyStore.currentLobby.hostUsername }}</span
                  >
                  <span v-if="lobbyStore.isHost" class="text-gruvbox-yellow"> (you)</span>
                </div>
                <div
                  :class="
                    lobbyStore.currentLobby.hostReady ? 'text-gruvbox-green' : 'text-gruvbox-red'
                  "
                >
                  [{{ lobbyStore.currentLobby.hostReady ? 'ready' : 'not ready' }}]
                </div>
              </div>
            </div>

            <!-- Guest -->
            <div class="border border-gruvbox-bg2 rounded px-3 py-3">
              <div class="flex items-center justify-between gap-3">
                <div class="text-gruvbox-fg">
                  <span class="text-gruvbox-fg4">Guest:</span>
                  <span
                    :class="
                      lobbyStore.currentLobby.guestUsername
                        ? 'text-gruvbox-orange'
                        : 'text-gruvbox-fg4'
                    "
                  >
                    {{ lobbyStore.currentLobby.guestUsername || '[waiting]' }}
                  </span>
                  <span v-if="lobbyStore.isGuest" class="text-gruvbox-yellow"> (you)</span>
                </div>
                <div
                  v-if="lobbyStore.currentLobby.guestUsername"
                  :class="
                    lobbyStore.currentLobby.guestReady ? 'text-gruvbox-green' : 'text-gruvbox-red'
                  "
                >
                  [{{ lobbyStore.currentLobby.guestReady ? 'ready' : 'not ready' }}]
                </div>
              </div>
            </div>
          </div>

          <div v-if="lobbyStore.error" class="text-gruvbox-red">{{ lobbyStore.error }}</div>

          <div class="space-y-3">
            <template v-if="lobbyStore.isParticipant">
              <!-- Toggle ready button -->
              <button
                @click="lobbyStore.toggleReady"
                class="btn-cta btn-cta--primary w-full cursor-pointer"
              >
                [toggle ready]
              </button>

              <!-- Start match button -->
              <button
                v-if="lobbyStore.isHost"
                @click="lobbyStore.startMatch"
                :disabled="lobbyStore.currentLobby.status !== 'ready'"
                class="btn-cta btn-cta--primary w-full cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed"
              >
                [start match]
              </button>
            </template>

            <template v-else>
              <button
                v-if="lobbyStore.currentLobby.status === 'waiting'"
                @click="handleJoin"
                class="btn-cta btn-cta--primary w-full cursor-pointer"
              >
                [join]
              </button>
              <div v-else class="text-gruvbox-yellow text-center">[notice] lobby is full</div>
            </template>

            <router-link
              to="/app"
              class="btn-cta btn-cta--secondary no-underline block text-center"
            >
              [Back]
            </router-link>

            <button
              v-if="lobbyStore.isParticipant"
              @click="handleLeave"
              class="btn-cta btn-cta--secondary w-full cursor-pointer"
            >
              [leave]
            </button>
          </div>
        </div>
      </template>

      <div v-else class="px-4 py-8 text-gruvbox-red text-center">Lobby not found.</div>
    </div>
  </section>
</template>

<script setup>
import { onMounted, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useLobbyStore } from '@/stores/lobby'

const route = useRoute()
const router = useRouter()
const lobbyStore = useLobbyStore()

watch(
  () => lobbyStore.lobbyClosed,
  (closed) => {
    if (closed) {
      lobbyStore.lobbyClosed = false
      lobbyStore.currentLobby = null
      router.push('/app')
    }
  },
)

// match_start msg
watch(
  () => lobbyStore.matchStart,
  (match) => {
    if (match) {
      lobbyStore.matchStart = null
      router.push(`/match/${match.matchId}`)
    }
  },
)

async function handleJoin() {
  await lobbyStore.joinLobby(lobbyStore.currentLobby.code)
  lobbyStore.connectWS(lobbyStore.currentLobby.code)
}

async function handleLeave() {
  await lobbyStore.leaveLobby()
  router.push('/app')
}

onMounted(async () => {
  await lobbyStore.fetchLobbyByCode(route.params.code)
  if (lobbyStore.isParticipant) {
    lobbyStore.connectWS(route.params.code)
  }
})

onUnmounted(() => {
  lobbyStore.disconnectWS()
})
</script>
