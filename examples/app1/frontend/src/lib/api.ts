import type { CurrentUser, Note, Notebook } from './types'

async function json<T>(input: string, init?: RequestInit): Promise<T> {
  const r = await fetch(input, { ...init, headers: { 'Content-Type': 'application/json', ...(init?.headers ?? {}) } })
  if (!r.ok) throw new Error(`${r.status} ${r.statusText}`)
  if (r.status === 204) return undefined as T
  return r.json() as Promise<T>
}

export const api = {
  me:        () => fetch('/api/me').then(r => (r.ok ? (r.json() as Promise<CurrentUser>) : null)),

  listNotebooks:  () => json<Notebook[]>('/api/notebooks'),
  createNotebook: (title: string) => json<Notebook>('/api/notebooks', {
    method: 'POST', body: JSON.stringify({ title }),
  }),
  renameNotebook: (id: number, title: string) => json<Notebook>(`/api/notebooks/${id}`, {
    method: 'PATCH', body: JSON.stringify({ title }),
  }),
  deleteNotebook: (id: number) => json<void>(`/api/notebooks/${id}`, { method: 'DELETE' }),

  listNotes:  (notebookId: number, signal?: AbortSignal) =>
    json<Note[]>(`/api/notes?notebookId=${notebookId}`, { signal }),
  searchNotes: (q: string, signal?: AbortSignal) =>
    json<Note[]>(`/api/notes/search?q=${encodeURIComponent(q)}`, { signal }),
  getNote:    (id: number) => json<Note>(`/api/notes/${id}`),
  createNote: (notebookId: number, title: string, bodyMd = '') =>
    json<Note>('/api/notes', { method: 'POST', body: JSON.stringify({ notebookId, title, bodyMd }) }),
  updateNote: (id: number, patch: { title?: string; bodyMd?: string }) =>
    json<Note>(`/api/notes/${id}`, { method: 'PATCH', body: JSON.stringify(patch) }),
  deleteNote: (id: number) => json<void>(`/api/notes/${id}`, { method: 'DELETE' }),
}
