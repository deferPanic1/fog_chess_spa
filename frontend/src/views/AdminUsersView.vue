<template>
  <div class="flex-1 w-full px-4 py-8">
    <div class="mx-auto w-full max-w-5xl border border-gruvbox-bg2 bg-gruvbox-bg1 rounded">
      <div
        class="flex flex-col gap-3 border-b border-gruvbox-bg2 px-4 py-3 text-sm text-gruvbox-fg4 md:flex-row md:items-center md:justify-between"
      >
        <div>[admin@fog-chess]$ users --search --paginate</div>
        <div class="text-xs text-gruvbox-fg4">total: {{ pagination.total }}</div>
      </div>

      <div class="border-b border-gruvbox-bg2 px-4 py-4">
        <form class="flex flex-col gap-3 md:flex-row" @submit.prevent="applySearch">
          <div class="input-wrapper flex-1">
            <input
              v-model.trim="searchDraft"
              type="text"
              class="input-terminal"
              placeholder="search by username"
            />
          </div>
          <div class="flex gap-3">
            <button type="submit" class="btn-cta btn-cta--primary px-4">[Search]</button>
            <button type="button" class="btn-cta btn-cta--secondary px-4" @click="resetSearch">
              [Reset]
            </button>
          </div>
        </form>
      </div>

      <div v-if="error" class="border-b border-gruvbox-bg2 px-4 py-3 text-sm text-gruvbox-red">
        {{ error }}
      </div>

      <div v-if="loading" class="px-4 py-6 text-sm text-gruvbox-fg4">Loading users...</div>

      <template v-else>
        <div v-if="users.length === 0" class="px-4 py-6 text-sm text-gruvbox-fg4">
          No users found.
        </div>

        <div v-else class="overflow-x-auto">
          <table class="min-w-full text-sm text-gruvbox-fg">
            <thead class="border-b border-gruvbox-bg2 text-left text-gruvbox-fg4">
              <tr>
                <th class="px-4 py-3 font-normal">username</th>
                <th class="px-4 py-3 font-normal">email</th>
                <th class="px-4 py-3 font-normal">role</th>
                <th class="px-4 py-3 font-normal">rating</th>
                <th class="px-4 py-3 font-normal">joined</th>
                <th class="px-4 py-3 font-normal text-right">action</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="user in users"
                :key="user.id"
                class="border-b border-gruvbox-bg2/60 last:border-b-0"
              >
                <td class="px-4 py-3 text-gruvbox-orange">{{ user.username }}</td>
                <td class="px-4 py-3">{{ user.email }}</td>
                <td class="px-4 py-3">
                  <span :class="user.role === 'admin' ? 'text-gruvbox-yellow' : 'text-gruvbox-fg'">
                    {{ user.role }}
                  </span>
                </td>
                <td class="px-4 py-3">{{ user.rating }}</td>
                <td class="px-4 py-3">{{ formatDate(user.createdAt) }}</td>
                <td class="px-4 py-3 text-right">
                  <button
                    class="cursor-pointer text-gruvbox-red hover:opacity-80 disabled:cursor-not-allowed disabled:opacity-50"
                    :disabled="deletingId === user.id || user.id === auth.user?.id"
                    @click="removeUser(user)"
                  >
                    [{{ deletingId === user.id ? 'Deleting...' : 'Delete' }}]
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <div
          class="flex flex-col gap-3 border-t border-gruvbox-bg2 px-4 py-4 text-sm text-gruvbox-fg md:flex-row md:items-center md:justify-between"
        >
          <div class="text-gruvbox-fg4">
            page {{ pagination.page }} / {{ pagination.totalPages || 1 }}
          </div>

          <div class="flex gap-3">
            <button
              class="btn-cta btn-cta--secondary px-4 disabled:cursor-not-allowed disabled:opacity-50"
              :disabled="loading || !pagination.hasPrev"
              @click="changePage(pagination.page - 1)"
            >
              [Prev]
            </button>
            <button
              class="btn-cta btn-cta--secondary px-4 disabled:cursor-not-allowed disabled:opacity-50"
              :disabled="loading || !pagination.hasNext"
              @click="changePage(pagination.page + 1)"
            >
              [Next]
            </button>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { userApi } from '@/api/user'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()

const users = ref([])
const loading = ref(true)
const deletingId = ref(null)
const error = ref('')
const searchDraft = ref('')
const filters = reactive({
  page: 1,
  limit: 10,
  username: '',
})

const pagination = reactive({
  page: 1,
  limit: 10,
  total: 0,
  totalPages: 0,
  hasPrev: false,
  hasNext: false,
})

function formatDate(date) {
  if (!date) return '—'
  return new Date(date).toLocaleDateString()
}

async function loadUsers() {
  loading.value = true
  error.value = ''

  try {
    const { data } = await userApi.adminList(filters)
    users.value = data.users
    Object.assign(pagination, data.pagination)
  } catch (err) {
    error.value = err.response?.data?.error ?? 'Failed to load users.'
  } finally {
    loading.value = false
  }
}

function applySearch() {
  filters.page = 1
  filters.username = searchDraft.value
  loadUsers()
}

function resetSearch() {
  searchDraft.value = ''
  filters.page = 1
  filters.username = ''
  loadUsers()
}

function changePage(page) {
  if (page < 1 || page === filters.page) return
  filters.page = page
  loadUsers()
}

async function removeUser(user) {
  const confirmed = window.confirm(`Delete user "${user.username}"?`)
  if (!confirmed) return

  deletingId.value = user.id
  error.value = ''

  try {
    await userApi.deleteById(user.id)

    if (users.value.length === 1 && filters.page > 1) {
      filters.page -= 1
    }

    await loadUsers()
  } catch (err) {
    error.value = err.response?.data?.error ?? 'Failed to delete user.'
  } finally {
    deletingId.value = null
  }
}

onMounted(() => {
  loadUsers()
})
</script>
