<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { api } from './api'
import type { Board, CurrentUser, Task, TaskColumn } from './types'

const user = ref<CurrentUser | null>(null)
const loading = ref(true)
const errorMsg = ref<string | null>(null)

const boards = ref<Board[]>([])
const selectedBoardId = ref<string | null>(null)
const tasks = ref<Task[]>([])

const newTaskInput = ref<Record<TaskColumn, string>>({ todo: '', doing: '', done: '' })

const COLUMNS: { id: TaskColumn; label: string; color: string }[] = [
  { id: 'todo',  label: 'To Do', color: 'bg-slate-100 text-slate-700' },
  { id: 'doing', label: 'Doing', color: 'bg-amber-100 text-amber-800' },
  { id: 'done',  label: 'Done',  color: 'bg-emerald-100 text-emerald-800' },
]

const displayName = computed(() => user.value?.display_name ?? user.value?.sub ?? '')

onMounted(async () => {
  const params = new URLSearchParams(window.location.search)
  const e = params.get('error')
  const loggedOut = params.get('logout')
  if (e) {
    errorMsg.value = e
    window.history.replaceState({}, '', '/')
  }
  if (loggedOut) {
    window.history.replaceState({}, '', '/')
    loading.value = false
    return
  }
  try {
    const r = await fetch('/api/me')
    if (r.ok) {
      user.value = (await r.json()) as CurrentUser
      loading.value = false
      await loadBoards()
      return
    }
    if (e) {
      loading.value = false
      return
    }
    window.location.href = '/login'
  } catch {
    loading.value = false
  }
})

async function loadBoards(): Promise<void> {
  const list = await api.listBoards()
  boards.value = list
  if (list.length > 0 && !selectedBoardId.value) {
    selectedBoardId.value = list[0]._id
  }
}

watch(selectedBoardId, async (id) => {
  if (!id) {
    tasks.value = []
    return
  }
  tasks.value = await api.listTasks(id)
})

function columnTasks(col: TaskColumn): Task[] {
  return tasks.value.filter(t => t.column === col).sort((a, b) => a.position - b.position)
}

async function addBoard(): Promise<void> {
  const title = window.prompt('새 보드 이름')?.trim()
  if (!title) return
  const created = await api.createBoard(title)
  boards.value = [created, ...boards.value]
  selectedBoardId.value = created._id
}

async function renameBoard(b: Board): Promise<void> {
  const title = window.prompt('보드 이름 변경', b.title)?.trim()
  if (!title || title === b.title) return
  const updated = await api.renameBoard(b._id, title)
  boards.value = boards.value.map(x => (x._id === b._id ? updated : x))
}

async function deleteBoard(b: Board): Promise<void> {
  if (!window.confirm(`"${b.title}" 와 안의 작업을 모두 삭제할까요?`)) return
  await api.deleteBoard(b._id)
  boards.value = boards.value.filter(x => x._id !== b._id)
  if (selectedBoardId.value === b._id) {
    selectedBoardId.value = boards.value[0]?._id ?? null
  }
}

async function addTask(column: TaskColumn): Promise<void> {
  const title = newTaskInput.value[column].trim()
  if (!title || !selectedBoardId.value) return
  const created = await api.createTask(selectedBoardId.value, title, column)
  tasks.value = [...tasks.value, created]
  newTaskInput.value[column] = ''
}

async function moveTask(t: Task, dir: -1 | 1): Promise<void> {
  const order: TaskColumn[] = ['todo', 'doing', 'done']
  const idx = order.indexOf(t.column)
  const next = order[idx + dir]
  if (!next) return
  const updated = await api.updateTask(t._id, { column: next })
  tasks.value = tasks.value.map(x => (x._id === t._id ? updated : x))
}

async function deleteTask(t: Task): Promise<void> {
  await api.deleteTask(t._id)
  tasks.value = tasks.value.filter(x => x._id !== t._id)
}

async function logout(): Promise<void> {
  await api.logout()
  window.location.href = '/?logout=1'
}
</script>

