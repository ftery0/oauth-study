import type { Board, CurrentUser, Task, TaskColumn } from './types'

async function json<T>(input: string, init?: RequestInit): Promise<T> {
  const r = await fetch(input, {
    ...init,
    headers: { 'Content-Type': 'application/json', ...(init?.headers ?? {}) },
  })
  if (!r.ok) throw new Error(`${r.status} ${r.statusText}`)
  if (r.status === 204) return undefined as T
  return r.json() as Promise<T>
}

export const api = {
  me:      () => fetch('/api/me').then(r => (r.ok ? (r.json() as Promise<CurrentUser>) : null)),
  logout:  () => fetch('/api/logout', { method: 'POST' }),

  listBoards:  () => json<Board[]>('/api/boards'),
  createBoard: (title: string) => json<Board>('/api/boards', { method: 'POST', body: JSON.stringify({ title }) }),
  renameBoard: (id: string, title: string) =>
    json<Board>(`/api/boards/${id}`, { method: 'PATCH', body: JSON.stringify({ title }) }),
  deleteBoard: (id: string) => json<void>(`/api/boards/${id}`, { method: 'DELETE' }),

  listTasks:   (boardId: string) => json<Task[]>(`/api/tasks?boardId=${boardId}`),
  createTask:  (boardId: string, title: string, column: TaskColumn = 'todo') =>
    json<Task>('/api/tasks', { method: 'POST', body: JSON.stringify({ boardId, title, column }) }),
  updateTask:  (id: string, patch: Partial<Pick<Task, 'title' | 'column' | 'position' | 'done'>>) =>
    json<Task>(`/api/tasks/${id}`, { method: 'PATCH', body: JSON.stringify(patch) }),
  deleteTask:  (id: string) => json<void>(`/api/tasks/${id}`, { method: 'DELETE' }),
}
