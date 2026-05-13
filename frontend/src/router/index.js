import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '@/views/HomeView.vue'
import AuthView from '@/views/AuthView.vue'
import AppView from '@/views/AppView.vue'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView,
      meta: { title: 'FOG OF WAR CHESS' },
    },
    {
      path: '/auth',
      name: 'auth',
      component: AuthView,
      meta: { title: 'Authentication | FOG-CHESS' },
    },
    {
      path: '/app',
      name: 'app',
      component: AppView,
      meta: { requiresAuth: true, title: 'Dashboard' },
    },
    {
      path: '/lobbies/:code',
      name: 'lobby',
      component: () => import('@/views/LobbyView.vue'), // создадим позже
      meta: { requiresAuth: true, title: 'Lobby' },
    },
    {
      path: '/match/:id',
      name: 'match',
      component: () => import('@/views/MatchView.vue'),
      meta: { requiresAuth: true, title: 'Match' },
    },
  ],
})

router.beforeEach((to, from) => {
  const auth = useAuthStore()

  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return '/auth'
  }
  if (to.name === 'auth' && auth.isAuthenticated) {
    return '/app'
  }
})

export default router
