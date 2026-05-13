<template>
  <header class="border-b border-gruvbox-bg2 px-6 py-2 flex items-center justify-between">
    <router-link
      :to="auth.isAuthenticated ? '/app' : '/'"
      class="text-gruvbox-orange font-bold text-lg no-underline hover:opacity-80"
    >
      &gt; FOG-CHESS
    </router-link>

    <div v-if="auth.isAuthenticated" class="flex items-center gap-2">
      <router-link
        v-if="auth.isAdmin"
        to="/admin/users"
        class="cursor-pointer text-gruvbox-fg text-sm px-3 py-1 hover:bg-gruvbox-fg hover:text-gruvbox-bg transition-colors duration-200"
      >
        [Admin]
      </router-link>
      <div class="text-gruvbox-fg text-sm px-3 py-1">[{{ auth.username }}]</div>
      <button
        @click="handleLogout"
        class="cursor-pointer text-gruvbox-fg text-sm px-3 py-1 hover:bg-gruvbox-fg hover:text-gruvbox-bg transition-colors duration-200"
      >
        [Logout]
      </button>
    </div>

    <router-link
      v-else
      to="/auth"
      class="text-gruvbox-fg text-sm px-3 py-1 hover:bg-gruvbox-fg hover:text-gruvbox-bg transition-colors duration-200"
    >
      [Login]
    </router-link>
  </header>
</template>

<script setup>
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'

const auth = useAuthStore()
const router = useRouter()

function handleLogout() {
  auth.logout()
  router.push('/')
}
</script>
