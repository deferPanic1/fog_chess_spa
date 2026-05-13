<template>
  <div class="flex-1 flex items-center justify-center px-4 py-12">
    <div
      class="border border-gruvbox-bg2 bg-gruvbox-bg1 rounded max-w-md w-full text-gruvbox-fg p-4"
    >
      <div class="text-sm mb-2">[user@fog-chess]$ authenticate</div>

      <!-- Вкладки -->
      <div class="flex gap-4 mb-4 border-b border-gruvbox-bg2 pb-1">
        <button
          @click="setTab('login')"
          class="cursor-pointer"
          :class="activeTab === 'login' ? 'text-gruvbox-orange font-bold' : 'text-gruvbox-fg4'"
        >
          [Login]
        </button>
        <button
          @click="setTab('register')"
          class="cursor-pointer"
          :class="activeTab === 'register' ? 'text-gruvbox-orange font-bold' : 'text-gruvbox-fg4'"
        >
          [Register]
        </button>
      </div>

      <!-- Общая ошибка -->
      <div v-if="auth.errors.General" class="text-red-500 text-sm mb-2">
        [Error] {{ auth.errors.General }}
      </div>

      <!-- Форма входа -->
      <form
        v-if="activeTab === 'login'"
        @submit.prevent="handleLogin"
        class="flex flex-col gap-3"
        novalidate
      >
        <div>
          <label class="text-gruvbox-fg4 text-sm block">username:</label>
          <div class="input-wrapper">
            <input
              v-model="loginForm.username"
              type="text"
              name="username"
              class="input-terminal"
              autocomplete="username"
              required
              minlength="3"
              maxlength="20"
            />
          </div>
          <div v-if="auth.errors.Username" class="text-red-500 text-xs mt-1">
            {{ auth.errors.Username }}
          </div>
        </div>

        <hr class="border-gruvbox-bg2" />

        <div>
          <label class="text-gruvbox-fg4 text-sm block">password:</label>
          <div class="input-wrapper">
            <input
              v-model="loginForm.password"
              type="password"
              name="password"
              class="input-terminal"
              autocomplete="current-password"
              required
              minlength="6"
            />
          </div>
          <div v-if="auth.errors.Password" class="text-red-500 text-xs mt-1">
            {{ auth.errors.Password }}
          </div>
        </div>

        <button type="submit" class="btn-terminal" :disabled="authLoading">
          [{{ authLoading ? '...' : 'Login' }}]
        </button>
      </form>

      <!-- Форма регистрации -->
      <form v-else @submit.prevent="handleRegister" class="flex flex-col gap-3" novalidate>
        <div>
          <label class="text-gruvbox-fg4 text-sm block">username:</label>
          <div class="input-wrapper">
            <input
              v-model="registerForm.username"
              type="text"
              name="username"
              class="input-terminal"
              required
              minlength="3"
              maxlength="20"
            />
          </div>
          <div v-if="auth.errors.Username" class="text-red-500 text-xs mt-1">
            {{ auth.errors.Username }}
          </div>
        </div>

        <hr class="border-gruvbox-bg2" />

        <div>
          <label class="text-gruvbox-fg4 text-sm block">email:</label>
          <div class="input-wrapper">
            <input
              v-model="registerForm.email"
              type="email"
              name="email"
              class="input-terminal"
              required
            />
          </div>
          <div v-if="auth.errors.Email" class="text-red-500 text-xs mt-1">
            {{ auth.errors.Email }}
          </div>
        </div>

        <hr class="border-gruvbox-bg2" />

        <div>
          <label class="text-gruvbox-fg4 text-sm block">password:</label>
          <div class="input-wrapper">
            <input
              v-model="registerForm.password"
              type="password"
              name="password"
              class="input-terminal"
              required
              minlength="6"
            />
          </div>
          <div v-if="auth.errors.Password" class="text-red-500 text-xs mt-1">
            {{ auth.errors.Password }}
          </div>
        </div>

        <button type="submit" class="btn-terminal" :disabled="authLoading">
          [{{ authLoading ? '...' : 'Create Account' }}]
        </button>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const authLoading = ref(false)

const activeTab = ref(route.query.tab === 'register' ? 'register' : 'login')

function setTab(tab) {
  activeTab.value = tab
  router.replace({ query: { tab } })
}

watch(
  () => route.query.tab,
  (newTab) => {
    if (newTab === 'register' || newTab === 'login') {
      activeTab.value = newTab
    }
  },
)

const loginForm = reactive({
  username: '',
  password: '',
})

const registerForm = reactive({
  username: '',
  email: '',
  password: '',
})

async function handleLogin() {
  authLoading.value = true
  try {
    await auth.loginAction({
      username: loginForm.username,
      password: loginForm.password,
    })
    router.push('/app')
  } catch (err) {
    console.error(err)
  } finally {
    authLoading.value = false
  }
}

async function handleRegister() {
  authLoading.value = true
  try {
    await auth.registerAction({
      username: registerForm.username,
      email: registerForm.email,
      password: registerForm.password,
    })
    router.push('/app')
  } catch (err) {
    console.error(err)
  } finally {
    authLoading.value = false
  }
}
</script>
