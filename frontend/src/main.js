import './assets/main.css'
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import { useAuthStore } from '@/stores/auth'

const app = createApp(App)
app.use(createPinia())

const auth = useAuthStore()
await auth.loadFromSession()

app.use(router)
await router.isReady()

app.mount('#app')
