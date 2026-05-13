<template>
  <div class="flex-1 flex flex-col">
    <AppTabs :activeTab="activeTab" @update:activeTab="setTab" />
    <div class="flex-1 flex items-center justify-center px-4 py-8">
      <PlayTab v-if="activeTab === 'play'" />
      <ProfileTab v-else-if="activeTab === 'profile'" />
      <HistoryTab v-else-if="activeTab === 'history'" />
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AppTabs from '@/components/app/AppTabs.vue'
import PlayTab from '@/components/app/PlayTab.vue'
import ProfileTab from '@/components/app/ProfileTab.vue'
import HistoryTab from '@/components/app/HistoryTab.vue'

const route = useRoute()
const router = useRouter()

const activeTab = ref(route.query.tab || 'play')

function setTab(tab) {
  activeTab.value = tab
  router.replace({ query: { tab } })
}

watch(
  () => route.query.tab,
  (newTab) => {
    if (newTab && ['play', 'profile', 'history'].includes(newTab)) {
      activeTab.value = newTab
    }
  },
)
</script>