<template>
  <div class="min-h-screen bg-slate-50 text-slate-900 antialiased">
    <main v-if="loading" class="min-h-screen flex items-center justify-center text-slate-500 text-sm">
      로딩 중...
    </main>

    <!-- 로그인 전 -->
    <main
      v-else-if="!user"
      class="min-h-screen flex items-center justify-center p-4"
    >
      <div class="w-full max-w-md bg-white rounded-2xl shadow-lg p-6 sm:p-8">
        <header class="mb-6">
          <div class="inline-flex items-center gap-2 mb-2">
            <span class="inline-block w-2 h-2 rounded-full bg-emerald-500" />
            <span class="text-xs font-medium text-emerald-700 uppercase tracking-wide">group-a</span>
          </div>
          <h1 class="text-2xl font-semibold">TaskBoard</h1>
          <p class="mt-1 text-sm text-slate-500">사내 칸반 · Node Express + Vue 3</p>
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
          <p>이미 같은 그룹(group-a)의 Notebook 에 로그인되어 있다면 폼 없이 자동 로그인됩니다.</p>
        </div>

        <a
          href="/login"
          class="block w-full text-center rounded-lg bg-emerald-600 hover:bg-emerald-700 active:bg-emerald-800 text-white font-medium px-4 py-3 transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-emerald-500"
        >
          OAuth 로그인
        </a>
      </div>
    </main>

    <!-- 로그인 후 -->
    <div v-else class="min-h-screen flex flex-col">
      <!-- Header -->
      <header class="bg-white border-b border-slate-200 px-4 py-3 flex items-center gap-3">
        <span class="inline-flex items-center gap-2">
          <span class="w-2 h-2 rounded-full bg-emerald-500" />
          <span class="text-xs font-medium text-emerald-700 uppercase tracking-wide">group-a</span>
        </span>
        <h1 class="font-semibold">TaskBoard</h1>
        <div class="flex-1" />
        <div class="flex items-center gap-2">
          <div class="w-8 h-8 rounded-full bg-emerald-600 text-white flex items-center justify-center font-semibold uppercase text-sm">
            {{ displayName[0] }}
          </div>
          <span class="text-sm text-slate-700">{{ displayName }}</span>
          <button
            @click="logout"
            class="ml-2 text-sm text-slate-600 hover:text-slate-900 rounded border border-slate-300 px-3 py-1.5 hover:bg-slate-100"
          >
            로그아웃
          </button>
        </div>
      </header>

      <!-- Body: Sidebar + Board -->
      <div class="flex-1 flex overflow-hidden">
        <!-- Sidebar: boards -->
        <aside class="w-56 bg-white border-r border-slate-200 flex flex-col">
          <div class="px-3 py-2 flex items-center justify-between border-b border-slate-200">
            <span class="text-xs uppercase tracking-wide text-slate-500">Boards</span>
            <button @click="addBoard" class="text-emerald-600 hover:text-emerald-800 text-lg leading-none">+</button>
          </div>
          <ul class="flex-1 overflow-y-auto">
            <li v-if="boards.length === 0" class="px-3 py-4 text-sm text-slate-400">보드가 없습니다</li>
            <li
              v-for="b in boards"
              :key="b._id"
              :class="[
                'px-3 py-2 text-sm cursor-pointer group flex items-center justify-between',
                selectedBoardId === b._id ? 'bg-emerald-50 text-emerald-900' : 'hover:bg-slate-50',
              ]"
              @click="selectedBoardId = b._id"
            >
              <span class="truncate">{{ b.title }}</span>
              <span class="hidden group-hover:flex gap-1 text-xs text-slate-500">
                <button @click.stop="renameBoard(b)" class="hover:text-slate-900">✎</button>
                <button @click.stop="deleteBoard(b)" class="hover:text-red-600">✕</button>
              </span>
            </li>
          </ul>
        </aside>

        <!-- Board (3 columns) -->
        <section class="flex-1 overflow-x-auto p-4">
          <div v-if="!selectedBoardId" class="h-full flex items-center justify-center text-slate-400 text-sm">
            보드를 만들거나 선택하세요
          </div>
          <div v-else class="flex gap-4 h-full min-w-fit">
            <div
              v-for="col in COLUMNS"
              :key="col.id"
              class="w-72 flex-shrink-0 flex flex-col bg-white border border-slate-200 rounded-lg"
            >
              <header class="px-3 py-2 border-b border-slate-200 flex items-center justify-between">
                <span :class="['text-xs px-2 py-0.5 rounded-full font-medium', col.color]">
                  {{ col.label }}
                </span>
                <span class="text-xs text-slate-400">{{ columnTasks(col.id).length }}</span>
              </header>
              <div class="p-2 flex flex-col gap-2 flex-1 overflow-y-auto">
                <article
                  v-for="t in columnTasks(col.id)"
                  :key="t._id"
                  class="bg-slate-50 border border-slate-200 rounded-md px-3 py-2 group"
                >
                  <div class="text-sm">{{ t.title }}</div>
                  <div class="mt-1 flex items-center justify-between text-xs text-slate-400">
                    <span>{{ new Date(t.updatedAt).toLocaleDateString('ko-KR') }}</span>
                    <div class="hidden group-hover:flex gap-1">
                      <button
                        v-if="col.id !== 'todo'"
                        @click="moveTask(t, -1)"
                        class="hover:text-slate-900"
                        title="이전 컬럼"
                      >◀</button>
                      <button
                        v-if="col.id !== 'done'"
                        @click="moveTask(t, 1)"
                        class="hover:text-slate-900"
                        title="다음 컬럼"
                      >▶</button>
                      <button @click="deleteTask(t)" class="hover:text-red-600" title="삭제">✕</button>
                    </div>
                  </div>
                </article>
              </div>
              <footer class="p-2 border-t border-slate-200">
                <form @submit.prevent="addTask(col.id)" class="flex gap-2">
                  <input
                    v-model="newTaskInput[col.id]"
                    type="text"
                    placeholder="+ 새 작업"
                    class="flex-1 text-sm rounded border border-slate-300 px-2 py-1 focus:outline-none focus:ring-2 focus:ring-emerald-500"
                  />
                </form>
              </footer>
            </div>
          </div>
        </section>
      </div>
    </div>
  </div>
</template>
