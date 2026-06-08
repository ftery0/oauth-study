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

interface ColumnDef {
  id: TaskColumn
  label: string
  bar: string   // top color bar
  badge: string // count badge bg/text
  placeholder: string
}

const COLUMNS: ColumnDef[] = [
  { id: 'todo',  label: 'TO DO',   bar: 'bg-slate-400',   badge: 'bg-slate-200 text-slate-700',     placeholder: '다음에 할 일' },
  { id: 'doing', label: 'DOING',   bar: 'bg-violet-500',  badge: 'bg-violet-100 text-violet-700',   placeholder: '진행 중' },
  { id: 'done',  label: 'DONE',    bar: 'bg-lime-500',    badge: 'bg-lime-100 text-lime-700',       placeholder: '완료' },
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
  if (!window.confirm(`"${b.title}" 보드와 안의 작업 전부 지울까요?`)) return
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

const selectedBoard = computed(() => boards.value.find(b => b._id === selectedBoardId.value))
</script>

<template>
  <!-- Loading -->
  <main v-if="loading" class="min-h-screen flex items-center justify-center bg-slate-900 text-slate-300">
    <div class="font-mono text-sm">› 워크스페이스 여는 중…</div>
  </main>

  <!-- 로그인 전 -->
  <main
    v-else-if="!user"
    class="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-900 via-violet-950 to-slate-900 p-4"
  >
    <div class="w-full max-w-md">
      <div class="mb-6 flex items-center gap-2">
        <span class="inline-block w-2.5 h-2.5 rounded-sm bg-lime-400" />
        <span class="text-xs font-bold uppercase tracking-[0.2em] text-lime-300">group · a</span>
      </div>

      <h1 class="text-5xl font-extrabold tracking-tight text-white mb-3">
        Task<span class="text-violet-400">Board</span>
      </h1>
      <p class="text-slate-400 mb-8">
        팀 작업을 한눈에. <span class="text-white font-semibold">빠르게 옮기고, 빠르게 끝내자.</span>
      </p>

      <div class="bg-slate-800/70 backdrop-blur border border-slate-700 rounded-lg p-6">
        <div
          v-if="errorMsg"
          class="mb-4 rounded bg-red-500/10 border border-red-500/30 px-3 py-2 text-sm text-red-300 font-mono"
          role="alert"
        >
          ! {{ errorMsg }}
        </div>

        <p class="text-sm text-slate-400 mb-5 leading-relaxed">
          Notebook 으로 이미 로그인했다면 폼 없이 바로 들어옵니다.
          <span class="text-slate-300">silent SSO 가 동작 중이에요.</span>
        </p>

        <a
          href="/login"
          class="group flex items-center justify-center gap-2 w-full rounded bg-violet-600 hover:bg-violet-500 active:bg-violet-700 text-white font-semibold px-4 py-3 transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-offset-slate-800 focus:ring-violet-400"
        >
          시작하기
          <span class="transition-transform group-hover:translate-x-0.5">→</span>
        </a>
      </div>

      <p class="mt-6 text-xs text-slate-500 font-mono">
        node · express · vue · mongo
      </p>
    </div>
  </main>

  <!-- 로그인 후 -->
  <div v-else class="min-h-screen flex flex-col bg-slate-100">
    <!-- Dark Top Bar -->
    <header class="bg-slate-900 text-slate-100 px-4 py-2.5 flex items-center gap-3 shadow-lg">
      <div class="flex items-center gap-2 pr-3">
        <span class="inline-block w-2.5 h-2.5 rounded-sm bg-lime-400" />
        <span class="font-extrabold tracking-tight text-base">
          Task<span class="text-violet-400">Board</span>
        </span>
      </div>

      <!-- Board tabs -->
      <nav class="flex-1 flex items-center gap-1 overflow-x-auto scrollbar-hide">
        <button
          v-for="b in boards"
          :key="b._id"
          @click="selectedBoardId = b._id"
          :class="[
            'group relative shrink-0 text-sm px-3 py-1.5 rounded transition-colors',
            selectedBoardId === b._id
              ? 'bg-violet-600 text-white font-semibold'
              : 'text-slate-400 hover:bg-slate-800 hover:text-slate-100',
          ]"
        >
          <span>{{ b.title }}</span>
          <span
            v-if="selectedBoardId === b._id"
            class="ml-2 inline-flex gap-1 text-xs opacity-70"
          >
            <button @click.stop="renameBoard(b)" class="hover:opacity-100" title="이름 변경">✎</button>
            <button @click.stop="deleteBoard(b)" class="hover:opacity-100 hover:text-red-300" title="삭제">✕</button>
          </span>
        </button>
        <button
          @click="addBoard"
          class="shrink-0 text-sm px-3 py-1.5 rounded text-slate-400 hover:bg-slate-800 hover:text-white font-mono"
        >
          + 보드
        </button>
      </nav>

      <div class="flex items-center gap-2 pl-3 border-l border-slate-700">
        <div class="w-7 h-7 rounded bg-violet-500 text-white flex items-center justify-center font-bold uppercase text-xs">
          {{ displayName[0] }}
        </div>
        <span class="text-sm text-slate-200 hidden sm:inline">{{ displayName }}</span>
        <button
          @click="logout"
          class="ml-1 text-xs text-slate-400 hover:text-white px-2 py-1 font-mono uppercase tracking-wider"
        >
          ↪ logout
        </button>
      </div>
    </header>

    <!-- Board area -->
    <section class="flex-1 overflow-hidden">
      <div v-if="!selectedBoardId" class="h-full flex flex-col items-center justify-center text-slate-500 gap-3">
        <div class="text-5xl">◫</div>
        <p class="text-sm">왼쪽 위 <span class="font-mono text-violet-600">+ 보드</span> 로 첫 보드를 만드세요</p>
      </div>

      <div v-else class="h-full flex flex-col">
        <!-- Board title strip -->
        <div class="px-4 py-3 border-b border-slate-200 bg-white flex items-center gap-3">
          <h2 class="font-bold text-slate-900 truncate">{{ selectedBoard?.title }}</h2>
          <span class="font-mono text-xs text-slate-500">{{ tasks.length }} tasks</span>
        </div>

        <!-- Columns -->
        <div class="flex-1 overflow-x-auto p-4 bg-gradient-to-br from-slate-100 to-violet-50">
          <div class="flex gap-4 h-full min-w-fit">
            <div
              v-for="col in COLUMNS"
              :key="col.id"
              class="w-80 flex-shrink-0 flex flex-col bg-white rounded shadow-card"
            >
              <!-- column top color bar -->
              <div :class="['h-1 rounded-t', col.bar]" />
              <!-- column header -->
              <header class="px-3 py-2.5 flex items-center justify-between border-b border-slate-100">
                <div class="flex items-center gap-2">
                  <span class="text-[11px] font-bold uppercase tracking-wider text-slate-700">
                    {{ col.label }}
                  </span>
                  <span
                    :class="['tag font-mono', col.badge]"
                  >
                    {{ columnTasks(col.id).length }}
                  </span>
                </div>
              </header>

              <!-- cards -->
              <div class="p-2 flex flex-col gap-2 flex-1 overflow-y-auto min-h-0">
                <article
                  v-for="t in columnTasks(col.id)"
                  :key="t._id"
                  class="group bg-white border border-slate-200 rounded shadow-card hover:shadow-cardHover hover:-translate-y-0.5 hover:border-violet-300 transition-all px-3 py-2.5 cursor-grab"
                >
                  <div class="text-sm font-medium text-slate-900 leading-snug">{{ t.title }}</div>
                  <div class="mt-2 flex items-center justify-between">
                    <span class="font-mono text-[10px] text-slate-400">
                      {{ new Date(t.updatedAt).toLocaleDateString('ko-KR', { month: '2-digit', day: '2-digit' }) }}
                    </span>
                    <div class="flex gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity">
                      <button
                        v-if="col.id !== 'todo'"
                        @click="moveTask(t, -1)"
                        class="w-6 h-6 flex items-center justify-center text-slate-400 hover:text-violet-600 hover:bg-violet-50 rounded text-xs"
                        title="이전 컬럼"
                      >◀</button>
                      <button
                        v-if="col.id !== 'done'"
                        @click="moveTask(t, 1)"
                        class="w-6 h-6 flex items-center justify-center text-slate-400 hover:text-violet-600 hover:bg-violet-50 rounded text-xs"
                        title="다음 컬럼"
                      >▶</button>
                      <button
                        @click="deleteTask(t)"
                        class="w-6 h-6 flex items-center justify-center text-slate-400 hover:text-red-600 hover:bg-red-50 rounded text-xs"
                        title="삭제"
                      >✕</button>
                    </div>
                  </div>
                </article>
                <div
                  v-if="columnTasks(col.id).length === 0"
                  class="text-xs text-slate-300 italic px-3 py-4 text-center"
                >
                  {{ col.placeholder }}
                </div>
              </div>

              <!-- add task -->
              <footer class="p-2 border-t border-slate-100 bg-slate-50/50">
                <form @submit.prevent="addTask(col.id)">
                  <input
                    v-model="newTaskInput[col.id]"
                    type="text"
                    placeholder="+ 추가 (Enter)"
                    class="w-full text-sm rounded border border-transparent bg-transparent hover:bg-white hover:border-slate-200 focus:bg-white focus:border-violet-400 px-2 py-1.5 focus:outline-none focus:ring-1 focus:ring-violet-400 placeholder:text-slate-400 transition-colors"
                  />
                </form>
              </footer>
            </div>

            <!-- spacer column to give breathing room when scrolling -->
            <div class="w-2 flex-shrink-0" />
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
.scrollbar-hide::-webkit-scrollbar {
  display: none;
}
.scrollbar-hide {
  scrollbar-width: none;
}
</style>
