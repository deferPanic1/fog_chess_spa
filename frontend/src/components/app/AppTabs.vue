<template>
  <nav class="border-b border-gruvbox-bg2 px-4 py-2 flex gap-4 text-sm">
    <button
      v-for="tab in tabs"
      :key="tab.value"
      @click="$emit('update:activeTab', tab.value)"
      class="cursor-pointer"
      :class="activeTab === tab.value ? 'text-gruvbox-orange font-bold' : 'text-gruvbox-fg4'"
    >
      [{{ tab.label }}]
    </button>
  </nav>
</template>

<script setup>
defineProps({
  activeTab: { type: String, required: true },
})

defineEmits(['update:activeTab'])

const tabs = [
  { label: 'Play', value: 'play' },
  { label: 'Profile', value: 'profile' },
  { label: 'History', value: 'history' },
]
</script>
