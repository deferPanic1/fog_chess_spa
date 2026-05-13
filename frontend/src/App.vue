<template>
  <AppHeader />
  <main class="flex-1 flex flex-col">
    <RouterView />
  </main>
  <AppFooter />
</template>

<script setup>
import AppHeader from '@/components/AppHeader.vue'
import AppFooter from '@/components/AppFooter.vue'
import { useAuthStore } from '@/stores/auth'
import { onMounted } from 'vue'

const auth = useAuthStore()

onMounted(() => {
  auth.loadFromSession()
})
</script>
