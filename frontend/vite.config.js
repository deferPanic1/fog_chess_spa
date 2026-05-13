import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'
import tailwindcss from '@tailwindcss/vite'

const apiTarget = process.env.VITE_API_TARGET || 'http://localhost:8080'
const wsTarget = process.env.VITE_WS_TARGET || 'ws://localhost:8080'

export default defineConfig({
  plugins: [vue(), vueDevTools({ launchEditor: 'zed' }), tailwindcss()],

  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },

  server: {
    host: '0.0.0.0',

    allowedHosts: ['acronym-ointment-blitz.ngrok-free.dev'],

    proxy: {
      '/api': {
        target: apiTarget,
        changeOrigin: true,
        secure: false,
      },

      '/ws': {
        target: wsTarget,
        ws: true,
        changeOrigin: true,

        rewrite: (path) => path.replace(/^\/ws/, ''),
      },
    },
  },
})
