<script setup>
import { ref, onMounted } from 'vue'

const user = ref(null)
const loading = ref(true)
const errorMsg = ref(null)

onMounted(async () => {
  const params = new URLSearchParams(window.location.search)
  const e = params.get('error')
  const loggedOut = params.get('logout')
  if (e) {
    errorMsg.value = e
    window.history.replaceState({}, '', '/')
  }
  if (loggedOut) {
    // 로그아웃 직후 자동 재로그인 루프 방지
    window.history.replaceState({}, '', '/')
    loading.value = false
    return
  }

  try {
    const r = await fetch('/api/me')
    if (r.ok) {
      user.value = await r.json()
      loading.value = false
      return
    }
    if (e) {
      loading.value = false
      return
    }
    // Keycloak 패턴: 401 + 에러 없음 → 즉시 OAuth 시작
    window.location.href = '/login'
  } catch {
    loading.value = false
  }
})

async function logout() {
  await fetch('/api/logout', { method: 'POST' })
  window.location.href = '/?logout=1'
}
</script>

<template>
  <div class="min-h-screen bg-slate-50 flex items-center justify-center p-4 antialiased text-slate-900">
    <main v-if="loading" class="text-slate-500 text-sm">로딩 중...</main>

    <!-- 로그인 전 -->
    <main
      v-else-if="!user"
      class="w-full max-w-md bg-white rounded-2xl shadow-lg p-6 sm:p-8"
    >
      <header class="mb-6">
        <div class="inline-flex items-center gap-2 mb-2">
          <span class="inline-block w-2 h-2 rounded-full bg-emerald-500"></span>
          <span class="text-xs font-medium text-emerald-700 uppercase tracking-wide">group-a</span>
        </div>
        <h1 class="text-2xl font-semibold">App 2</h1>
        <p class="mt-1 text-sm text-slate-500">Node Express + Vue 3</p>
      </header>

      <div
        v-if="errorMsg"
        class="mb-4 rounded-lg bg-red-50 border border-red-200 px-3 py-2 text-sm text-red-700"
        role="alert"
      >
        오류: {{ errorMsg }}
      </div>

      <div class="bg-emerald-50 border border-emerald-200 rounded-lg p-4 mb-5 text-sm text-emerald-900">
        <p class="font-medium mb-1">silent SSO 시연</p>
        <p>이미 같은 그룹(group-a)의 다른 앱에 로그인되어 있다면 폼 없이 자동 로그인됩니다.</p>
      </div>

      <a
        href="/login"
        class="block w-full text-center rounded-lg bg-emerald-600 hover:bg-emerald-700 active:bg-emerald-800 text-white font-medium px-4 py-3 transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-emerald-500"
      >
        OAuth 로그인
      </a>
    </main>

    <!-- 로그인 후 -->
    <main
      v-else
      class="w-full max-w-md bg-white rounded-2xl shadow-lg p-6 sm:p-8"
    >
      <div class="flex items-center gap-4 mb-6">
        <div
          class="w-12 h-12 rounded-full bg-emerald-600 text-white flex items-center justify-center font-semibold uppercase text-lg"
        >
          {{ user.sub[0] }}
        </div>
        <div>
          <h1 class="text-xl font-semibold">{{ user.sub }}</h1>
          <p class="text-sm text-slate-500">App 2 로그인 완료</p>
        </div>
      </div>

      <div class="space-y-2 mb-6 text-sm">
        <div class="flex justify-between border-b pb-2">
          <span class="text-slate-500">sub</span>
          <span class="font-mono text-slate-900">{{ user.sub }}</span>
        </div>
        <div class="flex justify-between border-b pb-2">
          <span class="text-slate-500">client_id</span>
          <span class="font-mono text-slate-900">{{ user.client_id }}</span>
        </div>
        <div class="flex justify-between">
          <span class="text-slate-500">scope</span>
          <span class="font-mono text-slate-900">{{ user.scope || '(none)' }}</span>
        </div>
      </div>

      <button
        @click="logout"
        class="w-full rounded-lg border border-slate-300 hover:bg-slate-100 active:bg-slate-200 text-slate-700 font-medium px-4 py-3 transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-slate-400"
      >
        로그아웃
      </button>
    </main>
  </div>
</template>
