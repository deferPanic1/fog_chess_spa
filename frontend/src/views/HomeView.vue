<template>
  <div class="flex flex-col items-center justify-center px-4 py-12">
    <!-- ASCII logo -->
    <pre
      class="ascii-logo text-gruvbox-orange text-[0.45rem] sm:text-xs md:text-sm lg:text-base tracking-tight drop-shadow-[0_0_15px_rgba(214,93,14,0.5)] animate-pulse-slow"
    >
███████╗ ██████╗  ██████╗
██╔════╝██╔═══██╗██╔════╝
█████╗  ██║   ██║██║  ███╗
██╔══╝  ██║   ██║██║   ██║
██║     ╚██████╔╝╚██████╔╝
╚═╝      ╚═════╝  ╚═════╝
        ─── of ───
██╗    ██╗ █████╗ ██████╗
██║    ██║██╔══██╗██╔══██╗
██║ █╗ ██║███████║██████╔╝
██║███╗██║██╔══██║██╔══██╗
╚███╔███╔╝██║  ██║██║  ██║
 ╚══╝╚══╝ ╚═╝  ╚═╝╚═╝  ╚═╝
    </pre>

    <div
      class="bg-gruvbox-bg1 border border-gruvbox-bg2 rounded p-6 max-w-2xl w-full mb-8 h-64 overflow-y-auto"
    >
      <div class="text-gruvbox-fg4 mb-2">$ fog-chess --help</div>
      <div ref="typewriterOutput" class="text-gruvbox-fg leading-tight"></div>
    </div>

    <div class="flex gap-4 max-w-2xl w-full">
      <router-link to="/auth" class="btn-cta btn-cta--primary">
        &gt; [Login / Register]
      </router-link>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'

const typewriterOutput = ref(null)

const text = `FOG OF WAR CHESS v0.1.0

A chess variant where you can only see squares
controlled by your pieces. Enemy pieces are hidden
in the fog until they enter your vision.

You see only what your pieces control.
Strike from the shadows. Trust your intuition.`

function type() {
  const output = typewriterOutput.value
  if (!output) return

  let i = 0
  let html = ''

  const typingCursor = '<span class="text-gruvbox-orange">|</span>'
  const doneCursor = '<span class="cursor text-gruvbox-orange">█</span>'

  function step() {
    if (i < text.length) {
      const char = text[i]

      if (char === '\n') {
        html += '<br>'
      } else {
        html += char
      }

      output.innerHTML = html + typingCursor
      i++

      let delay = 15
      if (char === '\n') delay = 120
      else if (char === '.') delay = 60
      else if (char === ',') delay = 40

      setTimeout(step, delay)
    } else {
      output.innerHTML = html + doneCursor
    }
  }

  step()
}

onMounted(() => {
  type()
})
</script>
